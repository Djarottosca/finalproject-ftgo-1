package product_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/product"
)

type mockRepository struct {
	products []models.Product
	total    int64
	findErr  error
}

func (m *mockRepository) FindAll(_ context.Context, _ product.ListFilter) ([]models.Product, int64, error) {
	if m.findErr != nil {
		return nil, 0, m.findErr
	}
	return m.products, m.total, nil
}

func (m *mockRepository) FindBySlug(_ context.Context, slug string) (*models.Product, error) {
	for _, p := range m.products {
		if p.ProductSlug == slug {
			return &p, nil
		}
	}
	if m.findErr != nil {
		return nil, m.findErr
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepository) FindByID(_ context.Context, id int) (*models.Product, error) {
	for _, p := range m.products {
		if p.ID == id {
			return &p, nil
		}
	}
	if m.findErr != nil {
		return nil, m.findErr
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepository) FindAllBySupplier(_ context.Context, supplierID int) ([]models.Product, error) {
	var result []models.Product
	for _, p := range m.products {
		if p.SupplierID == supplierID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockRepository) Create(_ context.Context, p *models.Product) error {
	m.products = append(m.products, *p)
	return nil
}

func (m *mockRepository) Update(_ context.Context, p *models.Product) error {
	for i := range m.products {
		if m.products[i].ID == p.ID {
			m.products[i] = *p
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *mockRepository) Delete(_ context.Context, id int) error {
	for i := range m.products {
		if m.products[i].ID == id {
			m.products = append(m.products[:i], m.products[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

type mockCache struct {
	store map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]string)}
}

func (c *mockCache) Get(_ context.Context, key string) (string, error) {
	val, ok := c.store[key]
	if !ok {
		return "", errCacheMiss
	}
	return val, nil
}

func (c *mockCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	c.store[key] = value
	return nil
}

type cacheMissErr struct{}

func (cacheMissErr) Error() string { return "cache miss" }

var errCacheMiss = cacheMissErr{}

func TestList_CacheMissThenHit(t *testing.T) {
	repo := &mockRepository{
		products: []models.Product{
			{ID: 1, ProductName: "Semen Tiga Roda 40kg", ProductSlug: "semen-tiga-roda-40kg-1", Price: 65000, Status: "active"},
		},
		total: 1,
	}
	cache := newMockCache()
	service := product.NewService(repo, cache)

	req := product.ListRequest{Page: 1, Limit: 20}

	res1, err := service.List(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res1.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(res1.Items))
	}
	if res1.Items[0].ProductName != "Semen Tiga Roda 40kg" {
		t.Errorf("unexpected product name: %s", res1.Items[0].ProductName)
	}

	repo.products[0].ProductName = "Nama Berubah"

	res2, err := service.List(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res2.Items[0].ProductName != "Semen Tiga Roda 40kg" {
		t.Errorf("expected cached result, got fresh data: %s", res2.Items[0].ProductName)
	}
}

func TestDetail_NotFound(t *testing.T) {
	repo := &mockRepository{findErr: nil} // kosong
	cache := newMockCache()
	service := product.NewService(repo, cache)

	_, err := service.Detail(context.Background(), "produk-yang-gak-ada")
	if err == nil {
		t.Fatal("expected error for non-existent product, got nil")
	}
}

func TestFinalPrice_PercentageDiscount(t *testing.T) {
	discountType := "percentage"
	discountAmount := 10.0
	p := models.Product{Price: 100000, DiscountType: &discountType, DiscountAmount: &discountAmount}

	got := p.FinalPrice()
	want := 90000.0

	if got != want {
		t.Errorf("expected final price %.2f, got %.2f", want, got)
	}
}

func TestFinalPrice_NoDiscount(t *testing.T) {
	p := models.Product{Price: 50000}

	got := p.FinalPrice()
	if got != 50000 {
		t.Errorf("expected final price to equal original price, got %.2f", got)
	}
}
