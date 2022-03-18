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
| NOTIFICATION_FILTERS                     | []                             | A list of filters that should trigger notifications to be sent to Slack (see section below) By default, notifications will be sent for all build messages |

##### NOTIFICATION_FILTERS (Optional)
The NOTIFICATION_FILTERS environment variable should be passed as a JSON string consisting of a list of objects as in the case below:
```javascript
[
    {
        "status": [
            "SUCCESS",
            "WORKING"
        ],
        "sources": [
            "org-name/repo-name1",
            "org-name/repo-name2"
        ],
        "branches": [
            "main",
            "production"
        ],
        "operator": "or"
    },
    {
        "status": [
            "FAILURE",
            "QUEUED"
        ],
        "sources": [
            "org-name/repo-name3",
            "org-name/repo-name4"
        ],
        "branches": [
            "dev"
        ],
        "operator": "and"
    }
]
```
Below is a description of the four(optional) parameters that it takes
| NAME                                  | DEFAULT                                | DESCRIPTION                                                                                          |
| ------------------------------------- | -------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| status                     | All statuses                             | A list of build status that should trigger notifications to be sent to Slack e.g. ["FAILURE", "INTERNAL_ERROR"]. By default, notifications will be sent for all statuses |
| sources                     | Any source                             | A list of substrings that will be used to match which source Repos should trigger notifications to be sent to Slack e.g. ["org-name/repo-name"] |
| branches                     | All branches                             | A list of branches whose builds should trigger notifications to be sent to Slack e.g. ["main", "production"]. By default, notifications will be sent for all branches |
| operator                     | "or"                             | The logical operation that should be used to trigger notifications. Options are: "and" or "or". When "and" is used, notifications will only be sent when the build matches all the other NOTIFICATION_FILTERS that have been provided. When "or" is used, notifications are triggered when the build matches any of the NOTIFICATION_FILTERS that have been provided.  |