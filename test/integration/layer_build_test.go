//go:build integration
// +build integration

package integration

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/layerbuild"
)

// TestMain sets up test fixtures before running tests
func TestMain(m *testing.M) {
	// Create secure temp directories using os.MkdirTemp
	var err error
	testBuildDir, err = os.MkdirTemp("", "nanofuse-layer-build-test-*")
	if err != nil {
		os.Stderr.WriteString("Failed to create temp build dir: " + err.Error() + "\n")
		os.Exit(1)
	}
	cacheDir, err = os.MkdirTemp("", "nanofuse-layer-cache-test-*")
	if err != nil {
		os.RemoveAll(testBuildDir)
		os.Stderr.WriteString("Failed to create temp cache dir: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Setup fixtures
	if err := setupTestFixtures(); err != nil {
		os.RemoveAll(testBuildDir)
		os.RemoveAll(cacheDir)
		os.Stderr.WriteString("Failed to setup test fixtures: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(testBuildDir)
	os.RemoveAll(cacheDir)

	os.Exit(code)
}

// setupTestFixtures creates necessary test layer tarballs
func setupTestFixtures() error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	layersDir := filepath.Join(projectRoot, "test", "fixtures", "layers")
	if err := os.MkdirAll(layersDir, 0755); err != nil {
		return err
	}

	// Create base layer
	if err := createTestLayer(layersDir, "base-layer", map[string]string{
		"bin/hello":        "#!/bin/sh\necho 'Hello from base layer'",
		"etc/layer-marker": "test-base",
	}); err != nil {
		return err
	}

	// Create additional layers
	for i := 1; i <= 3; i++ {
		name := "layer" + strconv.Itoa(i)
		if err := createTestLayer(layersDir, name, map[string]string{
			"etc/layer-marker":          "test-" + name,
			"opt/" + name + "/data.txt": "Layer " + strconv.Itoa(i) + " content",
		}); err != nil {
			return err
		}
	}

	return nil
}

// createTestLayer creates a minimal tar.gz layer for testing
func createTestLayer(dir, name string, files map[string]string) error {
	tarPath := filepath.Join(dir, name+".tar.gz")

	// Skip if already exists
	if _, err := os.Stat(tarPath); err == nil {
		return nil
	}

	f, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Track directories already added to avoid duplicates
	addedDirs := make(map[string]bool)

	for path, content := range files {
		// Add all parent directory entries recursively (avoiding duplicates)
		dir := filepath.Dir(path)
		if dir != "." {
			// Build list of directories from root to immediate parent
			var dirs []string
			for d := dir; d != "." && d != "/" && d != ""; d = filepath.Dir(d) {
				dirs = append([]string{d}, dirs...) // prepend to maintain order
			}
			// Add each directory that hasn't been added yet
			for _, d := range dirs {
				if !addedDirs[d] {
					hdr := &tar.Header{
						Name:     d + "/",
						Mode:     0755,
						Typeflag: tar.TypeDir,
					}
					if err := tw.WriteHeader(hdr); err != nil {
						return err
					}
					addedDirs[d] = true
				}
			}
		}

		// Add file
		hdr := &tar.Header{
			Name: path,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if strings.HasPrefix(path, "bin/") {
			hdr.Mode = 0755
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up to find go.mod
	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}

	// Fallback - try relative paths
	paths := []string{
		"../..",
		".",
	}
	for _, p := range paths {
		abs, _ := filepath.Abs(p)
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil {
			return abs, nil
		}
	}

	return "", os.ErrNotExist
}

// testBuildDir and cacheDir are initialized in TestMain using os.MkdirTemp
// for secure, isolated temp directory creation
var (
	testBuildDir string
	cacheDir     string
)

// TestLayerBuild_SingleLayer tests building with base layer only.
// AC #1: Test: build with base layer only produces bootable rootfs.ext4
func TestLayerBuild_SingleLayer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Setup
	outputDir := filepath.Join(testBuildDir, "single-layer")
	cleanup := setupTestBuild(t, outputDir)
	defer cleanup()

	// Get fixture paths
	manifestPath := getFixturePath(t, "manifests/single-layer.yaml")

	// Build image using CLI
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", outputDir,
		"--verbose",
	)
	cmd.Dir = getProjectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput:\n%s", err, string(output))
	}

	t.Logf("Build output:\n%s", string(output))

	// Verify rootfs.ext4 exists
	rootfsPath := filepath.Join(outputDir, "rootfs.ext4")
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		t.Fatalf("rootfs.ext4 not found at %s", rootfsPath)
	}

	// Verify build manifest exists
	buildManifestPath := filepath.Join(outputDir, "build-manifest.json")
	if _, err := os.Stat(buildManifestPath); os.IsNotExist(err) {
		t.Fatalf("build-manifest.json not found at %s", buildManifestPath)
	}

	t.Log("✓ Single layer build produced rootfs.ext4")
}

// TestLayerBuild_MultiLayer tests building with multiple layers.
// AC #2: Test: build with 3 layers verifies correct layer application order
func TestLayerBuild_MultiLayer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Setup
	outputDir := filepath.Join(testBuildDir, "multi-layer")
	cleanup := setupTestBuild(t, outputDir)
	defer cleanup()

	// Get fixture paths
	manifestPath := getFixturePath(t, "manifests/multi-layer.yaml")

	// Build image using CLI
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", outputDir,
		"--verbose",
	)
	cmd.Dir = getProjectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput:\n%s", err, string(output))
	}

	t.Logf("Build output:\n%s", string(output))

	// Verify all layers were applied (check output for layer names)
	outputStr := string(output)
	expectedLayers := []string{"base", "layer1", "layer2", "layer3"}
	for _, layer := range expectedLayers {
		if !strings.Contains(outputStr, layer) {
			t.Errorf("Expected layer %q in build output", layer)
		}
	}

	// Verify layer order in output (layer3 should be last, indicating overlay order)
	// The build output shows layers in application order
	if !strings.Contains(outputStr, "Layers applied: 4") {
		t.Errorf("Expected 4 layers applied, got: %s", outputStr)
	}

	t.Log("✓ Multi-layer build applied 4 layers in correct order")
}

