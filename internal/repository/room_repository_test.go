package repository

import (
	"context"
	"database/sql"
	"testing"

	"cloudflaredb/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDBWithRooms creates an in-memory SQLite database with users, room_types, and rooms tables
func setupTestDBWithRooms(t *testing.T) *sql.DB {
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

	CREATE TABLE room_types (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		size TEXT NOT NULL,
		style TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_room_types_size ON room_types(size);
	CREATE INDEX idx_room_types_style ON room_types(style);

	CREATE TABLE rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		room_type_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (room_type_id) REFERENCES room_types(id) ON DELETE SET NULL
	);
	CREATE INDEX idx_rooms_name ON rooms(name);
	CREATE INDEX idx_rooms_room_type_id ON rooms(room_type_id);

	CREATE TABLE user_rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		room_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
		UNIQUE(user_id, room_id)
	);
	CREATE INDEX idx_user_rooms_user_id ON user_rooms(user_id);
	CREATE INDEX idx_user_rooms_room_id ON user_rooms(room_id);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestRoomRepository_Create(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	repo := NewRoomRepository(db)
	roomTypeRepo := NewRoomTypeRepository(db)
	ctx := context.Background()

	// Create a test room type
	roomType, err := roomTypeRepo.Create(ctx, &models.CreateRoomTypeRequest{
		Size:  "large",
		Style: "conference",
	})
	if err != nil {
		t.Fatalf("Failed to create test room type: %v", err)
	}

	tests := []struct {
		name    string
		req     *models.CreateRoomRequest
		wantErr bool
	}{
		{
			name: "successful creation with room type",
			req: &models.CreateRoomRequest{
				Name:        "Conference Room A",
				Description: "Large conference room",
				RoomTypeID:  &roomType.ID,
			},
			wantErr: false,
		},
		{
			name: "successful creation without room type",
			req: &models.CreateRoomRequest{
				Name:        "Meeting Room",
				Description: "Small meeting room",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := repo.Create(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if room.ID == 0 {
					t.Error("Expected room ID to be set")
				}
				if room.Name != tt.req.Name {
					t.Errorf("Expected name %s, got %s", tt.req.Name, room.Name)
				}
				if tt.req.RoomTypeID != nil && (room.RoomTypeID == nil || *room.RoomTypeID != *tt.req.RoomTypeID) {
					t.Errorf("Expected room_type_id %d, got %v", *tt.req.RoomTypeID, room.RoomTypeID)
				}
			}
		})
	}
}

func TestRoomRepository_GetByID(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	room, err := repo.Create(ctx, &models.CreateRoomRequest{
		Name:        "Test Room",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing room",
			id:      room.ID,
			wantErr: false,
		},
		{
			name:    "non-existent room",
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
				if got.ID != room.ID {
					t.Errorf("Expected ID %d, got %d", room.ID, got.ID)
				}
				if got.Name != room.Name {
					t.Errorf("Expected name %s, got %s", room.Name, got.Name)
				}
			}
		})
	}
}

func TestRoomRepository_List(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create test rooms
	for i := 1; i <= 5; i++ {
		_, err := repo.Create(ctx, &models.CreateRoomRequest{
			Name: "Room " + string(rune(i+'0')),
		})
		if err != nil {
			t.Fatalf("Failed to create test room: %v", err)
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
			rooms, err := repo.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(rooms) != tt.wantCount {
				t.Errorf("Expected %d rooms, got %d", tt.wantCount, len(rooms))
			}
		})
	}
}

func TestRoomRepository_Update(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	repo := NewRoomRepository(db)
	roomTypeRepo := NewRoomTypeRepository(db)
	ctx := context.Background()

	// Create a test room type
	roomType, err := roomTypeRepo.Create(ctx, &models.CreateRoomTypeRequest{
		Size:  "medium",
		Style: "office",
	})
	if err != nil {
		t.Fatalf("Failed to create test room type: %v", err)
	}

	// Create a test room
	room, err := repo.Create(ctx, &models.CreateRoomRequest{
		Name:        "Original Room",
		Description: "Original Description",
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		req     *models.UpdateRoomRequest
		wantErr bool
	}{
		{
			name: "update name",
			id:   room.ID,
			req: &models.UpdateRoomRequest{
				Name: "Updated Room",
			},
			wantErr: false,
		},
		{
			name: "update room_type_id",
			id:   room.ID,
			req: &models.UpdateRoomRequest{
				RoomTypeID: &roomType.ID,
			},
			wantErr: false,
		},
		{
			name: "update non-existent room",
			id:   9999,
			req: &models.UpdateRoomRequest{
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
				if tt.req.RoomTypeID != nil && (updated.RoomTypeID == nil || *updated.RoomTypeID != *tt.req.RoomTypeID) {
					t.Errorf("Expected room_type_id %d, got %v", *tt.req.RoomTypeID, updated.RoomTypeID)
				}
			}
		})
	}
}

