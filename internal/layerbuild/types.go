// Package layerbuild provides types and interfaces for the NanoFuse layer build system.
// It defines the core data structures for image manifests, layers, and layer management.
package layerbuild

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested layer is not found in the cache.
// This allows callers to distinguish between "not found" and actual errors.
var ErrNotFound = errors.New("layer not found")

// ValidationError represents a single validation error with field context.
type ValidationError struct {
	Field   string
	Message string
	Line    int // Optional line number in source file
}

// Error implements the error interface for ValidationError.
func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s: %s", e.Line, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Add adds a validation error for the given field.
func (v *ValidationErrors) Add(field, message string) {
	*v = append(*v, ValidationError{Field: field, Message: message})
}

// AddWithLine adds a validation error with line number context.
func (v *ValidationErrors) AddWithLine(field, message string, line int) {
	*v = append(*v, ValidationError{Field: field, Message: message, Line: line})
}

// HasErrors returns true if there are any validation errors.
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// Error implements the error interface.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}
	if len(v) == 1 {
		return v[0].Error()
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d validation errors:\n", len(v)))
	for _, e := range v {
		sb.WriteString(fmt.Sprintf("  - %s\n", e.Error()))
	}
	return sb.String()
}

// validOutputFormats defines the supported output image formats.
// This is a package-level variable to avoid repeated allocations during validation.
var validOutputFormats = map[string]bool{
	"ext4":     true,
	"squashfs": true,
	"tar":      true,
}

// Compression type constants
const (
	CompressionNone = "none"
	CompressionGzip = "gzip"
	CompressionZstd = "zstd"
)

// LayerType represents the type of a layer in the image build system.
type LayerType string

const (
	// LayerTypeBase represents a base operating system layer (e.g., Alpine, Debian)
	LayerTypeBase LayerType = "base"

	// LayerTypeRuntime represents a language runtime layer (e.g., Python, Node.js, Go)
	LayerTypeRuntime LayerType = "runtime"

	// LayerTypeFeature represents a feature or capability layer (e.g., networking, debugging)
	LayerTypeFeature LayerType = "feature"

	// LayerTypeApplication represents an application-specific layer
	LayerTypeApplication LayerType = "application"
)

// String returns the string representation of the LayerType.
func (lt LayerType) String() string {
	return string(lt)
}

// IsValid returns true if the LayerType is a recognized type.
func (lt LayerType) IsValid() bool {
	switch lt {
	case LayerTypeBase, LayerTypeRuntime, LayerTypeFeature, LayerTypeApplication:
		return true
	default:
		return false
	}
}

// Valid is an alias for IsValid for compatibility.
func (lt LayerType) Valid() bool {
	return lt.IsValid()
}

// SourceType represents the type of layer source.
type SourceType string

const (
	// SourceTypeLocal represents a local filesystem source (local://)
	SourceTypeLocal SourceType = "local"

	// SourceTypeDocker represents a Docker image source (docker://)
	SourceTypeDocker SourceType = "docker"

	// SourceTypeRegistry represents an OCI registry source (registry://)
	SourceTypeRegistry SourceType = "registry"

	// SourceTypeURL represents an HTTP/HTTPS URL source (http:// or https://)
	SourceTypeURL SourceType = "url"
)

// String returns the string representation of the SourceType.
func (st SourceType) String() string {
	return string(st)
}

// Validate returns an error if the LayerType is not valid.
func (lt LayerType) Validate() error {
	if !lt.IsValid() {
		return fmt.Errorf("invalid layer type: %q (must be one of: base, runtime, feature, application)", string(lt))
	}
	return nil
}

