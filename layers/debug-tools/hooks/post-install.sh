#!/bin/bash
# Debug Tools Layer - Post-Install Hook
# This script runs after the layer rootfs is extracted to the target image.
#
# Environment variables available:
#   LAYER_NAME - Name of this layer
#   LAYER_VERSION - Version of this layer
#   ROOTFS_PATH - Path to the mounted rootfs
#   CONFIG_* - Layer config values from manifest

set -euo pipefail

echo "[debug-tools] Running post-install hook..."

# Ensure all debug tool binaries are executable
for bin in strace ltrace gdb tcpdump htop curl wget; do
    if [[ -f "${ROOTFS_PATH}/usr/bin/${bin}" ]]; then
        chmod 755 "${ROOTFS_PATH}/usr/bin/${bin}"
        echo "[debug-tools] Verified ${bin} is executable"
    fi
done

# SECURITY NOTE: We intentionally do NOT grant CAP_NET_RAW to tcpdump here.
# Granting cap_net_raw+ep via setcap would allow any user who can run tcpdump
# to capture network traffic, which is unsafe in multi-tenant or production
# environments. If packet capture is required, run tcpdump with explicit
# elevated privileges (e.g., as root) in a controlled debug context.
#
# For convenience, we install a runtime init script that operators can run
# manually if they explicitly want to enable this capability.

# Create runtime capabilities init script (for optional manual use)
cat > "${ROOTFS_PATH}/usr/local/bin/init-debug-capabilities" << 'EOF'
#!/bin/bash
# NanoFuse Debug Tools - Capabilities Initialization
# Run this script with root privileges to enable tcpdump packet capture
# for non-root users. Only use in trusted, isolated debug environments.
#
# WARNING: Enabling CAP_NET_RAW allows packet capture of sensitive traffic.
#
# Usage: init-debug-capabilities [--force]
#   --force   Skip interactive confirmation prompt

set -euo pipefail

FORCE=false
if [[ "${1:-}" == "--force" ]] || [[ "${1:-}" == "-f" ]]; then
    FORCE=true
fi

echo "[debug-tools] Initializing debug tool capabilities..."

if ! command -v setcap &>/dev/null; then
    echo "[debug-tools] Error: 'setcap' not found; install libcap2-bin"
    exit 1
fi

if [[ ! -f "/usr/bin/tcpdump" ]]; then
    echo "[debug-tools] Error: /usr/bin/tcpdump not found"
    exit 1
fi

if [[ $EUID -ne 0 ]]; then
    echo "[debug-tools] Error: This script must be run as root"
    exit 1
fi

if [[ "$FORCE" != "true" ]]; then
    # Check if stdin is a terminal for interactive prompt
    if [[ -t 0 ]]; then
        echo "[debug-tools] WARNING: Enabling CAP_NET_RAW on tcpdump allows any user to capture packets"
        read -p "Continue? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "[debug-tools] Aborted."
            exit 0
        fi
    else
        echo "[debug-tools] Error: Non-interactive mode requires --force flag"
        exit 1
    fi
fi

if setcap cap_net_raw+ep "/usr/bin/tcpdump"; then
    echo "[debug-tools] CAP_NET_RAW successfully applied to /usr/bin/tcpdump"
else
    echo "[debug-tools] Failed to set capabilities (filesystem may not support xattrs)"
    exit 1
fi
EOF
chmod 755 "${ROOTFS_PATH}/usr/local/bin/init-debug-capabilities"
echo "[debug-tools] Created /usr/local/bin/init-debug-capabilities for optional capability setup"

# Create a debug tools help script
cat > "${ROOTFS_PATH}/usr/local/bin/debug-help" << 'EOF'
#!/bin/bash
# NanoFuse Debug Tools - Quick Reference

cat << 'HELP'
NanoFuse Debug Tools Layer - Quick Reference
=============================================

TRACING:
  strace <command>         - Trace system calls
  strace -p <pid>          - Attach to running process
  ltrace <command>         - Trace library calls
  ltrace -p <pid>          - Attach to running process

DEBUGGING:
  gdb <program>            - Debug a program
  gdb -p <pid>             - Attach to running process
  gdb <program> <core>     - Analyze core dump

NETWORK:
  tcpdump -i eth0          - Capture packets on eth0
  tcpdump -w capture.pcap  - Write to file
  curl -v <url>            - Verbose HTTP request
  wget <url>               - Download file

PROCESS MONITORING:
  htop                     - Interactive process viewer
  htop -p <pid>            - Monitor specific process

COMMON EXAMPLES:
  strace -f -e trace=network curl example.com
  tcpdump -i any port 80
  gdb --batch -ex "bt" -p <pid>

HELP
EOF
chmod 755 "${ROOTFS_PATH}/usr/local/bin/debug-help"

echo "[debug-tools] Debug tools layer installed successfully"
