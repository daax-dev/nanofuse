#!/bin/bash
# debug-boot.sh - Boot debug kernel directly with Firecracker (bypasses nanofuse)
# Use this to verify Firecracker + kernel + rootfs works independently
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
FIXTURES="$PROJECT_ROOT/test/fixtures/debug-kernel"
WORK_DIR="/tmp/nanofuse-debug-$$"

# Check for kernel/rootfs
if [[ ! -f "$FIXTURES/vmlinux.bin" ]] || [[ ! -f "$FIXTURES/rootfs.ext4" ]]; then
    echo "ERROR: Debug kernel files not found in $FIXTURES"
    echo "Run: curl -L -o $FIXTURES/vmlinux.bin https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin"
    echo "     curl -L -o $FIXTURES/rootfs.ext4 https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/rootfs/bionic.rootfs.ext4"
    exit 1
fi

# Check for firecracker
if ! command -v firecracker &>/dev/null; then
    echo "ERROR: firecracker not found in PATH"
    exit 1
fi

# Check KVM access
if [[ ! -w /dev/kvm ]]; then
    echo "ERROR: Cannot access /dev/kvm - add user to kvm group"
    exit 1
fi

echo "=== NanoFuse Debug Boot ==="
echo "Kernel:  $FIXTURES/vmlinux.bin"
echo "Rootfs:  $FIXTURES/rootfs.ext4"
echo ""

# Setup work directory
mkdir -p "$WORK_DIR"
trap "rm -rf $WORK_DIR" EXIT

# Create a writable copy of rootfs (Firecracker needs write access)
echo "Creating writable rootfs copy..."
cp "$FIXTURES/rootfs.ext4" "$WORK_DIR/rootfs.ext4"

# Create Firecracker config
cat > "$WORK_DIR/config.json" << EOF
{
  "boot-source": {
    "kernel_image_path": "$FIXTURES/vmlinux.bin",
    "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "$WORK_DIR/rootfs.ext4",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 2,
    "mem_size_mib": 512,
    "smt": false
  }
}
EOF

echo "Config: $WORK_DIR/config.json"
echo ""
echo "Starting Firecracker (Ctrl+C to exit)..."
echo "========================================="

# Run Firecracker
firecracker --no-api --config-file "$WORK_DIR/config.json"
