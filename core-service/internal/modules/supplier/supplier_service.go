package supplier

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var (
	ErrNotFound        = errors.New("supplier not found")
	ErrAlreadyExists   = errors.New("user already registered as supplier")
	ErrInvalidStatus   = errors.New("status must be approved or rejected")
	slugNonAlnumRegexp = regexp.MustCompile(`[^a-z0-9]+`)
)

// Repository defines the persistence contract for suppliers, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	Create(supplier *models.Supplier) error
	FindByID(id int) (*models.Supplier, error)
	FindByUserID(userID int) (*models.Supplier, error)
	List(status string) ([]models.Supplier, error)
	Update(supplier *models.Supplier) error
}

// Service defines the supplier use cases exposed to the handler layer.
type Service interface {
	Register(userID int, req RegisterRequest) (*SupplierResponse, error)
	Get(id int) (*SupplierResponse, error)
	List(status string) ([]SupplierResponse, error)
	Review(id int, req ReviewRequest) (*SupplierResponse, error)
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Register(userID int, req RegisterRequest) (*SupplierResponse, error) {
	if _, err := s.repo.FindByUserID(userID); err == nil {
		return nil, ErrAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	supplier := &models.Supplier{
		UserID:       userID,
		StoreName:    req.StoreName,
		SupplierSlug: generateSlug(req.StoreName),
		Address:      req.Address,
		Status:       models.SupplierStatusPending,
	}
	if err := s.repo.Create(supplier); err != nil {
		return nil, err
	}

	return toResponse(supplier), nil
}

func (s *service) Get(id int) (*SupplierResponse, error) {
	supplier, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toResponse(supplier), nil
}

func (s *service) List(status string) ([]SupplierResponse, error) {
	suppliers, err := s.repo.List(status)
	if err != nil {
		return nil, err
	}

	res := make([]SupplierResponse, 0, len(suppliers))
	for _, sup := range suppliers {
		res = append(res, *toResponse(&sup))
	}
	return res, nil
}

func (s *service) Review(id int, req ReviewRequest) (*SupplierResponse, error) {
	if req.Status != models.SupplierStatusApproved && req.Status != models.SupplierStatusRejected {
		return nil, ErrInvalidStatus
	}

	supplier, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	supplier.Status = req.Status
	if err := s.repo.Update(supplier); err != nil {
		return nil, err
	}

	return toResponse(supplier), nil
}

// generateSlug builds a URL-safe, unique-enough slug from the store name.
// ponytail: no DB collision check beyond the unique constraint + random
// suffix. Good enough at this scale; add a retry-on-conflict loop if
// duplicate store names become common.
func generateSlug(storeName string) string {
	base := slugNonAlnumRegexp.ReplaceAllString(strings.ToLower(storeName), "-")
	base = strings.Trim(base, "-")

	suffix := make([]byte, 3)
	_, _ = rand.Read(suffix)

	return base + "-" + hex.EncodeToString(suffix)
}

func toResponse(supplier *models.Supplier) *SupplierResponse {
	return &SupplierResponse{
		ID:        supplier.ID,
		UserID:    supplier.UserID,
		StoreName: supplier.StoreName,
		Slug:      supplier.SupplierSlug,
		Address:   supplier.Address,
		Status:    supplier.Status,
	}
}
