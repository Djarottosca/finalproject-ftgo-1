package product

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toProductResponse(p models.Product) ProductResponse {
	return ProductResponse{
		ID:             p.ID,
		ProductName:    p.ProductName,
		ProductSlug:    p.ProductSlug,
		CategoryID:     p.CategoryID,
		CategoryName:   p.Category.CategoryName,
		Unit:           p.Unit,
		Stock:          p.Stock,
		Price:          p.Price,
		DiscountType:   p.DiscountType,
		DiscountAmount: p.DiscountAmount,
		FinalPrice:     p.FinalPrice(),
		Status:         p.Status,
	}
}

func toProductDetailResponse(p models.Product) ProductDetailResponse {
	images := make([]string, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, img.ImageURL)
	}

	return ProductDetailResponse{
		ProductResponse: toProductResponse(p),
		Description:     p.Description,
		SupplierID:      p.SupplierID,
		Images:          images,
	}
}
