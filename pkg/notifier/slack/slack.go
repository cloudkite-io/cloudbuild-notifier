package slack

import (
	"fmt"
	"log"
	"strings"

	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/cloudbuild"
	"github.com/slack-go/slack"
)

type FiltersType []struct {
	Status   []string
	Sources  []string
	Branches []string
	Operator string
}

type notifier struct {
	webhookURL             string
	notificationFiltersArr FiltersType
}

// New creates a slack notifier.
func New(webhookURL string, notificationFiltersArr FiltersType) cloudbuildnotifier.Notifier {
	return notifier{webhookURL, notificationFiltersArr}
}

func (n notifier) Send(cloudbuildResponse cloudbuildnotifier.CloudbuildResponse, buildParams cloudbuild.BuildParameters) error {

	// Check if there are any notification filters that have been passed
	// Notification filters are used to determine whether the build status should be sent to Slack channel
	if len(n.notificationFiltersArr) > 0 {
		var notify bool
		for i := 0; i < len(n.notificationFiltersArr); i++ {
			if n.notificationFiltersArr[i].Operator == "and" {
				notify = true
				// "and" case where all conditions passed must be fulfilled for messages to be sent to Slack
				if len(n.notificationFiltersArr[i].Status) > 0 {
					if !stringInSlice(cloudbuildResponse.Status, n.notificationFiltersArr[i].Status) {
						notify = false
					}
				}
				if len(n.notificationFiltersArr[i].Branches) > 0 {
					if !stringInSlice(buildParams.BRANCH_NAME, n.notificationFiltersArr[i].Branches) {
						notify = false
					}
				}
				if len(n.notificationFiltersArr[i].Sources) > 0 {
					for sc := 0; sc < len(n.notificationFiltersArr[i].Sources); sc++ {
						if !strings.Contains(buildParams.REPO_NAME, n.notificationFiltersArr[i].Sources[sc]) {
							notify = false
						}
					}
				}
			} else {
				notify = false
				// Default "or" operator case where if any condition is passed, notifications will be sent to Slack
				if len(n.notificationFiltersArr[i].Status) > 0 {
					if stringInSlice(cloudbuildResponse.Status, n.notificationFiltersArr[i].Status) {
						notify = true
					}
				}
				if len(n.notificationFiltersArr[i].Branches) > 0 {
					if stringInSlice(buildParams.BRANCH_NAME, n.notificationFiltersArr[i].Branches) {
						notify = true
					}
				}
				if len(n.notificationFiltersArr[i].Sources) > 0 {
					for sc := 0; sc < len(n.notificationFiltersArr[i].Sources); sc++ {
						if strings.Contains(buildParams.REPO_NAME, n.notificationFiltersArr[i].Sources[sc]) {
							notify = true
						}
					}
				}
			}
			if notify {
				break
			}
		}
		if !notify {
			return nil
		}

	}

	var color string
	switch {
	case stringInSlice(cloudbuildResponse.Status, []string{"FAILURE", "INTERNAL_ERROR", "TIMEOUT"}):
		color = "danger"
	case cloudbuildResponse.Status == "CANCELLED":
		color = "#C0C0C0"
	case cloudbuildResponse.Status == "SUCCESS":
		color = "good"
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
