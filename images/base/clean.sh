#!/bin/bash
set -euo pipefail

# Unified Clean Script for NanoFuse Base Image
# Intelligently cleans Docker, build artifacts, temp files, and processes
#
# Usage:
#   ./clean.sh          # Auto-detect if sudo is needed
#   ./clean.sh --check  # Dry-run: show what would be cleaned
#   ./clean.sh --sudo   # Force sudo mode (for root-owned files)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}✓${NC} $1"; }
log_warn() { echo -e "${YELLOW}⚠${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }
log_step() { echo ""; echo "========================================="; echo "$1"; echo "========================================="; }

# Parse arguments
CHECK_MODE=false
FORCE_SUDO=false

for arg in "$@"; do
    case $arg in
        --check)
            CHECK_MODE=true
            shift
            ;;
        --sudo)
            FORCE_SUDO=true
            shift
            ;;
        *)
            echo "Usage: $0 [--check] [--sudo]"
            echo "  --check : Dry-run mode, show what would be cleaned"
            echo "  --sudo  : Force sudo mode for root-owned files"
            exit 1
            ;;
    esac
done

# Check if we need sudo
NEEDS_SUDO=false
if [ -d "$SCRIPT_DIR/build" ] && [ ! -w "$SCRIPT_DIR/build" ]; then
    NEEDS_SUDO=true
fi
if [ -d "/tmp/nanofuse" ] && [ ! -w "/tmp/nanofuse" ]; then
    NEEDS_SUDO=true
fi

# Header
echo "========================================"
if $CHECK_MODE; then
    echo "NanoFuse Clean Check (Dry-Run)"
elif [[ $EUID -eq 0 ]] || $FORCE_SUDO; then
    echo "NanoFuse Complete Clean (with sudo)"
else
    echo "NanoFuse Complete Clean"
fi
echo "========================================"
echo ""

if $NEEDS_SUDO && [[ $EUID -ne 0 ]] && ! $CHECK_MODE; then
    log_warn "Some files are owned by root and require sudo"
    log_warn "Re-run with: sudo $0"
    echo ""
fi

# =============================================================================
# STEP 1: Kill processes
# =============================================================================
log_step "STEP 1: Check running processes"

if pgrep -f "nanofused" > /dev/null 2>&1; then
    if $CHECK_MODE; then
        log_warn "Would kill nanofused processes"
        ps aux | grep -v grep | grep "nanofused" | awk '{print "  " $2 " - " $11}'
    else
        log_warn "Killing nanofused processes..."
        pkill -9 -f "nanofused" || true
        sleep 1
    fi
else
    log_info "No nanofused processes running"
fi

if pgrep -f "firecracker" > /dev/null 2>&1; then
    if $CHECK_MODE; then
        log_warn "Would kill firecracker processes"
        ps aux | grep -v grep | grep "firecracker" | awk '{print "  " $2 " - " $11}'
    else
        log_warn "Killing firecracker processes..."
        pkill -9 -f "firecracker" || true
        sleep 1
    fi
else
    log_info "No firecracker processes running"
fi

# =============================================================================
# STEP 2: Clean Docker
# =============================================================================
log_step "STEP 2: Clean Docker images and containers"

if docker images | grep -q "nanofuse"; then
    if $CHECK_MODE; then
        log_warn "Would remove nanofuse Docker images:"
        docker images | grep "nanofuse"
    else
        log_warn "Removing nanofuse Docker images..."
        docker rmi -f $(docker images | grep "nanofuse" | awk '{print $3}') 2>/dev/null || true
    fi
else
    log_info "No nanofuse Docker images"
fi

if docker images | grep -q "kernel-builder"; then
    if $CHECK_MODE; then
        log_warn "Would remove kernel-builder images:"
        docker images | grep "kernel-builder"
    else
        log_warn "Removing kernel-builder images..."
        docker rmi -f $(docker images | grep "kernel-builder" | awk '{print $3}') 2>/dev/null || true
    fi
else
    log_info "No kernel-builder Docker images"
fi

if docker ps -a | grep -q "nanofuse"; then
    if $CHECK_MODE; then
        log_warn "Would remove nanofuse containers:"
        docker ps -a | grep "nanofuse"
    else
        log_warn "Removing nanofuse containers..."
        docker rm -f $(docker ps -a | grep "nanofuse" | awk '{print $1}') 2>/dev/null || true
    fi
else
    log_info "No nanofuse Docker containers"
fi

if ! $CHECK_MODE; then
    log_info "Pruning Docker system..."
    docker system prune -f --volumes > /dev/null 2>&1 || true
    log_info "Docker system pruned"
else
    log_warn "Would prune Docker system (docker system prune -f --volumes)"
fi

# =============================================================================
# STEP 3: Clean build artifacts
# =============================================================================
log_step "STEP 3: Clean build artifacts"

if [ -d "$SCRIPT_DIR/build" ]; then
    if $CHECK_MODE; then
        log_warn "Would remove build directory:"
        du -sh "$SCRIPT_DIR/build"
        ls -lh "$SCRIPT_DIR/build" | head -10
    else
        if [ -w "$SCRIPT_DIR/build" ] || [[ $EUID -eq 0 ]]; then
            log_info "Removing build directory..."
            rm -rf "$SCRIPT_DIR/build"
            log_info "Build directory removed"
        else
            log_warn "Build directory exists but not writable (owned by root)"
            log_warn "Re-run with: sudo $0"
        fi
    fi
