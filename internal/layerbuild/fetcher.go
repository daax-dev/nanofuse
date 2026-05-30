package layerbuild

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// FetchResult contains the result of a layer fetch operation.
type FetchResult struct {
	// TarballPath is the path to the fetched/created tarball
	TarballPath string

	// Digest is the SHA256 digest of the tarball
	Digest string

	// Metadata is the parsed layer.yaml metadata
	Metadata *LayerPackage

	// SizeBytes is the size of the tarball
	SizeBytes int64
}

// ProgressCallback is called during large fetches.
// Parameters: current bytes processed, total bytes
type ProgressCallback func(current, total int64)

// LocalFetcher fetches layers from local directories or tarballs.
type LocalFetcher struct {
	workDir          string
	progressCallback ProgressCallback
}

// NewLocalFetcher creates a new LocalFetcher with the given work directory.
func NewLocalFetcher(workDir string) *LocalFetcher {
	return &LocalFetcher{
		workDir: workDir,
	}
}

// SetProgressCallback sets the progress callback function.
func (f *LocalFetcher) SetProgressCallback(callback ProgressCallback) {
	f.progressCallback = callback
}

// Supports returns true if this fetcher handles the given source type.
func (f *LocalFetcher) Supports(sourceType SourceType) bool {
	return sourceType == SourceTypeLocal
}

// Fetch retrieves a layer from a local source (directory or tarball).
func (f *LocalFetcher) Fetch(source string) (*CachedLayer, error) {
	// Parse source URL
	if !strings.HasPrefix(source, "local://") {
		return nil, fmt.Errorf("invalid local source URL: %s (must start with local://)", source)
	}

	localPath := strings.TrimPrefix(source, "local://")
	if localPath == "" {
		return nil, fmt.Errorf("empty local path in source: %s", source)
	}

	// Check if path exists
	info, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access local path %s: %w", localPath, err)
	}

	// Handle directory vs tarball
	var result *FetchResult
	if info.IsDir() {
		result, err = f.fetchFromDirectory(localPath)
	} else if isTarball(localPath) {
		result, err = f.fetchFromTarball(localPath)
	} else {
		return nil, fmt.Errorf("local path %s is not a directory or tarball (.tar.gz)", localPath)
	}

	if err != nil {
		return nil, err
	}

	return result.ToCachedLayer(source), nil
}

// fetchFromDirectory creates a tarball from a layer directory.
func (f *LocalFetcher) fetchFromDirectory(dirPath string) (*FetchResult, error) {
	// Read layer.yaml
	layerYAMLPath := filepath.Join(dirPath, "layer.yaml")
	metadata, err := readLayerMetadata(layerYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read layer metadata from %s: %w", layerYAMLPath, err)
	}

	// Create output tarball path
	tarballPath := filepath.Join(f.workDir, fmt.Sprintf("%s-%s.tar.gz", metadata.Name, metadata.Version))

	// Create tarball from directory
	if err := f.createTarball(dirPath, tarballPath); err != nil {
		return nil, fmt.Errorf("failed to create tarball: %w", err)
	}

	// Calculate digest
	digest, size, err := calculateDigest(tarballPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}

	return &FetchResult{
		TarballPath: tarballPath,
		Digest:      digest,
		Metadata:    metadata,
		SizeBytes:   size,
	}, nil
}

