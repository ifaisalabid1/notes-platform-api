package admin

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"
)

type Admin struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Role        Role       `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AdminWithPassword struct {
	Admin
	PasswordHash string
}

type BootstrapOwnerInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAdminInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type UpdateAdminStatusInput struct {
	IsActive bool `json:"is_active"`
}
