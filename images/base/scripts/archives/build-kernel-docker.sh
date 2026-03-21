#!/bin/bash
set -euo pipefail

# Build Firecracker kernel using Docker
# This avoids having to install kernel build tools on your system

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="${SCRIPT_DIR}/build"
KERNEL_VERSION="6.1.90"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}✓${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}!${NC} $1"
}

echo "========================================="
echo "Firecracker Kernel Builder (Docker)"
echo "========================================="
echo ""
echo "Building Linux $KERNEL_VERSION kernel..."
echo "Using Firecracker's official microVM config"
echo "CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y (for Firecracker device injection)"
echo ""

# Create build directory
mkdir -p "$BUILD_DIR"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed"
    echo "Install Docker: https://docs.docker.com/engine/install/"
    exit 1
fi

log_info "Using Docker to build kernel"

echo ""
echo "[1/3] Building kernel Docker image..."
if docker build -f Dockerfile.kernel -t nanofuse-kernel-builder:latest . ; then
    log_info "Kernel builder image built"
else
    log_error "Failed to build kernel image"
    exit 1
fi

echo ""
echo "[2/3] Extracting kernel binary..."
# Extract kernel from Docker image
TEMP_KERNEL="/tmp/vmlinux-$$"
docker run --rm nanofuse-kernel-builder:latest cat /vmlinux > "$TEMP_KERNEL" || {
    log_error "Failed to extract kernel from image"
    rm -f "$TEMP_KERNEL"
    exit 1
}

echo ""
echo "[3/3] Placing kernel in build directory..."

# Ensure build directory exists
mkdir -p "$BUILD_DIR" 2>/dev/null || true

# Try to move kernel to build directory
KERNEL_FILE="$BUILD_DIR/vmlinux"
if mv "$TEMP_KERNEL" "$KERNEL_FILE" 2>/dev/null; then
    log_info "Kernel extracted to: $KERNEL_FILE"
elif cp "$TEMP_KERNEL" "$KERNEL_FILE" 2>/dev/null; then
    rm -f "$TEMP_KERNEL"
    log_info "Kernel copied to: $KERNEL_FILE"
else
    # Build directory not writable, keep in /tmp with consistent name
    KERNEL_FILE="/tmp/vmlinux-fresh-build"
    if mv "$TEMP_KERNEL" "$KERNEL_FILE" 2>/dev/null || cp "$TEMP_KERNEL" "$KERNEL_FILE" 2>/dev/null; then
        rm -f "$TEMP_KERNEL" 2>/dev/null || true
        log_warn "Build directory not writable, kernel saved to: $KERNEL_FILE"
        log_warn "Run 'sudo ./build.sh' to include this kernel in the final image"
    else
        log_error "Failed to save kernel to either $BUILD_DIR or /tmp"
        rm -f "$TEMP_KERNEL"
        exit 1
    fi
fi

if [ ! -f "$KERNEL_FILE" ]; then
    log_error "Kernel file not found at: $KERNEL_FILE"
    exit 1
fi

ls -lh "$KERNEL_FILE"
file "$KERNEL_FILE"

echo ""
echo "Kernel version:"
strings "$KERNEL_FILE" | grep "Linux version" || echo "Version info: built from Linux $KERNEL_VERSION"

echo ""
echo "========================================="
echo "✓ Kernel Build Complete!"
echo "========================================="
echo ""
echo "Kernel location: $KERNEL_FILE"
echo ""
