# NanoFuse Quick Start - Build & Test

## Prerequisites

- Docker installed and running
- Firecracker installed (`firecracker --version`)
- sudo access (for Docker and filesystem operations)
- At least 30GB free disk space (for kernel build and VM images)

## Two-Script System

This repository uses **two independent scripts** that should always be run together:

### 1. Build Script (`build-complete.sh`)
**Purpose:** Builds everything needed for NanoFuse VMs
**Time:** ~20-30 minutes (kernel build takes time)

```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
sudo ./build-complete.sh
```

**What it does:**
- STEP 1: Cleans old build artifacts
- STEP 2: Builds Linux 6.1.90 kernel (Firecracker optimized)
- STEP 3: Builds Docker image (nanofuse-base:latest)
- STEP 4: Exports filesystem from Docker image
- STEP 5: Creates ext4 rootfs image (2GB)
- STEP 6: Copies kernel to build directory
- STEP 7: Generates manifest.json metadata
- STEP 8: Verifies all artifacts

**Outputs:**
- `./build/vmlinux` - Linux kernel (39MB ELF binary)
- `./build/rootfs.ext4` - Filesystem image (2GB)
- `./build/manifest.json` - Image metadata

### 2. Test Script (`test-complete.sh`)
**Purpose:** Tests the built system end-to-end
**Time:** ~3-5 minutes

```bash
sudo /home/jpoley/src/_mine/nanofuse/test-complete.sh
```

**What it does:**
- STEP 1: Verifies all prerequisites (Docker, Firecracker, binaries)
- STEP 2: Stops any existing daemons/VMs
- STEP 3: Builds Docker image
- STEP 4: Starts nanofused daemon on port 8080
- **STEP 5:** **Extracts kernel and rootfs from Docker image** ✨ **CRITICAL FIX**
- STEP 6: Creates a VM instance
- STEP 7: Boots the VM
- STEP 8: Captures boot console output
- STEP 9: Verifies successful boot

**Expected Output:**
```
✓ Kernel detected: Linux version 6.1.90...
✓ VM boot sequence complete
✓ All tests passed
E2E TEST SUCCESSFUL
```

## Full Workflow

```bash
# Step 1: Build everything
cd /home/jpoley/src/_mine/nanofuse/images/base
sudo ./build-complete.sh

# Step 2: Test everything
sudo /home/jpoley/src/_mine/nanofuse/test-complete.sh
```

If both succeed, you have a working NanoFuse system!

## Understanding the Fix (CRITICAL)

The test script includes a **critical fix** in STEP 5 that wasn't in the original implementation:

### Problem
The daemon's registry client creates **metadata** for images but doesn't actually extract the kernel and rootfs files from Docker images. When VMs try to boot, they fail with:
```
Kernel Loader: failed to load ELF kernel image
InvalidElfMagicNumber
```

### Solution
STEP 5 manually extracts:
1. Docker image digest (unique ID)
2. Creates storage directory: `/tmp/nanofuse/images/{digest}/`
3. Exports Docker image rootfs to this directory
4. Copies kernel binary to this directory
5. Creates ext4 filesystem from exported files
6. Ensures both `vmlinux` and `rootfs.ext4` exist

### Why It Works
The daemon expects image files at: `/tmp/nanofuse/images/{image-digest}/vmlinux` and `rootfs.ext4`

Our fix ensures these files actually exist before VMs try to use them.

## Troubleshooting

### Kernel Build Fails
```bash
# Check Docker is running
docker ps

# Check disk space
df -h /

# Retry with clean cache
cd images/base && docker system prune -a
```

### Test Fails at "Create VM"
- Daemon might not be running on port 8080
- Check: `lsof -i :8080` or `netstat -tuln | grep 8080`
- Check logs: `tail -20 /tmp/nanofused-*.log`

### Test Fails at "Console Output"
- VM directory not found at `/tmp/nanofuse/vms/`
- Firecracker might have crashed
- Check: `dmesg | tail -20`

### Mount Permission Errors
- Script requires sudo for mount operations
- Don't run without `sudo`

## Cleanup

```bash
# Stop all VMs and daemon
./bin/nanofuse vm list --api-url http://127.0.0.1:8080
./bin/nanofuse vm stop <vm-name> --api-url http://127.0.0.1:8080
pkill -f nanofused

# Clean old test data
rm -rf /tmp/nanofuse/vms/*
rm -rf /tmp/nanofuse/images/*

# Clean Docker images and containers
docker system prune -a
```

## File Structure

```
nanofuse/
├── build-complete.sh              # Builds kernel, Docker image, rootfs
├── test-complete.sh               # Tests complete system (with fix!)
├── KERNEL_LOADING_FIX.md          # Detailed explanation of the fix
├── QUICK_START.md                 # This file
│
├── images/base/
│   ├── build-complete.sh          # Same as above (can run from here)
│   ├── build.sh                   # Original simpler build script
│   ├── build-kernel-docker.sh     # Builds kernel in Docker
│   ├── Dockerfile                 # Ubuntu 24.04 + systemd base image
│   ├── Dockerfile.kernel          # Multi-stage kernel build
│   ├── build/                     # Output directory
│   │   ├── vmlinux               # Linux kernel (ELF binary)
│   │   ├── rootfs.ext4           # Filesystem image
│   │   └── manifest.json         # Image metadata
│   └── units/                    # Systemd service files
│       ├── firstboot.service
│       └── http-test-server.service
│
└── bin/
    ├── nanofused                 # Daemon binary
    └── nanofuse                  # CLI binary
```

## Key Concepts

### ELF vs vmlinux.bin
- **ELF format** (uncompressed): Required by Firecracker
- **vmlinux.bin** (compressed): Not compatible with Firecracker
- Our kernel is ELF (verified with `file build/vmlinux`)

### Image Digest
- Unique SHA256 hash of Docker image contents
- Used as storage directory name
- Enables caching and versioning

### ext4 Filesystem
- Standard Linux filesystem format
- Works as VM's root disk
- Created from exported Docker container files
- 2GB default size (configurable)

### Firecracker Configuration
- Daemon generates JSON config for each VM
- Points to kernel and rootfs paths
- systemd boot service files

## Next Steps

After successful build and test:

1. **Create custom VMs:**
   ```bash
   ./bin/nanofuse vm create nanofuse-base:latest my-vm --memory 512 --vcpus 2
   ./bin/nanofuse vm start my-vm --api-url http://127.0.0.1:8080
   ```

2. **Access VM console:**
   ```bash
   VM_DIR=$(ls -td /tmp/nanofuse/vms/* | head -1)
   tail -f $VM_DIR/console.log
   ```

3. **Stop VM:**
   ```bash
   ./bin/nanofuse vm stop my-vm --api-url http://127.0.0.1:8080
   ```

## References

- **Firecracker:** https://github.com/firecracker-microvm/firecracker
- **Linux Kernel:** https://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git
- **systemd:** https://systemd.io/

## Support

See `KERNEL_LOADING_FIX.md` for detailed technical explanation of the fix.

---

**TL;DR:** Run `build-complete.sh`, then `test-complete.sh`. If both succeed, you have a working system!
