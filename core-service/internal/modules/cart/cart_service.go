package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrProductInactive   = errors.New("product is not available")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrItemNotFound      = errors.New("cart item not found")
)

// productLookup only needs to check the product still exists and its stock,
// so it doesn't depend on the full product.Repository interface.
type productLookup interface {
	FindByID(ctx context.Context, id int) (*models.Product, error)
}

// Service defines the cart use cases exposed to the handler layer.
type Service interface {
	AddItem(ctx context.Context, userID int, req AddItemRequest) (*CartResponse, error)
	UpdateItem(ctx context.Context, userID int, productID int, req UpdateItemRequest) (*CartResponse, error)
	RemoveItem(ctx context.Context, userID int, productID int) (*CartResponse, error)
	GetCart(ctx context.Context, userID int) (*CartResponse, error)
}

type service struct {
	repo     Repository
	products productLookup
}

// NewService returns the Service implementation.
func NewService(repo Repository, products productLookup) Service {
	return &service{repo: repo, products: products}
}

func (s *service) AddItem(ctx context.Context, userID int, req AddItemRequest) (*CartResponse, error) {
	product, err := s.getSellableProduct(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if product.Stock < req.Qty {
		return nil, ErrInsufficientStock
	}

	if err := s.repo.Upsert(ctx, userID, req.ProductID, req.Qty); err != nil {
		return nil, fmt.Errorf("add item to cart: %w", err)
	}
	return s.GetCart(ctx, userID)
}

func (s *service) UpdateItem(ctx context.Context, userID int, productID int, req UpdateItemRequest) (*CartResponse, error) {
	product, err := s.getSellableProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product.Stock < req.Qty {
		return nil, ErrInsufficientStock
	}

	rows, err := s.repo.UpdateQty(ctx, userID, productID, req.Qty)
	if err != nil {
		return nil, fmt.Errorf("update qty: %w", err)
	}
	if rows == 0 {
		return nil, ErrItemNotFound
	}
	return s.GetCart(ctx, userID)
}

func (s *service) RemoveItem(ctx context.Context, userID int, productID int) (*CartResponse, error) {
	rows, err := s.repo.Delete(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("delete item: %w", err)
	}
	if rows == 0 {
		return nil, ErrItemNotFound
	}
	return s.GetCart(ctx, userID)
}

func (s *service) GetCart(ctx context.Context, userID int) (*CartResponse, error) {
	carts, err := s.repo.FindAllByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	items := make([]CartItemResponse, 0, len(carts))
	var totalItems int
	var totalPrice float64
	for _, c := range carts {
		item := toCartItemResponse(c)
		items = append(items, item)
		totalItems += item.Qty
		totalPrice += item.Subtotal
	}

	return &CartResponse{
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
	}, nil
}

// getSellableProduct returns the product if it exists and is active.
func (s *service) getSellableProduct(ctx context.Context, productID int) (*models.Product, error) {
	product, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return nil, ErrProductNotFound
	}
	if product.Status != "active" {
		return nil, ErrProductInactive
	}
	return product, nil
}
