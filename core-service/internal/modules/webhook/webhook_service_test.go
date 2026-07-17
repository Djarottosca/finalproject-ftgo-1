package webhook_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/webhook"
)

// mockPaymentRepo implements payment.Repository in-memory, keyed by
// payment_reference to mirror how the webhook actually looks payments up.
type mockPaymentRepo struct {
	byReference map[string]*models.Payment
	orders      map[int]*models.Order
	users       map[int]*models.User
}

func (m *mockPaymentRepo) FindOrderByID(_ context.Context, id int) (*models.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return o, nil
}
func (m *mockPaymentRepo) FindUserByID(_ context.Context, id int) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}
func (m *mockPaymentRepo) FindPaymentByOrderID(context.Context, int) (*models.Payment, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockPaymentRepo) FindPaymentByReference(_ context.Context, ref string) (*models.Payment, error) {
	p, ok := m.byReference[ref]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return p, nil
}
func (m *mockPaymentRepo) CreatePayment(_ context.Context, p *models.Payment) error {
	m.byReference[*p.PaymentReference] = p
	return nil
}
func (m *mockPaymentRepo) UpdatePaymentStatus(_ context.Context, p *models.Payment) error {
	m.byReference[*p.PaymentReference] = p
	return nil
}

// mockOrderRepo only implements what webhook.Service calls (UpdateStatus);
// every other method is unused here and just needs to satisfy the interface.
type mockOrderRepo struct {
	orders map[int]*models.Order
}

func (m *mockOrderRepo) CreateWithItems(context.Context, *models.Order, []models.OrderItem) error {
	return nil
}
func (m *mockOrderRepo) FindByID(_ context.Context, id int) (*models.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return o, nil
}
func (m *mockOrderRepo) ItemsByOrderID(context.Context, int) ([]models.OrderItem, error) {
	return nil, nil
}
func (m *mockOrderRepo) ListByUserID(context.Context, int) ([]models.Order, error) { return nil, nil }
func (m *mockOrderRepo) ListBySupplierID(context.Context, int) ([]models.Order, error) {
	return nil, nil
}
func (m *mockOrderRepo) UpdateStatus(_ context.Context, o *models.Order) error {
	m.orders[o.ID].Status = o.Status
	return nil
}

func newTestService(payment *models.Payment, order *models.Order) webhook.Service {
	ref := *payment.PaymentReference
	paymentRepo := &mockPaymentRepo{
		byReference: map[string]*models.Payment{ref: payment},
		orders:      map[int]*models.Order{order.ID: order},
		users:       map[int]*models.User{order.UserID: {ID: order.UserID, FullName: "Test User", Email: "test@test.com"}},
	}
	orderRepo := &mockOrderRepo{orders: map[int]*models.Order{order.ID: order}}
	return webhook.NewService(paymentRepo, orderRepo, nil)
}

func TestHandlePaymentSession_Completed_MarksPaymentAndOrderPaid(t *testing.T) {
	ref := "sess-123"
	payment := &models.Payment{OrderID: 1, Status: models.PaymentStatusPending, PaymentReference: &ref, Amount: 50000}
	order := &models.Order{ID: 1, UserID: 10, Status: models.OrderStatusPending}
	svc := newTestService(payment, order)

	if err := svc.HandlePaymentSession(context.Background(), ref, "COMPLETED"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if payment.Status != models.PaymentStatusPaid {
		t.Fatalf("expected payment status paid, got %s", payment.Status)
	}
	if order.Status != models.OrderStatusPaid {
		t.Fatalf("expected order status paid, got %s", order.Status)
	}
}

func TestHandlePaymentSession_Expired_OnlyTouchesPayment(t *testing.T) {
	ref := "sess-456"
	payment := &models.Payment{OrderID: 2, Status: models.PaymentStatusPending, PaymentReference: &ref}
	order := &models.Order{ID: 2, UserID: 10, Status: models.OrderStatusPending}
	svc := newTestService(payment, order)

	if err := svc.HandlePaymentSession(context.Background(), ref, "EXPIRED"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if payment.Status != models.PaymentStatusExpired {
		t.Fatalf("expected payment status expired, got %s", payment.Status)
	}
	if order.Status != models.OrderStatusPending {
		t.Fatalf("order status should be untouched, got %s", order.Status)
	}
}

func TestHandlePaymentSession_UnknownReference_ReturnsErrPaymentNotFound(t *testing.T) {
	ref := "sess-known"
	payment := &models.Payment{OrderID: 1, Status: models.PaymentStatusPending, PaymentReference: &ref}
	order := &models.Order{ID: 1, UserID: 10, Status: models.OrderStatusPending}
	svc := newTestService(payment, order)

	err := svc.HandlePaymentSession(context.Background(), "sess-unknown", "COMPLETED")
	if !errors.Is(err, webhook.ErrPaymentNotFound) {
		t.Fatalf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestHandlePaymentSession_ActiveStatus_NoOp(t *testing.T) {
	ref := "sess-789"
	payment := &models.Payment{OrderID: 3, Status: models.PaymentStatusPending, PaymentReference: &ref}
	order := &models.Order{ID: 3, UserID: 10, Status: models.OrderStatusPending}
	svc := newTestService(payment, order)

	if err := svc.HandlePaymentSession(context.Background(), ref, "ACTIVE"); err != nil {
		t.Fatalf("expected success (no-op), got %v", err)
	}
	if payment.Status != models.PaymentStatusPending {
		t.Fatalf("payment status should be untouched, got %s", payment.Status)
	}
}
