package address

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for addresses, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(ctx context.Context, addr *models.Address) error
	FindByID(ctx context.Context, id int) (*models.Address, error)
	ListByUserID(ctx context.Context, userID int) ([]models.Address, error)
	Update(ctx context.Context, addr *models.Address) error
	Delete(ctx context.Context, id int) error
	ClearPrimary(ctx context.Context, userID int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, addr *models.Address) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id int) (*models.Address, error) {
	var addr models.Address
	if err := r.db.WithContext(ctx).First(&addr, id).Error; err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *gormRepository) ListByUserID(ctx context.Context, userID int) ([]models.Address, error) {
	var addrs []models.Address
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id").Find(&addrs).Error; err != nil {
		return nil, err
	}
	return addrs, nil
}

func (r *gormRepository) Update(ctx context.Context, addr *models.Address) error {
	return r.db.WithContext(ctx).Save(addr).Error
}

func (r *gormRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Address{}, id).Error
}

// ClearPrimary unsets is_primary on all of the user's addresses, so a new
// primary can be set without violating the one-primary-per-user index.
func (r *gormRepository) ClearPrimary(ctx context.Context, userID int) error {
	return r.db.WithContext(ctx).Model(&models.Address{}).Where("user_id = ? AND is_primary", userID).
		Update("is_primary", false).Error
}
