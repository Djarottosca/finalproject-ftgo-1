package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var ErrNotFound = errors.New("user not found")

// Service defines the user use cases exposed to the handler layer.
type Service interface {
	Create(req CreateUserRequest) (*UserResponse, error)
	Get(id int) (*UserResponse, error)
	List() ([]UserResponse, error)
	Update(id int, req UpdateUserRequest) (*UserResponse, error)
	Delete(id int) error
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(req CreateUserRequest) (*UserResponse, error) {
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
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:       user.ID,
		FullName: user.FullName,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
		RoleID:   user.RoleID,
	}, nil
}

func (s *service) Get(id int) (*UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &UserResponse{
		ID:       user.ID,
		FullName: user.FullName,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
		RoleID:   user.RoleID,
	}, nil
}

func (s *service) List() ([]UserResponse, error) {
	users, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	res := make([]UserResponse, 0, len(users))
	for _, u := range users {
		res = append(res, UserResponse{
			ID:       u.ID,
			FullName: u.FullName,
			Username: u.Username,
			Email:    u.Email,
			Status:   u.Status,
			RoleID:   u.RoleID,
		})
	}
	return res, nil
}

func (s *service) Update(id int, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	user.FullName = req.FullName
	user.Email = req.Email
	user.Status = req.Status
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:       user.ID,
		FullName: user.FullName,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
		RoleID:   user.RoleID,
	}, nil
}

func (s *service) Delete(id int) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.repo.Delete(user.ID); err != nil {
		return err
	}
	return nil
}
