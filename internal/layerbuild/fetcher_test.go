package layerbuild

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestLocalFetcher_Directory tests fetching a layer from a local directory
func TestLocalFetcher_Directory(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")
	rootfsDir := filepath.Join(layerDir, "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create layer.yaml
	metadata := &LayerPackage{
		Name:         "test-layer",
		Version:      "1.0.0",
		Description:  "Test layer for unit tests",
		Type:         LayerTypeBase,
		Dependencies: []string{},
		Provides:     []string{"test-capability"},
	}
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal layer metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(layerDir, "layer.yaml"), yamlData, 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	// Create some test files in rootfs
	if err := os.WriteFile(filepath.Join(rootfsDir, "test.txt"), []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create fetcher with work directory
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}
	fetcher := NewLocalFetcher(workDir)

	// Test Supports method
	if !fetcher.Supports(SourceTypeLocal) {
		t.Error("LocalFetcher should support SourceTypeLocal")
	}
	if fetcher.Supports(SourceTypeDocker) {
		t.Error("LocalFetcher should not support SourceTypeDocker")
	}

	// Fetch the layer
	source := "local://" + layerDir
	cached, err := fetcher.Fetch(source)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify result
	if cached.Metadata == nil {
		t.Fatal("Result metadata is nil")
	}
	if cached.Metadata.Name != "test-layer" {
		t.Errorf("Expected name 'test-layer', got '%s'", cached.Metadata.Name)
	}
	if cached.Metadata.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cached.Metadata.Version)
	}
	if cached.Name != "test-layer" {
		t.Errorf("Expected cached name 'test-layer', got '%s'", cached.Name)
	}
	if cached.Version != "1.0.0" {
		t.Errorf("Expected cached version '1.0.0', got '%s'", cached.Version)
	}
	if cached.Digest == "" {
		t.Error("Expected non-empty digest")
	}
	if cached.SizeBytes == 0 {
		t.Error("Expected non-zero size")
	}
	if cached.LocalPath == "" {
		t.Error("Expected non-empty local path")
	}
	if cached.SourceURL != source {
		t.Errorf("Expected source URL %s, got %s", source, cached.SourceURL)
	}

	// Verify tarball was created
	if _, err := os.Stat(cached.LocalPath); err != nil {
		t.Errorf("Tarball not created at %s: %v", cached.LocalPath, err)
	}

	// Verify tarball contains expected files
	verifyTarballContents(t, cached.LocalPath, []string{"layer.yaml", "rootfs/test.txt"})
}

// TestLocalFetcher_Tarball tests fetching a layer from a tarball
func TestLocalFetcher_Tarball(t *testing.T) {
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Create a test tarball
	tarballPath := filepath.Join(tmpDir, "test-layer.tar.gz")
	metadata := &LayerPackage{
		Name:        "tarball-layer",
		Version:     "2.0.0",
		Description: "Test tarball layer",
		Type:        LayerTypeRuntime,
	}
	createTestTarball(t, tarballPath, metadata, map[string]string{
		"rootfs/bin/test": "#!/bin/sh\necho test",
	})

	// Fetch the layer
	fetcher := NewLocalFetcher(workDir)
	source := "local://" + tarballPath
	cached, err := fetcher.Fetch(source)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify result
	if cached.Name != "tarball-layer" {
		t.Errorf("Expected name 'tarball-layer', got '%s'", cached.Name)
	}
	if cached.Metadata.Name != "tarball-layer" {
		t.Errorf("Expected metadata name 'tarball-layer', got '%s'", cached.Metadata.Name)
	}
	if cached.Digest == "" {
		t.Error("Expected non-empty digest")
	}

	// Verify the digest matches the original tarball
	originalDigest := calculateFileDigest(t, tarballPath)
	if cached.Digest != originalDigest {
		t.Errorf("Digest mismatch: expected %s, got %s", originalDigest, cached.Digest)
	}
}

// TestLocalFetcher_MissingLayerYAML tests error handling for missing layer.yaml
func TestLocalFetcher_MissingLayerYAML(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "invalid-layer")
	if err := os.MkdirAll(layerDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewLocalFetcher(workDir)
	source := "local://" + layerDir
	_, err := fetcher.Fetch(source)
	if err == nil {
		t.Error("Expected error for missing layer.yaml, got nil")
	}
}

