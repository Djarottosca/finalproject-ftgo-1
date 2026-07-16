package models

import "time"

// Role slugs — stable keys used in JWT claims and RBAC middleware, matching
// seeded rows in migration 000001.
const (
	RoleAdmin    = "admin"
	RoleSupplier = "supplier"
	RoleUser     = "user"
)

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
