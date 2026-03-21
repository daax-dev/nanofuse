package layerbuild

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewLayerCache verifies cache initialization
func TestNewLayerCache(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	dbPath := filepath.Join(dir, "test.db")

	cache, err := NewLayerCache(cacheDir, dbPath, 1024*1024*1024) // 1GB limit
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Verify cache directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Errorf("cache directory was not created: %s", cacheDir)
	}

	// Verify we can get stats from empty cache
	stats, err := cache.Stats()
	if err != nil {
		t.Errorf("Stats() failed: %v", err)
	}
	if stats.TotalLayers != 0 {
		t.Errorf("TotalLayers = %d, want 0", stats.TotalLayers)
	}
	if stats.TotalBytes != 0 {
		t.Errorf("TotalBytes = %d, want 0", stats.TotalBytes)
	}
}

// TestLayerCache_Put_Get verifies basic Put/Get operations
func TestLayerCache_Put_Get(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Create a test layer with tarball
	tarballPath := filepath.Join(dir, "test-layer.tar.gz")
	testData := []byte("fake tarball data for testing")
	if err := os.WriteFile(tarballPath, testData, 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer := &CachedLayer{
		Digest:     "sha256:abc123def456",
		Name:       "test-layer",
		Version:    "1.0.0",
		Type:       LayerTypeRuntime,
		SourceURL:  "docker://test:latest",
		LocalPath:  tarballPath,
		SizeBytes:  int64(len(testData)),
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
		Metadata: &LayerPackage{
			Name:        "test-layer",
			Version:     "1.0.0",
			Description: "Test layer",
			Type:        LayerTypeRuntime,
		},
	}

	// Put the layer in cache
	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// Get the layer back
	cached, err := cache.Get("sha256:abc123def456")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if cached == nil {
		t.Fatal("Get() returned nil for existing layer")
	}

	// Verify fields
	if cached.Digest != layer.Digest {
		t.Errorf("Digest = %q, want %q", cached.Digest, layer.Digest)
	}
	if cached.Name != layer.Name {
		t.Errorf("Name = %q, want %q", cached.Name, layer.Name)
	}
	if cached.Version != layer.Version {
		t.Errorf("Version = %q, want %q", cached.Version, layer.Version)
	}
	if cached.Type != layer.Type {
		t.Errorf("Type = %q, want %q", cached.Type, layer.Type)
	}
	if cached.SizeBytes != layer.SizeBytes {
		t.Errorf("SizeBytes = %d, want %d", cached.SizeBytes, layer.SizeBytes)
	}

	// Verify tarball was copied to cache
	expectedPath := filepath.Join(dir, "cache", "sha256:abc123def456.tar.gz")
	if cached.LocalPath != expectedPath {
		t.Errorf("LocalPath = %q, want %q", cached.LocalPath, expectedPath)
	}
	if _, err := os.Stat(cached.LocalPath); err != nil {
		t.Errorf("cached tarball does not exist: %v", err)
	}

	// Verify metadata was preserved
	if cached.Metadata == nil {
		t.Fatal("Metadata is nil")
	}
	if cached.Metadata.Name != layer.Metadata.Name {
		t.Errorf("Metadata.Name = %q, want %q", cached.Metadata.Name, layer.Metadata.Name)
	}
}

// TestLayerCache_Get_NotFound verifies cache miss behavior
func TestLayerCache_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	cached, err := cache.Get("sha256:nonexistent")
	if err != nil {
		t.Errorf("Get() should not error on cache miss: %v", err)
	}
	if cached != nil {
		t.Error("Get() should return nil for cache miss")
	}
}

