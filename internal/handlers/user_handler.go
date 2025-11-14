package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"cloudflaredb/internal/httputil"
	"cloudflaredb/internal/models"
	"cloudflaredb/internal/repository"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	repo *repository.UserRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// CreateUser handles POST /users
// Creates a new user with the provided email and name.
// Returns 400 if request body is invalid or required fields are missing.
// Returns 409 if a user with the same email already exists.
// Returns 201 with the created user on success.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Name == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Email and name are required")
		return
	}

	user, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			httputil.RespondError(w, http.StatusConflict, "User with this email already exists")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /users/{id}
// Retrieves a single user by their ID.
// Returns 400 if the ID format is invalid.
// Returns 404 if the user is not found.
// Returns 200 with the user data on success.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDFromPath(r.URL.Path, "/users/")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.RespondError(w, http.StatusNotFound, "User not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, user)
}

// ListUsers handles GET /users
// Returns a paginated list of users.
// Supports query parameters: limit (default: 10) and offset (default: 0).
// Returns 200 with an array of users on success.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters from query string
	params := httputil.ParsePaginationParams(r)

	users, err := h.repo.List(r.Context(), params.Limit, params.Offset)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list users: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, users)
}

// UpdateUser handles PUT /users/{id}
// Updates an existing user with the provided fields.
// Only non-zero/non-empty fields in the request will be updated.
// Returns 400 if the ID format or request body is invalid.
// Returns 404 if the user is not found.
// Returns 200 with the updated user on success.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDFromPath(r.URL.Path, "/users/")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.RespondError(w, http.StatusNotFound, "User not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update user: %v", err))
		return
	}

	httputil.RespondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}
// Permanently deletes a user from the system.
// Returns 400 if the ID format is invalid.
// Returns 404 if the user is not found.
// Returns 204 (No Content) on successful deletion.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDFromPath(r.URL.Path, "/users/")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.RespondError(w, http.StatusNotFound, "User not found")
			return
		}
		httputil.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete user: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