// validateSHA256 validates a SHA256 checksum string.
// If required is true, an empty string is considered an error.
// If required is false, an empty string is allowed.
// Returns an error if the checksum is not exactly 64 valid hexadecimal characters.
func validateSHA256(sha256 string, required bool) error {
	if sha256 == "" {
		if required {
			return errors.New("SHA256 checksum is required")
		}
		return nil
	}
	if len(sha256) != 64 {
		return fmt.Errorf("SHA256 must be 64 characters (got %d)", len(sha256))
	}
	// Use encoding/hex to validate hexadecimal format
	if _, err := hex.DecodeString(sha256); err != nil {
		return fmt.Errorf("SHA256 must be valid hexadecimal: %w", err)
	}
	return nil
}

// ImageManifest represents the top-level manifest for building VM images.
// It defines all the components needed to construct a bootable microVM image.
type ImageManifest struct {
	// Version of the manifest format (e.g., "1.0")
	Version string `yaml:"version"`

	// Name of the image being built
	Name string `yaml:"name"`

	// Description of the image
	Description string `yaml:"description,omitempty"`

	// Kernel configuration for the VM
	Kernel KernelConfig `yaml:"kernel"`

	// Layers to be composed into the final image
	Layers []LayerReference `yaml:"layers"`

	// Output configuration for the built image
	Output OutputConfig `yaml:"output"`
}

// Validate checks if the ImageManifest is valid.
func (im *ImageManifest) Validate() error {
	if im.Version == "" {
		return fmt.Errorf("manifest version is required")
	}
	if im.Name == "" {
		return fmt.Errorf("manifest name is required")
	}
	if err := im.Kernel.Validate(); err != nil {
		return fmt.Errorf("kernel config validation failed: %w", err)
	}
	if len(im.Layers) == 0 {
		return fmt.Errorf("at least one layer is required")
	}
	for i, layer := range im.Layers {
		if err := layer.Validate(); err != nil {
			return fmt.Errorf("layer %d validation failed: %w", i, err)
		}
	}
	if err := im.Output.Validate(); err != nil {
		return fmt.Errorf("output config validation failed: %w", err)
	}
	return nil
}

// KernelConfig represents the kernel configuration for a microVM.
type KernelConfig struct {
	// Version of the kernel (e.g., "5.10", "6.1")
	Version string `yaml:"version"`

	// Source location for the kernel binary (URL or local path)
	Source string `yaml:"source"`

	// SHA256 checksum of the kernel binary for verification
	SHA256 string `yaml:"sha256"`

	// Cmdline kernel command-line parameters
	Cmdline string `yaml:"cmdline"`
}

// Validate checks if the KernelConfig is valid.
func (kc *KernelConfig) Validate() error {
	if kc.Version == "" {
		return fmt.Errorf("kernel version is required")
	}
	if kc.Source == "" {
		return fmt.Errorf("kernel source is required")
	}
	if err := validateSHA256(kc.SHA256, true); err != nil {
		return fmt.Errorf("invalid kernel SHA256: %w", err)
	}
	return nil
}

// OutputConfig specifies how the built image should be output.
type OutputConfig struct {
	// Path where the image should be written
	Path string `yaml:"path"`

	// Format of the output image (e.g., "ext4", "squashfs")
	Format string `yaml:"format"`

	// SizeMB is the size of the output image in megabytes
	SizeMB int `yaml:"size_mb,omitempty"`

	// Compression algorithm to use (e.g., "gzip", "none")
	Compression string `yaml:"compression,omitempty"`
}

// DefaultOutputConfig returns the default output configuration.
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		Path:        "./build",
		Format:      "ext4",
		SizeMB:      2048, // 2GB default
		Compression: CompressionNone,
	}
}

// Validate checks if the OutputConfig is valid.
func (oc *OutputConfig) Validate() error {
	if oc.Path == "" {
		return fmt.Errorf("output path is required")
	}
	if oc.Format == "" {
		return fmt.Errorf("output format is required")
	}
	if !validOutputFormats[strings.ToLower(oc.Format)] {
		return fmt.Errorf("unsupported output format: %q (must be one of: ext4, squashfs, tar)", oc.Format)
	}
	return nil
}

