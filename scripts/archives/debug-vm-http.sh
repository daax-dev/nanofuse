#!/bin/bash
# Debug why HTTP server isn't starting in VM

set -e

API_URL="http://127.0.0.1:8080"
CLI_BIN="./bin/nanofuse"

# Get the port forward test VM name
VM_NAME="test-portforward-vm"

echo "=== VM HTTP Server Debug ==="
echo ""

# Check if VM exists
if ! $CLI_BIN --api-url "$API_URL" vm status "$VM_NAME" >/dev/null 2>&1; then
    echo "ERROR: VM '$VM_NAME' not found"
    echo "Run the test first: sudo ./scripts/test-network-e2e.sh"
    echo "Then run this script while the VM is still running (before cleanup)"
    exit 1
fi

# Get VM info
echo "1. VM Status:"
$CLI_BIN --api-url "$API_URL" vm status "$VM_NAME" || true

echo ""
echo "2. Full VM logs (looking for http-test-server):"
echo "---"
$CLI_BIN --api-url "$API_URL" vm logs "$VM_NAME" | grep -i "http-test-server" || echo "(no http-test-server logs found)"

echo ""
echo "3. Systemd startup logs:"
echo "---"
$CLI_BIN --api-url "$API_URL" vm logs "$VM_NAME" | grep -E "(systemd|Started|Failed|failed|error)" | tail -30

echo ""
echo "4. Network initialization logs:"
echo "---"
$CLI_BIN --api-url "$API_URL" vm logs "$VM_NAME" | grep -E "(network|eth|ens|ip|dhcp)" | tail -20

echo ""
echo "5. Last 50 lines of VM console:"
echo "---"
$CLI_BIN --api-url "$API_URL" vm logs "$VM_NAME" --tail 50

echo ""
echo "=== Analysis Tips ==="
echo "Look for:"
echo "  - 'Started NanoFuse HTTP Test Server' = service started successfully"
echo "  - 'Failed to start' = service failed"
echo "  - 'network-online.target' = network dependency"
echo "  - Python errors = python3 issue"
