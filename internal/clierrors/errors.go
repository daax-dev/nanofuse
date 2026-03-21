// Package clierrors provides user-friendly CLI error handling with actionable suggestions.
//
// The package provides structured error types that include:
//   - Clear, jargon-free error messages
//   - Suggested fixes for common error scenarios
//   - References to relevant documentation or commands
//   - Contextual information (operation, resource, reason)
//   - Consistent formatting without stack traces
package clierrors

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jpoley/nanofuse/internal/client"
)

// CLIError represents a user-friendly CLI error with actionable suggestions.
type CLIError struct {
	// Message is the primary error message shown to the user.
	Message string

	// Suggestion provides actionable guidance on how to resolve the error.
	Suggestion string

	// Context provides additional contextual information about the error.
	Context *ErrorContext

	// DocRef is an optional reference to documentation or help command.
	DocRef string

	// ExitCode is the exit code to use when exiting the CLI.
	ExitCode int

	// wrapped is the underlying error if this error wraps another.
	wrapped error
}

// ErrorContext provides structured context about where an error occurred.
type ErrorContext struct {
	// Operation is the operation being performed (e.g., "start VM", "pull image").
	Operation string

	// Resource is the resource being operated on (e.g., VM ID, image reference).
	Resource string

	// Reason provides additional details about why the operation failed.
	Reason string
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error.
func (e *CLIError) Unwrap() error {
	return e.wrapped
}

// Format writes the error to stderr in a user-friendly format.
// If useColor is true, the output will include ANSI color codes.
// All output goes to stderr regardless of color mode for consistent error output.
func (e *CLIError) Format(useColor bool) {
	// Create color printers that output to stderr
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)

	// Print the main error
	if useColor {
		red.Fprintf(os.Stderr, "Error: %s\n", e.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", e.Message)
	}

	// Print context if available
	if e.Context != nil {
		fmt.Fprintln(os.Stderr)
		if e.Context.Operation != "" {
			fmt.Fprintf(os.Stderr, "  Operation: %s\n", e.Context.Operation)
		}
		if e.Context.Resource != "" {
			fmt.Fprintf(os.Stderr, "  Resource:  %s\n", e.Context.Resource)
		}
		if e.Context.Reason != "" {
			fmt.Fprintf(os.Stderr, "  Reason:    %s\n", e.Context.Reason)
		}
	}

	// Print suggestion
	if e.Suggestion != "" {
		fmt.Fprintln(os.Stderr)
		if useColor {
			yellow.Fprintf(os.Stderr, "Suggestion: %s\n", e.Suggestion)
		} else {
			fmt.Fprintf(os.Stderr, "Suggestion: %s\n", e.Suggestion)
		}
	}

	// Print documentation reference
	if e.DocRef != "" {
		fmt.Fprintln(os.Stderr)
		if useColor {
			fmt.Fprintf(os.Stderr, "See: %s\n", cyan.Sprint(e.DocRef))
		} else {
			fmt.Fprintf(os.Stderr, "See: %s\n", e.DocRef)
		}
	}
}

// Error code string constants for common error types.
const (
	CodeVMNotFound             = "VM_NOT_FOUND"
	CodeImageNotFound          = "IMAGE_NOT_FOUND"
	CodeSnapshotNotFound       = "SNAPSHOT_NOT_FOUND"
	CodeDaemonNotRunning       = "API_UNREACHABLE"
	CodeInvalidImageRef        = "INVALID_IMAGE_REF"
	CodeMissingArgument        = "MISSING_ARGUMENT"
	CodeInvalidStateTransition = "INVALID_STATE_TRANSITION"
	CodeVMLocked               = "VM_LOCKED"
	CodeAuthFailed             = "REGISTRY_AUTH_FAILED"
	CodeValidationError        = "VALIDATION_ERROR"
)

// WrapVMNotFound creates a CLIError for a VM not found error.
func WrapVMNotFound(vmID string, operation string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("VM '%s' not found", vmID),
		Suggestion: "Run 'nanofuse vm list' to see available VMs",
		Context: &ErrorContext{
			Operation: operation,
			Resource:  vmID,
			Reason:    "The specified VM does not exist or has been deleted",
		},
		DocRef:   "nanofuse vm list --help",
		ExitCode: 4,
	}
}

