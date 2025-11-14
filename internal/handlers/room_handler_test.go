package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudflaredb/internal/models"
	"cloudflaredb/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDBForRooms creates an in-memory SQLite database for testing rooms
func setupTestDBForRooms(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

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
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestRoomHandler_CreateRoom(t *testing.T) {
	db := setupTestDBForRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewRoomRepository(db)
	handler := NewRoomHandler(repo)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful creation",
			requestBody: models.CreateRoomRequest{
				Name:        "Conference Room",
				Description: "Large meeting room",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var room models.Room
				if err := json.Unmarshal(body, &room); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if room.Name != "Conference Room" {
					t.Errorf("Expected name 'Conference Room', got %s", room.Name)
				}
			},
		},
		{
			name: "missing name",
			requestBody: models.CreateRoomRequest{
				Description: "Test",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.CreateRoom(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestRoomHandler_GetRoom(t *testing.T) {
	db := setupTestDBForRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewRoomRepository(db)
	handler := NewRoomHandler(repo)

	// Create a test room
	ctx := context.Background()
	room, err := repo.Create(ctx, &models.CreateRoomRequest{
		Name:        "Test Room",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	tests := []struct {
		name           string
		roomID         string
		expectedStatus int
	}{
		{
			name:           "existing room",
			roomID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent room",
			roomID:         "9999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid room ID",
			roomID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/rooms/"+tt.roomID, nil)
			w := httptest.NewRecorder()

			handler.GetRoom(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var gotRoom models.Room
				if err := json.Unmarshal(w.Body.Bytes(), &gotRoom); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if gotRoom.ID != room.ID {
					t.Errorf("Expected room ID %d, got %d", room.ID, gotRoom.ID)
				}
			}
		})
	}
}

func TestRoomHandler_ListRooms(t *testing.T) {
	db := setupTestDBForRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewRoomRepository(db)
	handler := NewRoomHandler(repo)

	// Create test rooms
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		_, err := repo.Create(ctx, &models.CreateRoomRequest{
			Name:        "Room " + string(rune('0'+i)),
			Description: "Test room description",
		})
		if err != nil {
			t.Fatalf("Failed to create test room: %v", err)
		}
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			name:           "limit 3",
			queryParams:    "?limit=3",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "offset 2",
			queryParams:    "?offset=2",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/rooms"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListRooms(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var rooms []*models.Room
				if err := json.Unmarshal(w.Body.Bytes(), &rooms); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(rooms) != tt.expectedCount {
					t.Errorf("Expected %d rooms, got %d", tt.expectedCount, len(rooms))
				}
			}
		})
	}
}

func TestRoomHandler_AssignUserToRoom(t *testing.T) {
	db := setupTestDBForRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	roomRepo := repository.NewRoomRepository(db)
	userRepo := repository.NewUserRepository(db)
	handler := NewRoomHandler(roomRepo)

	ctx := context.Background()

	// Create test user
	user, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})

	// Create test room
	room, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:        "Test Room",
		Description: "Test Description",
	})

	requestBody := map[string]int64{
		"user_id": user.ID,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/rooms/1/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.AssignUserToRoom(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify assignment
	rooms, err := roomRepo.GetUserRooms(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user rooms: %v", err)
	}

	if len(rooms) != 1 || rooms[0].ID != room.ID {
		t.Error("User not properly assigned to room")
	}
}

func TestRoomHandler_GetRoomUsers(t *testing.T) {
	db := setupTestDBForRooms(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	roomRepo := repository.NewRoomRepository(db)
	userRepo := repository.NewUserRepository(db)
	handler := NewRoomHandler(roomRepo)

	ctx := context.Background()

	// Create test room
	room, _ := roomRepo.Create(ctx, &models.CreateRoomRequest{
		Name:        "Test Room",
		Description: "Test Description",
	})

	// Create test users
	user1, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user1",
	})
	user2, _ := userRepo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user2",
	})

	// Assign users to room
	roomRepo.AssignUserToRoom(ctx, user1.ID, room.ID)
	roomRepo.AssignUserToRoom(ctx, user2.ID, room.ID)

	req := httptest.NewRequest(http.MethodGet, "/rooms/1/users", nil)
	w := httptest.NewRecorder()

	handler.GetRoomUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var roomWithUsers models.RoomWithUsers
	if err := json.Unmarshal(w.Body.Bytes(), &roomWithUsers); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(roomWithUsers.Users) != 2 {
		t.Errorf("Expected 2 users in room, got %d", len(roomWithUsers.Users))
	}
}
