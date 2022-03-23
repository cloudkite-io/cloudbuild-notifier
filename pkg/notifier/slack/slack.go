package slack

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/cloudbuild"
	"github.com/slack-go/slack"
)

type notifier struct {
	webhookURL          string
	notificationFilters string
}

// New creates a slack notifier.
func New(webhookURL string, notificationFilters string) cloudbuildnotifier.Notifier {
	return notifier{webhookURL, notificationFilters}
}

func (n notifier) Send(cloudbuildResponse cloudbuildnotifier.CloudbuildResponse, buildParams cloudbuild.BuildParameters) error {

	// Check if there are any notification filters that have been passed
	// Notification filters are used to determine whether the build status should be sent to Slack channel
	// <source regex>:<branch regex>:<status regex>,<source regex>:<branch regex>:<status regex>
	if n.notificationFilters != "" {
		notify := false
		notificationFiltersArray := strings.Split(n.notificationFilters, ",")
		for i := 0; i < len(notificationFiltersArray); i++ {
			splitNotificationFilters := strings.Split(notificationFiltersArray[i], ":")
			sourceRegex, _ := regexp.Compile(splitNotificationFilters[0])
			branchRegex, _ := regexp.Compile(splitNotificationFilters[1])
			statusRegex, _ := regexp.Compile(splitNotificationFilters[2])
			// check if all filters have been met
			if sourceRegex.MatchString(cloudbuildResponse.Source) && branchRegex.MatchString(buildParams.BRANCH_NAME) && statusRegex.MatchString(cloudbuildResponse.Status) {
				notify = true
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
