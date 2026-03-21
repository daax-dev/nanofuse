#!/bin/bash
# Minimal port forwarding debug script

set -e

VM_IP="172.16.0.11"  # Replace with actual VM IP
HOST_PORT="8888"
VM_PORT="8080"

echo "=== Port Forward Debug ==="
echo "VM IP: $VM_IP"
echo "Port forward: localhost:$HOST_PORT -> $VM_IP:$VM_PORT"
echo ""

# Check if HTTP server is responding on VM directly
echo "1. Testing direct connection to VM..."
if curl -s -m 2 "http://$VM_IP:$VM_PORT/" >/dev/null 2>&1; then
    echo "   ✓ VM HTTP server is responding"
else
    echo "   ✗ VM HTTP server not responding - test will fail"
    exit 1
fi

# Show current iptables rules
echo ""
echo "2. Current iptables NAT rules:"
echo "   PREROUTING chain:"
sudo iptables -t nat -L PREROUTING -n -v | grep "$HOST_PORT" || echo "   (no rules for port $HOST_PORT)"

echo ""
echo "   OUTPUT chain:"
sudo iptables -t nat -L OUTPUT -n -v | grep "$HOST_PORT" || echo "   (no rules for port $HOST_PORT)"

echo ""
echo "   POSTROUTING chain:"
sudo iptables -t nat -L POSTROUTING -n -v | grep "$VM_IP" || echo "   (no rules for $VM_IP)"

echo ""
echo "   FORWARD chain:"
sudo iptables -L FORWARD -n -v | grep "$VM_IP.*$VM_PORT" || echo "   (no rules for $VM_IP:$VM_PORT)"

# Test port forward from localhost
echo ""
echo "3. Testing port forward from localhost..."
echo "   Command: curl -v http://localhost:$HOST_PORT/"
curl -v -m 5 "http://localhost:$HOST_PORT/" 2>&1 | head -20

echo ""
echo "=== Debug Tips ==="
echo "If connection refused:"
echo "  - Check OUTPUT chain rule exists"
echo "  - Check MASQUERADE rule for return traffic"
echo ""
echo "If connection timeout:"
echo "  - Packet might be reaching VM but response not routing back"
echo "  - Check: sudo tcpdump -i nanofuse0 port $VM_PORT"
echo ""
echo "To watch traffic:"
echo "  Terminal 1: sudo tcpdump -i any port $HOST_PORT -n"
echo "  Terminal 2: curl http://localhost:$HOST_PORT/"
