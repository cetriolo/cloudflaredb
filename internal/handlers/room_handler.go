package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Room name is required")
		return
	}

	if req.Capacity < 1 {
		respondError(w, http.StatusBadRequest, "Capacity must be at least 1")
		return
	}

	room, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create room: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, room)
}

// GetRoom handles GET /rooms/{id}
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/rooms/")
	// Remove any trailing path segments
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Room not found")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get room: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, room)
}

// ListRooms handles GET /rooms
func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	rooms, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list rooms: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, rooms)
}

// UpdateRoom handles PUT /rooms/{id}
func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/rooms/")
	// Remove any trailing path segments
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	var req models.UpdateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	room, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Room not found")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update room: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, room)
}

// DeleteRoom handles DELETE /rooms/{id}
func (h *RoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/rooms/")
	// Remove any trailing path segments
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Room not found")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete room: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRoomUsers handles GET /rooms/{id}/users
func (h *RoomHandler) GetRoomUsers(w http.ResponseWriter, r *http.Request) {
	// Extract room ID from path: /rooms/{id}/users
	path := strings.TrimPrefix(r.URL.Path, "/rooms/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		respondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomWithUsers, err := h.repo.GetRoomWithUsers(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Room not found")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get room users: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, roomWithUsers)
}

// AssignUserToRoom handles POST /rooms/{id}/users
func (h *RoomHandler) AssignUserToRoom(w http.ResponseWriter, r *http.Request) {
	// Extract room ID from path: /rooms/{id}/users
	path := strings.TrimPrefix(r.URL.Path, "/rooms/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		respondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	roomID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.repo.AssignUserToRoom(r.Context(), req.UserID, roomID); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to assign user to room: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User assigned to room successfully"})
}

// RemoveUserFromRoom handles DELETE /rooms/{roomId}/users/{userId}
func (h *RoomHandler) RemoveUserFromRoom(w http.ResponseWriter, r *http.Request) {
	// Extract IDs from path: /rooms/{roomId}/users/{userId}
	path := strings.TrimPrefix(r.URL.Path, "/rooms/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		respondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	roomID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	userID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.repo.RemoveUserFromRoom(r.Context(), userID, roomID); err != nil {
		if strings.Contains(err.Error(), "not assigned") {
			respondError(w, http.StatusNotFound, "User not assigned to this room")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove user from room: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserRooms handles GET /users/{id}/rooms
func (h *RoomHandler) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path: /users/{id}/rooms
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		respondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	rooms, err := h.repo.GetUserRooms(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user rooms: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, rooms)
}
