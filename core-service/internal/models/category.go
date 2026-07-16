package models

import "time"

// tabel `categories`.
type Category struct {
	ID           int       `gorm:"column:id;primaryKey"`
	CategoryName string    `gorm:"column:category_name"`
	CategorySlug string    `gorm:"column:category_slug"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Category) TableName() string {
	return "categories"
}
