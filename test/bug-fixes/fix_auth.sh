#!/bin/bash
# Fix authentication for nanofused daemon
# The daemon runs as root, so it needs root's Docker config

echo "Copying Docker credentials to root user..."
sudo mkdir -p /root/.docker
sudo cp ~/.docker/config.json /root/.docker/

echo "Restarting daemon to pick up new credentials..."
sudo pkill -9 nanofused
sleep 2
sudo /home/jpoley/ps/nanofuse/bin/nanofused > /tmp/nanofused.log 2>&1 &
sleep 3

echo ""
echo "Testing image pull again..."
./bin/nanofuse --api-url http://localhost:8080 image pull --default

echo ""
echo "Check status with:"
echo "  ./bin/nanofuse --api-url http://localhost:8080 image list"
