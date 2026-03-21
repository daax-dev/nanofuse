package layerbuild

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// namePattern validates image names: lowercase alphanumeric with hyphens,
// cannot start or end with hyphen.
var namePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// conditionPattern matches ${VAR:-default} or ${VAR} syntax.
var conditionPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)(?::-([^}]*))?\}`)

// ParseManifest reads and parses a YAML image manifest from the given path.
// It returns the parsed manifest or an error if reading or parsing fails.
func ParseManifest(path string) (*ImageManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest %s: %w", path, err)
	}

	var manifest ImageManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest %s: %w", path, err)
	}

	// Apply defaults for output config
	if manifest.Output.Format == "" {
		manifest.Output.Format = DefaultOutputConfig().Format
	}
	if manifest.Output.SizeMB == 0 {
		manifest.Output.SizeMB = DefaultOutputConfig().SizeMB
	}
	if manifest.Output.Compression == "" {
		manifest.Output.Compression = DefaultOutputConfig().Compression
	}
	if manifest.Output.Path == "" {
		manifest.Output.Path = DefaultOutputConfig().Path
	}

	return &manifest, nil
}

// ValidateManifest validates an ImageManifest and returns any validation errors.
// Returns nil if the manifest is valid.
func ValidateManifest(m *ImageManifest) error {
	var errs ValidationErrors

	// Validate version
	if m.Version == "" {
		errs.Add("version", "is required")
	} else if m.Version != "1.0" {
		errs.Add("version", fmt.Sprintf("must be '1.0', got '%s'", m.Version))
	}

	// Validate name
	if m.Name == "" {
		errs.Add("name", "is required")
	} else if !namePattern.MatchString(m.Name) {
		errs.Add("name", "must be lowercase alphanumeric with hyphens, cannot start/end with hyphen")
	}

	// Validate kernel
	validateKernel(&errs, &m.Kernel)

	// Validate layers
	if len(m.Layers) == 0 {
		errs.Add("layers", "at least one layer is required")
	}

	for i, layer := range m.Layers {
		validateLayer(&errs, &layer, i)
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// validateKernel validates the kernel configuration.
func validateKernel(errs *ValidationErrors, k *KernelConfig) {
	if k.Version == "" {
		errs.Add("kernel.version", "is required")
	}
	if k.Source == "" {
		errs.Add("kernel.source", "is required")
	}
	if k.Cmdline == "" {
		errs.Add("kernel.cmdline", "is required")
	}
}

// validateLayer validates a single layer reference.
func validateLayer(errs *ValidationErrors, l *LayerReference, idx int) {
	prefix := fmt.Sprintf("layers[%d]", idx)

	if l.Name == "" {
		errs.Add(prefix+".name", "is required")
	}

	if l.Type == "" {
		errs.Add(prefix+".type", "is required")
	} else if !l.Type.Valid() {
		errs.Add(prefix+".type", fmt.Sprintf("invalid type '%s', must be one of: base, runtime, feature, application", l.Type))
	}

	if l.Source == "" {
		errs.Add(prefix+".source", "is required")
	} else {
		sourceType, ok := ParseSourceType(l.Source)
		if !ok {
			errs.Add(prefix+".source", "must start with docker://, registry://, local://, or url://")
		} else if isRemoteSource(sourceType) && l.SHA256 == "" {
			errs.Add(prefix+".sha256", "is required for remote sources (registry://, url://)")
		}
	}
}

// isRemoteSource returns true if the source type requires SHA256 verification.
func isRemoteSource(st SourceType) bool {
	return st == SourceTypeRegistry || st == SourceTypeURL
}

// ParseSourceType extracts the source type from a source URL.
// Returns the type and true if valid, empty string and false otherwise.
func ParseSourceType(source string) (SourceType, bool) {
	if source == "" {
		return "", false
	}

	schemes := map[string]SourceType{
		"docker://":   SourceTypeDocker,
		"registry://": SourceTypeRegistry,
		"local://":    SourceTypeLocal,
		"url://":      SourceTypeURL,
	}

	for prefix, st := range schemes {
		if strings.HasPrefix(source, prefix) {
			return st, true
		}
	}

	return "", false
}

// EvaluateConditions evaluates layer conditions and returns the list of active layers.
// Layers are included if:
// - They have Required=true (always included regardless of condition)
// - They have no condition
// - Their condition evaluates to "true", "1", "yes" (case-insensitive)
//
// Condition format: ${VAR_NAME:-default}
// - If VAR_NAME exists in env, use its value
// - Otherwise, use the default value
func EvaluateConditions(m *ImageManifest, env map[string]string) []LayerReference {
	if env == nil {
		env = make(map[string]string)
	}

	var active []LayerReference

	for _, layer := range m.Layers {
		if shouldIncludeLayer(layer, env) {
			active = append(active, layer)
		}
	}

	return active
}

// shouldIncludeLayer determines if a layer should be included based on its condition.
func shouldIncludeLayer(layer LayerReference, env map[string]string) bool {
	// Required layers are always included
	if layer.Required {
		return true
	}

	// No condition means always include
	if layer.Condition == "" {
		return true
	}

	// Evaluate the condition
	value := evaluateCondition(layer.Condition, env)
	return isTruthy(value)
}

// evaluateCondition expands ${VAR:-default} syntax using the environment.
func evaluateCondition(condition string, env map[string]string) string {
	matches := conditionPattern.FindStringSubmatch(condition)
	if len(matches) < 2 {
		// Not a valid condition pattern, treat as literal
		return condition
	}

	varName := matches[1]
	defaultValue := ""
	if len(matches) >= 3 {
		defaultValue = matches[2]
	}

	if value, ok := env[varName]; ok {
		return value
	}
	return defaultValue
}

// isTruthy returns true if the value represents a truthy boolean.
func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

// ResolveDependencies performs topological sort on layers based on their dependencies.
// Returns layers in dependency order (dependencies come before dependents).
// Returns an error if there are circular dependencies or missing dependencies.
func ResolveDependencies(layers []LayerReference) ([]LayerReference, error) {
	if len(layers) == 0 {
		return nil, nil
	}

	// Build maps for lookup
	layerMap := make(map[string]*LayerReference)
	for i := range layers {
		layerMap[layers[i].Name] = &layers[i]
	}

	// Check for missing dependencies and self-dependencies
	for _, layer := range layers {
		for _, dep := range layer.Dependencies {
			if dep == layer.Name {
				return nil, fmt.Errorf("layer '%s' has circular dependency (depends on itself)", layer.Name)
			}
			if _, exists := layerMap[dep]; !exists {
				return nil, fmt.Errorf("layer '%s' depends on '%s' which is not in the layer list", layer.Name, dep)
			}
		}
	}

	// Kahn's algorithm for topological sort
	// Count incoming edges (how many layers depend on this one)
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for _, layer := range layers {
		if _, exists := inDegree[layer.Name]; !exists {
			inDegree[layer.Name] = 0
		}
		for _, dep := range layer.Dependencies {
			inDegree[layer.Name]++
			dependents[dep] = append(dependents[dep], layer.Name)
		}
	}

	// Find all layers with no dependencies (preserve original order)
	var queue []string
	for _, layer := range layers {
		if inDegree[layer.Name] == 0 {
			queue = append(queue, layer.Name)
		}
	}

	var sorted []LayerReference
	for len(queue) > 0 {
		// Pop from queue
		name := queue[0]
		queue = queue[1:]

		sorted = append(sorted, *layerMap[name])

		// For each layer that depends on this one
		for _, dependent := range dependents[name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// If we didn't process all layers, there's a cycle
	if len(sorted) != len(layers) {
		return nil, fmt.Errorf("circular dependency detected in layers")
	}

	return sorted, nil
}
