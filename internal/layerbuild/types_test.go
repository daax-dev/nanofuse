package layerbuild

import (
	"errors"
	"testing"
)

func TestErrNotFound(t *testing.T) {
	// Verify ErrNotFound can be used with errors.Is
	err := ErrNotFound
	if !errors.Is(err, ErrNotFound) {
		t.Error("ErrNotFound should match itself with errors.Is")
	}

	// Verify wrapped error can be unwrapped
	wrapped := errors.New("cache: " + ErrNotFound.Error())
	if errors.Is(wrapped, ErrNotFound) {
		t.Error("Plain wrapped error should not match ErrNotFound")
	}
}

func TestValidateSHA256(t *testing.T) {
	validSHA := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	tests := []struct {
		name     string
		sha256   string
		required bool
		wantErr  bool
	}{
		{"valid lowercase", validSHA, true, false},
		{"valid uppercase", "ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789", true, false},
		{"valid mixed case", "AbCdEf0123456789abcdef0123456789ABCDEF0123456789abcdef0123456789", true, false},
		{"empty required", "", true, true},
		{"empty optional", "", false, false},
		{"too short", "abcdef", true, true},
		{"too long", validSHA + "00", true, true},
		{"invalid hex char g", "gbcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", true, true},
		{"invalid hex char z", "zbcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", true, true},
		{"63 chars", validSHA[:63], true, true},
		{"65 chars", validSHA + "a", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSHA256(tt.sha256, tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSHA256() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLayerType_String(t *testing.T) {
	tests := []struct {
		name     string
		lt       LayerType
		expected string
	}{
		{"base", LayerTypeBase, "base"},
		{"runtime", LayerTypeRuntime, "runtime"},
		{"feature", LayerTypeFeature, "feature"},
		{"application", LayerTypeApplication, "application"},
		{"custom", LayerType("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lt.String(); got != tt.expected {
				t.Errorf("LayerType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLayerType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		lt       LayerType
		expected bool
	}{
		{"base is valid", LayerTypeBase, true},
		{"runtime is valid", LayerTypeRuntime, true},
		{"feature is valid", LayerTypeFeature, true},
		{"application is valid", LayerTypeApplication, true},
		{"empty is invalid", LayerType(""), false},
		{"unknown is invalid", LayerType("unknown"), false},
		{"custom is invalid", LayerType("custom"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lt.IsValid(); got != tt.expected {
				t.Errorf("LayerType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLayerType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		lt      LayerType
		wantErr bool
	}{
		{"base is valid", LayerTypeBase, false},
		{"runtime is valid", LayerTypeRuntime, false},
		{"feature is valid", LayerTypeFeature, false},
		{"application is valid", LayerTypeApplication, false},
		{"empty is invalid", LayerType(""), true},
		{"unknown is invalid", LayerType("unknown"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lt.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LayerType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKernelConfig_Validate(t *testing.T) {
	validSHA := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	tests := []struct {
		name    string
		config  KernelConfig
		wantErr bool
	}{
		{
			name: "valid kernel config",
			config: KernelConfig{
				Version: "5.10",
				Source:  "https://example.com/kernel",
				SHA256:  validSHA,
				Cmdline: "console=ttyS0",
			},
			wantErr: false,
		},
		{
			name: "missing version",
			config: KernelConfig{
				Source:  "https://example.com/kernel",
				SHA256:  validSHA,
				Cmdline: "console=ttyS0",
			},
			wantErr: true,
		},
		{
			name: "missing source",
			config: KernelConfig{
				Version: "5.10",
				SHA256:  validSHA,
				Cmdline: "console=ttyS0",
			},
			wantErr: true,
		},
		{
			name: "missing sha256",
			config: KernelConfig{
				Version: "5.10",
				Source:  "https://example.com/kernel",
				Cmdline: "console=ttyS0",
			},
			wantErr: true,
		},
		{
			name: "invalid sha256 length",
			config: KernelConfig{
				Version: "5.10",
				Source:  "https://example.com/kernel",
				SHA256:  "short",
				Cmdline: "console=ttyS0",
			},
			wantErr: true,
		},
		{
			name: "invalid sha256 characters",
			config: KernelConfig{
				Version: "5.10",
				Source:  "https://example.com/kernel",
				SHA256:  "zzzzzz0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				Cmdline: "console=ttyS0",
			},
			wantErr: true,
		},
		{
			name: "empty cmdline is allowed",
			config: KernelConfig{
				Version: "5.10",
				Source:  "https://example.com/kernel",
				SHA256:  validSHA,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("KernelConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOutputConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  OutputConfig
		wantErr bool
	}{
		{
			name: "valid ext4 output",
			config: OutputConfig{
				Path:        "/output/image.ext4",
				Format:      "ext4",
				Compression: "gzip",
			},
			wantErr: false,
		},
		{
			name: "valid squashfs output",
			config: OutputConfig{
				Path:   "/output/image.squashfs",
				Format: "squashfs",
			},
			wantErr: false,
		},
		{
			name: "valid tar output",
			config: OutputConfig{
				Path:   "/output/image.tar",
				Format: "tar",
			},
			wantErr: false,
		},
		{
			name: "missing path",
			config: OutputConfig{
				Format: "ext4",
			},
			wantErr: true,
		},
		{
			name: "missing format",
			config: OutputConfig{
				Path: "/output/image.ext4",
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			config: OutputConfig{
				Path:   "/output/image.raw",
				Format: "raw",
			},
			wantErr: true,
		},
		{
			name: "format case insensitive",
			config: OutputConfig{
				Path:   "/output/image.ext4",
				Format: "EXT4",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OutputConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLayerReference_Validate(t *testing.T) {
	validSHA := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	tests := []struct {
		name    string
		ref     LayerReference
		wantErr bool
	}{
		{
			name: "valid layer reference",
			ref: LayerReference{
				Name:   "alpine-base",
				Type:   LayerTypeBase,
				Source: "docker://alpine:3.18",
				SHA256: validSHA,
			},
			wantErr: false,
		},
		{
			name: "valid without sha256",
			ref: LayerReference{
				Name:   "alpine-base",
				Type:   LayerTypeBase,
				Source: "docker://alpine:3.18",
			},
			wantErr: false,
		},
		{
			name: "valid with condition",
			ref: LayerReference{
				Name:      "debug-tools",
				Type:      LayerTypeFeature,
				Source:    "file:///layers/debug.tar",
				Condition: "DEBUG_MODE=true",
			},
			wantErr: false,
		},
		{
			name: "valid with config",
			ref: LayerReference{
				Name:   "python-runtime",
				Type:   LayerTypeRuntime,
				Source: "docker://python:3.11",
				Config: map[string]any{"version": "3.11"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			ref: LayerReference{
				Type:   LayerTypeBase,
				Source: "docker://alpine:3.18",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			ref: LayerReference{
				Name:   "alpine-base",
				Type:   LayerType("invalid"),
				Source: "docker://alpine:3.18",
			},
			wantErr: true,
		},
		{
			name: "missing source",
			ref: LayerReference{
				Name: "alpine-base",
				Type: LayerTypeBase,
			},
			wantErr: true,
		},
		{
			name: "invalid sha256 length",
			ref: LayerReference{
				Name:   "alpine-base",
				Type:   LayerTypeBase,
				Source: "docker://alpine:3.18",
				SHA256: "short",
			},
			wantErr: true,
		},
		{
			name: "invalid sha256 characters",
			ref: LayerReference{
				Name:   "alpine-base",
				Type:   LayerTypeBase,
				Source: "docker://alpine:3.18",
				SHA256: "zzzzzz0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ref.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LayerReference.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLayerPackage_Validate(t *testing.T) {
	validSHA := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	tests := []struct {
		name    string
		pkg     LayerPackage
		wantErr bool
	}{
		{
			name: "valid layer package",
			pkg: LayerPackage{
				Name:     "alpine-base",
				Version:  "3.18",
				Type:     LayerTypeBase,
				Size:     1024000,
				SHA256:   validSHA,
				RootFS:   "/cache/layers/alpine-base/rootfs",
				Metadata: map[string]string{"arch": "amd64"},
			},
			wantErr: false,
		},
		{
			name: "valid without metadata",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    1024000,
				SHA256:  validSHA,
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			pkg: LayerPackage{
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    1024000,
				SHA256:  validSHA,
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "missing version",
			pkg: LayerPackage{
				Name:   "alpine-base",
				Type:   LayerTypeBase,
				Size:   1024000,
				SHA256: validSHA,
				RootFS: "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerType("invalid"),
				Size:    1024000,
				SHA256:  validSHA,
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "zero size",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    0,
				SHA256:  validSHA,
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "negative size",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    -100,
				SHA256:  validSHA,
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "missing sha256",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    1024000,
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "invalid sha256 length",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    1024000,
				SHA256:  "short",
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "invalid sha256 characters",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    1024000,
				SHA256:  "zzzzzz0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				RootFS:  "/cache/layers/alpine-base/rootfs",
			},
			wantErr: true,
		},
		{
			name: "missing rootfs",
			pkg: LayerPackage{
				Name:    "alpine-base",
				Version: "3.18",
				Type:    LayerTypeBase,
				Size:    1024000,
				SHA256:  validSHA,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pkg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LayerPackage.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageManifest_Validate(t *testing.T) {
	validSHA := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	tests := []struct {
		name     string
		manifest ImageManifest
		wantErr  bool
	}{
		{
			name: "valid manifest",
			manifest: ImageManifest{
				Version: "1.0",
				Name:    "test-image",
				Kernel: KernelConfig{
					Version: "5.10",
					Source:  "https://example.com/kernel",
					SHA256:  validSHA,
					Cmdline: "console=ttyS0",
				},
				Layers: []LayerReference{
					{
						Name:   "alpine-base",
						Type:   LayerTypeBase,
						Source: "docker://alpine:3.18",
					},
				},
				Output: OutputConfig{
					Path:   "/output/image.ext4",
					Format: "ext4",
				},
			},
			wantErr: false,
		},
		{
			name: "valid with multiple layers",
			manifest: ImageManifest{
				Version: "1.0",
				Name:    "test-image",
				Kernel: KernelConfig{
					Version: "5.10",
					Source:  "https://example.com/kernel",
					SHA256:  validSHA,
					Cmdline: "console=ttyS0",
				},
				Layers: []LayerReference{
					{
						Name:   "alpine-base",
						Type:   LayerTypeBase,
						Source: "docker://alpine:3.18",
					},
					{
						Name:   "python-runtime",
						Type:   LayerTypeRuntime,
						Source: "docker://python:3.11",
					},
					{
						Name:   "my-app",
						Type:   LayerTypeApplication,
						Source: "file:///layers/app.tar",
					},
				},
				Output: OutputConfig{
					Path:   "/output/image.ext4",
					Format: "ext4",
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			manifest: ImageManifest{
				Name: "test-image",
				Kernel: KernelConfig{
					Version: "5.10",
					Source:  "https://example.com/kernel",
					SHA256:  validSHA,
				},
				Layers: []LayerReference{
					{
						Name:   "alpine-base",
						Type:   LayerTypeBase,
						Source: "docker://alpine:3.18",
					},
				},
				Output: OutputConfig{
					Path:   "/output/image.ext4",
					Format: "ext4",
				},
			},
			wantErr: true,
		},
		{
			name: "missing name",
			manifest: ImageManifest{
				Version: "1.0",
				Kernel: KernelConfig{
					Version: "5.10",
					Source:  "https://example.com/kernel",
					SHA256:  validSHA,
				},
				Layers: []LayerReference{
					{
						Name:   "alpine-base",
						Type:   LayerTypeBase,
						Source: "docker://alpine:3.18",
					},
				},
				Output: OutputConfig{
					Path:   "/output/image.ext4",
					Format: "ext4",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid kernel config",
			manifest: ImageManifest{
				Version: "1.0",
				Name:    "test-image",
				Kernel: KernelConfig{
					Version: "5.10",
					// Missing required fields
				},
				Layers: []LayerReference{
					{
						Name:   "alpine-base",
						Type:   LayerTypeBase,
						Source: "docker://alpine:3.18",
					},
				},
				Output: OutputConfig{
					Path:   "/output/image.ext4",
					Format: "ext4",
				},
			},
			wantErr: true,
		},
		{
			name: "no layers",
			manifest: ImageManifest{
				Version: "1.0",
				Name:    "test-image",
				Kernel: KernelConfig{
					Version: "5.10",
					Source:  "https://example.com/kernel",
					SHA256:  validSHA,
				},
				Layers: []LayerReference{},
				Output: OutputConfig{
					Path:   "/output/image.ext4",
					Format: "ext4",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid layer reference",
			manifest: ImageManifest{
				Version: "1.0",
				Name:    "test-image",
				Kernel: KernelConfig{
					Version: "5.10",
					Source:  "https://example.com/kernel",
					SHA256:  validSHA,
				},
				Layers: []LayerReference{
					{
						Name: "alpine-base",
						Type: LayerTypeBase,
						// Missing source
					},
				},
				Output: OutputConfig{
					Path:   "/output/image.ext4",
					Format: "ext4",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid output config",
			manifest: ImageManifest{
				Version: "1.0",
				Name:    "test-image",
				Kernel: KernelConfig{
					Version: "5.10",
					Source:  "https://example.com/kernel",
					SHA256:  validSHA,
				},
				Layers: []LayerReference{
					{
						Name:   "alpine-base",
						Type:   LayerTypeBase,
						Source: "docker://alpine:3.18",
					},
				},
				Output: OutputConfig{
					Path: "/output/image.ext4",
					// Missing format
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageManifest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Mock implementations for interface testing

type mockLayerFetcher struct {
	supportsFunc func(sourceType SourceType) bool
	fetchFunc    func(source string) (*CachedLayer, error)
}

func (m *mockLayerFetcher) Supports(sourceType SourceType) bool {
	if m.supportsFunc != nil {
		return m.supportsFunc(sourceType)
	}
	return true
}

func (m *mockLayerFetcher) Fetch(source string) (*CachedLayer, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(source)
	}
	return nil, nil
}

type mockLayerCache struct {
	getFunc    func(digest string) (*CachedLayer, error)
	putFunc    func(layer *CachedLayer) error
	existsFunc func(digest string) (bool, error)
	touchFunc  func(digest string) error
	evictFunc  func(targetBytes int64) (int64, error)
	statsFunc  func() (*CacheStats, error)
	closeFunc  func() error
}

func (m *mockLayerCache) Get(digest string) (*CachedLayer, error) {
	if m.getFunc != nil {
		return m.getFunc(digest)
	}
	return nil, nil
}

func (m *mockLayerCache) Put(layer *CachedLayer) error {
	if m.putFunc != nil {
		return m.putFunc(layer)
	}
	return nil
}

func (m *mockLayerCache) Exists(digest string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(digest)
	}
	return false, nil
}

func (m *mockLayerCache) Touch(digest string) error {
	if m.touchFunc != nil {
		return m.touchFunc(digest)
	}
	return nil
}

func (m *mockLayerCache) Evict(targetBytes int64) (int64, error) {
	if m.evictFunc != nil {
		return m.evictFunc(targetBytes)
	}
	return 0, nil
}

func (m *mockLayerCache) Stats() (*CacheStats, error) {
	if m.statsFunc != nil {
		return m.statsFunc()
	}
	return &CacheStats{}, nil
}

func (m *mockLayerCache) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestLayerFetcher_Interface(t *testing.T) {
	// Ensure mockLayerFetcher implements LayerFetcher
	var _ LayerFetcher = (*mockLayerFetcher)(nil)

	validSHA := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	fetcher := &mockLayerFetcher{
		supportsFunc: func(sourceType SourceType) bool {
			return sourceType == SourceTypeLocal
		},
		fetchFunc: func(source string) (*CachedLayer, error) {
			return &CachedLayer{
				Digest:    validSHA,
				Name:      "test-layer",
				Version:   "1.0",
				Type:      LayerTypeBase,
				SizeBytes: 1024,
				SourceURL: source,
				LocalPath: "/tmp/test.tar.gz",
			}, nil
		},
	}

	// Test Supports
	if !fetcher.Supports(SourceTypeLocal) {
		t.Error("Supports(SourceTypeLocal) = false, want true")
	}
	if fetcher.Supports(SourceTypeDocker) {
		t.Error("Supports(SourceTypeDocker) = true, want false")
	}

	// Test Fetch
	layer, err := fetcher.Fetch("local:///test")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if layer == nil {
		t.Fatal("Fetch() returned nil layer")
	}
	if layer.Name != "test-layer" {
		t.Errorf("Fetch() layer name = %v, want test-layer", layer.Name)
	}
}

func TestLayerCache_Interface(t *testing.T) {
	// Ensure mockLayerCache implements LayerCache
	var _ LayerCache = (*mockLayerCache)(nil)

	validSHA := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	testLayer := &CachedLayer{
		Digest:    validSHA,
		Name:      "test-layer",
		Version:   "1.0",
		Type:      LayerTypeBase,
		SizeBytes: 1024,
		LocalPath: "/tmp/test.tar.gz",
	}

	cache := &mockLayerCache{
		putFunc: func(layer *CachedLayer) error {
			return nil
		},
		getFunc: func(digest string) (*CachedLayer, error) {
			if digest == validSHA {
				return testLayer, nil
			}
			return nil, nil
		},
		existsFunc: func(digest string) (bool, error) {
			return digest == validSHA, nil
		},
	}

	// Test Put
	if err := cache.Put(testLayer); err != nil {
		t.Errorf("Put() error = %v", err)
	}

	// Test Get
	layer, err := cache.Get(validSHA)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if layer == nil {
		t.Error("Get() returned nil layer")
	}

	// Test Exists
	exists, err := cache.Exists(validSHA)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}

	// Test Exists with non-existent
	exists, err = cache.Exists("nonexistent")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false for non-existent")
	}
}
