#!/bin/bash
################################################################################
# FIX NGINX - Enable and verify nginx service
################################################################################
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT="/mnt/nginx-fix"

echo "=== FIXING NGINX IN ROOTFS ==="

mkdir -p "$MOUNT"
if mount | grep -q "$MOUNT"; then
    umount "$MOUNT" || true
fi

mount -o loop "$ROOTFS" "$MOUNT"

echo "Checking nginx installation..."
if [[ -f "${MOUNT}/usr/sbin/nginx" ]]; then
    echo "✓ nginx binary exists"
else
    echo "✗ nginx binary NOT FOUND"
    umount "$MOUNT"
    exit 1
fi

echo "Checking nginx service file..."
if [[ -f "${MOUNT}/lib/systemd/system/nginx.service" ]]; then
    echo "✓ nginx.service exists at /lib/systemd/system/"
    cat "${MOUNT}/lib/systemd/system/nginx.service"
elif [[ -f "${MOUNT}/usr/lib/systemd/system/nginx.service" ]]; then
    echo "✓ nginx.service exists at /usr/lib/systemd/system/"
    cat "${MOUNT}/usr/lib/systemd/system/nginx.service"
else
    echo "✗ nginx.service NOT FOUND"
fi

echo ""
echo "Enabling nginx service..."
mkdir -p "${MOUNT}/etc/systemd/system/multi-user.target.wants"

# Find where nginx.service actually is
if [[ -f "${MOUNT}/lib/systemd/system/nginx.service" ]]; then
    SERVICE_PATH="/lib/systemd/system/nginx.service"
elif [[ -f "${MOUNT}/usr/lib/systemd/system/nginx.service" ]]; then
    SERVICE_PATH="/usr/lib/systemd/system/nginx.service"
else
    echo "ERROR: Cannot find nginx.service"
    umount "$MOUNT"
    exit 1
fi

# Remove old symlink if exists
rm -f "${MOUNT}/etc/systemd/system/multi-user.target.wants/nginx.service"

# Create new symlink
ln -sf "$SERVICE_PATH" "${MOUNT}/etc/systemd/system/multi-user.target.wants/nginx.service"
echo "✓ Created symlink: multi-user.target.wants/nginx.service -> $SERVICE_PATH"

echo ""
echo "Verifying symlinks..."
ls -la "${MOUNT}/etc/systemd/system/multi-user.target.wants/" | grep -E "nginx|todo-backend"

umount "$MOUNT"
rmdir "$MOUNT"

echo ""
echo "=== NGINX FIXED ==="
echo ""
echo "Now restart the VM:"
echo "  nanofuse vm stop my-todo-app"
echo "  nanofuse vm start my-todo-app"
echo "  /home/jpoley/ps/nanofuse/scripts/test-todo-app.sh"