else
    log_info "No build directory"
fi

if [ -d "$SCRIPT_DIR/.docker" ]; then
    if $CHECK_MODE; then
        log_warn "Would remove .docker directory"
    else
        log_info "Removing .docker directory..."
        rm -rf "$SCRIPT_DIR/.docker"
    fi
else
    log_info "No .docker directory"
fi

# =============================================================================
# STEP 4: Clean temporary files
# =============================================================================
log_step "STEP 4: Clean temporary files"

if ls /tmp/vmlinux-* 2>/dev/null | head -1 > /dev/null; then
    if $CHECK_MODE; then
        log_warn "Would remove kernel binaries from /tmp:"
        ls -lh /tmp/vmlinux-* | awk '{print "  " $NF " (" $5 ")"}'
    else
        log_warn "Removing old kernel binaries from /tmp..."
        rm -f /tmp/vmlinux-* 2>/dev/null || true
    fi
else
    log_info "No temporary kernel binaries"
fi

if [ -d "/tmp/nanofuse" ]; then
    if $CHECK_MODE; then
        log_warn "Would remove /tmp/nanofuse:"
        du -sh /tmp/nanofuse
    else
        if [ -w "/tmp/nanofuse" ] || [[ $EUID -eq 0 ]]; then
            log_warn "Removing /tmp/nanofuse directory..."
            rm -rf /tmp/nanofuse
            log_info "Temporary nanofuse data cleaned"
        else
            log_warn "/tmp/nanofuse directory exists but owned by root"
            log_warn "Re-run with: sudo $0"
        fi
    fi
else
    log_info "No /tmp/nanofuse directory"
fi

# Remove rootfs test files
if [ -f "/tmp/rootfs-working.ext4" ]; then
    if $CHECK_MODE; then
        log_warn "Would remove /tmp/rootfs-working.ext4"
    else
        rm -f /tmp/rootfs-working.ext4 2>/dev/null || true
    fi
fi

# Remove test configs and logs
TEMP_FILES=(
    /tmp/test_boot_config.json
    /tmp/fc_boot_output.log
    /tmp/docker-build.log
    /tmp/kernel-build.log
    /tmp/nanofused-*.log
    /tmp/vm-create-*.log
    /tmp/vm-start-*.log
    /tmp/e2e-test.log
    /tmp/nanofused-test.yaml
)

FOUND_TEMPS=false
for pattern in "${TEMP_FILES[@]}"; do
    if ls $pattern 2>/dev/null | head -1 > /dev/null; then
        FOUND_TEMPS=true
        if $CHECK_MODE; then
            log_warn "Would remove: $pattern"
        fi
    fi
done

if ! $CHECK_MODE && $FOUND_TEMPS; then
    log_info "Removing build logs and test configs..."
    for pattern in "${TEMP_FILES[@]}"; do
        rm -f $pattern 2>/dev/null || true
    done
    log_info "Temporary files cleaned"
elif ! $FOUND_TEMPS; then
    log_info "No temporary test files"
fi

# =============================================================================
# STEP 5: Clean Docker buildkit cache
# =============================================================================
log_step "STEP 5: Clean Docker buildkit cache"

if ! $CHECK_MODE; then
    log_info "Clearing Docker buildkit cache..."
    docker buildx prune -f -a 2>/dev/null || true
    log_info "Docker buildkit cache pruned"
else
    log_warn "Would prune Docker buildkit cache (docker buildx prune -f -a)"
fi

# =============================================================================
# STEP 6: Verification
# =============================================================================
log_step "STEP 6: Verification"

REMAINING=0

if docker images | grep -q "nanofuse\|kernel-builder"; then
    log_warn "Docker images still present"
    REMAINING=$((REMAINING + 1))
else
    log_info "No Docker images remaining"
fi

if [ -d "$SCRIPT_DIR/build" ]; then
    log_warn "Build directory still exists"
    REMAINING=$((REMAINING + 1))
else
    log_info "No build directory"
fi

if [ -d "/tmp/nanofuse" ]; then
    log_warn "/tmp/nanofuse directory still exists"
    REMAINING=$((REMAINING + 1))
else
    log_info "No /tmp/nanofuse directory"
fi

if pgrep -f "nanofused|firecracker" > /dev/null 2>&1; then
    log_warn "Processes still running"
    REMAINING=$((REMAINING + 1))
else
    log_info "No processes running"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
log_step "Summary"

if $CHECK_MODE; then
    echo -e "${BLUE}ℹ${NC}  This was a dry-run. Run without --check to actually clean."
    echo ""
    echo "To clean:"
    if $NEEDS_SUDO; then
        echo "  sudo $0"
    else
        echo "  $0"
    fi
elif [ $REMAINING -eq 0 ]; then
    log_info "✓ System is clean!"
    log_info "✓ All Docker images removed"
    log_info "✓ All build artifacts removed"
    log_info "✓ All temporary files cleared"
    log_info "✓ All processes stopped"
    log_info "✓ Build cache cleared"
    echo ""
    log_info "Ready for fresh build!"
else
    log_warn "⚠ $REMAINING artifact(s) remain"
    if $NEEDS_SUDO && [[ $EUID -ne 0 ]]; then
        echo ""
        log_warn "Some files require sudo to remove"
        log_warn "Run: sudo $0"
    fi
fi

echo ""
