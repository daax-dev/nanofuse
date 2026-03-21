package types

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents an API error code
type ErrorCode string

const (
	ErrInvalidRequest         ErrorCode = "INVALID_REQUEST"
	ErrInvalidConfig          ErrorCode = "INVALID_CONFIG"
	ErrRegistryAuthFailed     ErrorCode = "REGISTRY_AUTH_FAILED"
	ErrVMNotFound             ErrorCode = "VM_NOT_FOUND"
	ErrImageNotFound          ErrorCode = "IMAGE_NOT_FOUND"
	ErrSnapshotNotFound       ErrorCode = "SNAPSHOT_NOT_FOUND"
	ErrRecordingNotFound      ErrorCode = "RECORDING_NOT_FOUND"
	ErrInvalidStateTransition ErrorCode = "INVALID_STATE_TRANSITION"
	ErrVMLocked               ErrorCode = "VM_LOCKED"
	ErrResourceInUse          ErrorCode = "RESOURCE_IN_USE"
	ErrValidationError        ErrorCode = "VALIDATION_ERROR"
	ErrResourceLimitExceeded  ErrorCode = "RESOURCE_LIMIT_EXCEEDED"
	ErrSnapshotIncompatible   ErrorCode = "SNAPSHOT_INCOMPATIBLE"
	ErrInternalError          ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable     ErrorCode = "SERVICE_UNAVAILABLE"
	ErrInsufficientStorage    ErrorCode = "INSUFFICIENT_STORAGE"
	ErrNotFound               ErrorCode = "NOT_FOUND"
)

// APIError represents an API error response
type APIError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, details map[string]interface{}) *APIError {
	return &APIError{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, statusCode int, code ErrorCode, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// We can't return an error after WriteHeader, so ignore encoding errors
	_ = json.NewEncoder(w).Encode(NewAPIError(code, message, details))
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status        string `json:"status"`
	Version       string `json:"version"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}
