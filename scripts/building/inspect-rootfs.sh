#!/bin/bash
# Inspect rootfs to diagnose missing systemd issue

set -e

ROOTFS_FILE="${1:-/home/jpoley/ps/nanofuse/examples/todo-app/output/rootfs.ext4}"

if [ ! -f "$ROOTFS_FILE" ]; then
    echo "Error: Rootfs file not found: $ROOTFS_FILE"
    exit 1
fi

echo "========================================="
echo "Inspecting Rootfs: $ROOTFS_FILE"
echo "========================================="
echo ""

MOUNT_POINT="/tmp/nanofuse-rootfs-inspect"
mkdir -p "$MOUNT_POINT"

echo "Mounting rootfs..."
mount -o loop "$ROOTFS_FILE" "$MOUNT_POINT"

echo ""
echo "=== Root directory structure ==="
ls -la "$MOUNT_POINT/" | head -20

echo ""
echo "=== Checking for systemd ==="
echo "Looking for /lib/systemd/systemd:"
if [ -f "$MOUNT_POINT/lib/systemd/systemd" ]; then
    echo "✓ FOUND: /lib/systemd/systemd"
    ls -la "$MOUNT_POINT/lib/systemd/systemd"
else
    echo "✗ NOT FOUND: /lib/systemd/systemd"
fi

echo ""
echo "Looking for /usr/lib/systemd/systemd:"
if [ -f "$MOUNT_POINT/usr/lib/systemd/systemd" ]; then
    echo "✓ FOUND: /usr/lib/systemd/systemd"
    ls -la "$MOUNT_POINT/usr/lib/systemd/systemd"
else
    echo "✗ NOT FOUND: /usr/lib/systemd/systemd"
fi

echo ""
echo "Looking for /sbin/init:"
if [ -f "$MOUNT_POINT/sbin/init" ] || [ -L "$MOUNT_POINT/sbin/init" ]; then
    echo "✓ FOUND: /sbin/init"
    ls -la "$MOUNT_POINT/sbin/init"
else
    echo "✗ NOT FOUND: /sbin/init"
fi

echo ""
echo "=== Checking /lib directory ==="
if [ -d "$MOUNT_POINT/lib" ]; then
    echo "/lib exists:"
    ls -la "$MOUNT_POINT/lib/" | head -10
else
    echo "/lib DOES NOT EXIST"
fi

echo ""
echo "=== Checking if /lib is a symlink ==="
if [ -L "$MOUNT_POINT/lib" ]; then
    echo "/lib is a symlink to:"
    readlink "$MOUNT_POINT/lib"
fi

echo ""
echo "=== Checking /usr/lib/systemd ==="
if [ -d "$MOUNT_POINT/usr/lib/systemd" ]; then
    echo "/usr/lib/systemd exists:"
    ls -la "$MOUNT_POINT/usr/lib/systemd/" | head -10
else
    echo "/usr/lib/systemd DOES NOT EXIST"
fi

echo ""
echo "=== Checking for nginx ==="
if [ -f "$MOUNT_POINT/usr/sbin/nginx" ]; then
    echo "✓ nginx found"
else
    echo "✗ nginx NOT found"
fi

echo ""
echo "=== Checking for todo-server ==="
if [ -f "$MOUNT_POINT/usr/local/bin/todo-server" ]; then
    echo "✓ todo-server found"
else
    echo "✗ todo-server NOT found"
fi

echo ""
echo "=== Checking systemd service files ==="
if [ -d "$MOUNT_POINT/etc/systemd/system" ]; then
    echo "Service files in /etc/systemd/system:"
    ls -la "$MOUNT_POINT/etc/systemd/system/" | grep -E "\.service|\.target"
else
    echo "/etc/systemd/system DOES NOT EXIST"
fi

echo ""
echo "Unmounting..."
umount "$MOUNT_POINT"
rmdir "$MOUNT_POINT"

echo ""
echo "========================================="
echo "Inspection Complete"
echo "========================================="
