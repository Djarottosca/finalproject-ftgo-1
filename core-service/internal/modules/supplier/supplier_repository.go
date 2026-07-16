package supplier

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for suppliers, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(ctx context.Context, supplier *models.Supplier) error
	FindByID(ctx context.Context, id int) (*models.Supplier, error)
	FindByUserID(ctx context.Context, userID int) (*models.Supplier, error)
	List(ctx context.Context, status string) ([]models.Supplier, error)
	Update(ctx context.Context, supplier *models.Supplier) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, supplier *models.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id int) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := r.db.WithContext(ctx).First(&supplier, id).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *gormRepository) FindByUserID(ctx context.Context, userID int) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&supplier).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *gormRepository) List(ctx context.Context, status string) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	q := r.db.WithContext(ctx).Order("id")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *gormRepository) Update(ctx context.Context, supplier *models.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}
