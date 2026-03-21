package layerbuild

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Composer composes layers into a bootable ext4 image.
type Composer struct {
	workDir   string
	outputDir string
	cache     LayerCache
	fetchers  []LayerFetcher
	verbose   bool
}

// ComposeOptions configures the composition process.
type ComposeOptions struct {
	Manifest   *ImageManifest
	Env        map[string]string
	Verbose    bool
	DryRun     bool
	OnProgress func(step string, current, total int)
}

// ComposeResult contains the build output.
type ComposeResult struct {
	RootfsPath    string   // Path to rootfs.ext4
	KernelPath    string   // Path to vmlinux
	ManifestPath  string   // Path to build-manifest.json
	LayersApplied []string // Names of applied layers
	TotalSize     int64    // Total image size
	BuildDuration time.Duration
}

// BuildManifest represents the build manifest stored in /etc/nanofuse/build-manifest.json
type BuildManifest struct {
	Version string               `json:"version"`
	Name    string               `json:"name"`
	BuiltAt time.Time            `json:"built_at"`
	Layers  []BuildManifestLayer `json:"layers"`
	Kernel  BuildManifestKernel  `json:"kernel"`
}

// BuildManifestLayer represents a layer in the build manifest
type BuildManifestLayer struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Digest    string    `json:"digest"`
	Type      string    `json:"type"`
	AppliedAt time.Time `json:"applied_at"`
}

// BuildManifestKernel represents kernel info in the build manifest
type BuildManifestKernel struct {
	Version string `json:"version"`
	Cmdline string `json:"cmdline"`
}

// NewComposer creates a new Composer instance.
func NewComposer(workDir, outputDir string, cache LayerCache, fetchers []LayerFetcher) *Composer {
	return &Composer{
		workDir:   workDir,
		outputDir: outputDir,
		cache:     cache,
		fetchers:  fetchers,
	}
}

