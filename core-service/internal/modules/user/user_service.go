package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var ErrNotFound = errors.New("user not found")

// Service defines the user use cases exposed to the handler layer.
// Self-service methods (Me/UpdateMe) act on the caller's own account;
// Admin* methods operate on an arbitrary user by ID and are reserved for
// admin-only routes.
type Service interface {
	AdminCreate(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	Me(ctx context.Context, id int) (*UserResponse, error)
	UpdateMe(ctx context.Context, id int, req UpdateProfileRequest) (*UserResponse, error)
	AdminList(ctx context.Context) ([]UserResponse, error)
	AdminGetByID(ctx context.Context, id int) (*UserResponse, error)
	AdminUpdate(ctx context.Context, id int, req UpdateUserRequest) (*UserResponse, error)
	AdminDelete(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// AdminCreate creates a user account with an arbitrary role_id — admin-only,
// since letting anyone pick their own role would let a caller self-declare
// as admin. Self-signup goes through auth.RegisterUser (role=user, fixed)
// or supplier.Register (role=supplier, fixed) instead.
func (s *service) AdminCreate(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FullName:     req.FullName,
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		Status:       models.UserStatusActive,
		RoleID:       req.RoleID,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toResponse(user), nil
}

// Me returns the caller's own profile (id comes from the JWT, not a path
// param — every role can read its own account this way).
func (s *service) Me(ctx context.Context, id int) (*UserResponse, error) {
	return s.AdminGetByID(ctx, id)
}

// UpdateMe lets the caller edit their own profile. Deliberately narrower
// than AdminUpdate: no Status field, so a user can't self-reactivate/ban.
func (s *service) UpdateMe(ctx context.Context, id int, req UpdateProfileRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	user.FullName = req.FullName
	user.Email = req.Email
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toResponse(user), nil
}

func (s *service) AdminGetByID(ctx context.Context, id int) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toResponse(user), nil
}

func (s *service) AdminList(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]UserResponse, 0, len(users))
	for _, u := range users {
		res = append(res, *toResponse(&u))
	}
	return res, nil
}

func (s *service) AdminUpdate(ctx context.Context, id int, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	user.FullName = req.FullName
	user.Email = req.Email
	user.Status = req.Status
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toResponse(user), nil
}

func (s *service) AdminDelete(ctx context.Context, id int) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.repo.Delete(ctx, user.ID); err != nil {
		return err
	}
	return nil
}
