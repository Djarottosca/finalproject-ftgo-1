package review

// POST /products/:id/reviews
type CreateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

type ReviewResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	ProductID int    `json:"product_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}
