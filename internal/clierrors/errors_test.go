package clierrors

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jpoley/nanofuse/internal/client"
)

func TestCLIError_Error(t *testing.T) {
	err := &CLIError{
		Message: "test error message",
	}

	if got := err.Error(); got != "test error message" {
		t.Errorf("Error() = %q, want %q", got, "test error message")
	}
}

func TestCLIError_Unwrap(t *testing.T) {
	wrapped := errors.New("wrapped error")
	err := &CLIError{
		Message: "cli error",
		wrapped: wrapped,
	}

	if got := err.Unwrap(); got != wrapped {
		t.Errorf("Unwrap() = %v, want %v", got, wrapped)
	}
}

func TestCLIError_Format(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		useColor bool
		wantMsg  []string // Substrings that should appear in output
	}{
		{
			name: "basic error",
			err: &CLIError{
				Message: "Something went wrong",
			},
			useColor: false,
			wantMsg:  []string{"Error: Something went wrong"},
		},
		{
			name: "error with suggestion",
			err: &CLIError{
				Message:    "VM not found",
				Suggestion: "Run 'nanofuse vm list'",
			},
			useColor: false,
			wantMsg:  []string{"Error: VM not found", "Suggestion: Run 'nanofuse vm list'"},
		},
		{
			name: "error with context",
			err: &CLIError{
				Message: "Failed to start",
				Context: &ErrorContext{
					Operation: "start VM",
					Resource:  "my-vm",
					Reason:    "VM is already running",
				},
			},
			useColor: false,
			wantMsg:  []string{"Error: Failed to start", "Operation: start VM", "Resource:  my-vm", "Reason:    VM is already running"},
		},
		{
			name: "error with doc ref",
			err: &CLIError{
				Message: "Invalid command",
				DocRef:  "nanofuse --help",
			},
			useColor: false,
			wantMsg:  []string{"Error: Invalid command", "See: nanofuse --help"},
		},
		{
			name: "full error",
			err: &CLIError{
				Message:    "VM 'test-vm' not found",
				Suggestion: "Run 'nanofuse vm list' to see available VMs",
				Context: &ErrorContext{
					Operation: "get VM status",
					Resource:  "test-vm",
					Reason:    "The VM does not exist",
				},
				DocRef: "nanofuse vm list --help",
			},
			useColor: false,
			wantMsg: []string{
				"Error: VM 'test-vm' not found",
				"Operation: get VM status",
				"Resource:  test-vm",
				"Reason:    The VM does not exist",
				"Suggestion: Run 'nanofuse vm list' to see available VMs",
				"See: nanofuse vm list --help",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr output
			old := os.Stderr
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stderr = w

			tt.err.Format(tt.useColor)

			w.Close()
			os.Stderr = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			for _, want := range tt.wantMsg {
				if !strings.Contains(output, want) {
					t.Errorf("Format() output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestWrapVMNotFound(t *testing.T) {
	err := WrapVMNotFound("my-vm", "start VM")

	if !strings.Contains(err.Message, "my-vm") {
		t.Errorf("Message should contain VM ID, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "nanofuse vm list") {
		t.Errorf("Suggestion should mention vm list command, got: %s", err.Suggestion)
	}
	if err.Context == nil || err.Context.Operation != "start VM" {
		t.Errorf("Context.Operation should be 'start VM'")
	}
	if err.ExitCode != 4 {
		t.Errorf("ExitCode = %d, want 4", err.ExitCode)
	}
}

func TestWrapImageNotFound(t *testing.T) {
	err := WrapImageNotFound("ghcr.io/test/image:v1", "create VM")

	if !strings.Contains(err.Message, "ghcr.io/test/image:v1") {
		t.Errorf("Message should contain image ref, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "nanofuse image pull") {
		t.Errorf("Suggestion should mention image pull, got: %s", err.Suggestion)
	}
	if err.ExitCode != 4 {
		t.Errorf("ExitCode = %d, want 4", err.ExitCode)
	}
}

func TestWrapSnapshotNotFound(t *testing.T) {
	err := WrapSnapshotNotFound("snap-123", "resume VM")

	if !strings.Contains(err.Message, "snap-123") {
		t.Errorf("Message should contain snapshot ID, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "snapshot list") {
		t.Errorf("Suggestion should mention snapshot list, got: %s", err.Suggestion)
	}
	if err.ExitCode != 4 {
		t.Errorf("ExitCode = %d, want 4", err.ExitCode)
	}
}

func TestWrapDaemonNotRunning(t *testing.T) {
	err := WrapDaemonNotRunning("/run/nanofused.sock")

	if !strings.Contains(err.Message, "connect") {
		t.Errorf("Message should mention connection, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "systemctl start nanofused") {
		t.Errorf("Suggestion should mention starting service, got: %s", err.Suggestion)
	}
	if err.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", err.ExitCode)
	}
}

func TestWrapInvalidImageRef(t *testing.T) {
	err := WrapInvalidImageRef("invalid:ref:format", "missing repository")

	if !strings.Contains(err.Message, "invalid:ref:format") {
		t.Errorf("Message should contain image ref, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "registry/repository:tag") {
		t.Errorf("Suggestion should show correct format, got: %s", err.Suggestion)
	}
	if err.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", err.ExitCode)
	}
}

func TestWrapMissingArgument(t *testing.T) {
	err := WrapMissingArgument("vm-id", "vm start", "nanofuse vm start my-vm")

	if !strings.Contains(err.Message, "vm-id") {
		t.Errorf("Message should contain argument name, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "nanofuse vm start my-vm") {
		t.Errorf("Suggestion should contain example, got: %s", err.Suggestion)
	}
	if err.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", err.ExitCode)
	}
}

func TestWrapInvalidStateTransition(t *testing.T) {
	err := WrapInvalidStateTransition("my-vm", "start", "running", []string{"stopped", "paused"})

	if !strings.Contains(err.Message, "running") {
		t.Errorf("Message should contain current state, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "stopped") || !strings.Contains(err.Suggestion, "paused") {
		t.Errorf("Suggestion should mention required states, got: %s", err.Suggestion)
	}
	if err.ExitCode != 5 {
		t.Errorf("ExitCode = %d, want 5", err.ExitCode)
	}
}

func TestWrapVMLocked(t *testing.T) {
	err := WrapVMLocked("my-vm", "delete")

	if !strings.Contains(err.Message, "my-vm") {
		t.Errorf("Message should contain VM ID, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "Wait") {
		t.Errorf("Suggestion should tell user to wait, got: %s", err.Suggestion)
	}
	if err.ExitCode != 5 {
		t.Errorf("ExitCode = %d, want 5", err.ExitCode)
	}
}

func TestWrapAuthFailed(t *testing.T) {
	err := WrapAuthFailed("ghcr.io")

	if !strings.Contains(err.Message, "ghcr.io") {
		t.Errorf("Message should contain registry, got: %s", err.Message)
	}
	if !strings.Contains(err.Suggestion, "docker login") {
		t.Errorf("Suggestion should mention docker login, got: %s", err.Suggestion)
	}
	if err.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", err.ExitCode)
	}
}

func TestWrapValidationError(t *testing.T) {
	err := WrapValidationError("vcpus", "must be between 1 and 32")

	if !strings.Contains(err.Message, "vcpus") {
		t.Errorf("Message should contain field name, got: %s", err.Message)
	}
	if !strings.Contains(err.Message, "must be between 1 and 32") {
		t.Errorf("Message should contain reason, got: %s", err.Message)
	}
	if err.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", err.ExitCode)
	}
}

func TestFromClientError(t *testing.T) {
	tests := []struct {
		name         string
		clientErr    *client.ClientError
		operation    string
		resource     string
		wantMsgPart  string
		wantExitCode int
	}{
		{
			name: "VM not found",
			clientErr: &client.ClientError{
				StatusCode: 404,
				Code:       "VM_NOT_FOUND",
				Message:    "VM not found",
			},
			operation:    "start VM",
			resource:     "test-vm",
			wantMsgPart:  "test-vm",
			wantExitCode: 4,
		},
		{
			name: "Image not found",
			clientErr: &client.ClientError{
				StatusCode: 404,
				Code:       "IMAGE_NOT_FOUND",
				Message:    "Image not found",
			},
			operation:    "create VM",
			resource:     "ghcr.io/test/image:v1",
			wantMsgPart:  "ghcr.io/test/image:v1",
			wantExitCode: 4,
		},
		{
			name: "Snapshot not found",
			clientErr: &client.ClientError{
				StatusCode: 404,
				Code:       "SNAPSHOT_NOT_FOUND",
				Message:    "Snapshot not found",
			},
			operation:    "resume VM",
			resource:     "snap-123",
			wantMsgPart:  "snap-123",
			wantExitCode: 4,
		},
		{
			name: "VM locked",
			clientErr: &client.ClientError{
				StatusCode: 409,
				Code:       "VM_LOCKED",
				Message:    "VM is locked",
			},
			operation:    "delete VM",
			resource:     "busy-vm",
			wantMsgPart:  "busy-vm",
			wantExitCode: 5,
		},
		{
			name: "Auth failed",
			clientErr: &client.ClientError{
				StatusCode: 401,
				Code:       "REGISTRY_AUTH_FAILED",
				Message:    "Authentication failed",
			},
			operation:    "pull image",
			resource:     "ghcr.io",
			wantMsgPart:  "ghcr.io",
			wantExitCode: 2,
		},
		{
			name: "Unknown error",
			clientErr: &client.ClientError{
				StatusCode: 500,
				Code:       "INTERNAL_ERROR",
				Message:    "Something went wrong",
			},
			operation:    "unknown",
			resource:     "resource",
			wantMsgPart:  "Something went wrong",
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliErr := FromClientError(tt.clientErr, tt.operation, tt.resource)

			if !strings.Contains(cliErr.Message, tt.wantMsgPart) {
				t.Errorf("Message = %q, want to contain %q", cliErr.Message, tt.wantMsgPart)
			}
			if cliErr.ExitCode != tt.wantExitCode {
				t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, tt.wantExitCode)
			}
		})
	}
}

func TestFromError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		operation    string
		resource     string
		wantMsgPart  string
		wantExitCode int
	}{
		{
			name:         "connection refused",
			err:          errors.New("dial unix /run/nanofused.sock: connect: connection refused"),
			operation:    "health check",
			resource:     "/run/nanofused.sock",
			wantMsgPart:  "connect",
			wantExitCode: 3,
		},
		{
			name:         "no such file (socket)",
			err:          errors.New("dial unix /run/nanofused.sock: connect: no such file or directory"),
			operation:    "health check",
			resource:     "/run/nanofused.sock",
			wantMsgPart:  "connect",
			wantExitCode: 3,
		},
		{
			name:         "unauthorized error",
			err:          errors.New("unauthorized: authentication required"),
			operation:    "pull image",
			resource:     "ghcr.io/test/image:v1",
			wantMsgPart:  "Authentication",
			wantExitCode: 2,
		},
		{
			name:         "403 forbidden",
			err:          errors.New("HTTP 403: forbidden"),
			operation:    "pull image",
			resource:     "ghcr.io/private/image:v1",
			wantMsgPart:  "Authentication",
			wantExitCode: 2,
		},
		{
			name:         "generic error",
			err:          errors.New("something unexpected happened"),
			operation:    "test operation",
			resource:     "test-resource",
			wantMsgPart:  "something unexpected happened",
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliErr := FromError(tt.err, tt.operation, tt.resource)

			if !strings.Contains(cliErr.Message, tt.wantMsgPart) {
				t.Errorf("Message = %q, want to contain %q", cliErr.Message, tt.wantMsgPart)
			}
			if cliErr.ExitCode != tt.wantExitCode {
				t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, tt.wantExitCode)
			}
		})
	}
}

func TestFromError_WithClientError(t *testing.T) {
	clientErr := &client.ClientError{
		StatusCode: 404,
		Code:       "VM_NOT_FOUND",
		Message:    "VM not found",
	}

	cliErr := FromError(clientErr, "start VM", "test-vm")

	if !strings.Contains(cliErr.Message, "test-vm") {
		t.Errorf("Should handle ClientError, got message: %s", cliErr.Message)
	}
	if cliErr.ExitCode != 4 {
		t.Errorf("ExitCode = %d, want 4", cliErr.ExitCode)
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		errMsg string
		want   bool
	}{
		{"dial unix /run/nanofused.sock: connect: connection refused", true},
		{"dial unix /run/nanofused.sock: no such file or directory", true},
		{"dial unix: missing address", true},
		{"dial tcp 127.0.0.1:8080: connect: connection refused", true},
		{"connect: no such file", true},
		{"connect: connection refused", true},
		{"no such file or directory", false},                                                        // Generic file error
		{"config failed: /etc/app/config.yaml: no such file or directory", false},                   // Config file error
		{"connect to database failed: config file /path/to/file: no such file or directory", false}, // False positive test
		{"VM not found", false},
		{"authentication failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			if got := IsConnectionError(tt.errMsg); got != tt.want {
				t.Errorf("IsConnectionError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		errMsg string
		want   bool
	}{
		{"unauthorized: authentication required", true},
		{"401 Unauthorized", true},
		{"HTTP 403 Forbidden", true},
		{"authentication failed", true},
		{"VM not found", false},
		{"connection refused", false},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			if got := IsAuthError(tt.errMsg); got != tt.want {
				t.Errorf("IsAuthError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		errMsg string
		want   bool
	}{
		{"VM not found", true},
		{"HTTP 404", true},
		{"resource does not exist", true},
		{"connection refused", false},
		{"unauthorized", false},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			if got := IsNotFoundError(tt.errMsg); got != tt.want {
				t.Errorf("IsNotFoundError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestExtractRegistry(t *testing.T) {
	tests := []struct {
		imageRef string
		want     string
	}{
		{"ghcr.io/user/repo:tag", "ghcr.io"},
		{"docker.io/library/alpine", "docker.io"},
		{"localhost:5000/myimage", "localhost:5000"},
		{"myimage:latest", "docker.io"},
		{"myimage:5000", "docker.io"}, // numeric tag; no "/" so ":" is treated as a tag separator, not a port, and registry defaults to docker.io
		{"library/nginx", "docker.io"},
		{"myorg/myimage", "docker.io"},      // org/image without registry defaults to docker.io
		{"myorg/myimage:v1.0", "docker.io"}, // org/image:tag without registry defaults to docker.io
		{"nginx", "docker.io"},              // simple image name
		{"", "container registry"},
	}

	for _, tt := range tests {
		t.Run(tt.imageRef, func(t *testing.T) {
			if got := extractRegistry(tt.imageRef); got != tt.want {
				t.Errorf("extractRegistry(%q) = %q, want %q", tt.imageRef, got, tt.want)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create pipe: %v", pipeErr)
	}
	os.Stderr = w

	err := errors.New("dial unix /run/nanofused.sock: connect: connection refused")
	exitCode := HandleError(err, "health check", "/run/nanofused.sock", false)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if exitCode != 3 {
		t.Errorf("HandleError() exit code = %d, want 3", exitCode)
	}

	output := buf.String()
	if !strings.Contains(output, "Error:") {
		t.Errorf("HandleError() should print error message, got: %s", output)
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// Valid ports
		{"1", true},     // Minimum valid port
		{"80", true},    // Common HTTP port
		{"443", true},   // Common HTTPS port
		{"5000", true},  // Common registry port
		{"8080", true},  // Common alt HTTP port
		{"65535", true}, // Maximum valid port

		// Invalid ports - boundary conditions
		{"0", false},      // Below minimum
		{"65536", false},  // Above maximum
		{"99999", false},  // Way above maximum
		{"100000", false}, // Six digits, definitely invalid

		// Invalid ports - non-numeric
		{"", false},      // Empty string
		{"abc", false},   // Letters
		{"123a", false},  // Mixed
		{"12.34", false}, // Contains dot
		{"-1", false},    // Negative (contains non-digit)
		{"5000:", false}, // Trailing colon
		{":5000", false}, // Leading colon

		// Edge cases - leading zeros are accepted
		{"00080", true},  // Leading zeros, still valid (80)
		{"01", true},     // Leading zero, valid (1)
		{"000001", true}, // Many leading zeros, still valid (1)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isValidPort(tt.input); got != tt.want {
				t.Errorf("isValidPort(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractRequiredStates(t *testing.T) {
	tests := []struct {
		name    string
		details map[string]interface{}
		want    []string
	}{
		{
			name:    "nil details",
			details: nil,
			want:    []string{"running", "stopped"},
		},
		{
			name:    "empty details map",
			details: map[string]interface{}{},
			want:    []string{"running", "stopped"},
		},
		{
			name:    "key missing",
			details: map[string]interface{}{"other_key": "value"},
			want:    []string{"running", "stopped"},
		},
		{
			name:    "nil value",
			details: map[string]interface{}{"required_states": nil},
			want:    []string{"running", "stopped"},
		},
		{
			name:    "string slice with values",
			details: map[string]interface{}{"required_states": []string{"running", "paused"}},
			want:    []string{"running", "paused"},
		},
		{
			name:    "empty string slice",
			details: map[string]interface{}{"required_states": []string{}},
			want:    []string{},
		},
		{
			name:    "interface slice with values",
			details: map[string]interface{}{"required_states": []interface{}{"stopped", "created"}},
			want:    []string{"stopped", "created"},
		},
		{
			name:    "empty interface slice",
			details: map[string]interface{}{"required_states": []interface{}{}},
			want:    []string{},
		},
		{
			name:    "comma-separated string",
			details: map[string]interface{}{"required_states": "running, stopped, paused"},
			want:    []string{"running", "stopped", "paused"},
		},
		{
			name:    "single state string",
			details: map[string]interface{}{"required_states": "running"},
			want:    []string{"running"},
		},
		{
			name:    "empty string",
			details: map[string]interface{}{"required_states": ""},
			want:    []string{},
		},
		{
			name:    "whitespace only string",
			details: map[string]interface{}{"required_states": "   ,  ,  "},
			want:    []string{},
		},
		{
			name:    "unsupported type returns fallback",
			details: map[string]interface{}{"required_states": 123},
			want:    []string{"running", "stopped"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRequiredStates(tt.details)
			if len(got) != len(tt.want) {
				t.Errorf("extractRequiredStates() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractRequiredStates()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWrapInvalidStateTransition_EmptyRequiredStates(t *testing.T) {
	err := WrapInvalidStateTransition("my-vm", "delete", "running", []string{})

	if !strings.Contains(err.Message, "running") {
		t.Errorf("Message should contain current state, got: %s", err.Message)
	}
	// With empty requiredStates, suggestion should NOT mention specific states
	if strings.Contains(err.Suggestion, "' or '") {
		t.Errorf("Suggestion should not list states when empty, got: %s", err.Suggestion)
	}
	if !strings.Contains(err.Suggestion, "not allowed in the current state") {
		t.Errorf("Suggestion should indicate operation not allowed, got: %s", err.Suggestion)
	}
	if !strings.Contains(err.Suggestion, "nanofuse vm status my-vm") {
		t.Errorf("Suggestion should include status command, got: %s", err.Suggestion)
	}
	if err.ExitCode != 5 {
		t.Errorf("ExitCode = %d, want 5", err.ExitCode)
	}
}
