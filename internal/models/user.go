package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID         int64     `json:"id"`
	ExternalID string    `json:"external_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateUserRequest represents the payload for creating a user
type CreateUserRequest struct {
	ExternalID string `json:"external_id"`
}

// UpdateUserRequest represents the payload for updating a user
type UpdateUserRequest struct {
	ExternalID string `json:"external_id,omitempty"`
}
