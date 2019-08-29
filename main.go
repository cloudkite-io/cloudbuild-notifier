package main

import (
	"context"
	"github.com/spf13/viper"
)

func init() {
	viper.AutomaticEnv()
}

type Config struct {
	projectID       string
	slackWebhookURL string
	subName         string
}

func main() {
	subName := viper.GetString("GCLOUD_PUBSUB_SUBSCRIPTION_NAME")
	if subName == "" {
		subName = "cloudbuild-notifier-subscription"
	}
	config := &Config{
		projectID:       viper.GetString("GCLOUD_PROJECT_ID"),
		subName:         subName,
		slackWebhookURL: viper.GetString("SLACK_WEBHOOK_URL"),
	}
	ctx := context.Background()
	notify(ctx, config)
}