// TestLocalFetcher_InvalidSource tests error handling for invalid source URLs
func TestLocalFetcher_InvalidSource(t *testing.T) {
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewLocalFetcher(workDir)

	testCases := []struct {
		name   string
		source string
	}{
		{"Empty source", ""},
		{"Missing prefix", "/some/path"},
		{"Wrong prefix", "docker://test"},
		{"Non-existent path", "local:///nonexistent/path"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fetcher.Fetch(tc.source)
			if err == nil {
				t.Errorf("Expected error for source '%s', got nil", tc.source)
			}
		})
	}
}

// TestDockerFetcher_Supports tests the Supports method
func TestDockerFetcher_Supports(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewDockerFetcher(tmpDir)

	if !fetcher.Supports(SourceTypeDocker) {
		t.Error("DockerFetcher should support SourceTypeDocker")
	}
	if fetcher.Supports(SourceTypeLocal) {
		t.Error("DockerFetcher should not support SourceTypeLocal")
	}
}

// TestDockerFetcher_InvalidSource tests error handling for invalid Docker sources
func TestDockerFetcher_InvalidSource(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewDockerFetcher(tmpDir)

	testCases := []struct {
		name   string
		source string
	}{
		{"Empty source", ""},
		{"Missing prefix", "alpine:latest"},
		{"Wrong prefix", "local://alpine"},
		{"Empty image name", "docker://"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fetcher.Fetch(tc.source)
			if err == nil {
				t.Errorf("Expected error for source '%s', got nil", tc.source)
			}
		})
	}
}

// TestVerifyDigest tests digest verification logic
func TestVerifyDigest(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content for digest verification")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Calculate expected digest
	hash := sha256.Sum256(testContent)
	expectedDigest := hex.EncodeToString(hash[:])

	// Test with matching digest
	err := verifyDigest(testFile, expectedDigest)
	if err != nil {
		t.Errorf("Digest verification failed for matching digest: %v", err)
	}

	// Test with mismatching digest
	wrongDigest := "0000000000000000000000000000000000000000000000000000000000000000"
	err = verifyDigest(testFile, wrongDigest)
	if err == nil {
		t.Error("Expected error for mismatching digest, got nil")
	}

	// Test with empty expected digest (should skip verification)
	err = verifyDigest(testFile, "")
	if err != nil {
		t.Errorf("Verification with empty digest should pass: %v", err)
	}
}

// TestProgressCallback tests progress reporting for large layers
func TestProgressCallback(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "large-layer")
	rootfsDir := filepath.Join(layerDir, "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create layer.yaml
	metadata := &LayerPackage{
		Name:    "large-layer",
		Version: "1.0.0",
		Type:    LayerTypeBase,
	}
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(layerDir, "layer.yaml"), yamlData, 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	// Create a large file (>100MB to trigger progress callback)
	largeFile := filepath.Join(rootfsDir, "large.bin")
	f, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}
	defer f.Close()

	// Write 101MB of data
	chunk := make([]byte, 1024*1024) // 1MB chunks
	for i := 0; i < 101; i++ {
		if _, err := f.Write(chunk); err != nil {
			t.Fatalf("Failed to write to large file: %v", err)
		}
	}
	f.Close()

	// Fetch with progress callback
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewLocalFetcher(workDir)
	callbackCalled := false
	fetcher.SetProgressCallback(func(current, total int64) {
		callbackCalled = true
		if current < 0 || total < 0 {
			t.Error("Progress values should be non-negative")
		}
		if current > total {
			t.Error("Current should not exceed total")
		}
	})

	source := "local://" + layerDir
	_, err = fetcher.Fetch(source)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if !callbackCalled {
		t.Error("Progress callback was not called for large layer")
	}
}

// Helper function to create a test tarball
func createTestTarball(t *testing.T, path string, metadata *LayerPackage, files map[string]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create tarball: %v", err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Write layer.yaml
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}
	if err := writeTarEntry(tw, "layer.yaml", yamlData); err != nil {
		t.Fatalf("Failed to write layer.yaml to tarball: %v", err)
	}

	// Write additional files
	for name, content := range files {
		if err := writeTarEntry(tw, name, []byte(content)); err != nil {
			t.Fatalf("Failed to write %s to tarball: %v", name, err)
		}
	}
}

