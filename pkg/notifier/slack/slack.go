package slack

import (
	"fmt"
	"log"
	"strings"

	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/cloudbuild"
	"github.com/slack-go/slack"
)

type notifier struct {
	webhookURL     string
	filteredStatus []string
	filteredBranch []string
	filteredSource string
}

// New creates a slack notifier.
func New(webhookURL string, filteredStatus []string, filteredBranch []string, filteredSource string) cloudbuildnotifier.Notifier {
	return notifier{webhookURL, filteredStatus, filteredBranch, filteredSource}
}

func (n notifier) Send(cloudbuildResponse cloudbuildnotifier.CloudbuildResponse, buildParams cloudbuild.BuildParameters) error {

	var color string
	switch {
	case stringInSlice(cloudbuildResponse.Status, []string{"FAILURE", "INTERNAL_ERROR", "TIMEOUT"}):
		color = "danger"
	case cloudbuildResponse.Status == "CANCELLED":
		color = "#C0C0C0"
	case cloudbuildResponse.Status == "SUCCESS":
		color = "good"
	}

	if len(n.filteredStatus) > 0 {
		if !stringInSlice(cloudbuildResponse.Status, n.filteredStatus) {
			return nil
		}
	}

	if len(n.filteredBranch) > 0 {
		if !stringInSlice(buildParams.BRANCH_NAME, n.filteredBranch) {
			return nil
		}
	}

	if n.filteredSource != "" {
		if !strings.Contains(buildParams.REPO_NAME, n.filteredSource) {
			return nil
		}
	}

	if cloudbuildResponse.Status == "QUEUED" || cloudbuildResponse.Status == "WORKING" {
		return nil
	}

	attachment := slack.Attachment{
		Title: fmt.Sprintf("Cloudbuild: %s", cloudbuildResponse.Status),
		Color: color,
		Text: fmt.Sprintf("Repo: %s\nBranch: %s\nCommit SHA: %s\nTrigger: %s\n",
			buildParams.REPO_NAME, buildParams.BRANCH_NAME, buildParams.COMMIT_SHA, buildParams.TRIGGER_NAME),
		Actions: []slack.AttachmentAction{
			{
				Text: "View Logs",
				Type: "button",
				URL:  cloudbuildResponse.LogURL,
			},
		},
	}

	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(n.webhookURL, &msg)
	if err != nil {
		return fmt.Errorf("failed posting to webhook %s: %s", n.webhookURL, err)
	}

	log.Printf("Sent %s Slack message for build %s\n", n.webhookURL, buildParams.Id)

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
