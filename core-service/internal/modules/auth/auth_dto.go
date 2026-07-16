package auth

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// RegisterRequest signs up a new "user"-role account via POST /user/register.
// Suppliers use POST /supplier/register instead (supplier module, creates
// user+supplier together); admins are seeded.
type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email" validate:"required,email"`
}
