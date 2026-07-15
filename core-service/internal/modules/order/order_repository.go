package order

import (
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

// Repository defines the persistence contract for orders, so tests can
// substitute a mock instead of hitting Postgres.
type Repository interface {
	CreateWithItems(o *models.Order, items []models.OrderItem) error
	FindByID(id int) (*models.Order, error)
	ItemsByOrderID(orderID int) ([]models.OrderItem, error)
	ListByUserID(userID int) ([]models.Order, error)
	ListBySupplierID(supplierID int) ([]models.Order, error)
	UpdateStatus(o *models.Order) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns the GORM-backed Repository implementation.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// CreateWithItems inserts the order and its items in one transaction, so a
// partial order (order row without items) can never be persisted.
func (r *gormRepository) CreateWithItems(o *models.Order, items []models.OrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = o.ID
		}
		return tx.Create(&items).Error
	})
}

func (r *gormRepository) FindByID(id int) (*models.Order, error) {
	var o models.Order
	if err := r.db.First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *gormRepository) ItemsByOrderID(orderID int) ([]models.OrderItem, error) {
	var items []models.OrderItem
	if err := r.db.Where("order_id = ?", orderID).Order("id").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *gormRepository) ListByUserID(userID int) ([]models.Order, error) {
	var orders []models.Order
	if err := r.db.Where("user_id = ?", userID).Order("id desc").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// ListBySupplierID returns orders that contain at least one product sold by
// the given supplier, so a supplier can see orders touching their products.
func (r *gormRepository) ListBySupplierID(supplierID int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Distinct("orders.*").
		Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.supplier_id = ?", supplierID).
		Order("orders.id desc").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *gormRepository) UpdateStatus(o *models.Order) error {
	return r.db.Model(o).Update("status", o.Status).Error
}
