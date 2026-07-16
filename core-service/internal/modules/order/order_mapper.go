package order

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(o *models.Order, items []models.OrderItem, shipment *models.Shipment) *OrderResponse {
	itemRes := make([]OrderItemResponse, 0, len(items))
	for _, item := range items {
		itemRes = append(itemRes, OrderItemResponse{
			ProductID: item.ProductID,
			Price:     item.Price,
			Qty:       item.Qty,
			Subtotal:  item.Subtotal,
		})
	}

	res := &OrderResponse{
		ID:         o.ID,
		UserID:     o.UserID,
		TotalPrice: o.TotalPrice,
		Discount:   o.Discount,
		TotalItems: o.TotalItems,
		FinalPrice: o.FinalPrice,
		Status:     o.Status,
		Items:      itemRes,
	}

	if shipment != nil {
		sr := &ShipmentResponse{Status: shipment.Status}
		if shipment.Courier != nil {
			sr.Courier = *shipment.Courier
		}
		if shipment.TrackingNumber != nil {
			sr.TrackingNumber = *shipment.TrackingNumber
		}
		res.Shipment = sr
	}

	return res
}
