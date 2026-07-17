package webhook

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/grpcclient"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/order"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/payment"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
)

var ErrPaymentNotFound = errors.New("payment not found for this session")

// EmailEnqueuer mirrors order.EmailEnqueuer's method shape. Declared locally
// (not imported from order) to keep this module's dependency surface small —
// task.Enqueuer already satisfies it structurally.
type EmailEnqueuer interface {
	EnqueueSendEmail(ctx context.Context, in grpcclient.EmailInput) error
}

// Service applies Xendit payment_session webhook events to our own tables.
// This is the only place that both writes payments (payment module's table)
// and flips an order to paid (order module's table) — everywhere else those
// stay in their own module, per the ownership notes in payment_repository.go.
type Service interface {
	HandlePaymentSession(ctx context.Context, sessionID, status string) error
}

type service struct {
	paymentRepo payment.Repository
	orderRepo   order.Repository
	emailQueue  EmailEnqueuer
}

// NewService: emailQueue boleh nil (mis. di test) — kalau nil, notifikasi
// dilewati tanpa error, sama seperti pola di order.Service.
func NewService(paymentRepo payment.Repository, orderRepo order.Repository, emailQueue EmailEnqueuer) Service {
	return &service{paymentRepo: paymentRepo, orderRepo: orderRepo, emailQueue: emailQueue}
}

// HandlePaymentSession applies one payment_session.* event. Statuses we don't
// act on (ACTIVE, CANCELED, anything unrecognized) are a silent no-op, not an
// error — Xendit only needs a 2xx to stop retrying.
func (s *service) HandlePaymentSession(ctx context.Context, sessionID, status string) error {
	p, err := s.paymentRepo.FindPaymentByReference(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}

	var newStatus string
	switch status {
	case "COMPLETED":
		newStatus = models.PaymentStatusPaid
	case "EXPIRED":
		newStatus = models.PaymentStatusExpired
	default:
		return nil
	}
	if newStatus == p.Status {
		// Already applied — Xendit retries the same event until it gets 2xx.
		return nil
	}

	p.Status = newStatus
	if err := s.paymentRepo.UpdatePaymentStatus(ctx, p); err != nil {
		return err
	}
	if newStatus != models.PaymentStatusPaid {
		return nil
	}

	o, err := s.paymentRepo.FindOrderByID(ctx, p.OrderID)
	if err != nil {
		logger.Log.Warn().Err(err).Int("order_id", p.OrderID).Msg("xendit webhook: order not found for paid payment")
		return nil
	}
	o.Status = models.OrderStatusPaid
	if err := s.orderRepo.UpdateStatus(ctx, o); err != nil {
		logger.Log.Warn().Err(err).Int("order_id", o.ID).Msg("xendit webhook: gagal update order status jadi paid")
	}

	s.notifyPaid(ctx, p, o)
	return nil
}

// notifyPaid is fire-and-forget: a failed lookup/enqueue only gets logged,
// same pattern as order.service.notifyStatusChange — the payment/order write
// already succeeded and matters more than the email.
func (s *service) notifyPaid(ctx context.Context, p *models.Payment, o *models.Order) {
	if s.emailQueue == nil {
		return
	}
	u, err := s.paymentRepo.FindUserByID(ctx, o.UserID)
	if err != nil {
		logger.Log.Warn().Err(err).Int("order_id", o.ID).Msg("xendit webhook: gagal ambil user buat notifikasi email")
		return
	}

	subject := fmt.Sprintf("Pembayaran Berhasil - Pesanan #%d", o.ID)
	html := fmt.Sprintf(
		"<h2>Pembayaran Berhasil</h2><p>Halo %s, pembayaran untuk pesanan <b>#%d</b> sebesar Rp%.0f telah kami terima.</p>",
		u.FullName, o.ID, p.Amount,
	)
	text := fmt.Sprintf("Halo %s, pembayaran untuk pesanan #%d sebesar Rp%.0f telah kami terima.", u.FullName, o.ID, p.Amount)

	if err := s.emailQueue.EnqueueSendEmail(ctx, grpcclient.EmailInput{
		ToEmail:     u.Email,
		ToName:      u.FullName,
		Subject:     subject,
		HTMLContent: html,
		TextContent: text,
	}); err != nil {
		logger.Log.Warn().Err(err).Int("order_id", o.ID).Msg("xendit webhook: gagal enqueue email konfirmasi pembayaran")
	}
}
