package subscriber

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
)

// Config is a notifiers configuration.
type Config struct {
	ProjectID string
	SubName   string
}

// Subscriber listens for cloudbuild events.
type Subscriber struct {
	config *Config
	client *pubsub.Client
	topic  *pubsub.Topic
}

// New creates a notifier
func New(config *Config) (Subscriber, error) {
	ctx := context.Background()
	log.Println("Creating pubsub client")

	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return Subscriber{}, fmt.Errorf("failed creating pubsub client: %s", err)
	}

	topic := client.Topic("cloud-builds")

	return Subscriber{config: config, client: client, topic: topic}, nil
}

// Receive listens for cloudbuild notifications.
func (s Subscriber) Receive(msg chan<- *pubsub.Message) (err error) {
	ctx := context.Background()
	subscription, err := s.getOrCreateSubscription(ctx, s.topic)
	if err != nil {
		return fmt.Errorf("failed getting/creating sub: %s", err)
	}

	log.Println("Starting to listen to events...")
	err = subscription.Receive(ctx, func(ctx context.Context, pbmsg *pubsub.Message) {
		msg <- pbmsg
	})
	if err != nil {
		return fmt.Errorf("error while receving pubsub message: %s", err)
	}

	return nil
}

func (s Subscriber) getOrCreateSubscription(ctx context.Context, topic *pubsub.Topic) (*pubsub.Subscription, error) {
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

func (s Subscriber) createSubscription(ctx context.Context, config pubsub.SubscriptionConfig) (*pubsub.Subscription, error) {
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
