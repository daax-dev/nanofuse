//go:build e2e
// +build e2e

// Package e2e provides gdt-based declarative E2E tests.
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jpoley/nanofuse/test/gdt/testutil"
)

// TestGdtE2ESuite runs all gdt E2E lifecycle tests.
// Requires: sudo, KVM, Firecracker, daemon running, built image
func TestGdtE2ESuite(t *testing.T) {
	// Check prerequisites
	if !isRoot() {
		t.Skip("E2E tests require root access (run with sudo)")
	}

	if !kvmAvailable() {
		t.Skip("KVM not available (/dev/kvm)")
	}

	if !firecrackerInstalled() {
		t.Skip("Firecracker not installed")
	}

	root := testutil.FindProjectRoot(t)
	if !testutil.BuildArtifactsExist(root) {
		t.Skip("Build artifacts not found - run build first")
	}

	if !daemonRunning() {
		t.Skip("Daemon not running - start nanofused first")
	}

	// Change to project root so relative paths in YAML work
	cleanup := testutil.SetupWorkingDir(t, root)
	defer cleanup()

	// Find the test directory
	testDir := filepath.Join(root, "test", "gdt", "e2e")

	t.Run("FullLifecycle", func(t *testing.T) {
		testutil.RunGdtTest(t, filepath.Join(testDir, "full_lifecycle.yaml"))
	})
}

// Helper functions

func isRoot() bool {
	return os.Geteuid() == 0
}

func kvmAvailable() bool {
	_, err := os.Stat("/dev/kvm")
	return err == nil
}

func firecrackerInstalled() bool {
	_, err := exec.LookPath("firecracker")
	return err == nil
}

// daemonRunning checks common socket locations in standard Linux order:
// /var/run (traditional), /run (modern systemd), /tmp (development/testing).
func daemonRunning() bool {
	socketPaths := []string{
		"/var/run/nanofused.sock",
		"/run/nanofused.sock",
		"/tmp/nanofused.sock",
	}

	for _, path := range socketPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	cmd := exec.Command("systemctl", "is-active", "--quiet", "nanofused")
	return cmd.Run() == nil
}
