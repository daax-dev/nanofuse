#!/bin/bash
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"

echo "=== Creating exec service override for debugging ==="

cat > /tmp/todo-backend.service <<'EOF'
[Unit]
Description=Todo App Backend Service
After=network.target

[Service]
Type=exec
User=root
WorkingDirectory=/data
ExecStart=/usr/local/bin/todo-server -db-path /data/todos.db -http-port 8080 -grpc-port 9090
Restart=no
StandardOutput=console
StandardError=console

[Install]
WantedBy=multi-user.target
EOF

debugfs -w -R "rm /etc/systemd/system/todo-backend.service" "$ROOTFS" 2>&1 || true
debugfs -w -R "write /tmp/todo-backend.service /etc/systemd/system/todo-backend.service" "$ROOTFS"

echo ""
echo "=== Testing backend binary directly ==="
# Extract and test the binary
debugfs -R "dump /usr/local/bin/todo-server /tmp/todo-server-test" "$ROOTFS" 2>&1
chmod +x /tmp/todo-server-test
echo "Binary extracted, checking..."
file /tmp/todo-server-test
ldd /tmp/todo-server-test 2>&1 | head -20 || echo "Binary may be statically linked"

echo ""
echo "✓ Service updated for debugging"
echo ""
echo "Now restart and the backend errors will be visible"
