package layerbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test fixtures
const validManifestYAML = `version: "1.0"
name: "test-image"
description: "Test microVM image"

kernel:
  version: "6.1.90"
  source: "local://build/vmlinux"
  sha256: "abc123def456"
  cmdline: "console=ttyS0 root=/dev/vda1 rw"

layers:
  - name: "base-os"
    type: "base"
    source: "docker://nanofuse-base:latest"
    sha256: "sha256:def456abc789"
    required: true

  - name: "python-runtime"
    type: "runtime"
    source: "registry://ghcr.io/nanofuse/layers/python:3.12"
    sha256: "sha256:789abc123"
    condition: "${INCLUDE_PYTHON:-false}"

  - name: "recording-agent"
    type: "feature"
    source: "local://layers/recording-agent"
    condition: "${INCLUDE_RECORDING:-false}"
    config:
      vsock_port: 52
      buffer_size_mb: 16

output:
  format: "ext4"
  size_mb: 2048
  compression: "none"
`

const minimalManifestYAML = `version: "1.0"
name: "minimal"

kernel:
  version: "6.1.90"
  source: "local://vmlinux"
  cmdline: "console=ttyS0"

layers:
  - name: "base"
    type: "base"
    source: "docker://base:latest"
`

func TestParseManifest_ValidComplete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.manifest.yaml")
	if err := os.WriteFile(path, []byte(validManifestYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	m, err := ParseManifest(path)
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}

	// Verify basic fields
	if m.Version != "1.0" {
		t.Errorf("Version = %q, want %q", m.Version, "1.0")
	}
	if m.Name != "test-image" {
		t.Errorf("Name = %q, want %q", m.Name, "test-image")
	}
	if m.Description != "Test microVM image" {
		t.Errorf("Description = %q, want %q", m.Description, "Test microVM image")
	}

	// Verify kernel
	if m.Kernel.Version != "6.1.90" {
		t.Errorf("Kernel.Version = %q, want %q", m.Kernel.Version, "6.1.90")
	}
	if m.Kernel.Source != "local://build/vmlinux" {
		t.Errorf("Kernel.Source = %q, want %q", m.Kernel.Source, "local://build/vmlinux")
	}

	// Verify layers
	if len(m.Layers) != 3 {
		t.Fatalf("len(Layers) = %d, want 3", len(m.Layers))
	}

	// Check base layer
	if m.Layers[0].Name != "base-os" {
		t.Errorf("Layers[0].Name = %q, want %q", m.Layers[0].Name, "base-os")
	}
	if m.Layers[0].Type != LayerTypeBase {
		t.Errorf("Layers[0].Type = %q, want %q", m.Layers[0].Type, LayerTypeBase)
	}
	if !m.Layers[0].Required {
		t.Error("Layers[0].Required = false, want true")
	}

	// Check conditional layer
	if m.Layers[1].Condition != "${INCLUDE_PYTHON:-false}" {
		t.Errorf("Layers[1].Condition = %q, want %q", m.Layers[1].Condition, "${INCLUDE_PYTHON:-false}")
	}

	// Check layer config
	if m.Layers[2].Config["vsock_port"] != 52 {
		t.Errorf("Layers[2].Config[vsock_port] = %v, want 52", m.Layers[2].Config["vsock_port"])
	}

	// Verify output
	if m.Output.Format != "ext4" {
		t.Errorf("Output.Format = %q, want %q", m.Output.Format, "ext4")
	}
	if m.Output.SizeMB != 2048 {
		t.Errorf("Output.SizeMB = %d, want 2048", m.Output.SizeMB)
	}
}

func TestParseManifest_MinimalValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "minimal.yaml")
	if err := os.WriteFile(path, []byte(minimalManifestYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	m, err := ParseManifest(path)
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}

	if m.Name != "minimal" {
		t.Errorf("Name = %q, want %q", m.Name, "minimal")
	}
	if len(m.Layers) != 1 {
		t.Errorf("len(Layers) = %d, want 1", len(m.Layers))
	}
}