// fetchFromTarball extracts and validates an existing tarball.
func (f *LocalFetcher) fetchFromTarball(tarballPath string) (*FetchResult, error) {
	// Extract to temporary directory to read layer.yaml
	tempDir := filepath.Join(f.workDir, "extract-"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract tarball
	if err := extractTarball(tarballPath, tempDir); err != nil {
		return nil, fmt.Errorf("failed to extract tarball: %w", err)
	}

	// Read layer.yaml from extracted content
	layerYAMLPath := filepath.Join(tempDir, "layer.yaml")
	metadata, err := readLayerMetadata(layerYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read layer metadata: %w", err)
	}

	// Calculate digest and size
	digest, size, err := calculateDigest(tarballPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}

	return &FetchResult{
		TarballPath: tarballPath,
		Digest:      digest,
		Metadata:    metadata,
		SizeBytes:   size,
	}, nil
}

// createTarball creates a gzipped tarball from a directory.
func (f *LocalFetcher) createTarball(srcDir, destPath string) error {
	// Create output file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzw := gzip.NewWriter(outFile)
	defer gzw.Close()

	// Create tar writer
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Get total size for progress reporting
	totalSize := int64(0)
	currentSize := int64(0)

	// Calculate total size if we have a progress callback
	if f.progressCallback != nil {
		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to calculate total size: %w", err)
		}
	}

	// Walk directory and add files to tarball
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = relPath

		// Write header
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		// If it's a file, write contents
		if !info.IsDir() {
			// path comes from Walk over a controlled local build-context dir;
			// os.Root would reject the symlinks that rootfs layers legitimately contain.
			srcFile, err := os.Open(path) //nolint:gosec // controlled build context; symlinks are intentional
			if err != nil {
				return err
			}

			written, err := io.Copy(tw, srcFile)
			closeErr := srcFile.Close() // Close immediately, not defer (avoid FD leak in loop)
			if err != nil {
				return err
			}
			if closeErr != nil {
				return fmt.Errorf("failed to close source file %s: %w", path, closeErr)
			}

			// Report progress for large layers (>100MB)
			if f.progressCallback != nil && totalSize > 100*1024*1024 {
				currentSize += written
				f.progressCallback(currentSize, totalSize)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	return nil
}

// DockerFetcher fetches layers from Docker images.
type DockerFetcher struct {
	workDir string
}

// NewDockerFetcher creates a new DockerFetcher with the given work directory.
func NewDockerFetcher(workDir string) *DockerFetcher {
	return &DockerFetcher{
		workDir: workDir,
	}
}

// Supports returns true if this fetcher handles the given source type.
func (f *DockerFetcher) Supports(sourceType SourceType) bool {
	return sourceType == SourceTypeDocker
}

// Fetch retrieves a layer from a Docker image.
func (f *DockerFetcher) Fetch(source string) (*CachedLayer, error) {
	// Parse source URL
	if !strings.HasPrefix(source, "docker://") {
		return nil, fmt.Errorf("invalid docker source URL: %s (must start with docker://)", source)
	}

	imageName := strings.TrimPrefix(source, "docker://")
	if imageName == "" {
		return nil, fmt.Errorf("empty image name in source: %s", source)
	}

	// Generate unique container name
	containerName := fmt.Sprintf("nanofuse-layer-%s", uuid.New().String())

	// Create container (without starting it)
	createCmd := exec.Command("docker", "create", "--name", containerName, imageName)
	if output, err := createCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create container from %s: %w (output: %s)", imageName, err, string(output))
	}

	// Ensure cleanup happens
	defer func() {
		rmCmd := exec.Command("docker", "rm", containerName)
		_ = rmCmd.Run() // Ignore errors during cleanup
	}()

	// Export container filesystem to tarball
	tarballPath := filepath.Join(f.workDir, containerName+".tar")
	exportCmd := exec.Command("docker", "export", containerName, "-o", tarballPath)
	if output, err := exportCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to export container %s: %w (output: %s)", containerName, err, string(output))
	}

	// Extract to temporary directory to read layer.yaml
	tempDir := filepath.Join(f.workDir, "extract-"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract tarball to read layer.yaml
	if err := extractTarball(tarballPath, tempDir); err != nil {
		return nil, fmt.Errorf("failed to extract tarball: %w", err)
	}

	// Read layer.yaml from extracted content
	layerYAMLPath := filepath.Join(tempDir, "layer.yaml")
	metadata, err := readLayerMetadata(layerYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read layer metadata: %w", err)
	}

	// Calculate digest
	digest, size, err := calculateDigest(tarballPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}

	result := &FetchResult{
		TarballPath: tarballPath,
		Digest:      digest,
		Metadata:    metadata,
		SizeBytes:   size,
	}

	return result.ToCachedLayer(source), nil
}

// RegistryFetcher fetches layers from OCI registries (ghcr.io, Docker Hub, etc.).
type RegistryFetcher struct {
	workDir          string
	keychain         authn.Keychain
	maxRetries       int
	baseBackoff      time.Duration
	progressCallback ProgressCallback
}

// RegistryFetcherOption configures a RegistryFetcher.
type RegistryFetcherOption func(*RegistryFetcher)

// WithKeychain sets a custom keychain for the registry fetcher.
func WithKeychain(keychain authn.Keychain) RegistryFetcherOption {
	return func(f *RegistryFetcher) {
		f.keychain = keychain
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) RegistryFetcherOption {
	return func(f *RegistryFetcher) {
		f.maxRetries = n
	}
}

// WithBaseBackoff sets the base backoff duration for retries.
func WithBaseBackoff(d time.Duration) RegistryFetcherOption {
	return func(f *RegistryFetcher) {
		f.baseBackoff = d
	}
}

// NewRegistryFetcher creates a new RegistryFetcher with the given work directory.
func NewRegistryFetcher(workDir string, opts ...RegistryFetcherOption) *RegistryFetcher {
	f := &RegistryFetcher{
		workDir:     workDir,
		maxRetries:  3,
		baseBackoff: 1 * time.Second,
	}

	for _, opt := range opts {
		opt(f)
	}

	// Default to using docker config for authentication
	if f.keychain == nil {
		f.keychain = authn.DefaultKeychain
	}

	return f
}

// SetProgressCallback sets the progress callback function.
func (f *RegistryFetcher) SetProgressCallback(callback ProgressCallback) {
	f.progressCallback = callback
}

// Supports returns true if this fetcher handles the given source type.
func (f *RegistryFetcher) Supports(sourceType SourceType) bool {
	return sourceType == SourceTypeRegistry
}

// Fetch retrieves a layer from an OCI registry.
// Source format: registry://ghcr.io/namespace/image:tag or registry://docker.io/library/alpine:latest
//
//nolint:gocyclo // Complex but well-structured registry fetch with proper error handling
func (f *RegistryFetcher) Fetch(source string) (*CachedLayer, error) {
	// Parse source URL
	if !strings.HasPrefix(source, "registry://") {
		return nil, fmt.Errorf("invalid registry source URL: %s (must start with registry://)", source)
	}

	imageRef := strings.TrimPrefix(source, "registry://")
	if imageRef == "" {
		return nil, fmt.Errorf("empty image reference in source: %s", source)
	}

	// Parse the image reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference %q: %w", imageRef, err)
	}

	// Fetch the image with retry logic
	img, err := f.fetchWithRetry(ref)
	if err != nil {
		return nil, err
	}

	// Get image digest for verification
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to get image digest: %w", err)
	}

	// Export image layers to a tarball
	tarballPath := filepath.Join(f.workDir, fmt.Sprintf("registry-%s.tar.gz", uuid.New().String()))

	// Create the tarball file
	outFile, err := os.Create(tarballPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	// Create gzip writer
	gzw := gzip.NewWriter(outFile)
	tw := tar.NewWriter(gzw)

	// Get the layers
	layers, err := img.Layers()
	if err != nil {
		tw.Close()
		gzw.Close()
		outFile.Close()
		os.Remove(tarballPath)
		return nil, fmt.Errorf("failed to get image layers: %w", err)
	}

	// Extract each layer into the tarball
	for i, layer := range layers {
		layerReader, err := layer.Uncompressed()
		if err != nil {
			tw.Close()
			gzw.Close()
			outFile.Close()
			os.Remove(tarballPath)
			return nil, fmt.Errorf("failed to get layer %d content: %w", i, err)
		}

		// Copy layer contents into our tarball
		layerTR := tar.NewReader(layerReader)
		for {
			header, err := layerTR.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				layerReader.Close()
				tw.Close()
				gzw.Close()
				outFile.Close()
				os.Remove(tarballPath)
				return nil, fmt.Errorf("failed to read layer %d entry: %w", i, err)
			}

			// Write the header
			if err := tw.WriteHeader(header); err != nil {
				layerReader.Close()
				tw.Close()
				gzw.Close()
				outFile.Close()
				os.Remove(tarballPath)
				return nil, fmt.Errorf("failed to write header for %s: %w", header.Name, err)
			}

			// Write the content for regular files
			if header.Typeflag == tar.TypeReg {
				// #nosec G110 -- Layer content is from trusted OCI registry with verified digest
				if _, err := io.Copy(tw, layerTR); err != nil {
					layerReader.Close()
					tw.Close()
					gzw.Close()
					outFile.Close()
					os.Remove(tarballPath)
					return nil, fmt.Errorf("failed to write content for %s: %w", header.Name, err)
				}
			}
		}
		layerReader.Close()
	}

	// Close writers in correct order
	if err := tw.Close(); err != nil {
		gzw.Close()
		outFile.Close()
		os.Remove(tarballPath)
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzw.Close(); err != nil {
		outFile.Close()
		os.Remove(tarballPath)
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	if err := outFile.Close(); err != nil {
		os.Remove(tarballPath)
		return nil, fmt.Errorf("failed to close output file: %w", err)
	}

	// Extract to temporary directory to read layer.yaml
	tempDir := filepath.Join(f.workDir, "extract-"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		os.Remove(tarballPath)
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract tarball to read layer.yaml
	if err := extractTarball(tarballPath, tempDir); err != nil {
		os.Remove(tarballPath)
		return nil, fmt.Errorf("failed to extract tarball: %w", err)
	}

	// Read layer.yaml from extracted content
	layerYAMLPath := filepath.Join(tempDir, "layer.yaml")
	metadata, err := readLayerMetadata(layerYAMLPath)
	if err != nil {
		// If no layer.yaml, create synthetic metadata from image reference
		metadata = &LayerPackage{
			Name:        ref.Context().RepositoryStr(),
			Version:     ref.Identifier(),
			Type:        LayerTypeBase,
			Description: fmt.Sprintf("Layer from OCI registry: %s", imageRef),
		}
	}

	// Calculate digest of the tarball
	tarballDigest, size, err := calculateDigest(tarballPath)
	if err != nil {
		os.Remove(tarballPath)
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}

	// Store image digest in metadata for verification
	if metadata.Metadata == nil {
		metadata.Metadata = make(map[string]string)
	}
	metadata.Metadata["oci-digest"] = digest.String()

	result := &FetchResult{
		TarballPath: tarballPath,
		Digest:      tarballDigest,
		Metadata:    metadata,
		SizeBytes:   size,
	}

	return result.ToCachedLayer(source), nil
}

// fetchWithRetry fetches an image with exponential backoff on rate limits.
func (f *RegistryFetcher) fetchWithRetry(ref name.Reference) (v1.Image, error) {
	var lastErr error
	backoff := f.baseBackoff

	for attempt := 0; attempt <= f.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		// Build remote options
		opts := []remote.Option{
			remote.WithContext(context.Background()),
		}

		// Add authentication from keychain
		if f.keychain != nil {
			opts = append(opts, remote.WithAuthFromKeychain(f.keychain))
		}

		img, err := remote.Image(ref, opts...)
		if err == nil {
			return img, nil
		}

		lastErr = err

		// Check if this is a rate limit error (429)
		if isRateLimitError(err) {
			continue // Retry with backoff
		}

		// For other errors, wrap with clear message and return immediately
		return nil, wrapRegistryError(err, ref.String())
	}

	return nil, fmt.Errorf("failed to fetch image %s after %d retries: %w", ref.String(), f.maxRetries, lastErr)
}

// isRateLimitError checks if the error is a rate limit (HTTP 429) error.
func isRateLimitError(err error) bool {
	var transportErr *transport.Error
	if errors.As(err, &transportErr) {
		// Check error codes in the diagnostics
		for _, diag := range transportErr.Errors {
			if diag.Code == transport.TooManyRequestsErrorCode {
				return true
			}
		}
		// Also check HTTP status code directly
		if transportErr.StatusCode == http.StatusTooManyRequests {
			return true
		}
	}
	return false
}

// wrapRegistryError wraps registry errors with clear, actionable messages.
func wrapRegistryError(err error, imageRef string) error {
	var transportErr *transport.Error
	if errors.As(err, &transportErr) {
		switch transportErr.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("authentication required for %s: configure docker credentials in ~/.docker/config.json or use 'docker login': %w", imageRef, err)
		case http.StatusForbidden:
			return fmt.Errorf("access denied for %s: check your permissions or ensure the image exists: %w", imageRef, err)
		case http.StatusNotFound:
			return fmt.Errorf("image not found: %s: verify the image name and tag are correct: %w", imageRef, err)
		case http.StatusTooManyRequests:
			return fmt.Errorf("rate limit exceeded for %s: wait and retry later or authenticate to increase limits: %w", imageRef, err)
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return fmt.Errorf("registry server error for %s: the registry may be experiencing issues, try again later: %w", imageRef, err)
		default:
			return fmt.Errorf("registry error for %s (HTTP %d): %w", imageRef, transportErr.StatusCode, err)
		}
	}

	// Check for DNS/network errors
	if strings.Contains(err.Error(), "no such host") {
		return fmt.Errorf("registry not found: %s: verify the registry hostname is correct: %w", imageRef, err)
	}
	if strings.Contains(err.Error(), "connection refused") {
		return fmt.Errorf("cannot connect to registry for %s: check network connectivity: %w", imageRef, err)
	}
	if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
		return fmt.Errorf("timeout connecting to registry for %s: check network connectivity or try again: %w", imageRef, err)
	}

	return fmt.Errorf("failed to fetch from registry %s: %w", imageRef, err)
}

