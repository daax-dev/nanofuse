#!/bin/bash
# Go Runtime Layer - Post-Install Hook
# This script runs after the layer rootfs is extracted to the target image.
#
# Environment variables available:
#   LAYER_NAME - Name of this layer
#   LAYER_VERSION - Version of this layer
#   ROOTFS_PATH - Path to the mounted rootfs
#   CONFIG_* - Layer config values from manifest

set -euo pipefail

echo "[go-runtime] Running post-install hook..."

# Ensure CA certificates are accessible
if [[ -d "${ROOTFS_PATH}/etc/ssl/certs" ]]; then
    chmod 755 "${ROOTFS_PATH}/etc/ssl/certs"
fi

# Set default timezone if specified
TIMEZONE="${CONFIG_TIMEZONE:-UTC}"
if [[ -f "${ROOTFS_PATH}/usr/share/zoneinfo/${TIMEZONE}" ]]; then
    ln -sf "/usr/share/zoneinfo/${TIMEZONE}" "${ROOTFS_PATH}/etc/localtime"
    echo "${TIMEZONE}" > "${ROOTFS_PATH}/etc/timezone"
fi

echo "[go-runtime] Go runtime layer installed successfully"
