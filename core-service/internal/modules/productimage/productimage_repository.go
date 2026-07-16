package productimage

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for product images, so tests
// can substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(ctx context.Context, image *models.ProductImage) error
	FindByID(ctx context.Context, id int) (*models.ProductImage, error)
	ListByProductID(ctx context.Context, productID int) ([]models.ProductImage, error)
	Delete(ctx context.Context, id int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, image *models.ProductImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id int) (*models.ProductImage, error) {
	var image models.ProductImage
	if err := r.db.WithContext(ctx).First(&image, id).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *gormRepository) ListByProductID(ctx context.Context, productID int) ([]models.ProductImage, error) {
	var images []models.ProductImage
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("id").Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

func (r *gormRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.ProductImage{}, id).Error
}
