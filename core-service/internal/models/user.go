package models

import "time"

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

type User struct {
	ID           int    `gorm:"primaryKey"`
	FullName     string `gorm:"column:full_name;not null"`
	Username     string `gorm:"column:username;not null;unique"`
	PasswordHash string `gorm:"column:password_hash;not null"`
	Email        string `gorm:"column:email;not null;unique"`
	Status       string `gorm:"column:status;not null;default:active"`
	RoleID       int    `gorm:"column:role_id;not null"`
	Role         Role   `gorm:"foreignKey:RoleID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "users"
}
