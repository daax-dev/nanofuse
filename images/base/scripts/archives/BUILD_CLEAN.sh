#!/bin/bash
set -euo pipefail

# COMPLETE CLEAN BUILD - Fresh from scratch
# No docker cache, full rebuild of everything

echo "=========================================="
echo "CLEAN KERNEL BUILD - Fresh from scratch"
echo "=========================================="
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Step 1: Clean Docker completely
echo "[1/3] Cleaning Docker cache and old images..."
docker system prune -af
docker image rm nanofuse-kernel:test 2>/dev/null || true
echo "✓ Docker cleaned"
echo ""

# Step 2: Build kernel fresh
echo "[2/3] Building kernel from scratch..."
echo "  - Cloning Linux 6.1.90"
echo "  - Downloading Firecracker config"
echo "  - Enabling CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y"
echo "  - Full kernel build (not just bzImage)"
echo ""

if docker build -f Dockerfile.kernel -t nanofuse-kernel:test . ; then
    echo "✓ Kernel built successfully"
else
    echo "✗ Build failed"
    exit 1
fi
echo ""

# Step 3: Extract vmlinux
echo "[3/3] Extracting vmlinux..."
KERNEL_PATH="/tmp/vmlinux-fresh-build"
docker run --rm nanofuse-kernel:test cat /vmlinux > "$KERNEL_PATH"
echo "✓ Kernel extracted to: $KERNEL_PATH"
file "$KERNEL_PATH"
echo ""

# Verify config in .config file
echo "Configuration verification:"
docker run --rm nanofuse-kernel:test sh -c "grep CONFIG_VIRTIO_MMIO .config"
echo ""

echo "=========================================="
echo "BUILD COMPLETE"
echo "=========================================="
echo "Kernel location: $KERNEL_PATH"
echo ""
