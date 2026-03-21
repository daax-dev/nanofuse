package layer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateLayer_ValidLayer(t *testing.T) {
	// Create a valid layer structure
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Scaffold a layer
	opts := ScaffoldOptions{
		Name:      "test-layer",
		Type:      LayerTypeFeature,
		OutputDir: tmpDir,
	}
	_, err := ScaffoldLayer(opts)
	if err != nil {
		t.Fatalf("ScaffoldLayer() error = %v", err)
	}

	// Validate it
	result := ValidateLayer(layerDir, ValidateOptions{})

	if !result.Valid {
		t.Errorf("ValidateLayer() Valid = false, want true")
		for _, issue := range result.Issues {
			t.Logf("Issue: [%s] %s: %s", issue.Severity, issue.Field, issue.Message)
		}
	}

	if result.Spec == nil {
		t.Error("ValidateLayer() should parse spec")
	}

	if result.Spec != nil && result.Spec.Name != "test-layer" {
		t.Errorf("Spec.Name = %q, want %q", result.Spec.Name, "test-layer")
	}
}

func TestValidateLayer_MissingDirectory(t *testing.T) {
	result := ValidateLayer("/nonexistent/path", ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for nonexistent directory")
	}

	errors := result.Errors()
	if len(errors) == 0 {
		t.Error("ValidateLayer() should have errors for nonexistent directory")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err.Message, "does not exist") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Error should mention directory does not exist")
	}
}

func TestValidateLayer_MissingLayerYAML(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create directory structure without layer.yaml
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layerDir, "hooks"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for missing layer.yaml")
	}

	errors := result.Errors()
	found := false
	for _, err := range errors {
		if strings.Contains(err.Message, "layer.yaml not found") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Error should mention layer.yaml not found")
	}
}

func TestValidateLayer_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layerDir, "hooks"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write invalid YAML
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	if err := os.WriteFile(yamlPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for invalid YAML")
	}

	errors := result.Errors()
	found := false
	for _, err := range errors {
		if strings.Contains(err.Message, "Invalid YAML") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Error should mention Invalid YAML")
	}
}

func TestValidateLayer_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layerDir, "hooks"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write YAML missing required fields
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	yamlContent := `description: "Test layer without required fields"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for missing required fields")
	}

	errors := result.Errors()
	// Should have errors for name, version, and type
	requiredFields := []string{"name", "version", "type"}
	for _, field := range requiredFields {
		found := false
		for _, err := range errors {
			if err.Field == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should have error for missing %s field", field)
		}
	}
}

func TestValidateLayer_InvalidLayerType(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layerDir, "hooks"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write YAML with invalid type
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "invalid-type"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for invalid layer type")
	}

	errors := result.Errors()
	found := false
	for _, err := range errors {
		if err.Field == "type" && strings.Contains(err.Message, "Invalid layer type") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have error for invalid layer type")
	}
}

func TestValidateLayer_MissingRootfs(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create directory without rootfs
	if err := os.MkdirAll(layerDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layerDir, "hooks"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write valid YAML
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for missing rootfs/")
	}

	errors := result.Errors()
	found := false
	for _, err := range errors {
		if strings.Contains(err.Message, "rootfs") && strings.Contains(err.Message, "not found") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have error for missing rootfs/ directory")
	}
}

func TestValidateLayer_NonExecutableHook(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	hooksDir := filepath.Join(layerDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write valid YAML
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	// Write non-executable hook
	hookPath := filepath.Join(hooksDir, "post-install.sh")
	hookContent := `#!/bin/bash
echo "test"
`
	if err := os.WriteFile(hookPath, []byte(hookContent), 0644); err != nil { // Note: not executable
		t.Fatalf("Failed to write hook: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for non-executable hook")
	}

	errors := result.Errors()
	found := false
	for _, err := range errors {
		if strings.Contains(err.Message, "not executable") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have error for non-executable hook")
	}
}

func TestValidateLayer_StrictMode_EmptyRootfs(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create valid but minimal layer structure
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layerDir, "hooks"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write valid YAML
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	// Validate without strict - should pass (empty rootfs is OK)
	result := ValidateLayer(layerDir, ValidateOptions{Strict: false})
	if !result.Valid {
		t.Error("ValidateLayer() without strict should pass for empty rootfs")
	}

	// Validate with strict - should warn about empty rootfs
	resultStrict := ValidateLayer(layerDir, ValidateOptions{Strict: true})
	// Still valid but should have warnings
	if !resultStrict.Valid {
		t.Error("ValidateLayer() with strict should still be valid but have warnings")
	}

	warnings := resultStrict.Warnings()
	found := false
	for _, warn := range warnings {
		if strings.Contains(warn.Message, "rootfs") && strings.Contains(warn.Message, "empty") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have warning for empty rootfs/ in strict mode")
	}
}

func TestValidateLayer_StrictMode_MissingShebang(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	hooksDir := filepath.Join(layerDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write valid YAML
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
description: "Test layer"
provides:
  - test
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	// Write executable hook without shebang
	hookPath := filepath.Join(hooksDir, "post-install.sh")
	hookContent := `echo "no shebang"
