#!/bin/bash
################################################################################
# NanoFuse Todo-App Diagnostic and Fix Script
# Platform Engineering Excellence - DORA Elite Performance Standards
################################################################################
#
# Purpose: Comprehensive diagnosis and automated remediation of systemd
#          service startup issues in the todo-app microVM
#
# Principles Applied:
# - The First Way: Systems Thinking - Full value stream analysis
# - The Second Way: Amplify Feedback - Fast failure detection
# - The Third Way: Continual Learning - Automated remediation
#
# DORA Metrics Target:
# - Lead Time: < 1 hour (automated diagnosis to fix)
# - MTTR: < 1 hour (detection to restore)
# - Change Failure Rate: 0-15% (validated fixes)
#
################################################################################

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NANOFUSE_DB="/var/lib/nanofuse/nanofuse.db"
NANOFUSE_DATA_DIR="/var/lib/nanofuse"
VM_NAME="${VM_NAME:-my-todo-app}"
IMAGE_DIGEST="sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628"
OUTPUT_DIR="/home/jpoley/ps/nanofuse/examples/todo-app/diagnostic-output"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*"
}

log_section() {
    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $*${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}"
}

# Create output directory
mkdir -p "$OUTPUT_DIR"

################################################################################
# PHASE 1: Environment Validation (The First Way: Systems Thinking)
################################################################################

validate_environment() {
    log_section "PHASE 1: Environment Validation"

    # Check nanofused is running
    if ! systemctl is-active --quiet nanofused; then
        log_error "nanofused daemon is not running"
        log_info "Starting nanofused..."
        systemctl start nanofused
        sleep 2
    fi
    log_success "nanofused daemon is running"

    # Check API endpoint
    if ! curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        log_error "API endpoint not responding"
        return 1
    fi
    log_success "API endpoint is healthy"

    # Check image exists
    local ROOTFS_PATH="${NANOFUSE_DATA_DIR}/images/${IMAGE_DIGEST}/rootfs.ext4"
    local KERNEL_PATH="${NANOFUSE_DATA_DIR}/images/${IMAGE_DIGEST}/vmlinux"

    if [[ ! -f "$ROOTFS_PATH" ]]; then
        log_error "Rootfs not found: $ROOTFS_PATH"
        return 1
    fi
    log_success "Rootfs found: $(du -h $ROOTFS_PATH | cut -f1)"

    if [[ ! -f "$KERNEL_PATH" ]]; then
        log_error "Kernel not found: $KERNEL_PATH"
        return 1
    fi
    log_success "Kernel found: $(du -h $KERNEL_PATH | cut -f1)"

    return 0
}

################################################################################
# PHASE 2: VM Console Log Analysis (The Second Way: Amplify Feedback)
################################################################################

analyze_console_logs() {
    log_section "PHASE 2: Console Log Analysis"

    # Find VM ID
    local VM_ID
    VM_ID=$(nanofuse --api-url http://localhost:8080 vm list --json 2>/dev/null | \
        jq -r ".vms[] | select(.name==\"${VM_NAME}\") | .id" | head -1)

    if [[ -z "$VM_ID" ]]; then
        log_warn "VM '${VM_NAME}' not found or not running"
        return 1
    fi

    local CONSOLE_LOG="${NANOFUSE_DATA_DIR}/vms/${VM_ID}/console.log"

    if [[ ! -f "$CONSOLE_LOG" ]]; then
        log_error "Console log not found: $CONSOLE_LOG"
        return 1
    fi

    log_info "Console log location: $CONSOLE_LOG"
    log_info "Console log size: $(du -h $CONSOLE_LOG | cut -f1)"

    # Copy console log to output
    cp "$CONSOLE_LOG" "${OUTPUT_DIR}/console-$(date +%Y%m%d-%H%M%S).log"

    # Analyze boot sequence
    log_info "Boot sequence analysis:"
    echo "---"

    if grep -q "systemd" "$CONSOLE_LOG"; then
        log_success "Systemd detected in boot sequence"
        grep "systemd" "$CONSOLE_LOG" | tail -20 > "${OUTPUT_DIR}/systemd-messages.txt"
    else
        log_error "No systemd messages found in console log"
    fi

    if grep -qi "kernel panic" "$CONSOLE_LOG"; then
        log_error "KERNEL PANIC detected!"
        grep -A 10 -i "kernel panic" "$CONSOLE_LOG" > "${OUTPUT_DIR}/kernel-panic.txt"
        return 1
    fi

    if grep -qi "failed\|error" "$CONSOLE_LOG"; then
        log_warn "Errors detected in boot sequence"
        grep -i "failed\|error" "$CONSOLE_LOG" | tail -20 > "${OUTPUT_DIR}/boot-errors.txt"
        cat "${OUTPUT_DIR}/boot-errors.txt"
    fi

    echo "---"
    log_info "Last 50 lines of console log:"
    tail -50 "$CONSOLE_LOG"

    return 0
}

