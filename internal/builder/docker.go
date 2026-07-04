package builder

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

// DockerBuilder extracts OCI images using Docker or Podman.
// This is the recommended builder for development and devcontainer environments.
type DockerBuilder struct {
	// runtime is "docker" or "podman"
	runtime string

	// dataDir is where to store extracted images
	dataDir string

	// verbose enables detailed logging
	verbose bool
}

// dockerInspect represents relevant fields from docker inspect output
type dockerInspect struct {
	ID           string `json:"Id"`
	Architecture string `json:"Architecture"`
	Config       struct {
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	RepoDigests []string `json:"RepoDigests"`
}

// NewDockerBuilder creates a builder that uses Docker or Podman.
func NewDockerBuilder(dataDir string, verbose bool) *DockerBuilder {
	return &DockerBuilder{
		dataDir: dataDir,
		verbose: verbose,
	}
}

// Available checks if Docker or Podman is available.
func (b *DockerBuilder) Available() error {
	// Try docker first
	if _, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command("docker", "info")
		if err := cmd.Run(); err == nil {
			b.runtime = "docker"
			return nil
		}
	}

	// Try podman
	if _, err := exec.LookPath("podman"); err == nil {
		cmd := exec.Command("podman", "info")
		if err := cmd.Run(); err == nil {
			b.runtime = "podman"
			return nil
		}
	}

	return fmt.Errorf("neither docker nor podman available")
}

// Extract pulls and extracts an OCI image to kernel + rootfs.
func (b *DockerBuilder) Extract(ctx context.Context, imageRef string, opts ExtractOptions) (*ExtractResult, error) {
	startTime := time.Now()

	// Ensure runtime is detected
	if b.runtime == "" {
		if err := b.Available(); err != nil {
			return nil, err
		}
	}

	// Apply defaults
	if opts.RootfsSizeMB == 0 {
		opts.RootfsSizeMB = 2048
	}
	if len(opts.KernelSearchPaths) == 0 {
		opts.KernelSearchPaths = DefaultExtractOptions().KernelSearchPaths
	}

	b.log("Using %s to extract %s", b.runtime, imageRef)

	// Step 1: Pull the image
	b.progress(opts, "Pulling image", 10)
	if err := b.pullImage(ctx, imageRef); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Step 2: Inspect image for metadata
	b.progress(opts, "Inspecting image", 20)
	inspect, err := b.inspectImage(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image: %w", err)
	}

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		// Use digest-based directory
		digest := b.extractDigest(inspect)
		if digest == "" {
			digest = fmt.Sprintf("local-%d", time.Now().UnixNano())
		}
		outputDir = filepath.Join(b.dataDir, "images", sanitizeDigest(digest))
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 3: Export container filesystem
	b.progress(opts, "Exporting filesystem", 30)
	tarPath := filepath.Join(outputDir, "rootfs.tar")
	if err := b.exportFilesystem(ctx, imageRef, tarPath); err != nil {
		return nil, fmt.Errorf("failed to export filesystem: %w", err)
	}
	defer os.Remove(tarPath) // Clean up tar after extraction

	// Step 4: Find kernel in the exported tar or use fallback
	b.progress(opts, "Locating kernel", 50)
	kernelPath, kernelVersion, err := b.extractKernel(ctx, tarPath, outputDir, opts.KernelSearchPaths)
	if err != nil {
		// No kernel in the image — fall back to the configured shared kernel.
		if opts.FallbackKernelPath == "" {
			return nil, fmt.Errorf("failed to extract kernel (no fallback configured): %w", err)
		}
		if fbErr := validateFallbackKernel(opts.FallbackKernelPath); fbErr != nil {
			return nil, fmt.Errorf("kernel not found in image and configured fallback kernel %q is unusable: %w", opts.FallbackKernelPath, fbErr)
		}
		kernelPath = opts.FallbackKernelPath
		kernelVersion = extractVersionFromPath(opts.FallbackKernelPath)
		b.log("Using fallback kernel: %s (version: %s)", kernelPath, kernelVersion)
	}

	// Step 5: Create ext4 rootfs
	b.progress(opts, "Creating ext4 rootfs", 60)
	rootfsPath := filepath.Join(outputDir, "rootfs.ext4")
	if err := b.createRootfs(ctx, tarPath, rootfsPath, opts.RootfsSizeMB); err != nil {
		return nil, fmt.Errorf("failed to create rootfs: %w", err)
	}

	// Step 6: Calculate total size
	b.progress(opts, "Finalizing", 90)
	var totalSize int64
	if info, err := os.Stat(kernelPath); err == nil {
		totalSize += info.Size()
	}
	if info, err := os.Stat(rootfsPath); err == nil {
		totalSize += info.Size()
	}

	b.progress(opts, "Complete", 100)

	result := &ExtractResult{
		KernelPath:    kernelPath,
		RootfsPath:    rootfsPath,
		Digest:        b.extractDigest(inspect),
		Architecture:  inspect.Architecture,
		Labels:        inspect.Config.Labels,
		KernelVersion: kernelVersion,
		SizeBytes:     totalSize,
		Duration:      time.Since(startTime),
	}

	b.log("Extraction complete in %v", result.Duration)
	b.log("  Kernel: %s", kernelPath)
	b.log("  Rootfs: %s", rootfsPath)

	return result, nil
}

