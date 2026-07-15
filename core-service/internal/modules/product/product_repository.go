package product

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// repository
type ListFilter struct {
	Keyword    string
	CategoryID uint64
	Page       int
	Limit      int
}

type Repository interface {
	FindAll(ctx context.Context, filter ListFilter) ([]models.Product, int64, error)
	FindBySlug(ctx context.Context, slug string) (*models.Product, error)
	FindByID(ctx context.Context, id uint64) (*models.Product, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// FindAll for active
func (r *gormRepository) FindAll(ctx context.Context, filter ListFilter) ([]models.Product, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Product{}).
		Preload("Category").
		Where("status = ?", "active")

	if filter.Keyword != "" {
		query = query.Where("product_name ILIKE ?", "%"+filter.Keyword+"%")
	}
	if filter.CategoryID != 0 {
		query = query.Where("category_id = ?", filter.CategoryID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit

	var products []models.Product
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *gormRepository) FindBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Images").
		Where("product_slug = ? AND status = ?", slug, "active").
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *gormRepository) FindByID(ctx context.Context, id uint64) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
