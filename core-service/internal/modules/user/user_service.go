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
type Service interface {
	Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	Get(ctx context.Context, id int) (*UserResponse, error)
	List(ctx context.Context) ([]UserResponse, error)
	Update(ctx context.Context, id int, req UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
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

func (s *service) Get(ctx context.Context, id int) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toResponse(user), nil
}

func (s *service) List(ctx context.Context) ([]UserResponse, error) {
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

func (s *service) Update(ctx context.Context, id int, req UpdateUserRequest) (*UserResponse, error) {
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

func (s *service) Delete(ctx context.Context, id int) error {
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
