# cloudbuild-notifer 
Cloudbuild Notifier is a simple go service that listens to `cloud-build` Google Cloud Pub/Sub topic, which exists by default.

### Docker image
cloudbuild-notifier is publicly accessible as a Docker image:

```
gcr.io/cloudkite-public/cloudbuild-notifier:latest
```

### How cloudbuild-notifier works
Cloudbuild Notifier filters out messages from Cloud Pub/Sub messages and send notifications to Slack if there failing builds.

A subscription to that topic is created automatically to receive build status messages and if builds have any fails (FAILURE, INTERNAL_ERROR, TIMEOUT, CANCELLED), a notification is sent to Slack and/or Email.
One can apply filters to determine when notifications should be sent based on: Build Status, Source Branch or Source Repo. 

#### Slack Setup

Follow instructions here https://cloud.google.com/cloud-build/docs/configure-third-party-notifications#slack_notifications

### Configuration

#### Environment variables

| NAME                                  | DEFAULT                                | DESCRIPTION                                                                                          |
| ------------------------------------- | -------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| GCLOUD_PROJECT_ID                     | no default                             | GCP project id                                                                                       |
| SLACK_WEBHOOK_URL                     | no default                             | Slack Webhook URL. Read more https://api.slack.com/incoming-webhooks                                 |
| GCLOUD_PUBSUB_SUBSCRIPTION_NAME       | cloudbuild-notifier-subscription       | Google Cloud Pub/Sub topic subscription. Read more: https://cloud.google.com/pubsub/docs/subscriber  |
| NOTIFICATION_FILTERS                     | ""                             | A string of regex filters that should trigger notifications to be sent to Slack (see section below) By default, notifications will be sent for all build messages |

##### NOTIFICATION_FILTERS (Optional)
The NOTIFICATION_FILTERS environment variable should be passed as a comma-separated string of regexes in the format:
```javascript
<source regex>:<branch regex>:<status regex>,<source regex>:<branch regex>:<status regex>:<tag regex>
e.g.
org-name/repo-name-(1|2):(main|production|dev-*):(SUCCESS|WORKING),org-name/repo-name-(3|4):dev:(FAILURE|QUEUED):monitoring-*
```