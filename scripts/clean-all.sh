#!/bin/bash
set -euo pipefail

# Complete NanoFuse cleanup script
# Run from any directory - cleans all nanofuse artifacts

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGES_BASE="$REPO_DIR/images/base"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}✓${NC} $1"; }
log_warn() { echo -e "${YELLOW}⚠${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }
log_step() { echo ""; echo "========================================="; echo "$1"; echo "========================================="; }

# Check for sudo
NEEDS_SUDO=0
if [ -d "$IMAGES_BASE/build" ] && [ ! -w "$IMAGES_BASE/build" ]; then
    NEEDS_SUDO=1
fi
if [ -d "/tmp/nanofuse" ] && [ ! -w "/tmp/nanofuse" ]; then
    NEEDS_SUDO=1
fi

if [ $NEEDS_SUDO -eq 1 ] && [[ $EUID -ne 0 ]]; then
    log_error "This script needs sudo to clean root-owned files"
    echo "Run with: sudo $0"
    exit 1
fi

echo "========================================"
echo "NanoFuse Complete Cleanup"
echo "========================================"
echo ""

# =============================================================================
# STEP 1: Kill processes
# =============================================================================
log_step "STEP 1: Stop all processes"

if pgrep -f "nanofused" > /dev/null 2>&1; then
    log_warn "Killing nanofused processes..."
    pkill -9 -f "nanofused" || true
    sleep 1
fi
log_info "nanofused stopped"

if pgrep -f "firecracker" > /dev/null 2>&1; then
    log_warn "Killing firecracker processes..."
    pkill -9 -f "firecracker" || true
    sleep 1
fi
log_info "firecracker stopped"

# =============================================================================
# STEP 2: Stop systemd daemon
# =============================================================================
log_step "STEP 2: Stop systemd service"

if command -v systemctl &> /dev/null; then
    if systemctl is-active --quiet nanofused 2>/dev/null; then
        log_warn "Stopping systemd nanofused service..."
        systemctl stop nanofused 2>/dev/null || true
        sleep 1
    fi
    log_info "systemd service stopped"
fi

# =============================================================================
# STEP 3: Clean Docker
# =============================================================================
log_step "STEP 3: Clean Docker"

if command -v docker &> /dev/null; then
    # Remove nanofuse images
    NANOFUSE_IMAGES=$(docker images | grep -E "nanofuse|kernel-builder" | awk '{print $3}' | tr '\n' ' ')
    if [ -n "$NANOFUSE_IMAGES" ]; then
        log_warn "Removing Docker images..."
        docker rmi -f $NANOFUSE_IMAGES 2>/dev/null || true
    fi
    log_info "Docker images cleaned"

    # Remove nanofuse containers
    NANOFUSE_CONTAINERS=$(docker ps -a | grep -E "nanofuse|kernel-builder" | awk '{print $1}' | tr '\n' ' ')
    if [ -n "$NANOFUSE_CONTAINERS" ]; then
        log_warn "Removing Docker containers..."
        docker rm -f $NANOFUSE_CONTAINERS 2>/dev/null || true
    fi
    log_info "Docker containers cleaned"

    # Prune system
    log_info "Pruning Docker system..."
    docker system prune -f --volumes > /dev/null 2>&1 || true
    docker buildx prune -f -a 2>/dev/null || true
    log_info "Docker system pruned"
fi

# =============================================================================
# STEP 4: Clean build artifacts
# =============================================================================
log_step "STEP 4: Clean build artifacts"

if [ -d "$IMAGES_BASE/build" ]; then
    log_info "Removing $IMAGES_BASE/build..."
    rm -rf "$IMAGES_BASE/build" 2>/dev/null || true
fi
log_info "Build artifacts cleaned"

# =============================================================================
# STEP 5: Clean temporary files
# =============================================================================
log_step "STEP 5: Clean temporary files"

log_info "Removing kernel binaries from /tmp..."
rm -f /tmp/vmlinux-* 2>/dev/null || true

log_info "Removing nanofuse temp data..."
rm -rf /tmp/nanofuse 2>/dev/null || true

log_info "Removing firecracker test directories..."
rm -rf /tmp/fc-test-* 2>/dev/null || true

log_info "Removing build logs..."
rm -f /tmp/docker-build.log 2>/dev/null || true
rm -f /tmp/kernel-build.log 2>/dev/null || true
rm -f /tmp/nanofused*.log 2>/dev/null || true
rm -f /tmp/vm-*.log 2>/dev/null || true
rm -f /tmp/e2e-test.log 2>/dev/null || true
rm -f /tmp/test-*.sh 2>/dev/null || true
rm -f /tmp/copy-kernel.sh 2>/dev/null || true
rm -f /tmp/fix-perms.sh 2>/dev/null || true

log_info "Temporary files cleaned"

# =============================================================================
# STEP 6: Verify clean
# =============================================================================
log_step "STEP 6: Verify clean state"

REMAINING=0

if docker images | grep -q -E "nanofuse|kernel-builder"; then
    log_warn "Docker images still present"
    REMAINING=$((REMAINING + 1))
else
    log_info "No nanofuse Docker images"
fi

if [ -d "$IMAGES_BASE/build" ]; then
    log_warn "Build directory still exists"
    REMAINING=$((REMAINING + 1))
else
    log_info "Build directory cleaned"
fi

if [ -d "/tmp/nanofuse" ]; then
    log_warn "/tmp/nanofuse still exists"
    REMAINING=$((REMAINING + 1))
else
    log_info "/tmp/nanofuse cleaned"
fi

if pgrep -f "nanofused" > /dev/null 2>&1; then
    log_warn "nanofused still running"
    REMAINING=$((REMAINING + 1))
else
    log_info "No nanofused processes"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
log_step "Cleanup Complete"

if [ $REMAINING -eq 0 ]; then
    log_info "✓ System completely cleaned"
    log_info "✓ All Docker images/containers removed"
    log_info "✓ All temporary files removed"
    log_info "✓ All build artifacts removed"
    log_info "✓ All processes stopped"
    echo ""
    log_info "Ready for fresh build: cd $IMAGES_BASE && sudo ./build.sh"
else
    log_warn "⚠ $REMAINING artifact(s) remain"
fi

echo ""
