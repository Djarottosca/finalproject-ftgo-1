package order

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(o *models.Order, items []models.OrderItem) *OrderResponse {
	itemRes := make([]OrderItemResponse, 0, len(items))
	for _, item := range items {
		itemRes = append(itemRes, OrderItemResponse{
			ProductID: item.ProductID,
			Price:     item.Price,
			Qty:       item.Qty,
			Subtotal:  item.Subtotal,
		})
	}

	return &OrderResponse{
		ID:         o.ID,
		UserID:     o.UserID,
		TotalPrice: o.TotalPrice,
		Discount:   o.Discount,
		TotalItems: o.TotalItems,
		FinalPrice: o.FinalPrice,
		Status:     o.Status,
		Items:      itemRes,
	}
}
