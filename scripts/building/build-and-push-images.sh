#!/bin/bash
# Build base and todo-app images with kernel 6.1.90, then push to GHCR
# This ensures we always have working images with the correct kernel

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BASE_IMAGE_DIR="$REPO_ROOT/images/base"
TODO_APP_DIR="$REPO_ROOT/examples/todo-app"

# GHCR configuration
GHCR_ORG="ghcr.io/peregrinesummit"
BASE_IMAGE_NAME="$GHCR_ORG/nanofuse/base"
TODO_APP_IMAGE_NAME="$GHCR_ORG/nanofuse/todo-app"

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

log_section() {
    echo ""
    echo "========================================="
    echo "$1"
    echo "========================================="
}

# Check prerequisites
if [ "$EUID" -ne 0 ]; then
    log_error "This script must be run as root (needs sudo for ext4 mounting)"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed"
    exit 1
fi

# Check GHCR authentication
# When running with sudo, check if root is logged in OR if we can use user's credentials
if ! docker info 2>&1 | grep -q "ghcr.io" && ! grep -q "ghcr.io" ~/.docker/config.json 2>/dev/null && ! grep -q "ghcr.io" /root/.docker/config.json 2>/dev/null; then
    log_error "Not logged in to GHCR."
    echo ""
    echo "Option 1 - Login as current user, then use sudo -E:"
    echo "  echo \$GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin"
    echo "  sudo -E ./scripts/building/build-and-push-images.sh"
    echo ""
    echo "Option 2 - Login as root:"
    echo "  sudo bash -c 'echo \$GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin'"
    echo "  sudo ./scripts/building/build-and-push-images.sh"
    exit 1
fi

log_section "Step 1: Build Base Image with Kernel 6.1.90"

cd "$BASE_IMAGE_DIR"

# Check if kernel already built
if [ ! -f "./build/vmlinux" ] || [ $(stat -c%s "./build/vmlinux") -lt 1000000 ]; then
    log_info "Building kernel 6.1.90 (this takes 10-15 minutes)..."
    ./scripts/build-kernel-docker.sh

    # build-kernel-docker.sh writes to scripts/build/vmlinux; stage it at ./build/vmlinux
    # (where the check below and build.sh expect it) before validating.
    mkdir -p build
    for k in scripts/build/vmlinux /tmp/vmlinux-fresh-build; do
        if [ -f "$k" ]; then cp "$k" build/vmlinux; break; fi
    done

    if [ ! -f "./build/vmlinux" ] || [ $(stat -c%s "./build/vmlinux") -lt 1000000 ]; then
        log_error "Kernel build failed or produced invalid kernel"
        exit 1
    fi

    log_info "Kernel built: $(du -h ./build/vmlinux | cut -f1)"
else
    log_info "Kernel already built: $(du -h ./build/vmlinux | cut -f1)"
fi

# Build base image
log_info "Building base image..."
./build.sh

# Verify kernel version
KERNEL_VERSION=$(strings ./build/vmlinux | grep "^Linux version" | head -1)
log_info "Kernel version: $KERNEL_VERSION"

if ! echo "$KERNEL_VERSION" | grep -q "6\.1\."; then
    log_error "Wrong kernel version! Expected 6.1.x, got: $KERNEL_VERSION"
    exit 1
fi

log_section "Step 2: Build Todo-App Image"

cd "$TODO_APP_DIR"

# Check if build script exists
if [ ! -f "./build-nanofuse-image.sh" ]; then
    log_error "Todo-app build script not found: ./build-nanofuse-image.sh"
    exit 1
fi

log_info "Building todo-app image with base kernel 6.1.90..."
./build-nanofuse-image.sh

# Verify output
if [ ! -f "./output/vmlinux" ] || [ ! -f "./output/rootfs.ext4" ]; then
    log_error "Todo-app build failed - missing output files"
    exit 1
fi

# Check kernel version in todo-app output
TODO_KERNEL_VERSION=$(strings ./output/vmlinux | grep "^Linux version" | head -1)
log_info "Todo-app kernel version: $TODO_KERNEL_VERSION"

if ! echo "$TODO_KERNEL_VERSION" | grep -q "6\.1\."; then
    log_error "Todo-app has wrong kernel version! Expected 6.1.x, got: $TODO_KERNEL_VERSION"
    exit 1
fi

log_section "Step 3: Tag Images for GHCR"

# Build Docker images for pushing to GHCR
cd "$BASE_IMAGE_DIR"
docker build -t "$BASE_IMAGE_NAME:latest" -f Dockerfile .
docker tag "$BASE_IMAGE_NAME:latest" "$BASE_IMAGE_NAME:6.1.90"

cd "$TODO_APP_DIR"
docker build -t "$TODO_APP_IMAGE_NAME:latest" -f docker/Dockerfile .
docker tag "$TODO_APP_IMAGE_NAME:latest" "$TODO_APP_IMAGE_NAME:$(date +%Y%m%d)"

log_section "Step 4: Push to GHCR"

log_info "Pushing base image to GHCR..."
docker push "$BASE_IMAGE_NAME:latest"
docker push "$BASE_IMAGE_NAME:6.1.90"

log_info "Pushing todo-app image to GHCR..."
docker push "$TODO_APP_IMAGE_NAME:latest"
docker push "$TODO_APP_IMAGE_NAME:$(date +%Y%m%d)"

log_section "Step 5: Create NanoFuse Image Metadata"

# Now we need to convert Docker images to NanoFuse format with correct metadata
# This involves extracting kernel and rootfs from the images

log_info "Creating NanoFuse-compatible image archives..."

# Base image
cd "$BASE_IMAGE_DIR"
MANIFEST_FILE="./build/manifest.json"
if [ -f "$MANIFEST_FILE" ]; then
    log_info "Base image manifest: $MANIFEST_FILE"
    cat "$MANIFEST_FILE"
fi

# Todo-app image
cd "$TODO_APP_DIR"
if [ -f "./output/manifest.json" ]; then
    log_info "Todo-app manifest: ./output/manifest.json"
    cat "./output/manifest.json"
fi

log_section "Success!"

echo ""
log_info "Images built and pushed to GHCR:"
echo "  - $BASE_IMAGE_NAME:latest (kernel 6.1.90)"
echo "  - $BASE_IMAGE_NAME:6.1.90"
echo "  - $TODO_APP_IMAGE_NAME:latest (kernel 6.1.90)"
echo "  - $TODO_APP_IMAGE_NAME:$(date +%Y%m%d)"
echo ""
log_info "To use these images:"
echo "  nanofuse image pull $TODO_APP_IMAGE_NAME:latest"
echo "  nanofuse vm create $TODO_APP_IMAGE_NAME:latest my-vm"
echo ""
log_info "Kernel version in all images: $KERNEL_VERSION"
