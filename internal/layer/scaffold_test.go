package layer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldLayer_Basic(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	opts := ScaffoldOptions{
		Name:      "test-layer",
		Type:      LayerTypeFeature,
		OutputDir: tmpDir,
	}

	result, err := ScaffoldLayer(opts)
	if err != nil {
		t.Fatalf("ScaffoldLayer() error = %v", err)
	}

	// Verify layer directory was created
	if result.LayerDir == "" {
		t.Error("LayerDir should not be empty")
	}
	expectedDir := filepath.Join(tmpDir, "test-layer")
	if result.LayerDir != expectedDir {
		t.Errorf("LayerDir = %q, want %q", result.LayerDir, expectedDir)
	}

	// Verify directories were created
	expectedDirs := []string{
		"test-layer",
		"test-layer/rootfs",
		"test-layer/hooks",
		"test-layer/tests",
	}
	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Errorf("directory %s not created or not a directory", dir)
		}
	}

	// Verify files were created
	expectedFiles := []string{
		"test-layer/layer.yaml",
		"test-layer/hooks/post-install.sh",
		"test-layer/rootfs/.gitkeep",
		"test-layer/tests/.gitkeep",
	}
	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("file %s not created: %v", file, err)
		}
	}
}

func TestScaffoldLayer_LayerYAMLContent(t *testing.T) {
	tmpDir := t.TempDir()

	opts := ScaffoldOptions{
		Name:        "my-runtime",
		Type:        LayerTypeRuntime,
		OutputDir:   tmpDir,
		Description: "My custom runtime layer",
		Version:     "1.2.3",
	}

	_, err := ScaffoldLayer(opts)
	if err != nil {
		t.Fatalf("ScaffoldLayer() error = %v", err)
	}

	// Read and verify layer.yaml content
	yamlPath := filepath.Join(tmpDir, "my-runtime", "layer.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("Failed to read layer.yaml: %v", err)
	}

	content := string(data)

	// Check required fields
	checks := []struct {
		name     string
		contains string
	}{
		{"name field", `name: "my-runtime"`},
		{"version field", `version: "1.2.3"`},
		{"type field", `type: "runtime"`},
		{"description", "My custom runtime layer"},
		{"dependencies", "base-os>=1.0.0"},
	}

	for _, check := range checks {
		if !strings.Contains(content, check.contains) {
			t.Errorf("layer.yaml should contain %s (%q)", check.name, check.contains)
		}
	}
}

func TestScaffoldLayer_PostInstallHookExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	opts := ScaffoldOptions{
		Name:      "test-layer",
		Type:      LayerTypeFeature,
		OutputDir: tmpDir,
	}

	_, err := ScaffoldLayer(opts)
	if err != nil {
		t.Fatalf("ScaffoldLayer() error = %v", err)
	}

	// Check hook is executable
	hookPath := filepath.Join(tmpDir, "test-layer", "hooks", "post-install.sh")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("Failed to stat post-install.sh: %v", err)
	}

	mode := info.Mode()
	if mode&0111 == 0 {
		t.Error("post-install.sh should be executable")
	}

	// Check shebang
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read post-install.sh: %v", err)
	}
	if !strings.HasPrefix(string(data), "#!/bin/bash") {
		t.Error("post-install.sh should start with #!/bin/bash shebang")
	}
}

func TestScaffoldLayer_AllTypes(t *testing.T) {
	for _, layerType := range ValidLayerTypes {
		t.Run(string(layerType), func(t *testing.T) {
			tmpDir := t.TempDir()

			opts := ScaffoldOptions{
				Name:      "test-" + string(layerType),
				Type:      layerType,
				OutputDir: tmpDir,
			}

			result, err := ScaffoldLayer(opts)
			if err != nil {
				t.Fatalf("ScaffoldLayer() error = %v", err)
			}

			if result.LayerDir == "" {
				t.Error("LayerDir should not be empty")
			}
		})
	}
}

func TestScaffoldLayer_DefaultDependencies(t *testing.T) {
	tests := []struct {
		layerType LayerType
		expectDep bool
	}{
		{LayerTypeBase, false},
		{LayerTypeRuntime, true},
		{LayerTypeFeature, true},
		{LayerTypeApplication, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.layerType), func(t *testing.T) {
			deps := DefaultDependencies(tt.layerType)
			hasDeps := len(deps) > 0

			if hasDeps != tt.expectDep {
				t.Errorf("DefaultDependencies(%s) has deps = %v, want %v", tt.layerType, hasDeps, tt.expectDep)
			}

			if tt.expectDep && len(deps) > 0 {
				found := false
				for _, d := range deps {
					if strings.Contains(d, "base-os") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Non-base layer should depend on base-os")
				}
			}
		})
	}
}

