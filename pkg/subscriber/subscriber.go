package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	notifier "github.com/cloudkite-io/cloudbuild-notifier"
)

// Config is a notifiers configuratios.
type Config struct {
	ProjectID string
	SubName   string
}

type subscriber struct {
	config *Config
	client *pubsub.Client
	topic  *pubsub.Topic
}

type cloudbuildResponse struct {
	Status     string    `json:"status"`
	CreateTime time.Time `json:"createTime"`
	LogURL     string    `json:"logUrl"`
}

// New creates a notifier
func New(config *Config) (notifier.Subscriber, error) {
	ctx := context.Background()
	log.Println("Creating pubsub client")

	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return subscriber{}, fmt.Errorf("failed creating pubsub client: %s", err)
	}

	topic := client.Topic("cloud-builds")

	return subscriber{config: config, client: client, topic: topic}, nil
}

// Receive listens for cloudbuild notificatios.
func (s subscriber) Receive() (text string, err error) {
	ctx := context.Background()
	subscription, err := s.getOrCreateSubscription(ctx, s.topic)
	if err != nil {
		return "", fmt.Errorf("failed getting/creating sub: %s", err)
	}

	log.Println("Starting to listen to events...")
	err = subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var cloudbuildResponse cloudbuildResponse
		err := json.Unmarshal(msg.Data, &cloudbuildResponse)
		if err != nil {
			err = fmt.Errorf("failed unmarshaling json from cloudbuild response: %s", err)
			return
		}

		if stringInSlice(cloudbuildResponse.Status, []string{"FAILURE", "INTERNAL_ERROR", "TIMEOUT", "CANCELLED"}) {
			text = fmt.Sprintf("Something went wrong in Cloudbuild! \nProject: %s \nStatus: %s \nLog URL: %s",
				s.config.ProjectID, cloudbuildResponse.Status, cloudbuildResponse.LogURL)
		}
		msg.Ack()
	})

	if err != nil {
		return "", err
	}

	return text, nil
}

func (s subscriber) getOrCreateSubscription(ctx context.Context, topic *pubsub.Topic) (*pubsub.Subscription, error) {
	exists, err := s.client.Subscription(s.config.SubName).Exists(ctx)
	if err != nil {
		return &pubsub.Subscription{}, fmt.Errorf("failed checking if subscription %s exists: %s ", s.config.SubName, err)
	}

	if !exists {
		log.Printf("Subscription does not exists. Creating new one with name %s\n", s.config.SubName)
		pubSubConfig := pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 10 * time.Second,
		}
		return s.createSubscription(ctx, pubSubConfig)
	}
	sub := s.client.Subscription(s.config.SubName)
	log.Printf("Subscription exists. Getting %s\n", s.config.SubName)

	return sub, nil
}

func (s subscriber) createSubscription(ctx context.Context, config pubsub.SubscriptionConfig) (*pubsub.Subscription, error) {
	sub, err := s.client.CreateSubscription(ctx, s.config.SubName, pubsub.SubscriptionConfig{
		Topic:       s.topic,
		AckDeadline: 10 * time.Second,
	})
	if err != nil {
		return &pubsub.Subscription{}, fmt.Errorf("failed creating subscription %s: %s", s.config.SubName, err)
	}
	log.Printf("Created subscription %s\n", s.config.SubName)
	return sub, nil
}

func stringInSlice(needle string, haystack []string) bool {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}
	return false
}
