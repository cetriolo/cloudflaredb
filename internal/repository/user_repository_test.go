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
		external_id TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_users_external_id ON users(external_id);
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
				ExternalID: "user123",
			},
			wantErr: false,
		},
		{
			name: "duplicate external_id",
			req: &models.CreateUserRequest{
				ExternalID: "user123",
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
				if user.ExternalID != tt.req.ExternalID {
					t.Errorf("Expected external_id %s, got %s", tt.req.ExternalID, user.ExternalID)
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
		ExternalID: "user123",
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
				if got.ExternalID != user.ExternalID {
					t.Errorf("Expected external_id %s, got %s", user.ExternalID, got.ExternalID)
				}
			}
		})
	}
}

func TestUserRepository_GetByExternalID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name       string
		externalID string
		wantErr    bool
	}{
		{
			name:       "existing external_id",
			externalID: user.ExternalID,
			wantErr:    false,
		},
		{
			name:       "non-existent external_id",
			externalID: "nonexistent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByExternalID(ctx, tt.externalID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByExternalID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.ExternalID != user.ExternalID {
					t.Errorf("Expected external_id %s, got %s", user.ExternalID, got.ExternalID)
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
			ExternalID: "user" + string(rune(i+'0')),
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
		ExternalID: "user123",
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
			name: "update external_id",
			id:   user.ID,
			req: &models.UpdateUserRequest{
				ExternalID: "updatedUser123",
			},
			wantErr: false,
		},
		{
			name: "update non-existent user",
			id:   9999,
			req: &models.UpdateUserRequest{
				ExternalID: "shouldFail",
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
				if tt.req.ExternalID != "" && updated.ExternalID != tt.req.ExternalID {
					t.Errorf("Expected external_id %s, got %s", tt.req.ExternalID, updated.ExternalID)
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
		ExternalID: "user123",
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
