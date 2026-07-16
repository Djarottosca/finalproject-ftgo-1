package shipping

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository is the persistence contract for the shipments table (1:1 with
// orders). It's separate from the RajaOngkir-backed Service in this same
// package: Service talks to the external API, Repository owns the DB row
// that tracks an order's shipment.
type Repository interface {
	Create(ctx context.Context, orderID int) (*models.Shipment, error)
	FindByOrderID(ctx context.Context, orderID int) (*models.Shipment, error)
	MarkShipped(ctx context.Context, orderID int, courier, trackingNumber string) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create inserts a pending shipment row for a freshly checked-out order.
// Courier/tracking number/cost are filled in later, once the order actually
// ships — out of scope until that flow is built.
func (r *gormRepository) Create(ctx context.Context, orderID int) (*models.Shipment, error) {
	s := &models.Shipment{
		OrderID: orderID,
		Status:  models.ShipmentStatusPending,
	}
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

func (r *gormRepository) FindByOrderID(ctx context.Context, orderID int) (*models.Shipment, error) {
	var s models.Shipment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// MarkShipped fills in courier/tracking_number and flips status to shipped,
// once the supplier has actually handed the order to a courier.
func (r *gormRepository) MarkShipped(ctx context.Context, orderID int, courier, trackingNumber string) error {
	return r.db.WithContext(ctx).Model(&models.Shipment{}).Where("order_id = ?", orderID).
		Updates(map[string]any{
			"status":          models.ShipmentStatusShipped,
			"courier":         courier,
			"tracking_number": trackingNumber,
		}).Error
}
