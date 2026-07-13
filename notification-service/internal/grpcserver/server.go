package grpcserver

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/Djarottosca/finalproject-ftgo-1/notification-service/internal/provider"
	notificationv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/notification/v1"
)

type Server struct {
	notificationv1.UnimplementedNotificationServiceServer
	sender provider.EmailSender
	log    zerolog.Logger
}

func NewServer(sender provider.EmailSender, log zerolog.Logger) *Server {
	return &Server{sender: sender, log: log}
}

func (s *Server) SendEmail(ctx context.Context, req *notificationv1.SendEmailRequest) (*notificationv1.SendEmailResponse, error) {
	if req.GetToEmail() == "" {
		return &notificationv1.SendEmailResponse{
			Success: false,
			Message: "to_email tidak boleh kosong",
		}, nil
	}
	if req.GetSubject() == "" {
		return &notificationv1.SendEmailResponse{
			Success: false,
			Message: "subject tidak boleh kosong",
		}, nil
	}

	result, err := s.sender.SendEmail(ctx, provider.SendEmailInput{
		ToEmail:     req.GetToEmail(),
		ToName:      req.GetToName(),
		Subject:     req.GetSubject(),
		HTMLContent: req.GetHtmlContent(),
		TextContent: req.GetTextContent(),
	})
	if err != nil {
		s.log.Error().Err(err).Str("to_email", req.GetToEmail()).Msg("gagal kirim email")
		return &notificationv1.SendEmailResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	s.log.Info().Str("to_email", req.GetToEmail()).Str("message_id", result.MessageID).Msg("email terkirim")

	return &notificationv1.SendEmailResponse{
		Success:   true,
		Message:   "email berhasil dikirim",
		MessageId: result.MessageID,
	}, nil
}