// LayerReference represents a reference to a layer in the manifest.
// It specifies where to fetch the layer and how to use it.
type LayerReference struct {
	// Name of the layer
	Name string `yaml:"name"`

	// Type of the layer (base, runtime, feature, application)
	Type LayerType `yaml:"type"`

	// Source location for the layer (URL, registry path, or local path)
	Source string `yaml:"source"`

	// SHA256 checksum for verification (optional if fetcher provides it)
	SHA256 string `yaml:"sha256,omitempty"`

	// Condition for conditional layer inclusion (optional)
	Condition string `yaml:"condition,omitempty"`

	// Required indicates if this layer is always included regardless of condition
	Required bool `yaml:"required,omitempty"`

	// Dependencies lists layer names that must be applied before this one
	Dependencies []string `yaml:"dependencies,omitempty"`

	// Config provides layer-specific configuration
	Config map[string]any `yaml:"config,omitempty"`
}

// Validate checks if the LayerReference is valid.
func (lr *LayerReference) Validate() error {
	if lr.Name == "" {
		return fmt.Errorf("layer name is required")
	}
	if err := lr.Type.Validate(); err != nil {
		return fmt.Errorf("layer type validation failed: %w", err)
	}
	if lr.Source == "" {
		return fmt.Errorf("layer source is required")
	}
	if err := validateSHA256(lr.SHA256, false); err != nil {
		return fmt.Errorf("invalid layer SHA256: %w", err)
	}
	return nil
}

