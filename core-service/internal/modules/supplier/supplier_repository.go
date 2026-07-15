package supplier

import (
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for suppliers, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(supplier *models.Supplier) error
	FindByID(id int) (*models.Supplier, error)
	FindByUserID(userID int) (*models.Supplier, error)
	List(status string) ([]models.Supplier, error)
	Update(supplier *models.Supplier) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(supplier *models.Supplier) error {
	return r.db.Create(supplier).Error
}

func (r *gormRepository) FindByID(id int) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := r.db.First(&supplier, id).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *gormRepository) FindByUserID(userID int) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := r.db.Where("user_id = ?", userID).First(&supplier).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *gormRepository) List(status string) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	q := r.db.Order("id")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *gormRepository) Update(supplier *models.Supplier) error {
	return r.db.Save(supplier).Error
}
