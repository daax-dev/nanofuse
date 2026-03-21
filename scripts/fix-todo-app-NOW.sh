#!/bin/bash
################################################################################
# FIX TODO-APP - ACTUALLY GET IT WORKING
################################################################################
set -euo pipefail

echo "=== FIXING TODO-APP ROOTFS AND CREATING WORKING VM ==="

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT="/mnt/todo-fix"

# 1. Mount rootfs
echo "[1/6] Mounting rootfs..."
mkdir -p "$MOUNT"
if mount | grep -q "$MOUNT"; then
    umount "$MOUNT" || true
fi
mount -o loop "$ROOTFS" "$MOUNT"

# 2. Fix /sbin/init symlink
echo "[2/6] Creating /sbin/init symlink..."
if [[ ! -L "${MOUNT}/sbin/init" ]]; then
    ln -sf /lib/systemd/systemd "${MOUNT}/sbin/init"
    echo "  ✓ Created /sbin/init -> /lib/systemd/systemd"
else
    echo "  ✓ /sbin/init already exists"
fi

# 3. Set default target
echo "[3/6] Setting default target to multi-user..."
if [[ ! -L "${MOUNT}/etc/systemd/system/default.target" ]]; then
    ln -sf /lib/systemd/system/multi-user.target "${MOUNT}/etc/systemd/system/default.target"
    echo "  ✓ Set default.target -> multi-user.target"
else
    echo "  ✓ default.target already set"
fi

# 4. Enable services
echo "[4/6] Enabling services..."
mkdir -p "${MOUNT}/etc/systemd/system/multi-user.target.wants"

if [[ ! -L "${MOUNT}/etc/systemd/system/multi-user.target.wants/todo-backend.service" ]]; then
    ln -sf /etc/systemd/system/todo-backend.service \
        "${MOUNT}/etc/systemd/system/multi-user.target.wants/todo-backend.service"
    echo "  ✓ Enabled todo-backend.service"
else
    echo "  ✓ todo-backend.service already enabled"
fi

if [[ ! -L "${MOUNT}/etc/systemd/system/multi-user.target.wants/nginx.service" ]]; then
    ln -sf /lib/systemd/system/nginx.service \
        "${MOUNT}/etc/systemd/system/multi-user.target.wants/nginx.service"
    echo "  ✓ Enabled nginx.service"
else
    echo "  ✓ nginx.service already enabled"
fi

# 5. Fix permissions
echo "[5/6] Fixing binary permissions..."
if [[ -f "${MOUNT}/usr/local/bin/todo-server" ]]; then
    chmod 755 "${MOUNT}/usr/local/bin/todo-server"
    chown root:root "${MOUNT}/usr/local/bin/todo-server"
    echo "  ✓ todo-server permissions fixed"
else
    echo "  ⚠ todo-server binary not found"
fi

# 6. Unmount
echo "[6/6] Unmounting..."
umount "$MOUNT"
rmdir "$MOUNT"

echo ""
echo "=== ROOTFS FIXED ==="
echo ""
echo "Now creating VM with correct kernel args..."

# Delete old VMs
nanofuse --api-url http://localhost:8080 vm delete todo-app-v2 2>/dev/null || true
nanofuse --api-url http://localhost:8080 vm delete my-todo-app 2>/dev/null || true
sleep 2

# Create VM with proper kernel args for Firecracker + systemd
nanofuse --api-url http://localhost:8080 vm create \
    sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628 \
    my-todo-app \
    --vcpus 2 \
    --memory 1024 \
    --kernel-args "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init"

echo "Starting VM..."
nanofuse --api-url http://localhost:8080 vm start my-todo-app

echo ""
echo "Waiting 20 seconds for boot and service startup..."
sleep 20

# Get IP and test
VM_IP=$(nanofuse --api-url http://localhost:8080 vm list --json | jq -r '.vms[] | select(.name=="my-todo-app") | .config.network.ip_address')

echo ""
echo "==============================================="
echo "VM IP: $VM_IP"
echo "==============================================="
echo ""

if [[ "$VM_IP" == "null" ]] || [[ -z "$VM_IP" ]]; then
    echo "ERROR: VM has no IP address!"
    echo "Checking VM status..."
    nanofuse --api-url http://localhost:8080 vm list
    exit 1
fi

echo "Testing connectivity..."
echo ""
echo "1. Ping test:"
ping -c 3 "$VM_IP"

echo ""
echo "2. Port 8080 (backend):"
curl -v "http://$VM_IP:8080/health" 2>&1 | head -20

echo ""
echo "3. Port 80 (nginx):"
curl -v "http://$VM_IP/health" 2>&1 | head -20

echo ""
echo "==============================================="
echo "If you see successful HTTP responses above,"
echo "the services are WORKING!"
echo "==============================================="
