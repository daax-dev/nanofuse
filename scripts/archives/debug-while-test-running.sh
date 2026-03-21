#!/bin/bash
# Run this in a second terminal WHILE the main test is running
# It will show you what's actually happening with port forwarding

echo "==========================================="
echo "Port Forward Debug (run while test active)"
echo "==========================================="
echo ""

echo "=== 1. Daemon Status ==="
ps aux | grep nanofused | grep -v grep || echo "No daemon running"
echo ""

echo "=== 2. VM Status ==="
./bin/nanofuse --api-url http://127.0.0.1:8080 vm list 2>/dev/null || echo "Cannot reach API"
echo ""

echo "=== 3. Daemon Logs (port forward related) ==="
grep -i "port forward\|iptables\|DNAT" /tmp/nanofuse/nanofused.log 2>/dev/null | tail -20 || echo "No logs found"
echo ""

echo "=== 4. iptables OUTPUT Chain (localhost) ==="
sudo iptables -t nat -L OUTPUT -n -v --line-numbers 2>/dev/null | head -30
echo ""

echo "=== 5. iptables PREROUTING Chain (external) ==="
sudo iptables -t nat -L PREROUTING -n -v --line-numbers 2>/dev/null | head -30
echo ""

echo "=== 6. iptables POSTROUTING Chain (SNAT/MASQUERADE) ==="
sudo iptables -t nat -L POSTROUTING -n -v --line-numbers 2>/dev/null | head -30
echo ""

echo "=== 7. iptables FORWARD Chain ==="
sudo iptables -L FORWARD -n -v --line-numbers 2>/dev/null | head -30
echo ""

echo "=== 8. Check for 8888 specifically ==="
echo "Looking for port 8888 in all chains..."
sudo iptables -t nat -L -n -v 2>/dev/null | grep 8888 || echo "No rules with port 8888 found"
echo ""

echo "=== 9. VM Details (if running) ==="
VM_STATUS=$(./bin/nanofuse --api-url http://127.0.0.1:8080 vm status test-portforward-vm 2>&1)
if echo "$VM_STATUS" | grep -q "IP Address"; then
    VM_IP=$(echo "$VM_STATUS" | grep "IP Address" | awk '{print $NF}')
    echo "VM IP: $VM_IP"

    echo ""
    echo "=== 10. Test Direct Connection to VM ==="
    curl -s --max-time 2 http://$VM_IP:8080 2>&1 || echo "Direct connection failed"

    echo ""
    echo "=== 11. Test Port Forward (localhost:8888) ==="
    curl -s --max-time 2 http://localhost:8888 2>&1 || echo "Port forward failed"
else
    echo "VM not running yet"
fi

echo ""
echo "==========================================="
echo "Done - compare this with expected behavior"
echo "==========================================="
