package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cloudflaredb/internal/httputil"
	"cloudflaredb/internal/models"
	"cloudflaredb/internal/repository"
)

// RoomHandler handles HTTP requests for room operations
type RoomHandler struct {
	repo *repository.RoomRepository
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(repo *repository.RoomRepository) *RoomHandler {
	return &RoomHandler{repo: repo}
}

// CreateRoom handles POST /rooms
// Creates a new room with the provided name and optional room_type_id.
// Returns 400 if request body is invalid or name is missing.
// Returns 201 with the created room on success.
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoomRequest
	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Room name is required")
		return
	}

	room, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create room: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, room)
}

// GetRoom handles GET /rooms/{id}
// Retrieves a single room by its ID.
// Returns 400 if the ID format is invalid.
// Returns 404 if the room is not found.
// Returns 200 with the room data on success.
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDFromPath(r.URL.Path, "/rooms/")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.RespondError(w, http.StatusNotFound, "Room not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get room: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, room)
}

// ListRooms handles GET /rooms
// Returns a paginated list of rooms.
// Supports query parameters: limit (default: 10) and offset (default: 0).
// Returns 200 with an array of rooms on success.
func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters from query string
	params := httputil.ParsePaginationParams(r)

	rooms, err := h.repo.List(r.Context(), params.Limit, params.Offset)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list rooms: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, rooms)
}

// UpdateRoom handles PUT /rooms/{id}
// Updates an existing room with the provided fields.
// Only non-zero/non-empty fields in the request will be updated.
// Returns 400 if the ID format or request body is invalid.
// Returns 404 if the room is not found.
// Returns 200 with the updated room on success.
func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDFromPath(r.URL.Path, "/rooms/")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	var req models.UpdateRoomRequest
	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	room, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.RespondError(w, http.StatusNotFound, "Room not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update room: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, room)
}

// DeleteRoom handles DELETE /rooms/{id}
// Permanently deletes a room from the system.
// Returns 400 if the ID format is invalid.
// Returns 404 if the room is not found.
// Returns 204 (No Content) on successful deletion.
func (h *RoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDFromPath(r.URL.Path, "/rooms/")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.RespondError(w, http.StatusNotFound, "Room not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete room: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRoomUsers handles GET /rooms/{id}/users
// Retrieves a room along with all users assigned to it.
// Returns 400 if the path format or ID is invalid.
// Returns 404 if the room is not found.
// Returns 200 with the room data including a list of users on success.
func (h *RoomHandler) GetRoomUsers(w http.ResponseWriter, r *http.Request) {
	// Extract room ID from path: /rooms/{id}/users
	parts := httputil.PathSegments(r.URL.Path, "/rooms/")

	if len(parts) < 2 {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomWithUsers, err := h.repo.GetRoomWithUsers(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.RespondError(w, http.StatusNotFound, "Room not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get room users: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, roomWithUsers)
}

// AssignUserToRoom handles POST /rooms/{id}/users
// Assigns a user to a specific room, creating a many-to-many relationship.
// Request body should contain: {"user_id": 123}
// Returns 400 if the path format, room ID, or user_id is invalid.
// Returns 200 with a success message on successful assignment.
func (h *RoomHandler) AssignUserToRoom(w http.ResponseWriter, r *http.Request) {
	// Extract room ID from path: /rooms/{id}/users
	parts := httputil.PathSegments(r.URL.Path, "/rooms/")

	if len(parts) < 2 {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	roomID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}

	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 {
		httputil.RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.repo.AssignUserToRoom(r.Context(), req.UserID, roomID); err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to assign user to room: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"message": "User assigned to room successfully"})
}

// RemoveUserFromRoom handles DELETE /rooms/{roomId}/users/{userId}
// Removes a user from a specific room, deleting the many-to-many relationship.
// Returns 400 if the path format or IDs are invalid.
// Returns 404 if the user is not assigned to the room.
// Returns 204 (No Content) on successful removal.
func (h *RoomHandler) RemoveUserFromRoom(w http.ResponseWriter, r *http.Request) {
	// Extract IDs from path: /rooms/{roomId}/users/{userId}
	parts := httputil.PathSegments(r.URL.Path, "/rooms/")

	if len(parts) < 3 {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	roomID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	userID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.repo.RemoveUserFromRoom(r.Context(), userID, roomID); err != nil {
		if strings.Contains(err.Error(), "not assigned") {
			httputil.RespondError(w, http.StatusNotFound, "User not assigned to this room")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove user from room: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserRooms handles GET /users/{id}/rooms
// Retrieves all rooms that a specific user is assigned to.
// Returns 400 if the path format or user ID is invalid.
// Returns 200 with an array of rooms on success.
func (h *RoomHandler) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path: /users/{id}/rooms
	parts := httputil.PathSegments(r.URL.Path, "/users/")

	if len(parts) < 2 {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	rooms, err := h.repo.GetUserRooms(r.Context(), userID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user rooms: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, rooms)
}
