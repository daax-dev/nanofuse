#!/bin/bash
################################################################################
# FIX NGINX - Debug and repair nginx service
################################################################################
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT="/tmp/nginx-fix-$$"

# Clean up on exit
cleanup() {
    umount "$MOUNT" 2>/dev/null || true
    rm -rf "$MOUNT"
}
trap cleanup EXIT

mkdir -p "$MOUNT"
mount -o loop "$ROOTFS" "$MOUNT"

echo "=== 1. Testing nginx config ==="
chroot "$MOUNT" /usr/sbin/nginx -t 2>&1 || echo "Config test failed"

echo ""
echo "=== 2. Checking nginx service file ==="
if [[ -f "${MOUNT}/lib/systemd/system/nginx.service" ]]; then
    cat "${MOUNT}/lib/systemd/system/nginx.service"
    SERVICE_PATH="${MOUNT}/lib/systemd/system/nginx.service"
elif [[ -f "${MOUNT}/usr/lib/systemd/system/nginx.service" ]]; then
    cat "${MOUNT}/usr/lib/systemd/system/nginx.service"
    SERVICE_PATH="${MOUNT}/usr/lib/systemd/system/nginx.service"
else
    echo "ERROR: nginx.service not found!"
    exit 1
fi

echo ""
echo "=== 3. Checking PID file directory ==="
# Nginx often fails if /run/nginx doesn't exist
if [[ ! -d "${MOUNT}/run/nginx" ]]; then
    echo "Creating /run/nginx directory..."
    mkdir -p "${MOUNT}/run/nginx"
    chown root:root "${MOUNT}/run/nginx"
    chmod 755 "${MOUNT}/run/nginx"
fi

echo ""
echo "=== 4. Checking log directory ==="
if [[ ! -d "${MOUNT}/var/log/nginx" ]]; then
    echo "Creating /var/log/nginx directory..."
    mkdir -p "${MOUNT}/var/log/nginx"
    chown root:root "${MOUNT}/var/log/nginx"
    chmod 755 "${MOUNT}/var/log/nginx"
fi

echo ""
echo "=== 5. Creating runtime directory setup ==="
# Create tmpfiles.d config for nginx
cat > "${MOUNT}/etc/tmpfiles.d/nginx.conf" <<'EOF'
d /run/nginx 0755 root root -
EOF

echo "Created /etc/tmpfiles.d/nginx.conf"

echo ""
echo "=== FIXES APPLIED ==="
echo "Now restart the VM to test:"
echo "  nanofuse vm delete my-todo-app --force"
echo "  nanofuse vm create sha256:0c8543... my-todo-app --vcpus 2 --memory 1024 --kernel-args 'console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init'"
echo "  nanofuse vm start my-todo-app"
echo "  sleep 20"
echo "  /home/jpoley/ps/nanofuse/scripts/test-todo-app.sh"
