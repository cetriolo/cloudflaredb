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

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (email, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Email, req.Name, now, now)
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

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT *
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
	err = scanUser(rows, &user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE email = ?
	`

	rows, err := r.db.QueryContext(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{}
	err = scanUser(rows, &user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return user, nil
}

// List retrieves all users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT *
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
		err := scanUser(rows, &user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
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

// Update updates a user's information
func (r *UserRepository) Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	query := `
		UPDATE users
		SET email = COALESCE(NULLIF(?, ''), email),
		    name = COALESCE(NULLIF(?, ''), name),
		    updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Email, req.Name, now, id)
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

// Delete removes a user from the database
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
