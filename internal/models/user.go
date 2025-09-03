package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system, integrated with Keycloak authentication
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	KeycloakID  string    `json:"keycloak_id" db:"keycloak_id"`
	Email       string    `json:"email" db:"email"`
	FirstName   *string   `json:"first_name" db:"first_name"`
	LastName    *string   `json:"last_name" db:"last_name"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	KeycloakID string  `json:"keycloak_id" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Email     *string `json:"email" validate:"omitempty,email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	IsActive  *bool   `json:"is_active"`
}

// UserResponse represents the response payload for user data (excludes sensitive info)
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
