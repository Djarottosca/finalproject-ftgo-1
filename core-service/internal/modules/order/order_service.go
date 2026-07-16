package order

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/cart"
)

var (
	ErrNotFound      = errors.New("order not found")
	ErrForbidden     = errors.New("order does not belong to this supplier")
	ErrEmptyCart     = errors.New("cart is empty")
	ErrInvalidStatus = errors.New("status must be processing or shipped")
)

// Service defines the order use cases exposed to the handler layer.
type Service interface {
	Checkout(ctx context.Context, userID int) (*OrderResponse, error)
	Get(ctx context.Context, id int) (*OrderResponse, error)
	ListMine(ctx context.Context, userID int) ([]OrderResponse, error)
	ListForSupplier(ctx context.Context, supplierID int) ([]OrderResponse, error)
	UpdateStatus(ctx context.Context, supplierID, orderID int, req UpdateOrderStatusRequest) (*OrderResponse, error)
}

type service struct {
	repo     Repository
	cartRepo cart.Repository
}

// NewService returns the Service implementation backed by the given repositories.
func NewService(repo Repository, cartRepo cart.Repository) Service {
	return &service{repo: repo, cartRepo: cartRepo}
}

// Checkout converts the user's cart into an order + order items, copying
// the current product price so later price changes don't affect this order.
// cartRepo.FindAllByUser already preloads Product, so no separate product
// lookup is needed here.
func (s *service) Checkout(ctx context.Context, userID int) (*OrderResponse, error) {
	cartItems, err := s.cartRepo.FindAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, ErrEmptyCart
	}

	items := make([]models.OrderItem, 0, len(cartItems))
	var totalPrice float64
	var totalItems int
	for _, ci := range cartItems {
		subtotal := ci.Product.Price * float64(ci.Qty)
		items = append(items, models.OrderItem{
			ProductID: ci.Product.ID,
			Price:     ci.Product.Price,
			Qty:       ci.Qty,
			Subtotal:  subtotal,
		})
		totalPrice += subtotal
		totalItems += ci.Qty
	}

	o := &models.Order{
		UserID:     userID,
		TotalPrice: totalPrice,
		Discount:   0,
		TotalItems: totalItems,
		FinalPrice: totalPrice,
		Status:     models.OrderStatusPending,
	}
	if err := s.repo.CreateWithItems(ctx, o, items); err != nil {
		return nil, err
	}

	for _, ci := range cartItems {
		if _, err := s.cartRepo.Delete(ctx, userID, ci.ProductID); err != nil {
			return nil, err
		}
	}

	return toResponse(o, items), nil
}

func (s *service) Get(ctx context.Context, id int) (*OrderResponse, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	items, err := s.repo.ItemsByOrderID(ctx, o.ID)
	if err != nil {
		return nil, err
	}

	return toResponse(o, items), nil
}

func (s *service) ListMine(ctx context.Context, userID int) ([]OrderResponse, error) {
	orders, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		items, err := s.repo.ItemsByOrderID(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, *toResponse(&o, items))
	}
	return res, nil
}

func (s *service) ListForSupplier(ctx context.Context, supplierID int) ([]OrderResponse, error) {
	orders, err := s.repo.ListBySupplierID(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	res := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		items, err := s.repo.ItemsByOrderID(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, *toResponse(&o, items))
	}
	return res, nil
}

// UpdateStatus lets a supplier move an order to processing or shipped, once
// it contains at least one of their products.
func (s *service) UpdateStatus(ctx context.Context, supplierID, orderID int, req UpdateOrderStatusRequest) (*OrderResponse, error) {
	if req.Status != models.OrderStatusProcessing && req.Status != models.OrderStatusShipped {
		return nil, ErrInvalidStatus
	}

	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	supplierOrders, err := s.repo.ListBySupplierID(ctx, supplierID)
	if err != nil {
		return nil, err
	}
	owns := false
	for _, so := range supplierOrders {
		if so.ID == orderID {
			owns = true
			break
		}
	}
	if !owns {
		return nil, ErrForbidden
	}

	o.Status = req.Status
	if err := s.repo.UpdateStatus(ctx, o); err != nil {
		return nil, err
	}

	items, err := s.repo.ItemsByOrderID(ctx, o.ID)
	if err != nil {
		return nil, err
	}

	return toResponse(o, items), nil
}
