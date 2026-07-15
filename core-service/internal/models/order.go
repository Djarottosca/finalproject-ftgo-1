package models

import "time"

const (
	OrderStatusPending    = "pending"
	OrderStatusPaid       = "paid"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusCompleted  = "completed"
	OrderStatusCancelled  = "cancelled"
)

// ponytail: money uses float64, not a decimal type. See models/product.go for
// the same tradeoff and upgrade path.
type Order struct {
	ID         int     `gorm:"primaryKey"`
	UserID     int     `gorm:"column:user_id;not null"`
	TotalPrice float64 `gorm:"column:total_price;not null"`
	Discount   float64 `gorm:"column:discount;not null;default:0"`
	TotalItems int     `gorm:"column:total_items;not null"`
	FinalPrice float64 `gorm:"column:final_price;not null"`
	Status     string  `gorm:"column:status;not null;default:pending"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Order) TableName() string {
	return "orders"
}
