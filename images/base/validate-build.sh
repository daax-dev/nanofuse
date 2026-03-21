#!/bin/bash
#
# validate-build.sh - Validate NanoFuse base image build artifacts
#
# This script validates that all build artifacts are correct:
# 1. Docker image built successfully
# 2. Rootfs.ext4 is valid ext4 filesystem
# 3. Kernel is valid uncompressed Linux kernel
# 4. Manifest.json is valid JSON with required fields
# 5. All required files present in rootfs
#
# Usage: ./validate-build.sh [build-dir]

set -e

BUILD_DIR=${1:-./build}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "======================================"
echo "NanoFuse Base Image Build Validation"
echo "======================================"
echo ""

ERRORS=0
WARNINGS=0

# Helper functions
error() {
    echo -e "${RED}✗ ERROR: $1${NC}"
    ERRORS=$((ERRORS + 1))
}

warning() {
    echo -e "${YELLOW}⚠ WARNING: $1${NC}"
    WARNINGS=$((WARNINGS + 1))
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check build directory exists
if [ ! -d "$BUILD_DIR" ]; then
    error "Build directory not found: $BUILD_DIR"
    echo ""
    echo "Run 'make build' first to create build artifacts"
    exit 1
fi

echo "Validating build artifacts in: $BUILD_DIR"
echo ""

# Test 1: Check rootfs.ext4 exists and is valid
echo "=== Rootfs Validation ==="
ROOTFS_FILE="$BUILD_DIR/rootfs.ext4"

if [ ! -f "$ROOTFS_FILE" ]; then
    error "Rootfs file not found: $ROOTFS_FILE"
else
    success "Rootfs file exists: $ROOTFS_FILE"

    # Check file size
    ROOTFS_SIZE=$(stat -c%s "$ROOTFS_FILE")
    ROOTFS_SIZE_MB=$((ROOTFS_SIZE / 1024 / 1024))
    info "Rootfs size: ${ROOTFS_SIZE_MB}MB"

    if [ $ROOTFS_SIZE_MB -lt 100 ]; then
        warning "Rootfs size seems too small (< 100MB)"
    fi

    # Check it's an ext4 filesystem
    if file "$ROOTFS_FILE" | grep -q "ext4"; then
        success "Rootfs is valid ext4 filesystem"
    else
        error "Rootfs is not ext4 filesystem: $(file $ROOTFS_FILE)"
    fi

    # Try to check filesystem integrity (requires root)
    if [ "$EUID" -eq 0 ]; then
        if e2fsck -n "$ROOTFS_FILE" > /dev/null 2>&1; then
            success "Rootfs filesystem integrity check passed"
        else
            error "Rootfs filesystem has errors"
        fi
    else
        info "Skipping filesystem integrity check (requires root)"
    fi
fi

echo ""

# Test 2: Check kernel exists and is valid
echo "=== Kernel Validation ==="
KERNEL_FILE="$BUILD_DIR/vmlinux"

if [ ! -f "$KERNEL_FILE" ]; then
    error "Kernel file not found: $KERNEL_FILE"
else
    success "Kernel file exists: $KERNEL_FILE"

    # Check file size
    KERNEL_SIZE=$(stat -c%s "$KERNEL_FILE")
    KERNEL_SIZE_MB=$((KERNEL_SIZE / 1024 / 1024))
    info "Kernel size: ${KERNEL_SIZE_MB}MB"

    if [ $KERNEL_SIZE_MB -lt 5 ]; then
        warning "Kernel size seems too small (< 5MB)"
    fi

    # Check it's an uncompressed Linux kernel
    FILE_OUTPUT=$(file "$KERNEL_FILE")
    if echo "$FILE_OUTPUT" | grep -q "Linux kernel"; then
        success "Kernel is valid Linux kernel"

        # Check if it's uncompressed (Firecracker requirement)
        if echo "$FILE_OUTPUT" | grep -q "bzImage"; then
            error "Kernel is compressed (bzImage). Firecracker requires uncompressed vmlinux"
        else
            success "Kernel is uncompressed (required for Firecracker)"
        fi
    else
        error "Kernel file is not a Linux kernel: $FILE_OUTPUT"
    fi
fi

echo ""

# Test 3: Check manifest.json exists and is valid
echo "=== Manifest Validation ==="
MANIFEST_FILE="$BUILD_DIR/manifest.json"

if [ ! -f "$MANIFEST_FILE" ]; then
    error "Manifest file not found: $MANIFEST_FILE"
else
    success "Manifest file exists: $MANIFEST_FILE"

    # Check it's valid JSON
    if jq empty "$MANIFEST_FILE" 2>/dev/null; then
        success "Manifest is valid JSON"

        # Check required fields
        REQUIRED_FIELDS=("version" "name" "architecture" "kernel" "rootfs")
        for field in "${REQUIRED_FIELDS[@]}"; do
            if jq -e ".$field" "$MANIFEST_FILE" > /dev/null 2>&1; then
                success "Manifest contains required field: $field"
            else
                error "Manifest missing required field: $field"
            fi
        done

        # Show manifest contents
        echo ""
        info "Manifest contents:"
        jq . "$MANIFEST_FILE"
    else
        error "Manifest is not valid JSON"
    fi
fi

echo ""

# Test 4: Check Docker image exists
echo "=== Docker Image Validation ==="
IMAGE_NAME="nanofuse-base:latest"

if docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "^$IMAGE_NAME$"; then
    success "Docker image exists: $IMAGE_NAME"

    # Show image info
    IMAGE_SIZE=$(docker images --format "{{.Size}}" "$IMAGE_NAME")
    info "Docker image size: $IMAGE_SIZE"
else
    warning "Docker image not found: $IMAGE_NAME (may have been removed after extraction)"
fi

echo ""

# Test 5: Validate rootfs contents (if we can mount it)
if [ "$EUID" -eq 0 ] && [ -f "$ROOTFS_FILE" ]; then
    echo "=== Rootfs Contents Validation ==="

    MOUNT_POINT=$(mktemp -d)
    if mount -o loop,ro "$ROOTFS_FILE" "$MOUNT_POINT" 2>/dev/null; then
        info "Mounted rootfs at: $MOUNT_POINT"

        # Check for required directories
        REQUIRED_DIRS=("/boot" "/etc" "/var" "/usr" "/sbin")
        for dir in "${REQUIRED_DIRS[@]}"; do
            if [ -d "$MOUNT_POINT$dir" ]; then
                success "Required directory exists: $dir"
            else
                error "Required directory missing: $dir"
            fi
        done

        # Check for systemd
        if [ -f "$MOUNT_POINT/sbin/init" ]; then
            success "systemd init exists: /sbin/init"
        else
            error "systemd init not found: /sbin/init"
        fi

        # Check for kernel in /boot
        if [ -f "$MOUNT_POINT/boot/vmlinux" ]; then
            success "Kernel exists in rootfs: /boot/vmlinux"
        else
            warning "Kernel not found in rootfs /boot/vmlinux"
        fi

        # Check for SSH config
        if [ -f "$MOUNT_POINT/etc/ssh/sshd_config" ]; then
            success "SSH config exists: /etc/ssh/sshd_config"
        else
            error "SSH config not found"
        fi

        # Check for network config
        if [ -f "$MOUNT_POINT/etc/systemd/network/20-wired.network" ]; then
            success "Network config exists: /etc/systemd/network/20-wired.network"
        else
            warning "Network config not found"
        fi

        # Check for firstboot service
        if [ -f "$MOUNT_POINT/etc/systemd/system/firstboot.service" ]; then
            success "First-boot service exists: /etc/systemd/system/firstboot.service"
        else
            warning "First-boot service not found"
        fi

        # List enabled systemd services
        echo ""
        info "Enabled systemd services:"
        ls -1 "$MOUNT_POINT/etc/systemd/system/multi-user.target.wants/" 2>/dev/null | sed 's/^/  - /' || echo "  (none found)"

        umount "$MOUNT_POINT"
        rmdir "$MOUNT_POINT"
    else
        info "Skipping rootfs contents check (unable to mount)"
    fi

    echo ""
else
    info "Skipping rootfs contents validation (requires root)"
    echo ""
fi

# Summary
echo "======================================"
echo "Validation Summary"
echo "======================================"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All validations passed!${NC}"
    echo ""
    echo "Build artifacts are ready for testing:"
    echo "  - Run 'make test' to boot image in Firecracker"
    echo "  - Run 'make push' to publish to registry"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ Validations passed with $WARNINGS warning(s)${NC}"
    echo ""
    echo "Build artifacts are ready but with warnings."
    echo "Review warnings above before proceeding."
    exit 0
else
    echo -e "${RED}✗ Validation failed with $ERRORS error(s) and $WARNINGS warning(s)${NC}"
    echo ""
    echo "Fix errors above before proceeding."
    echo "Run 'make clean' and 'make build' to rebuild."
    exit 1
fi
