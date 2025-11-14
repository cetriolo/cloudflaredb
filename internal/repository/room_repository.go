package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"cloudflaredb/internal/models"
)

// RoomRepository handles database operations for rooms
type RoomRepository struct {
	db *sql.DB
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create inserts a new room into the database.
// It automatically sets created_at and updated_at timestamps.
// Returns the created room with its generated ID, or an error if creation fails.
func (r *RoomRepository) Create(ctx context.Context, req *models.CreateRoomRequest) (*models.Room, error) {
	query := `
		INSERT INTO rooms (name, description, room_type_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Name, req.Description, req.RoomTypeID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Fetch the created room
	return r.GetByID(ctx, id)
}

// GetByID retrieves a single room by its ID.
// Returns an error with message "room not found" if the room doesn't exist.
// Uses the scanRoom helper to handle database type conversions.
func (r *RoomRepository) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	query := `
		SELECT id, name, description, room_type_id, created_at, updated_at
		FROM rooms
		WHERE id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query room: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	if !rows.Next() {
		return nil, fmt.Errorf("room not found")
	}

	room := &models.Room{}
	err = scanRoom(rows, &room.ID, &room.Name, &room.Description, &room.RoomTypeID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan room: %w", err)
	}

	return room, nil
}

// List retrieves all rooms with pagination support.
// Results are ordered by created_at in descending order (newest first).
// Parameters:
//   - limit: maximum number of rooms to return
//   - offset: number of rooms to skip before starting to return results
//
// Returns an empty slice if no rooms are found.
func (r *RoomRepository) List(ctx context.Context, limit, offset int) ([]*models.Room, error) {
	query := `
		SELECT id, name, description, room_type_id, created_at, updated_at
		FROM rooms
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	var rooms []*models.Room
	for rows.Next() {
		room := &models.Room{}
		err := scanRoom(rows, &room.ID, &room.Name, &room.Description, &room.RoomTypeID, &room.CreatedAt, &room.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return rooms, nil
}

// Update updates a room's information.
// Only non-empty/non-zero fields in the request will be updated (partial updates supported).
// The updated_at timestamp is automatically updated to the current time.
// For room_type_id: only updates if the new value is not nil.
// Returns the updated room object or an error if the room is not found.
func (r *RoomRepository) Update(ctx context.Context, id int64, req *models.UpdateRoomRequest) (*models.Room, error) {
	query := `
		UPDATE rooms
		SET name = COALESCE(NULLIF(?, ''), name),
		    description = COALESCE(NULLIF(?, ''), description),
		    room_type_id = CASE WHEN ? IS NOT NULL THEN ? ELSE room_type_id END,
		    updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Name, req.Description, req.RoomTypeID, req.RoomTypeID, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("room not found")
	}

	return r.GetByID(ctx, id)
}

// Delete permanently removes a room from the database.
// Returns an error with message "room not found" if the room doesn't exist.
// Note: This does not cascade delete related records like user_rooms.
// Consider handling related records before calling this method.
func (r *RoomRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM rooms WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("room not found")
	}

	return nil
}

// GetRoomWithUsers retrieves a room along with all users assigned to it.
// Returns a RoomWithUsers object containing the room details and a list of users.
// Users are ordered by their external_id.
// Returns an error if the room is not found.
// The Users slice will be empty if no users are assigned to the room.
func (r *RoomRepository) GetRoomWithUsers(ctx context.Context, roomID int64) (*models.RoomWithUsers, error) {
	// Get room details
	room, err := r.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Get all users assigned to this room
	query := `
		SELECT u.id, u.external_id, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_rooms ur ON u.id = ur.user_id
		WHERE ur.room_id = ?
		ORDER BY u.external_id
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room users: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := scanUser(rows, &user.ID, &user.ExternalID, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &models.RoomWithUsers{
		Room:  *room,
		Users: users,
	}, nil
}

// AssignUserToRoom creates a many-to-many relationship between a user and a room.
// A user can be assigned to multiple rooms, and a room can have multiple users.
// Returns an error if the user is already assigned to this room.
// The created_at timestamp is automatically set for the relationship.
func (r *RoomRepository) AssignUserToRoom(ctx context.Context, userID, roomID int64) error {
	// Check if assignment already exists
	checkQuery := `SELECT COUNT(*) FROM user_rooms WHERE user_id = ? AND room_id = ?`
	var countResult interface{}
	err := r.db.QueryRowContext(ctx, checkQuery, userID, roomID).Scan(&countResult)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}

	// Handle both int64 (SQLite3) and float64 (D1) count types
	var count int64
	switch v := countResult.(type) {
	case int64:
		count = v
	case float64:
		count = int64(v)
	case int:
		count = int64(v)
	default:
		return fmt.Errorf("unexpected count type: %T", countResult)
	}

	if count > 0 {
		return fmt.Errorf("user already assigned to this room")
	}

	// Assign user to room
	insertQuery := `
		INSERT INTO user_rooms (user_id, room_id, created_at)
		VALUES (?, ?, ?)
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, insertQuery, userID, roomID, now)
	if err != nil {
		return fmt.Errorf("failed to assign user to room: %w", err)
	}

	return nil
}

// RemoveUserFromRoom deletes the many-to-many relationship between a user and a room.
// Returns an error with message "user not assigned to this room" if the relationship doesn't exist.
// Only removes the relationship; the user and room records remain in the database.
func (r *RoomRepository) RemoveUserFromRoom(ctx context.Context, userID, roomID int64) error {
	query := `DELETE FROM user_rooms WHERE user_id = ? AND room_id = ?`

	result, err := r.db.ExecContext(ctx, query, userID, roomID)
	if err != nil {
		return fmt.Errorf("failed to remove user from room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not assigned to this room")
	}

	return nil
}

// RemoveUserFromAllRooms deletes all room assignments for a specific user.
// Useful when deleting a user to clean up all their room relationships.
// Returns an error with message "user not assigned to any rooms" if the user has no room assignments.
func (r *RoomRepository) RemoveUserFromAllRooms(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_rooms WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from rooms: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not assigned to any rooms")
	}

	return nil
}

// GetUserRooms retrieves all rooms that a specific user is assigned to.
// Results are ordered alphabetically by room name.
// Returns an empty slice if the user is not assigned to any rooms.
// Performs an INNER JOIN on the user_rooms table to get the relationship.
func (r *RoomRepository) GetUserRooms(ctx context.Context, userID int64) ([]*models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.room_type_id, r.created_at, r.updated_at
		FROM rooms r
		INNER JOIN user_rooms ur ON r.id = ur.room_id
		WHERE ur.user_id = ?
		ORDER BY r.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	var rooms []*models.Room
	for rows.Next() {
		room := &models.Room{}
		err := scanRoom(rows, &room.ID, &room.Name, &room.Description, &room.RoomTypeID, &room.CreatedAt, &room.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return rooms, nil
}
