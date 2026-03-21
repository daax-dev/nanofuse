# Build Process - Verified Working

## Overview
This document describes the **verified working** build and test process for the NanoFuse base image kernel.

## Prerequisites
- Docker installed and running
- Firecracker installed (`firecracker` in PATH)
- User in `kvm` group (for running Firecracker without sudo)
- sudo access (for Docker and mounting filesystems)
- ~10GB disk space

## Quick Start - 3 Commands That Work

### Option A: Step by Step
```bash
# Step 1: Build rootfs (needs sudo for mounting ext4)
sudo ./build.sh

# Step 2: Build kernel (needs sudo for Docker)
sudo ./BUILD_KERNEL_ONLY.sh

# Step 3: Test kernel (NO sudo - run as your user)
./TEST_KERNEL_ONLY.sh
```

### Option B: All-in-One
```bash
# Build kernel + test in one command
# Builds as root, automatically drops privileges for testing
sudo ./BUILD_AND_TEST.sh
```

## What Each Script Does

### `build.sh` - Build Complete Image
- **Requires**: sudo (for mounting ext4 filesystem)
- **Duration**: ~5 minutes
- **Output**:
  - `build/rootfs.ext4` (2GB root filesystem)
  - `build/vmlinux` (kernel binary)
  - `build/manifest.json` (image metadata)
- **What it does**:
  1. Builds Docker container with Ubuntu 24.04 + systemd
  2. Exports container filesystem
  3. Creates ext4 image and mounts it
  4. Copies filesystem into ext4 image
  5. Builds or copies kernel
  6. Generates manifest

### `BUILD_KERNEL_ONLY.sh` - Build Kernel Only
- **Requires**: sudo (for Docker)
- **Duration**: ~4 minutes
- **Output**: `/tmp/vmlinux-test`
- **What it does**:
  1. Removes old nanofuse-kernel-builder Docker image
  2. Builds kernel in Docker (Linux 6.1.90 with Firecracker config)
  3. Extracts vmlinux binary to /tmp
  4. Fixes ownership to your user

### `TEST_KERNEL_ONLY.sh` - Test Kernel Only
- **Requires**: NO sudo (must run as your user)
- **Duration**: ~35 seconds
- **Output**: Test results + boot log
- **What it does**:
  1. Auto-detects kernel and rootfs paths
  2. Creates Firecracker config
  3. Boots VM with 35 second timeout
  4. Validates boot sequence:
     - ✅ Firecracker starts
     - ✅ Kernel loads (Linux 6.1.90)
     - ✅ VIRTIO-MMIO device detected
     - ✅ VIRTIO_BLK driver loads
     - ✅ Block device [vda] detected
     - ✅ EXT4 filesystem mounts
     - ✅ No kernel panic

### `BUILD_AND_TEST.sh` - Combined Build + Test
- **Requires**: sudo (drops privileges internally for testing)
- **Duration**: ~4.5 minutes
- **Output**: Kernel + test results
- **What it does**:
  1. Checks it's running with sudo
  2. Builds kernel as root
  3. Fixes ownership to original user
  4. Drops privileges (`exec sudo -u $SUDO_USER`)
  5. Runs test as original user

## Critical: Why Sudo Matters

### ✅ Works
```bash
# Build with sudo, test without sudo
sudo ./BUILD_KERNEL_ONLY.sh
./TEST_KERNEL_ONLY.sh
```

### ❌ Fails
```bash
# Testing with sudo breaks Firecracker output capture
sudo ./TEST_KERNEL_ONLY.sh  # DON'T DO THIS
```

**Root Cause**: When Firecracker runs under sudo, its output redirection behaves differently and only captures 1 line instead of hundreds. The test script detects this and retries with background execution as a workaround, but it's better to just run tests as your user.

## Test Output Example

### Success
```
==========================================
TEST RESULTS
==========================================

✓ Step 1: Firecracker VM started and kernel booted
✓ Step 2: Linux 6.1 kernel loaded
✓ Step 3: VIRTIO-MMIO device detected
✓ Step 4: VIRTIO_BLK driver loaded
✓ Step 5: Block device [vda] detected
✓ Step 6: EXT4 filesystem mounted successfully
✓ Step 7: No kernel panic

==========================================

✓✓✓ ALL TESTS PASSED ✓✓✓
==========================================
```

