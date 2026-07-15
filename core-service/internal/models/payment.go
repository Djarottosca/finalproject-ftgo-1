package models

import "time"

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"
	PaymentStatusExpired = "expired"
)

// ponytail: amount/tax/shipping use float64 to mirror the other money models
// (order.go, product.go). The DB column is numeric(12,2); float64 is the
// portfolio-scope tradeoff, upgrade to shopspring/decimal if rounding ever bites.
type Payment struct {
	ID                   int      `gorm:"primaryKey"`
	OrderID              int      `gorm:"column:order_id;not null;unique"`
	Amount               float64  `gorm:"column:amount;not null"`
	Tax                  float64  `gorm:"column:tax;not null;default:0"`
	Status               string   `gorm:"column:status;not null;default:pending"`
	PaymentLink          *string  `gorm:"column:payment_link"`
	PaymentReference     *string  `gorm:"column:payment_reference"`
	ShippingCostEstimate *float64 `gorm:"column:shipping_cost_estimate"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (Payment) TableName() string {
	return "payments"
}
