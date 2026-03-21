// Package testutil provides shared test utilities for gdt-based tests.
// This package consolidates common test helpers across CLI, API, build, and E2E test suites
// to reduce duplication and ensure consistency.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gdt-dev/gdt"
)

// FindProjectRoot locates the project root by walking up directories until
// it finds go.mod. This is used to set up proper paths for test execution.
func FindProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root")
		}
		dir = parent
	}
}

// RunGdtTest loads and executes a gdt test suite from a YAML file.
// It skips the test if the file is not found.
func RunGdtTest(t *testing.T, yamlPath string) {
	t.Helper()

	if _, err := os.Stat(yamlPath); err != nil {
		t.Skipf("Test file not found: %s", yamlPath)
	}

	s, err := gdt.From(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load test suite from %s: %v", yamlPath, err)
	}

	ctx := gdt.NewContext()
	err = s.Run(ctx, t)
	if err != nil {
		t.Fatalf("Test suite failed: %v", err)
	}
}

// SetupWorkingDir changes to the project root and returns a cleanup function
// that restores the original directory. On failure to restore, it calls t.Errorf
// to ensure test failures are captured.
func SetupWorkingDir(t *testing.T, root string) func() {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}
}

// FindCLIBinary locates the nanofuse CLI binary in common locations.
// Returns empty string if not found.
func FindCLIBinary(root string) string {
	// Check project-relative locations (most common for development/CI)
	locations := []string{
		filepath.Join(root, "bin", "nanofuse"),
		filepath.Join(root, "nanofuse"),
		"/usr/local/bin/nanofuse",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Try PATH lookup as fallback
	if path, err := exec.LookPath("nanofuse"); err == nil {
		return path
	}

	return ""
}

// AddBinaryToPath adds the directory containing the binary to PATH.
// This ensures shell commands in gdt YAML tests can find the binary.
// Uses filepath.ListSeparator for cross-platform compatibility.
func AddBinaryToPath(binaryPath string) {
	if binaryPath == "" {
		return
	}
	binDir := filepath.Dir(binaryPath)
	currentPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(filepath.ListSeparator)+currentPath)
}

// BuildArtifactsExist checks if the required build artifacts (kernel and rootfs)
// exist in the expected locations.
func BuildArtifactsExist(root string) bool {
	kernelPath := filepath.Join(root, "images", "base", "build", "vmlinux")
	if _, err := os.Stat(kernelPath); err != nil {
		return false
	}

	rootfsPath := filepath.Join(root, "images", "base", "build", "rootfs.ext4")
	if _, err := os.Stat(rootfsPath); err != nil {
		return false
	}

	return true
}
