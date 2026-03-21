// Package layer provides functionality for creating and validating NanoFuse layer definitions.
//
// Layers are the building blocks of NanoFuse microVM images. This package provides:
//   - Scaffolding: Generate new layer directory structures with templates
//   - Validation: Verify layer configurations and directory structures
package layer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// LayerType represents the type of layer being created.
type LayerType string

const (
	// LayerTypeBase is the foundation layer (e.g., base-os).
	LayerTypeBase LayerType = "base"
	// LayerTypeRuntime is a language/runtime layer (e.g., python-runtime).
	LayerTypeRuntime LayerType = "runtime"
	// LayerTypeFeature is a feature layer (e.g., agent-tools).
	LayerTypeFeature LayerType = "feature"
	// LayerTypeApplication is an application layer.
	LayerTypeApplication LayerType = "application"
)

// ValidLayerTypes contains all valid layer type values.
var ValidLayerTypes = []LayerType{
	LayerTypeBase,
	LayerTypeRuntime,
	LayerTypeFeature,
	LayerTypeApplication,
}

// IsValidLayerType checks if a string is a valid layer type.
func IsValidLayerType(t string) bool {
	for _, valid := range ValidLayerTypes {
		if LayerType(t) == valid {
			return true
		}
	}
	return false
}

// ScaffoldOptions contains options for layer scaffolding.
type ScaffoldOptions struct {
	// Name is the layer name (required).
	Name string
	// Type is the layer type (required).
	Type LayerType
	// OutputDir is the directory to create the layer in.
	// If empty, uses current working directory.
	OutputDir string
	// Description is an optional description for the layer.
	Description string
	// Version is the initial version (default: "0.1.0").
	Version string
	// Dependencies is a list of layer dependencies.
	Dependencies []string
	// Provides is a list of capabilities this layer provides.
	Provides []string
	// Force overwrites existing files if true.
	Force bool
}

// ScaffoldResult contains the results of a scaffolding operation.
type ScaffoldResult struct {
	// LayerDir is the absolute path to the created layer directory.
	LayerDir string
	// CreatedFiles is a list of files created.
	CreatedFiles []string
	// CreatedDirs is a list of directories created.
	CreatedDirs []string
}

// DefaultDependencies returns default dependencies for a layer type.
func DefaultDependencies(t LayerType) []string {
	switch t {
	case LayerTypeBase:
		return nil
	case LayerTypeRuntime, LayerTypeFeature, LayerTypeApplication:
		return []string{"base-os>=1.0.0"}
	default:
		return nil
	}
}

// DefaultProvides returns default provides for a layer type and name.
func DefaultProvides(t LayerType, name string) []string {
	switch t {
	case LayerTypeBase:
		return []string{"init", "systemd", "networking"}
	case LayerTypeRuntime:
		// Extract runtime name from layer name (e.g., "python-runtime" -> "python")
		runtime := strings.TrimSuffix(name, "-runtime")
		return []string{runtime}
	case LayerTypeFeature:
		return []string{name}
	case LayerTypeApplication:
		return []string{name}
	default:
		return []string{name}
	}
}

