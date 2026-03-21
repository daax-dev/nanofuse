# Build Guide - Simplified

## The Quick Answer

### Want to build the complete image (rootfs + kernel)?
```bash
sudo ./build.sh
```
This is the main build script that creates the full NanoFuse base image.

### Want to test the build?
```bash
sudo ./test-boot.sh
```
Tests that the built image boots successfully in Firecracker.

### Want to validate build artifacts?
```bash
./validate-build.sh
```
Validates that all build artifacts are present and correct.

### Want to clean everything?
```bash
./clean.sh
```
Cleans all build artifacts and temporary files.

**Note**: Several specialized build and test scripts have been archived to `scripts/archives/` as they were used during initial development. The scripts above are the essential ones for normal usage.

## Build Scripts Overview

| Script | Purpose | Requires sudo? | What it does |
|--------|---------|----------------|--------------|
| `build.sh` | **Main build** | Yes (for mounting ext4) | Builds complete image: rootfs + kernel |
| `download-fc-kernel.sh` | Download kernel | No | Downloads pre-built Firecracker kernel |
| `test-boot.sh` | Test build | Yes (for Firecracker) | Tests kernel boots (supports --verbose, --check-virtio) |
| `validate-build.sh` | Validate artifacts | No | Validates build artifacts are present and correct |
| `clean.sh` | Clean all | Auto-detect | Cleans everything (supports --check, --sudo) |

## Typical Workflows

### 1. First time building
```bash
# Build everything (takes ~10 minutes)
sudo ./build.sh

# Test the build
sudo ./test-boot.sh
```

### 2. Validate your build
```bash
# Check all artifacts are present and correct
./validate-build.sh
```

### 3. Something is broken, start fresh
```bash
# Clean everything
./clean.sh

# Then do a full build
sudo ./build.sh
```

### 4. Test an existing build
```bash
# Test kernel (auto-detects paths)
sudo ./test-boot.sh

# Test with verbose output
sudo ./test-boot.sh --verbose

# Test specific kernel and rootfs
sudo ./test-boot.sh /path/to/vmlinux /path/to/rootfs.ext4
```

### 5. Clean everything
```bash
# Auto-detect if sudo is needed
./clean.sh

# Check what would be cleaned (dry-run)
./clean.sh --check

# Force sudo mode
sudo ./clean.sh
```

## Build Outputs

After building, you'll find artifacts in:

```
build/
├── vmlinux           # Kernel (uncompressed ELF)
├── rootfs.ext4       # Root filesystem (ext4 image)
└── manifest.json     # Image metadata

/tmp/
├── vmlinux-test           # From BUILD_AND_TEST.sh
├── vmlinux-fresh-build    # From build-kernel-docker.sh
├── rootfs-working.ext4    # Test rootfs
└── test_boot_config.json  # Firecracker config (temp)
```

## Fixing the Permission Issue

The original error you saw:
```
./TEST_BOOT_VERBOSE.sh: line 28: /tmp/test_boot_config.json: Permission denied
```

This was caused by running with sudo and having permission issues writing to `/tmp`.

**Fixed in this commit** by:
1. Adding `rm -f` before creating the config file
2. Fixing ownership of extracted kernel when using sudo
3. Using `$SUDO_UID` and `$SUDO_GID` to set correct ownership

## What Each Build Script Actually Does

### build.sh (Main Build)
1. Builds Docker image from Dockerfile
2. Exports container filesystem to tar
3. Creates empty ext4 image (2GB)
4. Mounts ext4 and copies filesystem
5. Builds or copies kernel
6. Generates manifest.json
7. **Outputs**: build/rootfs.ext4, build/vmlinux, build/manifest.json

### build-kernel-docker.sh (Kernel Only)
1. Builds Dockerfile.kernel in Docker
2. Extracts /vmlinux from container
3. Saves to build/vmlinux or /tmp/vmlinux-fresh-build
4. **Outputs**: build/vmlinux or /tmp/vmlinux-fresh-build

### BUILD_AND_TEST.sh (Build + Test)
1. Removes old Docker cache
2. Builds kernel using Dockerfile.kernel
3. Extracts to /tmp/vmlinux-test
4. Fixes ownership if using sudo
5. Runs TEST_BOOT_VERBOSE.sh
6. **Outputs**: /tmp/vmlinux-test + test results

### TEST_BOOT_VERBOSE.sh (Test)
1. Creates Firecracker config pointing to kernel and rootfs
2. Boots VM with 35 second timeout
3. Checks for:
   - Linux version message
   - VIRTIO-MMIO device registration
   - VIRTIO_BLK driver loaded
   - Block device [vda] detected
   - EXT4 filesystem mounted
   - No kernel panic
4. **Outputs**: Pass/fail results + /tmp/fc_boot_output.log

## Requirements

- Docker (for building)
- Firecracker (for testing)
- sudo access (for mount operations and Firecracker)
- ~10GB disk space (for kernel source + build artifacts)

## Troubleshooting

### "Permission denied" when creating config
**Fixed** - Update to latest BUILD_AND_TEST.sh and TEST_BOOT_VERBOSE.sh

### "Kernel not found"
```bash
# Check what kernels exist
ls -lh /tmp/vmlinux* build/vmlinux 2>/dev/null
```

### "Rootfs not found"
```bash
# Use build.sh to create rootfs
sudo ./build.sh
```

### Docker build fails
```bash
# Clean Docker cache
./BUILD_CLEAN.sh

# Or manually
docker system prune -af
```

## Next Steps

After building successfully:
1. Test the complete image: `sudo ./build.sh && make test`
2. Integrate with NanoFuse API (coming soon)
3. Create derived images (Trigger.dev web/worker)
