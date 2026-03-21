#!/bin/bash
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT="/tmp/nginx-inspect-$$"

cleanup() {
    umount "$MOUNT" 2>/dev/null || true
    rm -rf "$MOUNT"
}
trap cleanup EXIT

mkdir -p "$MOUNT"
mount -o loop,ro "$ROOTFS" "$MOUNT"

echo "=== NGINX SERVICE FILE ==="
if [[ -f "${MOUNT}/lib/systemd/system/nginx.service" ]]; then
    cat "${MOUNT}/lib/systemd/system/nginx.service"
elif [[ -f "${MOUNT}/usr/lib/systemd/system/nginx.service" ]]; then
    cat "${MOUNT}/usr/lib/systemd/system/nginx.service"
else
    echo "ERROR: nginx.service not found!"
    exit 1
fi

echo ""
echo "=== TESTING NGINX CONFIG ==="
chroot "$MOUNT" /usr/sbin/nginx -t 2>&1

echo ""
echo "=== CHECKING REQUIRED DIRECTORIES ==="
for dir in /var/log/nginx /var/lib/nginx /run/nginx /etc/nginx; do
    if [[ -d "${MOUNT}${dir}" ]]; then
        echo "✓ $dir exists"
        ls -ld "${MOUNT}${dir}"
    else
        echo "✗ $dir MISSING"
    fi
done

echo ""
echo "=== CHECKING NGINX BINARY ==="
ls -la "${MOUNT}/usr/sbin/nginx"
file "${MOUNT}/usr/sbin/nginx"
