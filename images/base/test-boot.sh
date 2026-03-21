#!/bin/bash
set -euo pipefail

# Unified Boot Test Script for NanoFuse Base Image
# Tests that the kernel boots successfully in Firecracker
#
# Usage:
#   ./test-boot.sh [options] [kernel] [rootfs]
#
# Options:
#   --verbose       Show detailed boot output and all checks
#   --check-virtio  Include VIRTIO-specific checks (for kernel development)
#   --help          Show this help message
#
# Arguments:
#   kernel          Path to kernel (default: ./build/vmlinux or /tmp/vmlinux-fresh-build)
#   rootfs          Path to rootfs (default: ./build/rootfs.ext4 or /tmp/rootfs-working.ext4)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Parse options
VERBOSE=false
CHECK_VIRTIO=false
KERNEL_PATH=""
ROOTFS_PATH=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose)
            VERBOSE=true
            shift
            ;;
        --check-virtio)
            CHECK_VIRTIO=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options] [kernel] [rootfs]"
            echo ""
            echo "Options:"
            echo "  --verbose       Show detailed boot output and all checks"
            echo "  --check-virtio  Include VIRTIO-specific checks"
            echo "  --help          Show this help message"
            echo ""
            echo "Arguments:"
            echo "  kernel          Path to kernel (default: auto-detect)"
            echo "  rootfs          Path to rootfs (default: auto-detect)"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Use default paths"
            echo "  $0 --verbose                          # Verbose output"
            echo "  $0 --check-virtio /tmp/vmlinux-test  # Check VIRTIO"
            echo "  $0 build/vmlinux build/rootfs.ext4   # Explicit paths"
            exit 0
            ;;
        *)
            if [ -z "$KERNEL_PATH" ]; then
                KERNEL_PATH="$1"
            elif [ -z "$ROOTFS_PATH" ]; then
                ROOTFS_PATH="$1"
            else
                echo "Error: Too many arguments"
                echo "Run '$0 --help' for usage"
                exit 1
            fi
            shift
            ;;
    esac
done

# Auto-detect kernel path if not provided
if [ -z "$KERNEL_PATH" ]; then
    if [ -f "./build/vmlinux" ]; then
        KERNEL_PATH="./build/vmlinux"
    elif [ -f "/tmp/vmlinux-fresh-build" ]; then
        KERNEL_PATH="/tmp/vmlinux-fresh-build"
    elif [ -f "/tmp/vmlinux-test" ]; then
        KERNEL_PATH="/tmp/vmlinux-test"
    else
        echo -e "${RED}✗ Error: Kernel not found${NC}"
        echo "Searched: ./build/vmlinux, /tmp/vmlinux-fresh-build, /tmp/vmlinux-test"
        echo "Specify kernel path: $0 /path/to/vmlinux"
        exit 1
    fi
fi

# Auto-detect rootfs path if not provided
if [ -z "$ROOTFS_PATH" ]; then
    if [ -f "./build/rootfs.ext4" ]; then
        ROOTFS_PATH="./build/rootfs.ext4"
    elif [ -f "/tmp/rootfs-working.ext4" ]; then
        ROOTFS_PATH="/tmp/rootfs-working.ext4"
    else
        echo -e "${RED}✗ Error: Rootfs not found${NC}"
        echo "Searched: ./build/rootfs.ext4, /tmp/rootfs-working.ext4"
        echo "Specify rootfs path: $0 $KERNEL_PATH /path/to/rootfs.ext4"
        exit 1
    fi
fi

# Validate files exist
if [ ! -f "$KERNEL_PATH" ]; then
    echo -e "${RED}✗ Error: Kernel not found: $KERNEL_PATH${NC}"
    exit 1
fi

if [ ! -f "$ROOTFS_PATH" ]; then
    echo -e "${RED}✗ Error: Rootfs not found: $ROOTFS_PATH${NC}"
    exit 1
fi

# Check firecracker
if ! command -v firecracker &> /dev/null; then
    echo -e "${RED}✗ Error: firecracker not found in PATH${NC}"
    echo "Install from: https://github.com/firecracker-microvm/firecracker/releases"
    exit 1
fi

# Header
echo "=========================================="
if $VERBOSE; then
    echo "KERNEL BOOT TEST - VERBOSE"
else
    echo "KERNEL BOOT TEST"
fi
echo "=========================================="
echo ""

echo "Test Configuration:"
echo "  Kernel:  $KERNEL_PATH"
echo "  Rootfs:  $ROOTFS_PATH"
if $CHECK_VIRTIO; then
    echo "  Mode:    VIRTIO validation enabled"
fi
echo ""

if $VERBOSE; then
    file "$KERNEL_PATH"
    echo ""
fi

