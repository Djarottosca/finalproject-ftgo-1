package auth

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/cache"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/user"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username already taken")
)

// Service defines the auth use cases exposed to the handler layer.
type Service interface {
	// Login authenticates and issues a token, but only if the account's
	// role matches expectedRole — keeps /user/login, /supplier/login,
	// /admin/login from cross-authenticating each other's accounts.
	Login(ctx context.Context, req LoginRequest, expectedRole string) (*LoginResponse, error)
	RegisterUser(ctx context.Context, req RegisterRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
}

type service struct {
	users     user.Repository
	db        *gorm.DB // only for resolving the "user" role id on self-register
	auth      *jwt.AuthManager
	blacklist *cache.TokenBlacklist
}

// NewService returns the Service implementation.
func NewService(users user.Repository, db *gorm.DB, auth *jwt.AuthManager, blacklist *cache.TokenBlacklist) Service {
	return &service{users: users, db: db, auth: auth, blacklist: blacklist}
}

func (s *service) Login(ctx context.Context, req LoginRequest, expectedRole string) (*LoginResponse, error) {
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

	// Wrong-portal login (e.g. a supplier account hitting /admin/login)
	// reported as the same generic error as bad credentials, so the
	// portal doesn't leak which roles exist for a given username.
	if u.Role.RoleSlug != expectedRole {
		return nil, ErrInvalidCredentials
	}

	token, err := s.auth.GenerateToken(u.ID, u.Role.RoleSlug)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: *token}, nil
}

// RegisterUser self-signs-up a "user"-role account and immediately logs
// them in. Suppliers register via supplier.Service.Register instead
// (creates the user + supplier profile together); admins are seeded.
func (s *service) RegisterUser(ctx context.Context, req RegisterRequest) (*LoginResponse, error) {
	if _, err := s.users.FindByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var role models.Role
	if err := s.db.WithContext(ctx).Where("role_slug = ?", models.RoleUser).First(&role).Error; err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		FullName:     req.FullName,
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		Status:       models.UserStatusActive,
		RoleID:       role.ID,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	token, err := s.auth.GenerateToken(u.ID, role.RoleSlug)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: *token}, nil
}

// Logout revokes the given token by blacklisting it in Redis for the
// remainder of its natural lifetime. Verification already happened in auth
// middleware; here we just read the expiry to size the blacklist TTL.
func (s *service) Logout(ctx context.Context, token string) error {
	claims, err := s.auth.VerifyToken(token)
	if err != nil {
		return err
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	return s.blacklist.Add(ctx, token, ttl)
}
