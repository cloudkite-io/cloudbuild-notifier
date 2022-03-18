package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/pubsub"
	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/cloudbuild"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/notifier/slack"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/subscriber"
	"github.com/spf13/viper"
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

	// Get NOTIFICATION_FILTERS environment variable and convert it to an Array
	notificationFilters := viper.GetString("NOTIFICATION_FILTERS")
	var notificationFiltersArr slack.FiltersType
	notificationFiltersErr := json.Unmarshal([]byte(notificationFilters), &notificationFiltersArr)
	if notificationFiltersErr != nil {
		log.Printf("No notification filters found")
	}
	notifier := slack.New(viper.GetString("SLACK_WEBHOOK_URL"), notificationFiltersArr)
	cloudbuildClient, _ := cloudbuild.New(config.ProjectID)

	// HTTP Handler
	http.HandleFunc("/", httpHandler(notifier, cloudbuildClient))
	http.HandleFunc("/status", statusHandler)
	go http.ListenAndServe(":8000", nil)

	// Pubsub handler
	for {
		err = handleMessage(<-msg, notifier, cloudbuildClient)
		if err != nil {
			log.Printf("failed handling pubsub message: %s", err)
		}
	}
}

type pubSubHTTPMessage struct {
	Message struct {
		Data       []byte            `json:"data,omitempty"`
		Attributes map[string]string `json:"attributes,omitempty"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Online")
}

func httpHandler(n cloudbuildnotifier.Notifier, c *cloudbuild.CloudbuildClient) http.HandlerFunc {
	pubsubHttp := &pubSubHTTPMessage{}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(pubsubHttp); err != nil {
			http.Error(w, fmt.Sprintf("Could not decode request body: %v", err), http.StatusBadRequest)
			return
		}

		m := &pubsub.Message{
			ID:         pubsubHttp.Message.Attributes["buildId"],
			Attributes: pubsubHttp.Message.Attributes,
			Data:       pubsubHttp.Message.Data,
		}

		if err := handleMessage(m, n, c); err != nil {
			errorMessage := fmt.Sprintf("Error handling http message: %s", err)
			fmt.Printf(errorMessage)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			return
		}
	}
}

func handleMessage(msg *pubsub.Message, notifier cloudbuildnotifier.Notifier, cloudbuild *cloudbuild.CloudbuildClient) error {
	defer func() {
		// Recover from a http message that can't be Acked
		if err := recover(); err != nil {
			fmt.Printf("PubSub HTTP Done.")
		}
	}()

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

	err = notifier.Send(resp, buildParams)
	if err != nil {
		msg.Nack()
		return fmt.Errorf("failed sending to slack: %s", err)
	}
	// This will panic if from an HTTP message since they can't be acked.
	msg.Ack()
	return nil
}