func writeTarEntry(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func verifyTarballContents(t *testing.T, tarballPath string, expectedFiles []string) {
	t.Helper()

	f, err := os.Open(tarballPath)
	if err != nil {
		t.Fatalf("Failed to open tarball: %v", err)
	}
	defer f.Close()

	var tr *tar.Reader
	if filepath.Ext(tarballPath) == ".gz" {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzr.Close()
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(f)
	}

	foundFiles := make(map[string]bool)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read tar entry: %v", err)
		}
		foundFiles[hdr.Name] = true
	}

	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected file '%s' not found in tarball", expected)
		}
	}
}

func calculateFileDigest(t *testing.T, path string) string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open file for digest: %v", err)
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		t.Fatalf("Failed to calculate digest: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// TestRegistryFetcher_Supports tests the Supports method
func TestRegistryFetcher_Supports(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewRegistryFetcher(tmpDir)

	if !fetcher.Supports(SourceTypeRegistry) {
		t.Error("RegistryFetcher should support SourceTypeRegistry")
	}
	if fetcher.Supports(SourceTypeLocal) {
		t.Error("RegistryFetcher should not support SourceTypeLocal")
	}
	if fetcher.Supports(SourceTypeDocker) {
		t.Error("RegistryFetcher should not support SourceTypeDocker")
	}
}

// TestRegistryFetcher_InvalidSource tests error handling for invalid registry sources
func TestRegistryFetcher_InvalidSource(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewRegistryFetcher(tmpDir)

	testCases := []struct {
		name   string
		source string
	}{
		{"Empty source", ""},
		{"Missing prefix", "ghcr.io/test/image:tag"},
		{"Wrong prefix - local", "local://test"},
		{"Wrong prefix - docker", "docker://test"},
		{"Empty image reference", "registry://"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fetcher.Fetch(tc.source)
			if err == nil {
				t.Errorf("Expected error for source '%s', got nil", tc.source)
			}
		})
	}
}

// TestRegistryFetcher_InvalidImageReference tests error handling for invalid image references
func TestRegistryFetcher_InvalidImageReference(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewRegistryFetcher(tmpDir)

	testCases := []struct {
		name   string
		source string
	}{
		{"Invalid characters", "registry://ghcr.io/test/image:tag@@@invalid"},
		{"Too many colons", "registry://ghcr.io:invalid:port/image:tag"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fetcher.Fetch(tc.source)
			if err == nil {
				t.Errorf("Expected error for source '%s', got nil", tc.source)
			}
		})
	}
}

// TestRegistryFetcher_Options tests the option functions
func TestRegistryFetcher_Options(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("WithMaxRetries", func(t *testing.T) {
		fetcher := NewRegistryFetcher(tmpDir, WithMaxRetries(5))
		if fetcher.maxRetries != 5 {
			t.Errorf("Expected maxRetries to be 5, got %d", fetcher.maxRetries)
		}
	})

	t.Run("WithBaseBackoff", func(t *testing.T) {
		backoff := 2 * time.Second
		fetcher := NewRegistryFetcher(tmpDir, WithBaseBackoff(backoff))
		if fetcher.baseBackoff != backoff {
			t.Errorf("Expected baseBackoff to be %v, got %v", backoff, fetcher.baseBackoff)
		}
	})

	t.Run("DefaultValues", func(t *testing.T) {
		fetcher := NewRegistryFetcher(tmpDir)
		if fetcher.maxRetries != 3 {
			t.Errorf("Expected default maxRetries to be 3, got %d", fetcher.maxRetries)
		}
		if fetcher.baseBackoff != 1*time.Second {
			t.Errorf("Expected default baseBackoff to be 1s, got %v", fetcher.baseBackoff)
		}
	})
}

// TestIsRateLimitError tests the rate limit error detection
func TestIsRateLimitError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if isRateLimitError(nil) {
			t.Error("nil error should not be rate limit error")
		}
	})

	t.Run("non-transport error", func(t *testing.T) {
		err := errors.New("some other error")
		if isRateLimitError(err) {
			t.Error("non-transport error should not be rate limit error")
		}
	})
}

