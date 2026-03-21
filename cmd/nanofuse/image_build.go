package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jpoley/nanofuse/internal/layerbuild"
	"github.com/spf13/cobra"
)

var (
	buildManifestPath string
	buildOutputDir    string
	buildVerbose      bool
	buildDryRun       bool
	buildNoCache      bool
	buildParallel     int
)

var imageBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an image from a manifest",
	Long: `Build a NanoFuse microVM image from a layer manifest.

This command composes layers into a bootable ext4 rootfs image according
to the specified manifest file.

Examples:
  # Build from a manifest file
  nanofuse image build --manifest image.manifest.yaml --output ./build/

  # Build with verbose output
  nanofuse image build -m image.manifest.yaml -o ./build/ --verbose

  # Validate manifest without building (dry-run)
  nanofuse image build -m image.manifest.yaml --dry-run

  # Build without layer caching
  nanofuse image build -m image.manifest.yaml -o ./build/ --no-cache

Note: Building images requires root privileges for filesystem operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runImageBuild(cmd.Context())
	},
}

func init() {
	imageBuildCmd.Flags().StringVarP(&buildManifestPath, "manifest", "m", "", "path to manifest file (required)")
	imageBuildCmd.Flags().StringVarP(&buildOutputDir, "output", "o", "./build", "output directory")
	imageBuildCmd.Flags().BoolVarP(&buildVerbose, "verbose", "v", false, "enable verbose output")
	imageBuildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "validate manifest without building")
	imageBuildCmd.Flags().BoolVar(&buildNoCache, "no-cache", false, "skip layer cache")
	imageBuildCmd.Flags().IntVar(&buildParallel, "parallel", 4, "number of parallel layer fetches")

	if err := imageBuildCmd.MarkFlagRequired("manifest"); err != nil {
		// This should never fail since we just defined the flag
		panic(fmt.Sprintf("failed to mark manifest flag as required: %v", err))
	}

	imageCmd.AddCommand(imageBuildCmd)
}

// buildPaths holds resolved paths for the build
type buildPaths struct {
	manifestPath string
	outputDir    string
}

// resolveBuildPaths resolves and validates input paths
func resolveBuildPaths() (*buildPaths, error) {
	manifestPath, err := filepath.Abs(buildManifestPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve manifest path: %w", err)
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file not found: %s", manifestPath)
	}

	outputDir, err := filepath.Abs(buildOutputDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve output path: %w", err)
	}

	return &buildPaths{manifestPath: manifestPath, outputDir: outputDir}, nil
}

// buildContext holds all resources needed for building
type buildContext struct {
	workDir  string
	cache    layerbuild.LayerCache
	composer *layerbuild.Composer
}

// setupBuildContext creates the build context with work directory, cache, and composer
func setupBuildContext(manifestDir, outputDir string) (*buildContext, func(), error) {
	workDir, err := os.MkdirTemp("", "nanofuse-build-")
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create work directory: %w", err)
	}

	cacheDir := filepath.Join(os.TempDir(), "nanofuse-layer-cache")
	if buildNoCache {
		cacheDir = filepath.Join(workDir, "cache")
	}
	dbPath := filepath.Join(cacheDir, "cache.db")

	cache, err := layerbuild.NewLayerCache(cacheDir, dbPath, layerbuild.DefaultCacheSizeLimit)
	if err != nil {
		os.RemoveAll(workDir)
		return nil, nil, fmt.Errorf("cannot create layer cache: %w", err)
	}

	fetchers := []layerbuild.LayerFetcher{
		layerbuild.NewLocalFetcher(manifestDir),
		layerbuild.NewDockerFetcher(workDir),
		layerbuild.NewRegistryFetcher(workDir),
	}

	composer := layerbuild.NewComposer(workDir, outputDir, cache, fetchers)

	cleanup := func() {
		cache.Close()
		os.RemoveAll(workDir)
	}

	return &buildContext{workDir: workDir, cache: cache, composer: composer}, cleanup, nil
}

// collectEnvVars collects environment variables for condition evaluation
func collectEnvVars() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				env[e[:i]] = e[i+1:]
				break
			}
		}
	}
	return env
}

// printDryRunSummary prints what would be built in dry-run mode
func printDryRunSummary(manifest *layerbuild.ImageManifest) {
	fmt.Println("Manifest validation passed!")
	fmt.Printf("\nWould build image: %s\n", manifest.Name)
	fmt.Printf("Layers to apply: %d\n", len(manifest.Layers))
	for i, layer := range manifest.Layers {
		status := ""
		if layer.Condition != "" {
			status = fmt.Sprintf(" (condition: %s)", layer.Condition)
		}
		fmt.Printf("  %d. %s (%s)%s\n", i+1, layer.Name, layer.Type, status)
	}
}

// printBuildSuccess prints the build success message
func printBuildSuccess(result *layerbuild.ComposeResult, startTime time.Time) {
	fmt.Println()
	fmt.Println("Build completed successfully!")
	fmt.Println()
	fmt.Printf("Output:\n")
	fmt.Printf("  Rootfs:   %s\n", result.RootfsPath)
	fmt.Printf("  Kernel:   %s\n", result.KernelPath)
	fmt.Printf("  Manifest: %s\n", result.ManifestPath)
	fmt.Println()
	fmt.Printf("Layers applied: %d\n", len(result.LayersApplied))
	for _, layer := range result.LayersApplied {
		fmt.Printf("  - %s\n", layer)
	}
	fmt.Println()
	fmt.Printf("Total size:  %.2f MB\n", float64(result.TotalSize)/(1024*1024))
	fmt.Printf("Build time:  %v\n", time.Since(startTime).Round(time.Millisecond))
}

func runImageBuild(ctx context.Context) error {
	startTime := time.Now()

	// Resolve and validate paths
	paths, err := resolveBuildPaths()
	if err != nil {
		return err
	}

	// Create output directory
	if !buildDryRun {
		if err := os.MkdirAll(paths.outputDir, 0755); err != nil {
			return fmt.Errorf("cannot create output directory: %w", err)
		}
	}

	if buildVerbose {
		fmt.Printf("Manifest:  %s\n", paths.manifestPath)
		fmt.Printf("Output:    %s\n", paths.outputDir)
		fmt.Printf("Dry-run:   %v\n", buildDryRun)
		fmt.Printf("No-cache:  %v\n", buildNoCache)
		fmt.Println()
	}

	// Parse and validate manifest
	manifest, err := layerbuild.ParseManifest(paths.manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse manifest: %v\n", err)
		return &exitCodeError{code: 2, message: "validation error"}
	}

	if buildVerbose {
		fmt.Printf("Image:     %s\n", manifest.Name)
		fmt.Printf("Layers:    %d\n", len(manifest.Layers))
		fmt.Println()
	}

	if err := layerbuild.ValidateManifest(manifest); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid manifest: %v\n", err)
		return &exitCodeError{code: 2, message: "validation error"}
	}

	if buildDryRun {
		printDryRunSummary(manifest)
		return nil
	}

	// Setup build context
	buildCtx, cleanup, err := setupBuildContext(filepath.Dir(paths.manifestPath), paths.outputDir)
	if err != nil {
		return err
	}
	defer cleanup()

	// Create progress reporter
	progressFn := func(step string, current, total int) {
		if buildVerbose {
			fmt.Printf("[%d/%d] %s\n", current, total, step)
		} else {
			fmt.Printf("\r[%d/%d] %s", current, total, step)
		}
	}

	// Compose layers
	opts := &layerbuild.ComposeOptions{
		Manifest:   manifest,
		Env:        collectEnvVars(),
		Verbose:    buildVerbose,
		DryRun:     buildDryRun,
		OnProgress: progressFn,
	}

	result, err := buildCtx.composer.Compose(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: build failed: %v\n", err)
		return &exitCodeError{code: 1, message: "build error"}
	}

	// Print success message
	if !buildVerbose {
		fmt.Println() // Clear progress line
	}

	printBuildSuccess(result, startTime)
	return nil
}

// exitCodeError is a simple error that carries an exit code
type exitCodeError struct {
	code    int
	message string
}

func (e *exitCodeError) Error() string {
	return e.message
}

func (e *exitCodeError) ExitCode() int {
	return e.code
}
