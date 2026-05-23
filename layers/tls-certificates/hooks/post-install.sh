#!/bin/bash
# TLS Certificates Layer - Post-Install Hook
# This script runs after the layer rootfs is extracted to the target image.
#
# Environment variables available:
#   LAYER_NAME - Name of this layer
#   LAYER_VERSION - Version of this layer
#   ROOTFS_PATH - Path to the mounted rootfs
#   CONFIG_CA_BUNDLE_PATH - CA bundle path
#   CONFIG_CERT_DIR - Certificate directory
#   CONFIG_AUTO_RENEW - Enable auto renewal

set -euo pipefail

echo "[tls-certificates] Running post-install hook..."

# Validate positive integer with optional upper bound.
# Usage: validate_positive_int VALUE NAME [MAX_VALUE]
# Note: This function is duplicated in observability/hooks/post-install.sh with identical logic.
# Duplication is intentional: each layer's post-install hook must be self-contained since layers
# may be installed independently. A shared library would add complexity and cross-layer dependencies.
validate_positive_int() {
    local value="$1"
    local name="$2"
    local max_value="${3:-}"
    if ! [[ "${value}" =~ ^[0-9]+$ ]] || [[ "${value}" -lt 1 ]]; then
        echo "[tls-certificates] Error: ${name} must be a positive integer, got '${value}'" >&2
        return 1
    fi
    if [[ -n "${max_value}" ]] && [[ "${value}" -gt "${max_value}" ]]; then
        echo "[tls-certificates] Error: ${name} must be between 1 and ${max_value}, got '${value}'" >&2
        return 1
    fi
    return 0
}

# Validate octal file mode (accepts 3-digit: 750, or 4-digit: 0750)
validate_octal_mode() {
    local mode="$1"
    if ! [[ "${mode}" =~ ^(0[0-7]{3}|[0-7]{3})$ ]]; then
        echo "[tls-certificates] Error: mode must be 3 or 4 octal digits (e.g., 750 or 0750), got '${mode}'" >&2
        return 1
    fi
    return 0
}

# Note: previously there was a duplicate validate_octal_mode definition here that only accepted
# 3-digit modes (000-777). It has been removed so that the original implementation above,
# which supports both 3-digit (750) and 4-digit (0750) octal modes, is used.

CA_BUNDLE_PATH="${CONFIG_CA_BUNDLE_PATH:-/etc/ssl/certs/ca-certificates.crt}"
CERT_DIR="${CONFIG_CERT_DIR:-/etc/nanofuse/certs}"
AUTO_RENEW="${CONFIG_AUTO_RENEW:-true}"

RENEWAL_DAYS="${CONFIG_RENEWAL_DAYS_BEFORE:-30}"
if ! validate_positive_int "${RENEWAL_DAYS}" "renewal_days_before" 3650; then
    echo "[tls-certificates] Using default renewal_days_before 30" >&2
    RENEWAL_DAYS="30"
fi

# Permissions: 750 allows group access; private keys use 600
CERT_DIR_MODE="${CONFIG_CERT_DIR_MODE:-750}"
if ! validate_octal_mode "${CERT_DIR_MODE}"; then
    echo "[tls-certificates] Using default mode 750" >&2
    CERT_DIR_MODE="750"
fi

# Create directories with configurable permissions
mkdir -p "${ROOTFS_PATH}${CERT_DIR}"
# Normalize mode to 4-digit format: if already starts with 0, use as-is; otherwise prepend 0
if [[ "${CERT_DIR_MODE}" =~ ^0 ]]; then
    chmod "${CERT_DIR_MODE}" "${ROOTFS_PATH}${CERT_DIR}"
else
    chmod "0${CERT_DIR_MODE}" "${ROOTFS_PATH}${CERT_DIR}"
fi

# Create cert-manager configuration file
mkdir -p "${ROOTFS_PATH}/etc/nanofuse"
printf 'RENEWAL_DAYS=%s\n' "${RENEWAL_DAYS}" > "${ROOTFS_PATH}/etc/nanofuse/cert-manager.conf"
chmod 644 "${ROOTFS_PATH}/etc/nanofuse/cert-manager.conf"

# Create certificate renewal timer
cat > "${ROOTFS_PATH}/etc/systemd/system/cert-renewal.timer" << EOF
[Unit]
Description=Certificate Renewal Timer
Documentation=https://github.com/daax-dev/nanofuse

[Timer]
OnCalendar=daily
RandomizedDelaySec=3600
Persistent=true

[Install]
WantedBy=timers.target
EOF

# Create certificate renewal service
cat > "${ROOTFS_PATH}/etc/systemd/system/cert-renewal.service" << EOF
[Unit]
Description=Certificate Renewal Service
Documentation=https://github.com/daax-dev/nanofuse
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/cert-manager renew
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable the timer if auto-renew is enabled
if [[ "$AUTO_RENEW" == "true" ]]; then
    mkdir -p "${ROOTFS_PATH}/etc/systemd/system/timers.target.wants"
    if ! ln -sf /etc/systemd/system/cert-renewal.timer \
        "${ROOTFS_PATH}/etc/systemd/system/timers.target.wants/cert-renewal.timer"; then
        echo "[tls-certificates] Warning: failed to enable cert-renewal.timer symlink" >&2
    else
        echo "[tls-certificates] Auto-renewal timer enabled"
    fi
else
    echo "[tls-certificates] Auto-renewal disabled"
fi

echo "[tls-certificates] Configuration:"
echo "  ca_bundle_path: ${CA_BUNDLE_PATH}"
echo "  cert_dir: ${CERT_DIR}"
echo "  auto_renew: ${AUTO_RENEW}"
echo "  renewal_days_before: ${RENEWAL_DAYS}"
echo "[tls-certificates] TLS certificates layer installed successfully"
