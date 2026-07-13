package product

type CreateProductRequest struct {
	ProductName string  `json:"product_name"`
	CategoryID  int     `json:"category_id"`
	Unit        string  `json:"unit"`
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type UpdateProductRequest struct {
	ProductName string  `json:"product_name"`
	CategoryID  int     `json:"category_id"`
	Unit        string  `json:"unit"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type DiscountRequest struct {
	DiscountType   string  `json:"discount_type"` // "percentage", "fixed", or "" to clear
	DiscountAmount float64 `json:"discount_amount"`
}

type StockAdjustRequest struct {
	Delta int `json:"delta"` // positive to add stock, negative to deduct
}

type ListFilter struct {
	CategoryID int
	SupplierID int
	Status     string
}

type ProductResponse struct {
	ID             int      `json:"id"`
	ProductName    string   `json:"product_name"`
	Slug           string   `json:"slug"`
	CategoryID     int      `json:"category_id"`
	Unit           string   `json:"unit"`
	Stock          int      `json:"stock"`
	SupplierID     int      `json:"supplier_id"`
	Price          float64  `json:"price"`
	Description    string   `json:"description"`
	DiscountType   *string  `json:"discount_type,omitempty"`
	DiscountAmount *float64 `json:"discount_amount,omitempty"`
	Status         string   `json:"status"`
}
