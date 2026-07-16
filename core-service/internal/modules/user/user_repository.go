package user

import (
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for users, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(user *models.User) error
	FindByID(id int) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	List() ([]models.User, error)
	Update(user *models.User) error
	Delete(id int) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *gormRepository) FindByID(id int) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Role").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) List() ([]models.User, error) {
	var users []models.User
	if err := r.db.Order("id").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *gormRepository) Delete(id int) error {
	return r.db.Delete(&models.User{}, id).Error
}
