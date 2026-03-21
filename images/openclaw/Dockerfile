# OpenClaw MicroVM Image
# Clawdbot running in Firecracker microVM via nanofuse
#
# Build: docker build -t openclaw:latest .
# Export: docker export $(docker create openclaw:latest) | tar -xf -

FROM ubuntu:24.04

ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF
ARG LICENSES=Apache-2.0
ARG VENDOR=peregrinesummit

LABEL org.opencontainers.image.title="OpenClaw MicroVM"
LABEL org.opencontainers.image.description="Clawdbot AI agent gateway in Firecracker microVM"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.source="https://github.com/peregrinesummit/daax"
LABEL org.opencontainers.image.url="https://github.com/peregrinesummit/daax"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.licenses="${LICENSES}"
LABEL org.opencontainers.image.vendor="${VENDOR}"
LABEL com.nanofuse.type="openclaw"
LABEL com.nanofuse.firecracker.compatible="true"

ENV DEBIAN_FRONTEND=noninteractive

# Install systemd and base packages (match nanofuse-base pattern)
RUN apt-get update && apt-get install -y --no-install-recommends \
    systemd \
    systemd-sysv \
    openssh-server \
    curl \
    ca-certificates \
    gnupg \
    git \
    jq \
    dbus \
    iproute2 \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 22.x LTS (required for Clawdbot)
# Use NodeSource signed apt repository instead of running remote scripts
RUN curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key \
      | gpg --dearmor -o /usr/share/keyrings/nodesource.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/nodesource.gpg] https://deb.nodesource.com/node_22.x nodistro main" \
      > /etc/apt/sources.list.d/nodesource.list \
    && apt-get update \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && node --version \
    && npm --version

# Install Clawdbot globally (pinned to vetted version for supply chain safety)
# Override at build time: docker build --build-arg CLAWDBOT_VERSION=2026.1.24-3 .
ARG CLAWDBOT_VERSION=2026.1.24-3
RUN npm install -g "clawdbot@${CLAWDBOT_VERSION}" \
    && clawdbot --version

# Create clawdbot user (non-root operation)
RUN useradd -m -s /bin/bash clawdbot \
    && mkdir -p /home/clawdbot/.clawdbot/agents/main/agent \
    && chown -R clawdbot:clawdbot /home/clawdbot

# Copy systemd services
COPY clawdbot.service /etc/systemd/system/clawdbot.service
COPY install-ssh-key.service /etc/systemd/system/install-ssh-key.service
COPY install-ssh-key.sh /usr/local/bin/install-ssh-key.sh
COPY generate-ssh-keys.service /etc/systemd/system/generate-ssh-keys.service
COPY generate-ssh-keys.sh /usr/local/bin/generate-ssh-keys.sh
RUN chmod +x /usr/local/bin/install-ssh-key.sh \
    && chmod +x /usr/local/bin/generate-ssh-keys.sh

# Enable services for VM boot (NOT start - they run at boot)
RUN systemctl enable ssh.service \
    && systemctl enable systemd-networkd.service \
    && systemctl enable clawdbot.service \
    && systemctl enable install-ssh-key.service \
    && systemctl enable generate-ssh-keys.service \
    && systemctl disable systemd-networkd-wait-online.service

# Configure SSH for VM access (key-only auth, matching nanofuse-base)
# Accounts are locked by default. Inject SSH keys at launch time via:
#   - Kernel cmdline: sshkey=<base64-encoded-pubkey> (read by install-ssh-key.service)
RUN mkdir -p /run/sshd \
    && sed -i 's/#PermitRootLogin.*/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config \
    && sed -i 's/#PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config \
    && passwd -l root \
    && passwd -l clawdbot

# Configure serial console for Firecracker (ttyS0)
RUN systemctl enable serial-getty@ttyS0.service

# Network configuration for systemd-networkd
RUN mkdir -p /etc/systemd/network \
    && printf '[Match]\nName=en* eth*\n\n[Network]\nDHCP=yes\n' > /etc/systemd/network/80-dhcp.network

# Remove pre-generated SSH host keys so each VM gets unique keys
# The openssh-server package creates /etc/ssh/ with correct permissions;
# generate-ssh-keys.service uses ssh-keygen -A to regenerate keys on first boot
RUN rm -f /etc/ssh/ssh_host_*

# Cleanup
RUN apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && rm -rf /var/cache/apt/archives/*

# Systemd init
CMD ["/sbin/init"]