// Compose composes layers into a bootable image according to the manifest.
// This function orchestrates the entire build process and necessarily has high complexity
// due to the many steps involved: validation, condition evaluation, dependency resolution,
// filesystem creation, layer application, hook execution, and manifest generation.
//
//nolint:gocyclo // Orchestrator function - complexity from coordinating multiple build phases
func (c *Composer) Compose(ctx context.Context, opts *ComposeOptions) (*ComposeResult, error) {
	startTime := time.Now()

	// Validate options
	if err := validateComposeOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid compose options: %w", err)
	}

	c.verbose = opts.Verbose

	// Create result structure
	result := &ComposeResult{
		LayersApplied: make([]string, 0),
	}

	// Validate manifest
	if err := ValidateManifest(opts.Manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	// Evaluate conditions to get active layers
	activeLayers := EvaluateConditions(opts.Manifest, opts.Env)
	if len(activeLayers) == 0 {
		return nil, fmt.Errorf("no active layers after condition evaluation")
	}

	if c.verbose {
		fmt.Printf("Active layers after condition evaluation: %d\n", len(activeLayers))
		for _, layer := range activeLayers {
			fmt.Printf("  - %s (%s)\n", layer.Name, layer.Type)
		}
	}

	// Resolve dependencies
	resolvedLayers, err := ResolveDependencies(activeLayers)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	if c.verbose {
		fmt.Printf("Layers in dependency order:\n")
		for i, layer := range resolvedLayers {
			fmt.Printf("  %d. %s\n", i+1, layer.Name)
		}
	}

	// Report progress
	totalSteps := len(resolvedLayers) + 4 // layers + create fs + manifest + kernel + finalize
	reportProgress := func(step string, current int) {
		if opts.OnProgress != nil {
			opts.OnProgress(step, current, totalSteps)
		}
	}

	// Step 1: Create ext4 filesystem
	reportProgress("Creating ext4 filesystem", 1)
	rootfsPath := filepath.Join(c.outputDir, fmt.Sprintf("%s-rootfs.ext4", opts.Manifest.Name))
	result.RootfsPath = rootfsPath

	if !opts.DryRun {
		if err := c.createExt4Filesystem(ctx, rootfsPath, opts.Manifest.Output.SizeMB); err != nil {
			return nil, fmt.Errorf("failed to create ext4 filesystem: %w", err)
		}
	} else if c.verbose {
		fmt.Printf("[DRY-RUN] Would create ext4 filesystem: %s (size: %dMB)\n",
			rootfsPath, opts.Manifest.Output.SizeMB)
	}

	// Step 2: Mount filesystem
	var mountPoint string
	if !opts.DryRun {
		mp, cleanup, err := c.mountFilesystem(ctx, rootfsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to mount filesystem: %w", err)
		}
		defer cleanup()
		mountPoint = mp
	} else {
		mountPoint = filepath.Join(c.workDir, "mnt")
		if c.verbose {
			fmt.Printf("[DRY-RUN] Would mount filesystem at: %s\n", mountPoint)
		}
	}

	// Step 3: Apply layers
	hookExecutor := NewHookExecutor(mountPoint, opts.DryRun, c.verbose)
	buildManifest := &BuildManifest{
		Version: "1.0",
		Name:    opts.Manifest.Name,
		BuiltAt: time.Now(),
		Layers:  make([]BuildManifestLayer, 0, len(resolvedLayers)),
		Kernel: BuildManifestKernel{
			Version: opts.Manifest.Kernel.Version,
			Cmdline: opts.Manifest.Kernel.Cmdline,
		},
	}

	for i, layer := range resolvedLayers {
		reportProgress(fmt.Sprintf("Applying layer: %s", layer.Name), i+2)

		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := c.applyLayer(ctx, &layer, mountPoint, hookExecutor, opts, buildManifest); err != nil {
			return nil, fmt.Errorf("failed to apply layer %s: %w", layer.Name, err)
		}

		result.LayersApplied = append(result.LayersApplied, layer.Name)
	}

	// Step 4: Generate build manifest
	reportProgress("Generating build manifest", len(resolvedLayers)+2)
	manifestPath := filepath.Join(c.outputDir, fmt.Sprintf("%s-build-manifest.json", opts.Manifest.Name))
	result.ManifestPath = manifestPath

	if !opts.DryRun {
		if err := c.writeBuildManifest(buildManifest, mountPoint); err != nil {
			return nil, fmt.Errorf("failed to write build manifest: %w", err)
		}

		// Also write to output directory
		if err := c.exportBuildManifest(buildManifest, manifestPath); err != nil {
			return nil, fmt.Errorf("failed to export build manifest: %w", err)
		}
	} else if c.verbose {
		fmt.Printf("[DRY-RUN] Would write build manifest to: %s\n", manifestPath)
	}

	// Step 5: Copy kernel
	reportProgress("Copying kernel", len(resolvedLayers)+3)
	kernelPath := filepath.Join(c.outputDir, "vmlinux")
	result.KernelPath = kernelPath

	if !opts.DryRun {
		if err := c.copyKernel(ctx, opts.Manifest.Kernel.Source, kernelPath); err != nil {
			return nil, fmt.Errorf("failed to copy kernel: %w", err)
		}
	} else if c.verbose {
		fmt.Printf("[DRY-RUN] Would copy kernel from %s to %s\n",
			opts.Manifest.Kernel.Source, kernelPath)
	}

	// Step 6: Finalize
	reportProgress("Finalizing image", totalSteps)

	// Get final size
	if !opts.DryRun {
		if info, err := os.Stat(rootfsPath); err == nil {
			result.TotalSize = info.Size()
		}
	}

	result.BuildDuration = time.Since(startTime)

	if c.verbose {
		fmt.Printf("\nBuild completed successfully in %v\n", result.BuildDuration)
		fmt.Printf("  Rootfs: %s\n", result.RootfsPath)
		fmt.Printf("  Kernel: %s\n", result.KernelPath)
		fmt.Printf("  Manifest: %s\n", result.ManifestPath)
		fmt.Printf("  Layers applied: %d\n", len(result.LayersApplied))
		if result.TotalSize > 0 {
			fmt.Printf("  Total size: %d bytes (%.2f MB)\n",
				result.TotalSize, float64(result.TotalSize)/(1024*1024))
		}
	}

	return result, nil
}

