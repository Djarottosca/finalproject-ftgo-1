package models

import "time"

const (
	SupplierStatusPending  = "pending"
	SupplierStatusApproved = "approved"
	SupplierStatusRejected = "rejected"
)

type Supplier struct {
	ID           int    `gorm:"primaryKey"`
	UserID       int    `gorm:"column:user_id;not null;unique"`
	User         User   `gorm:"foreignKey:UserID"`
	StoreName    string `gorm:"column:store_name;not null"`
	SupplierSlug string `gorm:"column:supplier_slug;not null;unique"`
	Address      string `gorm:"column:address;not null"`
	Status       string `gorm:"column:status;not null;default:pending"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Supplier) TableName() string {
	return "suppliers"
}
