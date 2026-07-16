package grpcclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	notificationv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/notification/v1"
)

// EmailInput: yang core kirim tiap kali mau notifikasi user lewat email.
type EmailInput struct {
	ToEmail     string
	ToName      string
	Subject     string
	HTMLContent string
	TextContent string // opsional
}

// EmailResult: hasil yang core terima balik dari notification-service.
type EmailResult struct {
	Success   bool
	Message   string
	MessageID string
}

// NotificationClient: anti-corruption layer di sisi core, sama polanya
// dengan PaymentClient. Nelen semua tipe proto, yang keluar cuma struct Go biasa.
type NotificationClient struct {
	client notificationv1.NotificationServiceClient
}

// NewNotificationClient: conn dioper dari luar (dibikin sekali di bootstrap),
// jangan bikin koneksi baru tiap request.
func NewNotificationClient(conn *grpc.ClientConn) *NotificationClient {
	return &NotificationClient{client: notificationv1.NewNotificationServiceClient(conn)}
}

// SendEmail motong komunikasi gRPC ke notification-service
func (c *NotificationClient) SendEmail(ctx context.Context, in EmailInput) (EmailResult, error) {
	req := &notificationv1.SendEmailRequest{
		ToEmail:     in.ToEmail,
		ToName:      in.ToName,
		Subject:     in.Subject,
		HtmlContent: in.HTMLContent,
		TextContent: in.TextContent,
	}

	resp, err := c.client.SendEmail(ctx, req)
	if err != nil {
		return EmailResult{}, fmt.Errorf("grpcclient: gagal kirim email: %w", err)
	}

	return EmailResult{
		Success:   resp.GetSuccess(),
		Message:   resp.GetMessage(),
		MessageID: resp.GetMessageId(),
	}, nil
}
