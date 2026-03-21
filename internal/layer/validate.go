package layer

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationSeverity represents the severity of a validation issue.
type ValidationSeverity string

const (
	// SeverityError indicates a critical issue that prevents layer use.
	SeverityError ValidationSeverity = "error"
	// SeverityWarning indicates an issue that should be fixed but doesn't block.
	SeverityWarning ValidationSeverity = "warning"
	// SeverityInfo indicates an informational notice.
	SeverityInfo ValidationSeverity = "info"
)

// ValidationIssue represents a single validation problem.
type ValidationIssue struct {
	// Severity indicates how critical the issue is.
	Severity ValidationSeverity
	// Field is the field or path that has the issue (may be empty).
	Field string
	// Message describes the issue.
	Message string
	// Suggestion provides a fix suggestion (may be empty).
	Suggestion string
}

// ValidationResult contains the results of layer validation.
type ValidationResult struct {
	// LayerPath is the path to the validated layer.
	LayerPath string
	// Valid is true if there are no errors (warnings are allowed).
	Valid bool
	// Issues is a list of all validation issues found.
	Issues []ValidationIssue
	// Spec is the parsed layer specification (may be nil if parsing failed).
	Spec *LayerSpec
}

// Errors returns only the error-severity issues.
func (r *ValidationResult) Errors() []ValidationIssue {
	var errors []ValidationIssue
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			errors = append(errors, issue)
		}
	}
	return errors
}

// Warnings returns only the warning-severity issues.
func (r *ValidationResult) Warnings() []ValidationIssue {
	var warnings []ValidationIssue
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			warnings = append(warnings, issue)
		}
	}
	return warnings
}

// Info returns only the info-severity issues.
func (r *ValidationResult) Info() []ValidationIssue {
	var infos []ValidationIssue
	for _, issue := range r.Issues {
		if issue.Severity == SeverityInfo {
			infos = append(infos, issue)
		}
	}
	return infos
}

// LayerSpec represents a layer.yaml file structure.
// This matches the structure used in internal/layerbuild.
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

// FileSpec describes a file in the layer.
type FileSpec struct {
	Path string `yaml:"path"`
	Mode string `yaml:"mode"`
}

// SystemdSpec describes systemd units to enable/mask.
type SystemdSpec struct {
	Enable []string `yaml:"enable"`
	Mask   []string `yaml:"mask"`
}

// ConfigFieldSpec describes a configuration field.
type ConfigFieldSpec struct {
	Type        string `yaml:"type"`
	Default     any    `yaml:"default"`
	Description string `yaml:"description"`
}

// ValidateOptions contains options for layer validation.
type ValidateOptions struct {
	// Strict enables stricter validation checks.
	Strict bool
}

// ValidateLayer validates a layer directory structure and configuration.
func ValidateLayer(layerPath string, opts ValidateOptions) *ValidationResult {
	result := &ValidationResult{
		LayerPath: layerPath,
		Valid:     true,
		Issues:    []ValidationIssue{},
	}

	// Check layer directory exists
	info, err := os.Stat(layerPath)
	if err != nil {
		result.addError("", "Layer directory does not exist: "+layerPath, "Create the layer directory or check the path")
		result.Valid = false
		return result
	}
	if !info.IsDir() {
		result.addError("", "Path is not a directory: "+layerPath, "Provide a directory path, not a file")
		result.Valid = false
		return result
	}

	// Validate layer.yaml
	layerYAMLPath := filepath.Join(layerPath, "layer.yaml")
	spec, specIssues := validateLayerYAML(layerYAMLPath, layerPath, opts)
	result.Issues = append(result.Issues, specIssues...)
	result.Spec = spec

	// Validate rootfs directory
	rootfsPath := filepath.Join(layerPath, "rootfs")
	rootfsIssues := validateRootfs(rootfsPath, opts)
	result.Issues = append(result.Issues, rootfsIssues...)

	// Validate hooks directory
	hooksPath := filepath.Join(layerPath, "hooks")
	hooksIssues := validateHooks(hooksPath, opts)
	result.Issues = append(result.Issues, hooksIssues...)

	// Check for any errors
	for _, issue := range result.Issues {
		if issue.Severity == SeverityError {
			result.Valid = false
			break
		}
	}

	return result
}

// validateLayerYAML validates the layer.yaml file.
func validateLayerYAML(path string, layerPath string, opts ValidateOptions) (*LayerSpec, []ValidationIssue) {
	var issues []ValidationIssue

	// Check file exists and parse YAML
	spec, parseIssues := parseLayerYAML(path)
	if len(parseIssues) > 0 {
		return nil, parseIssues
	}

	// Validate required fields
	issues = append(issues, validateSpecRequiredFields(spec, layerPath)...)

	// Validate dependencies format
	issues = append(issues, validateSpecDependencies(spec)...)

	// Strict mode checks
	if opts.Strict {
		issues = append(issues, validateSpecStrictMode(spec, layerPath)...)
	}

	return spec, issues
}

