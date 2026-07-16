package review

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(r *models.Review) *ReviewResponse {
	return &ReviewResponse{
		ID:        r.ID,
		UserID:    r.UserID,
		ProductID: r.ProductID,
		Rating:    r.Rating,
		Comment:   r.Comment,
	}
}
