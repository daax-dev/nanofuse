# Recording Agent Layer - Session recording via virtio-vsock
# This Dockerfile creates a minimal layer for recording agent infrastructure.
#
# The actual record-agent binary will be built from the nanofuse Go codebase
# and injected during image composition. This layer provides:
# - systemd service configuration
# - Required directories and permissions
# - Placeholder script for testing
#
# Build: docker build -t nanofuse-layer-recording-agent:latest .
# Extract: ./scripts/build-layer.sh recording-agent
#
# Target size: ~5MB (minimal)

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse Recording Agent Layer"
LABEL org.opencontainers.image.description="Session recording agent for NanoFuse microVMs"
LABEL org.opencontainers.image.version="1.0.0"
LABEL com.nanofuse.layer.type="feature"
LABEL com.nanofuse.layer.name="recording-agent"

ENV DEBIAN_FRONTEND=noninteractive

# Install minimal dependencies
# - socat: for vsock testing/debugging
# - coreutils: standard utilities
RUN apt-get update && apt-get install -y --no-install-recommends \
    socat \
    coreutils \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create directories for recording agent
RUN mkdir -p /var/log/nanofuse \
    && mkdir -p /var/lib/nanofuse/recordings \
    && mkdir -p /etc/nanofuse

# Create placeholder record-agent script
# This will be replaced by the actual Go binary during image composition
RUN cat > /usr/local/bin/record-agent << 'SCRIPT'
#!/bin/bash
# NanoFuse Recording Agent - Placeholder
# This script is a placeholder that will be replaced by the actual
# record-agent binary built from the nanofuse Go codebase.
#
# The real agent implements:
# - Terminal session capture via pty interception
# - File I/O capture via fanotify/inotify
# - vsock communication to host receiver
# - Buffered recording with configurable size

set -euo pipefail

VSOCK_PORT="${VSOCK_PORT:-52}"
BUFFER_SIZE_MB="${BUFFER_SIZE_MB:-16}"

echo "[record-agent] Starting recording agent (placeholder)"
echo "[record-agent] vsock port: ${VSOCK_PORT}"
echo "[record-agent] buffer size: ${BUFFER_SIZE_MB}MB"

# Keep the service running
while true; do
    sleep 60
    echo "[record-agent] heartbeat - waiting for vsock connection"
done
SCRIPT

RUN chmod +x /usr/local/bin/record-agent

# Create systemd service file
RUN cat > /etc/systemd/system/record-agent.service << 'SERVICE'
[Unit]
Description=NanoFuse Recording Agent
Documentation=https://github.com/peregrinesummit/nanofuse
After=network.target
Wants=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/record-agent
Restart=always
RestartSec=5
Environment=VSOCK_PORT=52
Environment=BUFFER_SIZE_MB=16

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/var/log/nanofuse /var/lib/nanofuse

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=record-agent

[Install]
WantedBy=multi-user.target
SERVICE

# Enable the service
RUN mkdir -p /etc/systemd/system/multi-user.target.wants \
    && ln -sf /etc/systemd/system/record-agent.service \
       /etc/systemd/system/multi-user.target.wants/record-agent.service

# Create layer marker
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: recording-agent" > /etc/nanofuse/layers/recording-agent.yaml \
    && echo "version: 1.0.0" >> /etc/nanofuse/layers/recording-agent.yaml \
    && echo "type: feature" >> /etc/nanofuse/layers/recording-agent.yaml

# Clean up
RUN rm -rf /var/log/* /tmp/* \
    && rm -rf /usr/share/doc/* /usr/share/man/*

CMD ["/bin/bash"]
