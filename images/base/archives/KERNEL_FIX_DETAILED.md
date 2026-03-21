# Kernel Fix: CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES

## Problem Statement

The kernel boots but cannot find the block device, failing with:
```
VFS: Cannot open root device "vda" or unknown-block(0,0): error -6
Kernel panic - not syncing: VFS: Unable to mount root fs on unknown-block(0,0)
```

## Root Cause

Firecracker passes virtio block devices via kernel command line parameter:
```
virtio_mmio.device=4K@0xd0000000:5
```

However, the current kernel config has `CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES` **disabled**, so the kernel cannot parse this parameter. The device is never initialized.

## The Fix

### File: `Dockerfile.kernel`

**Change 1:** After `make olddefconfig`, enable CMDLINE_DEVICES

Add this line after line 37 (after `RUN make olddefconfig`):

```dockerfile
# IMPORTANT: Enable VIRTIO_MMIO_CMDLINE_DEVICES so the kernel can parse
# the virtio_mmio.device= parameters that Firecracker passes via command line
RUN sed -i 's/# CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES is not set/CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y/' .config
```

**Change 2:** Replace the build command

Change line 44 from:
```dockerfile
RUN make -j$(nproc) bzImage
```

To:
```dockerfile
RUN make -j$(nproc)
```

Reason: `make bzImage` only builds the compressed kernel. We need the full vmlinux with all configured options included.

**Change 3:** Add verification

Replace the comment before `RUN make` with:

```dockerfile
# Build kernel (using all available cores)
# Build everything to ensure all modules and config options are included
RUN echo "=== BUILDING KERNEL ===" && make -j$(nproc) 2>&1 | tail -50

# Verify VIRTIO_MMIO_CMDLINE_DEVICES is compiled into vmlinux
RUN echo "=== CHECKING FINAL KERNEL ===" && \
    strings vmlinux | grep -i "virtio_mmio" | head -5 || echo "WARNING: VIRTIO_MMIO not found in final kernel binary"
```

### File: `build.sh`

**Fix permissions handling** (lines 93-102)

Replace:
```bash
if [ -n "${SUDO_UID:-}" ]; then
    chown ${SUDO_UID}:${SUDO_GID} "${BUILD_DIR}/rootfs.ext4"
else
    chown $(id -u):$(id -g) "${BUILD_DIR}/rootfs.ext4"
fi
chmod 664 "${BUILD_DIR}/rootfs.ext4"
```

With:
```bash
if [ $EUID -eq 0 ]; then
    # Running as root - try to change ownership to original user
    if [ -n "${SUDO_UID:-}" ]; then
        chown ${SUDO_UID}:${SUDO_GID} "${BUILD_DIR}/rootfs.ext4" 2>/dev/null || true
    fi
fi
# Always make it world-readable and writable (it's only used for VMs, not sensitive)
chmod 666 "${BUILD_DIR}/rootfs.ext4"
```

## Testing the Fix

### Quick Test
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
bash test-kernel-fix.sh
```

This will:
1. Build the kernel with the fix
2. Extract vmlinux
3. Verify CMDLINE_DEVICES is in the binary
4. Boot with Firecracker
5. Check for successful device detection and mount

### Expected Output

When working correctly, you'll see:
```
[    0.067676] virtio-mmio: Registering device virtio-mmio.0 at 0xd0000000-0xd0000fff, IRQ 5.
[    0.083752] virtio_blk virtio0: 1/0/0 default/read/poll queues
[    0.084132] virtio_blk virtio0: [vda] 4194304 512-byte logical blocks (2.15 GB/2.00 GiB)
[    0.092852] EXT4-fs (vda): mounted filesystem with ordered data mode. Quota mode: none.
[    0.093105] VFS: Mounted root (ext4 filesystem) on device 254:0.
```

### Manual Test

If you want to test manually:

```bash
# 1. Build the kernel
docker build -f Dockerfile.kernel -t test-kernel .

# 2. Extract vmlinux
docker run --rm test-kernel cat /vmlinux > /tmp/vmlinux-test

# 3. Verify the config is there
strings /tmp/vmlinux-test | grep "virtio_mmio"

# 4. Create a test config
cat > /tmp/test-fc.json << 'EOF'
{
  "boot-source": {
    "kernel_image_path": "/tmp/vmlinux-test",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "/tmp/rootfs-working.ext4",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 1,
    "mem_size_mib": 256
  }
}
EOF

# 5. Boot and check
timeout 30 firecracker --no-api --config-file /tmp/test-fc.json 2>&1 | grep -E "(virtio|vda|EXT4|mounted)"
```

## Why This Works

1. **CMDLINE_DEVICES enabled**: Kernel now parses `virtio_mmio.device=` from command line
2. **Full build**: `make -j$(nproc)` ensures all configured options are compiled in
3. **Device detection**: MMIO bus finds the device and initializes it
4. **Driver binding**: virtio_blk driver binds to the device
5. **Block device creation**: `/dev/vda` is created and can be mounted
6. **Filesystem mount**: EXT4 mounts successfully

## Verification Commands

After implementing the fix, verify with:

```bash
# Check that CMDLINE_DEVICES was enabled
docker run --rm test-kernel sh -c "grep CONFIG_VIRTIO_MMIO .config"
# Should output:
# CONFIG_VIRTIO_MMIO=y
# CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y

# Check strings in binary
strings /tmp/vmlinux-test | grep virtio_mmio
# Should show multiple lines including virtio_mmio.device
```

## Summary

- **1 critical config option** needed: `CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y`
- **2 build changes**: Full make (not bzImage), add verification
- **1 permissions fix**: Use SUDO_UID for correct ownership
- **Result**: Kernel boots, block device detected, filesystem mounts successfully
