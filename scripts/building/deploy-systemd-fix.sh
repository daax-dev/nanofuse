#!/bin/bash
# Deploy systemd init fix and test

set -e

echo "===== Deploying Systemd Init Fix ====="
echo ""
echo "Fix: Added init=/lib/systemd/systemd to default kernel args"
echo ""

if [ "$EUID" -ne 0 ]; then
    echo "❌ Must run as root"
    exit 1
fi

echo "Step 1: Stop daemon..."
systemctl stop nanofused
sleep 1

echo "Step 2: Backup and deploy new binary..."
cp /usr/local/bin/nanofused /usr/local/bin/nanofused.backup-systemd-$(date +%Y%m%d-%H%M%S)
cp /home/jpoley/ps/nanofuse/bin/nanofused /usr/local/bin/nanofused

echo "Step 3: Start daemon..."
systemctl start nanofused
sleep 2
systemctl status nanofused --no-pager | grep "Active:"

echo ""
echo "Step 4: Clean up old VMs..."
su - jpoley -c "nanofuse vm list --json | jq -r '.vms[].name' | xargs -I {} nanofuse vm delete {} -y 2>/dev/null || true"

echo ""
echo "Step 5: Create fresh VM..."
VM_NAME="systemd-test"
IMAGE="sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628"
su - jpoley -c "nanofuse vm create $IMAGE $VM_NAME --vcpus 2 --memory 1024"

echo ""
echo "Step 6: Start VM..."
su - jpoley -c "nanofuse vm start $VM_NAME"

echo ""
echo "Step 7: Wait for boot..."
sleep 15

echo ""
echo "Step 8: Check VM state..."
su - jpoley -c "nanofuse vm list"

echo ""
echo "Step 9: Test network..."
VM_IP=$(su - jpoley -c "nanofuse vm list --json" | jq -r '.vms[0].config.network.ip_address')
echo "VM IP: $VM_IP"
ping -c 2 $VM_IP

echo ""
echo "Step 10: Check services..."
echo "Port scan:"
nmap -p 80,8080 $VM_IP 2>&1 | grep -E "(open|closed)"

echo ""
echo "HTTP tests:"
echo -n "nginx (port 80): "
curl -s http://$VM_IP/health --max-time 3 && echo "✅ WORKS" || echo "❌ FAILED"

echo -n "backend (port 8080): "
curl -s http://$VM_IP:8080/health --max-time 3 && echo "✅ WORKS" || echo "❌ FAILED"

echo ""
echo "===== Test Complete ====="
echo ""
echo "To check console log:"
echo "  sudo ./scripts/building/check-vm-console.sh"
