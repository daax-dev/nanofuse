#!/bin/bash
# Check if HTTP test server is properly installed in base image

set -e

ROOTFS="images/base/build/rootfs.ext4"
MOUNT_POINT="/tmp/nanofuse-rootfs-check"

if [ ! -f "$ROOTFS" ]; then
    echo "ERROR: Base image not found at $ROOTFS"
    echo "Run: cd images/base && sudo ./build.sh"
    exit 1
fi

echo "Checking base image rootfs..."
echo ""

# Create mount point
sudo mkdir -p "$MOUNT_POINT"

# Mount rootfs
echo "Mounting $ROOTFS..."
sudo mount -o loop,ro "$ROOTFS" "$MOUNT_POINT"

# Check for HTTP service file
echo ""
echo "1. Checking for http-test-server.service..."
if [ -f "$MOUNT_POINT/etc/systemd/system/http-test-server.service" ]; then
    echo "   ✓ Service file exists"
    echo ""
    echo "   Content:"
    sudo cat "$MOUNT_POINT/etc/systemd/system/http-test-server.service" | sed 's/^/   /'
else
    echo "   ✗ Service file NOT FOUND"
    echo "   → Base image needs to be rebuilt"
fi

# Check if service is enabled
echo ""
echo "2. Checking if service is enabled..."
if [ -L "$MOUNT_POINT/etc/systemd/system/multi-user.target.wants/http-test-server.service" ]; then
    echo "   ✓ Service is enabled (symlink exists)"
else
    echo "   ✗ Service is NOT enabled"
    echo "   → Base image needs to be rebuilt"
fi

# Check for python3
echo ""
echo "3. Checking for python3..."
if [ -f "$MOUNT_POINT/usr/bin/python3" ]; then
    echo "   ✓ python3 is installed"
else
    echo "   ✗ python3 NOT FOUND"
    echo "   → Base image needs to be rebuilt"
fi

# Check systemd services that should be enabled
echo ""
echo "4. Enabled services in multi-user.target:"
sudo ls -la "$MOUNT_POINT/etc/systemd/system/multi-user.target.wants/" | grep "\.service" | awk '{print "   - " $9}' || echo "   (none)"

# Unmount
echo ""
echo "Unmounting..."
sudo umount "$MOUNT_POINT"
sudo rmdir "$MOUNT_POINT"

echo ""
echo "=== Summary ==="
echo "If any checks failed, rebuild the base image:"
echo "  cd images/base && sudo ./build.sh && cd ../.."
echo ""
echo "The build takes ~2-3 minutes and requires sudo for ext4 operations."
