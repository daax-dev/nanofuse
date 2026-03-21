#!/bin/bash
set -euo pipefail

# NanoFuse Base Image Build Script
# Builds Firecracker-compatible rootfs, downloads kernel, generates manifest

# Change to script directory so it works from any location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

BUILD_DIR="./build"
IMAGE_NAME="${IMAGE_NAME:-nanofuse-base}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ROOTFS_SIZE="${ROOTFS_SIZE:-2048}"  # MB
ARCHITECTURE="${ARCHITECTURE:-x86_64}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if running as root for mount operations
if [[ $EUID -ne 0 ]] && [[ "${SKIP_ROOT_CHECK:-}" != "1" ]]; then
    log_error "This script requires sudo for mounting ext4 filesystems"
    echo "Run with: sudo ./build.sh"
    echo "Or set SKIP_ROOT_CHECK=1 if you have passwordless sudo configured"
    exit 1
fi

echo "========================================"
echo "NanoFuse Base Image Build"
echo "========================================"
echo ""

# Step 1: Build Docker image
echo "[1/6] Building Docker image..."
if docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" . ; then
    log_info "Docker image built: ${IMAGE_NAME}:${IMAGE_TAG}"
else
    log_error "Docker build failed"
    exit 1
fi

# Step 2: Create container and export filesystem
echo "[2/6] Exporting container filesystem..."
mkdir -p "${BUILD_DIR}/rootfs"

# Remove any existing container with same name
docker rm nanofuse-build-temp 2>/dev/null || true

CONTAINER_ID=$(docker create --name nanofuse-build-temp "${IMAGE_NAME}:${IMAGE_TAG}")
log_info "Container created: ${CONTAINER_ID:0:12}"

docker export "${CONTAINER_ID}" | tar -C "${BUILD_DIR}/rootfs" -xf -
log_info "Filesystem exported"

docker rm "${CONTAINER_ID}" > /dev/null
log_info "Container cleaned up"

# Step 3: Create ext4 image
echo "[3/6] Creating ext4 filesystem image..."
dd if=/dev/zero of="${BUILD_DIR}/rootfs.ext4" bs=1M count="${ROOTFS_SIZE}" status=none
mkfs.ext4 -F -q -L nanofuse-root "${BUILD_DIR}/rootfs.ext4"
log_info "Created ${ROOTFS_SIZE}MB ext4 image"

# Step 4: Mount and copy files
echo "[4/6] Copying filesystem to ext4 image..."
mkdir -p "${BUILD_DIR}/mnt"
mount -o loop "${BUILD_DIR}/rootfs.ext4" "${BUILD_DIR}/mnt"
log_info "Mounted ext4 image"

cp -a "${BUILD_DIR}/rootfs/"* "${BUILD_DIR}/mnt/"
log_info "Files copied to ext4 image"

umount "${BUILD_DIR}/mnt"
rmdir "${BUILD_DIR}/mnt"
rm -rf "${BUILD_DIR}/rootfs"

# Set proper permissions so non-root users can use the image
# When run via sudo, use $SUDO_UID:$SUDO_GID to get the original user
if [ -n "${SUDO_UID:-}" ]; then
    chown ${SUDO_UID}:${SUDO_GID} "${BUILD_DIR}/rootfs.ext4"
else
    chown $(id -u):$(id -g) "${BUILD_DIR}/rootfs.ext4"
fi
chmod 664 "${BUILD_DIR}/rootfs.ext4"
log_info "Unmounted and cleaned up"

# Step 5: Build Firecracker kernel with modern config
echo "[5/6] Building Firecracker kernel..."

# Check for kernel in standard location
if [ -f "${BUILD_DIR}/vmlinux" ]; then
    log_info "Using pre-built kernel: $(du -h ${BUILD_DIR}/vmlinux | cut -f1)"
else
    # Check for kernel in /tmp with standard names
    TEMP_KERNEL=""
    for kernel_path in /tmp/vmlinux-fresh-build /tmp/vmlinux-final /tmp/vmlinux-*; do
        if [ -f "$kernel_path" ]; then
            TEMP_KERNEL="$kernel_path"
            break
        fi
    done

    if [ -n "$TEMP_KERNEL" ] && [ -f "$TEMP_KERNEL" ]; then
        log_info "Found kernel at $TEMP_KERNEL, copying to ${BUILD_DIR}/vmlinux..."
        cp "$TEMP_KERNEL" "${BUILD_DIR}/vmlinux"
        log_info "Kernel ready: $(du -h ${BUILD_DIR}/vmlinux | cut -f1)"
    else
        # Build kernel
        log_info "No pre-built kernel found, building with Firecracker config..."
        if ./scripts/archives/build-kernel-docker.sh; then
            # Check again for kernel in /tmp or build dir
            if [ -f "${BUILD_DIR}/vmlinux" ]; then
                log_info "Kernel built successfully: $(du -h ${BUILD_DIR}/vmlinux | cut -f1)"
            else
                # Try to find in /tmp again
                for kernel_path in /tmp/vmlinux-fresh-build /tmp/vmlinux-final /tmp/vmlinux-*; do
                    if [ -f "$kernel_path" ]; then
                        cp "$kernel_path" "${BUILD_DIR}/vmlinux"
                        log_info "Kernel built and copied: $(du -h ${BUILD_DIR}/vmlinux | cut -f1)"
                        break
                    fi
                done

                if [ ! -f "${BUILD_DIR}/vmlinux" ]; then
                    log_error "Kernel build failed - vmlinux not found in ${BUILD_DIR} or /tmp"
                    exit 1
                fi
            fi
        else
            log_error "Kernel build script failed"
            exit 1
        fi
    fi
fi

# Step 6: Generate manifest
echo "[6/6] Generating manifest.json..."
ROOTFS_SIZE_BYTES=$(stat -c%s "${BUILD_DIR}/rootfs.ext4")
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

cat > "${BUILD_DIR}/manifest.json" << EOF
{
  "version": "0.1.0",
  "name": "${IMAGE_NAME}",
  "tag": "${IMAGE_TAG}",
  "architecture": "${ARCHITECTURE}",
  "base_os": "ubuntu:24.04",
  "kernel": {
    "version": "6.1.90",
    "source": "Linux kernel (Firecracker config)",
    "cmdline": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k",
    "file": "vmlinux"
  },
  "rootfs": {
    "format": "ext4",
    "file": "rootfs.ext4",
    "size_bytes": ${ROOTFS_SIZE_BYTES}
  },
  "services": {
    "ssh": {"enabled": true, "port": 22},
    "systemd-networkd": {"enabled": true},
    "http-test-server": {"enabled": true, "port": 8080}
  },
  "built_at": "${BUILD_DATE}"
}
EOF
log_info "Manifest generated"

echo ""
echo "========================================"
echo "Build Complete!"
echo "========================================"
echo ""
echo "Artifacts created:"
ls -lh "${BUILD_DIR}/rootfs.ext4" "${BUILD_DIR}/vmlinux" "${BUILD_DIR}/manifest.json"
echo ""
echo "Verification:"
file "${BUILD_DIR}/rootfs.ext4" "${BUILD_DIR}/vmlinux"
echo ""
echo "Ready to use with Firecracker!"
