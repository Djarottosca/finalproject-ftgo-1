package models

import "time"

// ProductImage merepresentasikan tabel `product_images`.
type ProductImage struct {
	ID        int       `gorm:"column:id;primaryKey"`
	ProductID int       `gorm:"column:product_id"`
	ImageURL  string    `gorm:"column:image_url"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (ProductImage) TableName() string {
	return "product_images"
}
