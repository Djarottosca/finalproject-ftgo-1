package user

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(u *models.User) *UserResponse {
	return &UserResponse{
		ID:       u.ID,
		FullName: u.FullName,
		Username: u.Username,
		Email:    u.Email,
		Status:   u.Status,
		RoleID:   u.RoleID,
	}
}