# Create test config
TEST_CONFIG="/tmp/test_boot_config_$$.json"
rm -f "$TEST_CONFIG"
cat > "$TEST_CONFIG" << EOF
{
  "boot-source": {
    "kernel_image_path": "$KERNEL_PATH",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "$ROOTFS_PATH",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 1,
    "mem_size_mib": 256
  }
}
EOF

if $VERBOSE; then
    echo "Firecracker Config:"
    cat "$TEST_CONFIG"
    echo ""
fi

echo "=========================================="
echo "BOOTING KERNEL..."
echo "=========================================="
echo ""

# Boot with timeout
BOOT_LOG="/tmp/fc_boot_output_$$.log"
rm -f "$BOOT_LOG"

if $VERBOSE; then
    echo "Running: timeout 35 firecracker --no-api --config-file $TEST_CONFIG"
    echo "Output will be saved to: $BOOT_LOG"
    echo ""
fi

# Run firecracker and capture output
# Close unnecessary file descriptors and run in clean subshell
set +e
(
  exec 3>&- 4>&- 5>&- 6>&- 7>&- 8>&- 9>&- 2>/dev/null
  timeout 35 firecracker --no-api --config-file "$TEST_CONFIG" 2>&1
) > "$BOOT_LOG"
EXIT_STATUS=$?
set -e

# If log is still only 1 line, run directly in background to bypass any shell weirdness
if [ $(wc -l < "$BOOT_LOG") -le 1 ]; then
    if $VERBOSE; then
        echo "WARNING: Initial capture failed. Retrying with direct background execution..."
    fi
    rm -f "$BOOT_LOG"
    timeout 35 firecracker --no-api --config-file "$TEST_CONFIG" > "$BOOT_LOG" 2>&1 &
    FC_PID=$!
    wait $FC_PID || EXIT_STATUS=$?
fi

if $VERBOSE; then
    echo "Firecracker exit status: $EXIT_STATUS"
    echo "Boot output captured. Analyzing..."
    echo "Log file: $BOOT_LOG"
    wc -l "$BOOT_LOG"
    echo ""
fi

# =============================================================================
# TEST RESULTS
# =============================================================================
echo "=========================================="
echo "TEST RESULTS"
echo "=========================================="
echo ""

# Check 1: Firecracker started and booted kernel
if grep -q "Linux version" "$BOOT_LOG"; then
    echo -e "${GREEN}✓ Step 1: Firecracker VM started and kernel booted${NC}"
    if $VERBOSE; then
        grep "Linux version" "$BOOT_LOG" | head -1
    fi
else
    echo -e "${RED}✗ Step 1: Firecracker VM failed to start kernel${NC}"
    if $VERBOSE; then
        echo "  Output:"
        head -20 "$BOOT_LOG"
    fi
fi
echo ""

# Check 2: Kernel version
if grep -q "Linux version 6.1" "$BOOT_LOG"; then
    echo -e "${GREEN}✓ Step 2: Linux 6.1 kernel loaded${NC}"
    if $VERBOSE; then
        grep "Linux version" "$BOOT_LOG" | head -1
    fi
else
    ACTUAL_VERSION=$(grep "Linux version" "$BOOT_LOG" 2>/dev/null | head -1)
    if [ -z "$ACTUAL_VERSION" ]; then
        echo -e "${RED}✗ Step 2: Kernel version NOT FOUND in log${NC}"
        echo "  The kernel never booted - no version string in output"
        echo "  This means Firecracker started but the VM didn't boot"
    else
        echo -e "${YELLOW}⚠ Step 2: Kernel version mismatch${NC}"
        echo "  Expected: Linux 6.1.x"
        echo "  Found: $ACTUAL_VERSION"
    fi
fi
echo ""

# Check 3: VIRTIO-MMIO device detected (if --check-virtio)
if $CHECK_VIRTIO; then
    if grep -q "virtio-mmio: Registering device" "$BOOT_LOG"; then
        echo -e "${GREEN}✓ Step 3: VIRTIO-MMIO device detected${NC}"
        if $VERBOSE; then
            grep "virtio-mmio: Registering device" "$BOOT_LOG"
        fi
    else
        echo -e "${RED}✗ Step 3: VIRTIO-MMIO device NOT detected${NC}"
        echo "  This means CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES may not be working"
    fi
    echo ""
fi

# Check 4: VIRTIO_BLK driver loaded
if grep -q "virtio_blk virtio" "$BOOT_LOG"; then
    echo -e "${GREEN}✓ Step 4: VIRTIO_BLK driver loaded${NC}"
    if $VERBOSE; then
        grep "virtio_blk virtio" "$BOOT_LOG" | head -1
    fi
else
    echo -e "${RED}✗ Step 4: VIRTIO_BLK driver NOT loaded${NC}"
fi
echo ""

# Check 5: Block device detected
if grep -q "\[vda\]" "$BOOT_LOG"; then
    echo -e "${GREEN}✓ Step 5: Block device [vda] detected${NC}"
    if $VERBOSE; then
        grep "\[vda\]" "$BOOT_LOG"
    fi
