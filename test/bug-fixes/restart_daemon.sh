#!/bin/bash
# Restart the nanofused daemon

echo "Stopping any old daemon..."
sudo pkill -9 nanofused 2>/dev/null || true
sleep 2

echo "Removing old socket..."
sudo rm -f /tmp/nanofused.sock

echo "Starting daemon..."
sudo rm -f /tmp/nanofused.log
sudo /home/jpoley/ps/nanofuse/bin/nanofused > /tmp/nanofused.log 2>&1 &
sleep 3

echo ""
echo "Checking daemon status..."
if pgrep -f "nanofused" > /dev/null; then
    echo "✓ Daemon running"
else
    echo "✗ Daemon failed to start"
    if [ -f /tmp/nanofused.log ]; then
        echo "Logs:"
        sudo cat /tmp/nanofused.log
    fi
    exit 1
fi

echo ""
echo "Checking listeners..."
if [ -f /tmp/nanofused.log ]; then
    sudo grep -i "listening" /tmp/nanofused.log
fi

echo ""
echo "Testing API..."
curl -s http://localhost:8080/health | jq '.'

echo ""
echo "Daemon ready!"
