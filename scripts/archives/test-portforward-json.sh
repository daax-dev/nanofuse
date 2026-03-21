#!/bin/bash
# Quick test to verify port forward JSON encoding/decoding

set -euo pipefail

# Configuration
DATA_DIR="/tmp/nanofuse"
DB_PATH="$DATA_DIR/nanofuse.db"
IMAGE_NAME="nanofuse-base:latest"

# Start daemon if not running
if ! curl -s http://127.0.0.1:8080/health > /dev/null 2>&1; then
    echo "Daemon not running. Start it with: sudo ./bin/nanofused --config config.dev.yaml"
    exit 1
fi

# Check if image is registered
echo "Checking if image is registered..."
if ! sqlite3 "$DB_PATH" "SELECT digest FROM images WHERE digest = '$IMAGE_NAME';" 2>/dev/null | grep -q .; then
    echo "Image not registered. Registering..."

    # Check if image files exist
    if [ ! -f "images/base/build/rootfs.ext4" ] || [ ! -f "images/base/build/vmlinux" ]; then
        echo "Error: Base image files not found. Build them first:"
        echo "  sudo ./images/base/build.sh"
        exit 1
    fi

    # Create data directory structure
    sudo mkdir -p "$DATA_DIR/images/nanofuse-base/latest"

    # Copy image files
    sudo cp images/base/build/rootfs.ext4 "$DATA_DIR/images/nanofuse-base/latest/"
    sudo cp images/base/build/vmlinux "$DATA_DIR/images/nanofuse-base/latest/"
    sudo cp images/base/build/manifest.json "$DATA_DIR/images/nanofuse-base/latest/"

    # Register image
    sudo ./bin/register-local-image "$DB_PATH" "$IMAGE_NAME" \
        "$DATA_DIR/images/nanofuse-base/latest/rootfs.ext4" \
        "$DATA_DIR/images/nanofuse-base/latest/vmlinux"

    echo "✓ Image registered"
else
    echo "✓ Image already registered"
fi

echo ""

# Create a VM with port forwards using direct API call
echo "Creating VM with port forwards via API..."
RESPONSE=$(curl -s -X POST http://127.0.0.1:8080/vms \
    -H "Content-Type: application/json" \
    -d '{
  "name": "test-pf-json",
  "image": "nanofuse-base:latest",
  "config": {
    "vcpus": 1,
    "memory_mib": 256,
    "network": {
      "port_forwards": [
        {
          "host_port": 9999,
          "vm_port": 80,
          "protocol": "tcp"
        }
      ]
    }
  }
}')

echo "Response:"
echo "$RESPONSE" | jq .

# Check if it succeeded
if echo "$RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✓ VM created successfully with port forwards"
    VM_ID=$(echo "$RESPONSE" | jq -r '.id')

    # Inspect the VM to verify port forwards were saved
    echo ""
    echo "Inspecting VM..."
    curl -s "http://127.0.0.1:8080/vms/$VM_ID" | jq '.config.network.port_forwards'

    # Cleanup
    echo ""
    echo "Cleaning up..."
    curl -s -X DELETE "http://127.0.0.1:8080/vms/$VM_ID?force=true" > /dev/null
    echo "✓ Cleanup complete"
else
    echo "✗ Failed to create VM"
    echo "$RESPONSE" | jq .
    exit 1
fi
