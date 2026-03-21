#!/bin/bash
# Quick debug test for IP issue

# Kill any existing daemons
if pgrep -f "nanofused.*config.dev.yaml" > /dev/null; then
    echo "Stopping existing daemon..."
    pkill -f "nanofused.*config.dev.yaml"
    sleep 2
fi

echo "Setting up test environment..."
rm -rf /tmp/nanofuse
mkdir -p /tmp/nanofuse/images/nanofuse-base/latest
cp images/base/build/rootfs.ext4 /tmp/nanofuse/images/nanofuse-base/latest/
cp images/base/build/vmlinux /tmp/nanofuse/images/nanofuse-base/latest/
cp images/base/build/manifest.json /tmp/nanofuse/images/nanofuse-base/latest/
chmod 664 /tmp/nanofuse/images/nanofuse-base/latest/rootfs.ext4

echo "Starting daemon..."
./bin/nanofused --config ./config.dev.yaml > /tmp/debug-daemon.log 2>&1 &
DAEMON_PID=$!
sleep 5

echo "Registering image..."
./bin/register-local-image /tmp/nanofuse/nanofuse.db "nanofuse-base:latest" \
    /tmp/nanofuse/images/nanofuse-base/latest/rootfs.ext4 \
    /tmp/nanofuse/images/nanofuse-base/latest/vmlinux

echo ""
echo "Creating VM..."
curl -s -X POST "http://127.0.0.1:8080/vms" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "nanofuse-base:latest",
    "name": "debug-ip-test",
    "config": {
      "vcpus": 2,
      "memory_mib": 512
    }
  }' | jq '.'

echo ""
echo "=== Daemon log with debug info ==="
grep -E "DEBUG|Configured network" /tmp/debug-daemon.log

echo ""
echo "=== Database content ==="
sqlite3 /tmp/nanofuse/nanofuse.db "SELECT config_json FROM vms WHERE name='debug-ip-test';" | jq '.network'

echo ""
echo "Cleaning up..."
sudo kill $DAEMON_PID
rm -rf /tmp/nanofuse