// LayerPackage represents metadata for a packaged layer.
// This is the runtime representation of a layer after it has been fetched.
type LayerPackage struct {
	// Name of the layer
	Name string `yaml:"name" json:"name"`

	// Version of the layer
	Version string `yaml:"version" json:"version"`

	// Description of the layer
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Type of the layer
	Type LayerType `yaml:"type" json:"type"`

	// Size in bytes of the layer package
	Size int64 `yaml:"size" json:"size"`

	// SHA256 checksum of the layer
	SHA256 string `yaml:"sha256" json:"sha256"`

	// RootFS is the path to the extracted rootfs directory
	RootFS string `yaml:"rootfs" json:"rootfs"`

	// Dependencies lists layer names that must be present
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`

	// Provides lists capabilities this layer provides
	Provides []string `yaml:"provides,omitempty" json:"provides,omitempty"`

	// Files contains file entries for permission management
	Files []FileEntry `yaml:"files,omitempty" json:"files,omitempty"`

	// Systemd configuration for the layer
	Systemd *SystemdConfig `yaml:"systemd,omitempty" json:"systemd,omitempty"`

	// ConfigSchema defines configurable options for the layer
	ConfigSchema map[string]ConfigOption `yaml:"config_schema,omitempty" json:"config_schema,omitempty"`

	// Metadata provides additional key-value metadata about the layer
	Metadata map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// SystemdConfig represents systemd unit configuration for a layer.
type SystemdConfig struct {
	// Enable lists services to enable
	Enable []string `yaml:"enable,omitempty" json:"enable,omitempty"`

	// Mask lists services to mask
	Mask []string `yaml:"mask,omitempty" json:"mask,omitempty"`
}

// ConfigOption represents a configurable option for a layer.
type ConfigOption struct {
	// Type of the option (string, integer, boolean, etc.)
	Type string `yaml:"type" json:"type"`

	// Default value for the option
	Default any `yaml:"default,omitempty" json:"default,omitempty"`

	// Description of the option
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Required indicates if the option must be set
	Required bool `yaml:"required,omitempty" json:"required,omitempty"`
}

// Validate checks if the LayerPackage is valid.
func (lp *LayerPackage) Validate() error {
	if lp.Name == "" {
		return fmt.Errorf("layer package name is required")
	}
	if lp.Version == "" {
		return fmt.Errorf("layer package version is required")
	}
	if err := lp.Type.Validate(); err != nil {
		return fmt.Errorf("layer package type validation failed: %w", err)
	}
	if lp.Size <= 0 {
		return fmt.Errorf("layer package size must be positive (got %d)", lp.Size)
	}
	if err := validateSHA256(lp.SHA256, true); err != nil {
		return fmt.Errorf("invalid layer package SHA256: %w", err)
	}
	if lp.RootFS == "" {
		return fmt.Errorf("layer package rootfs path is required")
	}
	return nil
}

// LayerFetcher is the interface for fetching layers from various sources.
// Implementations might fetch from local filesystem, Docker registry, HTTP URLs, etc.
type LayerFetcher interface {
	// Supports returns true if this fetcher can handle the given source type.
	Supports(sourceType SourceType) bool

	// Fetch retrieves a layer from the specified source and returns a CachedLayer.
	// The source format depends on the implementation (e.g., "docker://alpine:3.18", "local:///path/to/layer").
	// Returns an error if the layer cannot be fetched or verified.
	Fetch(source string) (*CachedLayer, error)
}

// LayerCache is the interface for caching layers to avoid redundant fetches.
// Implementations might use local disk cache, distributed cache, etc.
type LayerCache interface {
	// Get retrieves a cached layer by its SHA256 digest.
	// Returns nil, nil if the layer is not cached.
	// Returns nil, error if an error occurred while checking.
	Get(digest string) (*CachedLayer, error)

	// Put stores a layer in the cache.
	// If a layer with the same digest already exists, it will be updated.
	// Returns an error if the layer cannot be stored.
	Put(layer *CachedLayer) error

	// Exists checks if a layer with the given SHA256 digest exists in the cache.
	// Returns (true, nil) if the layer exists.
	// Returns (false, nil) if the layer does not exist.
	// Returns (false, err) if an error occurred while checking.
	Exists(digest string) (bool, error)

	// Touch updates the LastUsedAt timestamp for LRU tracking.
	Touch(digest string) error

	// Evict removes layers to free space using LRU policy.
	// Returns the number of bytes freed.
	Evict(targetBytes int64) (int64, error)

	// Stats returns cache statistics.
	Stats() (*CacheStats, error)

	// Close releases any resources held by the cache.
	Close() error
}

// CachedLayer represents a layer stored in the cache.
// This is the runtime representation with cache-specific metadata.
type CachedLayer struct {
	// Digest is the SHA256 digest of the layer tarball
	Digest string

	// Name of the layer
	Name string

	// Version of the layer
	Version string

	// Type of the layer
	Type LayerType

	// SourceURL is the original source URL
	SourceURL string

	// LocalPath is the path to the cached tarball
	LocalPath string

	// SizeBytes is the size of the tarball in bytes
	SizeBytes int64

	// FetchedAt is when the layer was first cached
	FetchedAt time.Time

	// LastUsedAt is when the layer was last accessed
	LastUsedAt time.Time

	// Metadata contains the layer package metadata
	Metadata *LayerPackage

	// Files contains file entries for permission management
	Files []FileEntry
}

// CacheStats contains cache statistics.
type CacheStats struct {
	// TotalLayers is the number of layers in the cache
	TotalLayers int

	// TotalBytes is the total size of cached layers in bytes
	TotalBytes int64

	// OldestAccess is the timestamp of the least recently used layer
	OldestAccess time.Time
}

// FileEntry represents a file with custom permissions in a layer.
type FileEntry struct {
	// Path is the file path relative to the layer root
	Path string `yaml:"path"`

	// Mode is the file mode as an octal string (e.g., "0755")
	Mode string `yaml:"mode,omitempty"`

	// Owner is the file owner (e.g., "root")
	Owner string `yaml:"owner,omitempty"`

	// Group is the file group (e.g., "root")
	Group string `yaml:"group,omitempty"`
}
