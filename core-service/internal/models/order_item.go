package models

import "time"

// Price is copied from products.price at order creation time, so order
// history doesn't change if the supplier changes the product price later.
type OrderItem struct {
	ID        int     `gorm:"primaryKey"`
	OrderID   int     `gorm:"column:order_id;not null"`
	ProductID int     `gorm:"column:product_id;not null"`
	Price     float64 `gorm:"column:price;not null"`
	Qty       int     `gorm:"column:qty;not null"`
	Subtotal  float64 `gorm:"column:subtotal;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (OrderItem) TableName() string {
	return "order_items"
}
