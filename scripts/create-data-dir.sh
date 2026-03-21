#!/bin/bash
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"

echo "=== Creating /data directory in rootfs ==="
debugfs -w -R "mkdir /data" "$ROOTFS" 2>&1 || echo "Directory may already exist"
debugfs -w -R "sif /data mode 040755" "$ROOTFS" 2>&1
debugfs -w -R "sif /data uid 0" "$ROOTFS" 2>&1
debugfs -w -R "sif /data gid 0" "$ROOTFS" 2>&1

echo ""
echo "Verifying:"
debugfs -R "stat /data" "$ROOTFS" 2>&1 | head -5

echo ""
echo "✓ /data directory created"
echo ""
echo "Now restart VM and test again"
