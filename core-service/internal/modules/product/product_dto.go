package product

// GET /products.
type ListRequest struct {
	Keyword    string `query:"keyword"`
	CategoryID int    `query:"category_id"`
	Page       int    `query:"page" validate:"omitempty,min=1"`
	Limit      int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

// field-nya kosong (0)
func (r *ListRequest) Normalize() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
}

type ProductResponse struct {
	ID             int      `json:"id"`
	ProductName    string   `json:"product_name"`
	ProductSlug    string   `json:"product_slug"`
	CategoryID     int      `json:"category_id"`
	CategoryName   string   `json:"category_name"`
	Unit           string   `json:"unit"`
	Stock          int      `json:"stock"`
	Price          float64  `json:"price"`
	DiscountType   *string  `json:"discount_type,omitempty"`
	DiscountAmount *float64 `json:"discount_amount,omitempty"`
	FinalPrice     float64  `json:"final_price"`
	Status         string   `json:"status"`
}

type ProductDetailResponse struct {
	ProductResponse
	Description string   `json:"description"`
	SupplierID  int      `json:"supplier_id"`
	Images      []string `json:"images"`
}

type ListResponse struct {
	Items      []ProductResponse `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalItems int64             `json:"total_items"`
}

// POST /supplier/products
type CreateProductRequest struct {
	ProductName string  `json:"product_name" validate:"required"`
	CategoryID  int     `json:"category_id" validate:"required"`
	Unit        string  `json:"unit" validate:"required"`
	Stock       int     `json:"stock" validate:"min=0"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Description string  `json:"description"`
}

// PUT /supplier/products/:id
type UpdateProductRequest struct {
	ProductName string  `json:"product_name" validate:"required"`
	CategoryID  int     `json:"category_id" validate:"required"`
	Unit        string  `json:"unit" validate:"required"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Description string  `json:"description"`
}

// PATCH /supplier/products/:id/discount
type DiscountRequest struct {
	DiscountType   string  `json:"discount_type"` // "percentage", "fixed", or "" to clear
	DiscountAmount float64 `json:"discount_amount"`
}

// PATCH /supplier/products/:id/stock
type StockAdjustRequest struct {
	Delta int `json:"delta"` // positive to add stock, negative to deduct
}