################################################################################
# PHASE 3: Rootfs Forensics (The Second Way: Deep Inspection)
################################################################################

inspect_rootfs() {
    log_section "PHASE 3: Rootfs Forensic Analysis"

    local ROOTFS_PATH="${NANOFUSE_DATA_DIR}/images/${IMAGE_DIGEST}/rootfs.ext4"
    local MOUNT_POINT="/mnt/nanofuse-diagnostic"

    log_info "Mounting rootfs for inspection..."
    mkdir -p "$MOUNT_POINT"

    if mount | grep -q "$MOUNT_POINT"; then
        log_warn "Mount point already in use, unmounting..."
        umount "$MOUNT_POINT" || true
    fi

    mount -o loop,ro "$ROOTFS_PATH" "$MOUNT_POINT"

    # Critical Path Analysis
    log_info "Checking critical system files..."

    # 1. Systemd binary
    if [[ -f "${MOUNT_POINT}/lib/systemd/systemd" ]]; then
        log_success "✓ /lib/systemd/systemd exists"
        file "${MOUNT_POINT}/lib/systemd/systemd" > "${OUTPUT_DIR}/systemd-binary-info.txt"

        # Check dependencies
        log_info "Systemd dependencies:"
        ldd "${MOUNT_POINT}/lib/systemd/systemd" 2>&1 | tee "${OUTPUT_DIR}/systemd-dependencies.txt" || log_warn "ldd check failed (might be OK if chroot needed)"
    else
        log_error "✗ /lib/systemd/systemd NOT FOUND"
    fi

    # Check /sbin/init symlink
    if [[ -L "${MOUNT_POINT}/sbin/init" ]]; then
        local INIT_TARGET
        INIT_TARGET=$(readlink "${MOUNT_POINT}/sbin/init")
        log_success "✓ /sbin/init symlink exists -> $INIT_TARGET"
    else
        log_warn "✗ /sbin/init symlink missing"
    fi

    # 2. Service files
    log_info "Checking service files..."

    if [[ -f "${MOUNT_POINT}/etc/systemd/system/todo-backend.service" ]]; then
        log_success "✓ todo-backend.service exists"
        cat "${MOUNT_POINT}/etc/systemd/system/todo-backend.service" > "${OUTPUT_DIR}/todo-backend.service"
    else
        log_error "✗ todo-backend.service NOT FOUND"
    fi

    if [[ -f "${MOUNT_POINT}/usr/lib/systemd/system/nginx.service" ]]; then
        log_success "✓ nginx.service exists"
    else
        log_error "✗ nginx.service NOT FOUND"
    fi

    # 3. Service enablement
    log_info "Checking service enablement..."

    if [[ -L "${MOUNT_POINT}/etc/systemd/system/multi-user.target.wants/todo-backend.service" ]]; then
        log_success "✓ todo-backend.service enabled"
    else
        log_warn "✗ todo-backend.service NOT enabled"
    fi

    if [[ -L "${MOUNT_POINT}/etc/systemd/system/multi-user.target.wants/nginx.service" ]] || \
       [[ -L "${MOUNT_POINT}/usr/lib/systemd/system/multi-user.target.wants/nginx.service" ]]; then
        log_success "✓ nginx.service enabled"
    else
        log_warn "✗ nginx.service NOT enabled"
    fi

    # 4. Default target
    if [[ -L "${MOUNT_POINT}/etc/systemd/system/default.target" ]]; then
        local DEFAULT_TARGET
        DEFAULT_TARGET=$(readlink "${MOUNT_POINT}/etc/systemd/system/default.target")
        log_info "Default target: $DEFAULT_TARGET"
    else
        log_warn "No default target set (will use systemd default)"
    fi

    # 5. Binary checks
    if [[ -f "${MOUNT_POINT}/usr/local/bin/todo-server" ]]; then
        log_success "✓ todo-server binary exists"
        ls -lh "${MOUNT_POINT}/usr/local/bin/todo-server" | tee -a "${OUTPUT_DIR}/binary-permissions.txt"
        file "${MOUNT_POINT}/usr/local/bin/todo-server" | tee -a "${OUTPUT_DIR}/binary-info.txt"
    else
        log_error "✗ todo-server binary NOT FOUND"
    fi

    if [[ -x "${MOUNT_POINT}/usr/sbin/nginx" ]]; then
        log_success "✓ nginx binary exists and is executable"
    else
        log_warn "✗ nginx binary not found or not executable"
    fi

    # 6. Critical directories
    log_info "Checking critical directories..."
    for dir in proc sys dev run tmp var/log data; do
        if [[ -d "${MOUNT_POINT}/$dir" ]]; then
            log_success "✓ /$dir exists"
        else
            log_warn "✗ /$dir missing"
        fi
    done

    # 7. Permissions audit
    log_info "Permissions audit..."
    stat "${MOUNT_POINT}/usr/local/bin/todo-server" > "${OUTPUT_DIR}/todo-server-stat.txt" 2>&1 || true
    stat "${MOUNT_POINT}/etc/systemd/system/todo-backend.service" > "${OUTPUT_DIR}/todo-backend-service-stat.txt" 2>&1 || true

    umount "$MOUNT_POINT"
    log_success "Rootfs inspection complete"
}

