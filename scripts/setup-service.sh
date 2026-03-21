#!/bin/bash
# Setup nanofused systemd service

set -e

if [ "$EUID" -ne 0 ]; then
    echo "ERROR: Must run as root"
    echo "Usage: sudo bash setup-service.sh"
    exit 1
fi

echo "=== Setting up nanofused service ==="

# Create data directory
mkdir -p /var/lib/nanofuse
chown root:root /var/lib/nanofuse
chmod 755 /var/lib/nanofuse
echo "✓ Created /var/lib/nanofuse"

# Install service file
cp nanofused.service /etc/systemd/system/nanofused.service
echo "✓ Installed service file"

# Reload systemd
systemctl daemon-reload
echo "✓ Reloaded systemd"

# Enable and start service
systemctl enable nanofused
systemctl restart nanofused
echo "✓ Enabled and started service"

# Wait a moment for service to start
sleep 2

# Show status
echo ""
echo "=== Service Status ==="
systemctl status nanofused --no-pager -l || true

echo ""
if systemctl is-active --quiet nanofused; then
    echo "✓ nanofused is running"
    echo ""
    echo "You can now use nanofuse commands:"
    echo "  nanofuse image pull --default"
    echo "  nanofuse vm create myvm --image default --vcpus 2 --memory 512"
    echo "  nanofuse vm start myvm"
else
    echo "✗ nanofused failed to start"
    echo ""
    echo "Check logs with: journalctl -u nanofused -f"
    exit 1
fi
