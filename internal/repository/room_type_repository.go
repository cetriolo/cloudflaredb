package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"cloudflaredb/internal/models"
)

// RoomTypeRepository handles database operations for room types
type RoomTypeRepository struct {
	db *sql.DB
}

// NewRoomTypeRepository creates a new room type repository
func NewRoomTypeRepository(db *sql.DB) *RoomTypeRepository {
	return &RoomTypeRepository{db: db}
}

// Create inserts a new room type into the database.
// It automatically sets created_at and updated_at timestamps.
// Returns the created room type with its generated ID, or an error if creation fails.
func (r *RoomTypeRepository) Create(ctx context.Context, req *models.CreateRoomTypeRequest) (*models.RoomType, error) {
	query := `
		INSERT INTO room_types (size, style, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Size, req.Style, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create room type: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Fetch the created room type
	return r.GetByID(ctx, id)
}

// GetByID retrieves a single room type by its ID.
// Returns an error with message "room type not found" if the room type doesn't exist.
// Uses the scanRoomType helper to handle database type conversions.
func (r *RoomTypeRepository) GetByID(ctx context.Context, id int64) (*models.RoomType, error) {
	query := `
		SELECT id, size, style, created_at, updated_at
		FROM room_types
		WHERE id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query room type: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	if !rows.Next() {
		return nil, fmt.Errorf("room type not found")
	}

	roomType := &models.RoomType{}
	err = scanRoomType(rows, &roomType.ID, &roomType.Size, &roomType.Style, &roomType.CreatedAt, &roomType.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan room type: %w", err)
	}

	return roomType, nil
}

// List retrieves all room types with pagination support.
// Results are ordered by created_at in descending order (newest first).
// Parameters:
//   - limit: maximum number of room types to return
//   - offset: number of room types to skip before starting to return results
//
// Returns an empty slice if no room types are found.
func (r *RoomTypeRepository) List(ctx context.Context, limit, offset int) ([]*models.RoomType, error) {
	query := `
		SELECT id, size, style, created_at, updated_at
		FROM room_types
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list room types: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	var roomTypes []*models.RoomType
	for rows.Next() {
		roomType := &models.RoomType{}
		err := scanRoomType(rows, &roomType.ID, &roomType.Size, &roomType.Style, &roomType.CreatedAt, &roomType.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room type: %w", err)
		}
		roomTypes = append(roomTypes, roomType)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return roomTypes, nil
}

// Update updates a room type's information.
// Only non-empty fields in the request will be updated (partial updates supported).
// The updated_at timestamp is automatically updated to the current time.
// Returns the updated room type object or an error if the room type is not found.
func (r *RoomTypeRepository) Update(ctx context.Context, id int64, req *models.UpdateRoomTypeRequest) (*models.RoomType, error) {
	query := `
		UPDATE room_types
		SET size = COALESCE(NULLIF(?, ''), size),
		    style = COALESCE(NULLIF(?, ''), style),
		    updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Size, req.Style, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update room type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("room type not found")
	}

	return r.GetByID(ctx, id)
}

// Delete permanently removes a room type from the database.
// Returns an error with message "room type not found" if the room type doesn't exist.
// Note: Rooms referencing this room type will have their room_type_id set to NULL (ON DELETE SET NULL).
func (r *RoomTypeRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM room_types WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete room type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("room type not found")
	}

	return nil
}
