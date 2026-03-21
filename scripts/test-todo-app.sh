#!/bin/bash
################################################################################
# TEST TODO-APP - Check if services are responding
################################################################################
set -euo pipefail

VM_NAME="${1:-my-todo-app}"

echo "Testing VM: $VM_NAME"
echo ""

# Scan for IP by pinging the expected range
echo "Scanning for VM IP (172.16.0.10-20)..."
VM_IP=""
for i in {10..20}; do
    if timeout 0.3 ping -c 1 172.16.0.$i >/dev/null 2>&1; then
        VM_IP="172.16.0.$i"
        echo "✓ Found VM at: $VM_IP"
        break
    fi
done

if [[ -z "$VM_IP" ]]; then
    echo "ERROR: No VM responding on 172.16.0.0/24 network"
    exit 1
fi

echo ""
echo "==============================================="
echo "VM IP: $VM_IP"
echo "==============================================="
echo ""

echo "1. Ping test:"
ping -c 3 "$VM_IP"

echo ""
echo "2. Testing port 8080 (todo-backend):"
timeout 3 curl -v "http://$VM_IP:8080/health" 2>&1 | grep -E "HTTP|Connected|refused|timeout" || echo "No response on port 8080"

echo ""
echo "3. Testing port 80 (nginx):"
timeout 3 curl -v "http://$VM_IP/health" 2>&1 | grep -E "HTTP|Connected|refused|timeout" || echo "No response on port 80"

echo ""
echo "4. Port scan:"
if command -v nmap >/dev/null 2>&1; then
    nmap -p 22,80,8080 "$VM_IP" 2>/dev/null
else
    echo "nmap not installed, skipping port scan"
fi

echo ""
echo "==============================================="
echo "To see console logs, run:"
echo "  sudo cat /var/lib/nanofuse/vms/ea5bdbaa*/console.log"
echo "==============================================="