// Helper functions

// readLayerMetadata reads and parses a layer.yaml file.
func readLayerMetadata(path string) (*LayerPackage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read layer.yaml: %w", err)
	}

	var metadata LayerPackage
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse layer.yaml: %w", err)
	}

	return &metadata, nil
}

// calculateDigest calculates the SHA256 digest of a file and returns it with the file size.
func calculateDigest(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	hash := sha256.New()
	size, err := io.Copy(hash, f)
	if err != nil {
		return "", 0, err
	}

	digest := hex.EncodeToString(hash.Sum(nil))
	return digest, size, nil
}

// verifyDigest verifies that a file matches the expected SHA256 digest.
// If expectedDigest is empty, verification is skipped.
func verifyDigest(path, expectedDigest string) error {
	if expectedDigest == "" {
		return nil // Skip verification if no digest provided
	}

	actualDigest, _, err := calculateDigest(path)
	if err != nil {
		return fmt.Errorf("failed to calculate digest: %w", err)
	}

	if actualDigest != expectedDigest {
		return fmt.Errorf("digest mismatch: expected %s, got %s", expectedDigest, actualDigest)
	}

	return nil
}

// isTarball checks if a file is a tarball based on extension.
func isTarball(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".tar" || ext == ".gz" && strings.HasSuffix(path, ".tar.gz")
}

