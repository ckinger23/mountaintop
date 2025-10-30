package validation

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Message string
	Code    string
	Details map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]string) *ValidationError {
	return &ValidationError{
		Message: message,
		Code:    "VALIDATION_ERROR",
		Details: details,
	}
}

// RespondWithError sends a structured error response
func RespondWithError(w http.ResponseWriter, statusCode int, message string, code string, details map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(w http.ResponseWriter, err *ValidationError) {
	RespondWithError(w, http.StatusBadRequest, err.Message, err.Code, err.Details)
}
