package models

import "time"

//tabel `products`
type Product struct {
	ID             uint64         `gorm:"column:id;primaryKey"`
	ProductName    string         `gorm:"column:product_name"`
	ProductSlug    string         `gorm:"column:product_slug"`
	CategoryID     uint64         `gorm:"column:category_id"`
	Category       Category       `gorm:"foreignKey:CategoryID;references:ID"`
	Unit           string         `gorm:"column:unit"`
	Stock          int            `gorm:"column:stock"`
	SupplierID     uint64         `gorm:"column:supplier_id"`
	Price          float64        `gorm:"column:price"`
	Description    string         `gorm:"column:description"`
	DiscountType   *string        `gorm:"column:discount_type"`
	DiscountAmount *float64       `gorm:"column:discount_amount"`
	Status         string         `gorm:"column:status"`
	Images         []ProductImage `gorm:"foreignKey:ProductID;references:ID"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (Product) TableName() string {
	return "products"
}

// FinalPrice menghitung harga setelah diskon
func (p Product) FinalPrice() float64 {
	if p.DiscountType == nil || p.DiscountAmount == nil {
		return p.Price
	}

	switch *p.DiscountType {
	case "percentage":
		discount := p.Price * (*p.DiscountAmount / 100)
		return p.Price - discount
	case "fixed":
		final := p.Price - *p.DiscountAmount
		if final < 0 {
			return 0
		}
		return final
	default:
		return p.Price
	}
}