// validateFallbackKernel ensures the configured fallback kernel is a readable
// regular file, so a misconfigured kernel_path (a directory, or an unreadable
// file) fails here with a clear message instead of later at VM start.
func validateFallbackKernel(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}
	f, err := os.Open(path) //nolint:gosec // operator-configured fallback kernel path
	if err != nil {
		return err
	}
	return f.Close()
}

// pullImage pulls a container image.
func (b *DockerBuilder) pullImage(ctx context.Context, imageRef string) error {
	cmd := exec.CommandContext(ctx, b.runtime, "pull", imageRef) //nolint:gosec // runtime validated in Available(), imageRef from user input is expected
	if b.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// inspectImage gets image metadata.
func (b *DockerBuilder) inspectImage(ctx context.Context, imageRef string) (*dockerInspect, error) {
	cmd := exec.CommandContext(ctx, b.runtime, "inspect", imageRef) //nolint:gosec // runtime validated, imageRef from trusted source
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspects []dockerInspect
	if err := json.Unmarshal(output, &inspects); err != nil {
		return nil, err
	}

	if len(inspects) == 0 {
		return nil, fmt.Errorf("no inspect data returned")
	}

	return &inspects[0], nil
}

// exportFilesystem exports a container's filesystem to a tar file.
func (b *DockerBuilder) exportFilesystem(ctx context.Context, imageRef, tarPath string) error {
	// Create a temporary container
	containerName := fmt.Sprintf("nanofuse-extract-%d", time.Now().UnixNano())

	// Create container (don't start it)
	createCmd := exec.CommandContext(ctx, b.runtime, "create", "--name", containerName, imageRef) //nolint:gosec // runtime validated, containerName is timestamp-based
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Ensure cleanup
	defer func() {
		rmCmd := exec.Command(b.runtime, "rm", "-f", containerName) //nolint:gosec // runtime is validated in Available()
		_ = rmCmd.Run()                                             // Best effort cleanup
	}()

	// Export filesystem
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	exportCmd := exec.CommandContext(ctx, b.runtime, "export", containerName) //nolint:gosec // runtime validated, containerName is timestamp-based
	exportCmd.Stdout = tarFile
	if err := exportCmd.Run(); err != nil {
		return fmt.Errorf("failed to export container: %w", err)
	}

	return nil
}

// extractKernel finds and extracts the kernel from the tar archive.
func (b *DockerBuilder) extractKernel(ctx context.Context, tarPath, outputDir string, searchPaths []string) (string, string, error) {
	kernelDest := filepath.Join(outputDir, "vmlinux")

	// Try each search path
	for _, searchPath := range searchPaths {
		// Handle glob patterns
		if strings.Contains(searchPath, "*") {
			// List matching files in tar
			pattern := strings.TrimPrefix(searchPath, "/")
			listCmd := exec.CommandContext(ctx, "tar", "-tf", tarPath)
			output, err := listCmd.Output()
			if err != nil {
				continue
			}

			// Find matching files
			basePattern := filepath.Base(pattern)
			dirPattern := filepath.Dir(pattern)

			for _, line := range strings.Split(string(output), "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// Check if this file matches our pattern
				if filepath.Dir(line) == dirPattern || filepath.Dir(line) == dirPattern+"/" {
					matched, _ := filepath.Match(basePattern, filepath.Base(line))
					if matched {
						// Extract this kernel
						if err := b.extractFileFromTar(ctx, tarPath, line, kernelDest); err == nil {
							version := extractVersionFromPath(line)
							b.log("Found kernel: %s (version: %s)", line, version)
							return kernelDest, version, nil
						}
					}
				}
			}
		} else {
			// Exact path
			tarPath := strings.TrimPrefix(searchPath, "/")
			if err := b.extractFileFromTar(ctx, tarPath, tarPath, kernelDest); err == nil {
				version := extractVersionFromPath(searchPath)
				b.log("Found kernel: %s", searchPath)
				return kernelDest, version, nil
			}
		}
	}

	return "", "", fmt.Errorf("kernel not found in any of: %v", searchPaths)
}

// extractFileFromTar extracts a single file from a tar archive.
func (b *DockerBuilder) extractFileFromTar(ctx context.Context, tarPath, srcPath, destPath string) error {
	// Create temp dir for extraction
	tempDir, err := os.MkdirTemp("", "kernel-extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Extract the specific file
	cmd := exec.CommandContext(ctx, "tar", "-xf", tarPath, "-C", tempDir, srcPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Move to destination
	extractedPath := filepath.Join(tempDir, srcPath)
	if _, err := os.Stat(extractedPath); err != nil {
		return err
	}

	// Copy file (move might fail across filesystems)
	return copyFile(extractedPath, destPath)
}

// createRootfs creates an ext4 filesystem from the tar.
func (b *DockerBuilder) createRootfs(ctx context.Context, tarPath, rootfsPath string, sizeMB int) error {
	// Check if we can use fuse or need root
	if os.Geteuid() != 0 {
		// Try fuse-ext2 approach
		if _, err := exec.LookPath("fuse2fs"); err == nil {
			return b.createRootfsFuse(ctx, tarPath, rootfsPath, sizeMB)
		}
		return fmt.Errorf("rootfs creation requires root or fuse2fs")
	}

	// Root path: direct mount
	return b.createRootfsMount(ctx, tarPath, rootfsPath, sizeMB)
}

// createRootfsMount creates rootfs using loop mount (requires root).
func (b *DockerBuilder) createRootfsMount(ctx context.Context, tarPath, rootfsPath string, sizeMB int) error {
	// Create sparse file
	ddCmd := exec.CommandContext(ctx, "dd", "if=/dev/zero", fmt.Sprintf("of=%s", rootfsPath), //nolint:gosec // rootfsPath and sizeMB validated by caller
		"bs=1M", fmt.Sprintf("count=%d", sizeMB))
	if err := ddCmd.Run(); err != nil {
		return fmt.Errorf("dd failed: %w", err)
	}

	// Format as ext4
	mkfsCmd := exec.CommandContext(ctx, "mkfs.ext4", "-F", rootfsPath)
	if err := mkfsCmd.Run(); err != nil {
		return fmt.Errorf("mkfs.ext4 failed: %w", err)
	}

	// Create mount point
	mountPoint, err := os.MkdirTemp("", "rootfs-mount-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mountPoint)

	// Mount
	mountCmd := exec.CommandContext(ctx, "mount", "-o", "loop", rootfsPath, mountPoint)
	if err := mountCmd.Run(); err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}

	// Ensure unmount
	defer func() {
		umountCmd := exec.Command("umount", mountPoint) //nolint:gosec // mountPoint is from MkdirTemp
		_ = umountCmd.Run()                             // Best effort cleanup
	}()

	// Extract tar into mounted filesystem.
	// --numeric-owner avoids failures mapping uid/gid names that don't exist on
	// the host. Restore only the security.capability xattr (file capabilities) so
	// binaries like ping keep their caps — this pairs with CAP_SETFCAP granted to
	// the daemon. Limiting the include pattern avoids importing arbitrary/SELinux
	// xattrs from an untrusted rootfs. Capture combined output so failures are
	// actionable.
	tarCmd := exec.CommandContext(ctx, "tar", "--numeric-owner",
		"--xattrs", "--xattrs-include=security.capability", "-xf", tarPath, "-C", mountPoint)
	if out, err := tarCmd.CombinedOutput(); err != nil {
		if msg := strings.TrimSpace(string(out)); msg != "" {
			return fmt.Errorf("tar extraction failed: %w: %s", err, msg)
		}
		return fmt.Errorf("tar extraction failed: %w", err)
	}

	return nil
}

// createRootfsFuse creates rootfs using fuse2fs (unprivileged).
func (b *DockerBuilder) createRootfsFuse(ctx context.Context, tarPath, rootfsPath string, sizeMB int) error {
	// Create sparse file
	ddCmd := exec.CommandContext(ctx, "dd", "if=/dev/zero", fmt.Sprintf("of=%s", rootfsPath), //nolint:gosec // rootfsPath and sizeMB validated by caller
		"bs=1M", fmt.Sprintf("count=%d", sizeMB))
	if err := ddCmd.Run(); err != nil {
		return fmt.Errorf("dd failed: %w", err)
	}

	// Format as ext4
	mkfsCmd := exec.CommandContext(ctx, "mkfs.ext4", "-F", rootfsPath)
	if err := mkfsCmd.Run(); err != nil {
		return fmt.Errorf("mkfs.ext4 failed: %w", err)
	}

	// Create mount point
	mountPoint, err := os.MkdirTemp("", "rootfs-fuse-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mountPoint)

	// Mount with fuse2fs
	fuseCmd := exec.CommandContext(ctx, "fuse2fs", "-o", "rw", rootfsPath, mountPoint)
	if err := fuseCmd.Start(); err != nil {
		return fmt.Errorf("fuse2fs failed: %w", err)
	}

	// Give it a moment to mount
	time.Sleep(500 * time.Millisecond)

	// Ensure unmount
	defer func() {
		umountCmd := exec.Command("fusermount", "-u", mountPoint) //nolint:gosec // mountPoint is from MkdirTemp
		_ = umountCmd.Run()                                       // Best effort cleanup
	}()

	// Extract tar
	tarCmd := exec.CommandContext(ctx, "tar", "-xf", tarPath, "-C", mountPoint)
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("tar extraction failed: %w", err)
	}

	return nil
}

// Helper functions

func (b *DockerBuilder) log(format string, args ...interface{}) {
	if b.verbose {
		fmt.Printf("[DockerBuilder] "+format+"\n", args...)
	}
}

func (b *DockerBuilder) progress(opts ExtractOptions, stage string, percent int) {
	if opts.OnProgress != nil {
		opts.OnProgress(stage, percent)
	}
	b.log("%s (%d%%)", stage, percent)
}

func (b *DockerBuilder) extractDigest(inspect *dockerInspect) string {
	if len(inspect.RepoDigests) > 0 {
		// Format: repo@sha256:xxx
		parts := strings.Split(inspect.RepoDigests[0], "@")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	// Fallback to image ID
	return inspect.ID
}

func sanitizeDigest(digest string) string {
	// sha256:abc123 -> sha256-abc123
	return strings.ReplaceAll(digest, ":", "-")
}

func extractVersionFromPath(path string) string {
	// /boot/vmlinux-5.10.240 -> 5.10.240
	base := filepath.Base(path)
	if strings.HasPrefix(base, "vmlinux-") {
		return strings.TrimPrefix(base, "vmlinux-")
	}
	if strings.HasPrefix(base, "vmlinuz-") {
		return strings.TrimPrefix(base, "vmlinuz-")
	}
	return "unknown"
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644) //nolint:gosec // kernel files need to be world-readable for VM boot
}