// extractTarball extracts a tarball to the specified directory.
// This function handles gzip detection, tar entry iteration, directory traversal validation,
// and file extraction - necessarily complex due to security requirements.
//
//nolint:gocyclo // Tar extraction requires many security checks and error handling paths
func extractTarball(tarballPath, destDir string) error {
	f, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to open tarball: %w", err)
	}
	defer f.Close()

	// Create tar reader (with optional gzip decompression)
	var tr *tar.Reader
	if strings.HasSuffix(tarballPath, ".gz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzr.Close()
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(f)
	}

	// Extract files
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Validate and clean header name BEFORE constructing target path
		cleanName := filepath.Clean(hdr.Name)
		if cleanName == "." {
			// Skip top-level pseudo-entry
			continue
		}
		// Reject absolute paths and parent-directory traversals
		if filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid tar entry: %s (absolute or parent-directory path)", hdr.Name)
		}

		// Construct target path using validated, cleaned name
		target := filepath.Join(destDir, cleanName)

		// Double-check for directory traversal based on final target
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != filepath.Clean(destDir) {
			return fmt.Errorf("invalid tar entry: %s (potential directory traversal)", hdr.Name)
		}

		// Safe mode conversion with bounds check (hdr.Mode is int64, we need os.FileMode which is uint32)
		// #nosec G115 -- Mode is masked to 9 bits (0777), which safely fits in uint32
		fileMode := os.FileMode(hdr.Mode & 0777) // Mask to valid permission bits only

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, fileMode); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}

		case tar.TypeReg:
			// Create parent directory
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", target, err)
			}

			// Create file with safe mode
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}

			// Use LimitReader to prevent decompression bombs (max 10GB per file)
			const maxFileSize = 10 * 1024 * 1024 * 1024 // 10GB
			limitedReader := io.LimitReader(tr, maxFileSize)

			if _, err := io.Copy(outFile, limitedReader); err != nil {
				if cerr := outFile.Close(); cerr != nil {
					return fmt.Errorf("failed to write file %s: %v (also failed to close: %w)", target, err, cerr)
				}
				return fmt.Errorf("failed to write file %s: %w", target, err)
			}
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("failed to close file %s: %w", target, err)
			}
		}
	}

	return nil
}

// ConvertToCachedLayer converts a FetchResult to a CachedLayer.
func (fr *FetchResult) ToCachedLayer(sourceURL string) *CachedLayer {
	return &CachedLayer{
		Digest:     fr.Digest,
		Name:       fr.Metadata.Name,
		Version:    fr.Metadata.Version,
		Type:       fr.Metadata.Type,
		SourceURL:  sourceURL,
		LocalPath:  fr.TarballPath,
		SizeBytes:  fr.SizeBytes,
		FetchedAt:  time.Now(),
		LastUsedAt: time.Now(),
		Metadata:   fr.Metadata,
	}
}
