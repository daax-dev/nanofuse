#!/bin/bash
# register-debug-kernel.sh - Register debug kernel for testing (requires sudo)
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

DB_PATH="/var/lib/nanofuse/nanofuse.db"
FIXTURES="$PROJECT_ROOT/test/fixtures/debug-kernel"
REGISTER_TOOL="$PROJECT_ROOT/bin/register-local-image"

# Check if we're root
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root (database is owned by root)"
    echo "Run: sudo $0"
    exit 1
fi

# Check for kernel/rootfs
if [[ ! -f "$FIXTURES/vmlinux.bin" ]] || [[ ! -f "$FIXTURES/rootfs.ext4" ]]; then
    echo "ERROR: Debug kernel files not found in $FIXTURES"
    exit 1
fi

# Check for register tool
if [[ ! -f "$REGISTER_TOOL" ]]; then
    echo "Building register-local-image..."
    cd "$PROJECT_ROOT"
    CGO_ENABLED=1 go build -o bin/register-local-image ./register-local-image.go
fi

# Check database exists
if [[ ! -f "$DB_PATH" ]]; then
    echo "ERROR: Database not found at $DB_PATH"
    echo "Is the nanofuse daemon running?"
    exit 1
fi

echo "=== Registering Debug Kernel ==="
echo "DB:      $DB_PATH"
echo "Rootfs:  $FIXTURES/rootfs.ext4"
echo "Kernel:  $FIXTURES/vmlinux.bin"
echo ""

# Register the image
"$REGISTER_TOOL" "$DB_PATH" "debug:latest" "$FIXTURES/rootfs.ext4" "$FIXTURES/vmlinux.bin" x86_64

echo ""
echo "Now you can create VMs with: nanofuse vm create debug:latest test-vm"
