package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/notifier/slack"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/subscriber"
	"github.com/spf13/viper"

	"cloud.google.com/go/pubsub"
)

func init() {
	viper.AutomaticEnv()
}

func main() {
	subName := viper.GetString("GCLOUD_PUBSUB_SUBSCRIPTION_NAME")
	if subName == "" {
		subName = "cloudbuild-notifier-subscription"
	}

	config := &subscriber.Config{
		ProjectID: viper.GetString("GCLOUD_PROJECT_ID"),
		SubName:   subName,
	}

	sub, err := subscriber.New(config)
	if err != nil {
		log.Printf("failed creating cloudbuild subscriber: %s", err)
	}

	msg := make(chan *pubsub.Message)
	go func() {
		err = sub.Receive(msg)
		if err != nil {
			log.Printf("failed receiving cloudbuild notification: %s", err)
		}
	}()

	notifier := slack.New(viper.GetString("SLACK_WEBHOOK_URL"))
	for {
		err = handleMessage(config.ProjectID, <-msg, notifier)
		if err != nil {
			log.Printf("failed handling pubsub message: %s", err)
		}
	}
}

func handleMessage(projectID string, msg *pubsub.Message, notifier cloudbuildnotifier.Notifier) error {
	var resp cloudbuildResponse
	err := json.Unmarshal(msg.Data, &resp)
	if err != nil {
		return fmt.Errorf("failed unmarshaling json from cloudbuild response: %s", err)
	}

	if !stringInSlice(resp.Status, []string{"FAILURE", "INTERNAL_ERROR", "TIMEOUT", "CANCELLED"}) {
		return nil
	}

	text := fmt.Sprintf("Something went wrong in Cloudbuild! \nProject: %s \nStatus: %s \nLog URL: %s",
		projectID, resp.Status, resp.LogURL)

	err = notifier.Send(text)
	if err != nil {
		msg.Nack()
		return fmt.Errorf("failed sending to slack: %s", err)
	}

	msg.Ack()

	return nil
}

func stringInSlice(needle string, haystack []string) bool {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}
	return false
}

type cloudbuildResponse struct {
	Status     string    `json:"status"`
	CreateTime time.Time `json:"createTime"`
	LogURL     string    `json:"logUrl"`
}
