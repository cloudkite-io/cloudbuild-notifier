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
	subscription := client.Subscription(config.subName)
	log.Println("Getting pubsub subscription")
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

func stringInSlice(needle string, haystack []string) bool {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}
	return false
}