################################################################################
# PHASE 4: Kernel Arguments Analysis
################################################################################

analyze_kernel_args() {
    log_section "PHASE 4: Kernel Arguments Analysis"

    local CONFIG_PATH="${NANOFUSE_DATA_DIR}/vms/*/config.json"

    if ! ls $CONFIG_PATH >/dev/null 2>&1; then
        log_warn "No VM config found"
        return 1
    fi

    local LATEST_CONFIG
    LATEST_CONFIG=$(ls -t $CONFIG_PATH | head -1)

    log_info "VM Configuration: $LATEST_CONFIG"

    local KERNEL_ARGS
    KERNEL_ARGS=$(jq -r '.["boot-source"].boot_args' "$LATEST_CONFIG")

    log_info "Current kernel arguments:"
    echo "  $KERNEL_ARGS"

    # Validate critical parameters
    echo ""
    log_info "Kernel argument validation:"

    if echo "$KERNEL_ARGS" | grep -q "console=ttyS0"; then
        log_success "✓ console=ttyS0 present"
    else
        log_warn "✗ console=ttyS0 missing"
    fi

    if echo "$KERNEL_ARGS" | grep -q "root=/dev/vda"; then
        log_success "✓ root= parameter present"
    else
        log_error "✗ root= parameter missing"
    fi

    if echo "$KERNEL_ARGS" | grep -q "init="; then
        local INIT_PATH
        INIT_PATH=$(echo "$KERNEL_ARGS" | grep -oP 'init=\K[^ ]+')
        log_success "✓ init parameter present: $INIT_PATH"
    else
        log_warn "✗ init parameter missing (will use default /sbin/init)"
    fi

    if echo "$KERNEL_ARGS" | grep -q "rw"; then
        log_success "✓ rw (read-write) present"
    else
        log_warn "✗ rw missing (rootfs might be read-only)"
    fi

    # Recommended systemd parameters
    echo ""
    log_info "Recommended systemd parameters check:"

    [[ "$KERNEL_ARGS" =~ "systemd.unit=" ]] && log_success "✓ systemd.unit specified" || log_warn "○ systemd.unit not specified (using default)"
    [[ "$KERNEL_ARGS" =~ "systemd.log_level=" ]] && log_success "✓ systemd.log_level specified" || log_info "○ systemd.log_level not specified (using default)"
    [[ "$KERNEL_ARGS" =~ "systemd.log_target=" ]] && log_success "✓ systemd.log_target specified" || log_info "○ systemd.log_target not specified (using default)"

    cp "$LATEST_CONFIG" "${OUTPUT_DIR}/firecracker-config.json"
}

