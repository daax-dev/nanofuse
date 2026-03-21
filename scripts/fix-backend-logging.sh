#!/bin/bash
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"

echo "=== Fixing backend service to log to console ==="

# Create new service file with console logging
cat > /tmp/todo-backend.service <<'EOF'
[Unit]
Description=Todo App Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/usr/local/bin
ExecStart=/usr/local/bin/todo-server -db-path /data/todos.db -http-port 8080 -grpc-port 9090
Restart=always
RestartSec=5
StandardOutput=tty
StandardError=tty

[Install]
WantedBy=multi-user.target
EOF

# Write to rootfs
debugfs -w -R "rm /etc/systemd/system/todo-backend.service" "$ROOTFS" 2>&1 || true
debugfs -w -R "write /tmp/todo-backend.service /etc/systemd/system/todo-backend.service" "$ROOTFS"

echo "✓ Updated todo-backend service to log to console"
echo ""
echo "Now restart VM - you'll see backend errors in console log"
