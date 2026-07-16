package models

import "time"

// tabel `reviews`. Satu user cuma bisa review satu produk sekali
// (unique index user_id+product_id di migration 000013).
type Review struct {
	ID        int       `gorm:"primaryKey"`
	UserID    int       `gorm:"column:user_id;not null"`
	ProductID int       `gorm:"column:product_id;not null"`
	Rating    int       `gorm:"column:rating;not null"`
	Comment   string    `gorm:"column:comment"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Review) TableName() string {
	return "reviews"
}
