#!/bin/bash
# Run all layers of testing systematically
# Based on docs/building/COMPREHENSIVE_TESTING_PLAN.md

VM_ID="$1"

if [ -z "$VM_ID" ]; then
    echo "Usage: $0 <vm-id>"
    echo ""
    echo "Find VM ID with: nanofuse vm list"
    exit 1
fi

CONSOLE_LOG="/var/lib/nanofuse/vms/$VM_ID/console.log"

if [ ! -f "$CONSOLE_LOG" ]; then
    echo "Error: Console log not found at $CONSOLE_LOG"
    exit 1
fi

echo "========================================="
echo "Comprehensive VM Testing"
echo "VM ID: $VM_ID"
echo "========================================="
echo ""

echo "=== Layer 0: Kernel & Boot ==="
echo ""
echo "Kernel Version:"
sudo grep "Linux version" "$CONSOLE_LOG" | head -1
echo ""

echo "Systemd PID 1:"
sudo grep "systemd\[1\]:.*Detected" "$CONSOLE_LOG" | head -1
echo ""

echo "Boot Complete:"
sudo grep "Reached target multi-user" "$CONSOLE_LOG"
echo ""

echo "Kernel Panics/Errors:"
sudo grep -i "panic\|oops\|bug:" "$CONSOLE_LOG" | head -5
echo ""

echo "=== Layer 1: Service Status ==="
echo ""
echo "Nginx Service:"
sudo grep "nginx.service" "$CONSOLE_LOG" | tail -5
echo ""

echo "Todo-Backend Service:"
sudo grep "todo-backend.service" "$CONSOLE_LOG" | tail -5
echo ""

echo "Failed Services:"
sudo grep "\[FAILED\]" "$CONSOLE_LOG"
echo ""

echo "=== Layer 3: Network Configuration ==="
echo ""
echo "IP Assignment:"
sudo grep "ip=" "$CONSOLE_LOG" | head -1
echo ""

echo "Network Interface:"
sudo grep -E "eth0|ens|enp" "$CONSOLE_LOG" | tail -5
echo ""

echo "Network Target:"
sudo grep "network.*target" "$CONSOLE_LOG" | tail -3
echo ""

echo "=== Layer 4: Service-Specific Errors ==="
echo ""
echo "Permission Denied:"
sudo grep -i "permission denied" "$CONSOLE_LOG" | head -5
echo ""

echo "Missing Files:"
sudo grep "No such file" "$CONSOLE_LOG" | head -5
echo ""

echo "Port Conflicts:"
sudo grep "Address already in use" "$CONSOLE_LOG" | head -5
echo ""

echo "Database Issues:"
sudo grep -i "duckdb\|database\|/data" "$CONSOLE_LOG" | head -5
echo ""

echo "=== Layer 6: Kernel/Cgroup Support ==="
echo ""
echo "Cgroup Messages:"
sudo grep -i "cgroup" "$CONSOLE_LOG" | head -5
echo ""

echo "Module Loading:"
sudo grep "FATAL\|modprobe\|module" "$CONSOLE_LOG" | head -5
echo ""

echo "========================================="
echo "Full console log: $CONSOLE_LOG"
echo "========================================="
