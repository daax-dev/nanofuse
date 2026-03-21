package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/jpoley/nanofuse/internal/builder"
	"github.com/jpoley/nanofuse/internal/logging"
	"github.com/jpoley/nanofuse/internal/types"
)

// ClientOptions holds configuration for the registry client
type ClientOptions struct {
	PullTimeout  time.Duration // Total timeout for pull operation
	LayerTimeout time.Duration // Timeout per layer download
	UseDocker    bool          // Use Docker/Podman for extraction (default: true)
	Verbose      bool          // Enable verbose logging
}

// Client represents an OCI registry client
type Client struct {
	dataDir string
	logger  *logging.Logger
	opts    ClientOptions
	builder builder.Builder
}

// NewClient creates a new registry client
func NewClient(dataDir string, logger *logging.Logger, opts ClientOptions) *Client {
	// Apply defaults
	if opts.PullTimeout == 0 {
		opts.PullTimeout = 10 * time.Minute
	}
	if opts.LayerTimeout == 0 {
		opts.LayerTimeout = 5 * time.Minute
	}

	client := &Client{
		dataDir: dataDir,
		logger:  logger,
		opts:    opts,
	}

	// Initialize Docker builder if available (default behavior)
	// UseDocker defaults to true (false means not explicitly set), so we always try Docker
	dockerBuilder := builder.NewDockerBuilder(dataDir, opts.Verbose)
	if err := dockerBuilder.Available(); err == nil {
		client.builder = dockerBuilder
		logger.Info("Docker/Podman available for image extraction")
	} else {
		logger.Warn("Docker/Podman not available: %v - image pull will use metadata only", err)
	}

	return client
}

