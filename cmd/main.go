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
	"net/http"
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
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
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
			ID:   pubsubHttp.Message.ID,
			Data: pubsubHttp.Message.Data,
		}

		handleMessage(m, n, c)
	}
}

func handleMessage(msg *pubsub.Message, notifier cloudbuildnotifier.Notifier, cloudbuild *cloudbuild.CloudbuildClient) error {
	fmt.Printf("Handling msg: %v", msg.Data)
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
