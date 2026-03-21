// Package api provides gdt-based declarative tests for the nanofused API.
package api

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jpoley/nanofuse/test/gdt/testutil"
)

// TestGdtAPISuite runs all gdt API tests.
// These tests validate API error responses, validation, and boundary conditions.
// Note: Requires nanofused daemon to be running.
func TestGdtAPISuite(t *testing.T) {
	// Skip if GDT_SKIP is set
	if os.Getenv("GDT_SKIP") != "" {
		t.Skip("GDT_SKIP is set - skipping gdt tests")
	}

	// Check if daemon is running
	socketPath := findDaemonSocket()
	if socketPath == "" {
		t.Skip("nanofused daemon not running - start the daemon first")
	}

	root := testutil.FindProjectRoot(t)

	// Set environment for tests
	os.Setenv("PROJECT_ROOT", root)
	os.Setenv("NANOFUSED_SOCKET", socketPath)

	// Change to project root so relative paths work
	cleanup := testutil.SetupWorkingDir(t, root)
	defer cleanup()

	testDir := filepath.Join(root, "test", "gdt", "api")

	// Run error response tests
	t.Run("ErrorResponses", func(t *testing.T) {
		testutil.RunGdtTest(t, filepath.Join(testDir, "error_responses.yaml"))
	})

	// Run health endpoint tests
	t.Run("HealthEndpoint", func(t *testing.T) {
		testutil.RunGdtTest(t, filepath.Join(testDir, "health.yaml"))
	})
}

// daemonSocketPath is the expected socket location for gdt API tests.
// This matches the CLI default (cmd/nanofuse/main.go:97) and the hardcoded
// paths in the gdt YAML test files. Shell variable substitution does not
// work reliably in the gdt framework, so tests use this fixed path.
const daemonSocketPath = "/run/nanofused.sock"

// findDaemonSocket checks if the daemon is running at the expected socket path.
// Returns the socket path if daemon is reachable, empty string otherwise.
func findDaemonSocket() string {
	if canConnectSocket(daemonSocketPath) {
		return daemonSocketPath
	}
	return ""
}

func canConnectSocket(path string) bool {
	conn, err := net.DialTimeout("unix", path, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
