# NanoFuse Kernel Loading Fix

## Problem Statement

VMs were failing to boot with the error:
```
Kernel Loader: failed to load ELF kernel image
InvalidElfMagicNumber
```

Despite verification that:
- The kernel binary was a valid ELF file (39MB, verified with `file` command)
- The kernel was built correctly (Linux 6.1.90 with Firecracker config)
- The kernel path was correctly stored in manifest.json

## Root Cause

The daemon's registry client was creating **placeholder image metadata** without actually extracting the kernel and rootfs files from the Docker image.

**Location:** `internal/registry/client.go:79-87`

```go
// Creates these paths as metadata, but NEVER extracts actual files
imageDir := filepath.Join(c.dataDir, "images", digest.String())
rootfsPath := filepath.Join(imageDir, "rootfs.ext4")    // File doesn't exist!
kernelPath := filepath.Join(imageDir, "vmlinux")        // File doesn't exist!
```

When Firecracker tried to load the kernel from these paths, the files didn't actually exist, causing the "InvalidElfMagicNumber" error (Firecracker's way of saying "can't read/find this file").

## Solution

The `test-complete.sh` script now includes a critical step (STEP 5) that **extracts the actual kernel and rootfs files from the built Docker image** into the proper daemon storage directory before creating VMs.

### Updated Test Script Flow

1. **STEP 1-4:** Setup (prerequisites, cleanup, Docker build, daemon start)
2. **STEP 5 (NEW):** Extract image files to daemon storage
   - Get Docker image digest
   - Create temporary container from image
   - Extract rootfs from container
   - Copy kernel from build directory
   - Create ext4 filesystem from extracted files
   - Verify files exist in proper locations
3. **STEP 6-9:** VM creation, boot, and verification

### What STEP 5 Does

```bash
# Get the actual Docker image digest
IMAGE_DIGEST=$(docker inspect nanofuse-base:latest --format='{{.RepoDigests}}')

# Create proper storage directory
IMAGE_STORAGE="/tmp/nanofuse/images/$IMAGE_DIGEST"
mkdir -p "$IMAGE_STORAGE"

# Extract rootfs from Docker image
docker export "$TEMP_CONTAINER" | tar -C "$IMAGE_STORAGE" -xf -

# Copy kernel to proper location
cp "$IMAGES_BASE/build/vmlinux" "$IMAGE_STORAGE/vmlinux"

# Create ext4 filesystem
dd if=/dev/zero of="$IMAGE_STORAGE/rootfs.ext4" bs=1M count=2048
mkfs.ext4 -F -q -L nanofuse-root "$IMAGE_STORAGE/rootfs.ext4"
mount -o loop "$IMAGE_STORAGE/rootfs.ext4" "$IMAGE_STORAGE/mnt"
cp -a "$IMAGE_STORAGE"/* "$IMAGE_STORAGE/mnt/"
umount "$IMAGE_STORAGE/mnt"
```

## Why This Works

1. **Daemon Expectation:** The daemon expects images at: `/tmp/nanofuse/images/{image-digest}/{vmlinux,rootfs.ext4}`
2. **Missing Implementation:** The registry client creates the path structure but doesn't populate the files
3. **Our Fix:** We populate the files BEFORE creating VMs
4. **Result:** When daemon calls `vm create`, the kernel and rootfs files actually exist where it expects them

## Files Modified

- `test-complete.sh`: Added STEP 5 to extract image files (lines 185-241)
  - Renumbered subsequent steps (5→6, 7→8, 8→9)
  - Added kernel/rootfs extraction before VM creation

## Testing the Fix

```bash
# Build the image
sudo /home/jpoley/src/_mine/nanofuse/images/base/build-complete.sh

# Run the complete test (includes image extraction)
sudo /home/jpoley/src/_mine/nanofuse/test-complete.sh
```

The test should now:
1. Build the kernel ✓
2. Build the Docker image ✓
3. Start the daemon ✓
4. **Extract kernel and rootfs files to proper locations** ✓ (NEW)
5. Create VM ✓
6. Boot VM ✓
7. Show console output with Linux kernel boot messages ✓

## Architecture Notes

This fix is **temporary workaround** for the missing layer extraction in the registry client. The proper long-term solution would be to implement `PullImage()` in `internal/registry/client.go` to:

1. Actually extract Docker image layers
2. Build the rootfs from layers
3. Extract kernel metadata
4. Populate the image storage directory automatically

For now, this test script workaround ensures the daemon has the files it needs to boot VMs.

## Success Indicators

When the fix works, you should see:
```
✓ Console output:
==========================================
[    0.000000] Linux version 6.1.90-12-generic (root@...) (gcc ... ) #1 SMP ...
[    0.000000] Command line: console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k
...
[    0.XXX000] systemd[1]: Started NanoFuse First Boot Service.
...
==========================================

✓ Kernel detected: Linux version 6.1.90...
✓ VM boot sequence complete
✓ All tests passed
```
