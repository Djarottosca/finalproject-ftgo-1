package productimage

import (
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for product images, so tests
// can substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(image *models.ProductImage) error
	FindByID(id int) (*models.ProductImage, error)
	ListByProductID(productID int) ([]models.ProductImage, error)
	Delete(id int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(image *models.ProductImage) error {
	return r.db.Create(image).Error
}

func (r *gormRepository) FindByID(id int) (*models.ProductImage, error) {
	var image models.ProductImage
	if err := r.db.First(&image, id).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *gormRepository) ListByProductID(productID int) ([]models.ProductImage, error) {
	var images []models.ProductImage
	if err := r.db.Where("product_id = ?", productID).Order("id").Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

func (r *gormRepository) Delete(id int) error {
	return r.db.Delete(&models.ProductImage{}, id).Error
}
