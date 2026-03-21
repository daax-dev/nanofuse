# Firecracker MicroVM Image Build Guide

Complete reference for building, debugging, and customizing the NanoFuse Firecracker base image.

**Focus**: Container image build process, NOT the CLI/daemon (see [DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md) for that)

## Table of Contents

1. [Quick Start](#quick-start)
2. [Understanding the Build](#understanding-the-build)
3. [Build Process Breakdown](#build-process-breakdown)
4. [Dockerfile Deep Dive](#dockerfile-deep-dive)
5. [Debug the Build](#debug-the-build)
6. [Testing the Image](#testing-the-image)
7. [SSH Access to VMs](#ssh-access-to-vms)
8. [Testing Inside Running VMs](#testing-inside-running-vms)
9. [Building Custom Layers](#building-custom-layers)
10. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Prerequisites

```bash
# Docker installed and running
docker --version

# Root/sudo access (for ext4 filesystem operations)
sudo -v

# Linux kernel with KVM support (optional, for testing)
ls -la /dev/kvm

# Firecracker binary (optional, for boot testing)
which firecracker || echo "Not installed - optional"
```

### Build in 90 Seconds

```bash
cd images/base

# Build everything (Docker image → ext4 filesystem → kernel → manifest)
sudo make build

# Expected output:
# ✓ Docker image built: nanofuse-base:latest
# ✓ Rootfs extracted and formatted as ext4
# ✓ Kernel downloaded from Firecracker CI
# ✓ Manifest generated

# Check artifacts
ls -lh build/
# rootfs.ext4  (2GB ext4 filesystem)
# vmlinux      (Linux kernel 5.10.204)
# manifest.json (metadata)
```

---

## Understanding the Build

### Base Image Selection: Why Ubuntu 24.04?

You might wonder: "Isn't Ubuntu 24.04 massive for a microVM?"

**Answer: No.** Here's the actual data:

| Base Image | Base Size | With systemd+SSH | Why/Why Not |
|-----------|-----------|------------------|-----------|
| **Ubuntu 24.04** (CHOSEN) | 78 MB | **117 MB** | ✅ Perfect - small, has systemd, reliable |
| Debian bookworm-slim | 74 MB | 155 MB | ❌ Larger than Ubuntu, less common |
| Alpine | 8 MB | Can't use - **no systemd** | ❌ Missing core init system |
| Busybox | ~5 MB | Can't use - **no systemd** | ❌ Too minimal |

**Why not Alpine?**
- Alpine is only 8 MB but uses musl libc and OpenRC
- Firecracker VMs need `systemd` as PID 1 (init system)
- Alpine doesn't have systemd available
- Result: unusable for this use case

**Why Ubuntu 24.04 + minimal packages?**
- Only **117 MB** for Docker image (not massive)
- Converts to ~120-140 MB ext4 filesystem (reasonable)
- Has systemd, openssh, and all dependencies we need
- Latest LTS with long-term support
- Well-tested in production

**Key optimization: Use MINIMAL packages only**

Don't include: wget, vim, less, python3, kmod, udev, iputils-ping (unless needed)

Just install the 7 essentials:
1. systemd
2. systemd-sysv
3. openssh-server
4. ca-certificates
5. curl
6. dbus
7. iproute2

This keeps image at **117 MB** instead of 182 MB.

---

### What Gets Built

The build creates **three artifacts** for use with Firecracker:

| Artifact | Purpose | Size | Format |
|----------|---------|------|--------|
| **rootfs.ext4** | Root filesystem (OS, services, config) | 2GB | ext4 block device |
| **vmlinux** | Linux kernel | ~11MB | Uncompressed ELF |
| **manifest.json** | Image metadata | <1KB | JSON |

### The Build Flow

```
Dockerfile (ubuntu:24.04)
     ↓
docker build → Docker image (systemd, SSH, networking)
     ↓
docker create + docker export → Filesystem tarball
     ↓
mkfs.ext4 + mount + copy → rootfs.ext4 (block device)
     ↓
curl download → vmlinux (Firecracker kernel 5.10.204)
     ↓
generate metadata → manifest.json
     ↓
Ready for Firecracker!
```

### Why This Approach?

1. **Docker for building** - Familiar, reproducible, layer caching
2. **Export to ext4** - Firecracker doesn't support Docker images directly, needs block devices
3. **Separate kernel** - Kernel is downloaded, not included in rootfs (it's loaded separately by Firecracker)
4. **Manifest** - Machine-readable metadata for CLI/API integration

---

## Build Process Breakdown

### Step 1: Docker Build

**What happens:**
```bash
docker build -t nanofuse-base:latest .
```

**What the Dockerfile does:**
- Starts from `ubuntu:24.04`
- Installs packages: systemd, openssh-server, curl, networking tools
- Configures systemd services (enable ssh, networking, console)
- Configures SSH (key-only, no passwords)
- Creates firstboot service
- Creates network config files
- Sets `systemctl set-default multi-user.target`

**Time**: ~2-3 minutes (first build), ~30 seconds (cached)

**Key files created in container:**
- `/etc/systemd/system/ssh.service` → enabled
- `/etc/systemd/system/systemd-networkd.service` → enabled
- `/etc/systemd/network/20-wired.network` → DHCP config
- `/etc/systemd/system/serial-getty@ttyS0.service` → console
- `/etc/systemd/system/firstboot.service` → first-boot init
- `/etc/ssh/sshd_config` → SSH configuration
- `/sbin/init` → symlink to systemd

### Step 2: Container Export

**What happens:**
```bash
CONTAINER_ID=$(docker create nanofuse-base:latest)
docker export ${CONTAINER_ID} | tar -C ./build/rootfs -xf -
docker rm ${CONTAINER_ID}
```

**What this does:**
- Creates a stopped container from the Docker image
- Exports the entire filesystem as a tar stream
- Extracts to `build/rootfs/` directory
- Removes the temporary container

**Time**: ~30 seconds

**Output**: `build/rootfs/` directory with full filesystem:
```
build/rootfs/
├── bin/
├── etc/
│   ├── systemd/
│   │   ├── network/20-wired.network
│   │   └── system/
│   │       ├── serial-getty@ttyS0.service
│   │       ├── ssh.service
│   │       ├── systemd-networkd.service
│   │       └── firstboot.service
│   └── ssh/sshd_config
├── lib/
├── root/.ssh/          # Empty, will hold authorized_keys
├── sbin/init           # systemd init system
├── usr/
├── var/
└── ... (all Ubuntu files)
```

### Step 3: Create ext4 Filesystem

**What happens:**
```bash
# Create 2GB sparse file
dd if=/dev/zero of=./build/rootfs.ext4 bs=1M count=2048

# Format as ext4
mkfs.ext4 -F -q -L nanofuse-root ./build/rootfs.ext4
```

**Why ext4?**
- Firecracker requires block device format, not Docker layers
- ext4 is standard, reliable, well-supported
- Sparse files don't actually use 2GB on disk (only ~500-800MB)

**Time**: ~15 seconds

**Output**: `build/rootfs.ext4` - empty ext4 filesystem ready for data

### Step 4: Mount and Copy Filesystem

**What happens:**
```bash
# Mount the ext4 image
mount -o loop ./build/rootfs.ext4 ./build/mnt

# Copy all files from Docker export
cp -a ./build/rootfs/* ./build/mnt/

# Unmount
umount ./build/mnt

# Set permissions so non-root can use it
chown $(id -u):$(id -g) ./build/rootfs.ext4
chmod 664 ./build/rootfs.ext4
```

**What this does:**
- Mounts ext4 as loop device (no special hardware needed)
- Copies entire Ubuntu filesystem into ext4
- Preserves permissions, ownership, symlinks
- Cleans up and unmounts
- Makes file usable by non-root user

**Time**: ~30 seconds

**Result**: `build/rootfs.ext4` now contains complete, bootable filesystem

### Step 5: Download Firecracker Kernel

**What happens:**
```bash
curl -fsSL -o ./build/vmlinux \
  https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
```

**What this is:**
- Official Firecracker CI kernel version 5.10.204
- Uncompressed (vmlinux, not vmlinuz or bzImage)
- Proven compatible with Firecracker
- ~11MB
- Public download, no authentication

**Why separate from rootfs?**
- Firecracker loads kernel separately from block device
- Kernel config is fixed at boot time, not part of rootfs
- Enables fast kernel upgrades without rebuilding rootfs

**Time**: ~5 seconds

**Result**: `build/vmlinux` - bootable Linux kernel

### Step 6: Generate Manifest

**What happens:**
```bash
cat > ./build/manifest.json << EOF
{
  "version": "0.1.0",
  "name": "nanofuse-base",
  "tag": "latest",
  "architecture": "x86_64",
  "base_os": "ubuntu:24.04",
  "kernel": {
    "version": "5.10.204",
    "source": "Firecracker CI",
    "cmdline": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k",
    "file": "vmlinux"
  },
  "rootfs": {
    "format": "ext4",
    "file": "rootfs.ext4",
    "size_bytes": <ACTUAL_SIZE>
  },
  "services": {
    "ssh": {"enabled": true, "port": 22},
    "systemd-networkd": {"enabled": true}
  },
  "built_at": "2025-11-05T10:30:00Z"
}
EOF
```

**What this is:**
- Machine-readable image metadata
- Used by nanofuse CLI/daemon to configure VMs
- Contains kernel cmdline for Firecracker boot
- Records image version, size, enabled services

**Time**: <1 second

**Result**: `build/manifest.json` - metadata file

---

## Dockerfile Deep Dive

Location: `images/base/Dockerfile`

### Base Image Selection

```dockerfile
FROM ubuntu:24.04
```

**Why Ubuntu 24.04?**
- Latest LTS (long-term support)
- Modern systemd (v256+)
- Familiar package ecosystem
- Good security patch cadence

### Environment Setup

```dockerfile
ENV DEBIAN_FRONTEND=noninteractive
```

**Critical for automated builds** - prevents apt from asking interactive questions

### Package Installation

**IMPORTANT: Keep packages MINIMAL**

```dockerfile
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        systemd \
        systemd-sysv \
        openssh-server \
        ca-certificates \
        curl \
        dbus \
        iproute2 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
```

**Essential packages only** (others are optional):

| Package | Purpose | Essential? |
|---------|---------|-----------|
| `systemd` | Init system (PID 1) | ✅ YES |
| `systemd-sysv` | Provides `/sbin/init` symlink | ✅ YES |
| `openssh-server` | SSH daemon | ✅ YES |
| `ca-certificates` | TLS/HTTPS support | ✅ YES |
| `curl` | Download tool | ✅ YES |
| `dbus` | System message bus (required by systemd) | ✅ YES |
| `iproute2` | Network configuration (`ip` command) | ✅ YES |
| `wget` | Download tool | ❌ NO (curl is enough) |
| `iputils-ping` | Network diagnostics | ❌ NO (nice-to-have) |
| `kmod` | Kernel module management | ❌ NO (rarely needed) |
| `udev` | Device management | ❌ NO (systemd has built-in) |
| `vim-tiny, less` | Text editors | ❌ NO (vi/ed available) |
| `python3` | Optional scripting | ❌ NO (add if needed) |

**Size Impact:**

- **With minimal packages (7 only)**: 117 MB Docker image
- **With all packages (13)**: 182 MB Docker image
- **Difference**: ~65 MB (56% larger!)

**Recommendation**: Remove non-essential packages unless you specifically need them:
- Don't add `wget` (curl can do everything)
- Don't add `iputils-ping` (basic network test only)
- Don't add `kmod` (kernel modules rarely needed in Firecracker)
- Don't add `udev` (systemd handles devices)
- Don't add `vim, less` (vi/ed available, or install later if needed)
- Only add `python3` if your app specifically needs it

**`--no-install-recommends`** keeps image small by excluding optional dependencies

**`apt-get clean && rm -rf /var/lib/apt/lists/*`** removes package cache (~100-200MB saved)

### Systemd Configuration

```dockerfile
# Mask unnecessary services that cause issues in containers
RUN systemctl mask \
    systemd-resolved.service \
    systemd-timesyncd.service \
    systemd-logind.service \
    getty@.service
```

**Why mask these?**
- `systemd-resolved` - DNS resolution not needed (VM gets DNS from host)
- `systemd-timesyncd` - Time sync handled by host
- `systemd-logind` - Login manager for headless VM (not needed)
- `getty@.service` - Virtual terminal (using serial console instead)

**Difference between mask vs disable:**
- `mask`: Prevents service from running even if requested
- `disable`: Just removes from auto-start

### Enable Essential Services

```dockerfile
# Configure serial console on ttyS0 (required for Firecracker)
RUN systemctl enable serial-getty@ttyS0.service

# Enable SSH server (will start on VM boot)
RUN systemctl enable ssh.service

# Enable systemd-networkd for DHCP-based networking
RUN systemctl enable systemd-networkd.service
```

**Critical: Enable, don't start**

❌ **WRONG** - Services won't run in container context:
```dockerfile
RUN systemctl start ssh.service
```

✅ **RIGHT** - Services will start when VM boots:
```dockerfile
RUN systemctl enable ssh.service
```

### Network Configuration

```dockerfile
RUN mkdir -p /etc/systemd/network && \
    echo '[Match]' > /etc/systemd/network/20-wired.network && \
    echo 'Name=en*' >> /etc/systemd/network/20-wired.network && \
    echo '' >> /etc/systemd/network/20-wired.network && \
    echo '[Network]' >> /etc/systemd/network/20-wired.network && \
    echo 'DHCP=yes' >> /etc/systemd/network/20-wired.network && \
    echo 'LinkLocalAddressing=yes' >> /etc/systemd/network/20-wired.network
```

**File created: `/etc/systemd/network/20-wired.network`**

```ini
[Match]
Name=en*

[Network]
DHCP=yes
LinkLocalAddressing=yes
```

**What this does:**
- Matches any interface starting with `en` (ethernet)
- Gets IP via DHCP from Firecracker network
- Falls back to link-local if DHCP unavailable
- Number `20` controls priority (lower numbers processed first)

### SSH Configuration

```dockerfile
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
```

**What this configures:**
- Creates `/root/.ssh/` directory (empty, user adds keys later)
- Permits root login, but only with SSH key (no passwords)
- Disables password authentication entirely

**Security implications:**
- Users MUST inject their SSH public key to access VM
- No brute-force password attacks possible
- Standard for cloud images

### First-Boot Service

```dockerfile
COPY units/firstboot.service /etc/systemd/system/firstboot.service
RUN systemctl enable firstboot.service
```

**What it does:**
- Runs once on first VM boot
- Performs initialization tasks (DNS config, diagnostics logging)
- One-shot service (won't restart)

**See `images/base/units/firstboot.service` for details**

### Boot Target

```dockerfile
RUN systemctl set-default multi-user.target
```

**Sets default systemd target to `multi-user.target`**
- Multi-user system (CLI only, no GUI)
- All core services running
- No graphics/desktop environment

### Final CMD

```dockerfile
CMD ["/sbin/init"]
```

**Note:** This CMD won't be used by Firecracker

- Firecracker boots kernel directly, bypassing container runtime
- CMD is for container compatibility only
- In VM, the kernel loads `/sbin/init` (systemd) automatically

---

## Debug the Build

### Approach 1: Step-by-Step Builds

Stop at each build step to inspect:

```bash
cd images/base

# Step 1: Build Docker image only
docker build -t nanofuse-base:latest .

# Inspect what's inside
docker run --rm -it nanofuse-base:latest /bin/bash

# Check systemd services
docker run --rm nanofuse-base:latest systemctl list-unit-files | grep enabled

# Check network config
docker run --rm nanofuse-base:latest cat /etc/systemd/network/20-wired.network

# Check SSH config
docker run --rm nanofuse-base:latest grep -i "PermitRootLogin\|PasswordAuth" /etc/ssh/sshd_config
```

### Approach 2: Inspect Intermediate Artifacts

After Docker build, before ext4 creation:

```bash
# Create and export container manually
docker create -t nanofuse-base:latest tmp-inspect
docker export tmp-inspect | tar -tvf - | head -50

# Extract and inspect filesystem
mkdir -p /tmp/inspect
docker export tmp-inspect | tar -C /tmp/inspect -xf -

# Check files exist
ls -la /tmp/inspect/etc/systemd/network/
ls -la /tmp/inspect/etc/systemd/system/ | grep -E "(ssh|networkd|getty|firstboot)"
ls -la /tmp/inspect/root/.ssh/

# Verify kernel cmdline location
grep -r "console=ttyS0" /tmp/inspect/etc/ 2>/dev/null || echo "Check Dockerfile for kernel cmdline"

# Cleanup
docker rm tmp-inspect
rm -rf /tmp/inspect
```

### Approach 3: Interactive Docker Shell

```bash
# Start shell in built image
make shell

# Now inside the container:
root@abc1234# systemctl list-units --type=service
root@abc1234# cat /etc/systemd/network/20-wired.network
root@abc1234# cat /etc/ssh/sshd_config | grep -i "permit\|password"
root@abc1234# ls -la /root/.ssh/
root@abc1234# systemctl get-default
root@abc1234# file /sbin/init
root@abc1234# exit
```

### Approach 4: Modify and Rebuild

Make targeted changes to test:

```bash
# Edit Dockerfile
nano Dockerfile

# Example: Add debug package
# RUN apt-get install -y strace

# Rebuild (uses Docker cache up to change point)
docker build -t nanofuse-base:test .

# Test it
docker run --rm -it nanofuse-base:test /bin/bash
```

### Approach 5: Debug ext4 Creation

```bash
# Build Docker image only
docker build -t nanofuse-base:latest .

# Manually run steps 2-4 with debugging
cd images/base

# Step 2: Export with output
CONTAINER_ID=$(docker create nanofuse-base:latest)
echo "Container: $CONTAINER_ID"
docker export $CONTAINER_ID | tar -C ./build/rootfs -xvf - 2>&1 | head -20
docker rm $CONTAINER_ID

# Step 3: Create ext4 with output
dd if=/dev/zero of=./build/rootfs.ext4.debug bs=1M count=2048 status=progress
mkfs.ext4 -F -v -L nanofuse-root ./build/rootfs.ext4.debug 2>&1

# Step 4: Mount and copy with logging
mkdir -p ./build/mnt
sudo mount -v -o loop ./build/rootfs.ext4.debug ./build/mnt
sudo cp -av ./build/rootfs/* ./build/mnt/ 2>&1 | tail -20
sudo umount -v ./build/mnt
sudo chown $(id -u):$(id -g) ./build/rootfs.ext4.debug

# Inspect result
ls -lh ./build/rootfs.ext4.debug
file ./build/rootfs.ext4.debug

# Cleanup
rm ./build/rootfs.ext4.debug
```

### Approach 6: Validate Each Artifact

```bash
# Validate rootfs
file ./build/rootfs.ext4
fsck.ext4 -n ./build/rootfs.ext4  # Read-only check

# Check rootfs contents (requires mount)
sudo mkdir -p /tmp/inspect
sudo mount -o loop,ro ./build/rootfs.ext4 /tmp/inspect
sudo ls -la /tmp/inspect/etc/systemd/network/
sudo ls -la /tmp/inspect/etc/systemd/system/
sudo umount /tmp/inspect

# Validate kernel
file ./build/vmlinux
strings ./build/vmlinux | grep -i "linux version" | head -1

# Validate manifest
cat ./build/manifest.json | jq .
```

---

## Testing the Image

### Quick Validation

```bash
cd images/base

# Validate artifacts (no boot required)
sudo make validate

# Expected output:
# ✓ Rootfs file exists
# ✓ Rootfs is valid ext4 filesystem
# ✓ Kernel file exists
# ✓ Kernel is valid Linux kernel
# ✓ Manifest is valid JSON
```

### Boot Test in Firecracker

```bash
cd images/base

# Requires: firecracker binary, /dev/kvm access
sudo make test

# Expected output:
# ✓ VM booted successfully in 1.234s
# ✓ Test 1: VM boots successfully
# ✓ Test 2: Console output visible on ttyS0
# ✓ Test 3: systemd reaches multi-user.target
# ✓ Test 4: SSH daemon running
# ✓ Test 5: Network configured (systemd-networkd)
# ✓ Test 6: No failed systemd units
# ✓ Test 7: Boot time < 2s
# Overall: PASS
```

### Manual Boot Test

Create `vm-config.json`:

```json
{
  "boot-source": {
    "kernel_image_path": "/home/jpoley/src/_mine/nanofuse/images/base/build/vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "/home/jpoley/src/_mine/nanofuse/images/base/build/rootfs.ext4",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 2,
    "mem_size_mib": 512
  }
}
```

Start VM:

```bash
# Terminal 1: Start Firecracker
sudo firecracker --api-sock /tmp/fc.sock --config-file vm-config.json

# Watch console output
# Should see:
# - Linux kernel boot messages
# - systemd starting
# - "Reached target Multi-User System"
# - "systemd[1]: Started OpenSSH server daemon."
# - Login prompt

# Ctrl+C to stop
```

### Inspect Boot Logs

```bash
# After boot test, view full console
cat /tmp/nanofuse-test-*/console.log | less

# Look for:
# ✓ "Reached target Multi-User System" (systemd reached target)
# ✓ "Started OpenSSH server daemon" (SSH running)
# ✓ "Started Network Service" (networkd running)
# ✓ No "FAILED" entries in systemd logs
# ✓ Boot time metrics
```

### Test Boot Time

The boot test script measures boot time:

```bash
# Extract boot time from test output
sudo make test 2>&1 | grep -i "boot time\|in [0-9]"

# Target: < 2 seconds
# Acceptable: < 5 seconds
# Warning: > 5 seconds (investigate systemd slow units)
```

---

## SSH Access to VMs

### Important: SSH Host Keys

Each VM automatically gets **unique SSH host keys** on first boot:

1. Dockerfile removes pre-generated keys: `rm -f /etc/ssh/ssh_host_*_key*`
2. Firstboot service regenerates them: `dpkg-reconfigure openssh-server`
3. Result: Each VM has different SSH host keys (secure!)

This prevents SSH client warnings about "remote host identification changed".

See [SSH_HOST_KEYS_ISSUE.md](SSH_HOST_KEYS_ISSUE.md) for why this matters.

### Prerequisites for SSH Access

1. **SSH public key** - Add to image or inject at runtime
2. **VM network connectivity** - VM gets IP via DHCP
3. **SSH daemon running** - Enabled in Dockerfile (should auto-start)
4. **SSH host keys** - Auto-generated on first boot (no action needed)

### Method 1: Bake SSH Key Into Image

Best for testing, add your key to Dockerfile:

```dockerfile
# Add your SSH public key
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7VJT..." > /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys
```

Get your public key:

```bash
cat ~/.ssh/id_rsa.pub
# or
ssh-keyscan localhost

# or generate new key
ssh-keygen -t rsa -b 4096 -f ~/.ssh/fc-test -C "fc-test"
cat ~/.ssh/fc-test.pub
```

Rebuild image:

```bash
cd images/base
nano Dockerfile  # Add your key

sudo make build
```

Boot VM and SSH:

```bash
# Get VM IP from console or dhcp logs
VM_IP=172.16.0.10

# SSH with your key
ssh root@$VM_IP

# Or specify key explicitly
ssh -i ~/.ssh/fc-test root@$VM_IP
```

### Method 2: Inject Key at Runtime

For production, keys come from outside (cloud-init, API, metadata service):

Create `firstboot.sh`:

```bash
#!/bin/bash
# runs once on first boot

# Get SSH keys from metadata (example)
if [ -f /tmp/authorized_keys ]; then
    mkdir -p /root/.ssh
    cat /tmp/authorized_keys >> /root/.ssh/authorized_keys
    chmod 600 /root/.ssh/authorized_keys
    chmod 700 /root/.ssh
    logger "SSH keys injected from /tmp/authorized_keys"
fi

# Or generate a host key pair for testing
if [ ! -f /root/.ssh/id_rsa ]; then
    ssh-keygen -t rsa -N "" -f /root/.ssh/id_rsa
    logger "Generated SSH keypair"
fi
```

Update Dockerfile to copy and enable:

```dockerfile
COPY scripts/firstboot.sh /usr/local/bin/firstboot.sh
RUN chmod +x /usr/local/bin/firstboot.sh

RUN cat > /etc/systemd/system/firstboot.service << 'EOF'
[Unit]
Description=First Boot Initialization
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/firstboot.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF

RUN systemctl enable firstboot.service
```

### Method 3: Port Forwarding (Future)

Once NanoFuse supports port forwarding:

```bash
# Forward host port 2222 to VM port 22
nanofuse vm port-forward my-vm 2222:22

# SSH via forward
ssh -p 2222 root@127.0.0.1
```

### Debugging SSH Access

```bash
# Check SSH is running
./bin/nanofuse vm exec test-vm systemctl status ssh

# Check SSH is listening
./bin/nanofuse vm exec test-vm ss -tulpn | grep 22

# Check authorized_keys
./bin/nanofuse vm exec test-vm cat /root/.ssh/authorized_keys

# Test SSH verbose
ssh -vvv root@$VM_IP

# Check SSH config
./bin/nanofuse vm exec test-vm cat /etc/ssh/sshd_config | grep -i "permit\|password\|pubkey"

# View SSH logs
./bin/nanofuse vm exec test-vm journalctl -u ssh -n 20

# Manually start SSH (if stopped)
./bin/nanofuse vm exec test-vm systemctl restart ssh
```

---

## Testing Inside Running VMs

### Execute Commands

```bash
# Single command
./bin/nanofuse vm exec test-vm whoami
./bin/nanofuse vm exec test-vm hostname
./bin/nanofuse vm exec test-vm uname -a

# Check services
./bin/nanofuse vm exec test-vm systemctl status ssh
./bin/nanofuse vm exec test-vm systemctl status systemd-networkd

# Check network
./bin/nanofuse vm exec test-vm ip addr
./bin/nanofuse vm exec test-vm ip route
./bin/nanofuse vm exec test-vm ping -c 1 8.8.8.8

# Check disk
./bin/nanofuse vm exec test-vm df -h
./bin/nanofuse vm exec test-vm lsblk

# Check system
./bin/nanofuse vm exec test-vm free -h
./bin/nanofuse vm exec test-vm uptime
./bin/nanofuse vm exec test-vm ps aux
```

### View Logs

```bash
# Full console output
./bin/nanofuse vm logs test-vm | less

# Follow in real-time
./bin/nanofuse vm logs test-vm --follow

# Last 50 lines
./bin/nanofuse vm logs test-vm --tail 50

# Search logs
./bin/nanofuse vm logs test-vm | grep -i "error\|failed\|warning"
```

### Test Systemd

```bash
# List all units
./bin/nanofuse vm exec test-vm systemctl list-units

# Check unit status
./bin/nanofuse vm exec test-vm systemctl status ssh.service
./bin/nanofuse vm exec test-vm systemctl status systemd-networkd.service

# Show failed units
./bin/nanofuse vm exec test-vm systemctl list-units --failed

# View detailed logs
./bin/nanofuse vm exec test-vm journalctl -b -e -n 50

# Watch logs in real-time
./bin/nanofuse vm exec test-vm journalctl -f
```

### Test Network

```bash
# Get IP address
./bin/nanofuse vm exec test-vm ip addr show

# Test DNS
./bin/nanofuse vm exec test-vm nslookup google.com
./bin/nanofuse vm exec test-vm ping -c 3 google.com

# Check routing
./bin/nanofuse vm exec test-vm ip route
./bin/nanofuse vm exec test-vm traceroute 8.8.8.8

# Network statistics
./bin/nanofuse vm exec test-vm ss -s
./bin/nanofuse vm exec test-vm netstat -an
```

---

## Building Custom Layers

### Layer Architecture

```
Layer 0: Base Image (ubuntu:24.04 + systemd)
    ↓
Your Layer: Custom application
    ↓
Export: Dockerfile → ext4 image
    ↓
Ready: rootfs.ext4 + vmlinux for Firecracker
```

### Create Custom Image

Create `images/myapp/Dockerfile`:

```dockerfile
# Start from NanoFuse base (once published)
FROM ghcr.io/jpoley/nanofuse/base:latest

# Set environment
ENV DEBIAN_FRONTEND=noninteractive
ENV APP_HOME=/opt/myapp

# Install application dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        python3 \
        python3-pip \
        git \
        build-essential && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create app directory
RUN mkdir -p $APP_HOME && \
    cd $APP_HOME

# Copy application code (from local or git clone)
# COPY app/ $APP_HOME/
# OR
# RUN git clone https://github.com/myorg/myapp.git .

# Install application
# RUN pip install -r requirements.txt

# Create systemd service for app
RUN cat > /etc/systemd/system/myapp.service << 'EOF'
[Unit]
Description=My Application
After=network-online.target systemd-networkd.service
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/myapp
ExecStart=/usr/bin/python3 /opt/myapp/main.py
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable service (NEVER use 'start')
RUN systemctl enable myapp.service

# Expose ports if needed
EXPOSE 8000 8001

# Keep CMD from base image
```

### Build Custom Image

Create `images/myapp/build.sh`:

```bash
#!/bin/bash
set -euo pipefail

IMAGE_NAME="${IMAGE_NAME:-myapp}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ROOTFS_SIZE="${ROOTFS_SIZE:-4096}"  # 4GB for app

# Build Docker image
docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

# Step 2: Export
mkdir -p build/rootfs
CONTAINER_ID=$(docker create "${IMAGE_NAME}:${IMAGE_TAG}")
docker export "${CONTAINER_ID}" | tar -C build/rootfs -xf -
docker rm "${CONTAINER_ID}"

# Step 3: Create ext4
dd if=/dev/zero of=build/rootfs.ext4 bs=1M count="${ROOTFS_SIZE}" status=none
mkfs.ext4 -F -q -L myapp-root build/rootfs.ext4

# Step 4: Mount and copy
mkdir -p build/mnt
sudo mount -o loop build/rootfs.ext4 build/mnt
sudo cp -a build/rootfs/* build/mnt/
sudo umount build/mnt
sudo rmdir build/mnt
rm -rf build/rootfs
sudo chown $(id -u):$(id -g) build/rootfs.ext4

# Step 5: Download kernel (same as base)
curl -fsSL -o build/vmlinux \
    https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin

# Step 6: Manifest
cat > build/manifest.json << EOF
{
  "version": "0.1.0",
  "name": "${IMAGE_NAME}",
  "tag": "${IMAGE_TAG}",
  "architecture": "x86_64",
  "base_image": "ghcr.io/jpoley/nanofuse/base:latest",
  "services": {
    "ssh": {"enabled": true},
    "myapp": {"enabled": true}
  },
  "built_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "Build complete!"
ls -lh build/
```

Build it:

```bash
cd images/myapp
sudo bash build.sh

# Or with variables
IMAGE_TAG=v1.0.0 sudo bash build.sh
```

### Multi-Stage Builds (Optimized)

Keep final image small:

```dockerfile
# Stage 1: Build
FROM golang:1.21 AS builder
WORKDIR /build
COPY . .
RUN go build -o myapp .

# Stage 2: Runtime (includes NanoFuse base)
FROM ghcr.io/jpoley/nanofuse/base:latest

# Copy binary from build stage
COPY --from=builder /build/myapp /usr/local/bin/myapp

# Create service
RUN systemctl enable myapp.service

# Only app binary is in final image, not Go compiler
```

This keeps image small (just compiled binary, not build tools).

### Register Custom Image with NanoFuse

```bash
# Register locally
sudo ./bin/register-local-image \
  -name myapp:v1.0.0 \
  -kernel ./images/myapp/build/vmlinux \
  -rootfs ./images/myapp/build/rootfs.ext4

# List images
./bin/nanofuse image list

# Create VM
./bin/nanofuse vm create myapp:v1.0.0 test-app

# Start and test
./bin/nanofuse vm start test-app
./bin/nanofuse vm logs test-app --follow
```

---

## Troubleshooting

### Build Fails: "Permission denied"

**Error:**
```
mkfs.ext4: Permission denied
```

**Cause**: ext4 operations require root

**Solution**:
```bash
sudo make build
```

### Build Fails: "Docker build failed"

**Error:**
```
ERROR [build_stage 1/7]: failed to solve with frontend dockerfile.v0: rpc error...
```

**Causes**:
1. Docker daemon not running
2. Dockerfile syntax error
3. Internet connectivity issue

**Solutions**:
```bash
# Check Docker is running
sudo systemctl start docker
docker ps

# Check Dockerfile syntax
docker build --no-cache -t nanofuse-base:test .

# Check connectivity
curl -I https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
```

### Build Fails: "curl: (7) Failed to connect"

**Error**: Can't download kernel

**Cause**: Internet connectivity or URL changed

**Solution**:
```bash
# Verify kernel URL is accessible
curl -I https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin

# Download manually
mkdir -p images/base/build
curl -o images/base/build/vmlinux \
    https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin

# Check it's valid
file images/base/build/vmlinux
```

### Boot Test Fails: "No console output"

**Cause**: Kernel not uncompressed or rootfs invalid

**Solution**:
```bash
# Check kernel format
file images/base/build/vmlinux
# Should say: "Linux kernel" not "gzip compressed"

# Check rootfs format
file images/base/build/rootfs.ext4
# Should say: "ext4 filesystem data"

# Check file sizes are reasonable
ls -lh images/base/build/
# vmlinux should be ~10-15MB
# rootfs.ext4 should be ~2GB (sparse)
```

### Boot Test Fails: "systemd did not reach multi-user.target"

**Cause**: Systemd service failed to start

**Solution**:
```bash
# View full console log
cat /tmp/nanofuse-test-*/console.log | grep -i "failed\|error" | head -20

# Rebuild without Docker cache
sudo make clean
docker rmi nanofuse-base:latest
sudo make build

# Check systemd logs in container
docker run --rm nanofuse-base:latest journalctl -b 2>&1 | tail -50
```

### Boot Test Fails: "SSH daemon not running"

**Cause**: SSH service not enabled or failed to start

**Solution**:
```bash
# Check service is enabled in Dockerfile
docker run --rm nanofuse-base:latest systemctl is-enabled ssh.service

# Check service can start
docker run --rm nanofuse-base:latest systemctl status ssh.service || true

# Check sshd binary exists
docker run --rm nanofuse-base:latest which sshd

# Check SSH config is valid
docker run --rm nanofuse-base:latest sshd -T
```

### Boot Test Fails: "Network not configured"

**Cause**: systemd-networkd not running or config missing

**Solution**:
```bash
# Check networkd is enabled
docker run --rm nanofuse-base:latest systemctl is-enabled systemd-networkd.service

# Check network config exists
docker run --rm nanofuse-base:latest cat /etc/systemd/network/20-wired.network

# Check config is valid
docker run --rm nanofuse-base:latest networkctl status

# Check DHCP availability (in running VM)
./bin/nanofuse vm exec test-vm systemctl status systemd-networkd
./bin/nanofuse vm exec test-vm networkctl status
```

### Boot Time > 5 seconds

**Cause**: Slow services or system initialization

**Solution**:
```bash
# Analyze systemd startup time
./bin/nanofuse vm exec test-vm systemd-analyze

# See slowest units
./bin/nanofuse vm exec test-vm systemd-analyze critical-chain

# Profile boot sequence
./bin/nanofuse vm exec test-vm systemd-analyze plot > /tmp/boot.svg

# Disable slow services
# Edit Dockerfile to mask services that aren't needed
RUN systemctl mask systemd-resolved.service
```

### Can't SSH: "Connection refused"

**Cause**: SSH daemon not running or listening on port 22

**Solution**:
```bash
# Check SSH is running
./bin/nanofuse vm exec test-vm systemctl status ssh

# Check it's listening
./bin/nanofuse vm exec test-vm ss -tulpn | grep 22

# Check sshd config
./bin/nanofuse vm exec test-vm sshd -T | grep -i "port\|listen"

# Restart SSH
./bin/nanofuse vm exec test-vm systemctl restart ssh

# Check logs
./bin/nanofuse vm exec test-vm journalctl -u ssh -n 20
```

### Can't SSH: "Permission denied (publickey)"

**Cause**: SSH key not in authorized_keys or wrong permissions

**Solution**:
```bash
# Check authorized_keys exists and has content
./bin/nanofuse vm exec test-vm ls -la /root/.ssh/authorized_keys
./bin/nanofuse vm exec test-vm cat /root/.ssh/authorized_keys

# Check permissions are correct
./bin/nanofuse vm exec test-vm ls -la /root/.ssh/
# .ssh/ should be 700 (-rwx------)
# authorized_keys should be 600 (-rw-------)

# Add key manually (if VM is accessible)
./bin/nanofuse vm exec test-vm "echo 'ssh-rsa AAAA...' >> /root/.ssh/authorized_keys"

# Or rebuild image with your key in Dockerfile
```

### ext4 Image Corruption

**Symptoms**: `fsck` errors, mount failures

**Cause**: Interrupted build, bad copy, or filesystem errors

**Solution**:
```bash
# Check filesystem integrity
sudo fsck.ext4 -n images/base/build/rootfs.ext4

# If corrupted, rebuild
sudo make clean
sudo make build

# Verify after rebuild
sudo fsck.ext4 -n images/base/build/rootfs.ext4
```

### Out of Disk Space

**Error**: "No space left on device"

**Cause**: 2GB ext4 image uses significant space

**Solution**:
```bash
# Check available space
df -h | grep -E "/$|/home"

# Clean Docker (remove unused images)
docker system prune -a

# Clean old builds
sudo make clean

# Use smaller rootfs size
sudo make build ROOTFS_SIZE=1024  # 1GB instead of 2GB
```

---

## Quick Reference

### Most Used Commands

```bash
# Build
cd images/base
sudo make build              # Build everything

# Test
sudo make validate          # Check artifacts
sudo make test              # Boot test

# Develop
make shell                  # Shell in Docker image
make inspect                # Show artifact details
sudo make clean             # Remove artifacts

# Debug
docker build -t test .      # Step-by-step build
docker run -it test bash    # Inspect image
```

### Key Files

```
images/base/
├── Dockerfile           # Image definition (modify this to customize)
├── Makefile            # Build automation
├── build.sh            # Build script details
├── test-boot.sh        # Boot test logic
├── validate-build.sh   # Artifact validation
├── units/
│   ├── firstboot.service
│   └── http-test-server.service
├── build/              # Output directory
│   ├── rootfs.ext4     # Filesystem
│   ├── vmlinux         # Kernel
│   └── manifest.json   # Metadata
└── README.md           # Full docs
```

### Environment Variables

```bash
# Build customization
IMAGE_NAME=my-base
IMAGE_TAG=v1.0.0
ROOTFS_SIZE=4096        # MB

# Usage
sudo make build IMAGE_NAME=myimg ROOTFS_SIZE=4096
```

---

For CLI/daemon development, see [DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md).

For complete documentation, see `images/base/README.md`.
