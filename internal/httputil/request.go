package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ParseIDFromPath extracts and parses an ID from a URL path.
// It removes the specified prefix and parses the remaining part as an int64.
//
// Example:
//
//	id, err := ParseIDFromPath(r.URL.Path, "/users/")
//	// For path "/users/123", returns 123
func ParseIDFromPath(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)

	// Remove any trailing path segments (e.g., "/users/123/rooms" -> "123")
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format: %w", err)
	}

	return id, nil
}

// PaginationParams holds pagination parameters for list queries
type PaginationParams struct {
	Limit  int
	Offset int
}

// ParsePaginationParams extracts limit and offset from query parameters.
// If not provided, defaults to limit=10 and offset=0.
// Invalid values are ignored and defaults are used instead.
//
// Example:
//
//	params := ParsePaginationParams(r)
//	users, err := repo.List(ctx, params.Limit, params.Offset)
func ParsePaginationParams(r *http.Request) PaginationParams {
	params := PaginationParams{
		Limit:  10, // Default limit
		Offset: 0,  // Default offset
	}

	// Parse limit parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			params.Limit = l
		}
	}

	// Parse offset parameter
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			params.Offset = o
		}
	}

	return params
}

// DecodeJSONBody decodes the request body into the provided destination.
// Returns an error if the body is not valid JSON or cannot be decoded.
//
// Example:
//
//	var req CreateUserRequest
//	if err := DecodeJSONBody(r, &req); err != nil {
//	    RespondError(w, http.StatusBadRequest, "Invalid request body")
//	    return
//	}
func DecodeJSONBody(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}
	return nil
}

// PathSegments extracts path segments after removing a prefix.
// Useful for parsing complex nested routes like /rooms/{id}/users/{userId}.
//
// Example:
//
//	segments := PathSegments(r.URL.Path, "/rooms/")
//	// For path "/rooms/123/users/456", returns ["123", "users", "456"]
func PathSegments(path, prefix string) []string {
	trimmed := strings.TrimPrefix(path, prefix)
	return strings.Split(trimmed, "/")
}
