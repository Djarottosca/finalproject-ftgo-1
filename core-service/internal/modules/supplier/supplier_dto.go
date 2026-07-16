package supplier

// RegisterRequest self-signs-up a new "supplier"-role account in one step —
// no separate user registration needed first.
type RegisterRequest struct {
	FullName  string `json:"full_name" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
	Email     string `json:"email" validate:"required,email"`
	StoreName string `json:"store_name" validate:"required"`
	Address   string `json:"address" validate:"required"`
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

// RegisterResponse carries a login token alongside the created supplier, so
// the caller is immediately authenticated — same pattern as auth.Register.
type RegisterResponse struct {
	Token    string           `json:"token"`
	Supplier SupplierResponse `json:"supplier"`
}
