package repository

import (
	"context"
	"database/sql"
	"testing"

	"cloudflaredb/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create schema
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_users_email ON users(email);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *models.CreateUserRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			req: &models.CreateUserRequest{
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			req: &models.CreateUserRequest{
				Email: "test@example.com",
				Name:  "Another User",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.Create(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user.ID == 0 {
					t.Error("Expected user ID to be set")
				}
				if user.Email != tt.req.Email {
					t.Errorf("Expected email %s, got %s", tt.req.Email, user.Email)
				}
				if user.Name != tt.req.Name {
					t.Errorf("Expected name %s, got %s", tt.req.Name, user.Name)
				}
				if user.CreatedAt.IsZero() {
					t.Error("Expected created_at to be set")
				}
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "non-existent user",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.ID != user.ID {
					t.Errorf("Expected ID %d, got %d", user.ID, got.ID)
				}
				if got.Email != user.Email {
					t.Errorf("Expected email %s, got %s", user.Email, got.Email)
				}
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "existing email",
			email:   user.Email,
			wantErr: false,
		},
		{
			name:    "non-existent email",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByEmail(ctx, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Email != user.Email {
					t.Errorf("Expected email %s, got %s", user.Email, got.Email)
				}
			}
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create test users
	for i := 1; i <= 5; i++ {
		_, err := repo.Create(ctx, &models.CreateUserRequest{
			Email: "test" + string(rune(i)) + "@example.com",
			Name:  "Test User",
		})
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "get first 3",
			limit:     3,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "get last 2",
			limit:     10,
			offset:    3,
			wantCount: 2,
		},
		{
			name:      "get all",
			limit:     10,
			offset:    0,
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := repo.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(users) != tt.wantCount {
				t.Errorf("Expected %d users, got %d", tt.wantCount, len(users))
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		req     *models.UpdateUserRequest
		wantErr bool
	}{
		{
			name: "update name",
			id:   user.ID,
			req: &models.UpdateUserRequest{
				Name: "Updated Name",
			},
			wantErr: false,
		},
		{
			name: "update email",
			id:   user.ID,
			req: &models.UpdateUserRequest{
				Email: "updated@example.com",
			},
			wantErr: false,
		},
		{
			name: "update non-existent user",
			id:   9999,
			req: &models.UpdateUserRequest{
				Name: "Should Fail",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := repo.Update(ctx, tt.id, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.req.Name != "" && updated.Name != tt.req.Name {
					t.Errorf("Expected name %s, got %s", tt.req.Name, updated.Name)
				}
				if tt.req.Email != "" && updated.Email != tt.req.Email {
					t.Errorf("Expected email %s, got %s", tt.req.Email, updated.Email)
				}
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "delete existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent user",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify user is deleted
				_, err := repo.GetByID(ctx, tt.id)
				if err == nil {
					t.Error("Expected user to be deleted")
				}
			}
		})
	}
}