// PullImage pulls an image from a registry
func (c *Client) PullImage(ctx context.Context, imageRef string, progressChan chan<- *types.PullProgress) (*types.Image, error) {
	startTime := time.Now()
	c.logger.Info("Starting image pull: %s", imageRef)
	c.logger.Debug("Pull timeout: %v, Layer timeout: %v", c.opts.PullTimeout, c.opts.LayerTimeout)

	// Create context with timeout for the entire pull operation
	pullCtx, cancel := context.WithTimeout(ctx, c.opts.PullTimeout)
	defer cancel()

	// Parse image reference
	c.logger.Debug("Parsing image reference...")
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		c.logger.Error("Failed to parse image reference '%s': %v", imageRef, err)
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}
	c.logger.Debug("Parsed reference: registry=%s, repository=%s", ref.Context().RegistryStr(), ref.Context().RepositoryStr())

	// Get authenticator from Docker config
	c.logger.Debug("Resolving authentication for registry: %s", ref.Context().RegistryStr())
	auth, err := authn.DefaultKeychain.Resolve(ref.Context().Registry)
	if err != nil {
		c.logger.Error("Failed to get authentication for registry '%s': %v", ref.Context().RegistryStr(), err)
		return nil, fmt.Errorf("failed to get authentication: %w", err)
	}
	c.logger.Debug("Authentication resolved successfully")

	// Report initial progress
	if progressChan != nil {
		progressChan <- &types.PullProgress{
			CurrentBytes: 0,
			TotalBytes:   0,
			Percentage:   0,
		}
	}

	// Fetch image manifest (this is where network calls happen)
	c.logger.Info("Fetching image manifest from registry...")
	manifestStart := time.Now()
	img, err := remote.Image(ref, remote.WithAuth(auth), remote.WithContext(pullCtx))
	if err != nil {
		c.logger.Error("Failed to fetch image from registry after %v: %v", time.Since(manifestStart), err)
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}
	c.logger.Info("Manifest fetched in %v", time.Since(manifestStart))

	// Get image digest
	c.logger.Debug("Extracting image digest...")
	digest, err := img.Digest()
	if err != nil {
		c.logger.Error("Failed to get image digest: %v", err)
		return nil, fmt.Errorf("failed to get image digest: %w", err)
	}
	c.logger.Info("Image digest: %s", digest.String())

	// Get image size (from manifest, doesn't download layers)
	c.logger.Debug("Getting image size from manifest...")
	size, err := img.Size()
	if err != nil {
		c.logger.Error("Failed to get image size: %v", err)
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}
	c.logger.Info("Image size: %d bytes (%.2f MB)", size, float64(size)/(1024*1024))

	// Report size known
	if progressChan != nil {
		progressChan <- &types.PullProgress{
			CurrentBytes: 0,
			TotalBytes:   size,
			Percentage:   0,
		}
	}

	// Get image config (this fetches the config blob from registry)
	c.logger.Debug("Fetching image config blob...")
	configStart := time.Now()
	configFile, err := img.ConfigFile()
	if err != nil {
		c.logger.Error("Failed to get image config after %v: %v", time.Since(configStart), err)
		return nil, fmt.Errorf("failed to get image config: %w", err)
	}
	c.logger.Info("Config blob fetched in %v", time.Since(configStart))

	// Extract labels from config
	labels := make(map[string]string)
	if configFile != nil && configFile.Config.Labels != nil {
		labels = configFile.Config.Labels
		c.logger.Debug("Found %d labels in image config", len(labels))
	}

	// Extract architecture from config
	architecture := "unknown"
	if configFile != nil && configFile.Architecture != "" {
		architecture = configFile.Architecture
	}
	c.logger.Info("Image architecture: %s", architecture)

	// Get layer information
	layers, err := img.Layers()
	if err != nil {
		c.logger.Error("Failed to get image layers: %v", err)
		return nil, fmt.Errorf("failed to get image layers: %w", err)
	}
	c.logger.Info("Image has %d layers", len(layers))

	// Log layer details
	for i, layer := range layers {
		layerDigest, _ := layer.Digest()
		layerSize, _ := layer.Size()
		c.logger.Debug("  Layer %d: %s (%.2f MB)", i+1, layerDigest.String()[:16], float64(layerSize)/(1024*1024))
	}

	// Create image directory
	imageDir := filepath.Join(c.dataDir, "images", digest.String())
	c.logger.Debug("Creating image directory: %s", imageDir)
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		c.logger.Error("Failed to create image directory: %v", err)
		return nil, fmt.Errorf("failed to create image directory: %w", err)
	}

	var rootfsPath, kernelPath, kernelVersion string

	// Use Docker builder if available for actual layer extraction
	if c.builder != nil {
		c.logger.Info("Extracting image layers using Docker/Podman...")

		extractOpts := builder.ExtractOptions{
			OutputDir:    imageDir,
			RootfsSizeMB: 2048,
			Verbose:      c.opts.Verbose,
			OnProgress: func(stage string, percent int) {
				c.logger.Debug("Extraction: %s (%d%%)", stage, percent)
				if progressChan != nil {
					// Map extraction progress to 50-100% range (first 50% is metadata)
					adjustedPercent := 50 + (percent / 2)
					progressChan <- &types.PullProgress{
						CurrentBytes: int64(adjustedPercent) * size / 100,
						TotalBytes:   size,
						Percentage:   adjustedPercent,
					}
				}
			},
		}

		result, err := c.builder.Extract(ctx, imageRef, extractOpts)
		if err != nil {
			c.logger.Error("Failed to extract image layers: %v", err)
			return nil, fmt.Errorf("failed to extract image layers: %w", err)
		}

		rootfsPath = result.RootfsPath
		kernelPath = result.KernelPath
		kernelVersion = result.KernelVersion

		c.logger.Info("Extraction complete: kernel=%s, rootfs=%s", kernelPath, rootfsPath)
	} else {
		// Fallback: metadata only (no actual extraction)
		rootfsPath = filepath.Join(imageDir, "rootfs.ext4")
		kernelPath = filepath.Join(imageDir, "vmlinux")
		kernelVersion = "5.10.240"
		c.logger.Warn("No builder available - image metadata only, no layer extraction")
		c.logger.Warn("VMs using this image will fail to boot!")
		c.logger.Debug("  rootfs path: %s (placeholder)", rootfsPath)
		c.logger.Debug("  kernel path: %s (placeholder)", kernelPath)
	}

	// Report progress as complete
	if progressChan != nil {
		progressChan <- &types.PullProgress{
			CurrentBytes: size,
			TotalBytes:   size,
			Percentage:   100,
		}
	}

	// Create image metadata
	image := &types.Image{
		Digest:        digest.String(),
		Tags:          []string{imageRef},
		Architecture:  architecture,
		SizeBytes:     size,
		KernelVersion: kernelVersion,
		RootfsPath:    rootfsPath,
		KernelPath:    kernelPath,
		PulledAt:      time.Now(),
		Labels:        labels,
	}

	totalDuration := time.Since(startTime)
	c.logger.Info("Image pull completed in %v: %s -> %s", totalDuration, imageRef, digest.String())

	return image, nil
}

// ResolveDigest resolves an image reference to a digest
func (c *Client) ResolveDigest(ctx context.Context, imageRef string) (string, error) {
	c.logger.Debug("Resolving digest for: %s", imageRef)

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		c.logger.Error("Failed to parse image reference: %v", err)
		return "", fmt.Errorf("failed to parse image reference: %w", err)
	}

	auth, err := authn.DefaultKeychain.Resolve(ref.Context().Registry)
	if err != nil {
		c.logger.Error("Failed to get authentication: %v", err)
		return "", fmt.Errorf("failed to get authentication: %w", err)
	}

	// Create timeout context
	resolveCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	desc, err := remote.Get(ref, remote.WithAuth(auth), remote.WithContext(resolveCtx))
	if err != nil {
		c.logger.Error("Failed to get image descriptor: %v", err)
		return "", fmt.Errorf("failed to get image descriptor: %w", err)
	}

	c.logger.Debug("Resolved digest: %s", desc.Digest.String())
	return desc.Digest.String(), nil
}
