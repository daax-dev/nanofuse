#!/bin/bash
# Comprehensive port forwarding debug with tcpdump
# This script captures packets at every stage to find where forwarding breaks

set -e

VM_IP="${1:-172.16.0.11}"
HOST_PORT="${2:-8888}"
VM_PORT="${3:-8080}"

echo "=== Port Forward Packet Tracing ==="
echo "VM IP: $VM_IP"
echo "Port forward: localhost:$HOST_PORT -> $VM_IP:$VM_PORT"
echo ""

# Check prerequisites
if ! command -v tcpdump &> /dev/null; then
    echo "ERROR: tcpdump not installed"
    echo "Install: sudo apt-get install tcpdump"
    exit 1
fi

# Find TAP device for this VM
TAP_DEVICE=$(ip link show | grep "tap-" | head -1 | awk -F: '{print $2}' | tr -d ' ')
if [ -z "$TAP_DEVICE" ]; then
    echo "WARNING: No TAP device found, using nanofuse0 only"
    TAP_DEVICE="none"
fi

echo "Devices to monitor:"
echo "  - lo (localhost)"
echo "  - nanofuse0 (bridge)"
echo "  - $TAP_DEVICE (VM tap device)"
echo ""

# Create temp directory for captures
CAPTURE_DIR="/tmp/nanofuse-pcap-$(date +%s)"
mkdir -p "$CAPTURE_DIR"
echo "Packet captures will be saved to: $CAPTURE_DIR"
echo ""

# Start tcpdump on all relevant interfaces
echo "Starting packet capture on all interfaces..."
echo "Press Ctrl+C after curl completes to stop capture"
echo ""

# Capture on loopback
sudo tcpdump -i lo -n "port $HOST_PORT or port $VM_PORT" -w "$CAPTURE_DIR/lo.pcap" 2>/dev/null &
LO_PID=$!

# Capture on bridge
sudo tcpdump -i nanofuse0 -n "host $VM_IP and (port $HOST_PORT or port $VM_PORT)" -w "$CAPTURE_DIR/bridge.pcap" 2>/dev/null &
BRIDGE_PID=$!

# Capture on TAP device if it exists
if [ "$TAP_DEVICE" != "none" ]; then
    sudo tcpdump -i "$TAP_DEVICE" -n "port $VM_PORT" -w "$CAPTURE_DIR/tap.pcap" 2>/dev/null &
    TAP_PID=$!
fi

# Give tcpdump time to start
sleep 2

echo "Packet capture running..."
echo ""
echo "=== Test 1: Direct connection to VM (should work) ==="
echo "Command: curl -v -m 2 http://$VM_IP:$VM_PORT/"
echo ""
curl -v -m 2 "http://$VM_IP:$VM_PORT/" 2>&1 | head -15 || echo "FAILED"

echo ""
echo ""
echo "=== Test 2: Port forward from localhost (currently failing) ==="
echo "Command: curl -v -m 2 http://localhost:$HOST_PORT/"
echo ""
curl -v -m 2 "http://localhost:$HOST_PORT/" 2>&1 | head -15 || echo "FAILED"

echo ""
echo ""
echo "Stopping packet capture..."
sleep 1

# Stop all tcpdump processes
sudo kill $LO_PID 2>/dev/null || true
sudo kill $BRIDGE_PID 2>/dev/null || true
[ "$TAP_DEVICE" != "none" ] && sudo kill $TAP_PID 2>/dev/null || true

sleep 1

echo ""
echo "=== Packet Capture Analysis ==="
echo ""

# Analyze loopback capture
echo "1. Loopback interface (lo):"
echo "   Looking for traffic on port $HOST_PORT..."
LO_COUNT=$(sudo tcpdump -r "$CAPTURE_DIR/lo.pcap" 2>/dev/null | wc -l)
if [ "$LO_COUNT" -gt 0 ]; then
    echo "   ✓ Found $LO_COUNT packets on loopback"
    echo "   First 10 packets:"
    sudo tcpdump -r "$CAPTURE_DIR/lo.pcap" -n 2>/dev/null | head -10 | sed 's/^/     /'
else
    echo "   ✗ NO packets captured on loopback"
    echo "   → curl might not even be sending packets"
fi

echo ""
echo "2. Bridge interface (nanofuse0):"
echo "   Looking for traffic to $VM_IP..."
BRIDGE_COUNT=$(sudo tcpdump -r "$CAPTURE_DIR/bridge.pcap" 2>/dev/null | wc -l)
if [ "$BRIDGE_COUNT" -gt 0 ]; then
    echo "   ✓ Found $BRIDGE_COUNT packets on bridge"
    echo "   First 10 packets:"
    sudo tcpdump -r "$CAPTURE_DIR/bridge.pcap" -n 2>/dev/null | head -10 | sed 's/^/     /'

    # Check for SYN packets
    SYN_COUNT=$(sudo tcpdump -r "$CAPTURE_DIR/bridge.pcap" 'tcp[tcpflags] & tcp-syn != 0' 2>/dev/null | wc -l)
    echo "   SYN packets: $SYN_COUNT"

    # Check for ACK packets (would indicate successful connection)
    ACK_COUNT=$(sudo tcpdump -r "$CAPTURE_DIR/bridge.pcap" 'tcp[tcpflags] & tcp-ack != 0' 2>/dev/null | wc -l)
    echo "   ACK packets: $ACK_COUNT"
else
    echo "   ✗ NO packets captured on bridge"
    echo "   → Packets not reaching bridge from localhost"
    echo "   → DNAT in OUTPUT chain might not be working"
fi

