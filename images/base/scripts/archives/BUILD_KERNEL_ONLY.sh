#!/bin/bash
set -euo pipefail

# Step 1: Build kernel only (run with sudo)
# This part needs sudo for Docker

cd /home/jpoley/src/_mine/nanofuse/images/base

# Remove old kernel if it exists
rm -f /tmp/vmlinux-test 2>/dev/null || sudo rm -f /tmp/vmlinux-test 2>/dev/null || true

docker image rm nanofuse-kernel-builder:latest 2>/dev/null || true
docker build --no-cache -f Dockerfile.kernel -t nanofuse-kernel-builder:latest .
docker run --rm nanofuse-kernel-builder:latest cat /vmlinux > /tmp/vmlinux-test

# Fix ownership if running via sudo
if [ -n "${SUDO_UID:-}" ]; then
    chown ${SUDO_UID}:${SUDO_GID} /tmp/vmlinux-test
fi

echo "Kernel built: /tmp/vmlinux-test"
ls -lh /tmp/vmlinux-test