// parseLayerYAML reads and parses the layer.yaml file.
func parseLayerYAML(path string) (*LayerSpec, []ValidationIssue) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, []ValidationIssue{{
				Severity:   SeverityError,
				Field:      "layer.yaml",
				Message:    "layer.yaml not found",
				Suggestion: "Create a layer.yaml file with 'nanofuse layer create' or manually",
			}}
		}
		return nil, []ValidationIssue{{
			Severity:   SeverityError,
			Field:      "layer.yaml",
			Message:    "Cannot read layer.yaml: " + err.Error(),
			Suggestion: "Check file permissions",
		}}
	}

	var spec LayerSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, []ValidationIssue{{
			Severity:   SeverityError,
			Field:      "layer.yaml",
			Message:    "Invalid YAML syntax: " + err.Error(),
			Suggestion: "Check YAML syntax - common issues: incorrect indentation, missing quotes",
		}}
	}

	return &spec, nil
}

// validateSpecRequiredFields validates name, version, and type fields.
func validateSpecRequiredFields(spec *LayerSpec, layerPath string) []ValidationIssue {
	var issues []ValidationIssue

	if spec.Name == "" {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "name",
			Message:    "Layer name is required",
			Suggestion: "Add 'name: \"your-layer-name\"' to layer.yaml",
		})
	} else {
		if err := validateLayerName(spec.Name); err != nil {
			issues = append(issues, ValidationIssue{
				Severity:   SeverityError,
				Field:      "name",
				Message:    err.Error(),
				Suggestion: "Use lowercase letters, numbers, and hyphens only",
			})
		}
		dirName := filepath.Base(layerPath)
		if spec.Name != dirName {
			issues = append(issues, ValidationIssue{
				Severity:   SeverityWarning,
				Field:      "name",
				Message:    fmt.Sprintf("Layer name %q does not match directory name %q", spec.Name, dirName),
				Suggestion: "Rename directory to match layer name or update layer.yaml",
			})
		}
	}

	if spec.Version == "" {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "version",
			Message:    "Layer version is required",
			Suggestion: "Add 'version: \"1.0.0\"' to layer.yaml",
		})
	}

	if spec.Type == "" {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "type",
			Message:    "Layer type is required",
			Suggestion: fmt.Sprintf("Add 'type: \"<type>\"' where type is one of: %v", ValidLayerTypes),
		})
	} else if !IsValidLayerType(spec.Type) {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "type",
			Message:    fmt.Sprintf("Invalid layer type %q", spec.Type),
			Suggestion: fmt.Sprintf("Use one of: %v", ValidLayerTypes),
		})
	}

	return issues
}

// validateSpecDependencies validates the dependencies array.
func validateSpecDependencies(spec *LayerSpec) []ValidationIssue {
	var issues []ValidationIssue
	for i, dep := range spec.Dependencies {
		if dep == "" {
			issues = append(issues, ValidationIssue{
				Severity:   SeverityError,
				Field:      fmt.Sprintf("dependencies[%d]", i),
				Message:    "Empty dependency string",
				Suggestion: "Remove empty dependency or provide a valid layer name",
			})
		}
	}
	return issues
}

// validateSpecStrictMode performs additional checks when strict mode is enabled.
func validateSpecStrictMode(spec *LayerSpec, layerPath string) []ValidationIssue {
	var issues []ValidationIssue

	if spec.Description == "" {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityWarning,
			Field:      "description",
			Message:    "Layer description is empty",
			Suggestion: "Add a description to help users understand this layer's purpose",
		})
	}

	issues = append(issues, validateSpecProvides(spec)...)
	issues = append(issues, validateSpecSHA256(spec, layerPath)...)

	return issues
}

// validateSpecProvides validates the provides array in strict mode.
func validateSpecProvides(spec *LayerSpec) []ValidationIssue {
	var issues []ValidationIssue

	if len(spec.Provides) == 0 {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityWarning,
			Field:      "provides",
			Message:    "Layer provides no capabilities",
			Suggestion: "Add capabilities this layer provides (e.g., 'python', 'nodejs')",
		})
	} else {
		for i, p := range spec.Provides {
			if p == "" {
				issues = append(issues, ValidationIssue{
					Severity:   SeverityError,
					Field:      fmt.Sprintf("provides[%d]", i),
					Message:    "Empty provides string",
					Suggestion: "Remove empty provides entry or provide a valid capability name",
				})
			}
		}
	}

	return issues
}

