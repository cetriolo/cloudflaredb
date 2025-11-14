package repository

import (
	"context"
	"database/sql"
	"testing"

	"cloudflaredb/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDBWithRoomTypes creates an in-memory SQLite database with room_types table
func setupTestDBWithRoomTypes(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create schema
	schema := `
	CREATE TABLE room_types (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		size TEXT NOT NULL,
		style TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_room_types_size ON room_types(size);
	CREATE INDEX idx_room_types_style ON room_types(style);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestRoomTypeRepository_Create(t *testing.T) {
	db := setupTestDBWithRoomTypes(t)
	defer db.Close()

	repo := NewRoomTypeRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *models.CreateRoomTypeRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			req: &models.CreateRoomTypeRequest{
				Size:  "large",
				Style: "conference",
			},
			wantErr: false,
		},
		{
			name: "another successful creation",
			req: &models.CreateRoomTypeRequest{
				Size:  "small",
				Style: "office",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roomType, err := repo.Create(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if roomType.ID == 0 {
					t.Error("Expected room type ID to be set")
				}
				if roomType.Size != tt.req.Size {
					t.Errorf("Expected size %s, got %s", tt.req.Size, roomType.Size)
				}
				if roomType.Style != tt.req.Style {
					t.Errorf("Expected style %s, got %s", tt.req.Style, roomType.Style)
				}
				if roomType.CreatedAt.IsZero() {
					t.Error("Expected created_at to be set")
				}
			}
		})
	}
}

func TestRoomTypeRepository_GetByID(t *testing.T) {
	db := setupTestDBWithRoomTypes(t)
	defer db.Close()

	repo := NewRoomTypeRepository(db)
	ctx := context.Background()

	// Create a test room type
	roomType, err := repo.Create(ctx, &models.CreateRoomTypeRequest{
		Size:  "medium",
		Style: "meeting",
	})
	if err != nil {
		t.Fatalf("Failed to create test room type: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing room type",
			id:      roomType.ID,
			wantErr: false,
		},
		{
			name:    "non-existent room type",
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
				if got.ID != roomType.ID {
					t.Errorf("Expected ID %d, got %d", roomType.ID, got.ID)
				}
				if got.Size != roomType.Size {
					t.Errorf("Expected size %s, got %s", roomType.Size, got.Size)
				}
				if got.Style != roomType.Style {
					t.Errorf("Expected style %s, got %s", roomType.Style, got.Style)
				}
			}
		})
	}
}

func TestRoomTypeRepository_List(t *testing.T) {
	db := setupTestDBWithRoomTypes(t)
	defer db.Close()

	repo := NewRoomTypeRepository(db)
	ctx := context.Background()

	// Create test room types
	sizes := []string{"small", "medium", "large", "xlarge", "xxlarge"}
	for _, size := range sizes {
		_, err := repo.Create(ctx, &models.CreateRoomTypeRequest{
			Size:  size,
			Style: "office",
		})
		if err != nil {
			t.Fatalf("Failed to create test room type: %v", err)
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
			roomTypes, err := repo.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(roomTypes) != tt.wantCount {
				t.Errorf("Expected %d room types, got %d", tt.wantCount, len(roomTypes))
			}
		})
	}
}

func TestRoomTypeRepository_Update(t *testing.T) {
	db := setupTestDBWithRoomTypes(t)
	defer db.Close()

	repo := NewRoomTypeRepository(db)
	ctx := context.Background()

	// Create a test room type
	roomType, err := repo.Create(ctx, &models.CreateRoomTypeRequest{
		Size:  "small",
		Style: "office",
	})
	if err != nil {
		t.Fatalf("Failed to create test room type: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		req     *models.UpdateRoomTypeRequest
		wantErr bool
	}{
		{
			name: "update size",
			id:   roomType.ID,
			req: &models.UpdateRoomTypeRequest{
				Size: "large",
			},
			wantErr: false,
		},
		{
			name: "update style",
			id:   roomType.ID,
			req: &models.UpdateRoomTypeRequest{
				Style: "conference",
			},
			wantErr: false,
		},
		{
			name: "update non-existent room type",
			id:   9999,
			req: &models.UpdateRoomTypeRequest{
				Size: "huge",
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
				if tt.req.Size != "" && updated.Size != tt.req.Size {
					t.Errorf("Expected size %s, got %s", tt.req.Size, updated.Size)
				}
				if tt.req.Style != "" && updated.Style != tt.req.Style {
					t.Errorf("Expected style %s, got %s", tt.req.Style, updated.Style)
				}
			}
		})
	}
}

func TestRoomTypeRepository_Delete(t *testing.T) {
	db := setupTestDBWithRoomTypes(t)
	defer db.Close()

	repo := NewRoomTypeRepository(db)
	ctx := context.Background()

	// Create a test room type
	roomType, err := repo.Create(ctx, &models.CreateRoomTypeRequest{
		Size:  "medium",
		Style: "meeting",
	})
	if err != nil {
		t.Fatalf("Failed to create test room type: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "delete existing room type",
			id:      roomType.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent room type",
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
				// Verify room type is deleted
				_, err := repo.GetByID(ctx, tt.id)
				if err == nil {
					t.Error("Expected room type to be deleted")
				}
			}
		})
	}
}
