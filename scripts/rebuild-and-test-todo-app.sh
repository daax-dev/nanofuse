#!/bin/bash
################################################################################
# REBUILD AND TEST TODO-APP - Complete workflow
################################################################################
set -euo pipefail

echo "=========================================="
echo "REBUILD AND TEST TODO-APP"
echo "=========================================="
echo ""

# Step 1: Rebuild the image
echo "[1/6] Building fresh todo-app image..."
cd /home/jpoley/ps/nanofuse/examples/todo-app
./build-nanofuse-image.sh

IMAGE_DIGEST="sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628"

# Step 2: Delete old VM
echo ""
echo "[2/6] Cleaning up old VM..."
nanofuse --api-url http://localhost:8080 vm delete my-todo-app --force 2>/dev/null || echo "No old VM to delete"

# Step 3: Create VM with correct kernel args
echo ""
echo "[3/6] Creating VM..."
nanofuse --api-url http://localhost:8080 vm create \
    "$IMAGE_DIGEST" \
    my-todo-app \
    --vcpus 2 \
    --memory 1024 \
    --kernel-args "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init"

# Step 4: Start VM
echo ""
echo "[4/6] Starting VM..."
nanofuse --api-url http://localhost:8080 vm start my-todo-app

# Step 5: Wait for boot
echo ""
echo "[5/6] Waiting 25 seconds for boot and service startup..."
sleep 25

# Step 6: Test
echo ""
echo "[6/6] Testing services..."
/home/jpoley/ps/nanofuse/scripts/test-todo-app.sh

echo ""
echo "=========================================="
echo "To see console logs:"
echo "  sudo /home/jpoley/ps/nanofuse/scripts/full-console-log.sh"
echo "=========================================="
