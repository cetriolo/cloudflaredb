-- Migration: Create rooms and user_rooms tables
-- Created: 2025-11-13
-- Description: Add rooms table and many-to-many relationship with users

-- Create rooms table
CREATE TABLE IF NOT EXISTS rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    capacity INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index on room name for faster lookups
CREATE INDEX IF NOT EXISTS idx_rooms_name ON rooms(name);

-- Create index on capacity for filtering
CREATE INDEX IF NOT EXISTS idx_rooms_capacity ON rooms(capacity);

-- Create user_rooms junction table for many-to-many relationship
CREATE TABLE IF NOT EXISTS user_rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    room_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    UNIQUE(user_id, room_id)
);

-- Create indexes for user_rooms table
CREATE INDEX IF NOT EXISTS idx_user_rooms_user_id ON user_rooms(user_id);
CREATE INDEX IF NOT EXISTS idx_user_rooms_room_id ON user_rooms(room_id);

-- Note: Many-to-many relationship
-- A user can be assigned to multiple rooms
-- A room can have multiple users assigned to it
-- The UNIQUE constraint on (user_id, room_id) prevents duplicate assignments