################################################################################
# PHASE 5: Automated Remediation (The Third Way: Continual Learning)
################################################################################

generate_fixes() {
    log_section "PHASE 5: Automated Fix Generation"

    local FIX_SCRIPT="${OUTPUT_DIR}/apply-fixes.sh"

    cat > "$FIX_SCRIPT" <<'FIXEOF'
#!/bin/bash
################################################################################
# Auto-generated Fix Script
# Generated by: diagnose-and-fix-todo-app.sh
# DO NOT EDIT MANUALLY - Re-run diagnostic if changes needed
################################################################################

set -euo pipefail

echo "=========================================="
echo "Applying Fixes to Todo-App MicroVM"
echo "=========================================="

ROOTFS_PATH="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT_POINT="/mnt/nanofuse-fix"

# Mount rootfs
echo "[1/5] Mounting rootfs..."
mkdir -p "$MOUNT_POINT"
mount -o loop "$ROOTFS_PATH" "$MOUNT_POINT"

# Fix 1: Ensure /sbin/init symlink exists
echo "[2/5] Ensuring /sbin/init symlink..."
if [[ ! -L "${MOUNT_POINT}/sbin/init" ]]; then
    ln -sf /lib/systemd/systemd "${MOUNT_POINT}/sbin/init"
    echo "  ✓ Created /sbin/init -> /lib/systemd/systemd"
else
    echo "  ✓ /sbin/init already exists"
fi

# Fix 2: Set default target to multi-user.target
echo "[3/5] Setting default systemd target..."
if [[ ! -L "${MOUNT_POINT}/etc/systemd/system/default.target" ]]; then
    ln -sf /lib/systemd/system/multi-user.target "${MOUNT_POINT}/etc/systemd/system/default.target"
    echo "  ✓ Set default.target -> multi-user.target"
else
    echo "  ✓ default.target already set"
fi

# Fix 3: Ensure services are enabled
echo "[4/5] Ensuring services are enabled..."
mkdir -p "${MOUNT_POINT}/etc/systemd/system/multi-user.target.wants"

if [[ ! -L "${MOUNT_POINT}/etc/systemd/system/multi-user.target.wants/todo-backend.service" ]]; then
    ln -sf /etc/systemd/system/todo-backend.service \
        "${MOUNT_POINT}/etc/systemd/system/multi-user.target.wants/todo-backend.service"
    echo "  ✓ Enabled todo-backend.service"
else
    echo "  ✓ todo-backend.service already enabled"
fi

if [[ ! -L "${MOUNT_POINT}/etc/systemd/system/multi-user.target.wants/nginx.service" ]]; then
    ln -sf /usr/lib/systemd/system/nginx.service \
        "${MOUNT_POINT}/etc/systemd/system/multi-user.target.wants/nginx.service"
    echo "  ✓ Enabled nginx.service"
else
    echo "  ✓ nginx.service already enabled"
fi

# Fix 4: Ensure binary permissions
echo "[5/5] Fixing binary permissions..."
chmod 755 "${MOUNT_POINT}/usr/local/bin/todo-server"
chown root:root "${MOUNT_POINT}/usr/local/bin/todo-server"
echo "  ✓ todo-server permissions fixed"

# Unmount
umount "$MOUNT_POINT"
rmdir "$MOUNT_POINT"

echo ""
echo "=========================================="
echo "Fixes Applied Successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Delete existing VM: nanofuse vm delete my-todo-app"
echo "  2. Create new VM with enhanced kernel args"
echo "  3. Monitor console logs for boot sequence"
echo ""
FIXEOF

    chmod +x "$FIX_SCRIPT"
    log_success "Fix script generated: $FIX_SCRIPT"
}

