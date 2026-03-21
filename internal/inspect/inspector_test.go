package inspect

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewInspector(t *testing.T) {
	t.Parallel()

	inspector := NewInspector("/tmp/test", true)
	if inspector == nil {
		t.Fatal("NewInspector returned nil")
	}
	if inspector.workDir != "/tmp/test" {
		t.Errorf("workDir = %q, want %q", inspector.workDir, "/tmp/test")
	}
	if !inspector.verbose {
		t.Error("verbose should be true")
	}
}

func TestInspectResult_JSONSerialization(t *testing.T) {
	t.Parallel()

	builtAt := time.Date(2025, 12, 22, 10, 0, 0, 0, time.UTC)
	appliedAt := time.Date(2025, 12, 22, 10, 0, 5, 0, time.UTC)

	result := &InspectResult{
		Name:           "nanofuse-flowspec",
		BuiltAt:        builtAt,
		TotalSizeBytes: 600000000,
		HasMetadata:    true,
		ManifestPath:   "/etc/nanofuse/build-manifest.json",
		Kernel: KernelInfo{
			Version: "6.1.90",
			Cmdline: "console=ttyS0 reboot=k panic=1 pci=off",
		},
		Layers: []LayerInfo{
			{
				Name:      "base-os",
				Version:   "1.0.0",
				Digest:    "sha256:abc123def456",
				Type:      "base",
				SizeBytes: 500000000,
				AppliedAt: appliedAt,
			},
		},
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	// Verify JSON structure
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"name": "nanofuse-flowspec"`) {
		t.Error("JSON missing name field")
	}
	if !strings.Contains(jsonStr, `"total_size_bytes": 600000000`) {
		t.Error("JSON missing total_size_bytes field")
	}
	if !strings.Contains(jsonStr, `"version": "6.1.90"`) {
		t.Error("JSON missing kernel version field")
	}
	if !strings.Contains(jsonStr, `"base-os"`) {
		t.Error("JSON missing layer name")
	}

	// Deserialize back
	var decoded InspectResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if decoded.Name != result.Name {
		t.Errorf("decoded.Name = %q, want %q", decoded.Name, result.Name)
	}
	if decoded.TotalSizeBytes != result.TotalSizeBytes {
		t.Errorf("decoded.TotalSizeBytes = %d, want %d", decoded.TotalSizeBytes, result.TotalSizeBytes)
	}
	if len(decoded.Layers) != 1 {
		t.Errorf("decoded.Layers length = %d, want 1", len(decoded.Layers))
	}
}

func TestFormatText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     *InspectResult
		showLayers bool
		wantParts  []string
	}{
		{
			name: "with metadata and layers",
			result: &InspectResult{
				Name:           "test-image",
				BuiltAt:        time.Date(2025, 12, 22, 10, 0, 0, 0, time.UTC),
				TotalSizeBytes: 1073741824, // 1GB
				HasMetadata:    true,
				Kernel: KernelInfo{
					Version: "6.1.90",
					Cmdline: "console=ttyS0",
				},
				Layers: []LayerInfo{
					{Name: "base", Version: "1.0.0", Type: "base"},
					{Name: "app", Version: "2.0.0", Type: "application"},
				},
			},
			showLayers: true,
			wantParts: []string{
				"Image:         test-image",
				"Total Size:    1.0 GB",
				"Has Metadata:  true",
				"Kernel:",
				"Version:     6.1.90",
				"Layers:        2",
				"Layer Details:",
				"1. base",
				"2. app",
			},
		},
		{
			name: "without metadata",
			result: &InspectResult{
				Name:           "legacy-image.ext4",
				BuiltAt:        time.Date(2025, 12, 22, 10, 0, 0, 0, time.UTC),
				TotalSizeBytes: 536870912, // 512MB
				HasMetadata:    false,
				Layers:         []LayerInfo{},
			},
			showLayers: false,
			wantParts: []string{
				"Image:         legacy-image.ext4",
				"Has Metadata:  false",
				"does not contain NanoFuse layer metadata",
			},
		},
		{
			name: "layers hidden",
			result: &InspectResult{
				Name:           "test-image",
				BuiltAt:        time.Date(2025, 12, 22, 10, 0, 0, 0, time.UTC),
				TotalSizeBytes: 1073741824,
				HasMetadata:    true,
				Kernel:         KernelInfo{Version: "6.1.90"},
				Layers: []LayerInfo{
					{Name: "base", Version: "1.0.0"},
				},
			},
			showLayers: false,
			wantParts:  []string{"Layers:        1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := FormatText(tc.result, tc.showLayers, &buf)
			if err != nil {
				t.Fatalf("FormatText returned error: %v", err)
			}

			output := buf.String()
			for _, want := range tc.wantParts {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestFormatText_LayersNotShownByDefault(t *testing.T) {
	t.Parallel()

	result := &InspectResult{
		Name:           "test-image",
		BuiltAt:        time.Now(),
		TotalSizeBytes: 1073741824,
		HasMetadata:    true,
		Kernel:         KernelInfo{Version: "6.1.90"},
		Layers: []LayerInfo{
			{Name: "base", Version: "1.0.0", Type: "base"},
		},
	}

	var buf bytes.Buffer
	err := FormatText(result, false, &buf)
	if err != nil {
		t.Fatalf("FormatText returned error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Layer Details:") {
		t.Error("layer details should not be shown when showLayers=false")
	}
}

func TestFormatJSON(t *testing.T) {
	t.Parallel()

	result := &InspectResult{
		Name:           "test-image",
		BuiltAt:        time.Date(2025, 12, 22, 10, 0, 0, 0, time.UTC),
		TotalSizeBytes: 1073741824,
		HasMetadata:    true,
		Kernel: KernelInfo{
			Version: "6.1.90",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerInfo{
			{Name: "base", Version: "1.0.0", Digest: "sha256:abc123"},
		},
	}

	var buf bytes.Buffer
	err := FormatJSON(result, &buf)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	// Verify it's valid JSON
	var decoded InspectResult
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded.Name != result.Name {
		t.Errorf("decoded.Name = %q, want %q", decoded.Name, result.Name)
	}
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tc := range tests {
		got := formatBytes(tc.bytes)
		if got != tc.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

func TestInspector_InspectImage_FileNotFound(t *testing.T) {
	t.Parallel()

	inspector := NewInspector(t.TempDir(), false)
	_, err := inspector.InspectImage(context.Background(), "/nonexistent/path/image.ext4")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestInspector_InspectImage_DirectoryPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inspector := NewInspector(tmpDir, false)

	_, err := inspector.InspectImage(context.Background(), tmpDir)
	if err == nil {
		t.Error("expected error for directory path")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("error should mention 'directory', got: %v", err)
	}
}

func TestInspector_readManifestFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inspector := NewInspector(tmpDir, false)

	// Create a test manifest file
	manifest := BuildManifest{
		Version: "1.0",
		Name:    "test-image",
		BuiltAt: time.Date(2025, 12, 22, 10, 0, 0, 0, time.UTC),
		Kernel: BuildManifestKernel{
			Version: "6.1.90",
			Cmdline: "console=ttyS0",
		},
		Layers: []BuildManifestLayer{
			{
				Name:    "base",
				Version: "1.0.0",
				Digest:  "sha256:abc123",
				Type:    "base",
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "build-manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Read it back
	result, err := inspector.readManifestFile(manifestPath)
	if err != nil {
		t.Fatalf("readManifestFile returned error: %v", err)
	}

	if result.Name != manifest.Name {
		t.Errorf("result.Name = %q, want %q", result.Name, manifest.Name)
	}
	if result.Kernel.Version != manifest.Kernel.Version {
		t.Errorf("result.Kernel.Version = %q, want %q", result.Kernel.Version, manifest.Kernel.Version)
	}
	if len(result.Layers) != 1 {
		t.Errorf("len(result.Layers) = %d, want 1", len(result.Layers))
	}
}

func TestInspector_readManifestFile_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inspector := NewInspector(tmpDir, false)

	_, err := inspector.readManifestFile(filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestInspector_readManifestFile_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inspector := NewInspector(tmpDir, false)

	// Create an invalid JSON file
	manifestPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(manifestPath, []byte("not valid json {"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := inspector.readManifestFile(manifestPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should mention 'parse', got: %v", err)
	}
}

func TestBuildManifest_JSONCompatibility(t *testing.T) {
	t.Parallel()

	// Test JSON compatibility with the layerbuild package format
	jsonData := `{
		"version": "1.0",
		"name": "test-image",
		"built_at": "2025-12-22T10:00:00Z",
		"kernel": {
			"version": "6.1.90",
			"cmdline": "console=ttyS0 reboot=k panic=1"
		},
		"layers": [
			{
				"name": "base-os",
				"version": "1.0.0",
				"digest": "sha256:abc123def456",
				"type": "base",
				"applied_at": "2025-12-22T10:00:05Z"
			}
		]
	}`

	var manifest BuildManifest
	if err := json.Unmarshal([]byte(jsonData), &manifest); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if manifest.Name != "test-image" {
		t.Errorf("Name = %q, want %q", manifest.Name, "test-image")
	}
	if manifest.Kernel.Version != "6.1.90" {
		t.Errorf("Kernel.Version = %q, want %q", manifest.Kernel.Version, "6.1.90")
	}
	if len(manifest.Layers) != 1 {
		t.Fatalf("len(Layers) = %d, want 1", len(manifest.Layers))
	}
	if manifest.Layers[0].Name != "base-os" {
		t.Errorf("Layers[0].Name = %q, want %q", manifest.Layers[0].Name, "base-os")
	}
}
