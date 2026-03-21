#!/bin/bash
set -euo pipefail

# BUILD_AND_TEST.sh - Build kernel and test it
# This script should be run WITH sudo
# It builds the kernel as root, then drops privileges to test

cd /home/jpoley/src/_mine/nanofuse/images/base

echo "=========================================="
echo "BUILD AND TEST KERNEL"
echo "=========================================="
echo ""

# Check if running with sudo
if [ -z "${SUDO_USER:-}" ]; then
    echo "ERROR: This script must be run with sudo"
    echo "Run: sudo ./BUILD_AND_TEST.sh"
    exit 1
fi

echo "[1/2] Building kernel (as root)..."
echo ""

# Remove old kernel if it exists
rm -f /tmp/vmlinux-test 2>/dev/null || true

docker image rm nanofuse-kernel-builder:latest 2>/dev/null || true
docker build --no-cache -f Dockerfile.kernel -t nanofuse-kernel-builder:latest .
docker run --rm nanofuse-kernel-builder:latest cat /vmlinux > /tmp/vmlinux-test

# Fix ownership to original user
chown ${SUDO_UID}:${SUDO_GID} /tmp/vmlinux-test

echo ""
echo "[2/2] Testing kernel (as $SUDO_USER)..."
echo ""

# Drop privileges and run test as the original user
# This is critical - Firecracker doesn't work properly under sudo
exec sudo -u "$SUDO_USER" ./test-boot.sh --verbose --check-virtio /tmp/vmlinux-test