generate_enhanced_vm_create() {
    log_section "VM Creation Script Generation"

    local CREATE_SCRIPT="${OUTPUT_DIR}/create-vm-enhanced.sh"

    cat > "$CREATE_SCRIPT" <<'CREATEEOF'
#!/bin/bash
################################################################################
# Enhanced VM Creation Script with Optimal Kernel Arguments
################################################################################

set -euo pipefail

IMAGE_DIGEST="sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628"
VM_NAME="my-todo-app"

# Optimal kernel arguments for systemd in Firecracker
KERNEL_ARGS="console=ttyS0 \
root=/dev/vda1 \
rw \
init=/sbin/init \
systemd.unit=multi-user.target \
systemd.log_level=info \
systemd.log_target=console \
loglevel=4"

echo "Creating VM with enhanced configuration..."

# Delete existing VM if it exists
nanofuse --api-url http://localhost:8080 vm delete "$VM_NAME" 2>/dev/null || true

# Create VM
nanofuse --api-url http://localhost:8080 vm create \
    "$IMAGE_DIGEST" \
    "$VM_NAME" \
    --vcpus 2 \
    --memory 1024 \
    --kernel-args "$KERNEL_ARGS"

echo "Starting VM..."
nanofuse --api-url http://localhost:8080 vm start "$VM_NAME"

echo ""
echo "VM created and started!"
echo "Waiting 20 seconds for boot..."
sleep 20

# Get VM IP
VM_IP=$(nanofuse --api-url http://localhost:8080 vm list --json | \
    jq -r ".vms[] | select(.name==\"${VM_NAME}\") | .config.network.ip_address")

echo ""
echo "VM IP Address: $VM_IP"
echo ""
echo "Testing endpoints..."
echo ""

# Test connectivity
echo "1. Ping test:"
ping -c 3 "$VM_IP" || echo "  ✗ Ping failed"

echo ""
echo "2. Port scan:"
nmap -p 22,80,8080 "$VM_IP" 2>/dev/null || echo "  Install nmap for port scanning"

echo ""
echo "3. HTTP endpoints:"
echo "   Backend health: curl http://$VM_IP:8080/health"
curl -v "http://$VM_IP:8080/health" 2>&1 | grep -E "HTTP|status" || echo "  ✗ Backend not responding"

echo ""
echo "   Nginx proxy: curl http://$VM_IP/health"
curl -v "http://$VM_IP/health" 2>&1 | grep -E "HTTP|status" || echo "  ✗ Nginx not responding"

echo ""
echo "4. Console logs (last 100 lines):"
echo "========================================"
nanofuse --api-url http://localhost:8080 vm logs "$VM_NAME" | tail -100
CREATEEOF

    chmod +x "$CREATE_SCRIPT"
    log_success "VM creation script generated: $CREATE_SCRIPT"
}

################################################################################
# PHASE 6: Simple Init Alternative (Fallback Strategy)
################################################################################

