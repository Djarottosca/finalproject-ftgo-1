package supplier

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/slug"
)

var (
	ErrNotFound      = errors.New("supplier not found")
	ErrAlreadyExists = errors.New("user already registered as supplier")
	ErrInvalidStatus = errors.New("status must be approved or rejected")
)

// Service defines the supplier use cases exposed to the handler layer.
type Service interface {
	Register(ctx context.Context, userID int, req RegisterRequest) (*SupplierResponse, error)
	Get(ctx context.Context, id int) (*SupplierResponse, error)
	List(ctx context.Context, status string) ([]SupplierResponse, error)
	Review(ctx context.Context, id int, req ReviewRequest) (*SupplierResponse, error)
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, userID int, req RegisterRequest) (*SupplierResponse, error) {
	if _, err := s.repo.FindByUserID(ctx, userID); err == nil {
		return nil, ErrAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	supplier := &models.Supplier{
		UserID:       userID,
		StoreName:    req.StoreName,
		SupplierSlug: slug.Generate(req.StoreName),
		Address:      req.Address,
		Status:       models.SupplierStatusPending,
	}
	if err := s.repo.Create(ctx, supplier); err != nil {
		return nil, err
	}

	return toResponse(supplier), nil
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