// TestWrapRegistryError tests error wrapping for registry errors
func TestWrapRegistryError(t *testing.T) {
	testCases := []struct {
		name     string
		errMsg   string
		contains string
	}{
		{
			name:     "no such host",
			errMsg:   "dial tcp: lookup invalid.registry.example: no such host",
			contains: "registry not found",
		},
		{
			name:     "connection refused",
			errMsg:   "dial tcp 127.0.0.1:5000: connection refused",
			contains: "cannot connect to registry",
		},
		{
			name:     "timeout",
			errMsg:   "context deadline exceeded (timeout)",
			contains: "timeout connecting to registry",
		},
		{
			name:     "generic error",
			errMsg:   "some unknown error",
			contains: "failed to fetch from registry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := errors.New(tc.errMsg)
			wrapped := wrapRegistryError(err, "test/image:latest")
			if wrapped == nil {
				t.Fatal("Expected wrapped error, got nil")
			}
			if !containsString(wrapped.Error(), tc.contains) {
				t.Errorf("Expected error to contain %q, got %q", tc.contains, wrapped.Error())
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration tests for RegistryFetcher
// These tests require network access and are skipped in short mode

// TestRegistryFetcher_Integration_DockerHub tests fetching a public image from Docker Hub
func TestRegistryFetcher_Integration_DockerHub(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Use a very small public image from Docker Hub
	// busybox is one of the smallest public images
	fetcher := NewRegistryFetcher(workDir)
	source := "registry://docker.io/library/busybox:stable"

	cached, err := fetcher.Fetch(source)
	if err != nil {
		t.Fatalf("Failed to fetch from Docker Hub: %v", err)
	}

	// Verify basic properties
	if cached == nil {
		t.Fatal("Expected non-nil cached layer")
	}
	if cached.Digest == "" {
		t.Error("Expected non-empty digest")
	}
	if cached.SizeBytes <= 0 {
		t.Error("Expected positive size")
	}
	if cached.LocalPath == "" {
		t.Error("Expected non-empty local path")
	}
	if cached.SourceURL != source {
		t.Errorf("Expected source URL %s, got %s", source, cached.SourceURL)
	}

	// Verify tarball was created and is valid
	if _, err := os.Stat(cached.LocalPath); err != nil {
		t.Errorf("Tarball not created at %s: %v", cached.LocalPath, err)
	}

	// Verify OCI digest is stored in metadata
	if cached.Metadata == nil {
		t.Fatal("Expected non-nil metadata")
	}
	if cached.Metadata.Metadata == nil {
		t.Fatal("Expected non-nil metadata.Metadata")
	}
	ociDigest, ok := cached.Metadata.Metadata["oci-digest"]
	if !ok || ociDigest == "" {
		t.Error("Expected OCI digest to be stored in metadata")
	}

	t.Logf("Successfully fetched Docker Hub image: %s (size: %d bytes, digest: %s)",
		source, cached.SizeBytes, cached.Digest[:16]+"...")
}

// TestRegistryFetcher_Integration_GHCR tests fetching a public image from ghcr.io
func TestRegistryFetcher_Integration_GHCR(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Use a small, well-known public image from ghcr.io
	// helm/chartmuseum is a popular small image
	fetcher := NewRegistryFetcher(workDir)
	source := "registry://ghcr.io/helm/chartmuseum:v0.16.1"

	cached, err := fetcher.Fetch(source)
	if err != nil {
		// ghcr.io may require authentication for some images
		// Check if it's an auth error and provide helpful message
		if containsString(err.Error(), "authentication") || containsString(err.Error(), "UNAUTHORIZED") {
			t.Skipf("Skipping test: ghcr.io requires authentication: %v", err)
		}
		// Some images may not exist; skip gracefully
		if containsString(err.Error(), "not found") || containsString(err.Error(), "MANIFEST_UNKNOWN") {
			t.Skipf("Skipping test: ghcr.io image not found (may have been removed): %v", err)
		}
		t.Fatalf("Failed to fetch from ghcr.io: %v", err)
	}

	// Verify basic properties
	if cached == nil {
		t.Fatal("Expected non-nil cached layer")
	}
	if cached.Digest == "" {
		t.Error("Expected non-empty digest")
	}
	if cached.SizeBytes <= 0 {
		t.Error("Expected positive size")
	}

	t.Logf("Successfully fetched ghcr.io image: %s (size: %d bytes, digest: %s)",
		source, cached.SizeBytes, cached.Digest[:16]+"...")
}

// TestRegistryFetcher_Integration_NonExistentImage tests error handling for non-existent images
func TestRegistryFetcher_Integration_NonExistentImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewRegistryFetcher(workDir)
	source := "registry://docker.io/library/this-image-definitely-does-not-exist-12345:latest"

	_, err := fetcher.Fetch(source)
	if err == nil {
		t.Fatal("Expected error for non-existent image, got nil")
	}

	// Verify we get a helpful error message
	errStr := err.Error()
	if !containsString(errStr, "not found") && !containsString(errStr, "MANIFEST_UNKNOWN") {
		t.Logf("Error message: %s", errStr)
		// Don't fail - just log, as different registries may return different errors
	}

	t.Logf("Got expected error for non-existent image: %v", err)
}

// TestRegistryFetcher_Integration_DigestVerification tests that OCI digest is captured
func TestRegistryFetcher_Integration_DigestVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Fetch the same image twice to verify digest consistency
	fetcher := NewRegistryFetcher(workDir)
	source := "registry://docker.io/library/busybox:stable"

	cached1, err := fetcher.Fetch(source)
	if err != nil {
		t.Fatalf("First fetch failed: %v", err)
	}

	cached2, err := fetcher.Fetch(source)
	if err != nil {
		t.Fatalf("Second fetch failed: %v", err)
	}

	// OCI digests should be identical for the same tag
	ociDigest1 := cached1.Metadata.Metadata["oci-digest"]
	ociDigest2 := cached2.Metadata.Metadata["oci-digest"]

	if ociDigest1 != ociDigest2 {
		t.Errorf("OCI digests should match: %s vs %s", ociDigest1, ociDigest2)
	}

	t.Logf("OCI digest verification passed: %s", ociDigest1)
}

// TestFetchResult_ToCachedLayer tests the conversion from FetchResult to CachedLayer
func TestFetchResult_ToCachedLayer(t *testing.T) {
	metadata := &LayerPackage{
		Name:        "test-layer",
		Version:     "1.0.0",
		Description: "Test layer",
		Type:        LayerTypeBase,
	}

	result := &FetchResult{
		TarballPath: "/tmp/test-layer.tar.gz",
		Digest:      "abc123",
		Metadata:    metadata,
		SizeBytes:   1024,
	}

	sourceURL := "local:///tmp/test-layer"
	cached := result.ToCachedLayer(sourceURL)

	if cached.Digest != result.Digest {
		t.Errorf("Expected digest %s, got %s", result.Digest, cached.Digest)
	}
	if cached.Name != metadata.Name {
		t.Errorf("Expected name %s, got %s", metadata.Name, cached.Name)
	}
	if cached.Version != metadata.Version {
		t.Errorf("Expected version %s, got %s", metadata.Version, cached.Version)
	}
	if cached.Type != metadata.Type {
		t.Errorf("Expected type %s, got %s", metadata.Type, cached.Type)
	}
	if cached.SourceURL != sourceURL {
		t.Errorf("Expected source URL %s, got %s", sourceURL, cached.SourceURL)
	}
	if cached.LocalPath != result.TarballPath {
		t.Errorf("Expected local path %s, got %s", result.TarballPath, cached.LocalPath)
	}
	if cached.SizeBytes != result.SizeBytes {
		t.Errorf("Expected size %d, got %d", result.SizeBytes, cached.SizeBytes)
	}
	if cached.Metadata != metadata {
		t.Error("Metadata pointer mismatch")
	}
	if cached.FetchedAt.IsZero() {
		t.Error("FetchedAt should not be zero")
	}
	if cached.LastUsedAt.IsZero() {
		t.Error("LastUsedAt should not be zero")
	}
}
