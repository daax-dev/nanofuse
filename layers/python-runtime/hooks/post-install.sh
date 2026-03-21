#!/bin/bash
# Python Runtime Layer - Post-Install Hook
# This script runs after the layer rootfs is extracted to the target image.
#
# Environment variables available:
#   LAYER_NAME - Name of this layer
#   LAYER_VERSION - Version of this layer
#   ROOTFS_PATH - Path to the mounted rootfs
#   CONFIG_* - Layer config values from manifest

set -euo pipefail

echo "[python-runtime] Running post-install hook..."

# Create symlinks for python/pip if not present
if [[ ! -L "${ROOTFS_PATH}/usr/local/bin/python" ]]; then
    ln -sf /usr/bin/python3 "${ROOTFS_PATH}/usr/local/bin/python"
fi

if [[ ! -L "${ROOTFS_PATH}/usr/local/bin/pip" ]]; then
    ln -sf /usr/bin/pip3 "${ROOTFS_PATH}/usr/local/bin/pip"
fi

# Ensure pip is up to date (will run inside chroot during actual build)
echo "[python-runtime] Python runtime layer installed successfully"
