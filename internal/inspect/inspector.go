// Package inspect provides functionality for inspecting nanofuse ext4 images
// and extracting layer metadata from the build manifest.
package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// InspectResult represents the complete inspection result for an ext4 image.
// This struct is designed to match the JSON schema expected by the CLI output.
type InspectResult struct {
	Name           string      `json:"name"`
	BuiltAt        time.Time   `json:"built_at"`
	Kernel         KernelInfo  `json:"kernel"`
	Layers         []LayerInfo `json:"layers"`
	TotalSizeBytes int64       `json:"total_size_bytes"`
	HasMetadata    bool        `json:"has_metadata"`
	ManifestPath   string      `json:"manifest_path,omitempty"`
}

// KernelInfo represents kernel information extracted from the build manifest.
type KernelInfo struct {
	Version string `json:"version"`
	Cmdline string `json:"cmdline"`
}

// LayerInfo represents information about a single layer in the image.
type LayerInfo struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Digest    string    `json:"digest"`
	Type      string    `json:"type,omitempty"`
	SizeBytes int64     `json:"size_bytes,omitempty"`
	AppliedAt time.Time `json:"applied_at,omitempty"`
}

// BuildManifest represents the build manifest stored in /etc/nanofuse/build-manifest.json.
// This mirrors the BuildManifest struct from layerbuild package.
type BuildManifest struct {
	Version string               `json:"version"`
	Name    string               `json:"name"`
	BuiltAt time.Time            `json:"built_at"`
	Layers  []BuildManifestLayer `json:"layers"`
	Kernel  BuildManifestKernel  `json:"kernel"`
}

// BuildManifestLayer represents a layer in the build manifest.
type BuildManifestLayer struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Digest    string    `json:"digest"`
	Type      string    `json:"type"`
	AppliedAt time.Time `json:"applied_at"`
}

// BuildManifestKernel represents kernel info in the build manifest.
type BuildManifestKernel struct {
	Version string `json:"version"`
	Cmdline string `json:"cmdline"`
}

// Inspector provides methods for inspecting ext4 images.
type Inspector struct {
	// workDir is the temporary directory for mounting images
	workDir string
	// verbose enables verbose output
	verbose bool
}

// NewInspector creates a new Inspector instance.
func NewInspector(workDir string, verbose bool) *Inspector {
	return &Inspector{
		workDir: workDir,
		verbose: verbose,
	}
}

// InspectImage inspects an ext4 image file and returns the inspection result.
// It mounts the image read-only, reads the build manifest, and extracts layer information.
// If the image doesn't have layer metadata, it returns a minimal result with HasMetadata=false.
func (i *Inspector) InspectImage(ctx context.Context, imagePath string) (*InspectResult, error) {
	// Validate image path
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve image path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("image file not found: %s", absPath)
		}
		return nil, fmt.Errorf("failed to stat image file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not an image file: %s", absPath)
	}

	// Get image size
	imageSize := info.Size()

	// Try to mount and read manifest
	manifest, err := i.readManifestFromImage(ctx, absPath)
	if err != nil {
		// Image doesn't have metadata - return minimal result
		if i.verbose {
			fmt.Printf("No layer metadata found: %v\n", err)
		}
		return &InspectResult{
			Name:           filepath.Base(absPath),
			BuiltAt:        info.ModTime(),
			TotalSizeBytes: imageSize,
			HasMetadata:    false,
			Layers:         []LayerInfo{},
			Kernel:         KernelInfo{},
		}, nil
	}

	// Convert manifest to inspect result
	result := &InspectResult{
		Name:           manifest.Name,
		BuiltAt:        manifest.BuiltAt,
		TotalSizeBytes: imageSize,
		HasMetadata:    true,
		ManifestPath:   "/etc/nanofuse/build-manifest.json",
		Kernel: KernelInfo{
			Version: manifest.Kernel.Version,
			Cmdline: manifest.Kernel.Cmdline,
		},
		Layers: make([]LayerInfo, 0, len(manifest.Layers)),
	}

	for _, layer := range manifest.Layers {
		result.Layers = append(result.Layers, LayerInfo{
			Name:      layer.Name,
			Version:   layer.Version,
			Digest:    layer.Digest,
			Type:      layer.Type,
			AppliedAt: layer.AppliedAt,
		})
	}

	return result, nil
}

