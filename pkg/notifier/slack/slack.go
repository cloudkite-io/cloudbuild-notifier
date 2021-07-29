package slack

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"

	cloudbuildnotifier "github.com/cloudkite-io/cloudbuild-notifier"
	"github.com/cloudkite-io/cloudbuild-notifier/pkg/cloudbuild"
	"github.com/slack-go/slack"
)

type notifier struct {
	webhookURL    string
	githubUserURL string
}

// New creates a slack notifier.
func New(webhookURL, githubUserURL string) cloudbuildnotifier.Notifier {
	return notifier{
		webhookURL,
		githubUserURL,
	}
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

	if cloudbuildResponse.Status == "QUEUED" || cloudbuildResponse.Status == "WORKING" {
		return nil
	}

	commitShaURL, err := buildGitSourceURL(buildParams)
	if err != nil {
		return fmt.Errorf("failed posting to webhook %s: %s", n.webhookURL, err)
	}

	attachment := slack.Attachment{
		Title: fmt.Sprintf("Cloudbuild: %s", cloudbuildResponse.Status),
		Color: color,
		Text: fmt.Sprintf("Repo: %s\nBranch: %s\nCommit SHA: %s\nTrigger: %s\n",
			buildParams.REPO_NAME, buildParams.BRANCH_NAME, commitShaURL, buildParams.TRIGGER_NAME),
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

	err = slack.PostWebhook(n.webhookURL, &msg)
	if err != nil {
		return fmt.Errorf("failed posting to webhook %s: %s", n.webhookURL, err)
	}

	log.Printf("Sent %s Slack message for build %s\n", n.webhookURL, buildParams.Id)

	return nil
}

func buildGitSourceURL(buildParams cloudbuild.BuildParameters) (string, error) {
	u, err := url.Parse(buildParams.URL)
	if err != nil {
		return "", err
	}
	u.Path = strings.Trim(u.Path, ".git")
	u.Path = path.Join(u.Path, "commit", buildParams.COMMIT_SHA)
	return u.String(), nil
}

func stringInSlice(needle string, haystack []string) bool {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}
	return false
}
