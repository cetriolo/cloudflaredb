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

func TestUserHandler_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewUserRepository(db)
	handler := NewUserHandler(repo)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful creation",
			requestBody: models.CreateUserRequest{
				ExternalID: "user123",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var user models.User
				if err := json.Unmarshal(body, &user); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if user.ExternalID != "user123" {
					t.Errorf("Expected external_id user123, got %s", user.ExternalID)
				}
			},
		},
		{
			name: "missing external_id",
			requestBody: models.CreateUserRequest{
				ExternalID: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
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

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.CreateUser(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	db := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewUserRepository(db)
	handler := NewUserHandler(repo)

	// Create a test user
	ctx := context.Background()
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "existing user",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent user",
			userID:         "9999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.GetUser(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var gotUser models.User
				if err := json.Unmarshal(w.Body.Bytes(), &gotUser); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if gotUser.ID != user.ID {
					t.Errorf("Expected user ID %d, got %d", user.ID, gotUser.ID)
				}
			}
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	db := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewUserRepository(db)
	handler := NewUserHandler(repo)

	// Create test users
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		_, err := repo.Create(ctx, &models.CreateUserRequest{
			ExternalID: "user" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
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
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListUsers(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var users []*models.User
				if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(users) != tt.expectedCount {
					t.Errorf("Expected %d users, got %d", tt.expectedCount, len(users))
				}
			}
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewUserRepository(db)
	handler := NewUserHandler(repo)

	// Create a test user
	ctx := context.Background()
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:   "successful update",
			userID: "1",
			requestBody: models.UpdateUserRequest{
				ExternalID: "updated123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "non-existent user",
			userID: "9999",
			requestBody: models.UpdateUserRequest{
				ExternalID: "shouldfail",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid JSON",
			userID:         "1",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.UpdateUser(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}

	// Verify the update
	updatedUser, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	if updatedUser.ExternalID != "updated123" {
		t.Errorf("Expected external_id 'updated123', got '%s'", updatedUser.ExternalID)
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	repo := repository.NewUserRepository(db)
	handler := NewUserHandler(repo)

	// Create a test user
	ctx := context.Background()
	user, err := repo.Create(ctx, &models.CreateUserRequest{
		ExternalID: "user123",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			userID:         "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "non-existent user",
			userID:         "9999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.DeleteUser(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}

	// Verify the deletion
	_, err = repo.GetByID(ctx, user.ID)
	if err == nil {
		t.Error("Expected user to be deleted")
	}
}
