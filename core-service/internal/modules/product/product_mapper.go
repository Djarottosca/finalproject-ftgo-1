package product

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(product *models.Product) *ProductResponse {
	description := ""
	if product.Description != nil {
		description = *product.Description
	}

	return &ProductResponse{
		ID:             product.ID,
		ProductName:    product.ProductName,
		Slug:           product.ProductSlug,
		CategoryID:     product.CategoryID,
		Unit:           product.Unit,
		Stock:          product.Stock,
		SupplierID:     product.SupplierID,
		Price:          product.Price,
		Description:    description,
		DiscountType:   product.DiscountType,
		DiscountAmount: product.DiscountAmount,
		Status:         product.Status,
	}
}
