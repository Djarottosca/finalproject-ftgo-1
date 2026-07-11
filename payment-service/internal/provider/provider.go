package provider

import (
	"context"
	"errors"
	"time"
)

// Status internal provider, gak kenal proto.
type Status int

var ErrInvoiceNotFound = errors.New("provider: invoice not found")

const (
	StatusPending Status = iota
	StatusPaid
	StatusExpired
	StatusFailed
)

type CreateInvoiceParams struct {
	OrderID        int64
	UserID         int64
	CustomerName   string
	CustomerEmail  string
	Amount         int64 // rupiah
	Description    string
	Items          []InvoiceItem // opsional, display-only
	IdempotencyKey string        // opsional
}

type InvoiceItem struct {
	Name     string
	Quantity int32
	Price    int64
}

type Invoice struct {
	Reference   string
	PaymentLink string
	Status      Status
	ExpiresAt   time.Time
	OrderID     int64 // <- tambahan, biar GetPaymentStatus bisa balikin order_id
}

type WebhookEvent struct {
	Reference string
	OrderID   int64
	Status    Status
}

// Satu-satunya kontak ke dunia pembayaran (Xendit / simulasi).
type PaymentProvider interface {
	CreateInvoice(ctx context.Context, p CreateInvoiceParams) (Invoice, error)
	GetInvoice(ctx context.Context, reference string) (Invoice, error)
	ParseWebhook(payload []byte, signature string) (WebhookEvent, error)
}
