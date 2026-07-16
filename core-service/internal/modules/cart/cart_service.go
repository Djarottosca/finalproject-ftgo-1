package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var (
	ErrProductNotFound   = errors.New("produk tidak ditemukan")
	ErrProductInactive   = errors.New("produk sedang tidak tersedia")
	ErrInsufficientStock = errors.New("stok tidak mencukupi")
	ErrItemNotFound      = errors.New("item keranjang tidak ditemukan")
)

// product.Repository — cukup butuh cek produk masih ada & berapa stoknya.

type productLookup interface {
	FindByID(ctx context.Context, id uint64) (*models.Product, error)
}

// Service defines the cart use cases exposed to the handler layer.
type Service interface {
	AddItem(ctx context.Context, userID int, req AddItemRequest) (*CartResponse, error)
	UpdateItem(ctx context.Context, userID int, productID uint64, req UpdateItemRequest) (*CartResponse, error)
	RemoveItem(ctx context.Context, userID int, productID uint64) (*CartResponse, error)
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
		return nil, fmt.Errorf("gagal menambah item ke keranjang: %w", err)
	}
	return s.GetCart(ctx, userID)
}

func (s *service) UpdateItem(ctx context.Context, userID int, productID uint64, req UpdateItemRequest) (*CartResponse, error) {
	product, err := s.getSellableProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product.Stock < req.Qty {
		return nil, ErrInsufficientStock
	}

	rows, err := s.repo.UpdateQty(ctx, userID, productID, req.Qty)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah qty: %w", err)
	}
	if rows == 0 {
		return nil, ErrItemNotFound
	}
	return s.GetCart(ctx, userID)
}

func (s *service) RemoveItem(ctx context.Context, userID int, productID uint64) (*CartResponse, error) {
	rows, err := s.repo.Delete(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("gagal menghapus item: %w", err)
	}
	if rows == 0 {
		return nil, ErrItemNotFound
	}
	return s.GetCart(ctx, userID)
}

func (s *service) GetCart(ctx context.Context, userID int) (*CartResponse, error) {
	carts, err := s.repo.FindAllByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil keranjang: %w", err)
	}

	items := make([]CartItemResponse, 0, len(carts))
	var totalItems int
	var totalPrice float64
	for _, c := range carts {
		finalPrice := c.Product.FinalPrice()
		subtotal := finalPrice * float64(c.Qty)

		items = append(items, CartItemResponse{
			ProductID:   c.ProductID,
			ProductName: c.Product.ProductName,
			ProductSlug: c.Product.ProductSlug,
			Price:       c.Product.Price,
			FinalPrice:  finalPrice,
			Qty:         c.Qty,
			Subtotal:    subtotal,
			Stock:       c.Product.Stock,
		})
		totalItems += c.Qty
		totalPrice += subtotal
	}

	return &CartResponse{
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
	}, nil
}

// getSellableProduct: produk ada, statusnya "active"
func (s *service) getSellableProduct(ctx context.Context, productID uint64) (*models.Product, error) {
	product, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return nil, ErrProductNotFound
	}
	if product.Status != "active" {
		return nil, ErrProductInactive
	}
	return product, nil
}
