package layerbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// LayerSpec represents a layer.yaml file structure
type LayerSpec struct {
	Name         string                     `yaml:"name"`
	Version      string                     `yaml:"version"`
	SHA256       string                     `yaml:"sha256"`
	SizeMB       int                        `yaml:"size_mb"`
	Description  string                     `yaml:"description"`
	Type         string                     `yaml:"type"`
	Dependencies []string                   `yaml:"dependencies"`
	Provides     []string                   `yaml:"provides"`
	Files        []FileSpec                 `yaml:"files"`
	Systemd      SystemdSpec                `yaml:"systemd"`
	ConfigSchema map[string]ConfigFieldSpec `yaml:"config_schema"`
}

type FileSpec struct {
	Path string `yaml:"path"`
	Mode string `yaml:"mode"`
}

type SystemdSpec struct {
	Enable []string `yaml:"enable"`
	Mask   []string `yaml:"mask"`
}

type ConfigFieldSpec struct {
	Type        string `yaml:"type"`
	Default     any    `yaml:"default"`
	Description string `yaml:"description"`
}

func loadLayerSpec(t *testing.T, layerPath string) *LayerSpec {
	t.Helper()
	data, err := os.ReadFile(layerPath)
	if err != nil {
		t.Fatalf("failed to read layer.yaml: %v", err)
	}

	var spec LayerSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("failed to parse layer.yaml: %v", err)
	}
	return &spec
}

func layersDir(t *testing.T) string {
	// Find project root by looking for go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "layers")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("layers directory not found - skipping integration test")
		}
		dir = parent
	}
}

func TestPythonRuntimeLayer_Spec(t *testing.T) {
	layers := layersDir(t)
	specPath := filepath.Join(layers, "python-runtime", "layer.yaml")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("python-runtime layer.yaml not found")
	}

	spec := loadLayerSpec(t, specPath)

	// Validate layer spec
	if spec.Name != "python-runtime" {
		t.Errorf("Name = %q, want %q", spec.Name, "python-runtime")
	}
	if spec.Version == "" {
		t.Error("Version should not be empty")
	}
	if spec.Type != "runtime" {
		t.Errorf("Type = %q, want %q", spec.Type, "runtime")
	}
	if len(spec.Dependencies) == 0 {
		t.Error("python-runtime should depend on base-os")
	}
	if !contains(spec.Dependencies, "base-os>=1.0.0") {
		t.Errorf("Dependencies = %v, should contain 'base-os>=1.0.0'", spec.Dependencies)
	}

	// Check provides
	expectedProvides := []string{"python", "pip", "venv"}
	for _, p := range expectedProvides {
		if !contains(spec.Provides, p) {
			t.Errorf("Provides = %v, should contain %q", spec.Provides, p)
		}
	}

	// Check SHA256 is present if layer was built
	if spec.SHA256 != "" {
		if len(spec.SHA256) != 64 {
			t.Errorf("SHA256 length = %d, want 64", len(spec.SHA256))
		}
		if spec.SizeMB <= 0 {
			t.Errorf("SizeMB = %d, should be positive", spec.SizeMB)
		}
	}
}

func TestNodeRuntimeLayer_Spec(t *testing.T) {
	layers := layersDir(t)
	specPath := filepath.Join(layers, "node-runtime", "layer.yaml")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("node-runtime layer.yaml not found")
	}

	spec := loadLayerSpec(t, specPath)

	// Validate layer spec
	if spec.Name != "node-runtime" {
		t.Errorf("Name = %q, want %q", spec.Name, "node-runtime")
	}
	if spec.Version == "" {
		t.Error("Version should not be empty")
	}
	if spec.Type != "runtime" {
		t.Errorf("Type = %q, want %q", spec.Type, "runtime")
	}
	if len(spec.Dependencies) == 0 {
		t.Error("node-runtime should depend on base-os")
	}

	// Check provides
	expectedProvides := []string{"nodejs", "npm", "pnpm", "bun"}
	for _, p := range expectedProvides {
		if !contains(spec.Provides, p) {
			t.Errorf("Provides = %v, should contain %q", spec.Provides, p)
		}
	}

	// Check SHA256 is present if layer was built
	if spec.SHA256 != "" {
		if len(spec.SHA256) != 64 {
			t.Errorf("SHA256 length = %d, want 64", len(spec.SHA256))
		}
		if spec.SizeMB <= 0 {
			t.Errorf("SizeMB = %d, should be positive", spec.SizeMB)
		}
	}
}

func TestGoRuntimeLayer_Spec(t *testing.T) {
	layers := layersDir(t)
	specPath := filepath.Join(layers, "go-runtime", "layer.yaml")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("go-runtime layer.yaml not found")
	}

	spec := loadLayerSpec(t, specPath)

	// Validate layer spec
	if spec.Name != "go-runtime" {
		t.Errorf("Name = %q, want %q", spec.Name, "go-runtime")
	}
	if spec.Type != "runtime" {
		t.Errorf("Type = %q, want %q", spec.Type, "runtime")
	}

	// Check provides
	expectedProvides := []string{"go-runtime", "cgo-support", "tls-certificates"}
	for _, p := range expectedProvides {
		if !contains(spec.Provides, p) {
			t.Errorf("Provides = %v, should contain %q", spec.Provides, p)
		}
	}

	// Check SHA256 is present if layer was built
	if spec.SHA256 != "" {
		if len(spec.SHA256) != 64 {
			t.Errorf("SHA256 length = %d, want 64", len(spec.SHA256))
		}
		if spec.SizeMB <= 0 {
			t.Errorf("SizeMB = %d, should be positive", spec.SizeMB)
		}
	}
}

