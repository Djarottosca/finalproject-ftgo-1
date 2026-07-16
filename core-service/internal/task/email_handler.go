package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/grpcclient"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
)

// EmailHandler processes TypeSendEmail jobs by calling notification-service
// over gRPC. Registered against an *asynq.ServeMux in cmd/worker.
type EmailHandler struct {
	notifClient *grpcclient.NotificationClient
}

func NewEmailHandler(notifClient *grpcclient.NotificationClient) *EmailHandler {
	return &EmailHandler{notifClient: notifClient}
}

// ProcessTask implements asynq.Handler.
func (h *EmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		// malformed payload can never succeed on retry — skip retrying.
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	_, err := h.notifClient.SendEmail(ctx, grpcclient.EmailInput{
		ToEmail:     p.ToEmail,
		ToName:      p.ToName,
		Subject:     p.Subject,
		HTMLContent: p.HTMLContent,
		TextContent: p.TextContent,
	})
	if err != nil {
		logger.Log.Warn().Err(err).Str("to_email", p.ToEmail).Msg("gagal kirim email dari worker, akan diretry asynq")
		return err
	}

	logger.Log.Info().Str("to_email", p.ToEmail).Msg("email terkirim via worker")
	return nil
}
