package models

import "time"

type Cart struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
	ProductID int       `gorm:"column:product_id;primaryKey"`
	Product   Product   `gorm:"foreignKey:ProductID;references:ID"`
	Qty       int       `gorm:"column:qty"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Cart) TableName() string {
	return "carts"
}
