package cart

type AddToCartRequest struct {
	ProductID int `json:"product_id"`
	Qty       int `json:"qty"`
}

type UpdateCartRequest struct {
	Qty int `json:"qty"`
}

type CartItemResponse struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Qty         int     `json:"qty"`
	Subtotal    float64 `json:"subtotal"`
}