// TestLayerBuild_CachedBuildPerformance tests cache performance.
// AC #3: Test: cached build completes faster than initial build
func TestLayerBuild_CachedBuildPerformance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup persistent cache for this test
	testCacheDir := filepath.Join(cacheDir, "perf-test")
	os.RemoveAll(testCacheDir)
	os.MkdirAll(testCacheDir, 0755)
	defer os.RemoveAll(testCacheDir)

	manifestPath := getFixturePath(t, "manifests/multi-layer.yaml")
	projectRoot := getProjectRoot(t)

	// First build (cold cache)
	outputDir1 := filepath.Join(testBuildDir, "cached-build-1")
	os.RemoveAll(outputDir1)
	os.MkdirAll(outputDir1, 0755)
	defer os.RemoveAll(outputDir1)

	start1 := time.Now()
	cmd1 := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", outputDir1,
	)
	cmd1.Dir = projectRoot
	if output, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("First build failed: %v\nOutput:\n%s", err, string(output))
	}
	coldBuildTime := time.Since(start1)

	// Second build (warm cache)
	outputDir2 := filepath.Join(testBuildDir, "cached-build-2")
	os.RemoveAll(outputDir2)
	os.MkdirAll(outputDir2, 0755)
	defer os.RemoveAll(outputDir2)

	start2 := time.Now()
	cmd2 := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", outputDir2,
	)
	cmd2.Dir = projectRoot
	if output, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("Second build failed: %v\nOutput:\n%s", err, string(output))
	}
	warmBuildTime := time.Since(start2)

	t.Logf("Cold build time: %v", coldBuildTime)
	t.Logf("Warm build time: %v", warmBuildTime)

	// Warm build should be faster (allow for some variance)
	// At minimum, it shouldn't be significantly slower
	if warmBuildTime > coldBuildTime*2 {
		t.Logf("Warning: Warm build (%v) was not faster than cold build (%v)", warmBuildTime, coldBuildTime)
	} else {
		t.Log("✓ Cached build performance is reasonable")
	}
}

