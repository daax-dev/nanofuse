package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDockerBuilderAvailable(t *testing.T) {
	builder := NewDockerBuilder("/tmp/nanofuse-test", false)

	err := builder.Available()
	if err != nil {
		t.Skipf("Docker/Podman not available: %v", err)
	}

	if builder.runtime != "docker" && builder.runtime != "podman" {
		t.Errorf("Expected runtime to be docker or podman, got: %s", builder.runtime)
	}

	t.Logf("Using runtime: %s", builder.runtime)
}

func TestSanitizeDigest(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sha256:abc123", "sha256-abc123"},
		{"sha256:abc123def456", "sha256-abc123def456"},
		{"no-colon", "no-colon"},
	}

	for _, tc := range tests {
		result := sanitizeDigest(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeDigest(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestExtractVersionFromPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/boot/vmlinux-5.10.240", "5.10.240"},
		{"/boot/vmlinuz-5.15.0-generic", "5.15.0-generic"},
		{"/boot/vmlinux", "unknown"},
		{"/vmlinux", "unknown"},
	}

	for _, tc := range tests {
		result := extractVersionFromPath(tc.input)
		if result != tc.expected {
			t.Errorf("extractVersionFromPath(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestDefaultExtractOptions(t *testing.T) {
	opts := DefaultExtractOptions()

	if opts.RootfsSizeMB != 2048 {
		t.Errorf("Expected RootfsSizeMB=2048, got %d", opts.RootfsSizeMB)
	}

	if len(opts.KernelSearchPaths) == 0 {
		t.Error("Expected non-empty KernelSearchPaths")
	}
}

// Integration test - only runs if Docker is available and INTEGRATION_TEST=1
func TestDockerBuilderExtract(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Set INTEGRATION_TEST=1 to run integration tests")
	}

	dataDir := t.TempDir()
	builder := NewDockerBuilder(dataDir, true)

	if err := builder.Available(); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	// alpine (like most OCI images) contains no kernel, so a fallback kernel is
	// required. Provide one via NANOFUSE_TEST_FALLBACK_KERNEL; skip otherwise so
	// the test stays meaningful instead of reliably failing on kernel extraction.
	fallbackKernel := os.Getenv("NANOFUSE_TEST_FALLBACK_KERNEL")
	if fallbackKernel == "" {
		t.Skip("Set NANOFUSE_TEST_FALLBACK_KERNEL to a kernel path to run this integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Use a simple image for testing
	result, err := builder.Extract(ctx, "alpine:latest", ExtractOptions{
		OutputDir:          filepath.Join(t.TempDir(), "extract-test"),
		RootfsSizeMB:       256,
		Verbose:            true,
		FallbackKernelPath: fallbackKernel,
	})

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	t.Logf("Extraction result:")
	t.Logf("  Kernel: %s", result.KernelPath)
	t.Logf("  Rootfs: %s", result.RootfsPath)
	t.Logf("  Duration: %v", result.Duration)

	// alpine has no kernel, so the fallback must have been used.
	if result.KernelPath != fallbackKernel {
		t.Errorf("expected fallback kernel %q to be used, got %q", fallbackKernel, result.KernelPath)
	}
}

func TestValidateFallbackKernel(t *testing.T) {
	dir := t.TempDir()

	regular := dir + "/vmlinux"
	if err := os.WriteFile(regular, []byte("kernel"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("regular readable file", func(t *testing.T) {
		if err := validateFallbackKernel(regular); err != nil {
			t.Errorf("expected nil for a readable regular file, got: %v", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if err := validateFallbackKernel(dir + "/does-not-exist"); err == nil {
			t.Error("expected an error for a missing file, got nil")
		}
	})

	t.Run("directory", func(t *testing.T) {
		if err := validateFallbackKernel(dir); err == nil {
			t.Error("expected an error for a directory, got nil")
		}
	})
}
