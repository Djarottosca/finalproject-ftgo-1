package order

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
	// Courier and TrackingNumber are required when Status is "shipped";
	// ignored otherwise.
	Courier        string `json:"courier"`
	TrackingNumber string `json:"tracking_number"`
}

type OrderItemResponse struct {
	ProductID int     `json:"product_id"`
	Price     float64 `json:"price"`
	Qty       int     `json:"qty"`
	Subtotal  float64 `json:"subtotal"`
}

type OrderResponse struct {
	ID         int                 `json:"id"`
	UserID     int                 `json:"user_id"`
	TotalPrice float64             `json:"total_price"`
	Discount   float64             `json:"discount"`
	TotalItems int                 `json:"total_items"`
	FinalPrice float64             `json:"final_price"`
	Status     string              `json:"status"`
	Items      []OrderItemResponse `json:"items,omitempty"`
	Shipment   *ShipmentResponse   `json:"shipment,omitempty"`
}

type ShipmentResponse struct {
	Status         string `json:"status"`
	Courier        string `json:"courier,omitempty"`
	TrackingNumber string `json:"tracking_number,omitempty"`
}
