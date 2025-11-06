package domain

import "time"

// Role represents the user's role in the system
type Role string

const (
	// RoleEmployee represents an employee user
	RoleEmployee Role = "EMPLOYEE"
	// RoleManager represents a manager user
	RoleManager Role = "MANAGER"
)

// User represents a user in the time tracking system
type User struct {
	ID        uint      `json:"id" validate:"omitempty"`
	Username  string    `json:"username" validate:"required,min=3,max=50"`
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"-"`
	Role      Role      `json:"role" validate:"required,oneof=EMPLOYEE MANAGER"`
	CreatedAt time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt time.Time `json:"updated_at" validate:"omitempty"`
}
