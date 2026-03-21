# NanoFuse Base MicroVM Image
# Build FROM ubuntu:24.04 with systemd, following Slicer's Dockerfile patterns
# Bundles proven Slicer 5.10.240 kernel for Firecracker compatibility
#
# Critical Build Constraints:
# - NEVER use 'systemctl start' during Docker build (services not running in container)
# - ALWAYS use 'systemctl enable' to configure services for VM boot
# - Set DEBIAN_FRONTEND=noninteractive for all apt operations
# - Console MUST be on ttyS0 for Firecracker compatibility
# - Keep CMD unchanged (expects systemd)

FROM ubuntu:24.04

# Build arguments for OCI labels (passed from CI)
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF
ARG VCS_URL=https://github.com/jpoley/nanofuse

# OCI Image Labels (https://github.com/opencontainers/image-spec/blob/main/annotations.md)
LABEL org.opencontainers.image.title="NanoFuse Base Image"
LABEL org.opencontainers.image.description="Ubuntu 24.04 systemd-based microVM image for Firecracker with fast snapshot/resume support"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.authors="NanoFuse Contributors"
LABEL org.opencontainers.image.vendor="NanoFuse"
LABEL org.opencontainers.image.url="${VCS_URL}"
LABEL org.opencontainers.image.source="${VCS_URL}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.documentation="${VCS_URL}/blob/main/README.md"

# NanoFuse-specific labels
LABEL com.nanofuse.base-os="ubuntu:24.04"
LABEL com.nanofuse.kernel.version="5.10.204"
LABEL com.nanofuse.kernel.source="Firecracker CI"
LABEL com.nanofuse.architecture="x86_64"
LABEL com.nanofuse.type="base"
LABEL com.nanofuse.firecracker.compatible="true"
LABEL com.nanofuse.services.ssh="enabled"
LABEL com.nanofuse.services.systemd-networkd="enabled"

# Prevent interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive

# Install systemd and essential packages (MINIMAL set only)
# Only essentials to keep image small (~117 MB instead of 182 MB):
# - systemd: init system for VM (includes systemd-networkd in Ubuntu 24.04)
# - systemd-sysv: provides /sbin/init symlink
# - openssh-server: SSH access into VM
# - ca-certificates: TLS/HTTPS support
# - curl: download tool
# - dbus: required by systemd
# - iproute2: network configuration (ip command)
# - python3: required for http-test-server.service (port forwarding tests)
#
# NOT included (add only if needed):
# - wget: curl is sufficient
# - iputils-ping: basic ping only
# - kmod: rarely needed in Firecracker VMs
# - udev: systemd handles devices
# - vim-tiny, less: vi/ed available, can install later
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        systemd \
        systemd-sysv \
        openssh-server \
        ca-certificates \
        curl \
        dbus \
        iproute2 \
        python3 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Remove unnecessary systemd units that cause issues in containers
# Keep only essential services for VM boot
# Note: We do NOT mask serial-getty@.service because we need serial-getty@ttyS0
RUN systemctl mask \
    systemd-resolved.service \
    systemd-timesyncd.service \
    systemd-logind.service \
    getty@.service

# Configure serial console on ttyS0 (required for Firecracker)
# Enable getty on ttyS0 for console access
RUN systemctl enable serial-getty@ttyS0.service

# Enable SSH server (will start on VM boot)
RUN systemctl enable ssh.service

# Enable systemd-networkd for DHCP-based networking
RUN systemctl enable systemd-networkd.service

# Disable systemd-networkd-wait-online.service to prevent boot delays
# This service waits for network to be "online" which can timeout in VMs
# Our services will start when network is available, not waiting for "online" state
RUN systemctl disable systemd-networkd-wait-online.service

# Remove pre-generated SSH host keys
# They must be regenerated on first VM boot to ensure each VM has unique keys
# (See docs/SSH_HOST_KEYS_ISSUE.md for security implications)
RUN rm -f /etc/ssh/ssh_host_*_key /etc/ssh/ssh_host_*_key.pub

# Create network configuration for DHCP on primary interface
RUN mkdir -p /etc/systemd/network && \
    echo '[Match]' > /etc/systemd/network/20-wired.network && \
    echo 'Name=en*' >> /etc/systemd/network/20-wired.network && \
    echo '' >> /etc/systemd/network/20-wired.network && \
    echo '[Network]' >> /etc/systemd/network/20-wired.network && \
    echo 'DHCP=yes' >> /etc/systemd/network/20-wired.network && \
    echo 'LinkLocalAddressing=yes' >> /etc/systemd/network/20-wired.network

# Configure SSH for security and access
# - Permit root login with key only (no password)
# - Create authorized_keys directory
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Copy and enable SSH key injection script
# This installs SSH public keys passed via kernel cmdline (base64-encoded sshkey= parameter)
COPY scripts/install-ssh-key.sh /usr/local/bin/install-ssh-key.sh
RUN chmod +x /usr/local/bin/install-ssh-key.sh
COPY units/install-ssh-key.service /etc/systemd/system/install-ssh-key.service
RUN systemctl enable install-ssh-key.service

# Copy and enable first-boot systemd unit
# This runs once on first boot for initialization tasks
COPY units/firstboot.service /etc/systemd/system/firstboot.service
RUN systemctl enable firstboot.service

# Make SSH wait for firstboot to generate host keys
RUN mkdir -p /etc/systemd/system/ssh.service.d
COPY units/ssh-wait-firstboot.conf /etc/systemd/system/ssh.service.d/wait-firstboot.conf

# Copy and enable HTTP test server
# This provides a simple HTTP endpoint for testing port forwarding
COPY units/http-test-server.service /etc/systemd/system/http-test-server.service
RUN systemctl enable http-test-server.service

# Create directories for VM runtime data
RUN mkdir -p /var/log/nanofuse && \
    mkdir -p /run/nanofuse

# DNS configuration will be handled by firstboot service
# (Can't modify /etc/resolv.conf during Docker build - it's bind-mounted)

# Set default boot target to multi-user (no graphical target)
RUN systemctl set-default multi-user.target

# NOTE: Kernel is NOT included in Docker image
# The kernel is stored separately and passed to Firecracker directly
# Docker image only contains the rootfs (Ubuntu filesystem)
# The kernel binary will be at ./build/vmlinux (passed separately to VMs)

# Create manifest.json with image metadata
# This will be customized by build script with actual values
RUN echo '{}' > /manifest.json

# systemd expects to run as PID 1
# In the actual VM, Firecracker will boot the kernel which will run systemd
# This CMD is for container compatibility but won't be used in Firecracker VM
CMD ["/sbin/init"]