// TestLayerCache_Exists verifies Exists check
func TestLayerCache_Exists(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Check non-existent layer
	exists, err := cache.Exists("sha256:nothere")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if exists {
		t.Error("Exists() = true for non-existent layer")
	}

	// Add a layer
	tarballPath := filepath.Join(dir, "test.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer := &CachedLayer{
		Digest:     "sha256:exists",
		Name:       "test",
		Type:       LayerTypeBase,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  4,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// Check existing layer
	exists, err = cache.Exists("sha256:exists")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if !exists {
		t.Error("Exists() = false for existing layer")
	}
}

// TestLayerCache_Touch verifies LRU timestamp updates
func TestLayerCache_Touch(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Add a layer
	tarballPath := filepath.Join(dir, "test.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	initialTime := time.Now().Add(-1 * time.Hour) // 1 hour ago
	layer := &CachedLayer{
		Digest:     "sha256:touchtest",
		Name:       "test",
		Type:       LayerTypeBase,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  4,
		FetchedAt:  initialTime,
		LastUsedAt: initialTime,
	}

	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Touch the layer
	if err := cache.Touch("sha256:touchtest"); err != nil {
		t.Fatalf("Touch() failed: %v", err)
	}

	// Get the layer and verify LastUsedAt was updated
	cached, err := cache.Get("sha256:touchtest")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if !cached.LastUsedAt.After(initialTime) {
		t.Errorf("LastUsedAt was not updated: %v (should be after %v)", cached.LastUsedAt, initialTime)
	}
}

// TestLayerCache_Stats verifies cache statistics
func TestLayerCache_Stats(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Add multiple layers
	for i := 0; i < 3; i++ {
		testSubdir := filepath.Join(dir, "test")
		os.MkdirAll(testSubdir, 0755)
		tarballPath := filepath.Join(testSubdir, fmt.Sprintf("layer%d.tar.gz", i))
		data := make([]byte, 1024*(i+1)) // Different sizes
		if err := os.WriteFile(tarballPath, data, 0644); err != nil {
			t.Fatalf("failed to create test tarball: %v", err)
		}

		layer := &CachedLayer{
			Digest:     fmt.Sprintf("sha256:layer%d", i),
			Name:       fmt.Sprintf("layer%d", i),
			Type:       LayerTypeRuntime,
			SourceURL:  "local://test",
			LocalPath:  tarballPath,
			SizeBytes:  int64(len(data)),
			FetchedAt:  time.Now(),
			LastUsedAt: time.Now().Add(time.Duration(-i) * time.Hour), // Different access times
		}

		if err := cache.Put(layer); err != nil {
			t.Fatalf("Put() failed for layer %d: %v", i, err)
		}
	}

	// Get stats
	stats, err := cache.Stats()
	if err != nil {
		t.Fatalf("Stats() failed: %v", err)
	}

	if stats.TotalLayers != 3 {
		t.Errorf("TotalLayers = %d, want 3", stats.TotalLayers)
	}

	expectedBytes := int64(1024 + 2048 + 3072)
	if stats.TotalBytes != expectedBytes {
		t.Errorf("TotalBytes = %d, want %d", stats.TotalBytes, expectedBytes)
	}

	// OldestAccess should be roughly 2 hours ago (layer 2)
	if stats.OldestAccess.After(time.Now()) {
		t.Error("OldestAccess is in the future")
	}
}

// TestLayerCache_Evict_LRU verifies LRU eviction
func TestLayerCache_Evict_LRU(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Add 3 layers with different access times
	layers := []struct {
		digest    string
		ageHours  int
		sizeBytes int64
	}{
		{"sha256:newest", 1, 1024},  // Most recently used
		{"sha256:middle", 5, 2048},  // Middle
		{"sha256:oldest", 10, 3072}, // Least recently used
	}

	for _, l := range layers {
		tarballPath := filepath.Join(dir, l.digest+".tar.gz")
		data := make([]byte, l.sizeBytes)
		if err := os.WriteFile(tarballPath, data, 0644); err != nil {
			t.Fatalf("failed to create test tarball: %v", err)
		}

		layer := &CachedLayer{
			Digest:     l.digest,
			Name:       l.digest,
			Type:       LayerTypeRuntime,
			SourceURL:  "local://test",
			LocalPath:  tarballPath,
			SizeBytes:  l.sizeBytes,
			FetchedAt:  time.Now(),
			LastUsedAt: time.Now().Add(time.Duration(-l.ageHours) * time.Hour),
		}

		if err := cache.Put(layer); err != nil {
			t.Fatalf("Put() failed: %v", err)
		}
	}

	// Evict enough to free at least 3072 bytes (should remove oldest)
	freed, err := cache.Evict(3072)
	if err != nil {
		t.Fatalf("Evict() failed: %v", err)
	}

	if freed < 3072 {
		t.Errorf("Evict() freed %d bytes, want at least 3072", freed)
	}

	// Oldest should be gone
	exists, err := cache.Exists("sha256:oldest")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if exists {
		t.Error("oldest layer should have been evicted")
	}

	// Newer ones should still exist
	exists, err = cache.Exists("sha256:newest")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if !exists {
		t.Error("newest layer should not have been evicted")
	}

	exists, err = cache.Exists("sha256:middle")
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if !exists {
		t.Error("middle layer should not have been evicted")
	}
}

// TestLayerCache_Evict_Multiple verifies evicting multiple layers
func TestLayerCache_Evict_Multiple(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Add 5 layers, each 1KB, with different ages
	evictSubdir := filepath.Join(dir, "evict")
	os.MkdirAll(evictSubdir, 0755)
	for i := 0; i < 5; i++ {
		tarballPath := filepath.Join(evictSubdir, fmt.Sprintf("layer%d.tar.gz", i))
		data := make([]byte, 1024)
		if err := os.WriteFile(tarballPath, data, 0644); err != nil {
			t.Fatalf("failed to create test tarball: %v", err)
		}

		layer := &CachedLayer{
			Digest:     fmt.Sprintf("sha256:evict%d", i),
			Name:       fmt.Sprintf("evict%d", i),
			Type:       LayerTypeRuntime,
			SourceURL:  "local://test",
			LocalPath:  tarballPath,
			SizeBytes:  1024,
			FetchedAt:  time.Now(),
			LastUsedAt: time.Now().Add(time.Duration(-(i + 1)) * time.Hour),
		}

		if err := cache.Put(layer); err != nil {
			t.Fatalf("Put() failed: %v", err)
		}
	}

	// Evict 3KB worth (should remove 3 oldest layers)
	freed, err := cache.Evict(3 * 1024)
	if err != nil {
		t.Fatalf("Evict() failed: %v", err)
	}

	if freed < 3*1024 {
		t.Errorf("Evict() freed %d bytes, want at least %d", freed, 3*1024)
	}

	// Verify stats show reduced count
	stats, err := cache.Stats()
	if err != nil {
		t.Fatalf("Stats() failed: %v", err)
	}

	if stats.TotalLayers > 2 {
		t.Errorf("TotalLayers = %d, want 2 or less after eviction", stats.TotalLayers)
	}
}

// TestLayerCache_Evict_EmptyCache verifies eviction on empty cache
func TestLayerCache_Evict_EmptyCache(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	freed, err := cache.Evict(1024)
	if err != nil {
		t.Errorf("Evict() on empty cache should not error: %v", err)
	}
	if freed != 0 {
		t.Errorf("Evict() on empty cache freed %d bytes, want 0", freed)
	}
}

// TestLayerCache_Put_UpdateExisting verifies updating an existing layer
func TestLayerCache_Put_UpdateExisting(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Add initial layer
	tarballPath1 := filepath.Join(dir, "v1.tar.gz")
	if err := os.WriteFile(tarballPath1, []byte("version1"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer1 := &CachedLayer{
		Digest:     "sha256:update",
		Name:       "test",
		Version:    "1.0.0",
		Type:       LayerTypeRuntime,
		SourceURL:  "local://v1",
		LocalPath:  tarballPath1,
		SizeBytes:  8,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	if err := cache.Put(layer1); err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// Update with new version
	tarballPath2 := filepath.Join(dir, "v2.tar.gz")
	if err := os.WriteFile(tarballPath2, []byte("version2-updated"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer2 := &CachedLayer{
		Digest:     "sha256:update", // Same digest
		Name:       "test",
		Version:    "2.0.0", // Different version
		Type:       LayerTypeRuntime,
		SourceURL:  "local://v2",
		LocalPath:  tarballPath2,
		SizeBytes:  16,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	if err := cache.Put(layer2); err != nil {
		t.Fatalf("Put() update failed: %v", err)
	}

	// Get and verify it was updated
	cached, err := cache.Get("sha256:update")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if cached.Version != "2.0.0" {
		t.Errorf("Version = %q, want '2.0.0' (should be updated)", cached.Version)
	}
	if cached.SizeBytes != 16 {
		t.Errorf("SizeBytes = %d, want 16 (should be updated)", cached.SizeBytes)
	}
}

// TestLayerCache_MetadataSerialization verifies JSON serialization of metadata
func TestLayerCache_MetadataSerialization(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	tarballPath := filepath.Join(dir, "test.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	metadata := &LayerPackage{
		Name:         "test-layer",
		Version:      "1.0.0",
		Description:  "Test layer with complex metadata",
		Type:         LayerTypeFeature,
		Dependencies: []string{"base", "runtime"},
		Provides:     []string{"recording", "debugging"},
		Files: []FileEntry{
			{Path: "/usr/bin/test", Mode: "0755", Owner: "root:root"},
		},
		Systemd: &SystemdConfig{
			Enable: []string{"test.service"},
			Mask:   []string{"unwanted.service"},
		},
		ConfigSchema: map[string]ConfigOption{
			"port": {Type: "integer", Default: 8080, Description: "Port number"},
		},
	}

	layer := &CachedLayer{
		Digest:     "sha256:metadata",
		Name:       "test-layer",
		Version:    "1.0.0",
		Type:       LayerTypeFeature,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  4,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
		Metadata:   metadata,
	}

	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// Get and verify metadata was preserved
	cached, err := cache.Get("sha256:metadata")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if cached.Metadata == nil {
		t.Fatal("Metadata is nil")
	}

	// Verify complex fields
	if len(cached.Metadata.Dependencies) != 2 {
		t.Errorf("len(Dependencies) = %d, want 2", len(cached.Metadata.Dependencies))
	}
	if len(cached.Metadata.Files) != 1 {
		t.Errorf("len(Files) = %d, want 1", len(cached.Metadata.Files))
	}
	if cached.Metadata.Systemd == nil || len(cached.Metadata.Systemd.Enable) != 1 {
		t.Error("Systemd config not preserved")
	}
	// ConfigSchema default values come back as float64 from JSON unmarshaling
	if portOpt, ok := cached.Metadata.ConfigSchema["port"]; !ok {
		t.Error("ConfigSchema port not preserved")
	} else if portDefault, ok := portOpt.Default.(float64); !ok || portDefault != 8080 {
		t.Errorf("ConfigSchema port default = %v (type %T), want 8080", portOpt.Default, portOpt.Default)
	}
}

// TestLayerCache_NilMetadata verifies handling of nil metadata
func TestLayerCache_NilMetadata(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	tarballPath := filepath.Join(dir, "test.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer := &CachedLayer{
		Digest:     "sha256:nometa",
		Name:       "test",
		Type:       LayerTypeBase,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  4,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
		Metadata:   nil, // Explicitly nil
	}

	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() with nil metadata failed: %v", err)
	}

	cached, err := cache.Get("sha256:nometa")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if cached.Metadata != nil {
		t.Error("Metadata should be nil")
	}
}

// TestLayerCache_InvalidDigest verifies error handling for invalid operations
func TestLayerCache_InvalidDigest(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Touch non-existent layer should not error (idempotent)
	if err := cache.Touch("sha256:nonexistent"); err != nil {
		t.Errorf("Touch() on non-existent layer should not error: %v", err)
	}
}

// TestLayerCache_ConcurrentAccess verifies basic thread safety
func TestLayerCache_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Add a layer
	tarballPath := filepath.Join(dir, "test.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer := &CachedLayer{
		Digest:     "sha256:concurrent",
		Name:       "test",
		Type:       LayerTypeBase,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  4,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// Concurrent reads and touches
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Multiple operations
			cache.Exists("sha256:concurrent")
			cache.Get("sha256:concurrent")
			cache.Touch("sha256:concurrent")
			cache.Stats()
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify cache is still consistent
	exists, err := cache.Exists("sha256:concurrent")
	if err != nil {
		t.Errorf("Cache consistency check failed: %v", err)
	}
	if !exists {
		t.Error("Layer disappeared during concurrent access")
	}
}

// TestLayerCache_FileCleanup verifies tarball files are cleaned up on eviction
func TestLayerCache_FileCleanup(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	cache, err := NewLayerCache(cacheDir, filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Add a layer
	tarballPath := filepath.Join(dir, "source.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("test data"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer := &CachedLayer{
		Digest:     "sha256:cleanup",
		Name:       "test",
		Type:       LayerTypeBase,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  9,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now().Add(-10 * time.Hour),
	}

	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// Verify cached file exists
	cachedPath := filepath.Join(cacheDir, "sha256:cleanup.tar.gz")
	if _, err := os.Stat(cachedPath); err != nil {
		t.Fatalf("cached file should exist: %v", err)
	}

	// Evict the layer
	freed, err := cache.Evict(9)
	if err != nil {
		t.Fatalf("Evict() failed: %v", err)
	}
	if freed < 9 {
		t.Errorf("Evict() freed %d bytes, want at least 9", freed)
	}

	// Verify cached file was deleted
	if _, err := os.Stat(cachedPath); !os.IsNotExist(err) {
		t.Error("cached file should have been deleted on eviction")
	}
}

// Benchmark basic cache operations
func BenchmarkLayerCache_Put(b *testing.B) {
	dir := b.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 10*1024*1024*1024)
	if err != nil {
		b.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	tarballPath := filepath.Join(dir, "bench.tar.gz")
	data := make([]byte, 1024*1024) // 1MB
	if err := os.WriteFile(tarballPath, data, 0644); err != nil {
		b.Fatalf("failed to create test tarball: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layer := &CachedLayer{
			Digest:     "sha256:bench" + string(rune(i)),
			Name:       "bench",
			Type:       LayerTypeRuntime,
			SourceURL:  "local://test",
			LocalPath:  tarballPath,
			SizeBytes:  int64(len(data)),
			FetchedAt:  time.Now(),
			LastUsedAt: time.Now(),
		}
		cache.Put(layer)
	}
}

func BenchmarkLayerCache_Get(b *testing.B) {
	dir := b.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 10*1024*1024*1024)
	if err != nil {
		b.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	// Pre-populate
	tarballPath := filepath.Join(dir, "bench.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("bench"), 0644); err != nil {
		b.Fatalf("failed to create test tarball: %v", err)
	}

	layer := &CachedLayer{
		Digest:     "sha256:bench",
		Name:       "bench",
		Type:       LayerTypeRuntime,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  5,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}
	cache.Put(layer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("sha256:bench")
	}
}

// TestLayerCache_EmptyVersion verifies handling of empty version string
func TestLayerCache_EmptyVersion(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewLayerCache(filepath.Join(dir, "cache"), filepath.Join(dir, "test.db"), 1024*1024*1024)
	if err != nil {
		t.Fatalf("NewLayerCache failed: %v", err)
	}
	defer cache.Close()

	tarballPath := filepath.Join(dir, "test.tar.gz")
	if err := os.WriteFile(tarballPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	layer := &CachedLayer{
		Digest:     "sha256:noversion",
		Name:       "test",
		Version:    "", // Empty version
		Type:       LayerTypeBase,
		SourceURL:  "local://test",
		LocalPath:  tarballPath,
		SizeBytes:  4,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	if err := cache.Put(layer); err != nil {
		t.Fatalf("Put() with empty version failed: %v", err)
	}

	cached, err := cache.Get("sha256:noversion")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if cached.Version != "" {
		t.Errorf("Version = %q, want empty string", cached.Version)
	}
}

// TestLayerCache_MetadataJSONMarshaling tests direct JSON operations
func TestLayerCache_MetadataJSONMarshaling(t *testing.T) {
	metadata := &LayerPackage{
		Name:         "test",
		Version:      "1.0.0",
		Description:  "Test layer",
		Type:         LayerTypeRuntime,
		Dependencies: []string{"base"},
		Provides:     []string{"python"},
	}

	// Marshal to JSON
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded LayerPackage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify roundtrip
	if decoded.Name != metadata.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, metadata.Name)
	}
	if len(decoded.Dependencies) != len(metadata.Dependencies) {
		t.Errorf("Dependencies length mismatch")
	}
}
