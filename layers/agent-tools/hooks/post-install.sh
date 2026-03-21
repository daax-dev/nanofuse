#!/bin/bash
# Agent Tools Layer - Post-Install Hook
# This script runs after the layer rootfs is extracted to the target image.
#
# Environment variables available:
#   LAYER_NAME - Name of this layer
#   LAYER_VERSION - Version of this layer
#   ROOTFS_PATH - Path to the mounted rootfs
#   CONFIG_* - Layer config values from manifest

set -euo pipefail

echo "[agent-tools] Running post-install hook..."

# Create .claude config directory for all users
mkdir -p "${ROOTFS_PATH}/etc/skel/.claude"
mkdir -p "${ROOTFS_PATH}/root/.claude"

# Set up git config defaults
cat > "${ROOTFS_PATH}/etc/gitconfig" << 'EOF'
[safe]
    directory = *
[init]
    defaultBranch = main
[core]
    editor = nano
EOF

# Ensure agent binaries are executable
for bin in claude codex gemini; do
    if [[ -f "${ROOTFS_PATH}/usr/local/bin/${bin}" ]]; then
        chmod 755 "${ROOTFS_PATH}/usr/local/bin/${bin}"
    fi
done

echo "[agent-tools] Agent tools layer installed successfully"
