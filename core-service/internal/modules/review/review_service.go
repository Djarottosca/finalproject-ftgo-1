package review

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/product"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrAlreadyReviewed = errors.New("you have already reviewed this product")
)

// Service defines the review use cases exposed to the handler layer.
type Service interface {
	Create(ctx context.Context, userID, productID int, req CreateReviewRequest) (*ReviewResponse, error)
	ListByProduct(ctx context.Context, productID int) ([]ReviewResponse, error)
}

type service struct {
	repo        Repository
	productRepo product.Repository
}

// NewService returns the Service implementation backed by the given repositories.
func NewService(repo Repository, productRepo product.Repository) Service {
	return &service{repo: repo, productRepo: productRepo}
}

// Create adds a review for a product. One review per user per product,
// enforced here (friendly error) and by a unique index in the DB (source
// of truth under concurrent requests).
func (s *service) Create(ctx context.Context, userID, productID int, req CreateReviewRequest) (*ReviewResponse, error) {
	if _, err := s.productRepo.FindByID(ctx, productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if _, err := s.repo.FindByUserAndProduct(ctx, userID, productID); err == nil {
		return nil, ErrAlreadyReviewed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	review := &models.Review{
		UserID:    userID,
		ProductID: productID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}
	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	return toResponse(review), nil
}

func (s *service) ListByProduct(ctx context.Context, productID int) ([]ReviewResponse, error) {
	reviews, err := s.repo.ListByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	res := make([]ReviewResponse, 0, len(reviews))
	for _, r := range reviews {
		res = append(res, *toResponse(&r))
	}
	return res, nil
}
