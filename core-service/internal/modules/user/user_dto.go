package user

type CreateUserRequest struct {
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	RoleID   int    `json:"role_id"`
}

// UpdateUserRequest is admin-only: includes Status (activate/ban).
type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

// UpdateProfileRequest is the self-service variant: no Status field, so a
// user can't reactivate/ban their own account.
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
	RoleID   int    `json:"role_id"`
}