// createExt4Filesystem creates an ext4 filesystem of the specified size.
func (c *Composer) createExt4Filesystem(ctx context.Context, path string, sizeMB int) error {
	if c.verbose {
		fmt.Printf("Creating ext4 filesystem: %s (%dMB)\n", path, sizeMB)
	}

	// Create sparse file
	// #nosec G204 -- path and sizeMB are validated internal values, not user input
	ddCmd := exec.CommandContext(ctx, "dd",
		"if=/dev/zero",
		fmt.Sprintf("of=%s", path),
		"bs=1M",
		fmt.Sprintf("count=%d", sizeMB),
	)

	if output, err := ddCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dd failed: %w: %s", err, output)
	}

	// Format as ext4
	mkfsCmd := exec.CommandContext(ctx, "mkfs.ext4", "-F", path)
	if output, err := mkfsCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mkfs.ext4 failed: %w: %s", err, output)
	}

	if c.verbose {
		fmt.Printf("Ext4 filesystem created successfully\n")
	}

	return nil
}

// mountFilesystem mounts the ext4 image and returns the mount point and cleanup function.
func (c *Composer) mountFilesystem(ctx context.Context, imagePath string) (string, func(), error) {
	// Check for root privileges
	if os.Geteuid() != 0 {
		return "", nil, errRequiresRoot
	}

	mountPoint := filepath.Join(c.workDir, "mnt")
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create mount point: %w", err)
	}

	if c.verbose {
		fmt.Printf("Mounting %s at %s\n", imagePath, mountPoint)
	}

	mountCmd := exec.CommandContext(ctx, "mount", "-o", "loop", imagePath, mountPoint)
	if output, err := mountCmd.CombinedOutput(); err != nil {
		return "", nil, fmt.Errorf("mount failed: %w: %s", err, output)
	}

	cleanup := func() {
		if c.verbose {
			fmt.Printf("Unmounting %s\n", mountPoint)
		}
		umountCmd := exec.Command("umount", mountPoint)
		if output, err := umountCmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to unmount %s: %v: %s\n",
				mountPoint, err, output)
		}
	}

	return mountPoint, cleanup, nil
}

