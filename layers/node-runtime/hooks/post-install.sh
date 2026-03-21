#!/bin/bash
# Node.js Runtime Layer - Post-Install Hook
# This script runs after the layer rootfs is extracted to the target image.
#
# Environment variables available:
#   LAYER_NAME - Name of this layer
#   LAYER_VERSION - Version of this layer
#   ROOTFS_PATH - Path to the mounted rootfs
#   CONFIG_* - Layer config values from manifest

set -euo pipefail

echo "[node-runtime] Running post-install hook..."

# Verify Node.js is accessible
if [[ -x "${ROOTFS_PATH}/usr/local/bin/node" ]]; then
    echo "[node-runtime] Node.js binary found"
fi

# Set npm global config for non-root installs (if needed)
if [[ -d "${ROOTFS_PATH}/usr/local/lib/node_modules" ]]; then
    chmod -R 755 "${ROOTFS_PATH}/usr/local/lib/node_modules"
fi

echo "[node-runtime] Node.js runtime layer installed successfully"
