#!/bin/bash
# download-fixtures.sh - Download official Firecracker CI images (Ubuntu 24.04 + kernel 6.1)
# These are updated regularly by the Firecracker team and tested for compatibility
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
FIXTURES="$PROJECT_ROOT/test/fixtures/debug-kernel"

# Firecracker CI channel - use the maintained v1.15 channel prefix.
# The dated snapshot prefixes are pruned over time and 404 on the rootfs; the
# versioned channel (firecracker-ci/v1.15/x86_64/) is kept current by the
# Firecracker team and carries Ubuntu 24.04 + kernel 5.10/6.1.
CI_VERSION="v1.15"
ARCH="x86_64"
S3_BASE="https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/${CI_VERSION}/${ARCH}"

# Image versions
# NOTE: Use -no-acpi kernel variant because standard CI kernels require CONFIG_PCI
# which is only available in Amazon Linux microvm patches. The -no-acpi variant
# uses legacy MPTable boot that works with virtio-mmio block devices.
# See: https://github.com/firecracker-microvm/firecracker/issues/4881
KERNEL_VERSION="5.10.245-no-acpi"
UBUNTU_VERSION="24.04"

echo "=== NanoFuse Fixture Downloader ==="
echo "CI Version: ${CI_VERSION}"
echo "Kernel: ${KERNEL_VERSION}"
echo "Ubuntu: ${UBUNTU_VERSION}"
echo "Target: ${FIXTURES}"
echo ""

mkdir -p "$FIXTURES"

# Download kernel
KERNEL_FILE="vmlinux-${KERNEL_VERSION}"
if [[ -f "$FIXTURES/$KERNEL_FILE" ]]; then
    echo "[SKIP] Kernel already exists: $KERNEL_FILE"
else
    echo "[DOWNLOAD] Kernel: $KERNEL_FILE"
    curl -L -o "$FIXTURES/$KERNEL_FILE" "${S3_BASE}/${KERNEL_FILE}"
    echo "[OK] Downloaded kernel ($(du -h "$FIXTURES/$KERNEL_FILE" | cut -f1))"
fi

# Create symlink for compatibility
ln -sf "$KERNEL_FILE" "$FIXTURES/vmlinux.bin"
echo "[LINK] vmlinux.bin -> $KERNEL_FILE"

# Download rootfs (squashfs format)
SQUASHFS_FILE="ubuntu-${UBUNTU_VERSION}.squashfs"
EXT4_FILE="ubuntu-${UBUNTU_VERSION}.ext4"

if [[ -f "$FIXTURES/$EXT4_FILE" ]]; then
    echo "[SKIP] Rootfs already exists: $EXT4_FILE"
else
    echo "[DOWNLOAD] Rootfs: $SQUASHFS_FILE"
    curl -L -o "$FIXTURES/$SQUASHFS_FILE" "${S3_BASE}/${SQUASHFS_FILE}"
    echo "[OK] Downloaded squashfs ($(du -h "$FIXTURES/$SQUASHFS_FILE" | cut -f1))"

    # Convert squashfs to ext4
    echo "[CONVERT] Converting squashfs to ext4..."

    if ! command -v unsquashfs &>/dev/null; then
        echo "ERROR: unsquashfs not found. Install: sudo apt install squashfs-tools"
        exit 1
    fi

    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT

    # Extract squashfs
    echo "  Extracting squashfs..."
    unsquashfs -d "$TEMP_DIR/rootfs" "$FIXTURES/$SQUASHFS_FILE"

    # Calculate size needed (add 50% buffer)
    ROOTFS_SIZE_KB=$(du -sk "$TEMP_DIR/rootfs" | cut -f1)
    ROOTFS_SIZE_MB=$(( (ROOTFS_SIZE_KB * 3 / 2) / 1024 + 100 ))
    echo "  Creating ext4 image (${ROOTFS_SIZE_MB}MB)..."

    # Create ext4 image
    dd if=/dev/zero of="$FIXTURES/$EXT4_FILE" bs=1M count=$ROOTFS_SIZE_MB status=progress
    mkfs.ext4 -F "$FIXTURES/$EXT4_FILE"

    # Mount and copy
    MOUNT_DIR=$(mktemp -d)
    sudo mount -o loop "$FIXTURES/$EXT4_FILE" "$MOUNT_DIR"
    echo "  Copying files..."
    sudo cp -a "$TEMP_DIR/rootfs/." "$MOUNT_DIR/"
    sudo umount "$MOUNT_DIR"
    rmdir "$MOUNT_DIR"

    # Cleanup squashfs (optional - comment out to keep)
    rm "$FIXTURES/$SQUASHFS_FILE"

    echo "[OK] Created ext4 rootfs ($(du -h "$FIXTURES/$EXT4_FILE" | cut -f1))"
fi

# Create symlink for compatibility with old scripts
ln -sf "$EXT4_FILE" "$FIXTURES/rootfs.ext4"
echo "[LINK] rootfs.ext4 -> $EXT4_FILE"

echo ""
echo "=== Fixtures Ready ==="
ls -lh "$FIXTURES/"
echo ""
echo "Next steps:"
echo "  1. ./scripts/debug-boot.sh   # Test direct Firecracker boot"
echo "  2. sudo ./scripts/register-debug-kernel.sh  # Register with nanofuse"
