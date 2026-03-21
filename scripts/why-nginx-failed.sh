#!/bin/bash
# Quick check why nginx failed in the rootfs

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT="/mnt/nginx-debug"

mkdir -p "$MOUNT"
mount -o loop,ro "$ROOTFS" "$MOUNT"

echo "=== Testing nginx config ==="
chroot "$MOUNT" /usr/sbin/nginx -t 2>&1

echo ""
echo "=== Checking nginx binary ==="
ls -la "${MOUNT}/usr/sbin/nginx"

echo ""
echo "=== Checking if nginx can bind to port 80 ==="
echo "Nginx needs CAP_NET_BIND_SERVICE or to run as root to bind to port 80"
echo "In systemd service, check User= directive"

echo ""
echo "=== Service file ==="
cat "${MOUNT}/lib/systemd/system/nginx.service" 2>/dev/null || cat "${MOUNT}/usr/lib/systemd/system/nginx.service" 2>/dev/null || echo "Service file not found"

umount "$MOUNT"
rmdir "$MOUNT"
