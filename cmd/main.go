package main

import (
	"cloud.google.com/go/pubsub"
	"encoding/json"
	"fmt"
	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/cloudbuild"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/notifier/slack"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/subscriber"
	"github.com/spf13/viper"
	"log"
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
	cloudbuildClient, _ := cloudbuild.New(config.ProjectID)
	for {
		err = handleMessage(config.ProjectID, <-msg, notifier, cloudbuildClient)
		if err != nil {
			log.Printf("failed handling pubsub message: %s", err)
		}
	}
}

func handleMessage(projectID string, msg *pubsub.Message, notifier cloudbuildnotifier.Notifier, cloudbuild *cloudbuild.CloudbuildClient) error {
	var resp cloudbuildnotifier.CloudbuildResponse
	err := json.Unmarshal(msg.Data, &resp)
	if err != nil {
		return fmt.Errorf("failed unmarshaling json from cloudbuild response: %s", err)
	}

	buildParams, err := cloudbuild.GetBuildParameterss(msg.Attributes["buildId"])
	if err != nil {
		return fmt.Errorf("Failed getting build parameters from cloudbuild for build %s: %s",
			msg.Attributes["buildId"], err)
	}

	notifier.Send(resp, buildParams)
	if err != nil {
		msg.Nack()
		return fmt.Errorf("failed sending to slack: %s", err)
	}
	msg.Ack()
	return nil
}