func TestParseManifest_FileNotFound(t *testing.T) {
	_, err := ParseManifest("/nonexistent/path/manifest.yaml")
	if err == nil {
		t.Error("ParseManifest should fail for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read manifest") {
		t.Errorf("error should mention 'failed to read manifest', got: %v", err)
	}
}

func TestParseManifest_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(path, []byte("this: is: invalid: yaml: [}"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParseManifest(path)
	if err == nil {
		t.Error("ParseManifest should fail for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse manifest") {
		t.Errorf("error should mention 'failed to parse manifest', got: %v", err)
	}
}

func TestValidateManifest_MissingVersion(t *testing.T) {
	m := &ImageManifest{
		Name: "test",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{{Name: "base", Type: LayerTypeBase, Source: "docker://base:latest"}},
	}

	err := ValidateManifest(m)
	if err == nil {
		t.Error("ValidateManifest should fail for missing version")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("error should mention 'version', got: %v", err)
	}
}

func TestValidateManifest_InvalidVersion(t *testing.T) {
	m := &ImageManifest{
		Version: "2.0",
		Name:    "test",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{{Name: "base", Type: LayerTypeBase, Source: "docker://base:latest"}},
	}

	err := ValidateManifest(m)
	if err == nil {
		t.Error("ValidateManifest should fail for invalid version")
	}
	if !strings.Contains(err.Error(), "1.0") {
		t.Errorf("error should mention '1.0', got: %v", err)
	}
}

func TestValidateManifest_InvalidName(t *testing.T) {
	tests := []struct {
		name    string
		imgName string
		wantErr bool
	}{
		{"valid lowercase", "test-image", false},
		{"valid with numbers", "test123", false},
		{"valid single char", "x", false},
		{"empty", "", true},
		{"starts with hyphen", "-test", true},
		{"ends with hyphen", "test-", true},
		{"uppercase", "Test-Image", true},
		{"underscore", "test_image", true},
		{"space", "test image", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ImageManifest{
				Version: "1.0",
				Name:    tt.imgName,
				Kernel: KernelConfig{
					Version: "6.1",
					Source:  "local://vmlinux",
					Cmdline: "console=ttyS0",
				},
				Layers: []LayerReference{{Name: "base", Type: LayerTypeBase, Source: "docker://base:latest"}},
			}

			err := ValidateManifest(m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateManifest_MissingKernel(t *testing.T) {
	m := &ImageManifest{
		Version: "1.0",
		Name:    "test",
		Layers:  []LayerReference{{Name: "base", Type: LayerTypeBase, Source: "docker://base:latest"}},
	}

	err := ValidateManifest(m)
	if err == nil {
		t.Error("ValidateManifest should fail for missing kernel")
	}
	if !strings.Contains(err.Error(), "kernel") {
		t.Errorf("error should mention 'kernel', got: %v", err)
	}
}

func TestValidateManifest_NoLayers(t *testing.T) {
	m := &ImageManifest{
		Version: "1.0",
		Name:    "test",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{},
	}

	err := ValidateManifest(m)
	if err == nil {
		t.Error("ValidateManifest should fail for no layers")
	}
	if !strings.Contains(err.Error(), "layer") {
		t.Errorf("error should mention 'layer', got: %v", err)
	}
}

func TestValidateManifest_InvalidLayerType(t *testing.T) {
	m := &ImageManifest{
		Version: "1.0",
		Name:    "test",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{{Name: "bad", Type: "invalid", Source: "docker://base:latest"}},
	}

	err := ValidateManifest(m)
	if err == nil {
		t.Error("ValidateManifest should fail for invalid layer type")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("error should mention 'type', got: %v", err)
	}
}

func TestValidateManifest_MissingSHA256ForRemote(t *testing.T) {
	m := &ImageManifest{
		Version: "1.0",
		Name:    "test",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{
			{Name: "remote", Type: LayerTypeBase, Source: "registry://ghcr.io/test:latest"},
		},
	}

	err := ValidateManifest(m)
	if err == nil {
		t.Error("ValidateManifest should fail for remote source without sha256")
	}
	if !strings.Contains(err.Error(), "sha256") {
		t.Errorf("error should mention 'sha256', got: %v", err)
	}
}

func TestValidateManifest_LocalSourceWithoutSHA256_OK(t *testing.T) {
	m := &ImageManifest{
		Version: "1.0",
		Name:    "test",
		Kernel: KernelConfig{
			Version: "6.1",
			Source:  "local://vmlinux",
			Cmdline: "console=ttyS0",
		},
		Layers: []LayerReference{
			{Name: "local", Type: LayerTypeBase, Source: "local://layers/base"},
		},
	}

	err := ValidateManifest(m)
	if err != nil {
		t.Errorf("ValidateManifest should pass for local source without sha256: %v", err)
	}
}

func TestEvaluateConditions_NoConditions(t *testing.T) {
	m := &ImageManifest{
		Layers: []LayerReference{
			{Name: "base", Required: true},
			{Name: "feature"},
		},
	}

	active := EvaluateConditions(m, nil)
	if len(active) != 2 {
		t.Errorf("len(active) = %d, want 2", len(active))
	}
}

func TestEvaluateConditions_RequiredAlwaysIncluded(t *testing.T) {
	m := &ImageManifest{
		Layers: []LayerReference{
			{Name: "base", Required: true, Condition: "${NEVER:-false}"},
		},
	}

	env := map[string]string{"NEVER": "false"}
	active := EvaluateConditions(m, env)
	if len(active) != 1 {
		t.Errorf("len(active) = %d, want 1 (required layers always included)", len(active))
	}
}

func TestEvaluateConditions_TrueCondition(t *testing.T) {
	m := &ImageManifest{
		Layers: []LayerReference{
			{Name: "base"},
			{Name: "python", Condition: "${INCLUDE_PYTHON:-false}"},
		},
	}

	env := map[string]string{"INCLUDE_PYTHON": "true"}
	active := EvaluateConditions(m, env)
	if len(active) != 2 {
		t.Errorf("len(active) = %d, want 2", len(active))
	}
}

func TestEvaluateConditions_FalseCondition(t *testing.T) {
	m := &ImageManifest{
		Layers: []LayerReference{
			{Name: "base"},
			{Name: "python", Condition: "${INCLUDE_PYTHON:-false}"},
		},
	}

	env := map[string]string{"INCLUDE_PYTHON": "false"}
	active := EvaluateConditions(m, env)
	if len(active) != 1 {
		t.Errorf("len(active) = %d, want 1", len(active))
	}
	if active[0].Name != "base" {
		t.Errorf("active[0].Name = %q, want 'base'", active[0].Name)
	}
}

func TestEvaluateConditions_DefaultValue(t *testing.T) {
	m := &ImageManifest{
		Layers: []LayerReference{
			{Name: "base"},
			{Name: "debug", Condition: "${DEBUG:-true}"},
		},
	}

	// No DEBUG in env, should use default "true"
	active := EvaluateConditions(m, map[string]string{})
	if len(active) != 2 {
		t.Errorf("len(active) = %d, want 2 (default is true)", len(active))
	}
}

func TestEvaluateConditions_DefaultValueFalse(t *testing.T) {
	m := &ImageManifest{
		Layers: []LayerReference{
			{Name: "base"},
			{Name: "optional", Condition: "${OPT:-false}"},
		},
	}

	// No OPT in env, should use default "false"
	active := EvaluateConditions(m, map[string]string{})
	if len(active) != 1 {
		t.Errorf("len(active) = %d, want 1 (default is false)", len(active))
	}
}

func TestResolveDependencies_NoDependencies(t *testing.T) {
	layers := []LayerReference{
		{Name: "app"},
		{Name: "runtime"},
		{Name: "base"},
	}

	resolved, err := ResolveDependencies(layers)
	if err != nil {
		t.Fatalf("ResolveDependencies failed: %v", err)
	}

	// Order should be preserved when no dependencies
	if len(resolved) != 3 {
		t.Errorf("len(resolved) = %d, want 3", len(resolved))
	}
}

func TestResolveDependencies_SimpleDependency(t *testing.T) {
	layers := []LayerReference{
		{Name: "app", Dependencies: []string{"runtime"}},
		{Name: "runtime", Dependencies: []string{"base"}},
		{Name: "base"},
	}

	resolved, err := ResolveDependencies(layers)
	if err != nil {
		t.Fatalf("ResolveDependencies failed: %v", err)
	}

	// base must come before runtime, runtime before app
	baseIdx, runtimeIdx, appIdx := -1, -1, -1
	for i, l := range resolved {
		switch l.Name {
		case "base":
			baseIdx = i
		case "runtime":
			runtimeIdx = i
		case "app":
			appIdx = i
		}
	}

	if baseIdx > runtimeIdx {
		t.Errorf("base (%d) should come before runtime (%d)", baseIdx, runtimeIdx)
	}
	if runtimeIdx > appIdx {
		t.Errorf("runtime (%d) should come before app (%d)", runtimeIdx, appIdx)
	}
}

func TestResolveDependencies_CircularDependency(t *testing.T) {
	layers := []LayerReference{
		{Name: "a", Dependencies: []string{"b"}},
		{Name: "b", Dependencies: []string{"c"}},
		{Name: "c", Dependencies: []string{"a"}},
	}

	_, err := ResolveDependencies(layers)
	if err == nil {
		t.Error("ResolveDependencies should fail for circular dependency")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error should mention 'circular', got: %v", err)
	}
}

func TestResolveDependencies_MissingDependency(t *testing.T) {
	layers := []LayerReference{
		{Name: "app", Dependencies: []string{"nonexistent"}},
		{Name: "base"},
	}

	_, err := ResolveDependencies(layers)
	if err == nil {
		t.Error("ResolveDependencies should fail for missing dependency")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention 'nonexistent', got: %v", err)
	}
}

func TestResolveDependencies_SelfDependency(t *testing.T) {
	layers := []LayerReference{
		{Name: "self", Dependencies: []string{"self"}},
	}

	_, err := ResolveDependencies(layers)
	if err == nil {
		t.Error("ResolveDependencies should fail for self dependency")
	}
}

func TestParseSourceType(t *testing.T) {
	tests := []struct {
		source   string
		wantType SourceType
		wantOK   bool
	}{
		{"docker://image:tag", SourceTypeDocker, true},
		{"registry://ghcr.io/org/repo:tag", SourceTypeRegistry, true},
		{"local://./path/to/layer", SourceTypeLocal, true},
		{"url://https://example.com/layer.tar.gz", SourceTypeURL, true},
		{"invalid://something", "", false},
		{"no-scheme", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			got, ok := ParseSourceType(tt.source)
			if ok != tt.wantOK {
				t.Errorf("ParseSourceType(%q) ok = %v, want %v", tt.source, ok, tt.wantOK)
			}
			if ok && got != tt.wantType {
				t.Errorf("ParseSourceType(%q) = %q, want %q", tt.source, got, tt.wantType)
			}
		})
	}
}

func TestLayerType_Valid(t *testing.T) {
	tests := []struct {
		lt   LayerType
		want bool
	}{
		{LayerTypeBase, true},
		{LayerTypeRuntime, true},
		{LayerTypeFeature, true},
		{LayerTypeApplication, true},
		{LayerType("invalid"), false},
		{LayerType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.lt), func(t *testing.T) {
			if got := tt.lt.Valid(); got != tt.want {
				t.Errorf("LayerType(%q).Valid() = %v, want %v", tt.lt, got, tt.want)
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	var errs ValidationErrors

	// Empty errors
	if errs.HasErrors() {
		t.Error("Empty ValidationErrors should not have errors")
	}

	// Add errors
	errs.Add("field1", "message1")
	errs.AddWithLine("field2", "message2", 10)

	if !errs.HasErrors() {
		t.Error("ValidationErrors with entries should have errors")
	}

	if len(errs) != 2 {
		t.Errorf("len(errs) = %d, want 2", len(errs))
	}

	// Check error string format
	errStr := errs.Error()
	if !strings.Contains(errStr, "2 validation errors") {
		t.Errorf("Error() should mention count, got: %s", errStr)
	}

	// Single error case
	singleErr := ValidationErrors{{Field: "f", Message: "m"}}
	if singleErr.Error() != "f: m" {
		t.Errorf("Single error format wrong: %s", singleErr.Error())
	}

	// Error with line number
	lineErr := ValidationError{Field: "f", Message: "m", Line: 5}
	if !strings.Contains(lineErr.Error(), "line 5") {
		t.Errorf("Line error should include line number: %s", lineErr.Error())
	}
}

func TestDefaultOutputConfig(t *testing.T) {
	def := DefaultOutputConfig()

	if def.Format != "ext4" {
		t.Errorf("Format = %q, want 'ext4'", def.Format)
	}
	if def.SizeMB != 2048 {
		t.Errorf("SizeMB = %d, want 2048", def.SizeMB)
	}
	if def.Compression != CompressionNone {
		t.Errorf("Compression = %q, want 'none'", def.Compression)
	}
	if def.Path != "./build" {
		t.Errorf("Path = %q, want './build'", def.Path)
	}
}
