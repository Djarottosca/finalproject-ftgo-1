package user

import (
	"context"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for users, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) List(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Order("id").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *gormRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}