func TestPythonRuntimeLayer_Rootfs(t *testing.T) {
	layers := layersDir(t)
	rootfs := filepath.Join(layers, "python-runtime", "rootfs")
	if _, err := os.Stat(rootfs); os.IsNotExist(err) {
		t.Skip("python-runtime rootfs not built")
	}

	// Check essential files exist
	essentialFiles := []string{
		"usr/bin/python3.12",
		"usr/bin/pip3",
		"etc/nanofuse/layers/python-runtime.yaml",
	}

	for _, f := range essentialFiles {
		path := filepath.Join(rootfs, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("essential file missing: %s", f)
		}
	}
}

func TestNodeRuntimeLayer_Rootfs(t *testing.T) {
	layers := layersDir(t)
	rootfs := filepath.Join(layers, "node-runtime", "rootfs")
	if _, err := os.Stat(rootfs); os.IsNotExist(err) {
		t.Skip("node-runtime rootfs not built")
	}

	// Check essential files exist
	essentialFiles := []string{
		"usr/bin/node",
		"usr/local/bin/bun",
		"etc/nanofuse/layers/node-runtime.yaml",
	}

	for _, f := range essentialFiles {
		path := filepath.Join(rootfs, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("essential file missing: %s", f)
		}
	}
}

func TestGoRuntimeLayer_Rootfs(t *testing.T) {
	layers := layersDir(t)
	rootfs := filepath.Join(layers, "go-runtime", "rootfs")
	if _, err := os.Stat(rootfs); os.IsNotExist(err) {
		t.Skip("go-runtime rootfs not built")
	}

	// Check essential files/directories exist
	essentialPaths := []string{
		"etc/ssl/certs",
		"usr/share/zoneinfo",
		"etc/nanofuse/layers/go-runtime.yaml",
	}

	for _, p := range essentialPaths {
		path := filepath.Join(rootfs, p)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("essential path missing: %s", p)
		}
	}
}

func TestRuntimeLayerTarballs(t *testing.T) {
	layers := layersDir(t)

	runtimeLayers := []string{"python-runtime", "node-runtime", "go-runtime"}

	for _, layer := range runtimeLayers {
		t.Run(layer, func(t *testing.T) {
			tarball := filepath.Join(layers, layer, layer+".tar.gz")
			if _, err := os.Stat(tarball); os.IsNotExist(err) {
				t.Skipf("%s tarball not built", layer)
			}

			// Check tarball size is reasonable
			info, err := os.Stat(tarball)
			if err != nil {
				t.Fatalf("failed to stat tarball: %v", err)
			}
			if info.Size() == 0 {
				t.Error("tarball is empty")
			}
			if info.Size() < 1024 {
				t.Errorf("tarball is suspiciously small: %d bytes", info.Size())
			}

			// Verify corresponding layer.yaml has SHA256
			specPath := filepath.Join(layers, layer, "layer.yaml")
			spec := loadLayerSpec(t, specPath)
			if spec.SHA256 == "" {
				t.Error("layer.yaml should have SHA256 after tarball creation")
			}
		})
	}
}

func TestAllLayersHaveDockerfile(t *testing.T) {
	layers := layersDir(t)

	// These layers should have Dockerfiles
	layersWithDockerfile := []string{"python-runtime", "node-runtime", "go-runtime"}

	for _, layer := range layersWithDockerfile {
		t.Run(layer, func(t *testing.T) {
			dockerPath := filepath.Join(layers, layer, "Dockerfile")
			if _, err := os.Stat(dockerPath); os.IsNotExist(err) {
				t.Errorf("Dockerfile missing for %s", layer)
			}
		})
	}
}

func TestLayerNamingConvention(t *testing.T) {
	layers := layersDir(t)
	entries, err := os.ReadDir(layers)
	if err != nil {
		t.Skip("layers directory not accessible")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			// Check naming convention: lowercase, hyphens only
			if strings.ToLower(name) != name {
				t.Errorf("layer name %q should be lowercase", name)
			}
			if strings.Contains(name, "_") {
				t.Errorf("layer name %q should use hyphens, not underscores", name)
			}
			if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
				t.Errorf("layer name %q should not start or end with hyphen", name)
			}
		})
	}
}

func TestLayerYAMLConsistency(t *testing.T) {
	layers := layersDir(t)
	entries, err := os.ReadDir(layers)
	if err != nil {
		t.Skip("layers directory not accessible")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		layerPath := filepath.Join(layers, entry.Name(), "layer.yaml")
		if _, err := os.Stat(layerPath); os.IsNotExist(err) {
			continue // Skip if no layer.yaml
		}

		t.Run(entry.Name(), func(t *testing.T) {
			spec := loadLayerSpec(t, layerPath)

			// Name in yaml should match directory name
			if spec.Name != entry.Name() {
				t.Errorf("layer.yaml name %q does not match directory name %q", spec.Name, entry.Name())
			}

			// Type should be valid
			validTypes := map[string]bool{"base": true, "runtime": true, "feature": true, "application": true}
			if !validTypes[spec.Type] {
				t.Errorf("invalid layer type: %q", spec.Type)
			}

			// Version should be set
			if spec.Version == "" {
				t.Error("version should not be empty")
			}
		})
	}
}

// helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
