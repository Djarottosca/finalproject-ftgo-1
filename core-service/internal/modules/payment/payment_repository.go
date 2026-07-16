package payment

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository is the persistence contract for the payment module.
//
// It reads orders and users READ-ONLY (to know how much to charge and who the
// customer is) and owns the payments table for writes. Reading another
// module's table read-only follows the same pattern as admin reading products
// and order joining products; it never writes orders or users.
type Repository interface {
	FindOrderByID(ctx context.Context, id int) (*models.Order, error)
	FindUserByID(ctx context.Context, id int) (*models.User, error)
	FindPaymentByOrderID(ctx context.Context, orderID int) (*models.Payment, error)
	CreatePayment(ctx context.Context, p *models.Payment) error
	UpdatePaymentStatus(ctx context.Context, p *models.Payment) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) FindOrderByID(ctx context.Context, id int) (*models.Order, error) {
	var o models.Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *gormRepository) FindUserByID(ctx context.Context, id int) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormRepository) FindPaymentByOrderID(ctx context.Context, orderID int) (*models.Payment, error) {
	var p models.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *gormRepository) CreatePayment(ctx context.Context, p *models.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// UpdatePaymentStatus only touches the status column, so a status refresh can't
// accidentally overwrite the link or reference stored at creation time.
func (r *gormRepository) UpdatePaymentStatus(ctx context.Context, p *models.Payment) error {
	return r.db.WithContext(ctx).Model(p).Update("status", p.Status).Error
}
