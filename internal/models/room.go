package models

import (
	"time"
)

// Room represents a room in the system
type Room struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRoomRequest represents the payload for creating a room
type CreateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

// UpdateRoomRequest represents the payload for updating a room
type UpdateRoomRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Capacity    int    `json:"capacity,omitempty"`
}

// UserRoom represents the many-to-many relationship between users and rooms
type UserRoom struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	RoomID    int64     `json:"room_id"`
	CreatedAt time.Time `json:"created_at"`
}

// RoomWithUsers represents a room with its associated users
type RoomWithUsers struct {
	Room
	Users []*User `json:"users"`
}

// UserWithRooms represents a user with their associated rooms
type UserWithRooms struct {
	User
	Rooms []*Room `json:"rooms"`
}

// AssignUserToRoomRequest represents the payload for assigning a user to a room
type AssignUserToRoomRequest struct {
	UserID int64 `json:"user_id"`
	RoomID int64 `json:"room_id"`
}
