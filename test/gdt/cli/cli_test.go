// Package cli provides gdt-based declarative tests for the nanofuse CLI.
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daax-dev/nanofuse/test/gdt/testutil"
)

// TestGdtCLISuite runs all gdt CLI tests.
// These tests validate CLI error boundaries, argument parsing, and output formatting.
// Note: Most tests do NOT require the daemon to be running - they test CLI behavior itself.
func TestGdtCLISuite(t *testing.T) {
	// Skip if GDT_SKIP is set
	if os.Getenv("GDT_SKIP") != "" {
		t.Skip("GDT_SKIP is set - skipping gdt tests")
	}

	root := testutil.FindProjectRoot(t)

	// Check if CLI binary exists
	cliBinary := testutil.FindCLIBinary(root)
	if cliBinary == "" {
		t.Skip("nanofuse CLI not found - run 'mage cli' first")
	}

	// Add binary directory to PATH so shell commands in YAML tests can find it
	testutil.AddBinaryToPath(cliBinary)

	// Set environment for tests
	os.Setenv("PROJECT_ROOT", root)
	os.Setenv("NANOFUSE_BIN", cliBinary)

	// Change to project root so relative paths work
	cleanup := testutil.SetupWorkingDir(t, root)
	defer cleanup()

	testDir := filepath.Join(root, "test", "gdt", "cli")

	// Run error handling tests
	t.Run("ErrorHandling", func(t *testing.T) {
		testutil.RunGdtTest(t, filepath.Join(testDir, "error_handling.yaml"))
	})

	// Run help and usage tests
	t.Run("HelpUsage", func(t *testing.T) {
		testutil.RunGdtTest(t, filepath.Join(testDir, "help_usage.yaml"))
	})
}
