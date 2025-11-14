-- Migration: Create room_types table
-- Created: 2025-11-14
-- Description: Add room_types table to define room categories with size and style

-- Create room_types table
CREATE TABLE IF NOT EXISTS room_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    size TEXT NOT NULL,
    style TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index on size for filtering
CREATE INDEX IF NOT EXISTS idx_room_types_size ON room_types(size);

-- Create index on style for filtering
CREATE INDEX IF NOT EXISTS idx_room_types_style ON room_types(style);
