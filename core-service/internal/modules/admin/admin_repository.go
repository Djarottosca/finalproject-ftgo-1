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
	SalesSummary(ctx context.Context) (SalesSummary, error)
	OrdersByStatus(ctx context.Context) ([]StatusCount, error)
	ListTransactions(ctx context.Context) ([]TransactionItem, error)
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

// SalesSummary sums revenue from PAID orders only. It joins orders to payments
// so "revenue" reflects money actually received, not orders still pending.
func (r *gormRepository) SalesSummary(ctx context.Context) (SalesSummary, error) {
	var summary SalesSummary
	err := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Joins("JOIN payments ON payments.order_id = orders.id").
		Where("payments.status = ?", models.PaymentStatusPaid).
		Select(`
			COALESCE(SUM(orders.final_price), 0) AS total_revenue,
			COUNT(*) AS paid_orders,
			COALESCE(AVG(orders.final_price), 0) AS avg_order_value`).
		Scan(&summary).Error
	if err != nil {
		return SalesSummary{}, err
	}
	return summary, nil
}

// OrdersByStatus counts orders grouped by their payment status, giving admin a
// funnel view (how many pending vs paid vs failed).
func (r *gormRepository) OrdersByStatus(ctx context.Context) ([]StatusCount, error) {
	var counts []StatusCount
	err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("status, COUNT(*) AS count").
		Group("status").
		Order("status").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	return counts, nil
}

// ListTransactions returns every order with its payment status, newest first,
// for the read-only monitoring view. LEFT JOIN so an order without a payment
// row still shows up.
func (r *gormRepository) ListTransactions(ctx context.Context) ([]TransactionItem, error) {
	var items []TransactionItem
	err := r.db.WithContext(ctx).
		Table("orders").
		Joins("LEFT JOIN payments ON payments.order_id = orders.id").
		Select(`
			orders.id AS order_id,
			orders.user_id AS user_id,
			orders.final_price AS final_price,
			orders.status AS order_status,
			COALESCE(payments.status, 'none') AS payment_status,
			COALESCE(payments.payment_reference, '') AS payment_ref`).
		Order("orders.id DESC").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
