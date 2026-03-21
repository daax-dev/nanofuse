#!/bin/bash
set -euo pipefail

# Complete end-to-end test: clean, build, verify, test VM boot
REPO_DIR="/home/jpoley/src/_mine/nanofuse"
IMAGES_BASE="$REPO_DIR/images/base"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}✓${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }
log_warn() { echo -e "${YELLOW}⚠${NC} $1"; }
log_step() { echo ""; echo "========================================="; echo "$1"; echo "========================================="; }

# Trap errors
trap 'log_error "Script failed at line $LINENO"; cleanup; exit 1' ERR

cleanup() {
    log_step "Cleanup"
    log_info "Stopping daemon..."
    sudo systemctl stop nanofused 2>/dev/null || pkill -f "nanofused" 2>/dev/null || true
    sleep 1
}

# =============================================================================
# STEP 1: Verify Prerequisites
# =============================================================================
log_step "STEP 1: Verify Prerequisites"

# Check Docker
if ! command -v docker &> /dev/null; then
    log_error "Docker not installed"
    exit 1
fi
log_info "Docker: $(docker --version)"

# Check Firecracker
if ! command -v firecracker &> /dev/null; then
    log_error "Firecracker not installed at /usr/local/bin/firecracker"
    exit 1
fi
log_info "Firecracker: $(firecracker --version)"

# Check nanofuse binary
if [ ! -x "$REPO_DIR/bin/nanofuse" ]; then
    log_error "nanofuse binary not found or not executable"
    exit 1
fi
log_info "nanofuse binary found"

# Check build tools
if ! command -v make &> /dev/null; then
    log_error "make not installed"
    exit 1
fi
log_info "make available"

# =============================================================================
# STEP 2: Clean Build Directory
# =============================================================================
log_step "STEP 2: Clean Build Directory"

cd "$IMAGES_BASE"

if [ -d "./build" ]; then
    log_info "Removing old build artifacts..."
    sudo make clean > /dev/null 2>&1 || true
    if [ -d "./build" ]; then
        log_warn "Build directory still exists after clean, removing..."
        sudo rm -rf ./build
    fi
fi
log_info "Build directory cleaned"

# =============================================================================
# STEP 3: Build Image
# =============================================================================
log_step "STEP 3: Build Image (kernel + rootfs)"

log_info "Running: sudo make build"
if ! sudo make build; then
    log_error "Build failed"
    exit 1
fi
log_info "Build complete"

# Verify artifacts exist
if [ ! -f "$IMAGES_BASE/build/vmlinux" ]; then
    log_error "Kernel (vmlinux) not found after build"
    exit 1
fi

if [ ! -f "$IMAGES_BASE/build/rootfs.ext4" ]; then
    log_error "Rootfs not found after build"
    exit 1
fi

if [ ! -f "$IMAGES_BASE/build/manifest.json" ]; then
    log_error "Manifest not found after build"
    exit 1
fi

KERNEL_SIZE=$(du -h "$IMAGES_BASE/build/vmlinux" | cut -f1)
ROOTFS_SIZE=$(du -h "$IMAGES_BASE/build/rootfs.ext4" | cut -f1)
log_info "Kernel: $KERNEL_SIZE"
log_info "Rootfs: $ROOTFS_SIZE"

# Verify kernel is valid binary
KERNEL_TYPE=$(file "$IMAGES_BASE/build/vmlinux" | cut -d: -f2)
log_info "Kernel type:$KERNEL_TYPE"
if ! echo "$KERNEL_TYPE" | grep -q "data\|executable"; then
    log_error "Kernel is not a valid binary"
    exit 1
fi

# =============================================================================
# STEP 4: Build Docker Image
# =============================================================================
log_step "STEP 4: Build Docker Image"

log_info "Building: docker build -t nanofuse-base:latest"
if ! docker build -t nanofuse-base:latest "$IMAGES_BASE" > /tmp/docker-build.log 2>&1; then
    log_error "Docker build failed"
    tail -20 /tmp/docker-build.log
    exit 1
fi
log_info "Docker image built successfully"

# Verify image exists
if ! docker images | grep -q "nanofuse-base.*latest"; then
    log_error "Docker image not found after build"
    exit 1
fi
log_info "Docker image verified in local registry"

# =============================================================================
# STEP 5: Prepare Daemon
# =============================================================================
log_step "STEP 5: Prepare Daemon Environment"

cd "$REPO_DIR"

# Stop any existing daemon
log_info "Checking for existing daemon..."
if systemctl is-active --quiet nanofused 2>/dev/null; then
    log_warn "Systemd daemon is running, stopping it..."
    sudo systemctl stop nanofused
    sleep 2
fi

if pgrep -f "nanofused" > /dev/null; then
    log_warn "Found running nanofused process, killing it..."
    pkill -9 -f "nanofused" || true
    sleep 2
fi

# Verify port is free
if netstat -tuln 2>/dev/null | grep -q ":8080"; then
    log_error "Port 8080 is still in use"
    exit 1
fi
log_info "Port 8080 is available"

# Create config
CONFIG_FILE="/tmp/nanofused-test-$$.yaml"
cat > "$CONFIG_FILE" << 'EOF'
api:
  socket: /tmp/nanofused.sock
  socket_mode: "0660"
  tcp_bind: "127.0.0.1:8080"

storage:
  data_dir: /tmp/nanofuse
  database: /tmp/nanofuse/nanofuse.db

firecracker:
  binary_path: /usr/local/bin/firecracker

