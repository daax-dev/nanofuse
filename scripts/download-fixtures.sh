#!/bin/bash
# download-fixtures.sh - Download official Firecracker CI images (Ubuntu 24.04;
# kernel 5.10.x/6.1.x, defaulting to 5.10.245 — see KERNEL_VERSION below).
# These are updated regularly by the Firecracker team and tested for compatibility
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
FIXTURES="$PROJECT_ROOT/test/fixtures/debug-kernel"

# Use sudo only when not already root, so this works on minimal root WSL/containers
# (where sudo may not be installed) as well as normal non-root shells.
SUDO=""
if [[ "$(id -u)" -ne 0 ]]; then
    if ! command -v sudo >/dev/null 2>&1; then
        echo "ERROR: not running as root and 'sudo' is not installed; run this script as root or install sudo." >&2
        exit 1
    fi
    SUDO="sudo"
fi

# Firecracker CI channel - use the maintained v1.15 channel prefix.
# The dated snapshot prefixes are pruned over time and 404 on the rootfs; the
# versioned channel (firecracker-ci/v1.15/x86_64/) is kept current by the
# Firecracker team and carries Ubuntu 24.04 plus 5.10.x and 6.1.x kernels.
# This script defaults to the 5.10.245 kernel (see KERNEL_VERSION below).
CI_VERSION="v1.15"

# Derive the Firecracker CI arch from the host (override with FIXTURES_ARCH).
ARCH="${FIXTURES_ARCH:-}"
if [[ -z "$ARCH" ]]; then
    case "$(uname -m)" in
        x86_64 | amd64) ARCH="x86_64" ;;
        aarch64 | arm64) ARCH="aarch64" ;;
        *) echo "ERROR: unsupported architecture $(uname -m); set FIXTURES_ARCH" >&2; exit 1 ;;
    esac
fi
S3_BASE="https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/${CI_VERSION}/${ARCH}"

# Image versions
# NOTE: On x86_64 use the -no-acpi kernel variant because standard CI kernels
# require CONFIG_PCI (only in Amazon Linux microvm patches); the -no-acpi variant
# uses legacy MPTable boot that works with virtio-mmio block devices.
# See: https://github.com/firecracker-microvm/firecracker/issues/4881
# aarch64 boots via device tree and uses the plain kernel.
if [[ "$ARCH" == "x86_64" ]]; then
    KERNEL_VERSION="${KERNEL_VERSION:-5.10.245-no-acpi}"
else
    KERNEL_VERSION="${KERNEL_VERSION:-5.10.245}"
fi
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

    # Mount and copy. Install a cleanup trap before/right after mounting so a
    # failure under set -e always unmounts and removes the temp dir.
    MOUNT_DIR=$(mktemp -d)
    trap '$SUDO umount "$MOUNT_DIR" 2>/dev/null || true; rmdir "$MOUNT_DIR" 2>/dev/null || true; rm -rf "$TEMP_DIR"' EXIT
    $SUDO mount -o loop "$FIXTURES/$EXT4_FILE" "$MOUNT_DIR"
    echo "  Copying files..."
    $SUDO cp -a "$TEMP_DIR/rootfs/." "$MOUNT_DIR/"
    $SUDO umount "$MOUNT_DIR"
    rmdir "$MOUNT_DIR"
    # Restore the original TEMP_DIR-only cleanup now that the mount is released.
    trap 'rm -rf "$TEMP_DIR"' EXIT

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