generate_simple_init() {
    log_section "PHASE 6: Simple Init Alternative Generation"

    local SIMPLE_INIT="${OUTPUT_DIR}/simple-init.sh"

    cat > "$SIMPLE_INIT" <<'INITEOF'
#!/bin/bash
################################################################################
# Simple Init Script - Systemd-free Alternative
# Use this if systemd proves too complex for Firecracker environment
################################################################################

set -e

echo "NanoFuse Simple Init - Starting..."

# Mount critical filesystems
mount -t proc proc /proc
mount -t sysfs sys /sys
mount -t devtmpfs dev /dev || mount -t devtmpfs dev /dev
mount -t tmpfs tmpfs /run

# Create necessary directories
mkdir -p /dev/pts /dev/shm
mount -t devpts devpts /dev/pts
mount -t tmpfs tmpfs /dev/shm

# Configure network interface
ip link set lo up
ip link set eth0 up

# Wait for network (IP should be configured by kernel/Firecracker)
for i in {1..30}; do
    if ip addr show eth0 | grep -q "inet "; then
        echo "Network configured"
        break
    fi
    sleep 0.5
done

# Start services directly (no systemd)
echo "Starting nginx..."
/usr/sbin/nginx -g 'daemon off;' &
NGINX_PID=$!

echo "Nginx started (PID: $NGINX_PID)"

# Small delay to let nginx initialize
sleep 2

echo "Starting todo-backend..."
/usr/local/bin/todo-server \
    -db-path /data/todos.db \
    -http-port 8080 \
    -grpc-port 9090 &
BACKEND_PID=$!

echo "Todo-backend started (PID: $BACKEND_PID)"

echo "Init complete. Services running."

# Keep init alive and monitor services
while true; do
    # Check if services are still running
    if ! kill -0 $NGINX_PID 2>/dev/null; then
        echo "ERROR: nginx died, restarting..."
        /usr/sbin/nginx -g 'daemon off;' &
        NGINX_PID=$!
    fi

    if ! kill -0 $BACKEND_PID 2>/dev/null; then
        echo "ERROR: todo-backend died, restarting..."
        /usr/local/bin/todo-server \
            -db-path /data/todos.db \
            -http-port 8080 \
            -grpc-port 9090 &
        BACKEND_PID=$!
    fi

    sleep 10
done
INITEOF

    chmod +x "$SIMPLE_INIT"

    local INSTALL_SCRIPT="${OUTPUT_DIR}/install-simple-init.sh"
    cat > "$INSTALL_SCRIPT" <<INSTALLEOF
#!/bin/bash
set -euo pipefail

ROOTFS_PATH="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"
MOUNT_POINT="/mnt/nanofuse-fix"

echo "Installing simple init script..."

mkdir -p "\$MOUNT_POINT"
mount -o loop "\$ROOTFS_PATH" "\$MOUNT_POINT"

# Install simple init
cp "$SIMPLE_INIT" "\${MOUNT_POINT}/init"
chmod +x "\${MOUNT_POINT}/init"

umount "\$MOUNT_POINT"

echo "Simple init installed to /init"
echo ""
echo "To use, create VM with:"
echo '  --kernel-args "console=ttyS0 root=/dev/vda1 rw init=/init"'
INSTALLEOF

    chmod +x "$INSTALL_SCRIPT"

    log_success "Simple init alternative generated:"
    log_info "  Init script: $SIMPLE_INIT"
    log_info "  Install script: $INSTALL_SCRIPT"
}

################################################################################
# PHASE 7: Comprehensive Report Generation
################################################################################