func TestScaffoldLayer_ValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		opts    ScaffoldOptions
		wantErr string
	}{
		{
			name:    "empty name",
			opts:    ScaffoldOptions{Name: "", Type: LayerTypeFeature, OutputDir: tmpDir},
			wantErr: "layer name is required",
		},
		{
			name:    "empty type",
			opts:    ScaffoldOptions{Name: "test", Type: "", OutputDir: tmpDir},
			wantErr: "layer type is required",
		},
		{
			name:    "invalid type",
			opts:    ScaffoldOptions{Name: "test", Type: "invalid", OutputDir: tmpDir},
			wantErr: "invalid layer type",
		},
		{
			name:    "uppercase name",
			opts:    ScaffoldOptions{Name: "TestLayer", Type: LayerTypeFeature, OutputDir: tmpDir},
			wantErr: "must be lowercase",
		},
		{
			name:    "underscore in name",
			opts:    ScaffoldOptions{Name: "test_layer", Type: LayerTypeFeature, OutputDir: tmpDir},
			wantErr: "must use hyphens",
		},
		{
			name:    "leading hyphen",
			opts:    ScaffoldOptions{Name: "-test", Type: LayerTypeFeature, OutputDir: tmpDir},
			wantErr: "must not start or end with hyphen",
		},
		{
			name:    "trailing hyphen",
			opts:    ScaffoldOptions{Name: "test-", Type: LayerTypeFeature, OutputDir: tmpDir},
			wantErr: "must not start or end with hyphen",
		},
		{
			name:    "invalid character",
			opts:    ScaffoldOptions{Name: "test.layer", Type: LayerTypeFeature, OutputDir: tmpDir},
			wantErr: "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ScaffoldLayer(tt.opts)
			if err == nil {
				t.Errorf("ScaffoldLayer() should return error")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ScaffoldLayer() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestScaffoldLayer_ExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create layer first time
	opts := ScaffoldOptions{
		Name:      "test-layer",
		Type:      LayerTypeFeature,
		OutputDir: tmpDir,
	}

	_, err := ScaffoldLayer(opts)
	if err != nil {
		t.Fatalf("First ScaffoldLayer() error = %v", err)
	}

	// Try to create again without force
	_, err = ScaffoldLayer(opts)
	if err == nil {
		t.Error("ScaffoldLayer() should fail when directory exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}

	// Create with force
	opts.Force = true
	_, err = ScaffoldLayer(opts)
	if err != nil {
		t.Errorf("ScaffoldLayer() with force should succeed: %v", err)
	}
}

func TestIsValidLayerType(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"base", true},
		{"runtime", true},
		{"feature", true},
		{"application", true},
		{"invalid", false},
		{"Base", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsValidLayerType(tt.input); got != tt.want {
				t.Errorf("IsValidLayerType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateLayerName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-name", false},
		{"valid123", false},
		{"a-b-c", false},
		{"a1b2c3", false},
		{"", true},
		{"UPPERCASE", true},
		{"under_score", true},
		{"-leading", true},
		{"trailing-", true},
		{"special.char", true},
		{"space name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLayerName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLayerName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestDefaultProvides(t *testing.T) {
	tests := []struct {
		layerType LayerType
		name      string
		want      []string
	}{
		{LayerTypeBase, "base-os", []string{"init", "systemd", "networking"}},
		{LayerTypeRuntime, "python-runtime", []string{"python"}},
		{LayerTypeRuntime, "nodejs-runtime", []string{"nodejs"}},
		{LayerTypeFeature, "debug-tools", []string{"debug-tools"}},
		{LayerTypeApplication, "my-app", []string{"my-app"}},
		{"unknown", "custom", []string{"custom"}}, // default case
	}

	for _, tt := range tests {
		t.Run(string(tt.layerType)+"/"+tt.name, func(t *testing.T) {
			got := DefaultProvides(tt.layerType, tt.name)
			if len(got) != len(tt.want) {
				t.Errorf("DefaultProvides(%v, %q) = %v, want %v", tt.layerType, tt.name, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("DefaultProvides(%v, %q)[%d] = %q, want %q", tt.layerType, tt.name, i, got[i], tt.want[i])
				}
			}
		})
	}
}
