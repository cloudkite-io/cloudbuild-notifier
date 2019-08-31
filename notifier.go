package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"log"
	"time"
)

type CloudbuildResponse struct {
	Status     string    `json:"status"`
	CreateTime time.Time `json:"createTime"`
	LogUrl     string    `json:"logUrl"`
}

func notify(ctx context.Context, config *Config) {
	log.Println("Creating pubsub client")
	client, err := pubsub.NewClient(ctx, config.projectID)
	if err != nil {
		log.Println(err)
	}

	topic := client.Topic("cloud-builds")
	err, subscription := getOrCreateSubscription(client, config, ctx, topic)

	cloudbuildResponse := CloudbuildResponse{}
	log.Println("Starting to listen to events...")
	err = subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		_ = json.Unmarshal(msg.Data, &cloudbuildResponse)
		if stringInSlice(cloudbuildResponse.Status, []string{"FAILURE", "INTERNAL_ERROR", "TIMEOUT", "CANCELLED"}) {
			sendToSlack(cloudbuildResponse, config)
		}
		msg.Ack()
	})
	if err != nil {
		log.Println(err)
	}
}

func getOrCreateSubscription(client *pubsub.Client, config *Config, ctx context.Context, topic *pubsub.Topic) (error, *pubsub.Subscription) {
	exists, err := client.Subscription(config.subName).Exists(ctx)
	var subscription *pubsub.Subscription
	if !exists {
		log.Printf("Subscription does not exists. Creating new one with name %s\n", config.subName)
		subscription, err = client.CreateSubscription(ctx, config.subName, pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 10 * time.Second,
		})
		if err != nil {
			log.Println(err)
		}
		log.Printf("Created subscription %s\n", config.subName)
	} else {
		subscription = client.Subscription(config.subName)
		log.Printf("Subscription exists. Getting %s\n", config.subName)
	}
	if err != nil {
		log.Println(err)
	}
	return err, subscription
}

func stringInSlice(needle string, haystack []string) bool {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}
	return false
}
