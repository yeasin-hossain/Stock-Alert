package dto

import (
	"time"
)

// UserResponse is the DTO used for API responses
type UserResponse struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserCreateRequest is the DTO for creating a new user
type UserCreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserUpdateRequest is the DTO for updating an existing user
type UserUpdateRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}
