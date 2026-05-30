# Nanofuse Quick Start Guide

Get Nanofuse running and launch your first microVM in 5 minutes.

## Prerequisites

Before you begin, ensure you have:

- Linux host with KVM support
- x86_64 architecture
- Root access (required for networking)

Verify KVM is available:

```bash
ls -l /dev/kvm
# Should show: crw-rw----+ 1 root kvm ...
```

## Step 1: Install Nanofuse

### Option A: Download Pre-built Binaries

```bash
# Download latest release
VERSION=v0.1.0
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofuse
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofused

# Make executable and install
chmod +x nanofuse nanofused
sudo mv nanofuse nanofused /usr/local/bin/

# Verify installation
nanofuse --version
```

### Option B: Build from Source

```bash
# Clone repository
git clone https://github.com/daax-dev/nanofuse.git
cd nanofuse

# Install build tool (Mage)
./scripts/ensure-mage.sh

# Build binaries
mage all

# Install to ~/bin (recommended)
mage installUser

# Or install system-wide
sudo mage install
```

## Step 2: Start the Daemon

The `nanofused` daemon manages microVM lifecycle and networking.

### Manual Start (for testing)

```bash
sudo nanofused
```

### Production Start (systemd)

```bash
# Install systemd service
sudo cp systemd/nanofused.service /etc/systemd/system/
sudo systemctl daemon-reload

# Start and enable on boot
sudo systemctl enable --now nanofused

# Check status
sudo systemctl status nanofused
```

## Step 3: Authenticate with GHCR

Nanofuse images are hosted on GitHub Container Registry and require authentication.

```bash
# Create a GitHub token with read:packages scope
# https://github.com/settings/tokens/new?scopes=read:packages

# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin
```

## Step 4: Pull a Base Image

```bash
# Pull the default base image
nanofuse image pull --default

# Or pull a specific version
nanofuse image pull --default --tag v1.0.0

# Verify image was pulled
nanofuse image list
```

## Step 5: Create and Start a microVM

```bash
# Create a VM named "my-vm" from the default image
nanofuse vm run default my-vm

# Check VM status
nanofuse vm status my-vm

# View console output
nanofuse vm logs my-vm
```

## Step 6: Access Your microVM

### Via SSH

```bash
# Get VM IP address
nanofuse vm inspect my-vm | grep ip

# SSH into the VM (if SSH keys are configured)
ssh root@<VM_IP>
```

### Via Console Logs

```bash
# Follow console output in real-time
nanofuse vm logs my-vm --follow
```

## Step 7: Stop and Clean Up

```bash
# Stop the VM gracefully
nanofuse vm stop my-vm

# Delete the VM
nanofuse vm delete my-vm

# Verify deletion
nanofuse vm list
```

## Building Custom Images with Layers

NanoFuse uses a layer-based architecture for building microVM images. You can compose custom images from reusable layers.

### Available Layers

| Layer | Type | Size | Description |
|-------|------|------|-------------|
| `base-os` | base | ~100MB | Ubuntu 24.04 with systemd, SSH |
| `python-runtime` | runtime | ~159MB | Python 3.12 with pip, venv |
| `node-runtime` | runtime | ~466MB | Node.js 22 LTS with npm, pnpm, bun |
| `go-runtime` | runtime | ~87MB | Runtime for compiled Go binaries |
| `recording-agent` | feature | ~81MB | Session recording via vsock |

### Build a Layer

```bash
# Build a specific layer
./scripts/build-layer.sh python-runtime

# Build all layers
./scripts/build-layer.sh all

# List available layers
./scripts/build-layer.sh --list
```

### Create an Image Manifest

Create `my-image.manifest.yaml`:

```yaml
version: "1.0"
name: "my-custom-image"
description: "Custom microVM image with Python and recording"

kernel:
  version: "6.1.102"
  source: "local://test/fixtures/kernel/vmlinux"
  cmdline: "console=ttyS0 root=/dev/vda rw init=/sbin/init"

layers:
  - name: "base-os"
    type: "base"
    source: "local://layers/base-os"
    required: true

  - name: "python-runtime"
    type: "runtime"
    source: "local://layers/python-runtime"
    required: true
    dependencies:
      - "base-os"

  - name: "recording-agent"
    type: "feature"
    source: "local://layers/recording-agent"
    condition: "${INCLUDE_RECORDING:-true}"
    dependencies:
      - "base-os"

output:
  path: "./build/my-custom-image"
  format: "ext4"
  size_mb: 2048
```

### Build the Image

