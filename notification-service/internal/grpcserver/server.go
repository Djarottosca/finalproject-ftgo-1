// Package grpcserver mengimplementasikan notificationv1.NotificationServiceServer
// sesuai kontrak di proto/notification/v1/notification.proto.
//
// CATATAN: setelah `make proto` dijalankan, kode hasil generate protoc akan
// muncul di proto/notification/v1/ (notification.pb.go & notification_grpc.pb.go).
package grpcserver

import (
	"context"

	"github.com/rs/zerolog"

	notificationv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/notification/v1"

	"github.com/Djarottosca/finalproject-ftgo-1/notification-service/internal/provider"
)

// Server mengimplementasikan notificationv1.NotificationServiceServer.
// Sengaja tipis: gak ada business logic di sini, cuma validasi input dan
// delegasi ke provider.EmailSender (Mailjet). Ini bikin server gampang
// di-unit-test tanpa mock library eksternal, cukup mock EmailSender.
type Server struct {
	notificationv1.UnimplementedNotificationServiceServer
	sender provider.EmailSender
	log    zerolog.Logger
}

// NewServer membuat instance grpc server baru.
func NewServer(sender provider.EmailSender, log zerolog.Logger) *Server {
	return &Server{sender: sender, log: log}
}

// SendEmail dipanggil oleh core-service (langsung, atau lewat Asynq worker
// saat order berubah status jadi "paid") untuk mengirim satu email.
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
