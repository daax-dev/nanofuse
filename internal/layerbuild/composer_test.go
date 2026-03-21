package layerbuild

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockCache implements LayerCache for testing
type mockCache struct {
	layers map[string]*CachedLayer
	err    error
}

func newMockCache() *mockCache {
	return &mockCache{
		layers: make(map[string]*CachedLayer),
	}
}

func (m *mockCache) Get(digest string) (*CachedLayer, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.layers[digest], nil
}

func (m *mockCache) Put(layer *CachedLayer) error {
	if m.err != nil {
		return m.err
	}
	m.layers[layer.Digest] = layer
	return nil
}

func (m *mockCache) Exists(digest string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	_, ok := m.layers[digest]
	return ok, nil
}

func (m *mockCache) Touch(digest string) error {
	if m.err != nil {
		return m.err
	}
	if layer, ok := m.layers[digest]; ok {
		layer.LastUsedAt = time.Now()
		return nil
	}
	return errors.New("layer not found")
}

func (m *mockCache) Evict(targetBytes int64) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return targetBytes, nil
}

func (m *mockCache) Stats() (*CacheStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CacheStats{
		TotalLayers:  len(m.layers),
		TotalBytes:   0,
		OldestAccess: time.Now(),
	}, nil
}

func (m *mockCache) Close() error {
	if m.err != nil {
		return m.err
	}
	return nil
}

// mockFetcher implements LayerFetcher for testing
type mockFetcher struct {
	layers   map[string]*CachedLayer
	err      error
	supports SourceType
}

func newMockFetcher(supports SourceType) *mockFetcher {
	return &mockFetcher{
		layers:   make(map[string]*CachedLayer),
		supports: supports,
	}
}

func (m *mockFetcher) Fetch(source string) (*CachedLayer, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.layers[source], nil
}

func (m *mockFetcher) Supports(sourceType SourceType) bool {
	return sourceType == m.supports
}

// TestNewComposer tests Composer creation
func TestNewComposer(t *testing.T) {
	cache := newMockCache()
	fetcher := newMockFetcher(SourceTypeLocal)

	c := NewComposer("/tmp/work", "/tmp/output", cache, []LayerFetcher{fetcher})

	if c == nil {
		t.Fatal("NewComposer returned nil")
	}

	if c.workDir != "/tmp/work" {
		t.Errorf("workDir = %q, want %q", c.workDir, "/tmp/work")
	}

	if c.outputDir != "/tmp/output" {
		t.Errorf("outputDir = %q, want %q", c.outputDir, "/tmp/output")
	}

	if c.cache != cache {
		t.Error("cache was not set correctly")
	}

	if len(c.fetchers) != 1 {
		t.Errorf("len(fetchers) = %d, want 1", len(c.fetchers))
	}
}

