// Package builder provides image extraction from OCI registries.
// This package handles converting container images into bootable microVM artifacts
// (kernel + rootfs). It is designed for eventual migration to the provenance project.
package builder

import (
	"context"
	"time"
)

// Builder extracts OCI images into bootable microVM artifacts.
type Builder interface {
	// Extract pulls an OCI image and extracts kernel + rootfs.
	// Returns paths to the extracted artifacts.
	Extract(ctx context.Context, imageRef string, opts ExtractOptions) (*ExtractResult, error)

	// Available checks if this builder can run in the current environment.
	Available() error
}

// ExtractOptions configures the extraction process.
type ExtractOptions struct {
	// OutputDir is where to place the extracted artifacts.
	// If empty, uses a temp directory under the data dir.
	OutputDir string

	// RootfsSizeMB is the size of the rootfs ext4 image in MB.
	// Default: 2048 (2GB)
	RootfsSizeMB int

	// KernelSearchPaths are paths inside the container to look for the kernel.
	// Default: ["/boot/vmlinux", "/boot/vmlinuz", "/vmlinux"]
	KernelSearchPaths []string

	// FallbackKernelPath is used when no kernel is found in the container.
	// Most container images don't include kernels, so this provides a default.
	FallbackKernelPath string

	// Verbose enables detailed logging.
	Verbose bool

	// OnProgress is called with progress updates.
	OnProgress func(stage string, percent int)
}

// ExtractResult contains the extracted artifacts.
type ExtractResult struct {
	// KernelPath is the absolute path to the extracted kernel.
	KernelPath string

	// RootfsPath is the absolute path to the rootfs ext4 image.
	RootfsPath string

	// Digest is the image digest (sha256:...).
	Digest string

	// Architecture is the image architecture (amd64, arm64, etc.).
	Architecture string

	// Labels are the image labels from the container config.
	Labels map[string]string

	// KernelVersion extracted from the image or detected from kernel binary.
	KernelVersion string

	// SizeBytes is the total size of extracted artifacts.
	SizeBytes int64

	// Duration is how long extraction took.
	Duration time.Duration
}

// DefaultExtractOptions returns sensible defaults for extraction.
func DefaultExtractOptions() ExtractOptions {
	return ExtractOptions{
		RootfsSizeMB: 2048,
		KernelSearchPaths: []string{
			"/boot/vmlinux",
			"/boot/vmlinuz",
			"/vmlinux",
			"/boot/vmlinux-*",
		},
		Verbose: false,
	}
}