// WrapImageNotFound creates a CLIError for an image not found error.
func WrapImageNotFound(imageRef string, operation string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("Image '%s' not found in local cache", imageRef),
		Suggestion: fmt.Sprintf("Pull the image first: nanofuse image pull %s", imageRef),
		Context: &ErrorContext{
			Operation: operation,
			Resource:  imageRef,
			Reason:    "The image is not cached locally and needs to be pulled from the registry",
		},
		DocRef:   "nanofuse image pull --help",
		ExitCode: 4,
	}
}

// WrapSnapshotNotFound creates a CLIError for a snapshot not found error.
func WrapSnapshotNotFound(snapshotID string, operation string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("Snapshot '%s' not found", snapshotID),
		Suggestion: "Run 'nanofuse vm snapshot list <vm-id>' to see available snapshots",
		Context: &ErrorContext{
			Operation: operation,
			Resource:  snapshotID,
			Reason:    "The specified snapshot does not exist or has been deleted",
		},
		DocRef:   "nanofuse vm snapshot list --help",
		ExitCode: 4,
	}
}

// WrapDaemonNotRunning creates a CLIError when the nanofused daemon is not reachable.
func WrapDaemonNotRunning(socketPath string) *CLIError {
	return &CLIError{
		Message:    "Cannot connect to nanofused API",
		Suggestion: "Start the nanofused service: sudo systemctl start nanofused",
		Context: &ErrorContext{
			Operation: "connect to API",
			Resource:  socketPath,
			Reason:    "The nanofused daemon is not running or the socket is inaccessible",
		},
		DocRef:   "sudo systemctl status nanofused",
		ExitCode: 3,
	}
}

// WrapInvalidImageRef creates a CLIError for an invalid image reference.
func WrapInvalidImageRef(imageRef string, reason string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("Invalid image reference: %s", imageRef),
		Suggestion: "Use format: registry/repository:tag (e.g., ghcr.io/jpoley/nanofuse/base:latest)",
		Context: &ErrorContext{
			Operation: "parse image reference",
			Resource:  imageRef,
			Reason:    reason,
		},
		DocRef:   "nanofuse image pull --help",
		ExitCode: 2,
	}
}

// WrapMissingArgument creates a CLIError for a missing required argument.
func WrapMissingArgument(argName string, command string, example string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("Missing required argument: %s", argName),
		Suggestion: fmt.Sprintf("Example: %s", example),
		Context: &ErrorContext{
			Operation: command,
			Reason:    fmt.Sprintf("The '%s' argument is required but was not provided", argName),
		},
		DocRef:   fmt.Sprintf("nanofuse %s --help", command),
		ExitCode: 2,
	}
}

// WrapInvalidStateTransition creates a CLIError for an invalid VM state transition.
func WrapInvalidStateTransition(vmID string, operation string, currentState string, requiredStates []string) *CLIError {
	var suggestion string
	if len(requiredStates) == 0 {
		// Empty requiredStates means the operation is not allowed in any state
		suggestion = fmt.Sprintf("This operation is not allowed in the current state. Check status: nanofuse vm status %s", vmID)
	} else {
		statesStr := strings.Join(requiredStates, "' or '")
		suggestion = fmt.Sprintf("VM must be in '%s' state. Check status: nanofuse vm status %s", statesStr, vmID)
	}
	return &CLIError{
		Message:    fmt.Sprintf("Cannot %s VM in '%s' state", operation, currentState),
		Suggestion: suggestion,
		Context: &ErrorContext{
			Operation: operation,
			Resource:  vmID,
			Reason:    fmt.Sprintf("Current state '%s' does not allow this operation", currentState),
		},
		DocRef:   "nanofuse vm status --help",
		ExitCode: 5,
	}
}

// WrapVMLocked creates a CLIError when a VM is locked by another operation.
func WrapVMLocked(vmID string, operation string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("VM '%s' is locked by another operation", vmID),
		Suggestion: "Wait for the current operation to complete and try again",
		Context: &ErrorContext{
			Operation: operation,
			Resource:  vmID,
			Reason:    "Another operation is currently in progress on this VM",
		},
		DocRef:   fmt.Sprintf("nanofuse vm status %s", vmID),
		ExitCode: 5,
	}
}

