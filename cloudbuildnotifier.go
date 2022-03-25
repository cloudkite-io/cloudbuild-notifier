package cloudbuildnotifier

import (
	"time"

	"github.com/cloudkite-io/cloudbuild-notifier/pkg/cloudbuild"
)

// Notifier sends messages
type Notifier interface {
	Send(cloudbuildResponse CloudbuildResponse, buildParams cloudbuild.BuildParameters) error
}

type CloudbuildResponse struct {
	Status     string    `json:"status"`
	CreateTime time.Time `json:"createTime"`
	LogURL     string    `json:"logUrl"`
}
