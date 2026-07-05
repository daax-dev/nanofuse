#!/bin/bash
set -e

BUILD_DIR="./build"
mkdir -p "$BUILD_DIR"

echo "Downloading official Firecracker v1.7.0 kernel..."
echo "NOTE: Pre-built 6.1.x kernels not available from Firecracker."
echo "Use ./scripts/build-kernel-docker.sh to build kernel 6.1.90 instead."

# This is the exact kernel Firecracker v1.7.0 uses in CI
# URL pattern: https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v{VERSION}/x86_64/vmlinux-{KERNEL_VERSION}

# Try 6.1 kernel (likely doesn't exist as pre-built)
KERNEL_URL="https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.7/x86_64/vmlinux-6.1"

echo "Trying: $KERNEL_URL"
if curl -fL "$KERNEL_URL" -o "$BUILD_DIR/vmlinux" 2>/dev/null; then
    if [ -s "$BUILD_DIR/vmlinux" ] && file "$BUILD_DIR/vmlinux" | grep -q "ELF"; then
        echo "Success!"
        file "$BUILD_DIR/vmlinux"
        ls -lh "$BUILD_DIR/vmlinux"
        exit 0
    fi
fi

# If that didn't work, try from GitHub releases
echo "Trying GitHub release..."
GITHUB_URL="https://github.com/firecracker-microvm/firecracker/releases/download/v1.7.0/vmlinux-6.1"

if curl -fL "$GITHUB_URL" -o "$BUILD_DIR/vmlinux" 2>/dev/null; then
    if [ -s "$BUILD_DIR/vmlinux" ] && file "$BUILD_DIR/vmlinux" | grep -q "ELF"; then
        echo "Success!"
        file "$BUILD_DIR/vmlinux"
        ls -lh "$BUILD_DIR/vmlinux"
        exit 0
    fi
fi

echo ""
echo "Failed to download pre-built kernel."
echo "Pre-built 6.1.x kernels are not available from Firecracker."
echo ""
echo "Build the kernel instead:"
echo "  cd /home/jpoley/ps/nanofuse/images/base"
echo "  sudo ./scripts/build-kernel-docker.sh"
echo ""
exit 1
