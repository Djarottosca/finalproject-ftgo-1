package supplier

type RegisterRequest struct {
	StoreName string `json:"store_name"`
	Address   string `json:"address"`
}

type ReviewRequest struct {
	Status string `json:"status"` // "approved" or "rejected"
}

type SupplierResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	StoreName string `json:"store_name"`
	Slug      string `json:"slug"`
	Address   string `json:"address"`
	Status    string `json:"status"`
}