// validateSpecSHA256 validates the SHA256 hash for built layers.
func validateSpecSHA256(spec *LayerSpec, layerPath string) []ValidationIssue {
	var issues []ValidationIssue

	tarballPath := filepath.Join(layerPath, spec.Name+".tar.gz")
	if _, err := os.Stat(tarballPath); err != nil {
		return issues // Tarball doesn't exist, skip SHA256 check
	}

	if spec.SHA256 == "" {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityWarning,
			Field:      "sha256",
			Message:    "Built layer missing SHA256 hash",
			Suggestion: "Rebuild the layer to generate SHA256 hash",
		})
	} else if len(spec.SHA256) != 64 {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "sha256",
			Message:    fmt.Sprintf("Invalid SHA256 hash length: %d (expected 64)", len(spec.SHA256)),
			Suggestion: "Rebuild the layer to generate valid SHA256 hash",
		})
	} else if _, err := hex.DecodeString(spec.SHA256); err != nil {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "sha256",
			Message:    "Invalid SHA256 hash: contains non-hexadecimal characters",
			Suggestion: "Rebuild the layer to generate valid SHA256 hash",
		})
	}

	return issues
}

// validateRootfs validates the rootfs directory.
func validateRootfs(path string, opts ValidateOptions) []ValidationIssue {
	var issues []ValidationIssue

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, ValidationIssue{
				Severity:   SeverityError,
				Field:      "rootfs/",
				Message:    "rootfs/ directory not found",
				Suggestion: "Create a rootfs/ directory containing the layer's filesystem contents",
			})
		} else {
			issues = append(issues, ValidationIssue{
				Severity:   SeverityError,
				Field:      "rootfs/",
				Message:    "Cannot access rootfs/: " + err.Error(),
				Suggestion: "Check directory permissions",
			})
		}
		return issues
	}

	if !info.IsDir() {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "rootfs",
			Message:    "rootfs is not a directory",
			Suggestion: "Remove rootfs file and create rootfs/ directory",
		})
		return issues
	}

	// Check if rootfs is empty (only .gitkeep is allowed)
	entries, err := os.ReadDir(path)
	if err != nil {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityWarning,
			Field:      "rootfs/",
			Message:    "Cannot read rootfs/ contents: " + err.Error(),
			Suggestion: "Check directory permissions",
		})
	} else {
		contentFileCount := 0
		for _, entry := range entries {
			if entry.Name() != ".gitkeep" {
				contentFileCount++
			}
		}
		if contentFileCount == 0 && opts.Strict {
			issues = append(issues, ValidationIssue{
				Severity:   SeverityWarning,
				Field:      "rootfs/",
				Message:    "rootfs/ directory is empty",
				Suggestion: "Add filesystem contents or build the layer",
			})
		}
	}

	return issues
}

// validateHooks validates the hooks directory and its scripts.
func validateHooks(path string, opts ValidateOptions) []ValidationIssue {
	var issues []ValidationIssue

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// hooks/ is optional but recommended
			if opts.Strict {
				issues = append(issues, ValidationIssue{
					Severity:   SeverityInfo,
					Field:      "hooks/",
					Message:    "No hooks/ directory found",
					Suggestion: "Consider adding hooks for post-install customization",
				})
			}
		}
		return issues
	}

	if !info.IsDir() {
		issues = append(issues, ValidationIssue{
			Severity:   SeverityError,
			Field:      "hooks",
			Message:    "hooks is not a directory",
			Suggestion: "Remove hooks file and create hooks/ directory",
		})
		return issues
	}

	// Check each hook file
	entries, err := os.ReadDir(path)
	if err != nil {
		return issues
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".sh") {
			continue
		}

		hookPath := filepath.Join(path, entry.Name())
		hookInfo, err := os.Stat(hookPath)
		if err != nil {
			continue
		}

		// Check if hook is executable
		mode := hookInfo.Mode()
		if mode&0111 == 0 {
			issues = append(issues, ValidationIssue{
				Severity:   SeverityError,
				Field:      "hooks/" + entry.Name(),
				Message:    fmt.Sprintf("Hook %s is not executable", entry.Name()),
				Suggestion: fmt.Sprintf("Run: chmod +x %s", hookPath),
			})
		}

		// Check shebang in strict mode
		if opts.Strict {
			data, err := os.ReadFile(hookPath)
			if err == nil {
				content := string(data)
				if !strings.HasPrefix(content, "#!/") {
					issues = append(issues, ValidationIssue{
						Severity:   SeverityWarning,
						Field:      "hooks/" + entry.Name(),
						Message:    fmt.Sprintf("Hook %s missing shebang", entry.Name()),
						Suggestion: "Add '#!/bin/bash' or '#!/bin/sh' at the start of the file",
					})
				}
			}
		}
	}

	return issues
}

// Helper methods for ValidationResult

func (r *ValidationResult) addError(field, message, suggestion string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Severity:   SeverityError,
		Field:      field,
		Message:    message,
		Suggestion: suggestion,
	})
}
