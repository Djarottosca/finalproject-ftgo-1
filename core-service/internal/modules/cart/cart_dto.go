package cart

// POST /cart/items
type AddItemRequest struct {
	ProductID int `json:"product_id" validate:"required"`
	Qty       int `json:"qty" validate:"required,min=1"`
}

// PUT /cart/items/:product_id
type UpdateItemRequest struct {
	Qty int `json:"qty" validate:"required,min=1"`
}

type CartItemResponse struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSlug string  `json:"product_slug"`
	Price       float64 `json:"price"`
	FinalPrice  float64 `json:"final_price"`
	Qty         int     `json:"qty"`
	Subtotal    float64 `json:"subtotal"`
	Stock       int     `json:"stock"`
}

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice float64            `json:"total_price"`
}
