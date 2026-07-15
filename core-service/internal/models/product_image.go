package models

import "time"

type ProductImage struct {
	ID        int    `gorm:"primaryKey"`
	ProductID int    `gorm:"column:product_id;not null"`
	ImageURL  string `gorm:"column:image_url;not null"`
	CreatedAt time.Time
}

func (ProductImage) TableName() string {
	return "product_images"
}
