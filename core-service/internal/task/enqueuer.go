package task

import (
	"context"

	"github.com/hibiken/asynq"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/grpcclient"
)

// Enqueuer wraps an *asynq.Client so callers (e.g. order module) only see a
// narrow "send this email eventually" interface — they don't need to know
// about asynq.Task/queue names.
type Enqueuer struct {
	client *asynq.Client
}

func NewEnqueuer(client *asynq.Client) *Enqueuer {
	return &Enqueuer{client: client}
}

// EnqueueSendEmail queues an email job instead of calling notification-service
// synchronously, so a slow/unavailable mail provider never blocks the HTTP
// request that triggered it (e.g. order checkout, status update).
func (e *Enqueuer) EnqueueSendEmail(ctx context.Context, in grpcclient.EmailInput) error {
	t, err := NewSendEmailTask(SendEmailPayload{
		ToEmail:     in.ToEmail,
		ToName:      in.ToName,
		Subject:     in.Subject,
		HTMLContent: in.HTMLContent,
		TextContent: in.TextContent,
	})
	if err != nil {
		return err
	}
	_, err = e.client.EnqueueContext(ctx, t)
	return err
}
