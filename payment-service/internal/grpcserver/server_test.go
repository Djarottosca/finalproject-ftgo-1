package grpcserver

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/provider"
	paymentv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/payment/v1"
)

// fakeProvider: implementasi palsu PaymentProvider, buat nguji grpcserver
// tanpa nyentuh simulasi maupun Xendit.
type fakeProvider struct {
	invoice    provider.Invoice
	err        error
	lastParams provider.CreateInvoiceParams // buat mastiin mapping proto->provider bener
}

func (f *fakeProvider) CreateInvoice(ctx context.Context, p provider.CreateInvoiceParams) (provider.Invoice, error) {
	f.lastParams = p
	if f.err != nil {
		return provider.Invoice{}, f.err
	}
	return f.invoice, nil
}

func (f *fakeProvider) GetInvoice(ctx context.Context, reference string) (provider.Invoice, error) {
	if f.err != nil {
		return provider.Invoice{}, f.err
	}
	return f.invoice, nil
}

func discardLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

func TestCreatePayment_MappingBenar(t *testing.T) {
	expiry := time.Now().Add(30 * time.Minute)
	fake := &fakeProvider{
		invoice: provider.Invoice{
			Reference:   "ps-abc",
			PaymentLink: "https://xen.to/abc",
			Status:      provider.StatusPending,
			ExpiresAt:   expiry,
			OrderID:     123,
		},
	}
	srv := NewPaymentServer(fake, discardLogger())

	resp, err := srv.CreatePayment(context.Background(), &paymentv1.CreatePaymentRequest{
		OrderId:       123,
		UserId:        1,
		CustomerName:  "Kevin",
		CustomerEmail: "kevin@example.com",
		Amount:        250000,
		Items: []*paymentv1.PaymentItem{
			{Name: "Semen", Quantity: 5, Price: 50000},
		},
	})
	if err != nil {
		t.Fatalf("CreatePayment error: %v", err)
	}

	// proto -> provider: params yang diterima provider harus sesuai request
	if fake.lastParams.OrderID != 123 || fake.lastParams.Amount != 250000 {
		t.Errorf("params ke provider salah: %+v", fake.lastParams)
	}
	if len(fake.lastParams.Items) != 1 || fake.lastParams.Items[0].Name != "Semen" {
		t.Errorf("items gak kemapping bener: %+v", fake.lastParams.Items)
	}

	// provider -> proto: response harus bawa data dari invoice
	if resp.GetPaymentReference() != "ps-abc" {
		t.Errorf("reference salah: %s", resp.GetPaymentReference())
	}
	if resp.GetPaymentLink() != "https://xen.to/abc" {
		t.Errorf("payment link salah: %s", resp.GetPaymentLink())
	}
	if resp.GetStatus() != paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING {
		t.Errorf("status salah: %v", resp.GetStatus())
	}
	if !resp.GetExpiresAt().AsTime().Equal(expiry.UTC()) {
		t.Errorf("expires_at gak kemapping bener")
	}
}

func TestCreatePayment_ProviderError(t *testing.T) {
	fake := &fakeProvider{err: errors.New("xendit meledak")}
	srv := NewPaymentServer(fake, discardLogger())

	_, err := srv.CreatePayment(context.Background(), &paymentv1.CreatePaymentRequest{OrderId: 1})
	if status.Code(err) != codes.Internal {
		t.Errorf("harusnya codes.Internal, dapet %v", status.Code(err))
	}
	// detail error provider TIDAK boleh bocor ke pemanggil
	if err != nil && err.Error() == "xendit meledak" {
		t.Error("detail error provider bocor ke core")
	}
}

func TestGetPaymentStatus_NotFound(t *testing.T) {
	fake := &fakeProvider{err: provider.ErrInvoiceNotFound}
	srv := NewPaymentServer(fake, discardLogger())

	_, err := srv.GetPaymentStatus(context.Background(), &paymentv1.GetPaymentStatusRequest{
		PaymentReference: "gak-ada",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("ErrInvoiceNotFound harusnya jadi codes.NotFound, dapet %v", status.Code(err))
	}
}

// SimulatePayment harus DITOLAK kalau provider aktif bukan simulasi.
// Ini penjaga produksi: jalur demo mati sendiri kalau pakai Xendit.
func TestSimulatePayment_DitolakKalauBukanSimulasi(t *testing.T) {
	srv := NewPaymentServer(&fakeProvider{}, discardLogger())

	_, err := srv.SimulatePayment(context.Background(), &paymentv1.SimulatePaymentRequest{
		PaymentReference: "apapun",
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Errorf("harusnya codes.FailedPrecondition, dapet %v", status.Code(err))
	}
}

// SimulatePayment jalan kalau provider aktif memang simulasi.
func TestSimulatePayment_JalanKalauSimulasi(t *testing.T) {
	sim := provider.NewSimulation("http://localhost:9001", 24*time.Hour)
	srv := NewPaymentServer(sim, discardLogger())
	ctx := context.Background()

	created, err := srv.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{
		OrderId: 123,
		Amount:  250000,
	})
	if err != nil {
		t.Fatalf("CreatePayment error: %v", err)
	}

	paid, err := srv.SimulatePayment(ctx, &paymentv1.SimulatePaymentRequest{
		PaymentReference: created.GetPaymentReference(),
	})
	if err != nil {
		t.Fatalf("SimulatePayment error: %v", err)
	}
	if paid.GetStatus() != paymentv1.PaymentStatus_PAYMENT_STATUS_PAID {
		t.Errorf("harusnya PAID, dapet %v", paid.GetStatus())
	}
	if paid.GetOrderId() != 123 {
		t.Errorf("order_id harusnya 123, dapet %d", paid.GetOrderId())
	}
}

func TestStatusToProto(t *testing.T) {
	cases := []struct {
		in   provider.Status
		want paymentv1.PaymentStatus
	}{
		{provider.StatusPending, paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING},
		{provider.StatusPaid, paymentv1.PaymentStatus_PAYMENT_STATUS_PAID},
		{provider.StatusExpired, paymentv1.PaymentStatus_PAYMENT_STATUS_EXPIRED},
		{provider.StatusFailed, paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED},
	}
	for _, c := range cases {
		if got := statusToProto(c.in); got != c.want {
			t.Errorf("statusToProto(%v) = %v, mau %v", c.in, got, c.want)
		}
	}
}
