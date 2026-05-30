#!/bin/bash
# Test script for default nanofuse base image
set -e

NANOFUSE="${NANOFUSE:-./bin/nanofuse}"
VM_NAME="test-default-$$"
CLEANUP=true

cleanup() {
    if [ "$CLEANUP" = true ] && [ -n "$VM_NAME" ]; then
        echo
        echo "=== Cleanup ==="
        $NANOFUSE vm rm -f "$VM_NAME" 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo "=== NanoFuse Default Image Test ==="
echo

# Check if nanofuse binary exists
if [ ! -x "$NANOFUSE" ]; then
    echo "ERROR: nanofuse binary not found at $NANOFUSE"
    exit 1
fi

# Check daemon is running
echo "1. Checking daemon..."
if ! $NANOFUSE health &>/dev/null; then
    echo "   ERROR: Daemon not responding"
    echo "   Run: sudo systemctl start nanofused"
    exit 1
fi
echo "   OK"

# Pull default image if needed
echo
echo "2. Checking default image..."
if ! $NANOFUSE image list --json | grep -q "ghcr.io/daax-dev/nanofuse/base"; then
    echo "   Pulling default image..."
    $NANOFUSE image pull --default
fi
echo "   OK"

# Create VM
echo
echo "3. Creating VM: $VM_NAME"
$NANOFUSE vm create ghcr.io/daax-dev/nanofuse/base:latest "$VM_NAME"
echo "   OK"

# Start VM
echo
echo "4. Starting VM..."
$NANOFUSE vm start "$VM_NAME"
VM_IP=$($NANOFUSE vm list --json | jq -r ".vms[] | select(.name==\"$VM_NAME\") | .config.network.ip_address")
echo "   IP: $VM_IP"

# Wait for VM to boot
echo
echo "5. Waiting for VM to boot (10s)..."
sleep 10

# Test connectivity
echo
echo "6. Testing connectivity..."

echo "   Ping:"
if ping -c 3 -W 2 "$VM_IP" &>/dev/null; then
    echo "   OK - VM responds to ping"
else
    echo "   WARN - Ping failed (might be blocked)"
fi

echo
echo "   Bridge status:"
ip link show nanofuse0 2>/dev/null | head -1 || echo "   Bridge not found"
ip addr show nanofuse0 2>/dev/null | grep inet || echo "   No IP on bridge"

echo
echo "   NAT rules:"
sudo iptables -t nat -L POSTROUTING -n 2>/dev/null | grep -E "172\.16|MASQ" | head -3 || echo "   No NAT rules found"

echo
echo "   Firecracker process:"
ps aux | grep -E "[f]irecracker.*$VM_NAME" | head -1 || ps aux | grep "[f]irecracker" | head -1 || echo "   Not found"

echo
echo "   Port scan (common ports):"
for port in 22 80 443 8080; do
    if timeout 2 bash -c "echo >/dev/tcp/$VM_IP/$port" 2>/dev/null; then
        echo "   Port $port: OPEN"
    else
        echo "   Port $port: closed"
    fi
done

# Check VM logs
echo
echo "7. VM console logs (last 20 lines):"
$NANOFUSE vm logs "$VM_NAME" --tail 20 2>/dev/null || echo "   No logs available"

echo
echo "=== Test Complete ==="
echo
echo "VM is running at $VM_IP"
echo "To keep VM running, press Ctrl+C now (within 5s)"
echo "Otherwise VM will be cleaned up..."
sleep 5
