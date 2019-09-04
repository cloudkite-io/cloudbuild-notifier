package slack

import (
	"fmt"
	"log"

	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/nlopes/slack"
)

type notifier struct {
	webhookURL string
}

// New creates a slack notifier.
func New(webhookURL string) cloudbuildnotifier.Notifier {
	return notifier{webhookURL}
}

func (n notifier) Send(text string) error {
	attachment := slack.Attachment{
		Color: "danger",
		Text:  text,
	}

	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(n.webhookURL, &msg)
	if err != nil {
		return fmt.Errorf("failed posting to webhook %s: %s", n.webhookURL, err)
	}

	log.Printf("Sent %s Slack message for %s\n", n.webhookURL, text)
	return nil
}
