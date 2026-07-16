package admin

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository is the read-only persistence contract for admin reporting.
// It reads tables owned by other modules (products, and later orders/payments)
// directly, following the same cross-table pattern as order.ListBySupplierID.
type Repository interface {
	ListProductsByStock(ctx context.Context) ([]models.Product, error)
	StockSummary(ctx context.Context, threshold int) (StockSummary, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// ListProductsByStock returns all products ordered by stock ascending, so the
// items most at risk of running out appear first.
func (r *gormRepository) ListProductsByStock(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).Order("stock asc").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// StockSummary computes the aggregate rollup in a single query, instead of
// pulling every row into memory and counting in Go.
func (r *gormRepository) StockSummary(ctx context.Context, threshold int) (StockSummary, error) {
	var summary StockSummary
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select(`
			COUNT(*) AS total_products,
			COALESCE(SUM(stock), 0) AS total_stock,
			COUNT(*) FILTER (WHERE stock < ?) AS low_stock_count,
			COUNT(*) FILTER (WHERE stock = 0) AS out_of_stock_count`,
			threshold).
		Scan(&summary).Error
	if err != nil {
		return StockSummary{}, err
	}
	return summary, nil
}
