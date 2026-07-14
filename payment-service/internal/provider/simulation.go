package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Simulation adalah PaymentProvider tanpa Xendit, buat dev & demo.
// Dia mimik ingatan Xendit pakai map in-memory. Kalau service restart, data
// simulasi ilang, dan itu gak masalah: sumber kebenaran tetap di core.
type Simulation struct {
	baseURL string // base URL payment-service, buat nyusun link /simulation/pay

	mu    sync.RWMutex
	store map[string]simRecord
}

type simRecord struct {
	invoice Invoice
	orderID int64
}

// invoiceTTL: umur invoice simulasi sebelum dianggap kadaluarsa.
const invoiceTTL = 24 * time.Hour

// NewSimulation bikin provider simulasi. baseURL diambil dari config
// (mis. http://localhost:8081), dipakai buat nyusun payment_link.
func NewSimulation(baseURL string) *Simulation {
	return &Simulation{
		baseURL: baseURL,
		store:   make(map[string]simRecord),
	}
}

// pastikan Simulation memenuhi kontrak PaymentProvider di compile time.
var _ PaymentProvider = (*Simulation)(nil)

func (s *Simulation) CreateInvoice(ctx context.Context, p CreateInvoiceParams) (Invoice, error) {
	// Provider tetap dumb: tiap panggilan = invoice baru. Idempotency urusan core.
	// (Xendit asli bakal pakai p.IdempotencyKey buat nolak duplikat di sini.)
	ref := newReference()
	inv := Invoice{
		Reference:   ref,
		PaymentLink: fmt.Sprintf("%s/simulation/pay?ref=%s", s.baseURL, ref),
		Status:      StatusPending,
		ExpiresAt:   time.Now().Add(invoiceTTL),
		OrderID:     p.OrderID, // <- tambahan
	}

	s.mu.Lock()
	s.store[ref] = simRecord{invoice: inv, orderID: p.OrderID}
	s.mu.Unlock()

	return inv, nil
}

func (s *Simulation) GetInvoice(ctx context.Context, reference string) (Invoice, error) {
	s.mu.RLock()
	rec, ok := s.store[reference]
	s.mu.RUnlock()
	if !ok {
		return Invoice{}, ErrInvoiceNotFound
	}
	return rec.invoice, nil
}

// ParseWebhook: provider simulasi gak nerima webhook eksternal. Jalur "paid"-nya
// lewat MarkPaid (dipicu endpoint /simulation/pay), bukan lewat sini.
func (s *Simulation) ParseWebhook(payload []byte, signature string) (WebhookEvent, error) {
	return WebhookEvent{}, errors.New("provider: simulasi tidak menerima webhook eksternal, pakai MarkPaid via /simulation/pay")
}

// MarkPaid nandain invoice jadi paid. DI LUAR interface PaymentProvider,
// khusus simulasi, dipanggil lewat RPC SimulatePayment pas demo.
// Idempotent: kalau udah paid, tetap balikin invoice yang sama tanpa error.
func (s *Simulation) MarkPaid(reference string) (Invoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.store[reference]
	if !ok {
		return Invoice{}, ErrInvoiceNotFound
	}

	rec.invoice.Status = StatusPaid
	s.store[reference] = rec

	return rec.invoice, nil
}

func newReference() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b) // crypto/rand
	return "sim-inv-" + hex.EncodeToString(b)
}
