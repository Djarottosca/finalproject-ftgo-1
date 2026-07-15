package models

import "time"

const (
	ProductStatusActive   = "active"
	ProductStatusInactive = "inactive"

	ProductDiscountPercentage = "percentage"
	ProductDiscountFixed      = "fixed"
)

// ponytail: price/discount use float64, not a decimal type. No decimal lib
// in go.mod and this is portfolio scope; upgrade to shopspring/decimal if
// rounding errors on money ever become a real bug.
type Product struct {
	ID             int      `gorm:"primaryKey"`
	ProductName    string   `gorm:"column:product_name;not null"`
	ProductSlug    string   `gorm:"column:product_slug;not null;unique"`
	CategoryID     int      `gorm:"column:category_id;not null"`
	Unit           string   `gorm:"column:unit;not null"`
	Stock          int      `gorm:"column:stock;not null;default:0"`
	SupplierID     int      `gorm:"column:supplier_id;not null"`
	Price          float64  `gorm:"column:price;not null"`
	Description    *string  `gorm:"column:description"`
	DiscountType   *string  `gorm:"column:discount_type"`
	DiscountAmount *float64 `gorm:"column:discount_amount"`
	Status         string   `gorm:"column:status;not null;default:active"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Product) TableName() string {
	return "products"
}
