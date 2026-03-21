#!/bin/bash
#
# Fix Unix Socket Issue - Phase 1A
#
# Problem: PrivateTmp=true in systemd service isolates /tmp, making socket inaccessible
# Solution: Change socket path from /tmp/nanofused.sock to /run/nanofused.sock
#

set -e

echo "===== Phase 1A: Fix Unix Socket Issue ====="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Error: This script must be run as root (use sudo)"
    exit 1
fi

echo "Step 1: Backup current config..."
cp /etc/nanofuse/nanofused.yaml /etc/nanofuse/nanofused.yaml.backup
echo "✅ Backed up to /etc/nanofuse/nanofused.yaml.backup"
echo ""

echo "Step 2: Update socket path in config..."
sed -i 's|socket: /tmp/nanofused.sock|socket: /run/nanofused.sock|g' /etc/nanofuse/nanofused.yaml
echo "✅ Changed socket path to /run/nanofused.sock"
echo ""

echo "Step 3: Verify config change..."
grep "socket:" /etc/nanofuse/nanofused.yaml
echo ""

echo "Step 4: Restart nanofused daemon..."
systemctl restart nanofused
sleep 2
echo "✅ Daemon restarted"
echo ""

echo "Step 5: Check daemon status..."
systemctl status nanofused --no-pager | grep -E "(Active:|socket|TCP)" || true
echo ""

echo "Step 6: Verify socket file exists..."
if [ -S /run/nanofused.sock ]; then
    ls -la /run/nanofused.sock
    echo "✅ Unix socket created successfully!"
else
    echo "❌ Socket file not found at /run/nanofused.sock"
    echo "Check daemon logs: journalctl -u nanofused -n 50"
    exit 1
fi
echo ""

echo "Step 7: Test Unix socket connectivity..."
if curl --unix-socket /run/nanofused.sock http://localhost/health -s > /dev/null 2>&1; then
    echo "✅ Unix socket is accessible!"
    curl --unix-socket /run/nanofused.sock http://localhost/health -s | jq '.'
else
    echo "❌ Failed to connect to Unix socket"
    exit 1
fi
echo ""

echo "Step 8: Test TCP connectivity (verify both work)..."
if curl http://localhost:8080/health -s > /dev/null 2>&1; then
    echo "✅ TCP listener is accessible!"
    curl http://localhost:8080/health -s | jq '.'
else
    echo "❌ Failed to connect to TCP listener"
    exit 1
fi
echo ""

echo "Step 9: Test CLI without --api-url flag..."
if su - jpoley -c "nanofuse image list" > /dev/null 2>&1; then
    echo "✅ CLI works with Unix socket!"
    su - jpoley -c "nanofuse image list"
else
    echo "❌ CLI failed to connect via socket"
    echo "Run: nanofuse --debug image list"
    exit 1
fi
echo ""

echo "======================================"
echo "✅ Phase 1A Fix Complete!"
echo "======================================"
echo ""
echo "Both TCP and Unix socket are now working:"
echo "  - TCP:         http://localhost:8080"
echo "  - Unix Socket: /run/nanofused.sock"
echo ""
echo "You can now use the CLI without --api-url flag:"
echo "  nanofuse image list"
echo "  nanofuse vm list"
echo ""