// applyLayer applies a single layer to the mounted filesystem.
// This function handles fetching, extracting, copying files, running hooks, and
// applying file permissions - necessarily complex due to the multi-step layer application.
//
//nolint:gocyclo // Layer application requires many conditional steps for different layer types
func (c *Composer) applyLayer(ctx context.Context, layer *LayerReference, mountPoint string,
	hookExecutor *HookExecutor, opts *ComposeOptions, buildManifest *BuildManifest) error {

	if c.verbose {
		fmt.Printf("\nApplying layer: %s (type: %s)\n", layer.Name, layer.Type)
	}

	// Get layer from cache
	digest := layer.SHA256
	if digest == "" {
		// For local sources without digest, use name as key
		digest = layer.Name
	}

	cachedLayer, err := c.cache.Get(digest)
	if err != nil {
		return fmt.Errorf("failed to get layer from cache: %w", err)
	}

	if cachedLayer == nil {
		// Layer not in cache, fetch it
		if c.verbose {
			fmt.Printf("Layer %s not in cache, fetching...\n", layer.Name)
		}

		fetched := false
		sourceType, _ := ParseSourceType(layer.Source)

		for _, fetcher := range c.fetchers {
			if fetcher.Supports(sourceType) {
				cachedLayer, err = fetcher.Fetch(layer.Source)
				if err != nil {
					return fmt.Errorf("failed to fetch layer: %w", err)
				}
				fetched = true
				break
			}
		}

		if !fetched {
			return fmt.Errorf("no fetcher available for source type: %s", sourceType)
		}

		// Add to cache
		if err := c.cache.Put(cachedLayer); err != nil {
			return fmt.Errorf("failed to cache layer: %w", err)
		}
	}

	// Touch layer to update LRU
	if err := c.cache.Touch(digest); err != nil && c.verbose {
		fmt.Printf("Warning: failed to update layer access time: %v\n", err)
	}

	// Extract layer to temporary directory
	layerExtractDir := filepath.Join(c.workDir, "extract", layer.Name)
	if !opts.DryRun {
		if err := os.MkdirAll(layerExtractDir, 0755); err != nil {
			return fmt.Errorf("failed to create extract directory: %w", err)
		}

		if err := c.extractLayer(ctx, cachedLayer.LocalPath, layerExtractDir); err != nil {
			return fmt.Errorf("failed to extract layer: %w", err)
		}
	} else if c.verbose {
		fmt.Printf("[DRY-RUN] Would extract layer from %s to %s\n",
			cachedLayer.LocalPath, layerExtractDir)
	}

	// Execute pre-install hook
	if c.verbose {
		fmt.Printf("Checking for pre-install hook...\n")
	}
	if err := hookExecutor.ExecuteLayerHooks(ctx, layerExtractDir, layer.Name, opts.Env, "pre-install"); err != nil {
		return fmt.Errorf("pre-install hook failed: %w", err)
	}

	// Copy layer rootfs to mount point
	layerRootfs := filepath.Join(layerExtractDir, "rootfs")
	if !opts.DryRun {
		if err := c.copyLayerFiles(ctx, layerRootfs, mountPoint); err != nil {
			return fmt.Errorf("failed to copy layer files: %w", err)
		}
	} else if c.verbose {
		fmt.Printf("[DRY-RUN] Would copy files from %s to %s\n", layerRootfs, mountPoint)
	}

	// Apply file permissions from layer metadata
	if cachedLayer.Metadata != nil && len(cachedLayer.Metadata.Files) > 0 && !opts.DryRun {
		if err := c.applyFilePermissions(mountPoint, cachedLayer.Metadata.Files); err != nil {
			return fmt.Errorf("failed to apply file permissions: %w", err)
		}
	}

	// Execute post-install hook
	if c.verbose {
		fmt.Printf("Checking for post-install hook...\n")
	}
	if err := hookExecutor.ExecuteLayerHooks(ctx, layerExtractDir, layer.Name, opts.Env, "post-install"); err != nil {
		return fmt.Errorf("post-install hook failed: %w", err)
	}

	// Record layer in /etc/nanofuse/layers/
	if !opts.DryRun {
		layerRecordPath := filepath.Join(mountPoint, "etc", "nanofuse", "layers", layer.Name+".json")
		if err := c.recordLayer(layerRecordPath, cachedLayer); err != nil {
			return fmt.Errorf("failed to record layer: %w", err)
		}
	}

	// Add to build manifest
	buildManifest.Layers = append(buildManifest.Layers, BuildManifestLayer{
		Name:      layer.Name,
		Version:   cachedLayer.Version,
		Digest:    cachedLayer.Digest,
		Type:      string(layer.Type),
		AppliedAt: time.Now(),
	})

	if c.verbose {
		fmt.Printf("Layer %s applied successfully\n", layer.Name)
	}

	return nil
}

// extractLayer extracts a layer tarball to the specified directory.
func (c *Composer) extractLayer(ctx context.Context, tarPath, destDir string) error {
	if c.verbose {
		fmt.Printf("Extracting %s to %s\n", tarPath, destDir)
	}

	tarCmd := exec.CommandContext(ctx, "tar", "-xzf", tarPath, "-C", destDir)
	if output, err := tarCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tar extraction failed: %w: %s", err, output)
	}

	return nil
}

// copyLayerFiles copies files from layer rootfs to mount point.
func (c *Composer) copyLayerFiles(ctx context.Context, srcDir, dstDir string) error {
	if c.verbose {
		fmt.Printf("Copying layer files from %s to %s\n", srcDir, dstDir)
	}

	// Check if source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		// No rootfs directory - layer might be metadata-only
		if c.verbose {
			fmt.Printf("No rootfs directory in layer, skipping file copy\n")
		}
		return nil
	}

	// Use cp -a to preserve permissions and ownership
	// #nosec G204 -- srcDir and dstDir are validated internal paths from layer extraction
	cpCmd := exec.CommandContext(ctx, "cp", "-a", srcDir+"/.", dstDir+"/")
	if output, err := cpCmd.CombinedOutput(); err != nil {
		// Log warning but check if it's just file conflicts
		if strings.Contains(string(output), "overwrite") {
			if c.verbose {
				fmt.Printf("Warning: files overwritten (last layer wins): %s\n", output)
			}
		} else {
			return fmt.Errorf("cp failed: %w: %s", err, output)
		}
	}

	return nil
}

