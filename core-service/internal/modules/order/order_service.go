package order

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/grpcclient"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/cart"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/shipping"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/user"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
)

var (
	ErrNotFound            = errors.New("order not found")
	ErrForbidden           = errors.New("order does not belong to this supplier")
	ErrEmptyCart           = errors.New("cart is empty")
	ErrInvalidStatus       = errors.New("status must be processing or shipped")
	ErrMissingShipmentInfo = errors.New("courier and tracking_number are required when status is shipped")
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
	repo         Repository
	cartRepo     cart.Repository
	userRepo     user.Repository
	shippingRepo shipping.Repository
	notifClient  *grpcclient.NotificationClient
}

// NewService returns the Service implementation backed by the given
// repositories. notifClient boleh nil (mis. di unit test) — kalau nil,
// pengiriman email dilewati tanpa error.
func NewService(repo Repository, cartRepo cart.Repository, userRepo user.Repository, shippingRepo shipping.Repository, notifClient *grpcclient.NotificationClient) Service {
	return &service{repo: repo, cartRepo: cartRepo, userRepo: userRepo, shippingRepo: shippingRepo, notifClient: notifClient}
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

	// Every order gets a pending shipment row up front; courier/tracking are
	// filled in later once the supplier actually ships it. Best-effort: a
	// failure here shouldn't roll back an already-placed order.
	var shipment *models.Shipment
	if s.shippingRepo != nil {
		shipment, err = s.shippingRepo.Create(ctx, o.ID)
		if err != nil {
			logger.Log.Warn().Err(err).Int("order_id", o.ID).Msg("gagal membuat shipment record untuk order")
		}
	}

	for _, ci := range cartItems {
		if _, err := s.cartRepo.Delete(ctx, userID, ci.ProductID); err != nil {
			return nil, err
		}
	}

	return toResponse(o, items, shipment), nil
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

	return toResponse(o, items, s.findShipment(ctx, o.ID)), nil
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
		res = append(res, *toResponse(&o, items, s.findShipment(ctx, o.ID)))
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
		res = append(res, *toResponse(&o, items, s.findShipment(ctx, o.ID)))
	}
	return res, nil
}

// UpdateStatus lets a supplier move an order to processing or shipped, once
// it contains at least one of their products. Moving to shipped also fills
// in the shipment's courier/tracking_number.
func (s *service) UpdateStatus(ctx context.Context, supplierID, orderID int, req UpdateOrderStatusRequest) (*OrderResponse, error) {
	if req.Status != models.OrderStatusProcessing && req.Status != models.OrderStatusShipped {
		return nil, ErrInvalidStatus
	}
	if req.Status == models.OrderStatusShipped && (req.Courier == "" || req.TrackingNumber == "") {
		return nil, ErrMissingShipmentInfo
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

	if req.Status == models.OrderStatusShipped && s.shippingRepo != nil {
		if err := s.shippingRepo.MarkShipped(ctx, o.ID, req.Courier, req.TrackingNumber); err != nil {
			logger.Log.Warn().Err(err).Int("order_id", o.ID).Msg("gagal update shipment record")
		}
	}

	items, err := s.repo.ItemsByOrderID(ctx, o.ID)
	if err != nil {
		return nil, err
	}

	s.notifyStatusChange(ctx, o)

	return toResponse(o, items, s.findShipment(ctx, o.ID)), nil
}

// findShipment is a best-effort lookup: a missing/failed shipment row
// shouldn't break rendering the order itself, so errors are swallowed and
// callers get nil (Shipment omitted from the response).
func (s *service) findShipment(ctx context.Context, orderID int) *models.Shipment {
	if s.shippingRepo == nil {
		return nil
	}
	shipment, err := s.shippingRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil
	}
	return shipment
}

// notifyStatusChange kirim email ke user pemilik order lewat notification-service.
// Sengaja fire-and-forget: gagal kirim email cuma di-log, gak bikin
// UpdateStatus ikut gagal — status order di DB sudah berhasil berubah.
func (s *service) notifyStatusChange(ctx context.Context, o *models.Order) {
	if s.notifClient == nil {
		return
	}

	u, err := s.userRepo.FindByID(ctx, o.UserID)
	if err != nil {
		logger.Log.Warn().Err(err).Int("order_id", o.ID).Msg("gagal ambil data user buat notifikasi email")
		return
	}

	subject := fmt.Sprintf("Update Pesanan #%d", o.ID)
	html := fmt.Sprintf(
		"<p>Halo %s,</p><p>Status pesanan <b>#%d</b> kamu sekarang: <b>%s</b>.</p>",
		u.FullName, o.ID, o.Status,
	)
	text := fmt.Sprintf("Halo %s, status pesanan #%d kamu sekarang: %s.", u.FullName, o.ID, o.Status)

	if _, err := s.notifClient.SendEmail(ctx, grpcclient.EmailInput{
		ToEmail:     u.Email,
		ToName:      u.FullName,
		Subject:     subject,
		HTMLContent: html,
		TextContent: text,
	}); err != nil {
		logger.Log.Warn().Err(err).Int("order_id", o.ID).Msg("gagal kirim email notifikasi status order")
	}
}
