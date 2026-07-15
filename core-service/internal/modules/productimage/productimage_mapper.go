package productimage

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(image *models.ProductImage) *ImageResponse {
	return &ImageResponse{
		ID:        image.ID,
		ProductID: image.ProductID,
		ImageURL:  image.ImageURL,
	}
}
