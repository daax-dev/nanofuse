# Base OS Layer - Ubuntu 24.04 Minimal with systemd
# This is the foundation layer that all other NanoFuse layers depend on.
#
# Provides:
# - Ubuntu 24.04 minimal rootfs
# - systemd init system (configured for Firecracker)
# - Basic networking (iproute2, netplan)
# - Essential utilities
# - SSH server (optional, for debugging)
#
# Build: docker build -t nanofuse-layer-base-os:latest .
# Extract: ./scripts/build-layer.sh base-os
#
# Target size: ~150MB (compressed)

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse Base OS Layer"
LABEL org.opencontainers.image.description="Ubuntu 24.04 minimal with systemd for NanoFuse microVMs"
LABEL org.opencontainers.image.version="1.0.0"
LABEL com.nanofuse.layer.type="base"
LABEL com.nanofuse.layer.name="base-os"

ENV DEBIAN_FRONTEND=noninteractive

# Install essential packages for a bootable system
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Init system
    systemd \
    systemd-sysv \
    dbus \
    # Networking
    iproute2 \
    iputils-ping \
    netplan.io \
    # Essential utilities
    coreutils \
    util-linux \
    procps \
    kmod \
    # Filesystem utilities
    e2fsprogs \
    # Terminal/shell
    bash \
    less \
    # SSH for debugging (optional but useful)
    openssh-server \
    # Certificates for HTTPS
    ca-certificates \
    # Cleanup
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Configure systemd for Firecracker
# Remove unnecessary systemd units that won't work in microVM
RUN systemctl mask \
    systemd-resolved.service \
    systemd-networkd-wait-online.service \
    systemd-firstboot.service \
    systemd-machine-id-commit.service \
    plymouth-quit-wait.service \
    plymouth-quit.service \
    plymouth-start.service \
    multipathd.service \
    multipathd.socket \
    ModemManager.service \
    NetworkManager-wait-online.service \
    NetworkManager.service \
    udisks2.service \
    accounts-daemon.service \
    power-profiles-daemon.service \
    2>/dev/null || true

# Enable essential services
# Note: Use ssh.socket (not ssh.service) for Ubuntu 24.04 socket activation
RUN systemctl enable \
    systemd-networkd.service \
    ssh.socket \
    2>/dev/null || true

# Configure serial console for Firecracker (ttyS0)
RUN mkdir -p /etc/systemd/system/serial-getty@ttyS0.service.d \
    && echo '[Service]' > /etc/systemd/system/serial-getty@ttyS0.service.d/override.conf \
    && echo 'ExecStart=' >> /etc/systemd/system/serial-getty@ttyS0.service.d/override.conf \
    && echo 'ExecStart=-/sbin/agetty -o "-p -- \\u" --keep-baud 115200,38400,9600 --noclear --autologin root ttyS0 $TERM' >> /etc/systemd/system/serial-getty@ttyS0.service.d/override.conf \
    && systemctl enable serial-getty@ttyS0.service

# Create default netplan configuration for DHCP
RUN mkdir -p /etc/netplan \
    && echo 'network:' > /etc/netplan/01-netcfg.yaml \
    && echo '  version: 2' >> /etc/netplan/01-netcfg.yaml \
    && echo '  renderer: networkd' >> /etc/netplan/01-netcfg.yaml \
    && echo '  ethernets:' >> /etc/netplan/01-netcfg.yaml \
    && echo '    eth0:' >> /etc/netplan/01-netcfg.yaml \
    && echo '      dhcp4: true' >> /etc/netplan/01-netcfg.yaml \
    && chmod 600 /etc/netplan/01-netcfg.yaml

# Configure SSH (disable password auth, allow root login for debugging)
RUN mkdir -p /etc/ssh/sshd_config.d \
    && echo 'PermitRootLogin prohibit-password' > /etc/ssh/sshd_config.d/nanofuse.conf \
    && echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config.d/nanofuse.conf \
    && echo 'PubkeyAuthentication yes' >> /etc/ssh/sshd_config.d/nanofuse.conf \
    && mkdir -p /root/.ssh \
    && chmod 700 /root/.ssh

# Create standard directories
RUN mkdir -p /var/log /tmp /run \
    && chmod 1777 /tmp

# Set hostname
RUN echo 'nanofuse' > /etc/hostname

# Note: /etc/resolv.conf is managed by Docker during build
# It will be configured by systemd-networkd at runtime via DHCP
# or can be set via post-install hook

# Create fstab
RUN echo '# /etc/fstab: static file system information' > /etc/fstab \
    && echo 'proc /proc proc defaults 0 0' >> /etc/fstab \
    && echo 'sysfs /sys sysfs defaults 0 0' >> /etc/fstab \
    && echo 'devtmpfs /dev devtmpfs defaults 0 0' >> /etc/fstab

# Create layer marker
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: base-os" > /etc/nanofuse/layers/base-os.yaml \
    && echo "version: 1.0.0" >> /etc/nanofuse/layers/base-os.yaml \
    && echo "type: base" >> /etc/nanofuse/layers/base-os.yaml \
    && echo "ubuntu: 24.04" >> /etc/nanofuse/layers/base-os.yaml

# Set default target to multi-user (no GUI)
RUN systemctl set-default multi-user.target

# Clean up
RUN rm -rf /var/log/* /tmp/* \
    && rm -rf /usr/share/doc/* /usr/share/man/* \
    && find /usr/share/locale -mindepth 1 -maxdepth 1 ! -name 'en*' -exec rm -rf {} + 2>/dev/null || true

CMD ["/sbin/init"]