else
    echo -e "${RED}✗ Step 5: Block device [vda] NOT detected${NC}"
fi
echo ""

# Check 6: EXT4 filesystem mounted
if grep -q "EXT4-fs (vda): mounted filesystem" "$BOOT_LOG" && grep -q "VFS: Mounted root" "$BOOT_LOG"; then
    echo -e "${GREEN}✓ Step 6: EXT4 filesystem mounted successfully${NC}"
    if $VERBOSE; then
        grep "EXT4-fs (vda): mounted filesystem" "$BOOT_LOG" | head -1
    fi
else
    echo -e "${RED}✗ Step 6: EXT4 filesystem NOT mounted${NC}"
fi
echo ""

# Check 7: No kernel panic
if grep -q "Kernel panic" "$BOOT_LOG"; then
    echo -e "${RED}✗ Step 7: KERNEL PANIC DETECTED${NC}"
    if $VERBOSE; then
        grep "Kernel panic" "$BOOT_LOG"
    fi
else
    echo -e "${GREEN}✓ Step 7: No kernel panic${NC}"
fi
echo ""

# =============================================================================
# FINAL VERDICT
# =============================================================================
echo "=========================================="

# Count successes - grep -c returns 0 on no match, so we don't need || echo
VIRTIO_MMIO=$(grep -c "virtio-mmio: Registering device" "$BOOT_LOG" 2>/dev/null)
VIRTIO_BLK=$(grep -c "virtio_blk virtio" "$BOOT_LOG" 2>/dev/null)
BLOCK_DEVICE=$(grep -cE "\[vda\]" "$BOOT_LOG" 2>/dev/null)
EXT4_MOUNT=$(grep -cE "EXT4-fs.*mounted|VFS: Mounted root" "$BOOT_LOG" 2>/dev/null)
KERNEL_PANIC=$(grep -c "Kernel panic" "$BOOT_LOG" 2>/dev/null)

# Default to 0 if grep failed completely
VIRTIO_MMIO=${VIRTIO_MMIO:-0}
VIRTIO_BLK=${VIRTIO_BLK:-0}
BLOCK_DEVICE=${BLOCK_DEVICE:-0}
EXT4_MOUNT=${EXT4_MOUNT:-0}
KERNEL_PANIC=${KERNEL_PANIC:-0}

# Determine pass/fail
# ALL required checks must pass
REQUIRED_CHECKS_PASS=true

# Check if we have any output at all
if [ $(wc -l < "$BOOT_LOG") -le 1 ]; then
    echo -e "${RED}CRITICAL: Firecracker produced no output. Log has only 1 line.${NC}"
    echo "This means Firecracker started but the VM never booted."
    echo ""
    REQUIRED_CHECKS_PASS=false
fi

# VIRTIO-MMIO only required if --check-virtio
if $CHECK_VIRTIO; then
    if [ "$VIRTIO_MMIO" -eq 0 ]; then
        REQUIRED_CHECKS_PASS=false
    fi
fi

# These are ALWAYS required for boot success
if [ "$VIRTIO_BLK" -eq 0 ]; then
    REQUIRED_CHECKS_PASS=false
fi

if [ "$BLOCK_DEVICE" -eq 0 ]; then
    REQUIRED_CHECKS_PASS=false
fi

if [ "$EXT4_MOUNT" -eq 0 ]; then
    REQUIRED_CHECKS_PASS=false
fi

if [ "$KERNEL_PANIC" -gt 0 ]; then
    REQUIRED_CHECKS_PASS=false
fi

# Display final result
echo ""
if $REQUIRED_CHECKS_PASS; then
    echo -e "${GREEN}✓✓✓ ALL TESTS PASSED ✓✓✓${NC}"
    echo "=========================================="
    EXIT_CODE=0

    # Clean up on success unless verbose
    if ! $VERBOSE; then
        rm -f "$BOOT_LOG" "$TEST_CONFIG"
    else
        echo ""
        echo "Artifacts preserved:"
        echo "  Boot log: $BOOT_LOG"
        echo "  Config:   $TEST_CONFIG"
    fi
else
    echo -e "${RED}✗✗✗ TESTS FAILED ✗✗✗${NC}"
    echo "=========================================="
    echo ""
    echo "Boot log: $BOOT_LOG ($(wc -l < $BOOT_LOG) lines)"
    echo "Config:   $TEST_CONFIG"
    echo ""
    echo "To debug:"
    echo "  1. Check log: cat $BOOT_LOG"
    echo "  2. Run manually: sudo firecracker --no-api --config-file $TEST_CONFIG"
    echo "  3. Re-run verbose: sudo $0 --verbose $KERNEL_PATH $ROOTFS_PATH"

    EXIT_CODE=1
fi

echo ""
exit $EXIT_CODE
