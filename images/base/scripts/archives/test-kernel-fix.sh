#!/bin/bash
set -euo pipefail

# Test script to validate the VIRTIO_MMIO_CMDLINE_DEVICES kernel fix
# This script:
# 1. Builds the kernel with CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y
# 2. Extracts the vmlinux binary
# 3. Tests it with Firecracker
# 4. Validates that the block device is detected and mounted

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

echo "=========================================="
echo "Kernel Fix Validation Test"
echo "=========================================="
echo ""

# Step 1: Build the kernel with the fix
echo "[1/5] Building kernel with CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y..."
BUILD_LOG=$(mktemp)
if docker build -f Dockerfile.kernel -t nanofuse-kernel:test . > "$BUILD_LOG" 2>&1; then
    log_info "Kernel built successfully"
else
    log_error "Kernel build failed"
    tail -50 "$BUILD_LOG"
    rm -f "$BUILD_LOG"
    exit 1
fi
rm -f "$BUILD_LOG"

# Step 2: Extract vmlinux
echo "[2/5] Extracting vmlinux from Docker image..."
KERNEL_PATH="/tmp/vmlinux-test-$$"
if docker run --rm nanofuse-kernel:test cat /vmlinux > "$KERNEL_PATH" 2>&1; then
    log_info "Kernel extracted to $KERNEL_PATH"
    file "$KERNEL_PATH"
else
    log_error "Failed to extract kernel"
    exit 1
fi

# Step 3: Verify VIRTIO_MMIO_CMDLINE_DEVICES is in the binary
echo "[3/5] Verifying CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES is compiled in..."
if strings "$KERNEL_PATH" | grep -q "virtio_mmio.device"; then
    log_info "CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES found in kernel binary"
else
    log_error "CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES NOT found in kernel binary"
    exit 1
fi

# Step 4: Create Firecracker test config
echo "[4/5] Creating Firecracker test configuration..."
TEST_CONFIG="/tmp/test-kernel-fix-$$.json"
cat > "$TEST_CONFIG" << EOF
{
  "boot-source": {
    "kernel_image_path": "$KERNEL_PATH",
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
log_info "Test config created: $TEST_CONFIG"

# Step 5: Boot with Firecracker and check for success indicators
echo "[5/5] Booting with Firecracker and validating..."
echo ""
echo "Expected boot sequence:"
echo "  1. 'virtio-mmio: Registering device' - Device detected"
echo "  2. 'virtio_blk' - Block driver loaded"
echo "  3. '[vda]' - Block device discovered"
echo "  4. 'EXT4-fs' or 'mounted filesystem' - Filesystem mounted"
echo ""

BOOT_OUTPUT=$(mktemp)
timeout 30 firecracker --no-api --config-file "$TEST_CONFIG" 2>&1 | tee "$BOOT_OUTPUT" | grep -E "(virtio|vda|EXT4|mounted|Kernel panic)" | head -15

echo ""
echo "Analyzing boot results..."
echo ""

VIRTIO_MMIO=$(grep -c "virtio-mmio:" "$BOOT_OUTPUT" || echo 0)
VIRTIO_BLK=$(grep -c "virtio_blk" "$BOOT_OUTPUT" || echo 0)
BLOCK_DEVICE=$(grep -c "\[vda\]" "$BOOT_OUTPUT" || echo 0)
EXT4_MOUNT=$(grep -c "EXT4-fs.*mounted\|VFS: Mounted root" "$BOOT_OUTPUT" || echo 0)
KERNEL_PANIC=$(grep -c "Kernel panic" "$BOOT_OUTPUT" || echo 0)

echo "Boot Analysis Results:"
echo "  virtio-mmio device detected: $([ $VIRTIO_MMIO -gt 0 ] && echo "YES ✓" || echo "NO ✗")"
echo "  virtio_blk driver loaded: $([ $VIRTIO_BLK -gt 0 ] && echo "YES ✓" || echo "NO ✗")"
echo "  Block device [vda] found: $([ $BLOCK_DEVICE -gt 0 ] && echo "YES ✓" || echo "NO ✗")"
echo "  EXT4 filesystem mounted: $([ $EXT4_MOUNT -gt 0 ] && echo "YES ✓" || echo "NO ✗")"
echo "  Kernel panic: $([ $KERNEL_PANIC -gt 0 ] && echo "YES ✗" || echo "NO ✓")"
echo ""

# Cleanup
rm -f "$TEST_CONFIG" "$BOOT_OUTPUT" "$KERNEL_PATH"

if [ $VIRTIO_MMIO -gt 0 ] && [ $VIRTIO_BLK -gt 0 ] && [ $BLOCK_DEVICE -gt 0 ] && [ $EXT4_MOUNT -gt 0 ] && [ $KERNEL_PANIC -eq 0 ]; then
    echo "=========================================="
    log_info "ALL TESTS PASSED ✓"
    echo "=========================================="
    exit 0
else
    echo "=========================================="
    log_error "TESTS FAILED ✗"
    echo "=========================================="
    exit 1
fi
