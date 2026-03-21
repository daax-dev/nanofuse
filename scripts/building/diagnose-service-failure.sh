#!/bin/bash
# Diagnose why nginx and services are failing in the VM
# This helps us understand if kernel version is actually the problem

set -e

VM_ID="${1:-$(nanofuse vm list --json | jq -r '.vms[0].id')}"

if [ -z "$VM_ID" ] || [ "$VM_ID" = "null" ]; then
    echo "Error: No VM found. Create a VM first."
    exit 1
fi

CONSOLE_LOG="/var/lib/nanofuse/vms/$VM_ID/console.log"

if [ ! -r "$CONSOLE_LOG" ]; then
    echo "Error: Cannot read console log: $CONSOLE_LOG"
    echo "Run with sudo: sudo $0 $VM_ID"
    exit 1
fi

echo "========================================="
echo "Service Failure Diagnostic"
echo "========================================="
echo ""
echo "VM ID: $VM_ID"
echo "Console log: $CONSOLE_LOG"
echo ""

# Check kernel version
echo "=== Kernel Version ==="
KERNEL_VERSION=$(grep "Linux version" "$CONSOLE_LOG" | head -1)
echo "$KERNEL_VERSION"

if echo "$KERNEL_VERSION" | grep -q "5\.10\."; then
    echo "⚠️  WARNING: Old kernel 5.10.x detected"
    echo "   Ubuntu 24.04 expects kernel 6.x+"
elif echo "$KERNEL_VERSION" | grep -q "6\.[0-9]\."; then
    echo "✓ Kernel 6.x detected (good)"
else
    echo "? Unknown kernel version"
fi
echo ""

# Check if systemd started
echo "=== Systemd Startup ==="
if grep -q "systemd\[1\]:" "$CONSOLE_LOG"; then
    echo "✓ Systemd started as PID 1"
    grep "systemd\[1\]:.*Detected" "$CONSOLE_LOG" | head -3
else
    echo "✗ Systemd did not start"
fi
echo ""

# Check which services were attempted
echo "=== Service Start Attempts ==="
grep -E "Starting|Started" "$CONSOLE_LOG" | grep -E "(nginx|todo-backend)" | tail -10
echo ""

# Check for service failures
echo "=== Service Failures ==="
grep -E "\[FAILED\]|Failed to start" "$CONSOLE_LOG" | tail -10
echo ""

# Look for actual error messages
echo "=== Nginx Error Messages ==="
grep -i "nginx" "$CONSOLE_LOG" | grep -iE "error|fail|cannot|denied|missing" | tail -15
echo ""

# Check for systemd journal errors
echo "=== Systemd Journal Errors ==="
grep -E "systemd\[1\]:" "$CONSOLE_LOG" | grep -iE "error|fail|cannot" | tail -15
echo ""

# Check for cgroup issues
echo "=== Cgroup Issues ==="
if grep -q "cgroup" "$CONSOLE_LOG"; then
    grep -i "cgroup" "$CONSOLE_LOG" | grep -iE "error|fail|warn" | tail -10
else
    echo "(no cgroup errors found)"
fi
echo ""

# Check todo-backend status
echo "=== Todo-Backend Service ==="
grep "todo-backend" "$CONSOLE_LOG" | tail -10
echo ""

# Check for library/binary issues
echo "=== Binary/Library Issues ==="
grep -iE "not found|no such file|cannot execute" "$CONSOLE_LOG" | tail -10
echo ""

# Final system state
echo "=== Final System State ==="
grep "Reached target" "$CONSOLE_LOG" | tail -5
echo ""

echo "========================================="
echo "Diagnostic Summary"
echo "========================================="
echo ""

# Determine likely cause
HAS_SYSTEMD=$(grep -c "systemd\[1\]:" "$CONSOLE_LOG" || echo "0")
HAS_NGINX_FAIL=$(grep -c "Failed to start nginx" "$CONSOLE_LOG" || echo "0")
HAS_BACKEND_OK=$(grep -c "Started todo-backend" "$CONSOLE_LOG" || echo "0")

echo "Systemd messages: $HAS_SYSTEMD"
echo "Nginx failures: $HAS_NGINX_FAIL"
echo "Backend started: $HAS_BACKEND_OK"
echo ""

if [ "$HAS_SYSTEMD" -gt 0 ]; then
    echo "✓ Systemd IS running (kernel compatible enough to boot Ubuntu 24.04)"
    echo ""

    if [ "$HAS_NGINX_FAIL" -gt 0 ]; then
        echo "Issue: Nginx specifically is failing"
        echo ""
        echo "Possible causes:"
        echo "1. Missing nginx binary or libraries"
        echo "2. Port 80 bind failure"
        echo "3. Configuration file syntax error"
        echo "4. Missing /var/www/html directory"
        echo "5. Permission issues"
        echo ""
        echo "Check the 'Nginx Error Messages' section above for specifics."
    fi

    if [ "$HAS_BACKEND_OK" -gt 0 ]; then
        echo "Note: todo-backend service started successfully"
        echo "This suggests the issue is specific to nginx, not a general systemd/kernel problem"
    fi
else
    echo "✗ Systemd NOT running"
    echo "This is a fundamental boot failure (kernel incompatibility or init issue)"
fi

echo ""
echo "Full console log: $CONSOLE_LOG"
echo "View with: sudo cat $CONSOLE_LOG"
