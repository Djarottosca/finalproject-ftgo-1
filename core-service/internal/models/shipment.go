package models

import "time"

const (
	ShipmentStatusPending   = "pending"
	ShipmentStatusShipped   = "shipped"
	ShipmentStatusDelivered = "delivered"
)

type Shipment struct {
	ID                 int      `gorm:"primaryKey"`
	OrderID            int      `gorm:"column:order_id;not null;unique"`
	ShippingID         *string  `gorm:"column:shipping_id"`
	Status             string   `gorm:"column:status;not null;default:pending"`
	TrackingNumber     *string  `gorm:"column:tracking_number"`
	Courier            *string  `gorm:"column:courier"`
	ActualShippingCost *float64 `gorm:"column:actual_shipping_cost"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (Shipment) TableName() string {
	return "shipments"
}
