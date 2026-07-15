package provider

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestSim() *Simulation {
	return NewSimulation("http://localhost:9001", 24*time.Hour)
}

func TestSimulation_CreateInvoice(t *testing.T) {
	sim := newTestSim()

	inv, err := sim.CreateInvoice(context.Background(), CreateInvoiceParams{
		OrderID: 123,
		Amount:  250000,
	})
	if err != nil {
		t.Fatalf("CreateInvoice error: %v", err)
	}

	if inv.Status != StatusPending {
		t.Errorf("invoice baru harusnya PENDING, dapet %v", inv.Status)
	}
	if inv.OrderID != 123 {
		t.Errorf("OrderID harusnya 123, dapet %d", inv.OrderID)
	}
	if inv.Reference == "" {
		t.Error("Reference gak boleh kosong")
	}
	if inv.PaymentLink == "" {
		t.Error("PaymentLink gak boleh kosong")
	}
}

func TestSimulation_CreateInvoice_ReferenceUnik(t *testing.T) {
	sim := newTestSim()
	ctx := context.Background()

	a, _ := sim.CreateInvoice(ctx, CreateInvoiceParams{OrderID: 1, Amount: 1000})
	b, _ := sim.CreateInvoice(ctx, CreateInvoiceParams{OrderID: 2, Amount: 2000})

	if a.Reference == b.Reference {
		t.Error("dua invoice harus punya reference berbeda")
	}
}

func TestSimulation_GetInvoice_TidakKetemu(t *testing.T) {
	sim := newTestSim()

	_, err := sim.GetInvoice(context.Background(), "gak-ada")
	if !errors.Is(err, ErrInvoiceNotFound) {
		t.Errorf("harusnya ErrInvoiceNotFound, dapet %v", err)
	}
}

func TestSimulation_MarkPaid(t *testing.T) {
	sim := newTestSim()
	ctx := context.Background()

	inv, _ := sim.CreateInvoice(ctx, CreateInvoiceParams{OrderID: 123, Amount: 250000})

	paid, err := sim.MarkPaid(inv.Reference)
	if err != nil {
		t.Fatalf("MarkPaid error: %v", err)
	}
	if paid.Status != StatusPaid {
		t.Errorf("harusnya PAID, dapet %v", paid.Status)
	}
	if paid.OrderID != 123 {
		t.Errorf("OrderID harusnya kebawa, dapet %d", paid.OrderID)
	}

	// perubahan status harus persisten: GetInvoice setelahnya juga PAID
	got, err := sim.GetInvoice(ctx, inv.Reference)
	if err != nil {
		t.Fatalf("GetInvoice error: %v", err)
	}
	if got.Status != StatusPaid {
		t.Errorf("status harusnya tersimpan sebagai PAID, dapet %v", got.Status)
	}
}

func TestSimulation_MarkPaid_Idempotent(t *testing.T) {
	sim := newTestSim()
	ctx := context.Background()

	inv, _ := sim.CreateInvoice(ctx, CreateInvoiceParams{OrderID: 123, Amount: 250000})

	if _, err := sim.MarkPaid(inv.Reference); err != nil {
		t.Fatalf("MarkPaid pertama error: %v", err)
	}

	// bayar lagi: harus tetap sukses & tetap PAID, bukan error
	second, err := sim.MarkPaid(inv.Reference)
	if err != nil {
		t.Fatalf("MarkPaid kedua harusnya idempotent, malah error: %v", err)
	}
	if second.Status != StatusPaid {
		t.Errorf("harusnya tetap PAID, dapet %v", second.Status)
	}
}

func TestSimulation_MarkPaid_TidakKetemu(t *testing.T) {
	sim := newTestSim()

	_, err := sim.MarkPaid("gak-ada")
	if !errors.Is(err, ErrInvoiceNotFound) {
		t.Errorf("harusnya ErrInvoiceNotFound, dapet %v", err)
	}
}
