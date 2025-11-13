package repository

import (
	"context"
	"database/sql"
	"testing"

	"cloudflaredb/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDBWithRooms creates an in-memory SQLite database with users and rooms tables
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
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_users_email ON users(email);

	CREATE TABLE rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		capacity INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_rooms_name ON rooms(name);

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
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *models.CreateRoomRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			req: &models.CreateRoomRequest{
				Name:        "Conference Room A",
				Description: "Large conference room",
				Capacity:    20,
			},
			wantErr: false,
		},
		{
			name: "with minimal fields",
			req: &models.CreateRoomRequest{
				Name:     "Meeting Room",
				Capacity: 5,
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
				if room.Capacity != tt.req.Capacity {
					t.Errorf("Expected capacity %d, got %d", tt.req.Capacity, room.Capacity)
				}
			}
		})
	}
}

func TestRoomRepository_GetByID(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	room, err := repo.Create(ctx, &models.CreateRoomRequest{
		Name:        "Test Room",
		Description: "Test Description",
		Capacity:    10,
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
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create test rooms
	for i := 1; i <= 5; i++ {
		_, err := repo.Create(ctx, &models.CreateRoomRequest{
			Name:     "Room " + string(rune(i)),
			Capacity: i * 10,
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
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	room, err := repo.Create(ctx, &models.CreateRoomRequest{
		Name:        "Original Room",
		Description: "Original Description",
		Capacity:    10,
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
			name: "update capacity",
			id:   room.ID,
			req: &models.UpdateRoomRequest{
				Capacity: 25,
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
				if tt.req.Capacity > 0 && updated.Capacity != tt.req.Capacity {
					t.Errorf("Expected capacity %d, got %d", tt.req.Capacity, updated.Capacity)
				}
			}
		})
	}
}

func TestRoomRepository_Delete(t *testing.T) {
	db := setupTestDBWithRooms(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	room, err := repo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Test Room",
		Capacity: 10,
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
	defer db.Close()

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user, err := userRepo.Create(ctx, &models.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test rooms
	room1, err := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Room 1",
		Capacity: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	room2, err := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Room 2",
		Capacity: 20,
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
	defer db.Close()

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test room
	room, err := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Test Room",
		Capacity: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	// Create test users
	user1, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		Email: "user1@example.com",
		Name:  "User 1",
	})
	user2, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		Email: "user2@example.com",
		Name:  "User 2",
	})

	// Assign users to room
	roomRepo.AssignUserToRoom(ctx, user1.ID, room.ID)
	roomRepo.AssignUserToRoom(ctx, user2.ID, room.ID)

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
	defer db.Close()

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user and room
	user, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Test User",
	})
	room, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Test Room",
		Capacity: 10,
	})

	// Assign user to room
	roomRepo.AssignUserToRoom(ctx, user.ID, room.ID)

	// Remove user from room
	err := roomRepo.RemoveUserFromRoom(ctx, user.ID, room.ID)
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
	defer db.Close()

	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Test User",
	})

	// Create test rooms
	room1, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Room 1",
		Capacity: 10,
	})
	room2, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Room 2",
		Capacity: 20,
	})
	room3, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:     "Room 3",
		Capacity: 30,
	})

	// Assign user to multiple rooms
	roomRepo.AssignUserToRoom(ctx, user.ID, room1.ID)
	roomRepo.AssignUserToRoom(ctx, user.ID, room2.ID)
	roomRepo.AssignUserToRoom(ctx, user.ID, room3.ID)

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
