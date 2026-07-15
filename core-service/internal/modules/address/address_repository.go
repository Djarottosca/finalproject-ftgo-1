package address

import (
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for addresses, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(addr *models.Address) error
	FindByID(id int) (*models.Address, error)
	ListByUserID(userID int) ([]models.Address, error)
	Update(addr *models.Address) error
	Delete(id int) error
	ClearPrimary(userID int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(addr *models.Address) error {
	return r.db.Create(addr).Error
}

func (r *gormRepository) FindByID(id int) (*models.Address, error) {
	var addr models.Address
	if err := r.db.First(&addr, id).Error; err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *gormRepository) ListByUserID(userID int) ([]models.Address, error) {
	var addrs []models.Address
	if err := r.db.Where("user_id = ?", userID).Order("id").Find(&addrs).Error; err != nil {
		return nil, err
	}
	return addrs, nil
}

func (r *gormRepository) Update(addr *models.Address) error {
	return r.db.Save(addr).Error
}

func (r *gormRepository) Delete(id int) error {
	return r.db.Delete(&models.Address{}, id).Error
}

// ClearPrimary unsets is_primary on all of the user's addresses, so a new
// primary can be set without violating the one-primary-per-user index.
func (r *gormRepository) ClearPrimary(userID int) error {
	return r.db.Model(&models.Address{}).Where("user_id = ? AND is_primary", userID).
		Update("is_primary", false).Error
}