func TestRoomRepository_Delete(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	room, err := repo.Create(ctx, &models.CreateRoomRequest{
		Name: "Test Room",
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "delete existing room",
			id:      room.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent room",
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
				// Verify room is deleted
				_, err := repo.GetByID(ctx, tt.id)
				if err == nil {
					t.Error("Expected room to be deleted")
				}
			}
		})
	}
}

func TestRoomRepository_AssignUserToRoom(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user, err := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test rooms
	room1, err := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name: "Room 1",
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	room2, err := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name: "Room 2",
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	// Test assigning user to first room
	err = roomRepo.AssignUserToRoom(ctx, user.ID, room1.ID)
	if err != nil {
		t.Errorf("Failed to assign user to room 1: %v", err)
	}

	// Test assigning same user to second room (should succeed - many-to-many)
	err = roomRepo.AssignUserToRoom(ctx, user.ID, room2.ID)
	if err != nil {
		t.Errorf("Failed to assign user to room 2: %v", err)
	}

	// Test assigning user to same room again (should fail - duplicate)
	err = roomRepo.AssignUserToRoom(ctx, user.ID, room1.ID)
	if err == nil {
		t.Error("Expected error when assigning user to same room twice")
	}

	// Verify user is assigned to both rooms
	rooms, err := roomRepo.GetUserRooms(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user rooms: %v", err)
	}

	if len(rooms) != 2 {
		t.Errorf("Expected user to be in 2 rooms, got %d", len(rooms))
	}
}

func TestRoomRepository_GetRoomWithUsers(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test room
	room, err := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name: "Test Room",
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	// Create test users
	user1, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user1",
	})
	user2, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user2",
	})

	// Assign users to room
	err = roomRepo.AssignUserToRoom(ctx, user1.ID, room.ID)
	if err != nil {
		t.Errorf("Failed to assign user1 to room: %v", err)
		return
	}
	err = roomRepo.AssignUserToRoom(ctx, user2.ID, room.ID)
	if err != nil {
		t.Errorf("Failed to assign user2 to room: %v", err)
		return
	}

	// Get room with users
	roomWithUsers, err := roomRepo.GetRoomWithUsers(ctx, room.ID)
	if err != nil {
		t.Fatalf("Failed to get room with users: %v", err)
	}

	if len(roomWithUsers.Users) != 2 {
		t.Errorf("Expected 2 users in room, got %d", len(roomWithUsers.Users))
	}
}

func TestRoomRepository_RemoveUserFromRoom(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user and room
	user, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})
	room, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name: "Test Room",
	})

	// Assign user to room
	err := roomRepo.AssignUserToRoom(ctx, user.ID, room.ID)
	if err != nil {
		t.Fatalf("Failed to assign user to room: %v", err)
		return
	}

	// Remove user from room
	err = roomRepo.RemoveUserFromRoom(ctx, user.ID, room.ID)
	if err != nil {
		t.Errorf("Failed to remove user from room: %v", err)
	}

	// Verify user is removed
	rooms, err := roomRepo.GetUserRooms(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user rooms: %v", err)
	}

	if len(rooms) != 0 {
		t.Errorf("Expected user to have 0 rooms, got %d", len(rooms))
	}

	// Try removing again (should fail)
	err = roomRepo.RemoveUserFromRoom(ctx, user.ID, room.ID)
	if err == nil {
		t.Error("Expected error when removing user from room they're not in")
	}
}

func TestRoomRepository_GetUserRooms(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(db)

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})

	// Create test rooms
	room1, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name: "Room 1",
	})
	room2, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name: "Room 2",
	})
	room3, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name: "Room 3",
	})

	// Assign user to multiple rooms
	err := roomRepo.AssignUserToRoom(ctx, user.ID, room1.ID)
	if err != nil {
		t.Fatalf("Failed to assign user rooms: %v", err)
		return
	}
	err = roomRepo.AssignUserToRoom(ctx, user.ID, room2.ID)
	if err != nil {
		t.Fatalf("Failed to assign user rooms: %v", err)
		return
	}
	err = roomRepo.AssignUserToRoom(ctx, user.ID, room3.ID)
	if err != nil {
		t.Fatalf("Failed to assign user rooms: %v", err)
		return
	}

	// Get user rooms
	rooms, err := roomRepo.GetUserRooms(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user rooms: %v", err)
	}

	if len(rooms) != 3 {
		t.Errorf("Expected user to have 3 rooms, got %d", len(rooms))
	}

	// Verify rooms are sorted by name
	if rooms[0].Name != "Room 1" {
		t.Errorf("Expected first room to be 'Room 1', got '%s'", rooms[0].Name)
	}
}
