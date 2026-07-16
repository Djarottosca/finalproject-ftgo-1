package supplier

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/slug"
)

var (
	ErrNotFound      = errors.New("supplier not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidStatus = errors.New("status must be approved or rejected")
)

// Service defines the supplier use cases exposed to the handler layer.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	Get(ctx context.Context, id int) (*SupplierResponse, error)
	List(ctx context.Context, status string) ([]SupplierResponse, error)
	Review(ctx context.Context, id int, req ReviewRequest) (*SupplierResponse, error)
}

type service struct {
	repo Repository
	db   *gorm.DB // transactional user+supplier create, and role_id lookup
	auth *jwt.AuthManager
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository, db *gorm.DB, auth *jwt.AuthManager) Service {
	return &service{repo: repo, db: db, auth: auth}
}

// Register self-signs-up a "supplier"-role user account and its supplier
// profile in one step — no prior /admin/users or /user/register call needed.
func (s *service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	var user models.User
	var supplier models.Supplier

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.User
		if err := tx.Where("username = ?", req.Username).First(&existing).Error; err == nil {
			return ErrUsernameTaken
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var role models.Role
		if err := tx.Where("role_slug = ?", models.RoleSupplier).First(&role).Error; err != nil {
			return err
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		user = models.User{
			FullName:     req.FullName,
			Username:     req.Username,
			PasswordHash: string(hash),
			Email:        req.Email,
			Status:       models.UserStatusActive,
			RoleID:       role.ID,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		supplier = models.Supplier{
			UserID:       user.ID,
			StoreName:    req.StoreName,
			SupplierSlug: slug.Generate(req.StoreName),
			Address:      req.Address,
			Status:       models.SupplierStatusPending,
		}
		return tx.Create(&supplier).Error
	})
	if err != nil {
		return nil, err
	}

	token, err := s.auth.GenerateToken(user.ID, models.RoleSupplier)
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{Token: *token, Supplier: *toResponse(&supplier)}, nil
}

func (s *service) Get(ctx context.Context, id int) (*SupplierResponse, error) {
	supplier, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toResponse(supplier), nil
}

func (s *service) List(ctx context.Context, status string) ([]SupplierResponse, error) {
	suppliers, err := s.repo.List(ctx, status)
	if err != nil {
		return nil, err
	}

	res := make([]SupplierResponse, 0, len(suppliers))
	for _, sup := range suppliers {
		res = append(res, *toResponse(&sup))
	}
	return res, nil
}

func (s *service) Review(ctx context.Context, id int, req ReviewRequest) (*SupplierResponse, error) {
	if req.Status != models.SupplierStatusApproved && req.Status != models.SupplierStatusRejected {
		return nil, ErrInvalidStatus
	}

	supplier, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	supplier.Status = req.Status
	if err := s.repo.Update(ctx, supplier); err != nil {
		return nil, err
	}

	return toResponse(supplier), nil
}