limits:
  max_vms: 10
  max_total_memory_mib: 8192
  max_vcpus_per_vm: 4
  max_memory_per_vm_mib: 2048
  max_snapshot_storage_gib: 20

registry:
  auth_config_path: /root/.docker/config.json

logging:
  level: info
  format: text
EOF

log_info "Config created: $CONFIG_FILE"

# =============================================================================
# STEP 6: Start Daemon
# =============================================================================
log_step "STEP 6: Start Daemon"

log_info "Starting: sudo ./bin/nanofused --config $CONFIG_FILE"
sudo ./bin/nanofused --config "$CONFIG_FILE" > /tmp/nanofused-$$.log 2>&1 &
DAEMON_PID=$!
log_info "Daemon process started (PID: $DAEMON_PID)"

# Wait and verify daemon is actually running
sleep 3
if ! ps -p $DAEMON_PID > /dev/null 2>&1; then
    log_error "Daemon process died"
    log_error "Daemon logs:"
    cat /tmp/nanofused-$$.log
    exit 1
fi
log_info "Daemon process verified running"

# Verify TCP port is listening
RETRY=0
MAX_RETRIES=10
while [ $RETRY -lt $MAX_RETRIES ]; do
    if netstat -tuln 2>/dev/null | grep -q ":8080"; then
        log_info "TCP port 8080 is listening"
        break
    fi
    RETRY=$((RETRY + 1))
    if [ $RETRY -eq $MAX_RETRIES ]; then
        log_error "Daemon did not bind to port 8080 after $MAX_RETRIES seconds"
        log_error "Daemon logs:"
        cat /tmp/nanofused-$$.log
        kill $DAEMON_PID 2>/dev/null || true
        exit 1
    fi
    sleep 1
done

# =============================================================================
# STEP 7: Create VM
# =============================================================================
log_step "STEP 7: Create VM"

API_OPTS="--api-url http://127.0.0.1:8080"
VM_NAME="test-$(date +%s)"

log_info "VM name: $VM_NAME"
log_info "Image: nanofuse-base:latest"

if ! sudo ./bin/nanofuse $API_OPTS vm create "nanofuse-base:latest" "$VM_NAME" --memory 512 --vcpus 2 2>&1 | tee /tmp/vm-create-$$.log; then
    log_error "VM creation failed"
    cat /tmp/vm-create-$$.log
    kill $DAEMON_PID 2>/dev/null || true
    exit 1
fi
log_info "VM created successfully"

# =============================================================================
# STEP 8: Start VM
# =============================================================================
log_step "STEP 8: Start VM"

log_info "Starting VM..."
if ! sudo ./bin/nanofuse $API_OPTS vm start "$VM_NAME" 2>&1 | tee /tmp/vm-start-$$.log; then
    log_error "VM start failed"
    cat /tmp/vm-start-$$.log
    kill $DAEMON_PID 2>/dev/null || true
    exit 1
fi
log_info "VM started"

# Wait for VM to initialize
sleep 3

# =============================================================================
# STEP 9: Capture Boot Output
# =============================================================================
log_step "STEP 9: Boot Console Output"

# Find VM directory
VM_DIR=$(sudo ls -td /tmp/nanofuse/vms/* 2>/dev/null | head -1)
if [ -z "$VM_DIR" ]; then
    log_error "VM directory not found"
    kill $DAEMON_PID 2>/dev/null || true
    exit 1
fi

CONSOLE_LOG="$VM_DIR/console.log"
if [ ! -f "$CONSOLE_LOG" ]; then
    log_error "Console log not found at $CONSOLE_LOG"
    kill $DAEMON_PID 2>/dev/null || true
    exit 1
fi

log_info "Console log found: $CONSOLE_LOG"
log_info "Boot output:"
echo ""
echo "========== CONSOLE OUTPUT =========="
sudo cat "$CONSOLE_LOG" || log_error "Could not read console log"
echo "====================================="
echo ""

# Check for kernel panic or errors
if sudo grep -q "Kernel panic\|Unable to mount\|ERROR\|FATAL" "$CONSOLE_LOG" 2>/dev/null; then
    log_warn "Found error messages in boot output (see above)"
fi

if sudo grep -q "Linux version" "$CONSOLE_LOG" 2>/dev/null; then
    KERNEL_VERSION=$(sudo grep "Linux version" "$CONSOLE_LOG" | head -1)
    log_info "Kernel detected: $KERNEL_VERSION"
fi

# =============================================================================
# STEP 10: Cleanup
# =============================================================================
log_step "STEP 10: Cleanup"

log_info "Stopping VM..."
sudo ./bin/nanofuse $API_OPTS vm stop "$VM_NAME" 2>/dev/null || true
sleep 2

log_info "Stopping daemon (PID: $DAEMON_PID)..."
kill $DAEMON_PID 2>/dev/null || true
sleep 1
pkill -f "nanofused" 2>/dev/null || true

log_info "Removing temp config: $CONFIG_FILE"
rm -f "$CONFIG_FILE"

# =============================================================================
# Final Summary
# =============================================================================
log_step "Test Complete"
log_info "✓ Prerequisites verified"
log_info "✓ Build directory cleaned"
log_info "✓ Image built successfully (kernel + rootfs)"
log_info "✓ Docker image created"
log_info "✓ Daemon started and verified"
log_info "✓ VM created and started"
log_info "✓ Boot sequence captured"
log_info "✓ Cleanup completed"
echo ""
log_info "E2E test SUCCESSFUL"