if [ "$TAP_DEVICE" != "none" ]; then
    echo ""
    echo "3. TAP device ($TAP_DEVICE):"
    echo "   Looking for traffic to port $VM_PORT..."
    TAP_COUNT=$(sudo tcpdump -r "$CAPTURE_DIR/tap.pcap" 2>/dev/null | wc -l)
    if [ "$TAP_COUNT" -gt 0 ]; then
        echo "   ✓ Found $TAP_COUNT packets on TAP device"
        echo "   First 10 packets:"
        sudo tcpdump -r "$CAPTURE_DIR/tap.pcap" -n 2>/dev/null | head -10 | sed 's/^/     /'
    else
        echo "   ✗ NO packets captured on TAP device"
        echo "   → Bridge not forwarding to TAP"
    fi
fi

echo ""
echo "=== Routing and iptables Analysis ==="
echo ""

echo "4. Routing table:"
echo "   Routes to 172.16.0.0/24:"
ip route show | grep "172.16.0" | sed 's/^/     /' || echo "     (no routes found)"

echo ""
echo "5. iptables NAT rules for port $HOST_PORT:"
echo "   PREROUTING chain:"
sudo iptables -t nat -L PREROUTING -n -v | grep "$HOST_PORT" | sed 's/^/     /' || echo "     (no rules)"

echo ""
echo "   OUTPUT chain:"
sudo iptables -t nat -L OUTPUT -n -v | grep "$HOST_PORT" | sed 's/^/     /' || echo "     (no rules)"

echo ""
echo "   POSTROUTING chain (to $VM_IP):"
sudo iptables -t nat -L POSTROUTING -n -v | grep "$VM_IP" | sed 's/^/     /' || echo "     (no rules)"

echo ""
echo "6. Connection tracking:"
echo "   Active connections for port $HOST_PORT or $VM_PORT:"
sudo conntrack -L 2>/dev/null | grep -E "($HOST_PORT|$VM_PORT)" | sed 's/^/     /' || echo "     (no connections tracked)"

echo ""
echo "7. Kernel parameters:"
echo "   IP forwarding: $(sysctl -n net.ipv4.ip_forward)"
echo "   RP filter (all): $(sysctl -n net.ipv4.conf.all.rp_filter)"
echo "   RP filter (nanofuse0): $(sysctl -n net.ipv4.conf.nanofuse0.rp_filter 2>/dev/null || echo 'N/A')"

if [ -e /proc/sys/net/bridge/bridge-nf-call-iptables ]; then
    echo "   Bridge netfilter: $(sysctl -n net.bridge.bridge-nf-call-iptables)"
fi

echo ""
echo "=== Diagnosis ==="
echo ""

if [ "$LO_COUNT" -eq 0 ]; then
    echo "❌ PROBLEM: No packets on loopback"
    echo "   → curl is not sending to localhost:$HOST_PORT"
    echo "   → Check if curl is actually running"

elif [ "$BRIDGE_COUNT" -eq 0 ]; then
    echo "❌ PROBLEM: Packets on loopback but NOT on bridge"
    echo "   → DNAT in OUTPUT chain is not working"
    echo "   → OR routing to 172.16.0.0/24 is missing"
    echo ""
    echo "   Fixes to try:"
    echo "   1. Check iptables OUTPUT rule: sudo iptables -t nat -L OUTPUT -n -v"
    echo "   2. Add explicit route: sudo ip route add 172.16.0.0/24 dev nanofuse0"
    echo "   3. Try matching on dest: iptables -t nat -A OUTPUT -d 127.0.0.1 -p tcp --dport $HOST_PORT -j DNAT --to $VM_IP:$VM_PORT"

elif [ "$TAP_DEVICE" != "none" ] && [ "$TAP_COUNT" -eq 0 ]; then
    echo "❌ PROBLEM: Packets on bridge but NOT on TAP device"
    echo "   → Bridge is not forwarding to TAP"
    echo "   → Check bridge configuration"
    echo ""
    echo "   Fixes to try:"
    echo "   1. Check bridge ports: bridge link show"
    echo "   2. Check FORWARD chain: sudo iptables -L FORWARD -n -v"
    echo "   3. Check bridge filtering: sudo sysctl net.bridge.bridge-nf-call-iptables"

else
    echo "✓ Packets are flowing through the network stack"
    echo ""
    if [ "$ACK_COUNT" -eq 0 ]; then
        echo "⚠️  But no ACK packets seen"
        echo "   → VM might not be responding"
        echo "   → OR response packets are not routing back"
        echo ""
        echo "   Fixes to try:"
        echo "   1. Check SNAT/MASQUERADE rule in POSTROUTING"
        echo "   2. Check if HTTP server is actually listening in VM"
        echo "   3. Try tcpdump inside VM: ssh root@$VM_IP tcpdump -i eth0 port $VM_PORT"
    else
        echo "✓ Connection seems to establish (ACKs present)"
        echo "   → Problem might be at application layer"
        echo "   → Check HTTP server logs in VM"
    fi
fi

echo ""
echo "=== Saved Files ==="
echo "PCAP files saved to: $CAPTURE_DIR"
echo ""
echo "To analyze further:"
echo "  sudo tcpdump -r $CAPTURE_DIR/lo.pcap -n -vv"
echo "  sudo tcpdump -r $CAPTURE_DIR/bridge.pcap -n -vv"
[ "$TAP_DEVICE" != "none" ] && echo "  sudo tcpdump -r $CAPTURE_DIR/tap.pcap -n -vv"
echo ""
echo "To open in Wireshark:"
echo "  wireshark $CAPTURE_DIR/lo.pcap"
echo ""
echo "=== End of Diagnosis ==="