// applyFilePermissions applies file permissions from layer metadata.
func (c *Composer) applyFilePermissions(mountPoint string, files []FileEntry) error {
	if c.verbose {
		fmt.Printf("Applying file permissions for %d files\n", len(files))
	}

	for _, file := range files {
		fullPath := filepath.Join(mountPoint, file.Path)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			if c.verbose {
				fmt.Printf("Warning: file %s not found, skipping permissions\n", file.Path)
			}
			continue
		}

		// Apply mode if specified
		if file.Mode != "" {
			var mode os.FileMode
			if _, err := fmt.Sscanf(file.Mode, "%o", &mode); err != nil {
				return fmt.Errorf("invalid mode %s for %s: %w", file.Mode, file.Path, err)
			}

			if err := os.Chmod(fullPath, mode); err != nil {
				return fmt.Errorf("failed to chmod %s: %w", file.Path, err)
			}
		}

		// Apply owner if specified (requires root)
		if file.Owner != "" && os.Geteuid() == 0 {
			parts := strings.Split(file.Owner, ":")
			if len(parts) == 2 {
				// #nosec G204 -- file.Owner is from layer metadata, fullPath is validated internal path
				chownCmd := exec.Command("chown", file.Owner, fullPath)
				if output, err := chownCmd.CombinedOutput(); err != nil {
					return fmt.Errorf("failed to chown %s: %w: %s", file.Path, err, output)
				}
			}
		}
	}

	return nil
}

// recordLayer records layer metadata in /etc/nanofuse/layers/
func (c *Composer) recordLayer(path string, layer *CachedLayer) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(layer.Metadata, "", "  ")
	if err != nil {
		return err
	}

	// #nosec G306 -- layer metadata must be world-readable inside the VM
	return os.WriteFile(path, data, 0644)
}

// writeBuildManifest writes the build manifest to /etc/nanofuse/build-manifest.json
func (c *Composer) writeBuildManifest(manifest *BuildManifest, mountPoint string) error {
	manifestPath := filepath.Join(mountPoint, "etc", "nanofuse", "build-manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	// #nosec G306 -- build manifest must be world-readable inside the VM for introspection
	return os.WriteFile(manifestPath, data, 0644)
}

// exportBuildManifest exports the build manifest to output directory
func (c *Composer) exportBuildManifest(manifest *BuildManifest, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	// #nosec G306 -- exported build manifest should be readable for CI/CD integration
	return os.WriteFile(path, data, 0644)
}

// copyKernel copies the kernel to the output directory
func (c *Composer) copyKernel(ctx context.Context, source, dest string) error {
	if c.verbose {
		fmt.Printf("Copying kernel from %s to %s\n", source, dest)
	}

	// Parse source type
	sourceType, ok := ParseSourceType(source)
	if !ok {
		return fmt.Errorf("invalid kernel source: %s", source)
	}

	var kernelPath string

	switch sourceType {
	case SourceTypeLocal:
		// Strip local:// prefix
		kernelPath = strings.TrimPrefix(source, "local://")
	default:
		return fmt.Errorf("unsupported kernel source type: %s", sourceType)
	}

	// Copy file
	cpCmd := exec.CommandContext(ctx, "cp", kernelPath, dest)
	if output, err := cpCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy kernel: %w: %s", err, output)
	}

	return nil
}

// validateComposeOptions validates compose options
func validateComposeOptions(opts *ComposeOptions) error {
	if opts == nil {
		return fmt.Errorf("options cannot be nil")
	}

	if opts.Manifest == nil {
		return fmt.Errorf("manifest cannot be nil")
	}

	if opts.Env == nil {
		opts.Env = make(map[string]string)
	}

	return nil
}