### Failure
```
==========================================
TEST RESULTS
==========================================

✗ Step 1: Firecracker VM failed to start kernel
✗ Step 2: Kernel version NOT FOUND in log
  The kernel never booted - no version string in output
✗ Step 3: VIRTIO-MMIO device NOT detected
...

==========================================
CRITICAL: Firecracker produced no output. Log has only 1 line.
This means Firecracker started but the VM never booted.

✗✗✗ TESTS FAILED ✗✗✗
==========================================

Boot log: /tmp/fc_boot_output_XXX.log (1 lines)
Config:   /tmp/test_boot_config_XXX.json

To debug:
  1. Check log: cat /tmp/fc_boot_output_XXX.log
  2. Run manually: sudo firecracker --no-api --config-file /tmp/test_boot_config_XXX.json
  3. Re-run verbose: sudo ./test-boot.sh --verbose /tmp/vmlinux-test /tmp/rootfs-working.ext4
```

## Troubleshooting

### "Kernel not found"
```bash
# Check what exists
ls -lh /tmp/vmlinux* build/vmlinux 2>/dev/null

# Build kernel
sudo ./BUILD_KERNEL_ONLY.sh
```

### "Rootfs not found"
```bash
# Check what exists
ls -lh build/rootfs.ext4 /tmp/rootfs-working.ext4 2>/dev/null

# Build rootfs
sudo ./build.sh
```

### "Permission denied" on /tmp/vmlinux-test
```bash
# File owned by root from previous sudo run
sudo rm -f /tmp/vmlinux-test
sudo ./BUILD_KERNEL_ONLY.sh
```

### "Docker build failed"
```bash
# Clean Docker cache
./BUILD_CLEAN.sh

# Or manually
docker system prune -af
```

### Tests fail but manual run works
```bash
# Manual test to verify kernel is good
timeout 5 firecracker --no-api --config-file /tmp/test_boot_config_XXX.json 2>&1 | head -20

# If that works, the kernel is fine - test script has an issue
```

## Clean Everything
```bash
# Remove all build artifacts, Docker images, temp files
./clean.sh

# Check what would be cleaned (dry-run)
./clean.sh --check

# Force clean with sudo (for root-owned files)
sudo ./clean.sh
```

## File Locations

### Build Artifacts
- `build/vmlinux` - Kernel binary (if using build.sh)
- `build/rootfs.ext4` - Root filesystem (2GB)
- `build/manifest.json` - Image metadata

### Temporary Files
- `/tmp/vmlinux-test` - Kernel from BUILD_KERNEL_ONLY.sh
- `/tmp/vmlinux-fresh-build` - Kernel from build-kernel-docker.sh
- `/tmp/rootfs-working.ext4` - Test rootfs (if created)
- `/tmp/fc_boot_output_*.log` - Firecracker boot logs (from tests)
- `/tmp/test_boot_config_*.json` - Firecracker configs (from tests)

## Kernel Configuration

The kernel is built with:
- **Version**: Linux 6.1.90
- **Base Config**: Firecracker's official microVM config
- **Critical Setting**: `CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y`
  - This allows Firecracker to inject virtio devices via kernel command line
  - Without this, the block device won't be detected
  - Boot will fail with "VFS: Unable to mount root fs"

## Next Steps

After successful build and test:
1. ✅ Kernel boots in Firecracker
2. ✅ All VIRTIO devices work
3. ✅ Filesystem mounts
4. ⏭️ Test with actual workload
5. ⏭️ Integrate with NanoFuse API
6. ⏭️ Create GitHub Actions workflow
7. ⏭️ Publish to registry

## References

- Kernel Dockerfile: `Dockerfile.kernel`
- Main build script: `build.sh`
- Test script: `test-boot.sh`
- Clean script: `clean.sh`
- Full guide: `BUILD_GUIDE.md`
- Script inventory: `SCRIPT_INVENTORY.md`
- Consolidation summary: `CONSOLIDATION_SUMMARY.md`
