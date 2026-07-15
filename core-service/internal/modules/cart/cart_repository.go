package cart

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for cart items, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Upsert(item *models.Cart) error
	FindByUserAndProduct(userID, productID int) (*models.Cart, error)
	ListByUserID(userID int) ([]models.Cart, error)
	Delete(userID, productID int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Upsert inserts the cart item, or updates qty if the (user_id, product_id)
// row already exists.
func (r *gormRepository) Upsert(item *models.Cart) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"qty", "updated_at"}),
	}).Create(item).Error
}

func (r *gormRepository) FindByUserAndProduct(userID, productID int) (*models.Cart, error) {
	var item models.Cart
	if err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormRepository) ListByUserID(userID int) ([]models.Cart, error) {
	var items []models.Cart
	if err := r.db.Where("user_id = ?", userID).Order("product_id").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *gormRepository) Delete(userID, productID int) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Cart{}).Error
}
