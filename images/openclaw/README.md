# openclaw

Clawdbot AI agent gateway running in a Firecracker microVM.

## Overview

This image builds a minimal Ubuntu 24.04-based rootfs containing:

- **Clawdbot** gateway service (Node.js 22.x LTS, pinned version)
- **systemd** for service management
- **SSH** for remote access (key-only auth)
- **SSH key injection** via kernel cmdline (`sshkey=<base64>`)

## User Accounts

Both `root` and `clawdbot` accounts are **locked by default** with SSH key-only authentication (password auth is disabled).

### Configuring SSH Access

**Via kernel cmdline (Firecracker boot_args) — recommended:**

The `install-ssh-key.service` reads a base64-encoded public key from the kernel cmdline at boot and installs it for both `root` and `clawdbot` users:

```bash
# Encode your SSH public key
SSHKEY=$(base64 -w0 < ~/.ssh/id_ed25519.pub)

# Pass via Firecracker boot_args
"boot_args": "console=ttyS0 sshkey=${SSHKEY}"
```

**Via authorized_keys in Dockerfile (build time):**
```bash
# Add a COPY instruction to the Dockerfile before the final CMD:
# COPY my_authorized_keys /home/clawdbot/.ssh/authorized_keys
# RUN chown clawdbot:clawdbot /home/clawdbot/.ssh/authorized_keys \
#     && chmod 600 /home/clawdbot/.ssh/authorized_keys
```

## Building

```bash
sudo ./build.sh
```

This produces an ext4 rootfs image in `./build/rootfs.ext4` suitable for Firecracker.

**With different clawdbot version:**
```bash
# Override the default pinned version via env var:
sudo CLAWDBOT_VERSION=2026.2.1 ./build.sh
```

## Services

| Service | Port | User | Description |
|---------|------|------|-------------|
| clawdbot | 3000 | clawdbot | AI agent gateway |
| sshd | 22 | any | SSH access (key-only) |
| install-ssh-key | - | root | Injects SSH key from kernel cmdline at boot (root + clawdbot) |
| generate-ssh-keys | - | root | Generates SSH host keys on first boot |
| serial-getty@ttyS0 | - | - | Serial console for Firecracker |

## Resources

Recommended VM configuration:

- **Memory:** 1024 MB
- **vCPUs:** 2
- **Rootfs:** 2048 MB