`
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		t.Fatalf("Failed to write hook: %v", err)
	}

	// Validate with strict mode
	result := ValidateLayer(layerDir, ValidateOptions{Strict: true})

	// Should be valid but have warning
	if !result.Valid {
		t.Error("ValidateLayer() should be valid despite missing shebang")
	}

	warnings := result.Warnings()
	found := false
	for _, warn := range warnings {
		if strings.Contains(warn.Message, "shebang") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have warning for missing shebang in strict mode")
	}
}

func TestValidateLayer_NameMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "different-name")

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layerDir, "hooks"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write YAML with different name
	yamlPath := filepath.Join(layerDir, "layer.yaml")
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	// Should be valid but have warning
	if !result.Valid {
		t.Error("ValidateLayer() should be valid despite name mismatch")
	}

	warnings := result.Warnings()
	found := false
	for _, warn := range warnings {
		if strings.Contains(warn.Message, "does not match directory") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have warning for name mismatch")
	}
}

func TestValidationResult_Errors(t *testing.T) {
	result := &ValidationResult{
		Issues: []ValidationIssue{
			{Severity: SeverityError, Message: "error 1"},
			{Severity: SeverityWarning, Message: "warning 1"},
			{Severity: SeverityError, Message: "error 2"},
			{Severity: SeverityInfo, Message: "info 1"},
		},
	}

	errors := result.Errors()
	if len(errors) != 2 {
		t.Errorf("Errors() returned %d errors, want 2", len(errors))
	}
}

func TestValidationResult_Warnings(t *testing.T) {
	result := &ValidationResult{
		Issues: []ValidationIssue{
			{Severity: SeverityError, Message: "error 1"},
			{Severity: SeverityWarning, Message: "warning 1"},
			{Severity: SeverityWarning, Message: "warning 2"},
			{Severity: SeverityInfo, Message: "info 1"},
		},
	}

	warnings := result.Warnings()
	if len(warnings) != 2 {
		t.Errorf("Warnings() returned %d warnings, want 2", len(warnings))
	}
}

func TestValidateLayer_ExistingLayers(t *testing.T) {
	// Test against real layers in the repository if they exist
	// Find project root
	dir, err := os.Getwd()
	if err != nil {
		t.Skip("Cannot get working directory")
	}

	// Walk up to find layers directory
	for {
		layersDir := filepath.Join(dir, "layers")
		if _, err := os.Stat(layersDir); err == nil {
			// Found layers directory
			entries, err := os.ReadDir(layersDir)
			if err != nil {
				t.Skip("Cannot read layers directory")
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				t.Run(entry.Name(), func(t *testing.T) {
					layerPath := filepath.Join(layersDir, entry.Name())
					result := ValidateLayer(layerPath, ValidateOptions{})

					// Log issues for debugging
					for _, issue := range result.Issues {
						t.Logf("[%s] %s: %s", issue.Severity, issue.Field, issue.Message)
					}

					// Existing layers should be valid (or at least have parseable specs)
					if result.Spec == nil {
						// Only fail if layer.yaml exists but couldn't be parsed
						yamlPath := filepath.Join(layerPath, "layer.yaml")
						if _, err := os.Stat(yamlPath); err == nil {
							t.Errorf("Layer %s has layer.yaml but spec could not be parsed", entry.Name())
						}
					}
				})
			}
			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("layers directory not found")
		}
		dir = parent
	}
}

func TestValidateLayer_EmptyDependency(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write YAML with empty dependency string
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
dependencies:
  - "valid-dep"
  - ""
  - "another-dep"
`
	if err := os.WriteFile(filepath.Join(layerDir, "layer.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{})

	if result.Valid {
		t.Error("ValidateLayer() should fail for empty dependency string")
	}

	errors := result.Errors()
	found := false
	for _, err := range errors {
		if strings.Contains(err.Message, "Empty dependency") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have error for empty dependency string")
	}
}

func TestValidateLayer_EmptyProvides(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write YAML with empty provides string (strict mode only)
	yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
provides:
  - "valid-capability"
  - ""
`
	if err := os.WriteFile(filepath.Join(layerDir, "layer.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write layer.yaml: %v", err)
	}

	result := ValidateLayer(layerDir, ValidateOptions{Strict: true})

	if result.Valid {
		t.Error("ValidateLayer() should fail for empty provides string in strict mode")
	}

	errors := result.Errors()
	found := false
	for _, err := range errors {
		if strings.Contains(err.Message, "Empty provides") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have error for empty provides string")
	}
}

func TestValidateLayer_InvalidSHA256(t *testing.T) {
	tmpDir := t.TempDir()
	layerDir := filepath.Join(tmpDir, "test-layer")

	if err := os.MkdirAll(filepath.Join(layerDir, "rootfs"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a fake tarball to trigger SHA256 validation
	if err := os.WriteFile(filepath.Join(layerDir, "test-layer.tar.gz"), []byte("fake"), 0644); err != nil {
		t.Fatalf("Failed to create tarball: %v", err)
	}

	tests := []struct {
		name    string
		sha256  string
		wantErr string
	}{
		{
			name:    "too short",
			sha256:  "abc123",
			wantErr: "Invalid SHA256 hash length",
		},
		{
			name:    "non-hex characters",
			sha256:  "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			wantErr: "non-hexadecimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlContent := `name: "test-layer"
version: "1.0.0"
type: "feature"
sha256: "` + tt.sha256 + `"
`
			if err := os.WriteFile(filepath.Join(layerDir, "layer.yaml"), []byte(yamlContent), 0644); err != nil {
				t.Fatalf("Failed to write layer.yaml: %v", err)
			}

			result := ValidateLayer(layerDir, ValidateOptions{Strict: true})

			if result.Valid {
				t.Error("ValidateLayer() should fail for invalid SHA256")
			}

			errors := result.Errors()
			found := false
			for _, err := range errors {
				if strings.Contains(err.Message, tt.wantErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Should have error containing %q", tt.wantErr)
				for _, err := range errors {
					t.Logf("Got error: %s", err.Message)
				}
			}
		})
	}
}
