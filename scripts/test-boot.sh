#!/bin/bash
# test-boot.sh - Test boot a composed NanoFuse image in Firecracker
#
# Usage:
#   ./scripts/test-boot.sh [image-path] [kernel-path]
#
# Examples:
#   ./scripts/test-boot.sh build/test-boot.ext4
#   ./scripts/test-boot.sh build/falcondev-agents.ext4 test/fixtures/debug-kernel/vmlinux-6.1.155

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# Default paths
DEFAULT_IMAGE="${PROJECT_ROOT}/build/test-boot.ext4"
DEFAULT_KERNEL="${PROJECT_ROOT}/test/fixtures/debug-kernel/vmlinux-6.1.155"
SOCKET_PATH="/tmp/firecracker-test-$$.sock"

cleanup() {
    log_info "Cleaning up..."
    rm -f "$SOCKET_PATH" 2>/dev/null || true
    # Kill any lingering firecracker process
    pkill -f "firecracker.*$SOCKET_PATH" 2>/dev/null || true
}

trap cleanup EXIT

check_requirements() {
    if ! command -v firecracker &>/dev/null; then
        log_error "firecracker not found. Please install it first."
        exit 1
    fi

    # Check for /dev/kvm
    if [[ ! -c /dev/kvm ]]; then
        log_warn "/dev/kvm not available - VM will be slow without KVM"
    elif [[ ! -w /dev/kvm ]]; then
        log_warn "/dev/kvm not writable - add user to kvm group: sudo usermod -aG kvm $USER"
    fi
}

create_config() {
    local image_path="$1"
    local kernel_path="$2"
    local config_file="$3"

    cat > "$config_file" << EOF
{
  "boot-source": {
    "kernel_image_path": "${kernel_path}",
    "boot_args": "console=ttyS0 root=/dev/vda rw init=/sbin/init panic=1 reboot=k quiet loglevel=3"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "${image_path}",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 2,
    "mem_size_mib": 1024
  }
}
EOF
}

run_firecracker() {
    local image_path="$1"
    local kernel_path="$2"

    local config_file=$(mktemp)
    trap "rm -f '$config_file'" RETURN

    log_info "Creating Firecracker config..."
    create_config "$image_path" "$kernel_path" "$config_file"

    log_info "Starting Firecracker..."
    log_info "  Kernel: $kernel_path"
    log_info "  Image: $image_path"
    log_info "  Memory: 1024 MiB"
    log_info "  vCPUs: 2"
    echo ""
    log_info "=== VM Console Output ==="
    log_info "(Press Ctrl+C to stop the VM)"
    echo ""

    # Run firecracker with the config
    firecracker --no-api --config-file "$config_file"
}

usage() {
    echo "Usage: $0 [image-path] [kernel-path]"
    echo ""
    echo "Test boot a composed NanoFuse image in Firecracker."
    echo ""
    echo "Arguments:"
    echo "  image-path    Path to ext4 rootfs image (default: build/test-boot.ext4)"
    echo "  kernel-path   Path to kernel binary (default: test/fixtures/debug-kernel/vmlinux-6.1.155)"
    echo ""
    echo "Examples:"
    echo "  $0"
    echo "  $0 build/test-boot.ext4"
    echo "  $0 build/falcondev-agents.ext4 test/fixtures/debug-kernel/vmlinux-5.10.245-no-acpi"
}

main() {
    local image_path="${1:-$DEFAULT_IMAGE}"
    local kernel_path="${2:-$DEFAULT_KERNEL}"

    if [[ "$image_path" == "-h" || "$image_path" == "--help" ]]; then
        usage
        exit 0
    fi

    check_requirements

    # Resolve relative paths
    if [[ ! "$image_path" = /* ]]; then
        image_path="${PROJECT_ROOT}/${image_path}"
    fi
    if [[ ! "$kernel_path" = /* ]]; then
        kernel_path="${PROJECT_ROOT}/${kernel_path}"
    fi

    # Verify files exist
    if [[ ! -f "$image_path" ]]; then
        log_error "Image not found: $image_path"
        log_info "Run ./scripts/compose-image.sh first to create an image"
        exit 1
    fi

    if [[ ! -f "$kernel_path" ]]; then
        log_error "Kernel not found: $kernel_path"
        exit 1
    fi

    run_firecracker "$image_path" "$kernel_path"
}

main "$@"
