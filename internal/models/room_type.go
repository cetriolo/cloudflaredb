package models

import (
	"time"
)

// RoomType represents a room type in the system
type RoomType struct {
	ID        int64     `json:"id"`
	Size      string    `json:"size"`
	Style     string    `json:"style"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateRoomTypeRequest represents the payload for creating a room type
type CreateRoomTypeRequest struct {
	Size  string `json:"size"`
	Style string `json:"style"`
}

// UpdateRoomTypeRequest represents the payload for updating a room type
type UpdateRoomTypeRequest struct {
	Size  string `json:"size,omitempty"`
	Style string `json:"style,omitempty"`
}