```bash
# Build from manifest
nanofuse build -m my-image.manifest.yaml

# Build with recording disabled
INCLUDE_RECORDING=false nanofuse build -m my-image.manifest.yaml
```

### Run Custom Image

```bash
# Register the built image
nanofuse image register ./build/my-custom-image/rootfs.ext4 my-custom-image

# Run a VM from your custom image
nanofuse vm run my-custom-image test-vm
```

### Layer Composition Example

Here's a complete example for an AI agent workstation:

```yaml
version: "1.0"
name: "ai-agent-workstation"
description: "Full-featured AI agent execution environment"

kernel:
  version: "6.1.102"
  source: "local://test/fixtures/kernel/vmlinux"
  cmdline: "console=ttyS0 root=/dev/vda rw init=/sbin/init"

layers:
  # Core OS
  - name: "base-os"
    type: "base"
    source: "local://layers/base-os"
    required: true
    config:
      timezone: "UTC"

  # Python for scripts and tools
  - name: "python-runtime"
    type: "runtime"
    source: "local://layers/python-runtime"
    required: true
    dependencies: ["base-os"]

  # Node.js for Claude Code CLI
  - name: "node-runtime"
    type: "runtime"
    source: "local://layers/node-runtime"
    required: true
    dependencies: ["base-os"]

  # Session recording for audit
  - name: "recording-agent"
    type: "feature"
    source: "local://layers/recording-agent"
    condition: "${INCLUDE_RECORDING:-true}"
    dependencies: ["base-os"]
    config:
      vsock_port: 52
      capture_modes:
        - "terminal"
        - "file_io"

output:
  path: "./build/ai-agent-workstation"
  format: "ext4"
  size_mb: 4096
```

For more details on layer authoring, see the [Layer Authoring Guide](LAYER_AUTHORING.md).

## Enabling Session Recording

NanoFuse can capture terminal sessions and file operations for auditing and debugging.

### Quick Recording Setup

1. Include the recording-agent layer in your image (see above)
2. Start the daemon with recording enabled:
   ```bash
   sudo nanofused --recording-enabled
   ```
3. Run a VM - recording starts automatically
4. View recordings via API:
   ```bash
   curl http://localhost:8080/api/v1/recordings
   ```

For full recording documentation, see the [Recording Guide](RECORDING.md).

## Common Commands Reference

| Command | Description |
|---------|-------------|
| `nanofuse vm list` | List all VMs |
| `nanofuse vm run <image> <name>` | Create and start a VM |
| `nanofuse vm status <name>` | Check VM status |
| `nanofuse vm logs <name>` | View console output |
| `nanofuse vm stop <name>` | Stop a running VM |
| `nanofuse vm delete <name>` | Delete a VM |
| `nanofuse image list` | List pulled images |
| `nanofuse image pull <ref>` | Pull an image |

## Image Shortcuts

Nanofuse supports convenient image shortcuts:

| Shortcut | Full Reference |
|----------|----------------|
| `default` | `ghcr.io/daax-dev/nanofuse/base:latest` |
| `default:v1.0.0` | `ghcr.io/daax-dev/nanofuse/base:v1.0.0` |
| `base` | `ghcr.io/daax-dev/nanofuse/base:latest` |

## Troubleshooting

### Daemon fails to start

```bash
# Check daemon logs
sudo journalctl -u nanofused -n 50

# Common causes:
# - /dev/kvm not accessible
# - Port already in use
# - Missing Firecracker binary
```

### VM fails to boot

```bash
# Check console output for errors
nanofuse vm logs my-vm

# Common causes:
# - Image not pulled
# - Network bridge not created
# - Insufficient permissions
```

### Network connectivity issues

```bash
# Verify bridge interface exists
ip addr show nanofuse0

# Check NAT rules
sudo iptables -t nat -L -n | grep 172.16
```

For more troubleshooting help, see [Troubleshooting Guide](TROUBLESHOOTING.md).

## Next Steps

- [Layer Authoring Guide](LAYER_AUTHORING.md) - Create custom layers
- [Recording Guide](RECORDING.md) - Set up session recording
- [API Quick Start](API_QUICK_START.md) - Use the REST API directly
- [Developer Guide](DEVELOPER_GUIDE.md) - Build and debug Nanofuse
- [SSH Access Guide](SSH_ACCESS_QUICK_START.md) - Configure SSH access
- [Project Goals](GOALS.md) - Understand the architecture and roadmap

---

*See also: [README](../README.md) | [FAQ](FAQ.md) | [Troubleshooting](TROUBLESHOOTING.md)*
