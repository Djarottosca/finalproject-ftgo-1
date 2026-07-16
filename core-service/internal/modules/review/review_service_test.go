package review_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/product"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/review"
)

// mockProductRepo implements product.Repository, only FindByID is exercised.
type mockProductRepo struct{ products map[int]models.Product }

func (m *mockProductRepo) FindAll(context.Context, product.ListFilter) ([]models.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) FindBySlug(context.Context, string) (*models.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) FindByID(_ context.Context, id int) (*models.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &p, nil
}
func (m *mockProductRepo) FindAllBySupplier(context.Context, int) ([]models.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) FindAllAdmin(context.Context) ([]models.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) Create(context.Context, *models.Product) error { return nil }
func (m *mockProductRepo) Update(context.Context, *models.Product) error { return nil }
func (m *mockProductRepo) Delete(context.Context, int) error             { return nil }

// mockReviewRepo implements review.Repository in-memory.
type mockReviewRepo struct {
	byUserProduct map[[2]int]models.Review
}

func (m *mockReviewRepo) Create(_ context.Context, r *models.Review) error {
	if m.byUserProduct == nil {
		m.byUserProduct = map[[2]int]models.Review{}
	}
	m.byUserProduct[[2]int{r.UserID, r.ProductID}] = *r
	return nil
}
func (m *mockReviewRepo) FindByUserAndProduct(_ context.Context, userID, productID int) (*models.Review, error) {
	r, ok := m.byUserProduct[[2]int{userID, productID}]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &r, nil
}
func (m *mockReviewRepo) ListByProductID(context.Context, int) ([]models.Review, error) {
	return nil, nil
}

// TestCreate_RejectsDuplicateReview verifies a user can't review the same
// product twice, and can't review a product that doesn't exist.
func TestCreate_RejectsDuplicateReview(t *testing.T) {
	productRepo := &mockProductRepo{products: map[int]models.Product{1: {ID: 1}}}
	reviewRepo := &mockReviewRepo{}
	svc := review.NewService(reviewRepo, productRepo)

	req := review.CreateReviewRequest{Rating: 5, Comment: "great"}
	if _, err := svc.Create(context.Background(), 100, 1, req); err != nil {
		t.Fatalf("expected first review to succeed, got %v", err)
	}

	if _, err := svc.Create(context.Background(), 100, 1, req); !errors.Is(err, review.ErrAlreadyReviewed) {
		t.Fatalf("expected ErrAlreadyReviewed, got %v", err)
	}

	if _, err := svc.Create(context.Background(), 100, 999, req); !errors.Is(err, review.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}
