package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/user"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

// Service defines the auth use cases exposed to the handler layer.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

type service struct {
	users user.Repository
	auth  *jwt.AuthManager
}

// NewService returns the Service implementation.
func NewService(users user.Repository, auth *jwt.AuthManager) Service {
	return &service{users: users, auth: auth}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	u, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.auth.GenerateToken(u.ID, u.Role.RoleSlug)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: *token}, nil
}
