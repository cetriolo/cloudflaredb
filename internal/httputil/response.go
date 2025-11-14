package httputil

import (
	"encoding/json"
	"net/http"
)

// RespondJSON sends a JSON response with the given status code and data.
// It automatically sets the Content-Type header to application/json.
//
// Example:
//
//	RespondJSON(w, http.StatusOK, user)
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RespondError sends a standardized JSON error response.
// The error is wrapped in an object with an "error" field.
//
// Example:
//
//	RespondError(w, http.StatusBadRequest, "Invalid input")
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{"error": message})
}

// ErrorResponse represents a standardized error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}