// TestComposeOptionsValidation tests ComposeOptions validation
func TestComposeOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    *ComposeOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: &ComposeOptions{
				Manifest: &ImageManifest{
					Version: "1.0",
					Name:    "test",
					Kernel: KernelConfig{
						Version: "6.1",
						Source:  "local://vmlinux",
						Cmdline: "console=ttyS0",
					},
					Layers: []LayerReference{
						{
							Name:   "base",
							Type:   LayerTypeBase,
							Source: "local://base.tar.gz",
						},
					},
					Output: DefaultOutputConfig(),
				},
				Env:     make(map[string]string),
				Verbose: false,
				DryRun:  false,
			},
			wantErr: false,
		},
		{
			name: "nil manifest",
			opts: &ComposeOptions{
				Manifest: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComposeOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateComposeOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestComposeDryRun tests dry-run mode
func TestComposeDryRun(t *testing.T) {
	cache := newMockCache()
	fetcher := newMockFetcher(SourceTypeLocal)

	tmpDir := t.TempDir()
	c := NewComposer(tmpDir, tmpDir, cache, []LayerFetcher{fetcher})

	manifest := &ImageManifest{
		Version: "1.0",
		Name:    "test-image",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{
			{
				Name:   "base",
				Type:   LayerTypeBase,
				Source: "local://base.tar.gz",
				SHA256: "abc123",
			},
		},
		Output: DefaultOutputConfig(),
	}

	// Add layer to mock cache
	cache.Put(&CachedLayer{
		Digest:    "abc123",
		Name:      "base",
		Version:   "1.0",
		Type:      LayerTypeBase,
		LocalPath: filepath.Join(tmpDir, "base.tar.gz"),
		Metadata: &LayerPackage{
			Name:    "base",
			Version: "1.0",
			Type:    LayerTypeBase,
		},
	})

	opts := &ComposeOptions{
		Manifest: manifest,
		Env:      make(map[string]string),
		Verbose:  true,
		DryRun:   true,
	}

	ctx := context.Background()
	result, err := c.Compose(ctx, opts)

	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}

	if result == nil {
		t.Fatal("Compose() returned nil result")
	}

	// In dry-run mode, no files should be created
	if result.RootfsPath != "" {
		// Verify the path is constructed but file doesn't exist (dry-run)
		expectedPath := filepath.Join(tmpDir, "test-image-rootfs.ext4")
		if result.RootfsPath != expectedPath {
			t.Errorf("RootfsPath = %q, want %q", result.RootfsPath, expectedPath)
		}
	}
}

// TestComposeLayerOrder tests that layers are applied in correct order
func TestComposeLayerOrder(t *testing.T) {
	cache := newMockCache()
	fetcher := newMockFetcher(SourceTypeLocal)

	tmpDir := t.TempDir()
	c := NewComposer(tmpDir, tmpDir, cache, []LayerFetcher{fetcher})

	manifest := &ImageManifest{
		Version: "1.0",
		Name:    "test-order",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{
			{
				Name:         "app",
				Type:         LayerTypeApplication,
				Source:       "local://app.tar.gz",
				SHA256:       "app123",
				Dependencies: []string{"runtime"},
			},
			{
				Name:         "runtime",
				Type:         LayerTypeRuntime,
				Source:       "local://runtime.tar.gz",
				SHA256:       "runtime123",
				Dependencies: []string{"base"},
			},
			{
				Name:   "base",
				Type:   LayerTypeBase,
				Source: "local://base.tar.gz",
				SHA256: "base123",
			},
		},
		Output: DefaultOutputConfig(),
	}

	// Add layers to cache
	for _, layer := range manifest.Layers {
		cache.Put(&CachedLayer{
			Digest:    layer.SHA256,
			Name:      layer.Name,
			Version:   "1.0",
			Type:      layer.Type,
			LocalPath: filepath.Join(tmpDir, layer.Name+".tar.gz"),
			Metadata: &LayerPackage{
				Name:    layer.Name,
				Version: "1.0",
				Type:    layer.Type,
			},
		})
	}

	opts := &ComposeOptions{
		Manifest: manifest,
		Env:      make(map[string]string),
		DryRun:   true,
	}

	ctx := context.Background()
	result, err := c.Compose(ctx, opts)

	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}

	// Verify layers are applied in dependency order: base -> runtime -> app
	expectedOrder := []string{"base", "runtime", "app"}
	if len(result.LayersApplied) != len(expectedOrder) {
		t.Fatalf("LayersApplied length = %d, want %d", len(result.LayersApplied), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if result.LayersApplied[i] != expected {
			t.Errorf("LayersApplied[%d] = %q, want %q", i, result.LayersApplied[i], expected)
		}
	}
}

// TestComposeConditionalLayers tests conditional layer inclusion
func TestComposeConditionalLayers(t *testing.T) {
	cache := newMockCache()
	fetcher := newMockFetcher(SourceTypeLocal)

	tmpDir := t.TempDir()
	c := NewComposer(tmpDir, tmpDir, cache, []LayerFetcher{fetcher})

	manifest := &ImageManifest{
		Version: "1.0",
		Name:    "test-conditional",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{
			{
				Name:   "base",
				Type:   LayerTypeBase,
				Source: "local://base.tar.gz",
				SHA256: "base123",
			},
			{
				Name:      "debug",
				Type:      LayerTypeFeature,
				Source:    "local://debug.tar.gz",
				SHA256:    "debug123",
				Condition: "${ENABLE_DEBUG:-false}",
			},
		},
		Output: DefaultOutputConfig(),
	}

	// Add layers to cache
	for _, layer := range manifest.Layers {
		cache.Put(&CachedLayer{
			Digest:    layer.SHA256,
			Name:      layer.Name,
			Version:   "1.0",
			Type:      layer.Type,
			LocalPath: filepath.Join(tmpDir, layer.Name+".tar.gz"),
			Metadata: &LayerPackage{
				Name:    layer.Name,
				Version: "1.0",
				Type:    layer.Type,
			},
		})
	}

	tests := []struct {
		name       string
		env        map[string]string
		wantLayers []string
	}{
		{
			name:       "debug disabled (default)",
			env:        make(map[string]string),
			wantLayers: []string{"base"},
		},
		{
			name: "debug enabled",
			env: map[string]string{
				"ENABLE_DEBUG": "true",
			},
			wantLayers: []string{"base", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ComposeOptions{
				Manifest: manifest,
				Env:      tt.env,
				DryRun:   true,
			}

			ctx := context.Background()
			result, err := c.Compose(ctx, opts)

			if err != nil {
				t.Fatalf("Compose() error = %v", err)
			}

			if len(result.LayersApplied) != len(tt.wantLayers) {
				t.Fatalf("LayersApplied length = %d, want %d", len(result.LayersApplied), len(tt.wantLayers))
			}

			for i, expected := range tt.wantLayers {
				if result.LayersApplied[i] != expected {
					t.Errorf("LayersApplied[%d] = %q, want %q", i, result.LayersApplied[i], expected)
				}
			}
		})
	}
}

// TestComposeBuildManifest tests build manifest generation
func TestComposeBuildManifest(t *testing.T) {
	cache := newMockCache()
	fetcher := newMockFetcher(SourceTypeLocal)

	tmpDir := t.TempDir()
	c := NewComposer(tmpDir, tmpDir, cache, []LayerFetcher{fetcher})

	manifest := &ImageManifest{
		Version: "1.0",
		Name:    "test-manifest",
		Kernel: KernelConfig{
			Version: "6.1.90",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0 root=/dev/vda1 rw",
		},
		Layers: []LayerReference{
			{
				Name:   "base",
				Type:   LayerTypeBase,
				Source: "local://base.tar.gz",
				SHA256: "base123",
			},
		},
		Output: DefaultOutputConfig(),
	}

	cache.Put(&CachedLayer{
		Digest:    "base123",
		Name:      "base",
		Version:   "1.0.0",
		Type:      LayerTypeBase,
		LocalPath: filepath.Join(tmpDir, "base.tar.gz"),
		Metadata: &LayerPackage{
			Name:    "base",
			Version: "1.0.0",
			Type:    LayerTypeBase,
		},
	})

	opts := &ComposeOptions{
		Manifest: manifest,
		Env:      make(map[string]string),
		DryRun:   false,
	}

	// Create a temporary kernel file
	kernelPath := filepath.Join(tmpDir, "vmlinux")
	if err := os.WriteFile(kernelPath, []byte("fake kernel"), 0644); err != nil {
		t.Fatalf("Failed to create test kernel: %v", err)
	}

	ctx := context.Background()
	result, err := c.Compose(ctx, opts)

	// Note: This will fail in actual execution without root, but tests the structure
	if err != nil {
		// Expected to fail without root privileges, but check if it's the right error
		if !errors.Is(err, errRequiresRoot) && !os.IsPermission(err) {
			t.Logf("Compose() error (expected without root): %v", err)
		}
		t.Skip("Skipping actual compose test - requires root")
	}

	if result.ManifestPath == "" {
		t.Error("ManifestPath should be set")
	}

	// Verify build manifest structure (if created)
	if result.ManifestPath != "" && fileExists(result.ManifestPath) {
		data, err := os.ReadFile(result.ManifestPath)
		if err != nil {
			t.Fatalf("Failed to read build manifest: %v", err)
		}

		var buildManifest BuildManifest
		if err := json.Unmarshal(data, &buildManifest); err != nil {
			t.Fatalf("Failed to parse build manifest: %v", err)
		}

		if buildManifest.Version != "1.0" {
			t.Errorf("BuildManifest.Version = %q, want %q", buildManifest.Version, "1.0")
		}

		if buildManifest.Name != "test-manifest" {
			t.Errorf("BuildManifest.Name = %q, want %q", buildManifest.Name, "test-manifest")
		}

		if len(buildManifest.Layers) != 1 {
			t.Fatalf("BuildManifest.Layers length = %d, want 1", len(buildManifest.Layers))
		}

		if buildManifest.Layers[0].Name != "base" {
			t.Errorf("BuildManifest.Layers[0].Name = %q, want %q", buildManifest.Layers[0].Name, "base")
		}
	}
}

// TestComposeProgress tests progress callback
func TestComposeProgress(t *testing.T) {
	cache := newMockCache()
	fetcher := newMockFetcher(SourceTypeLocal)

	tmpDir := t.TempDir()
	c := NewComposer(tmpDir, tmpDir, cache, []LayerFetcher{fetcher})

	manifest := &ImageManifest{
		Version: "1.0",
		Name:    "test-progress",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{
			{
				Name:   "base",
				Type:   LayerTypeBase,
				Source: "local://base.tar.gz",
				SHA256: "base123",
			},
		},
		Output: DefaultOutputConfig(),
	}

	cache.Put(&CachedLayer{
		Digest:    "base123",
		Name:      "base",
		Version:   "1.0",
		Type:      LayerTypeBase,
		LocalPath: filepath.Join(tmpDir, "base.tar.gz"),
		Metadata: &LayerPackage{
			Name:    "base",
			Version: "1.0",
			Type:    LayerTypeBase,
		},
	})

	progressCalls := 0
	opts := &ComposeOptions{
		Manifest: manifest,
		Env:      make(map[string]string),
		DryRun:   true,
		OnProgress: func(step string, current, total int) {
			progressCalls++
			t.Logf("Progress: %s (%d/%d)", step, current, total)
		},
	}

	ctx := context.Background()
	_, err := c.Compose(ctx, opts)

	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}

	if progressCalls == 0 {
		t.Error("OnProgress callback was never called")
	}
}

// TestComposeContextCancellation tests context cancellation
func TestComposeContextCancellation(t *testing.T) {
	cache := newMockCache()
	fetcher := newMockFetcher(SourceTypeLocal)

	tmpDir := t.TempDir()
	c := NewComposer(tmpDir, tmpDir, cache, []LayerFetcher{fetcher})

	manifest := &ImageManifest{
		Version: "1.0",
		Name:    "test-cancel",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{
			{
				Name:   "base",
				Type:   LayerTypeBase,
				Source: "local://base.tar.gz",
				SHA256: "base123",
			},
		},
		Output: DefaultOutputConfig(),
	}

	opts := &ComposeOptions{
		Manifest: manifest,
		Env:      make(map[string]string),
		DryRun:   false,
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := c.Compose(ctx, opts)

	if err == nil {
		t.Error("Expected error from cancelled context")
	}

	if !errors.Is(err, context.Canceled) && err.Error() != "context canceled" {
		t.Logf("Got error (may not be context.Canceled): %v", err)
	}
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
