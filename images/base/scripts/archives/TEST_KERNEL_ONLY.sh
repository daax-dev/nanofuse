#!/bin/bash
set -euo pipefail

# Step 2: Test kernel (run WITHOUT sudo as your user)
# This part should NOT use sudo

cd /home/jpoley/src/_mine/nanofuse/images/base

echo "Testing kernel: /tmp/vmlinux-test"
./test-boot.sh --verbose --check-virtio /tmp/vmlinux-test
