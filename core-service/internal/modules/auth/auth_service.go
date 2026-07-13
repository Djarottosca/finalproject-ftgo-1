package auth

import (
	"errors"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type credential struct {
	Password string
	UserID   int
	Role     string
}

// ponytail: hardcoded credentials, no users table yet. Swap for a repository
// lookup + bcrypt compare once the user module lands.
var staticUsers = map[string]credential{
	"admin":    {Password: "admin123", UserID: 1, Role: "admin"},
	"supplier": {Password: "supplier123", UserID: 2, Role: "supplier"},
	"user":     {Password: "user123", UserID: 3, Role: "user"},
}

// Service defines the auth use cases exposed to the handler layer.
type Service interface {
	Login(req LoginRequest) (*LoginResponse, error)
}

type service struct {
	auth *jwt.AuthManager
}

// NewService returns the Service implementation.
func NewService(auth *jwt.AuthManager) Service {
	return &service{auth: auth}
}

func (s *service) Login(req LoginRequest) (*LoginResponse, error) {
	cred, ok := staticUsers[req.Username]
	if !ok || cred.Password != req.Password {
		return nil, ErrInvalidCredentials
	}

	token, err := s.auth.GenerateToken(cred.UserID, cred.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: *token}, nil
}
