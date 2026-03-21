# OpenClaw (Clawdbot) MicroVM Image Plan

*Created: 2026-02-01*

---

## Goal

Run OpenClaw/Clawdbot in a Firecracker microVM using nanofuse, following the container → microVM workflow.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Host (galway)                           │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Firecracker microVM                      │  │
│  │  ┌────────────────────────────────────────────────┐  │  │
│  │  │  Ubuntu 24.04 + systemd                        │  │  │
│  │  │  ├── Node.js 22.x                              │  │  │
│  │  │  ├── Clawdbot (global npm)                     │  │  │
│  │  │  ├── clawdbot.service (systemd)                │  │  │
│  │  │  └── SSH access                                │  │  │
│  │  └────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                          │                                  │
│                   TAP networking                            │
│                   (172.16.x.x)                              │
└─────────────────────────────────────────────────────────────┘
```

## Workflow

### Phase 1: Containerize Clawdbot

```
Dockerfile → docker build → docker image
```

### Phase 2: Export to MicroVM Image

```
docker create → docker export → tar extract → ext4 rootfs
```

### Phase 3: Run with Nanofuse

```
Build rootfs → Launch Firecracker VM → SSH/API access
```

---

## Phase 1: Dockerfile

### Requirements

| Component | Version | Notes |
|-----------|---------|-------|
| Base | Ubuntu 24.04 | Match nanofuse base |
| Node.js | 22.x LTS | Required for Clawdbot |
| Clawdbot | 2026.1.24-3 | `npm install -g clawdbot` (pinned, override via build arg) |
| systemd | included | Init system for VM |
| SSH | openssh-server | Remote access |

### Dockerfile

```dockerfile
# OpenClaw MicroVM Image
# Simplified example - see actual Dockerfile for complete implementation

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="OpenClaw MicroVM"
LABEL org.opencontainers.image.description="Clawdbot running in Firecracker microVM"

ENV DEBIAN_FRONTEND=noninteractive

# Install systemd and base packages (match nanofuse-base)
RUN apt-get update && apt-get install -y --no-install-recommends \
    systemd \
    systemd-sysv \
    openssh-server \
    curl \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 22.x (via signed apt repo)
RUN curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key \
      | gpg --dearmor -o /usr/share/keyrings/nodesource.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/nodesource.gpg] https://deb.nodesource.com/node_22.x nodistro main" \
      > /etc/apt/sources.list.d/nodesource.list \
    && apt-get update && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Clawdbot globally
ARG CLAWDBOT_VERSION=2026.1.24-3
RUN npm install -g "clawdbot@${CLAWDBOT_VERSION}"

# Create clawdbot user (non-root operation)
RUN useradd -m -s /bin/bash clawdbot \
    && mkdir -p /home/clawdbot/.clawdbot \
    && chown -R clawdbot:clawdbot /home/clawdbot

# Copy systemd service
COPY clawdbot.service /etc/systemd/system/clawdbot.service

# Enable services for VM boot
RUN systemctl enable ssh \
    && systemctl enable systemd-networkd \
    && systemctl enable clawdbot

# Configure SSH (key-only auth, matching nanofuse-base security posture)
# Accounts are locked by default; use SSH keys for access
RUN mkdir -p /run/sshd \
    && sed -i 's/#PermitRootLogin.*/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config \
    && sed -i 's/#PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config \
    && passwd -l root \
    && passwd -l clawdbot \
    && rm -f /etc/ssh/ssh_host_*

# Configure serial console for Firecracker (ttyS0)
RUN systemctl enable serial-getty@ttyS0.service

# Cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

CMD ["/sbin/init"]
```

### systemd Service

```ini
# /etc/systemd/system/clawdbot.service
[Unit]
Description=Clawdbot Gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=clawdbot
WorkingDirectory=/home/clawdbot
ExecStart=/usr/bin/env clawdbot gateway
Restart=always
RestartSec=10
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
```

---

## Phase 2: Build Process

### Directory Structure

```
nanofuse/images/openclaw/
├── Dockerfile
├── clawdbot.service
├── build.sh           # Adapted from base/build.sh
├── manifest.json
└── PLAN.md           # This file
```

### Build Script (simplified)

```bash
#!/bin/bash
set -euo pipefail

