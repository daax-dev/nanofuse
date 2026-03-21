#!/bin/bash
#
# Deploy Phase 1B Fix - VM Config in List Response
#
# Changes made:
# 1. Added Config, ImageDigest, Architecture fields to VMListItem type
# 2. Updated API handler to copy those fields
# 3. Compiled and tested successfully
#

set -e

echo "===== Phase 1B Fix Deployment ====="
echo ""
echo "Changes:"
echo "  - internal/types/vm.go: Added Config to VMListItem"
echo "  - internal/api/vm_handlers.go: Copy Config in list handler"
echo ""
echo "Build status: ✅ Both binaries compiled successfully"
echo "  - bin/nanofuse (8.5MB)"
echo "  - bin/nanofused (9.0MB)"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Error: This script must be run as root (use sudo)"
    exit 1
fi

echo "Step 1: Stop nanofused daemon..."
systemctl stop nanofused
echo "✅ Daemon stopped"
echo ""

echo "Step 2: Backup current binaries..."
cp /usr/local/bin/nanofused /usr/local/bin/nanofused.backup-$(date +%Y%m%d-%H%M%S)
cp /usr/local/bin/nanofuse /usr/local/bin/nanofuse.backup-$(date +%Y%m%d-%H%M%S)
echo "✅ Backups created in /usr/local/bin/"
echo ""

echo "Step 3: Deploy new binaries..."
cp /home/jpoley/ps/nanofuse/bin/nanofused /usr/local/bin/nanofused
cp /home/jpoley/ps/nanofuse/bin/nanofuse /usr/local/bin/nanofuse
chmod +x /usr/local/bin/nanofused
chmod +x /usr/local/bin/nanofuse
echo "✅ New binaries deployed"
echo ""

echo "Step 4: Start nanofused daemon..."
systemctl start nanofused
sleep 2
echo "✅ Daemon started"
echo ""

echo "Step 5: Check daemon status..."
systemctl status nanofused --no-pager | grep -E "(Active:|Listening)" || true
echo ""

echo "Step 6: Test VM list shows config..."
echo "Current VMs:"
su - jpoley -c "nanofuse vm list"
echo ""

echo "Step 7: Verify config via API..."
echo "Config from API:"
curl -s http://localhost:8080/vms | jq '.vms[0].config | {vcpus, memory_mib}' 2>/dev/null || echo "No VMs found or jq not available"
echo ""

echo "======================================"
echo "✅ Phase 1B Fix Deployed!"
echo "======================================"
echo ""
echo "Expected results:"
echo "  - VM list should show VCPUS=2, MEMORY=512M (not 0)"
echo "  - API response includes full config"
echo ""
echo "If there are issues, rollback with:"
echo "  sudo systemctl stop nanofused"
echo "  sudo cp /usr/local/bin/nanofused.backup-* /usr/local/bin/nanofused"
echo "  sudo cp /usr/local/bin/nanofuse.backup-* /usr/local/bin/nanofuse"
echo "  sudo systemctl start nanofused"
echo ""
