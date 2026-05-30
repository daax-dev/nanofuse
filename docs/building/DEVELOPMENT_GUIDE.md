# NanoFuse Development Guide

Complete guide to building NanoFuse, debugging with VS Code, and working with Firecracker microVM images.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Building NanoFuse](#building-nanofuse)
3. [Firecracker Image Build](#firecracker-image-build)
4. [Debugging with VS Code](#debugging-with-vs-code)
5. [Running and Testing VMs](#running-and-testing-vms)
6. [SSH Access](#ssh-access)
7. [Testing Inside Running Containers](#testing-inside-running-containers)
8. [Building Custom Layers](#building-custom-layers)
9. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Prerequisites

- **Linux host** with KVM support (`/dev/kvm`)
- **Go 1.24+** installed
- **Docker** for building images
- **Firecracker** binary (optional, needed for VM testing)
- **VS Code** with Go extension (optional, for debugging)

### Get Started in 5 Minutes

```bash
# Clone repository
git clone https://github.com/daax-dev/nanofuse.git
cd nanofuse

# Install build tool (mage)
./scripts/ensure-mage.sh

# Build all binaries (CLI and daemon)
mage all

# Binaries are now in ./bin/
ls -lh bin/
```

---

## Building NanoFuse

NanoFuse consists of two main components:
- **nanofuse** (CLI) - Command-line tool for managing VMs
- **nanofused** (daemon) - Background service that manages VM lifecycle

### Full Build

```bash
# Build all components
mage all

# Output:
# - ./bin/nanofuse     (CLI, ~10MB)
# - ./bin/nanofused    (daemon, ~15MB)
# - ./bin/register-local-image (utility, ~12MB)
```

### Build Individual Components

```bash
# Build only CLI
mage CLI

# Build only daemon
mage Daemon

# Build specific tool
mage RegisterLocalImage
```

### Build with Custom Flags

```bash
# Build with custom version/metadata
go build -ldflags="-X main.Version=v0.2.0 -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o ./bin/nanofuse ./cmd/nanofuse

# Build with race detector (catches concurrent bugs)
CGO_ENABLED=1 go build -race -o ./bin/nanofused ./cmd/nanofused
```

### Testing the Build

```bash
# Run unit tests
mage test

# Run with verbose output
mage testVerbose

# Run integration tests (requires running daemon)
mage testIntegration

# Generate coverage report
mage testCoverage
```

---

## Firecracker Image Build

The base microVM image is a critical component. It's a Ubuntu 24.04 image with systemd, SSH, and networking pre-configured.

### What Gets Built

```
images/base/build/
├── rootfs.ext4          # 2GB ext4 filesystem image
├── vmlinux              # Linux 5.10.204 kernel (Firecracker CI)
└── manifest.json        # Image metadata
```

### Building the Base Image

```bash
cd images/base

# Simple build (handles Docker + ext4 filesystem + kernel)
sudo make build

# Build with custom size
sudo make build ROOTFS_SIZE=4G

# Build with custom registry
make build REGISTRY=myregistry.io/myorg

# All-in-one: build, validate, and test
sudo make all
```

### Build Process Breakdown

The build.sh script performs 6 sequential steps:

1. **Docker Image Build** - Creates Docker image from Dockerfile
   - Installs systemd, OpenSSH, networking packages
   - Configures serial console on ttyS0
   - Enables required systemd services

2. **Container Export** - Extracts filesystem from Docker container
   - Creates rootfs/ directory with full filesystem
   - Captures all installed packages and configurations

3. **ext4 Creation** - Creates a 2GB ext4 filesystem
   - Uses `dd` + `mkfs.ext4`
   - Suitable for Firecracker block device

4. **Filesystem Copy** - Copies Docker filesystem to ext4 image
   - Mounts ext4 image
   - Copies all files from Docker container
   - Unmounts and cleans up

5. **Kernel Download** - Downloads proven Firecracker kernel
   - Source: Firecracker CI (v5.10.204)
   - S3: `spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin`
   - ~4.5MB compressed, ~10MB uncompressed

6. **Manifest Generation** - Creates metadata file
   - Records kernel version, sizes, services, build date
   - Used by nanofuse for VM configuration

### Understanding the Build Script

Key sections in `build.sh`:

```bash
# Step 1: Docker build
docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

# Step 2: Export filesystem
docker create --name nanofuse-build-temp "${IMAGE_NAME}:${IMAGE_TAG}"
docker export "${CONTAINER_ID}" | tar -C "${BUILD_DIR}/rootfs" -xf -

# Step 3: Create ext4
dd if=/dev/zero of="${BUILD_DIR}/rootfs.ext4" bs=1M count="${ROOTFS_SIZE}"
mkfs.ext4 -F -q -L nanofuse-root "${BUILD_DIR}/rootfs.ext4"

# Step 4: Mount and copy
mount -o loop "${BUILD_DIR}/rootfs.ext4" "${BUILD_DIR}/mnt"
cp -a "${BUILD_DIR}/rootfs/"* "${BUILD_DIR}/mnt/"

# Step 5: Download kernel
curl -fsSL -o "${BUILD_DIR}/vmlinux" https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin

# Step 6: Generate manifest
cat > "${BUILD_DIR}/manifest.json" << EOF
{ "version": "0.1.0", "name": "${IMAGE_NAME}", ... }
EOF
```

### Image Dockerfile Components

The `Dockerfile` is carefully designed to work with Firecracker:

```dockerfile
# Start with Ubuntu 24.04
FROM ubuntu:24.04

# Install core packages
RUN apt-get install -y systemd openssh-server ca-certificates curl ...

# Critical: Remove services that break in containers but keep serial-getty
RUN systemctl mask systemd-resolved systemd-timesyncd systemd-logind

# Configure serial console (required for Firecracker)
RUN systemctl enable serial-getty@ttyS0.service

# Enable SSH and networking
RUN systemctl enable ssh.service systemd-networkd.service

# Configure DHCP
RUN mkdir -p /etc/systemd/network && \
    echo '[Match]' > /etc/systemd/network/20-wired.network && \
    echo 'Name=en*' >> /etc/systemd/network/20-wired.network && \
    echo '[Network]' >> /etc/systemd/network/20-wired.network && \
    echo 'DHCP=yes' >> /etc/systemd/network/20-wired.network

# SSH configuration
RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config
```

**Critical Build Constraints:**

❌ **NEVER** use `systemctl start` during Docker build (services won't run in container context)

✅ **ALWAYS** use `systemctl enable` (services will start when VM boots)

✅ **ALWAYS** set `DEBIAN_FRONTEND=noninteractive` for apt operations

### Validating the Build

```bash
# Verify artifacts
make validate

# Or manually check:
file images/base/build/rootfs.ext4    # Should show: ext4 filesystem data
file images/base/build/vmlinux        # Should show: Linux kernel

# Check sizes
ls -lh images/base/build/
# Example output:
# -rw-r--r--  1 user group 2.0G Nov  5 10:42 rootfs.ext4
# -rw-r--r--  1 user group  11M Nov  5 10:43 vmlinux
# -rw-r--r--  1 user group 1.2K Nov  5 10:43 manifest.json
```

### Testing the Image Boots

```bash
# Automated boot test in Firecracker
cd images/base
make test

# Manual boot test (if you have firecracker binary)
sudo firecracker --api-sock /tmp/firecracker.sock --config-file vm-config.json
```

### Publishing to GHCR

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Push image
cd images/base
make push REGISTRY=ghcr.io/jpoley
```

---

## Debugging with VS Code

### Setup

#### 1. Install Go Extension

In VS Code:
- Extensions → Search "Go"
- Install "Go" by Go Team at Google
- Click "Install" and wait for gopls installation

#### 2. Create Launch Configuration

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "nanofuse CLI",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/nanofuse",
      "args": ["vm", "list"],
      "cwd": "${workspaceFolder}",
      "env": {
        "NANOFUSED_SOCKET": "/tmp/nanofused.sock"
      },
      "showLog": true
    },
    {
      "name": "nanofused daemon",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/nanofused",
      "args": [],
      "cwd": "${workspaceFolder}",
      "env": {
        "LOG_LEVEL": "debug"
      },
      "showLog": true,
      "preLaunchTask": "make-build",
      "postDebugTask": "cleanup-daemon"
    },
    {
      "name": "nanofused (with sudo)",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/nanofused",
      "args": [],
      "cwd": "${workspaceFolder}",
      "env": {
        "LOG_LEVEL": "debug"
      },
      "showLog": true,
      "sudo": true,
      "preLaunchTask": "ensure-socket-dir"
    },
    {
      "name": "Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/...",
      "args": ["-test.v", "-test.run", "TestMain"]
    }
  ]
}
```

#### 3. Create Build Tasks

Create `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "make-build",
      "type": "shell",
      "command": "mage",
      "args": ["all"],
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "problemMatcher": ["$go"]
    },
    {
      "label": "ensure-socket-dir",
      "type": "shell",
      "command": "mkdir",
      "args": ["-p", "/tmp/nanofuse"],
      "isBackground": false
    },
    {
      "label": "cleanup-daemon",
      "type": "shell",
      "command": "pkill",
      "args": ["-f", "nanofused"],
      "isBackground": false,
      "problemMatcher": []
    },
    {
      "label": "run-tests",
      "type": "shell",
      "command": "mage",
      "args": ["test"],
      "group": {
        "kind": "test",
        "isDefault": true
      }
    }
  ]
}
```

### Debugging Workflows

#### Debugging the CLI

```bash
# In VS Code: F5 → Select "nanofuse CLI"
# Breakpoints work automatically
# Modify args in launch.json to test different commands
# Example args: ["vm", "list"], ["image", "list"], ["vm", "status", "my-vm"]
```

#### Debugging the Daemon

```bash
# Terminal 1: Start daemon in debugger
# VS Code: F5 → Select "nanofused daemon"
#
# Terminal 2: Run CLI commands against it
$ ./bin/nanofuse vm list
$ ./bin/nanofuse image pull --default

# Breakpoints in daemon code will be hit
```

#### Common Debugging Tasks

**Setting breakpoints:**
- Click in the left margin next to line numbers
- Red dot indicates breakpoint is set

**Conditional breakpoints:**
- Right-click on breakpoint → Edit Breakpoint
- Enter condition: `vm.Name == "my-vm"`
- Breakpoint only triggers when condition is true

**Watch variables:**
- Debug panel → Watch section
- Click `+` to add expression
- Example: `vm.State`, `network.Bridge`

**Step through code:**
- F10: Step over (skip function calls)
- F11: Step into (enter function calls)
- Shift+F11: Step out (exit current function)
- F5: Continue

**Debug console:**
- Use the Debug Console at bottom
- Type variable names to inspect them
- Example: `vm`, `config`, `err`

---

## Running and Testing VMs

### Manual VM Testing

#### 1. Start the Daemon

```bash
# Build first if you haven't
mage all

# Start daemon (requires sudo for networking)
sudo ./bin/nanofused &

# Or in a separate terminal for easier debugging:
# Terminal 1:
sudo ./bin/nanofused

# Terminal 2:
export NANOFUSED_SOCKET=/tmp/nanofused.sock
./bin/nanofuse ...
```

#### 2. Pull a Base Image

```bash
# Pull the default pre-built image
./bin/nanofuse image pull --default

# Or pull from a specific registry
./bin/nanofuse image pull ghcr.io/daax-dev/nanofuse/base:latest

# View pulled images
./bin/nanofuse image list
```

#### 3. Create a VM

```bash
# Create VM named "test-vm" from default image
./bin/nanofuse vm create default test-vm

# Or specify image version
./bin/nanofuse vm create default:v1.0.0 test-vm

# Check VM was created
./bin/nanofuse vm list
./bin/nanofuse vm inspect test-vm
```

#### 4. Start and Monitor VM

```bash
# Start the VM
./bin/nanofuse vm start test-vm

# Check status (should show "running")
./bin/nanofuse vm status test-vm

# View console output
./bin/nanofuse vm logs test-vm

# Follow logs in real-time
./bin/nanofuse vm logs test-vm --follow

# Get detailed VM info
./bin/nanofuse vm inspect test-vm
```

#### 5. Stop and Clean Up

```bash
# Stop VM gracefully (10s timeout before kill)
./bin/nanofuse vm stop test-vm

# Force stop if needed
./bin/nanofuse vm stop test-vm --force

# Delete VM
./bin/nanofuse vm delete test-vm

# Verify deleted
./bin/nanofuse vm list
```

### Testing with Integration Tests

```bash
# Run all integration tests
mage testIntegration

# Run specific test
go test -v -run TestVMLifecycle ./test/integration/...

# Run with race detector (slow but catches concurrency bugs)
go test -race -v ./test/integration/...

# Run with verbose output and show all passes
go test -v -count=1 ./test/integration/...
```

---

## SSH Access

### Prerequisites for SSH Access

The base image includes OpenSSH server pre-configured, but you need:

1. **SSH key in the image** - Add your public key
2. **Network access** - VM must be on reachable network
3. **IP address** - Determine VM's IP address

### Method 1: SSH with Injected Keys (Recommended)

#### A. Add SSH Key to Image During Build

Modify `images/base/Dockerfile`:

```dockerfile
# Add your public key to image
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..." > /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys
```

Rebuild the image:

```bash
cd images/base
sudo make clean build
```

#### B. SSH Into VM

```bash
# Determine VM's IP address
./bin/nanofuse vm inspect test-vm | grep -i ip

# SSH to VM
ssh root@<VM_IP>

# Example:
ssh root@172.16.0.10
```

### Method 2: SSH Key Injection at Runtime

#### Option A: Inject via firstboot.service

Create/modify `images/base/units/firstboot.service` to inject keys:

```ini
[Unit]
Description=NanoFuse First Boot Initialization
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/firstboot.sh
RemainAfterExit=yes
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Create `images/base/scripts/firstboot.sh`:

```bash
#!/bin/bash
set -e

# Add SSH keys from metadata service or config file
if [ -f /tmp/authorized_keys ]; then
    mkdir -p /root/.ssh
    cat /tmp/authorized_keys >> /root/.ssh/authorized_keys
    chmod 600 /root/.ssh/authorized_keys
    chmod 700 /root/.ssh
fi

# Other initialization tasks...
logger "First boot complete"
```

#### Option B: Pass Keys via Socket API

```bash
# Example (if API supports it):
curl -X POST http://localhost:8080/api/vm/test-vm/ssh-keys \
  -H "Content-Type: application/json" \
  -d '{"authorized_keys": "ssh-rsa AAAA..."}'
```

### Method 3: SSH via Port Forwarding

NanoFuse supports port forwarding (currently in progress):

```bash
# Forward local port 2222 to VM's port 22
./bin/nanofuse vm port-forward test-vm 2222:22

# SSH via forwarded port
ssh -p 2222 root@127.0.0.1
```

### Debugging SSH Access

```bash
# Check SSH is running in VM
./bin/nanofuse vm logs test-vm | grep -i ssh

# Test network connectivity
ping <VM_IP>

# Check SSH is listening
./bin/nanofuse vm exec test-vm ss -tulpn | grep 22

# Verify SSH configuration
./bin/nanofuse vm exec test-vm cat /etc/ssh/sshd_config | grep -i "permit"

# Manual SSH test with verbose output
ssh -v root@<VM_IP>
```

---

## Testing Inside Running Containers

### Method 1: Execute Commands in VM

```bash
# Run single command
./bin/nanofuse vm exec test-vm whoami
./bin/nanofuse vm exec test-vm hostname
./bin/nanofuse vm exec test-vm uname -a

# Run more complex commands
./bin/nanofuse vm exec test-vm "curl -s http://localhost:8080"
./bin/nanofuse vm exec test-vm "systemctl status ssh"
./bin/nanofuse vm exec test-vm "df -h"
```

### Method 2: Interactive Shell

```bash
# Get interactive shell inside VM (if supported by API)
./bin/nanofuse vm shell test-vm

# Or via SSH:
ssh root@<VM_IP> /bin/bash
```

### Method 3: Console Access

```bash
# View full console output
./bin/nanofuse vm logs test-vm

# Follow in real-time
./bin/nanofuse vm logs test-vm --follow

# Get last 50 lines
./bin/nanofuse vm logs test-vm --tail 50
```

### Testing Systemd Services

```bash
# List all services
./bin/nanofuse vm exec test-vm systemctl list-units --type=service

# Check specific service status
./bin/nanofuse vm exec test-vm systemctl status ssh
./bin/nanofuse vm exec test-vm systemctl status systemd-networkd

# View service logs
./bin/nanofuse vm exec test-vm journalctl -u ssh -n 20

# Enable/disable service (for testing)
./bin/nanofuse vm exec test-vm systemctl restart ssh
```

### Testing Network

```bash
# Check IP address
./bin/nanofuse vm exec test-vm ip addr

# Check routes
./bin/nanofuse vm exec test-vm ip route

# Test DNS
./bin/nanofuse vm exec test-vm nslookup google.com

# Test connectivity
./bin/nanofuse vm exec test-vm ping -c 3 8.8.8.8

# Check listening ports
./bin/nanofuse vm exec test-vm ss -tulpn
```

### Testing HTTP Server (if enabled)

The base image includes an HTTP test server. Test it:

```bash
# Get VM IP
VM_IP=$(./bin/nanofuse vm inspect test-vm | grep -i ip | head -1 | awk '{print $NF}')

# Test HTTP server from host
curl http://$VM_IP:8080

# Or from inside VM
./bin/nanofuse vm exec test-vm curl http://localhost:8080
```

---

## Building Custom Layers

### Understanding the Image Layers

The complete NanoFuse stack:

```
Layer 0 (Base): Ubuntu 24.04 (from Docker Hub)
  ↓
Layer 1: System packages (systemd, SSH, networking)
  ↓
Layer 2: Configuration (services, network, SSH keys)
  ↓
Layer 3: NanoFuse utils (firstboot.service, monitoring)
  ↓
Layer N: Your custom application layer
```

### Creating a Custom Image

#### Step 1: Create a Custom Dockerfile

Create `images/myapp/Dockerfile`:

```dockerfile
# Start from NanoFuse base image (once published to GHCR)
FROM ghcr.io/daax-dev/nanofuse/base:latest

# Install your application dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        python3 \
        python3-pip \
        git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install your application
RUN mkdir -p /opt/myapp && \
    cd /opt/myapp

# Copy application files (if building locally)
# COPY app/ /opt/myapp/

# Create systemd service for your app
RUN cat > /etc/systemd/system/myapp.service << 'EOF'
[Unit]
Description=My Application
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/bin/python3 /opt/myapp/main.py
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable the service (NEVER use 'start')
RUN systemctl enable myapp.service

# Expose port if needed
EXPOSE 8000

# Keep CMD as inherited from base (systemd)
```

#### Step 2: Create Build Script

Create `images/myapp/build.sh` (similar to base):

```bash
#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

BUILD_DIR="./build"
IMAGE_NAME="${IMAGE_NAME:-myapp}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ROOTFS_SIZE="${ROOTFS_SIZE:-4096}"  # 4GB for app image

echo "Building $IMAGE_NAME:$IMAGE_TAG..."

# Build Docker image
docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

# Export and create ext4 (same as base image)
mkdir -p "${BUILD_DIR}/rootfs"
CONTAINER_ID=$(docker create "${IMAGE_NAME}:${IMAGE_TAG}")
docker export "${CONTAINER_ID}" | tar -C "${BUILD_DIR}/rootfs" -xf -
docker rm "${CONTAINER_ID}"

# Create ext4 image
dd if=/dev/zero of="${BUILD_DIR}/rootfs.ext4" bs=1M count="${ROOTFS_SIZE}"
mkfs.ext4 -F -q -L myapp-root "${BUILD_DIR}/rootfs.ext4"

mkdir -p "${BUILD_DIR}/mnt"
mount -o loop "${BUILD_DIR}/rootfs.ext4" "${BUILD_DIR}/mnt"
cp -a "${BUILD_DIR}/rootfs/"* "${BUILD_DIR}/mnt/"
umount "${BUILD_DIR}/mnt"
rmdir "${BUILD_DIR}/mnt"
rm -rf "${BUILD_DIR}/rootfs"

# Download kernel (same as base)
curl -fsSL -o "${BUILD_DIR}/vmlinux" \
    https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin

# Generate manifest
cat > "${BUILD_DIR}/manifest.json" << EOF
{
  "version": "0.1.0",
  "name": "${IMAGE_NAME}",
  "tag": "${IMAGE_TAG}",
  "architecture": "x86_64",
  "base_image": "ghcr.io/daax-dev/nanofuse/base:latest",
  "built_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "Build complete! Artifacts in $BUILD_DIR/"
ls -lh "$BUILD_DIR/"
```

#### Step 3: Build Your Custom Image

```bash
cd images/myapp
sudo bash build.sh

# Or with custom tags
IMAGE_NAME=myapp IMAGE_TAG=v1.0.0 sudo bash build.sh
```

#### Step 4: Register Custom Image

```bash
# Register locally with NanoFuse
sudo ./bin/register-local-image \
  -name myapp:v1.0.0 \
  -kernel ./images/myapp/build/vmlinux \
  -rootfs ./images/myapp/build/rootfs.ext4

# Test it
./bin/nanofuse image list
./bin/nanofuse vm create myapp:v1.0.0 test-app
```

### Advanced: Multi-Stage Builds

For efficient builds, use Docker multi-stage:

```dockerfile
# Stage 1: Build application
FROM golang:1.21 AS builder
WORKDIR /build
COPY . .
RUN go build -o myapp .

# Stage 2: Runtime
FROM ghcr.io/daax-dev/nanofuse/base:latest
COPY --from=builder /build/myapp /usr/local/bin/myapp

RUN systemctl enable myapp.service
```

This keeps the final image small by excluding build tools.

### Publishing Custom Images

```bash
# Tag for registry
docker tag myapp:latest ghcr.io/myorg/myapp:v1.0.0

# Login
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Push
docker push ghcr.io/myorg/myapp:v1.0.0

# Others can now pull
nanofuse image pull ghcr.io/myorg/myapp:v1.0.0
```

---

## Troubleshooting

### Build Issues

#### "This script requires sudo"

```bash
# Solution: Run with sudo
sudo make build

# Or add yourself to docker group (not recommended for security)
sudo usermod -aG docker $USER
newgrp docker
```

#### "Docker build failed: permission denied"

```bash
# Solution: Check Docker daemon is running
sudo systemctl start docker

# Or check Docker socket permissions
ls -la /var/run/docker.sock
```

#### "Unable to download kernel"

```bash
# Verify URL accessibility
curl -I https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin

# Check internet connection
ping 8.8.8.8

# Try with wget as fallback
wget https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
```

### VM Boot Issues

#### "VM fails to boot / No console output"

```bash
# Check artifact validity
file images/base/build/vmlinux        # Must show "Linux kernel"
file images/base/build/rootfs.ext4    # Must show "ext4 filesystem"

# Check Firecracker logs
journalctl -u firecracker -n 50

# Verify kernel is uncompressed
strings images/base/build/vmlinux | head -1  # Should show "Linux version"
```

#### "systemd did not reach multi-user.target"

```bash
# View full console output
./bin/nanofuse vm logs test-vm

# Check for failed services
./bin/nanofuse vm logs test-vm | grep -i "failed"

# Rebuild without docker cache (fresh build)
cd images/base
docker build --no-cache -t nanofuse-base:latest .
sudo make build
```

#### "Network not working in VM"

```bash
# Check bridge interface on host
ip addr show nanofuse0

# Check DHCP server running
ps aux | grep dhcp

# View VM network config
./bin/nanofuse vm exec test-vm cat /etc/systemd/network/20-wired.network

# Test network from within VM
./bin/nanofuse vm exec test-vm ping -c 3 8.8.8.8
./bin/nanofuse vm exec test-vm ip addr
```

### Debugging Problems

#### "Breakpoint not hit in VS Code"

```bash
# Check build mode
mage all

# Verify you're debugging correct binary
file ./bin/nanofused | grep -i debug

# Try explicit debug build
CGO_ENABLED=1 go build -gcflags="all=-N -l" -o ./bin/nanofused ./cmd/nanofused

# Restart VS Code
```

#### "Socket permission denied"

```bash
# Check socket exists and permissions
ls -la /tmp/nanofused.sock

# Recreate socket directory
sudo mkdir -p /tmp/nanofuse
sudo chmod 755 /tmp/nanofuse

# Or run daemon with proper permissions
sudo ./bin/nanofused
```

### SSH Access Issues

#### "Connection refused on port 22"

```bash
# Check SSH is running in VM
./bin/nanofuse vm logs test-vm | grep -i ssh

# Restart SSH service
./bin/nanofuse vm exec test-vm systemctl restart ssh

# Check listening ports
./bin/nanofuse vm exec test-vm ss -tulpn | grep 22
```

#### "Permission denied (publickey)"

```bash
# Verify authorized_keys in image
./bin/nanofuse vm exec test-vm cat /root/.ssh/authorized_keys

# Check SSH key permissions in image
./bin/nanofuse vm exec test-vm ls -la /root/.ssh/

# Rebuild image with your public key
# (See SSH Access section above)
```

### Performance Issues

#### "VM takes too long to boot"

```bash
# Check console output for slowdowns
./bin/nanofuse vm logs test-vm

# Look for "A stop job is running..." messages (usually disk-related)
# Reduce rootfs size or use faster storage

# Profile systemd startup
./bin/nanofuse vm exec test-vm systemd-analyze

# Check for network timeouts
./bin/nanofuse vm logs test-vm | grep -i "timeout"
```

#### "High CPU usage while VM idle"

```bash
# Check for spinning processes
./bin/nanofuse vm exec test-vm top

# View system load
./bin/nanofuse vm exec test-vm uptime

# Check for busy-waiting in journalctl
./bin/nanofuse vm exec test-vm journalctl -b -p err
```

---

## Quick Reference

### Most Used Commands

```bash
# Building
mage all                                 # Build all binaries
cd images/base && sudo make build        # Build base image

# Running
sudo ./bin/nanofused                     # Start daemon
./bin/nanofuse vm create default vm1     # Create VM
./bin/nanofuse vm start vm1              # Start VM
./bin/nanofuse vm logs vm1 --follow      # Watch VM boot

# Debugging
# F5 in VS Code → Select configuration
./bin/nanofuse vm exec vm1 <command>     # Run command in VM

# Cleanup
./bin/nanofuse vm stop vm1               # Stop VM
./bin/nanofuse vm delete vm1             # Delete VM
mage clean                               # Remove binaries
```

### Key Directories

```
nanofuse/
├── cmd/                 # Source code (CLI, daemon)
├── images/base/         # Base microVM image definition
├── internal/            # Core packages (API, networking, firecracker)
├── docs/                # Documentation
├── scripts/             # Helper scripts
├── test/                # Integration tests
└── .vscode/             # VS Code configuration (create this)
```

### Environment Variables

```bash
# Runtime
NANOFUSED_SOCKET=/tmp/nanofused.sock    # Daemon socket path
LOG_LEVEL=debug                          # Logging level

# Build
CGO_ENABLED=1                            # Required for daemon (SQLite)
GORACE=halt_on_error=1                   # Crash on race condition

# Image build
IMAGE_NAME=nanofuse-base                 # Docker image name
IMAGE_TAG=latest                         # Docker image tag
ROOTFS_SIZE=2048                         # MB for ext4 image
```

---

## Next Steps

- **Run your first VM**: Follow "Running and Testing VMs" section
- **Debug the daemon**: Set up VS Code and practice breakpoints
- **Build a custom image**: Create your own layer following the examples
- **Run integration tests**: Execute `mage testIntegration` to validate everything
- **Check latest docs**: See `docs/` directory for phase-specific guides

For questions or issues, check the main [README.md](../README.md) and [PHASE_1_COMPLETE.md](PHASE_1_COMPLETE.md).
