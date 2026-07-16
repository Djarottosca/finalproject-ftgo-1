package product

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// repository
type ListFilter struct {
	Keyword    string
	CategoryID int
	Page       int
	Limit      int
}

type Repository interface {
	FindAll(ctx context.Context, filter ListFilter) ([]models.Product, int64, error)
	FindBySlug(ctx context.Context, slug string) (*models.Product, error)
	FindByID(ctx context.Context, id int) (*models.Product, error)
	FindAllBySupplier(ctx context.Context, supplierID int) ([]models.Product, error)
	Create(ctx context.Context, product *models.Product) error
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id int) error
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

func (r *gormRepository) FindByID(ctx context.Context, id int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Preload("Category").First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// FindAllBySupplier returns every product owned by a supplier regardless of
// status, so the supplier can also see their own inactive listings.
func (r *gormRepository) FindAllBySupplier(ctx context.Context, supplierID int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("supplier_id = ?", supplierID).
		Order("created_at DESC").
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *gormRepository) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *gormRepository) Update(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *gormRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, id).Error
}
