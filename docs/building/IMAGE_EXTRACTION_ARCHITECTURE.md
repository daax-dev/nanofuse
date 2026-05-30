# Image Extraction Architecture

## Problem Statement

The registry client (`internal/registry/client.go`) fetches OCI image metadata but never extracts the actual layers. It sets placeholder paths that don't exist, causing Firecracker to fail with "kernel file not found".

## Current State

```
┌─────────────────────────────────────────────────────────────────┐
│                     Image Sources                                │
├─────────────────────┬───────────────────────────────────────────┤
│ Local Registration  │ OCI Registry Pull                         │
│ (WORKS)             │ (BROKEN - placeholder paths)              │
├─────────────────────┼───────────────────────────────────────────┤
│ register-local-image│ registry/client.go PullImage()            │
│ → real file paths   │ → placeholder paths → kernel not found    │
└─────────────────────┴───────────────────────────────────────────┘
```

## Design Principle: Separation of Concerns

**Running MicroVMs** (nanofuse core) should be isolated from **Building MicroVMs** (image/layer processing).

### Runtime Layer (nanofuse)
- VM lifecycle management
- Network management
- Storage management
- Image registry (references to pre-built artifacts)

### Build Layer (future: provenance project)
- OCI layer fetching
- Layer composition
- Rootfs generation (ext4 creation)
- Kernel extraction

## Solution Options

### Option A: Docker-based Extraction (Recommended)

Use Docker/Podman to extract OCI layers (works in devcontainers):

```go
// 1. Pull container image
docker pull ghcr.io/daax-dev/nanofuse/base:latest

// 2. Export filesystem
docker create --name temp-extract <image>
docker export temp-extract > rootfs.tar
docker rm temp-extract

// 3. Convert to ext4 (requires privileged or fuse)
dd if=/dev/zero of=rootfs.ext4 bs=1M count=2048
mkfs.ext4 rootfs.ext4
mount -o loop rootfs.ext4 /mnt
tar -xf rootfs.tar -C /mnt
umount /mnt

// 4. Extract kernel from known location
// e.g., /boot/vmlinux in the container
```

**Pros:**
- Works with Docker & devcontainers
- No direct OCI layer manipulation
- Reuses existing container tooling

**Cons:**
- Requires Docker/Podman installed
- Privileged for ext4 mounting (or use fuse-ext2)

### Option B: go-containerregistry Direct Extraction

Use the existing go-containerregistry library to download and extract layers:

```go
// Already have:
img, _ := remote.Image(ref, remote.WithAuth(auth))
layers, _ := img.Layers()

// Need to add:
for _, layer := range layers {
    rc, _ := layer.Compressed()  // or Uncompressed()
    // Extract tar to staging directory
}
```

**Pros:**
- Pure Go, no external dependencies
- Already partially implemented

**Cons:**
- Need to handle layer application order
- Need ext4 creation (requires root or fuse)

### Option C: Hybrid Approach

Use go-containerregistry for metadata + Docker for extraction:

1. Validate image exists and get metadata (current code)
2. Shell out to Docker for actual extraction
3. Register extracted artifacts

## Recommended Implementation

### Phase 1: Quick Fix (Docker-based)

Add to `registry/client.go`:

```go
func (c *Client) extractWithDocker(ctx context.Context, imageRef, outputDir string) error {
    // 1. docker pull
    // 2. docker create + export
    // 3. mkfs.ext4 + mount + extract (or fuse-ext2)
    // 4. locate kernel in /boot/
}
```

### Phase 2: Abstract the Builder Interface

```go
// internal/builder/interface.go
type ImageBuilder interface {
    // Extract OCI image to kernel + rootfs
    Extract(ctx context.Context, imageRef string) (*ExtractResult, error)
}

type ExtractResult struct {
    KernelPath string
    RootfsPath string
    Manifest   *BuildManifest
}

// Implementations:
// - DockerBuilder (uses docker export)
// - NativeBuilder (pure Go, needs root)
// - FuseBuilder (uses fuse-ext2, unprivileged)
```

### Phase 3: Move to Provenance

The entire `internal/builder/` and `internal/layerbuild/` packages move to the `provenance` project. Nanofuse only deals with pre-extracted images.

## File Locations

```
nanofuse/
├── internal/
│   ├── registry/
│   │   └── client.go      # OCI metadata + download
│   ├── builder/           # NEW: extraction logic
│   │   ├── interface.go   # Builder interface
│   │   ├── docker.go      # Docker-based extraction
│   │   └── native.go      # Pure Go extraction (future)
│   ├── layerbuild/        # Existing layer composition
│   └── ...
└── test/
    └── fixtures/
        └── debug-kernel/  # Known-good fallback for testing
            ├── vmlinux.bin
            └── rootfs.ext4
```

## Compatibility Matrix

| Environment      | Docker Available | Root Access | Recommended Builder |
|-----------------|------------------|-------------|---------------------|
| Local dev       | Yes              | Yes         | DockerBuilder       |
| Devcontainer    | Yes (DinD/socket)| Maybe       | DockerBuilder       |
| CI/CD           | Yes              | Varies      | DockerBuilder       |
| Production      | No               | Yes         | NativeBuilder       |

## Next Steps

1. [ ] Verify debug kernel boots (`./scripts/debug-boot.sh`)
2. [ ] Register debug kernel (`sudo ./scripts/register-debug-kernel.sh`)
3. [ ] Implement DockerBuilder for extraction
4. [ ] Wire into registry/client.go PullImage
5. [ ] Add tests for Docker-based extraction
6. [ ] Plan provenance project migration
