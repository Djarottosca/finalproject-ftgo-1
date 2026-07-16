package cart_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/cart"
)

// mockRepository
type mockRepository struct {
	rows map[string]models.Cart
}

func newMockRepository() *mockRepository {
	return &mockRepository{rows: make(map[string]models.Cart)}
}

func key(userID int, productID int) string {
	return fmt.Sprintf("%d:%d", userID, productID)
}

func (m *mockRepository) Upsert(_ context.Context, userID int, productID int, qty int) error {
	k := key(userID, productID)
	row, exists := m.rows[k]
	if exists {
		row.Qty += qty
	} else {
		row = models.Cart{UserID: userID, ProductID: productID, Qty: qty}
	}
	m.rows[k] = row
	return nil
}

func (m *mockRepository) UpdateQty(_ context.Context, userID int, productID int, qty int) (int64, error) {
	k := key(userID, productID)
	row, exists := m.rows[k]
	if !exists {
		return 0, nil
	}
	row.Qty = qty
	m.rows[k] = row
	return 1, nil
}

func (m *mockRepository) Delete(_ context.Context, userID int, productID int) (int64, error) {
	k := key(userID, productID)
	if _, exists := m.rows[k]; !exists {
		return 0, nil
	}
	delete(m.rows, k)
	return 1, nil
}

func (m *mockRepository) FindAllByUser(_ context.Context, userID int) ([]models.Cart, error) {
	var result []models.Cart
	for _, row := range m.rows {
		if row.UserID == userID {
			result = append(result, row)
		}
	}
	return result, nil
}

// mockProductLookup
type mockProductLookup struct {
	products map[int]models.Product
}

func newMockProductLookup() *mockProductLookup {
	return &mockProductLookup{products: make(map[int]models.Product)}
}

func (m *mockProductLookup) FindByID(_ context.Context, id int) (*models.Product, error) {
	p, ok := m.products[id]
	if !ok {

		return nil, errors.New("record not found")
	}
	return &p, nil
}

func TestAddItem_Success(t *testing.T) {
	repo := newMockRepository()
	products := newMockProductLookup()
	products.products[1] = models.Product{ID: 1, ProductName: "Semen Tiga Roda 40kg", ProductSlug: "semen-1", Price: 65000, Stock: 10, Status: "active"}

	svc := cart.NewService(repo, products)

	res, err := svc.AddItem(context.Background(), 100, cart.AddItemRequest{ProductID: 1, Qty: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.TotalItems != 2 {
		t.Errorf("expected total items 2, got %d", res.TotalItems)
	}
	if res.TotalPrice != 130000 {
		t.Errorf("expected total price 130000, got %.2f", res.TotalPrice)
	}
}

func TestAddItem_InsufficientStock(t *testing.T) {
	repo := newMockRepository()
	products := newMockProductLookup()
	products.products[1] = models.Product{ID: 1, Price: 65000, Stock: 1, Status: "active"}

	svc := cart.NewService(repo, products)

	_, err := svc.AddItem(context.Background(), 100, cart.AddItemRequest{ProductID: 1, Qty: 5})
	if !errors.Is(err, cart.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestAddItem_ProductNotFound(t *testing.T) {
	repo := newMockRepository()
	products := newMockProductLookup() // kosong

	svc := cart.NewService(repo, products)

	_, err := svc.AddItem(context.Background(), 100, cart.AddItemRequest{ProductID: 99, Qty: 1})
	if !errors.Is(err, cart.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestAddItem_ProductInactive(t *testing.T) {
	repo := newMockRepository()
	products := newMockProductLookup()
	products.products[1] = models.Product{ID: 1, Price: 65000, Stock: 10, Status: "inactive"}

	svc := cart.NewService(repo, products)

	_, err := svc.AddItem(context.Background(), 100, cart.AddItemRequest{ProductID: 1, Qty: 1})
	if !errors.Is(err, cart.ErrProductInactive) {
		t.Fatalf("expected ErrProductInactive, got %v", err)
	}
}

func TestUpdateItem_NotInCart(t *testing.T) {
	repo := newMockRepository()
	products := newMockProductLookup()
	products.products[1] = models.Product{ID: 1, Price: 65000, Stock: 10, Status: "active"}

	svc := cart.NewService(repo, products)

	// belum pernah AddItem, langsung UpdateItem
	_, err := svc.UpdateItem(context.Background(), 100, 1, cart.UpdateItemRequest{Qty: 3})
	if !errors.Is(err, cart.ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
}

func TestRemoveItem_Success(t *testing.T) {
	repo := newMockRepository()
	products := newMockProductLookup()
	products.products[1] = models.Product{ID: 1, Price: 65000, Stock: 10, Status: "active"}

	svc := cart.NewService(repo, products)
	if _, err := svc.AddItem(context.Background(), 100, cart.AddItemRequest{ProductID: 1, Qty: 1}); err != nil {
		t.Fatalf("setup AddItem failed: %v", err)
	}

	res, err := svc.RemoveItem(context.Background(), 100, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.Items) != 0 {
		t.Errorf("expected empty cart, got %d items", len(res.Items))
	}
}