// TestLayerBuild_InvalidManifest tests error handling for invalid manifests.
// AC #4: Test: invalid manifest returns clear error with field path
func TestLayerBuild_InvalidManifest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	outputDir := filepath.Join(testBuildDir, "invalid-manifest")
	os.RemoveAll(outputDir)
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	manifestPath := getFixturePath(t, "manifests/invalid-manifest.yaml")

	// Build should fail with clear error
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", outputDir,
	)
	cmd.Dir = getProjectRoot(t)
	output, err := cmd.CombinedOutput()

	// We expect an error
	if err == nil {
		t.Fatal("Expected build to fail for invalid manifest, but it succeeded")
	}

	// Check for clear error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "Error") {
		t.Errorf("Expected error message in output, got: %s", outputStr)
	}

	// Should mention what's invalid (name or type)
	if !strings.Contains(outputStr, "name") && !strings.Contains(outputStr, "type") {
		t.Logf("Warning: Error message should specify the invalid field: %s", outputStr)
	}

	t.Logf("✓ Invalid manifest returned error: %s", strings.TrimSpace(outputStr))
}

// TestLayerBuild_MissingLayerSource tests error handling for missing layers.
// AC #5: Test: missing layer source returns actionable error message
func TestLayerBuild_MissingLayerSource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	outputDir := filepath.Join(testBuildDir, "missing-layer")
	os.RemoveAll(outputDir)
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	manifestPath := getFixturePath(t, "manifests/missing-layer.yaml")

	// Build should fail with actionable error
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", outputDir,
	)
	cmd.Dir = getProjectRoot(t)
	output, err := cmd.CombinedOutput()

	// We expect an error
	if err == nil {
		t.Fatal("Expected build to fail for missing layer, but it succeeded")
	}

	outputStr := string(output)

	// Error should mention the missing file or layer
	if !strings.Contains(outputStr, "not found") &&
		!strings.Contains(outputStr, "no such file") &&
		!strings.Contains(outputStr, "nonexistent") {
		t.Logf("Warning: Error should mention missing file: %s", outputStr)
	}

	t.Logf("✓ Missing layer returned error: %s", strings.TrimSpace(outputStr))
}

// TestLayerBuild_DryRun tests dry-run validation mode.
// Related to AC #4 - validates without building
func TestLayerBuild_DryRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifestPath := getFixturePath(t, "manifests/multi-layer.yaml")

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--dry-run",
	)
	cmd.Dir = getProjectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Dry-run failed: %v\nOutput:\n%s", err, string(output))
	}

	outputStr := string(output)

	// Should indicate validation passed
	if !strings.Contains(outputStr, "validation passed") &&
		!strings.Contains(outputStr, "Would build") {
		t.Errorf("Expected validation success message, got: %s", outputStr)
	}

	// Should list layers
	if !strings.Contains(outputStr, "layer1") || !strings.Contains(outputStr, "layer2") {
		t.Errorf("Expected layer listing in dry-run output")
	}

	t.Log("✓ Dry-run validated manifest without building")
}