// WrapAuthFailed creates a CLIError for registry authentication failures.
func WrapAuthFailed(registry string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("Authentication failed for registry: %s", registry),
		Suggestion: fmt.Sprintf("Run 'docker login %s' to authenticate", registry),
		Context: &ErrorContext{
			Operation: "authenticate",
			Resource:  registry,
			Reason:    "Invalid or missing credentials for the container registry",
		},
		DocRef:   "docker login --help",
		ExitCode: 2,
	}
}

// WrapValidationError creates a CLIError for validation failures.
func WrapValidationError(field string, reason string) *CLIError {
	return &CLIError{
		Message:    fmt.Sprintf("Invalid value for '%s': %s", field, reason),
		Suggestion: "Check the command help for valid values",
		Context: &ErrorContext{
			Operation: "validate input",
			Reason:    reason,
		},
		ExitCode: 2,
	}
}

// extractRequiredStates parses required_states from error details.
// Supports []string, []interface{}, and comma-separated string formats.
//
// Behavior:
//   - Key not present or nil: returns fallback ["running", "stopped"]
//   - Empty slice explicitly provided: returns empty slice (API says no valid states)
//   - Non-empty slice: returns the provided states
//   - Unsupported types (int, bool, etc.): returns fallback silently
//
// Note: Unsupported types silently fall back to default states rather than failing.
// This is intentional to maintain CLI usability even if the API contract changes.
// The API should ideally always provide the actual required states.
func extractRequiredStates(details map[string]interface{}) []string {
	// Fallback used when API doesn't specify required states or sends an unsupported type.
	// Most VM lifecycle operations need running or stopped state.
	fallbackStates := []string{"running", "stopped"}
	raw, ok := details["required_states"]
	if !ok || raw == nil {
		return fallbackStates
	}

	switch v := raw.(type) {
	case []string:
		// Return as-is, even if empty (API explicitly provided empty slice)
		return v
	case []interface{}:
		// Convert to []string, preserving empty slice if that's what API sent
		rs := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				rs = append(rs, s)
			}
		}
		return rs
	case string:
		if v == "" {
			return []string{} // Explicitly empty string means no valid states
		}
		parts := strings.Split(v, ",")
		var rs []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				rs = append(rs, p)
			}
		}
		if len(rs) == 0 {
			return []string{} // All parts were empty/whitespace
		}
		return rs
	}
	return fallbackStates
}

// FromClientError converts a client.ClientError to a user-friendly CLIError.
// The operation parameter provides context about what was being attempted.
func FromClientError(err *client.ClientError, operation string, resource string) *CLIError {
	switch err.Code {
	case CodeVMNotFound:
		return WrapVMNotFound(resource, operation)

	case CodeImageNotFound:
		return WrapImageNotFound(resource, operation)

	case CodeSnapshotNotFound:
		return WrapSnapshotNotFound(resource, operation)

	case CodeDaemonNotRunning:
		return WrapDaemonNotRunning(resource)

	case CodeInvalidStateTransition:
		// Default to "unknown" if current_state is missing or not a string.
		// Silent fallback is intentional for CLI resilience - we don't want to
		// spam users with warnings about API contract issues. The error message
		// will still be actionable even with "unknown" state.
		currentState := "unknown"
		if state, ok := err.Details["current_state"].(string); ok {
			currentState = state
		}
		requiredStates := extractRequiredStates(err.Details)
		return WrapInvalidStateTransition(resource, operation, currentState, requiredStates)

	case CodeVMLocked:
		return WrapVMLocked(resource, operation)

	case CodeAuthFailed:
		return WrapAuthFailed(resource)

	case CodeValidationError:
		return &CLIError{
			Message:    err.Message,
			Suggestion: "Check the command help for valid input format",
			Context: &ErrorContext{
				Operation: operation,
				Resource:  resource,
			},
			ExitCode: 2,
			wrapped:  err,
		}

	default:
		// Generic error for unknown codes
		return &CLIError{
			Message: err.Message,
			Context: &ErrorContext{
				Operation: operation,
				Resource:  resource,
			},
			ExitCode: err.ExitCode(),
			wrapped:  err,
		}
	}
}

