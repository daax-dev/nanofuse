package types

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents an API error code
type ErrorCode string

const (
	ErrInvalidRequest         ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized           ErrorCode = "UNAUTHORIZED"
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

// CapabilitiesResponse describes the daemon host and runtime capabilities.
type CapabilitiesResponse struct {
	Status  string                   `json:"status"`
	Version string                   `json:"version"`
	Host    HostCapabilities         `json:"host"`
	Runtime RuntimeCapabilities      `json:"runtime"`
	API     APITransportCapabilities `json:"api"`
}

// HostCapabilities describes host-level platform support.
type HostCapabilities struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	KVMDevice    string `json:"kvm_device"`
	KVMExists    bool   `json:"kvm_exists"`
	KVMReadWrite bool   `json:"kvm_read_write"`
	KVMError     string `json:"kvm_error,omitempty"`
}

// RuntimeCapabilities describes the microVM runtime available to nanofused.
type RuntimeCapabilities struct {
	NativeRuntime        bool   `json:"native_runtime"`
	FirecrackerBinary    string `json:"firecracker_binary"`
	FirecrackerAvailable bool   `json:"firecracker_available"`
	RootRequired         bool   `json:"root_required"`
	NetworkSetupRequired bool   `json:"network_setup_required"`
	Message              string `json:"message"`
}

// APITransportCapabilities describes how clients can reach the daemon.
type APITransportCapabilities struct {
	UnixSocket string `json:"unix_socket,omitempty"`
	TCPBind    string `json:"tcp_bind,omitempty"`
}
