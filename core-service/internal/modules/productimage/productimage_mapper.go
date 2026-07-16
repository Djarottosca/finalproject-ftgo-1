package productimage

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(image *models.ProductImage) *ImageResponse {
	return &ImageResponse{
		ID:        int(image.ID),
		ProductID: int(image.ProductID),
		ImageURL:  image.ImageURL,
	}
}
