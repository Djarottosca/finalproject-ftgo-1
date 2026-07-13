package models

import "time"

type Role struct {
	ID        int    `gorm:"primaryKey"`
	RoleName  string `gorm:"column:role_name;not null;unique"`
	RoleSlug  string `gorm:"column:role_slug;not null;unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Role) TableName() string {
	return "roles"
}