// ScaffoldLayer creates a new layer directory structure with template files.
func ScaffoldLayer(opts ScaffoldOptions) (*ScaffoldResult, error) {
	// Validate inputs
	if opts.Name == "" {
		return nil, fmt.Errorf("layer name is required")
	}
	if err := validateLayerName(opts.Name); err != nil {
		return nil, err
	}
	if opts.Type == "" {
		return nil, fmt.Errorf("layer type is required")
	}
	if !IsValidLayerType(string(opts.Type)) {
		return nil, fmt.Errorf("invalid layer type %q: must be one of %v", opts.Type, ValidLayerTypes)
	}

	// Set defaults
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}
	if opts.Dependencies == nil {
		opts.Dependencies = DefaultDependencies(opts.Type)
	}
	if opts.Provides == nil {
		opts.Provides = DefaultProvides(opts.Type, opts.Name)
	}
	if opts.Description == "" {
		caser := cases.Title(language.English)
		opts.Description = fmt.Sprintf("%s layer for NanoFuse microVMs", caser.String(string(opts.Type)))
	}

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
	}

	// Create layer directory path
	layerDir := filepath.Join(outputDir, opts.Name)

	// Check if layer already exists
	if _, err := os.Stat(layerDir); err == nil && !opts.Force {
		return nil, fmt.Errorf("layer directory already exists: %s (use --force to overwrite)", layerDir)
	}

	result := &ScaffoldResult{
		LayerDir: layerDir,
	}

	// Create directories
	dirs := []string{
		layerDir,
		filepath.Join(layerDir, "rootfs"),
		filepath.Join(layerDir, "hooks"),
		filepath.Join(layerDir, "tests"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
		result.CreatedDirs = append(result.CreatedDirs, dir)
	}

	// Create layer.yaml
	layerYAMLPath := filepath.Join(layerDir, "layer.yaml")
	if err := createLayerYAML(layerYAMLPath, opts); err != nil {
		return nil, fmt.Errorf("create layer.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, layerYAMLPath)

	// Create hooks/post-install.sh
	hookPath := filepath.Join(layerDir, "hooks", "post-install.sh")
	if err := createPostInstallHook(hookPath, opts); err != nil {
		return nil, fmt.Errorf("create post-install hook: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, hookPath)

	// Create .gitkeep in rootfs and tests
	for _, dir := range []string{"rootfs", "tests"} {
		gitkeepPath := filepath.Join(layerDir, dir, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil { //nolint:gosec // G306: 0644 is safe for empty .gitkeep placeholder files
			return nil, fmt.Errorf("create .gitkeep in %s: %w", dir, err)
		}
		result.CreatedFiles = append(result.CreatedFiles, gitkeepPath)
	}

	return result, nil
}

// validateLayerName validates a layer name follows naming conventions.
func validateLayerName(name string) error {
	if name == "" {
		return fmt.Errorf("layer name cannot be empty")
	}
	if strings.ToLower(name) != name {
		return fmt.Errorf("layer name %q must be lowercase", name)
	}
	if strings.Contains(name, "_") {
		return fmt.Errorf("layer name %q must use hyphens, not underscores", name)
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("layer name %q must not start or end with hyphen", name)
	}
	// Check for valid characters
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return fmt.Errorf("layer name %q contains invalid character %q: only lowercase letters, numbers, and hyphens allowed", name, c)
		}
	}
	return nil
}

// layerYAMLTemplate is the template for layer.yaml files.
const layerYAMLTemplate = `# {{.Name}} Layer
# {{.Description}}
#
# Type: {{.Type}}
# Version: {{.Version}}

name: "{{.Name}}"
version: "{{.Version}}"
sha256: ""
size_mb: 0
description: "{{.Description}}"
type: "{{.Type}}"

dependencies:{{if .Dependencies}}
{{- range .Dependencies}}
  - "{{.}}"
{{- end}}
{{else}}
  []
{{end}}
provides:{{if .Provides}}
{{- range .Provides}}
  - "{{.}}"
{{- end}}
{{else}}
  []
{{end}}

# List key files provided by this layer
# Example:
# - path: "/usr/bin/example"
#   mode: "0755"
files: []

systemd:
  enable: []
  mask: []

config_schema:
  # Define configuration options for this layer
  # example_option:
  #   type: string
  #   default: "value"
  #   description: "Example configuration option"
`

func createLayerYAML(path string, opts ScaffoldOptions) error {
	tmpl, err := template.New("layer.yaml").Parse(layerYAMLTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %s: %w", path, err)
	}

	if err := tmpl.Execute(f, opts); err != nil {
		cerr := f.Close()
		rerr := os.Remove(path) // Cleanup incomplete file
		if cerr != nil && rerr != nil {
			return fmt.Errorf("execute template for %s: %v; close error: %v; cleanup error: %w", path, err, cerr, rerr)
		}
		if cerr != nil {
			return fmt.Errorf("execute template for %s: %v; close error: %w", path, err, cerr)
		}
		if rerr != nil {
			return fmt.Errorf("execute template for %s: %v; cleanup error: %w", path, err, rerr)
		}
		return fmt.Errorf("execute template for %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		if rerr := os.Remove(path); rerr != nil {
			return fmt.Errorf("close file %s: %v; cleanup error: %w", path, err, rerr)
		}
		return fmt.Errorf("close file %s: %w", path, err)
	}

	return nil
}

// postInstallHookTemplate is the template for hooks/post-install.sh.
const postInstallHookTemplate = `#!/bin/bash
# {{.Name}} Layer - Post-Install Hook
# This script runs after the layer rootfs is extracted to the target image.
#
# Environment variables available:
#   LAYER_NAME - Name of this layer
#   LAYER_VERSION - Version of this layer
#   ROOTFS_PATH - Path to the mounted rootfs
#   CONFIG_* - Layer config values from manifest

set -euo pipefail

echo "[{{.Name}}] Running post-install hook..."

# Add your post-install steps here
# Example: Create symlinks, set permissions, generate configs

echo "[{{.Name}}] Post-install completed successfully"
`

func createPostInstallHook(path string, opts ScaffoldOptions) error {
	tmpl, err := template.New("post-install.sh").Parse(postInstallHookTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create file %s: %w", path, err)
	}

	if err := tmpl.Execute(f, opts); err != nil {
		cerr := f.Close()
		rerr := os.Remove(path) // Cleanup incomplete file
		if cerr != nil && rerr != nil {
			return fmt.Errorf("execute template for %s: %v; close error: %v; cleanup error: %w", path, err, cerr, rerr)
		}
		if cerr != nil {
			return fmt.Errorf("execute template for %s: %v; close error: %w", path, err, cerr)
		}
		if rerr != nil {
			return fmt.Errorf("execute template for %s: %v; cleanup error: %w", path, err, rerr)
		}
		return fmt.Errorf("execute template for %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		if rerr := os.Remove(path); rerr != nil {
			return fmt.Errorf("close file %s: %v; cleanup error: %w", path, err, rerr)
		}
		return fmt.Errorf("close file %s: %w", path, err)
	}

	return nil
}