// FromError converts a generic error to a CLIError, detecting common patterns.
// The operation and resource parameters provide context about what was being attempted.
func FromError(err error, operation string, resource string) *CLIError {
	// Check if it's already a ClientError
	if cerr, ok := err.(*client.ClientError); ok {
		return FromClientError(cerr, operation, resource)
	}

	errMsg := err.Error()

	// Detect connection refused (daemon not running)
	if IsConnectionError(errMsg) {
		socketPath := resource
		if socketPath == "" {
			socketPath = "/run/nanofused.sock"
		}
		return WrapDaemonNotRunning(socketPath)
	}

	// Detect authentication errors
	if IsAuthError(errMsg) {
		registry := extractRegistry(resource)
		return WrapAuthFailed(registry)
	}

	// Generic error
	return &CLIError{
		Message: err.Error(),
		Context: &ErrorContext{
			Operation: operation,
			Resource:  resource,
		},
		ExitCode: 1,
		wrapped:  err,
	}
}

// IsConnectionError checks if an error message indicates a connection failure.
// It specifically looks for socket/dial-related errors to avoid false positives
// from general file-not-found errors (e.g., config files).
func IsConnectionError(errMsg string) bool {
	lower := strings.ToLower(errMsg)
	// Match common connection error patterns:
	// - "connection refused" - server not listening
	// - "dial unix" - Unix socket issues (covers "dial unix /path: no such file...")
	// - "dial tcp" - TCP connection issues
	// - "connect: no such file" - short form socket error
	// Note: We don't match generic "no such file or directory" as it could be config files
	return strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "dial unix") ||
		strings.Contains(lower, "dial tcp") ||
		strings.Contains(lower, "connect: no such file") ||
		strings.Contains(lower, "connect: connection refused")
}

// IsAuthError checks if an error message indicates an authentication failure.
func IsAuthError(errMsg string) bool {
	lower := strings.ToLower(errMsg)
	return strings.Contains(lower, "authentication") ||
		strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "forbidden") ||
		strings.Contains(lower, "401") ||
		strings.Contains(lower, "403")
}

// IsNotFoundError checks if an error message indicates a resource not found.
func IsNotFoundError(errMsg string) bool {
	lower := strings.ToLower(errMsg)
	return strings.Contains(lower, "not found") ||
		strings.Contains(lower, "404") ||
		strings.Contains(lower, "does not exist")
}

// extractRegistry attempts to extract a registry name from an image reference.
// Docker Hub is the default for images without an explicit registry hostname.
func extractRegistry(imageRef string) string {
	if imageRef == "" {
		return "container registry"
	}

	// Check for common registry patterns
	parts := strings.Split(imageRef, "/")
	if len(parts) > 1 {
		// There's at least one "/" - the first part could be a registry
		first := parts[0]
		// If it contains a dot, treat it as a registry hostname (optionally with a port),
		// e.g., ghcr.io, docker.io, my.registry:5000
		if strings.Contains(first, ".") {
			return first
		}
		// If it has a colon but no dot, check if it looks like a port (localhost:5000)
		// Only do this check when there's a "/" because "myimage:5000" is image:tag, not registry
		if strings.Contains(first, ":") {
			// Check if the part after colon is a valid port (not a tag)
			colonParts := strings.Split(first, ":")
			if len(colonParts) == 2 && isValidPort(colonParts[1]) {
				return first // localhost:5000
			}
		}
	}

	// Default to Docker Hub for images without explicit registry:
	// - "nginx" (single part)
	// - "library/nginx" (library prefix)
	// - "myorg/myimage" (org/image format without registry hostname)
	return "docker.io"
}

// isValidPort checks if a string is a valid TCP/UDP port number (1-65535).
// It accepts leading zeros (e.g., "00080" is valid as port 80).
func isValidPort(s string) bool {
	if len(s) == 0 {
		return false
	}
	port := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		port = port*10 + int(c-'0')
		// Early exit if port exceeds maximum to prevent overflow on very long strings
		if port > 65535 {
			return false
		}
	}
	return port >= 1
}

// HandleError is a convenience function that formats and handles an error.
// It returns the exit code that should be used.
func HandleError(err error, operation string, resource string, useColor bool) int {
	cliErr := FromError(err, operation, resource)
	cliErr.Format(useColor)
	return cliErr.ExitCode
}
