package product

import (
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for products, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(product *models.Product) error
	FindByID(id int) (*models.Product, error)
	List(filter ListFilter) ([]models.Product, error)
	Update(product *models.Product) error
	Delete(id int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *gormRepository) FindByID(id int) (*models.Product, error) {
	var product models.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *gormRepository) List(filter ListFilter) ([]models.Product, error) {
	var products []models.Product
	q := r.db.Order("id")
	if filter.CategoryID != 0 {
		q = q.Where("category_id = ?", filter.CategoryID)
	}
	if filter.SupplierID != 0 {
		q = q.Where("supplier_id = ?", filter.SupplierID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if err := q.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *gormRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *gormRepository) Delete(id int) error {
	return r.db.Delete(&models.Product{}, id).Error
}
