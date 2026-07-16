package cart

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toCartItemResponse(c models.Cart) CartItemResponse {
	finalPrice := c.Product.FinalPrice()
	return CartItemResponse{
		ProductID:   c.ProductID,
		ProductName: c.Product.ProductName,
		ProductSlug: c.Product.ProductSlug,
		Price:       c.Product.Price,
		FinalPrice:  finalPrice,
		Qty:         c.Qty,
		Subtotal:    finalPrice * float64(c.Qty),
		Stock:       c.Product.Stock,
	}
}
