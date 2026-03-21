# Networking Tools Layer Dockerfile
# Provides iptables, nftables, and advanced iproute2 tools
#
# Build: docker build -t nanofuse-layer-networking-tools .
# Export: docker export $(docker create nanofuse-layer-networking-tools) | tar -C rootfs -xf -

FROM ubuntu:24.04

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install networking tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    iptables \
    nftables \
    iproute2 \
    iputils-ping \
    conntrack \
    ebtables \
    bridge-utils \
    tcpdump \
    netcat-openbsd \
    dnsutils \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create sysctl configuration for gateway
RUN mkdir -p /etc/sysctl.d && \
    echo '# Gateway networking settings' > /etc/sysctl.d/99-gateway.conf && \
    echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.d/99-gateway.conf && \
    echo 'net.ipv4.conf.all.forwarding = 1' >> /etc/sysctl.d/99-gateway.conf && \
    echo 'net.ipv6.conf.all.forwarding = 1' >> /etc/sysctl.d/99-gateway.conf && \
    echo 'net.ipv4.conf.all.proxy_arp = 0' >> /etc/sysctl.d/99-gateway.conf && \
    echo 'net.ipv4.conf.all.send_redirects = 0' >> /etc/sysctl.d/99-gateway.conf && \
    echo 'net.ipv4.conf.default.send_redirects = 0' >> /etc/sysctl.d/99-gateway.conf

# Create default nftables configuration
RUN mkdir -p /etc/nftables && \
    echo '#!/usr/sbin/nft -f' > /etc/nftables.conf && \
    echo '' >> /etc/nftables.conf && \
    echo '# NanoFuse Gateway nftables configuration' >> /etc/nftables.conf && \
    echo 'flush ruleset' >> /etc/nftables.conf && \
    echo '' >> /etc/nftables.conf && \
    echo 'table inet filter {' >> /etc/nftables.conf && \
    echo '    chain input {' >> /etc/nftables.conf && \
    echo '        type filter hook input priority 0; policy accept;' >> /etc/nftables.conf && \
    echo '        # Accept established connections' >> /etc/nftables.conf && \
    echo '        ct state established,related accept' >> /etc/nftables.conf && \
    echo '        # Accept loopback' >> /etc/nftables.conf && \
    echo '        iif lo accept' >> /etc/nftables.conf && \
    echo '    }' >> /etc/nftables.conf && \
    echo '' >> /etc/nftables.conf && \
    echo '    chain forward {' >> /etc/nftables.conf && \
    echo '        type filter hook forward priority 0; policy accept;' >> /etc/nftables.conf && \
    echo '    }' >> /etc/nftables.conf && \
    echo '' >> /etc/nftables.conf && \
    echo '    chain output {' >> /etc/nftables.conf && \
    echo '        type filter hook output priority 0; policy accept;' >> /etc/nftables.conf && \
    echo '    }' >> /etc/nftables.conf && \
    echo '}' >> /etc/nftables.conf && \
    chmod 644 /etc/nftables.conf