IMAGE_NAME="openclaw"
BUILD_DIR="./build"
ROOTFS_SIZE="2048"  # 2GB for Node.js + npm packages

# 1. Build Docker image
docker build -t "${IMAGE_NAME}:latest" .

# 2. Export filesystem
mkdir -p "${BUILD_DIR}/rootfs"
CONTAINER_ID=$(docker create "${IMAGE_NAME}:latest")
docker export "${CONTAINER_ID}" | tar -C "${BUILD_DIR}/rootfs" -xf -
docker rm "${CONTAINER_ID}"

# 3. Create ext4 image
dd if=/dev/zero of="${BUILD_DIR}/rootfs.ext4" bs=1M count="${ROOTFS_SIZE}"
mkfs.ext4 -F -L openclaw-root "${BUILD_DIR}/rootfs.ext4"

# 4. Mount and copy
mkdir -p "${BUILD_DIR}/mnt"
sudo mount "${BUILD_DIR}/rootfs.ext4" "${BUILD_DIR}/mnt"
sudo cp -a "${BUILD_DIR}/rootfs/." "${BUILD_DIR}/mnt/"
sudo umount "${BUILD_DIR}/mnt"

# 5. Use shared kernel from base image
# (or download Firecracker kernel)

echo "✓ Build complete: ${BUILD_DIR}/rootfs.ext4"
```

---

## Phase 3: Run with Nanofuse

### Register Image

```bash
# Use the rootfs directly with Firecracker
firecracker --config-file vm-config.json
# Set root_drive.path_on_host to ./build/rootfs.ext4

# Or push to a container registry for distribution
docker push ghcr.io/<org>/openclaw:latest
```

### Launch VM

```bash
# Launch via Firecracker config (see nanofuse docs for helper commands)
# The rootfs and kernel paths are specified in the VM config JSON

# SSH into VM after launch
ssh clawdbot@<vm-ip>
```

### Configure Clawdbot Inside VM

```bash
# SSH in
ssh clawdbot@<vm-ip>

# Configure (first time)
clawdbot init
clawdbot configure --provider anthropic

# Check status
clawdbot status
```

---

## Configuration Injection

> **Note**: The `nanofuse vm run` commands below represent the **planned CLI interface**
> and are not yet implemented. Currently, use Firecracker config JSON directly.

### Option A: Volume Mount (config at runtime)

```bash
# Mount host config directory
nanofuse vm run openclaw my-clawdbot \
  --mount /home/jpoley/.clawdbot:/home/clawdbot/.clawdbot
```

### Option B: Baked Config (config in image)

```dockerfile
# Add to Dockerfile
COPY config.yaml /home/clawdbot/.clawdbot/config.yaml
COPY auth-profiles.json /home/clawdbot/.clawdbot/auth-profiles.json
```

### Option C: Environment Variables

```bash
# Pass via env
nanofuse vm run openclaw my-clawdbot \
  --env ANTHROPIC_API_KEY=sk-xxx \
  --env TELEGRAM_BOT_TOKEN=xxx
```

---

## Networking Considerations

### Port Forwarding

Clawdbot gateway typically runs on port 3000. Forward it:

```bash
nanofuse vm run openclaw my-clawdbot \
  --port-forward 3000:3000
```

### Webhook Access

For Telegram webhooks, the VM needs external access:

```bash
# NAT mode with port forward
nanofuse vm run openclaw my-clawdbot \
  --network-mode nat \
  --port-forward 443:3000
```

---

## Open Questions

1. **Config persistence**: Mount volume vs bake into image?
2. **Auth tokens**: How to inject secrets securely?
3. **Multi-tenant**: One VM per user or shared?
4. **Snapshot/Resume**: Leverage nanofuse snapshots for fast startup?
5. **Logging**: Where do logs go? Journald → host?

---

## Task Breakdown

### Immediate (Today)

- [ ] Create `nanofuse/images/openclaw/` directory
- [ ] Write Dockerfile
- [ ] Write clawdbot.service
- [ ] Adapt build.sh from base

### Next

- [ ] Test Docker build locally
- [ ] Test docker export → ext4 workflow
- [ ] Register with nanofuse
- [ ] Boot VM and verify clawdbot starts

### Future

- [ ] Config injection strategy
- [ ] GHCR CI/CD pipeline
- [ ] Snapshot optimization
- [ ] Documentation
