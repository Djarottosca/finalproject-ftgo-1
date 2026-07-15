package models

import "time"

type Category struct {
	ID           int    `gorm:"primaryKey"`
	CategoryName string `gorm:"column:category_name;not null;unique"`
	CategorySlug string `gorm:"column:category_slug;not null;unique"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Category) TableName() string {
	return "categories"
}
