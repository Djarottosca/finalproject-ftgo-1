package models

import "time"

type Address struct {
	ID          int    `gorm:"primaryKey"`
	UserID      int    `gorm:"column:user_id;not null"`
	Label       string `gorm:"column:label;not null"`
	FullAddress string `gorm:"column:full_address;not null"`
	City        string `gorm:"column:city;not null"`
	District    string `gorm:"column:district;not null"`
	PostalCode  string `gorm:"column:postal_code;not null"`
	IsPrimary   bool   `gorm:"column:is_primary;not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Address) TableName() string {
	return "addresses"
}
