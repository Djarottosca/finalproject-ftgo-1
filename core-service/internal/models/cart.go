package models

import "time"

// Cart has no own id; one row per (user, product), per migration 000008.
type Cart struct {
	UserID    int `gorm:"column:user_id;primaryKey"`
	ProductID int `gorm:"column:product_id;primaryKey"`
	Qty       int `gorm:"column:qty;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Cart) TableName() string {
	return "carts"
}
