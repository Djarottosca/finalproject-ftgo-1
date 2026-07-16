package review

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for reviews, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(ctx context.Context, review *models.Review) error
	FindByUserAndProduct(ctx context.Context, userID, productID int) (*models.Review, error)
	ListByProductID(ctx context.Context, productID int) ([]models.Review, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, review *models.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *gormRepository) FindByUserAndProduct(ctx context.Context, userID, productID int) (*models.Review, error) {
	var review models.Review
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *gormRepository) ListByProductID(ctx context.Context, productID int) ([]models.Review, error) {
	var reviews []models.Review
	if err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}