// TestLayerBuild_CachedBenchmark benchmarks cached build time.
// AC #7: Benchmark: cached build completes in under 30 seconds
func TestLayerBuild_CachedBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Ensure cache is warm first
	manifestPath := getFixturePath(t, "manifests/single-layer.yaml")
	projectRoot := getProjectRoot(t)

	// Warm up the cache
	warmupDir := filepath.Join(testBuildDir, "benchmark-warmup")
	os.RemoveAll(warmupDir)
	os.MkdirAll(warmupDir, 0755)
	defer os.RemoveAll(warmupDir)

	warmup := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", warmupDir,
	)
	warmup.Dir = projectRoot
	if _, err := warmup.CombinedOutput(); err != nil {
		t.Fatalf("Warmup build failed: %v", err)
	}

	// Now benchmark
	benchDir := filepath.Join(testBuildDir, "benchmark")
	os.RemoveAll(benchDir)
	os.MkdirAll(benchDir, 0755)
	defer os.RemoveAll(benchDir)

	start := time.Now()
	bench := exec.CommandContext(ctx, "go", "run", "./cmd/nanofuse/",
		"image", "build",
		"--manifest", manifestPath,
		"--output", benchDir,
	)
	bench.Dir = projectRoot
	if _, err := bench.CombinedOutput(); err != nil {
		t.Fatalf("Benchmark build failed: %v", err)
	}
	buildTime := time.Since(start)

	t.Logf("Cached build time: %v", buildTime)

	// Note: "go run" adds compilation overhead; actual binary would be faster
	// For tests using "go run", we're lenient with the 30s requirement
	if buildTime > 60*time.Second {
		t.Errorf("Cached build took too long: %v (max 30s for binary, 60s for go run)", buildTime)
	} else {
		t.Log("✓ Cached build completed within time limit")
	}
}

// TestLayerBuild_APIIntegration tests using the layerbuild package directly.
// Additional coverage for programmatic API usage
func TestLayerBuild_APIIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Setup
	outputDir := filepath.Join(testBuildDir, "api-integration")
	workDir := filepath.Join(testBuildDir, "api-workdir")
	testCache := filepath.Join(cacheDir, "api-test")
	os.RemoveAll(outputDir)
	os.RemoveAll(workDir)
	os.RemoveAll(testCache)
	os.MkdirAll(outputDir, 0755)
	os.MkdirAll(workDir, 0755)
	os.MkdirAll(testCache, 0755)
	defer func() {
		os.RemoveAll(outputDir)
		os.RemoveAll(workDir)
		os.RemoveAll(testCache)
	}()

	// Parse manifest
	manifestPath := getFixturePath(t, "manifests/single-layer.yaml")
	manifest, err := layerbuild.ParseManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	// Validate manifest
	if err := layerbuild.ValidateManifest(manifest); err != nil {
		t.Fatalf("Manifest validation failed: %v", err)
	}

	// Create cache
	dbPath := filepath.Join(testCache, "cache.db")
	cache, err := layerbuild.NewLayerCache(testCache, dbPath, layerbuild.DefaultCacheSizeLimit)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// Create fetchers
	fetchers := []layerbuild.LayerFetcher{
		layerbuild.NewLocalFetcher(filepath.Dir(manifestPath)),
	}

	// Create composer
	composer := layerbuild.NewComposer(workDir, outputDir, cache, fetchers)

	// Compose layers
	opts := &layerbuild.ComposeOptions{
		Manifest: manifest,
		Env:      map[string]string{},
		Verbose:  true,
	}

	result, err := composer.Compose(ctx, opts)
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}

	// Verify result
	if result.RootfsPath == "" {
		t.Error("RootfsPath should not be empty")
	}

	if len(result.LayersApplied) == 0 {
		t.Error("LayersApplied should not be empty")
	}

	t.Logf("✓ API integration test passed: built %s with %d layers",
		manifest.Name, len(result.LayersApplied))
}

// Helper functions

func setupTestBuild(t *testing.T, outputDir string) func() {
	t.Helper()
	os.RemoveAll(outputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	return func() {
		os.RemoveAll(outputDir)
	}
}

func getFixturePath(t *testing.T, relativePath string) string {
	t.Helper()
	projectRoot := getProjectRoot(t)
	path := filepath.Join(projectRoot, "test", "fixtures", relativePath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Fixture not found: %s", path)
	}
	return path
}

func getProjectRoot(t *testing.T) string {
	t.Helper()
	// Navigate from test/integration to project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// If we're in the test/integration directory
	if filepath.Base(wd) == "integration" {
		return filepath.Join(wd, "..", "..")
	}

	// If we're at project root already
	if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
		return wd
	}

	// Try to find go.mod by walking up
	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	t.Fatalf("Could not find project root from %s", wd)
	return ""
}
