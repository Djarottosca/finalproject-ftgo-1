package grpcclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	paymentv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/payment/v1"
)

// PaymentStatus: status pembayaran dalam bahasa core, bukan proto.
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusPaid    PaymentStatus = "PAID"
	PaymentStatusExpired PaymentStatus = "EXPIRED"
	PaymentStatusFailed  PaymentStatus = "FAILED"
	PaymentStatusUnknown PaymentStatus = "UNKNOWN"
)

// CreatePaymentInput: yang core kirim pas checkout.
type CreatePaymentInput struct {
	OrderID        int64
	UserID         int64
	CustomerName   string
	CustomerEmail  string
	Amount         int64 // rupiah utuh, integer
	Description    string
	Items          []PaymentItem // opsional, cuma buat tampilan invoice
	IdempotencyKey string        // opsional
}

type PaymentItem struct {
	Name     string
	Quantity int32
	Price    int64
}

// Payment: hasil yang core terima. Gak ada tipe proto di sini.
type Payment struct {
	Reference   string // simpan ini di tabel payments (payment_reference)
	PaymentLink string // redirect user ke sini
	Status      PaymentStatus
	ExpiresAt   time.Time // session Xendit umurnya pendek (default 30 menit)
	OrderID     int64
}

// PaymentClient: anti-corruption layer di sisi core.
// Nelen semua kerumitan proto, yang keluar cuma struct Go biasa.
type PaymentClient struct {
	client paymentv1.PaymentServiceClient
}

// NewPaymentClient: conn dioper dari luar (dibikin sekali di main core),
// jangan bikin koneksi baru tiap request.
func NewPaymentClient(conn *grpc.ClientConn) *PaymentClient {
	return &PaymentClient{client: paymentv1.NewPaymentServiceClient(conn)}
}

// CreatePayment dipanggil pas checkout. Balikin link bayar buat di-redirect ke user.
func (c *PaymentClient) CreatePayment(ctx context.Context, in CreatePaymentInput) (Payment, error) {
	req := &paymentv1.CreatePaymentRequest{
		OrderId:        in.OrderID,
		UserId:         in.UserID,
		CustomerName:   in.CustomerName,
		CustomerEmail:  in.CustomerEmail,
		Amount:         in.Amount,
		Description:    in.Description,
		IdempotencyKey: in.IdempotencyKey,
		Items:          itemsToProto(in.Items),
	}

	resp, err := c.client.CreatePayment(ctx, req)
	if err != nil {
		return Payment{}, fmt.Errorf("grpcclient: gagal bikin pembayaran: %w", err)
	}

	return Payment{
		Reference:   resp.GetPaymentReference(),
		PaymentLink: resp.GetPaymentLink(),
		Status:      statusFromProto(resp.GetStatus()),
		ExpiresAt:   resp.GetExpiresAt().AsTime(),
		OrderID:     in.OrderID,
	}, nil
}

// GetPaymentStatus di-poll job Asynq sampai status berubah dari PENDING.
// Ini SATU-SATUNYA cara core tau user udah bayar (gak ada webhook).
func (c *PaymentClient) GetPaymentStatus(ctx context.Context, reference string) (Payment, error) {
	resp, err := c.client.GetPaymentStatus(ctx, &paymentv1.GetPaymentStatusRequest{
		PaymentReference: reference,
	})
	if err != nil {
		return Payment{}, fmt.Errorf("grpcclient: gagal ambil status pembayaran: %w", err)
	}

	return Payment{
		Reference: resp.GetPaymentReference(),
		Status:    statusFromProto(resp.GetStatus()),
		OrderID:   resp.GetOrderId(),
	}, nil
}

// SimulatePayment: HANYA buat demo/dev, maksa pembayaran jadi PAID tanpa Xendit.
// Balikin error kalau payment-service lagi pakai provider xendit.
func (c *PaymentClient) SimulatePayment(ctx context.Context, reference string) (Payment, error) {
	resp, err := c.client.SimulatePayment(ctx, &paymentv1.SimulatePaymentRequest{
		PaymentReference: reference,
	})
	if err != nil {
		return Payment{}, fmt.Errorf("grpcclient: gagal simulasi pembayaran: %w", err)
	}

	return Payment{
		Reference: resp.GetPaymentReference(),
		Status:    statusFromProto(resp.GetStatus()),
		OrderID:   resp.GetOrderId(),
	}, nil
}

// --- mapping proto -> core ---

func statusFromProto(s paymentv1.PaymentStatus) PaymentStatus {
	switch s {
	case paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING:
		return PaymentStatusPending
	case paymentv1.PaymentStatus_PAYMENT_STATUS_PAID:
		return PaymentStatusPaid
	case paymentv1.PaymentStatus_PAYMENT_STATUS_EXPIRED:
		return PaymentStatusExpired
	case paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED:
		return PaymentStatusFailed
	default:
		return PaymentStatusUnknown
	}
}

func itemsToProto(items []PaymentItem) []*paymentv1.PaymentItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]*paymentv1.PaymentItem, 0, len(items))
	for _, it := range items {
		out = append(out, &paymentv1.PaymentItem{
			Name:     it.Name,
			Quantity: it.Quantity,
			Price:    it.Price,
		})
	}
	return out
}