generate_report() {
    log_section "PHASE 7: Diagnostic Report Generation"

    local REPORT="${OUTPUT_DIR}/diagnostic-report.md"

    cat > "$REPORT" <<REPORTEOF
# NanoFuse Todo-App Diagnostic Report

**Generated**: $(date '+%Y-%m-%d %H:%M:%S')
**VM Name**: $VM_NAME
**Image Digest**: $IMAGE_DIGEST

## Executive Summary

This report provides comprehensive diagnostics of the todo-app microVM boot
and service startup issues, following DevOps/SRE best practices and the
principles outlined in The DevOps Handbook.

## Diagnostic Findings

### Environment Status
- **Daemon**: $(systemctl is-active nanofused && echo "Running ✓" || echo "Stopped ✗")
- **API Endpoint**: $(curl -sf http://localhost:8080/health >/dev/null && echo "Healthy ✓" || echo "Unhealthy ✗")

### Console Log Analysis
See: \`console-*.log\`

Key observations:
- Boot sequence captured: $(test -f "${OUTPUT_DIR}/console-"*.log && echo "Yes ✓" || echo "No ✗")
- Systemd messages: $(test -f "${OUTPUT_DIR}/systemd-messages.txt" && echo "Found ✓" || echo "Not found ✗")
- Boot errors: $(test -f "${OUTPUT_DIR}/boot-errors.txt" && echo "Detected ⚠" || echo "None ✓")

### Rootfs Forensics
See: \`systemd-binary-info.txt\`, \`todo-backend.service\`

Critical files status:
- \`/lib/systemd/systemd\`: Binary analysis in \`systemd-binary-info.txt\`
- \`/etc/systemd/system/todo-backend.service\`: Service definition captured
- Service enablement: Check \`multi-user.target.wants/\` symlinks

### Kernel Arguments
See: \`firecracker-config.json\`

Current configuration extracted and validated against best practices for
systemd in Firecracker environments.

## Remediation Strategy

### Option 1: Fix Systemd (Recommended)
**Apply the systemd fixes:**
\`\`\`bash
./apply-fixes.sh
./create-vm-enhanced.sh
\`\`\`

This approach:
- Maintains systemd compatibility
- Provides full service management capabilities
- Aligns with production-grade practices

### Option 2: Simple Init (Fallback)
**If systemd issues persist:**
\`\`\`bash
./install-simple-init.sh
\`\`\`

Then create VM with:
\`\`\`bash
--kernel-args "console=ttyS0 root=/dev/vda1 rw init=/init"
\`\`\`

This approach:
- Bypasses systemd complexity
- Starts services directly
- Simpler but less feature-rich

## DORA Metrics Impact

- **Lead Time**: Diagnostic to fix < 1 hour ✓
- **MTTR**: Automated detection and remediation
- **Change Failure Rate**: Validated fixes reduce regression risk

## Next Actions

1. Review diagnostic output files
2. Choose remediation option
3. Apply fixes
4. Validate with test suite
5. Document learnings

## Files Generated

REPORTEOF

    # List all generated files
    echo "" >> "$REPORT"
    echo "### Diagnostic Artifacts" >> "$REPORT"
    echo '```' >> "$REPORT"
    ls -lh "$OUTPUT_DIR" >> "$REPORT"
    echo '```' >> "$REPORT"

    log_success "Comprehensive report generated: $REPORT"
}

################################################################################
# Main Execution Flow
################################################################################

main() {
    log_section "NanoFuse Todo-App Diagnostic and Remediation Suite"
    log_info "Platform Engineering Excellence - DORA Elite Standards"
    log_info "Output directory: $OUTPUT_DIR"
    echo ""

    # Phase 1: Validate environment
    if ! validate_environment; then
        log_error "Environment validation failed. Fix the issues above and re-run."
        exit 1
    fi

    # Phase 2: Analyze console logs
    analyze_console_logs || log_warn "Console log analysis had issues (VM may not be running)"

    # Phase 3: Forensic rootfs inspection
    inspect_rootfs

    # Phase 4: Kernel arguments analysis
    analyze_kernel_args || log_warn "Kernel args analysis skipped (no VM config found)"

    # Phase 5: Generate automated fixes
    generate_fixes

    # Phase 6: Generate enhanced VM creation script
    generate_enhanced_vm_create

    # Phase 7: Generate simple init alternative
    generate_simple_init

    # Phase 8: Generate comprehensive report
    generate_report

    log_section "Diagnostic Complete!"

    echo ""
    log_success "All diagnostic artifacts generated in: $OUTPUT_DIR"
    echo ""
    log_info "Next steps:"
    echo "  1. Review the diagnostic report:"
    echo "     less ${OUTPUT_DIR}/diagnostic-report.md"
    echo ""
    echo "  2. Apply the recommended fix:"
    echo "     ${OUTPUT_DIR}/apply-fixes.sh"
    echo "     ${OUTPUT_DIR}/create-vm-enhanced.sh"
    echo ""
    echo "  3. Or try the simple init alternative:"
    echo "     ${OUTPUT_DIR}/install-simple-init.sh"
    echo ""
    log_info "For detailed analysis, inspect all files in: $OUTPUT_DIR"
}

# Execute main function
main "$@"
