package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
)

func sendToSlack(cloudbuildResponse CloudbuildResponse, config *Config) {
	text := fmt.Sprintf("Something went wrong in Cloudbuild! \nProject: %s \nStatus: %s \nLog URL: %s",
		config.projectID, cloudbuildResponse.Status, cloudbuildResponse.LogUrl)
	attachment := slack.Attachment{
		Color: "danger",
		Text:  text,
	}

	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.PostWebhook(config.slackWebhookURL, &msg)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Sent %s Slack message for %s\n", cloudbuildResponse.Status, cloudbuildResponse.LogUrl)
}
