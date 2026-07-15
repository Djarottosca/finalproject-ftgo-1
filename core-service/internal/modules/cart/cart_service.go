package cart

import (
	"errors"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/product"
)

var (
	ErrNotFound        = errors.New("cart item not found")
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidQty      = errors.New("qty must be greater than zero")
)

// Service defines the cart use cases exposed to the handler layer.
type Service interface {
	Add(userID int, req AddToCartRequest) error
	List(userID int) ([]CartItemResponse, error)
	Update(userID, productID int, req UpdateCartRequest) error
	Delete(userID, productID int) error
}

type service struct {
	repo        Repository
	productRepo product.Repository
}

// NewService returns the Service implementation backed by the given repositories.
func NewService(repo Repository, productRepo product.Repository) Service {
	return &service{repo: repo, productRepo: productRepo}
}

func (s *service) Add(userID int, req AddToCartRequest) error {
	if req.Qty <= 0 {
		return ErrInvalidQty
	}

	if _, err := s.productRepo.FindByID(req.ProductID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	qty := req.Qty
	existing, err := s.repo.FindByUserAndProduct(userID, req.ProductID)
	if err == nil {
		qty += existing.Qty
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.repo.Upsert(&models.Cart{
		UserID:    userID,
		ProductID: req.ProductID,
		Qty:       qty,
	})
}

func (s *service) List(userID int) ([]CartItemResponse, error) {
	items, err := s.repo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	res := make([]CartItemResponse, 0, len(items))
	for _, item := range items {
		prod, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			return nil, err
		}
		res = append(res, CartItemResponse{
			ProductID:   prod.ID,
			ProductName: prod.ProductName,
			Price:       prod.Price,
			Qty:         item.Qty,
			Subtotal:    prod.Price * float64(item.Qty),
		})
	}
	return res, nil
}

func (s *service) Update(userID, productID int, req UpdateCartRequest) error {
	if req.Qty <= 0 {
		return ErrInvalidQty
	}

	if _, err := s.repo.FindByUserAndProduct(userID, productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	return s.repo.Upsert(&models.Cart{
		UserID:    userID,
		ProductID: productID,
		Qty:       req.Qty,
	})
}

func (s *service) Delete(userID, productID int) error {
	if _, err := s.repo.FindByUserAndProduct(userID, productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.Delete(userID, productID)
}
