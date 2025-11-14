package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cloudflaredb/internal/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database.
// It automatically sets created_at and updated_at timestamps.
// Returns the created user with its generated ID, or an error if creation fails.
// Will return an error if a user with the same external_id already exists (UNIQUE constraint).
func (r *UserRepository) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (external_id, created_at, updated_at)
		VALUES (?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.ExternalID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Fetch the created user
	return r.GetByID(ctx, id)
}

// GetByID retrieves a single user by their ID.
// Returns an error with message "user not found" if the user doesn't exist.
// Uses the scanUser helper to handle database type conversions.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, external_id, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{}
	err = scanUser(rows, &user.ID, &user.ExternalID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return user, nil
}

// GetByExternalID retrieves a single user by their external ID.
// Returns an error with message "user not found" if no user has the specified external ID.
// Useful for authentication and external ID uniqueness checks.
func (r *UserRepository) GetByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	query := `
		SELECT id, external_id, created_at, updated_at
		FROM users
		WHERE external_id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{}
	err = scanUser(rows, &user.ID, &user.ExternalID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return user, nil
}

// List retrieves all users with pagination support.
// Results are ordered by created_at in descending order (newest first).
// Parameters:
//   - limit: maximum number of users to return
//   - offset: number of users to skip before starting to return results
//
// Returns an empty slice if no users are found.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, external_id, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

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

	return users, nil
}

// Update updates a user's information.
// Only non-empty fields in the request will be updated (partial updates supported).
// The updated_at timestamp is automatically updated to the current time.
// Returns the updated user object or an error if the user is not found.
// Uses COALESCE and NULLIF to handle partial updates gracefully.
func (r *UserRepository) Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	query := `
		UPDATE users
		SET external_id = COALESCE(NULLIF(?, ''), external_id),
		    updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.ExternalID, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return r.GetByID(ctx, id)
}

// Delete permanently removes a user from the database.
// Returns an error with message "user not found" if the user doesn't exist.
// Note: This does not cascade delete related records like user_rooms.
// Consider handling related records before calling this method.
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
