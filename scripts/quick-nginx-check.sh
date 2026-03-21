#!/bin/bash
# Quick nginx check without mounting
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"

echo "=== Extracting nginx service file ==="
debugfs -R "cat /lib/systemd/system/nginx.service" "$ROOTFS" 2>/dev/null || \
debugfs -R "cat /usr/lib/systemd/system/nginx.service" "$ROOTFS" 2>/dev/null || \
echo "Could not extract service file"

echo ""
echo "=== The issue is likely ==="
echo "Nginx needs /run/nginx directory created at boot"
echo "The fix-nginx-now.sh script should have added tmpfiles.d config"
echo ""
echo "Let me create a comprehensive fix..."
