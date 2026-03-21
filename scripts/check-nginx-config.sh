#!/bin/bash
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT="/mnt/check-nginx"

mkdir -p "$MOUNT"
mount -o loop,ro "$ROOTFS" "$MOUNT"

echo "=== NGINX CONFIG ==="
cat "${MOUNT}/etc/nginx/sites-available/default"

echo ""
echo "=== NGINX ERROR LOG (if exists) ==="
if [[ -f "${MOUNT}/var/log/nginx/error.log" ]]; then
    cat "${MOUNT}/var/log/nginx/error.log"
else
    echo "No error log found"
fi

echo ""
echo "=== TEST NGINX CONFIG ==="
chroot "$MOUNT" /usr/sbin/nginx -t 2>&1 || echo "Config test failed"

umount "$MOUNT"
rmdir "$MOUNT"
