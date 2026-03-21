#!/bin/bash
# Test port forwarding with alternative services
# If HTTP doesn't work, maybe SSH or netcat will reveal the issue

set -e

VM_IP="${1:-172.16.0.11}"

echo "=== Alternative Service Port Forward Tests ==="
echo "VM IP: $VM_IP"
echo ""

# Test 1: SSH Port Forward (port 22)
echo "Test 1: SSH Port Forward"
echo "========================"
echo "SSH is guaranteed to be running in the VM (enabled in base image)"
echo ""

# Check if we can reach SSH directly
echo "  Testing direct SSH connection to VM..."
if timeout 3 nc -z "$VM_IP" 22 2>/dev/null; then
    echo "  ✓ SSH is accessible on $VM_IP:22"

    # Now test port forward 2222 -> 22
    echo ""
    echo "  Setting up port forward: localhost:2222 -> $VM_IP:22"

    # This would need to be done via API, but we can test manually
    echo "  Manual test command:"
    echo "    sudo iptables -t nat -A OUTPUT -p tcp --dport 2222 -j DNAT --to-destination $VM_IP:22"
    echo "    ssh -p 2222 root@localhost"
    echo ""
else
    echo "  ✗ SSH not accessible on $VM_IP:22"
    echo "    Waiting longer for VM to boot, or SSH failed to start"
fi

# Test 2: Simple Netcat Server
echo ""
echo "Test 2: Netcat Echo Server"
echo "=========================="
echo "Use netcat to create simplest possible server"
echo ""

echo "  To test manually:"
echo "  1. SSH into VM: ssh root@$VM_IP"
echo "  2. Start netcat server: nc -l -p 9999 -e /bin/cat"
echo "  3. Set up port forward: localhost:9999 -> $VM_IP:9999"
echo "  4. Test: echo 'hello' | nc localhost 9999"
echo ""

# Test 3: Python HTTP Server on Different Port
echo ""
echo "Test 3: Python HTTP Server (different port)"
echo "==========================================="
echo "Try HTTP on port 9090 instead of 8080"
echo ""

echo "  To test manually:"
echo "  1. SSH into VM: ssh root@$VM_IP"
echo "  2. Start HTTP server: python3 -m http.server 9090"
echo "  3. Set up port forward: localhost:9090 -> $VM_IP:9090"
echo "  4. Test: curl http://localhost:9090"
echo ""

# Test 4: socat as Alternative to iptables
echo ""
echo "Test 4: socat Port Forward (known to work)"
echo "=========================================="
echo "Use socat instead of iptables for localhost forwarding"
echo ""

if ! command -v socat &> /dev/null; then
    echo "  socat not installed"
    echo "  Install: sudo apt-get install socat"
else
    echo "  socat is installed"
    echo ""
    echo "  Test commands:"
    echo "    # Start socat forwarder"
    echo "    socat TCP-LISTEN:8888,reuseaddr,fork TCP:$VM_IP:8080 &"
    echo ""
    echo "    # Test the forward"
    echo "    curl http://localhost:8888"
    echo ""
    echo "    # Stop socat"
    echo "    pkill socat"
fi

echo ""
echo "Test 5: Direct TCP Connection Test"
echo "==================================="
echo "Use netcat to test raw TCP forwarding"
echo ""

echo "  Testing direct TCP connection to VM HTTP server..."
if timeout 3 nc -z "$VM_IP" 8080 2>/dev/null; then
    echo "  ✓ Port 8080 is open on $VM_IP"

    echo ""
    echo "  Sending HTTP request via netcat:"
    echo "  echo -e 'GET / HTTP/1.0\r\n\r\n' | nc $VM_IP 8080"
    echo ""
    echo -e 'GET / HTTP/1.0\r\n\r\n' | timeout 2 nc "$VM_IP" 8080 | head -10 || echo "  (no response)"
else
    echo "  ✗ Port 8080 is not open on $VM_IP"
fi

echo ""
echo "=== Summary ==="
echo ""
echo "Next steps:"
echo "1. If SSH works: Port forward SSH first, verify that works"
echo "2. If netcat works: Problem is specific to HTTP"
echo "3. If socat works: Use socat instead of iptables for localhost"
echo "4. If nothing works: Problem is with DNAT/routing, not service"
echo ""
echo "Most Likely Solution:"
echo "  Use socat for localhost port forwarding (more reliable than iptables)"
echo "  Keep iptables for external connections (those work better)"
