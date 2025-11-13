package repository

import (
	"context"
	"database/sql"
	"fmt"
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

// Create inserts a new room into the database
func (r *RoomRepository) Create(ctx context.Context, req *models.CreateRoomRequest) (*models.Room, error) {
	query := `
		INSERT INTO rooms (name, description, capacity, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, name, description, capacity, created_at, updated_at
	`

	now := time.Now()
	row := r.db.QueryRowContext(ctx, query, req.Name, req.Description, req.Capacity, now, now)

	room := &models.Room{}
	err := row.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return room, nil
}

// GetByID retrieves a room by ID
func (r *RoomRepository) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	query := `
		SELECT id, name, description, capacity, created_at, updated_at
		FROM rooms
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	room := &models.Room{}
	err := row.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found")
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return room, nil
}

// List retrieves all rooms with pagination
func (r *RoomRepository) List(ctx context.Context, limit, offset int) ([]*models.Room, error) {
	query := `
		SELECT id, name, description, capacity, created_at, updated_at
		FROM rooms
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*models.Room
	for rows.Next() {
		room := &models.Room{}
		err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt, &room.UpdatedAt)
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

// Update updates a room's information
func (r *RoomRepository) Update(ctx context.Context, id int64, req *models.UpdateRoomRequest) (*models.Room, error) {
	query := `
		UPDATE rooms
		SET name = COALESCE(NULLIF(?, ''), name),
		    description = COALESCE(NULLIF(?, ''), description),
		    capacity = CASE WHEN ? > 0 THEN ? ELSE capacity END,
		    updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Name, req.Description, req.Capacity, req.Capacity, now, id)
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

// Delete removes a room from the database
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

// GetRoomWithUsers retrieves a room with all its assigned users
func (r *RoomRepository) GetRoomWithUsers(ctx context.Context, roomID int64) (*models.RoomWithUsers, error) {
	// Get room details
	room, err := r.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Get all users assigned to this room
	query := `
		SELECT u.id, u.email, u.name, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_rooms ur ON u.id = ur.user_id
		WHERE ur.room_id = ?
		ORDER BY u.name
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
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

// AssignUserToRoom assigns a user to a room (user can have multiple rooms)
func (r *RoomRepository) AssignUserToRoom(ctx context.Context, userID, roomID int64) error {
	// Check if assignment already exists
	checkQuery := `SELECT COUNT(*) FROM user_rooms WHERE user_id = ? AND room_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, checkQuery, userID, roomID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
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

// RemoveUserFromRoom removes a user from a specific room
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

// RemoveUserFromAllRooms removes a user from all their assigned rooms
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

// GetUserRooms retrieves all rooms assigned to a user
func (r *RoomRepository) GetUserRooms(ctx context.Context, userID int64) ([]*models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.capacity, r.created_at, r.updated_at
		FROM rooms r
		INNER JOIN user_rooms ur ON r.id = ur.room_id
		WHERE ur.user_id = ?
		ORDER BY r.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*models.Room
	for rows.Next() {
		room := &models.Room{}
		err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt, &room.UpdatedAt)
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
