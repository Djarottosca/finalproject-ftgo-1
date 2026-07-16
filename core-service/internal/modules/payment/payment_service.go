package payment

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/grpcclient"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrForbidden       = errors.New("order does not belong to this user")
	ErrOrderNotPayable = errors.New("order is not in a payable state")
	ErrNotFound        = errors.New("payment not found")
)

// paymentGateway is the slice of grpcclient.PaymentClient this module needs.
// Declaring it as an interface (instead of taking the concrete client) lets
// tests substitute a fake without a live payment-service.
type paymentGateway interface {
	CreatePayment(ctx context.Context, in grpcclient.CreatePaymentInput) (grpcclient.Payment, error)
	GetPaymentStatus(ctx context.Context, reference string) (grpcclient.Payment, error)
}

// Service defines the payment use cases exposed to the handler layer.
type Service interface {
	CreateForOrder(ctx context.Context, orderID, userID int) (*PaymentResponse, error)
	GetStatus(ctx context.Context, orderID, userID int) (*PaymentResponse, error)
}

type service struct {
	repo    Repository
	gateway paymentGateway
}

// NewService wires the repository and the payment-service gateway (grpcclient).
func NewService(repo Repository, gateway paymentGateway) Service {
	return &service{repo: repo, gateway: gateway}
}

// CreateForOrder starts a payment for an already-created order. It's idempotent:
// calling it twice for the same order returns the existing payment instead of
// creating a duplicate (the payments table also enforces one payment per order).
func (s *service) CreateForOrder(ctx context.Context, orderID, userID int) (*PaymentResponse, error) {
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// A user may only pay for their own order.
	if order.UserID != userID {
		return nil, ErrForbidden
	}

	// Only a pending order can start a payment; paid/cancelled can't.
	if order.Status != models.OrderStatusPending {
		return nil, ErrOrderNotPayable
	}

	// Idempotency: if a payment already exists for this order, return it.
	if existing, err := s.repo.FindPaymentByOrderID(ctx, orderID); err == nil {
		return toResponse(existing), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Customer name/email come from the user; Xendit shows them on checkout.
	user, err := s.repo.FindUserByID(ctx, order.UserID)
	if err != nil {
		return nil, err
	}

	// Amount is authoritative from the order. FinalPrice is float64 rupiah;
	// the gateway takes integer rupiah, and IDR has no sub-rupiah, so the
	// int64 conversion is exact for real amounts.
	result, err := s.gateway.CreatePayment(ctx, grpcclient.CreatePaymentInput{
		OrderID:       int64(order.ID),
		UserID:        int64(order.UserID),
		CustomerName:  user.FullName,
		CustomerEmail: user.Email,
		Amount:        int64(order.FinalPrice),
		Description:   fmt.Sprintf("Pembayaran order #%d", order.ID),
	})
	if err != nil {
		return nil, err
	}

	link := result.PaymentLink
	ref := result.Reference
	payment := &models.Payment{
		OrderID:          order.ID,
		Amount:           order.FinalPrice,
		Tax:              0,
		Status:           toModelStatus(result.Status),
		PaymentLink:      &link,
		PaymentReference: &ref,
	}
	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	return toResponse(payment), nil
}

// GetStatus returns the stored payment and, if still pending, refreshes it from
// payment-service. This is the manual version of what the Asynq worker will do
// automatically: since there's no webhook, asking is the only way to learn the
// user paid.
//
// NOTE: this deliberately does NOT flip the order to "paid" — orders is another
// module's table to write. That transition belongs to the order module / Asynq
// worker and needs coordination, not a silent cross-module write here.
func (s *service) GetStatus(ctx context.Context, orderID, userID int) (*PaymentResponse, error) {
	payment, err := s.repo.FindPaymentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Ownership: confirm the order behind this payment belongs to the caller.
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrForbidden
	}

	// Only ask the gateway if we might still learn something new.
	if payment.PaymentReference != nil && payment.Status == models.PaymentStatusPending {
		latest, err := s.gateway.GetPaymentStatus(ctx, *payment.PaymentReference)
		if err == nil {
			newStatus := toModelStatus(latest.Status)
			if newStatus != payment.Status {
				payment.Status = newStatus
				// Best-effort persist; returning the fresh status matters more
				// than a write error here.
				_ = s.repo.UpdatePaymentStatus(ctx, payment)
			}
		}
	}

	return toResponse(payment), nil
}

// toModelStatus maps the gateway's status vocabulary (UPPERCASE, core-side) to
// the lowercase values the payments table's CHECK constraint allows.
func toModelStatus(s grpcclient.PaymentStatus) string {
	switch s {
	case grpcclient.PaymentStatusPaid:
		return models.PaymentStatusPaid
	case grpcclient.PaymentStatusExpired:
		return models.PaymentStatusExpired
	case grpcclient.PaymentStatusFailed:
		return models.PaymentStatusFailed
	default:
		// PENDING and UNKNOWN both map to pending: safer to treat an
		// unrecognized status as "not yet paid" than to assume paid.
		return models.PaymentStatusPending
	}
}
