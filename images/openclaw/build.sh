#!/bin/bash
set -euo pipefail

# OpenClaw MicroVM Image Build Script
# Builds Firecracker-compatible rootfs from Docker container

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

IMAGE_NAME="${IMAGE_NAME:-openclaw}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
BUILD_DIR="./build"
ROOTFS_SIZE="${ROOTFS_SIZE:-2048}"  # 2GB for Node.js + npm packages
KERNEL_PATH="${KERNEL_PATH:-../base/build/vmlinux}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}✓${NC} $1"; }
log_warn() { echo -e "${YELLOW}⚠${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }

# Check root for mount operations
if [[ $EUID -ne 0 ]] && [[ "${SKIP_ROOT_CHECK:-}" != "1" ]]; then
    log_error "This script requires sudo for mounting ext4 filesystems"
    echo "Run with: sudo ./build.sh"
    exit 1
fi

echo "========================================"
echo "OpenClaw MicroVM Image Build"
echo "========================================"
echo ""

# Ensure kernel exists
if [[ ! -f "$KERNEL_PATH" ]]; then
    log_warn "Kernel not found at $KERNEL_PATH"
    echo "Run: cd ../base && sudo ./build.sh  (to build kernel)"
    echo "Or set KERNEL_PATH to existing vmlinux"
    SKIP_KERNEL=1
else
    log_info "Using kernel: $KERNEL_PATH"
    SKIP_KERNEL=0
fi

# Step 1: Build Docker image
echo ""
echo "[1/5] Building Docker image..."
BUILD_ARGS=()
if [[ -n "${CLAWDBOT_VERSION:-}" ]]; then
    BUILD_ARGS+=(--build-arg "CLAWDBOT_VERSION=${CLAWDBOT_VERSION}")
fi
if docker build "${BUILD_ARGS[@]}" -t "${IMAGE_NAME}:${IMAGE_TAG}" . ; then
    log_info "Docker image built: ${IMAGE_NAME}:${IMAGE_TAG}"
else
    log_error "Docker build failed"
    exit 1
fi

# Step 2: Export container filesystem
echo ""
echo "[2/5] Exporting container filesystem..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}/rootfs"

docker rm openclaw-build-temp 2>/dev/null || true
CONTAINER_ID=$(docker create --name openclaw-build-temp "${IMAGE_NAME}:${IMAGE_TAG}")
log_info "Container created: ${CONTAINER_ID:0:12}"

docker export "${CONTAINER_ID}" | tar -C "${BUILD_DIR}/rootfs" -xf -
log_info "Filesystem exported to ${BUILD_DIR}/rootfs"

docker rm "${CONTAINER_ID}" > /dev/null
log_info "Container cleaned up"

# Step 3: Create ext4 image
echo ""
echo "[3/5] Creating ext4 filesystem image (${ROOTFS_SIZE}MB)..."
dd if=/dev/zero of="${BUILD_DIR}/rootfs.ext4" bs=1M count="${ROOTFS_SIZE}" status=progress
mkfs.ext4 -F -q -L openclaw-root "${BUILD_DIR}/rootfs.ext4"
log_info "Created ext4 image: ${BUILD_DIR}/rootfs.ext4"

# Step 4: Mount and copy
echo ""
echo "[4/5] Copying filesystem to ext4 image..."
mkdir -p "${BUILD_DIR}/mnt"
mount -o loop "${BUILD_DIR}/rootfs.ext4" "${BUILD_DIR}/mnt"

# Ensure cleanup on failure (unmount if still mounted)
cleanup_mount() {
    umount "${BUILD_DIR}/mnt" 2>/dev/null || true
    rmdir "${BUILD_DIR}/mnt" 2>/dev/null || true
}
trap cleanup_mount EXIT

cp -a "${BUILD_DIR}/rootfs/." "${BUILD_DIR}/mnt/"
cleanup_mount
trap - EXIT
log_info "Filesystem copied"

# Cleanup extracted rootfs to save space
rm -rf "${BUILD_DIR}/rootfs"
log_info "Cleaned up temporary files"

# Fix ownership if run via sudo (match base image build pattern)
if [[ -n "${SUDO_UID:-}" ]] && [[ -n "${SUDO_GID:-}" ]]; then
    chown "${SUDO_UID}:${SUDO_GID}" "${BUILD_DIR}/rootfs.ext4"
    log_info "Fixed ownership for rootfs.ext4"
fi

# Step 5: Copy or link kernel
echo ""
echo "[5/5] Setting up kernel..."
if [[ "$SKIP_KERNEL" == "0" ]]; then
    cp "$KERNEL_PATH" "${BUILD_DIR}/vmlinux"
    if [[ -n "${SUDO_UID:-}" ]] && [[ -n "${SUDO_GID:-}" ]]; then
        chown "${SUDO_UID}:${SUDO_GID}" "${BUILD_DIR}/vmlinux"
    fi
    log_info "Kernel copied to ${BUILD_DIR}/vmlinux"
else
    log_warn "Skipped kernel (not found)"
fi

# Summary
echo ""
echo "========================================"
echo "Build Complete!"
echo "========================================"
echo ""
echo "Artifacts:"
echo "  Rootfs:  ${BUILD_DIR}/rootfs.ext4 ($(du -h "${BUILD_DIR}/rootfs.ext4" | cut -f1))"
if [[ -f "${BUILD_DIR}/vmlinux" ]]; then
    echo "  Kernel:  ${BUILD_DIR}/vmlinux"
fi
echo ""
echo "Next steps:"
echo "  1. Use the rootfs with Firecracker directly:"
echo "     firecracker --config-file vm-config.json"
echo "     # Set root_drive.path_on_host to: ${SCRIPT_DIR}/${BUILD_DIR}/rootfs.ext4"
echo ""
echo "  2. Or push to a container registry:"
echo "     docker push ghcr.io/<org>/openclaw:latest"
echo ""
echo "  3. SSH into VM (after launching):"
echo "     ssh clawdbot@<vm-ip>"
