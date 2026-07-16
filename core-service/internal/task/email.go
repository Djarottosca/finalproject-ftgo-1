// Package task defines Asynq task types enqueued by core-service's HTTP
// server and processed by its worker (cmd/worker), decoupling slow email
// sends from the request path.
package task

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// TypeSendEmail is the Asynq task type for outbound email jobs.
const TypeSendEmail = "email:send"

// SendEmailPayload mirrors grpcclient.EmailInput — kept as its own type so
// this package doesn't depend on grpcclient (avoids an import cycle since
// grpcclient is also used directly by modules).
type SendEmailPayload struct {
	ToEmail     string `json:"to_email"`
	ToName      string `json:"to_name"`
	Subject     string `json:"subject"`
	HTMLContent string `json:"html_content"`
	TextContent string `json:"text_content"`
}

// NewSendEmailTask builds an *asynq.Task ready to enqueue.
func NewSendEmailTask(p SendEmailPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendEmail, payload), nil
}
