#!/bin/bash
# Cleanup the bad GHCR push that has wrong kernel version
# This script undoes the damage and prepares for correct push

set -e

if [ "$EUID" -ne 0 ]; then
    echo "Error: Must run as root (needs access to /var/lib/nanofuse)"
    exit 1
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_info() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}!${NC} $1"
}

echo "========================================="
echo "Cleanup Bad GHCR Push & Fix"
echo "========================================="
echo ""

# Step 1: Delete ALL VMs first
echo "Step 1: Deleting ALL VMs..."
echo ""
nanofuse vm list --json | jq -r '.vms[].name' | while read -r vm; do
    log_warn "Deleting VM: $vm"
    echo "y" | nanofuse vm delete "$vm" 2>/dev/null || true
done

# Step 2: Delete ALL local images
echo ""
echo "Step 2: Deleting ALL local cached images..."
echo ""
sqlite3 /var/lib/nanofuse/nanofuse.db "SELECT digest FROM images;" | while read -r digest; do
    log_warn "Deleting $digest"
    nanofuse image remove "$digest" 2>/dev/null || true
done

# Also clean image storage directories
log_info "Cleaning image storage directories..."
rm -rf /var/lib/nanofuse/images/*

echo ""
echo "Step 3: Verify local database is clean"
echo ""
REMAINING=$(sqlite3 /var/lib/nanofuse/nanofuse.db "SELECT COUNT(*) FROM images;")
if [ "$REMAINING" -eq 0 ]; then
    log_info "Local database clean (0 images)"
else
    log_error "Still has $REMAINING images in database!"
fi

echo ""
echo "Step 4: Delete bad Docker images from GHCR"
echo ""
log_warn "Delete these packages manually via GitHub UI:"
echo "  https://github.com/peregrinesummit?tab=packages"
echo ""
echo "  - nanofuse/base:latest (wrong kernel)"
echo "  - nanofuse/base:6.1.90 (wrong kernel)"
echo "  - nanofuse/todo-app:latest (wrong kernel)"
echo "  - nanofuse/todo-app:20251119 (wrong kernel)"
echo ""

echo "Step 5: Rebuild todo-app image with kernel 6.1.90"
echo ""
cd /home/jpoley/ps/nanofuse/examples/todo-app

# Clean old build
log_info "Cleaning old build artifacts..."
rm -rf ./output

# Rebuild with correct kernel
log_info "Rebuilding todo-app image..."
./build-nanofuse-image.sh

echo ""
echo "Step 6: Verify new image has kernel 6.1.90"
echo ""
KERNEL_VERSION=$(strings ./output/vmlinux | grep "Linux version" | head -1)
echo "Kernel: $KERNEL_VERSION"

if echo "$KERNEL_VERSION" | grep -q "6.1.90"; then
    log_info "✓ Kernel 6.1.90 confirmed!"
else
    log_error "✗ Wrong kernel version!"
    exit 1
fi

echo ""
echo "Step 7: Check registered image"
echo ""
nanofuse image list

echo ""
echo "========================================="
echo "Cleanup & Rebuild Complete"
echo "========================================="
echo ""
log_info "Local images cleaned"
log_info "Todo-app rebuilt with kernel 6.1.90"
log_warn "GHCR packages need manual deletion via web UI"
echo ""
echo "Next: We need a PROPER way to push nanofuse images to GHCR"
echo "      (Not Docker images - actual kernel + rootfs artifacts)"