// readManifestFromImage mounts the image and reads the build manifest.
func (i *Inspector) readManifestFromImage(ctx context.Context, imagePath string) (*BuildManifest, error) {
	// Check if we're running as root (required for mounting)
	if os.Geteuid() != 0 {
		// Try to use debugfs to read the file without mounting
		return i.readManifestWithDebugfs(ctx, imagePath)
	}

	// Create mount point
	mountPoint := filepath.Join(i.workDir, "mnt")
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return nil, fmt.Errorf("failed to create mount point: %w", err)
	}
	defer os.RemoveAll(mountPoint)

	// Mount read-only
	if i.verbose {
		fmt.Printf("Mounting %s at %s (read-only)\n", imagePath, mountPoint)
	}

	mountCmd := exec.CommandContext(ctx, "mount", "-o", "loop,ro", imagePath, mountPoint)
	if output, err := mountCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to mount image: %w: %s", err, string(output))
	}

	// Ensure unmount on exit
	defer func() {
		if i.verbose {
			fmt.Printf("Unmounting %s\n", mountPoint)
		}
		umountCmd := exec.Command("umount", mountPoint)
		if output, err := umountCmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to unmount %s: %v: %s\n",
				mountPoint, err, string(output))
		}
	}()

	// Read manifest
	manifestPath := filepath.Join(mountPoint, "etc", "nanofuse", "build-manifest.json")
	return i.readManifestFile(manifestPath)
}

// readManifestWithDebugfs uses debugfs to read the manifest without mounting.
// This works for non-root users.
func (i *Inspector) readManifestWithDebugfs(ctx context.Context, imagePath string) (*BuildManifest, error) {
	if i.verbose {
		fmt.Printf("Using debugfs to read manifest from %s\n", imagePath)
	}

	// Use debugfs to cat the file
	// debugfs -R 'cat /etc/nanofuse/build-manifest.json' image.ext4
	cmd := exec.CommandContext(ctx, "debugfs", "-R", "cat /etc/nanofuse/build-manifest.json", imagePath)
	output, err := cmd.Output()
	if err != nil {
		// Check if debugfs is available
		if strings.Contains(err.Error(), "executable file not found") {
			return nil, fmt.Errorf("debugfs not found (install e2fsprogs); alternatively run as root to mount the image")
		}
		return nil, fmt.Errorf("failed to read manifest with debugfs: %w", err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("manifest file not found in image")
	}

	var manifest BuildManifest
	if err := json.Unmarshal(output, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// readManifestFile reads and parses a manifest file from the given path.
func (i *Inspector) readManifestFile(path string) (*BuildManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// FormatText formats the inspection result as human-readable text.
func FormatText(result *InspectResult, showLayers bool, w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}

	// Image name and build info
	fmt.Fprintf(w, "Image:         %s\n", result.Name)
	fmt.Fprintf(w, "Built:         %s\n", result.BuiltAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(w, "Total Size:    %s\n", formatBytes(result.TotalSizeBytes))
	fmt.Fprintf(w, "Has Metadata:  %v\n", result.HasMetadata)

	if !result.HasMetadata {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Note: This image does not contain NanoFuse layer metadata.")
		fmt.Fprintln(w, "      It may be a legacy image or was built without the layer system.")
		return nil
	}

	// Kernel info
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Kernel:")
	fmt.Fprintf(w, "  Version:     %s\n", result.Kernel.Version)
	if result.Kernel.Cmdline != "" {
		// Truncate cmdline if too long
		cmdline := result.Kernel.Cmdline
		if len(cmdline) > 60 {
			cmdline = cmdline[:57] + "..."
		}
		fmt.Fprintf(w, "  Cmdline:     %s\n", cmdline)
	}

	// Layer summary
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Layers:        %d\n", len(result.Layers))

	if showLayers && len(result.Layers) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Layer Details:")
		for i, layer := range result.Layers {
			fmt.Fprintf(w, "  %d. %s\n", i+1, layer.Name)
			fmt.Fprintf(w, "     Version: %s\n", layer.Version)
			if layer.Type != "" {
				fmt.Fprintf(w, "     Type:    %s\n", layer.Type)
			}
			if layer.Digest != "" {
				digest := layer.Digest
				if len(digest) > 20 {
					digest = digest[:17] + "..."
				}
				fmt.Fprintf(w, "     Digest:  %s\n", digest)
			}
			if layer.SizeBytes > 0 {
				fmt.Fprintf(w, "     Size:    %s\n", formatBytes(layer.SizeBytes))
			}
		}
	}

	return nil
}

// FormatJSON formats the inspection result as JSON.
func FormatJSON(result *InspectResult, w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// formatBytes formats a byte count as a human-readable string.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
