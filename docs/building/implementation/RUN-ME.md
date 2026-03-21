# NanoFuse - Complete Run Guide

**How to build, run, and test the entire NanoFuse system end-to-end.**

---

## Prerequisites

- Go 1.22+
- Docker
- Linux with KVM support
- Firecracker binary (optional, for VM testing)
- sudo access (for building images)

---

## Quick Start (5 Minutes)

### 1. Build Everything

```bash
# Build Go binaries (CLI + API daemon)
go install github.com/magefile/mage@latest
mage all

# Build base Firecracker image
cd images/base
sudo ./build.sh
cd ../..
```

### 2. Start API Daemon

```bash
# Create config directory
mkdir -p ~/.config/nanofuse

# Create basic config
cat > ~/.config/nanofuse/config.yaml << 'EOF'
api:
  address: "127.0.0.1:8080"

storage:
  data_dir: "/tmp/nanofuse"
  images_dir: "/tmp/nanofuse/images"
  vms_dir: "/tmp/nanofuse/vms"

firecracker:
  binary_path: "/usr/bin/firecracker"
  socket_dir: "/tmp/nanofuse/sockets"
EOF

# Start daemon (in background or separate terminal)
./bin/nanofused --config ~/.config/nanofuse/config.yaml
```

### 3. Use CLI

```bash
# Check daemon health
./bin/nanofuse health

# List images (will be empty)
./bin/nanofuse image list

# Check version
./bin/nanofuse version
```

---

## Full End-to-End Workflow

### Step 1: Build the Base Image

```bash
cd images/base

# Clean any previous builds
sudo rm -rf build/
docker rmi nanofuse-base:latest 2>/dev/null || true

# Build from scratch
sudo ./build.sh

# Verify artifacts
ls -lh build/
# Should show:
# - rootfs.ext4 (2GB)
# - vmlinux (21MB)
# - manifest.json (metadata)

# Verify file types
file build/rootfs.ext4  # Should be: Linux ext4 filesystem
file build/vmlinux      # Should be: ELF 64-bit LSB executable

cd ../..
```

**Expected output:**
```
========================================
NanoFuse Base Image Build
========================================

[1/6] Building Docker image...
✓ Docker image built: nanofuse-base:latest
[2/6] Exporting container filesystem...
✓ Container created: 7480fa21f2db
✓ Filesystem exported
✓ Container cleaned up
[3/6] Creating ext4 filesystem image...
✓ Created 2048MB ext4 image
[4/6] Copying filesystem to ext4 image...
✓ Mounted ext4 image
✓ Files copied to ext4 image
✓ Unmounted and cleaned up
[5/6] Downloading Firecracker kernel...
✓ Kernel downloaded: 21M
[6/6] Generating manifest.json...
✓ Manifest generated

========================================
Build Complete!
========================================
```

### Step 2: Build Go Binaries

```bash
# Install Mage if not already installed
go install github.com/magefile/mage@latest

# Build both binaries
mage all

# Verify builds
ls -lh bin/
# Should show:
# - nanofuse (CLI, ~13MB)
# - nanofused (daemon, ~15MB)

# Test CLI
./bin/nanofuse version
```

**Expected output:**
```
Building nanofuse CLI...
✓ Built: bin/nanofuse (13.2 MB)

Building nanofused daemon...
✓ Built: bin/nanofused (14.8 MB)

Build complete!
```

### Step 3: Configure the System

```bash
# Create config directory
mkdir -p ~/.config/nanofuse
mkdir -p /tmp/nanofuse/{images,vms,snapshots,sockets}

# Create configuration file
cat > ~/.config/nanofuse/config.yaml << 'EOF'
api:
  address: "127.0.0.1:8080"
  log_level: "info"

storage:
  data_dir: "/tmp/nanofuse"
  images_dir: "/tmp/nanofuse/images"
  vms_dir: "/tmp/nanofuse/vms"
  snapshots_dir: "/tmp/nanofuse/snapshots"

firecracker:
  binary_path: "/usr/bin/firecracker"
  socket_dir: "/tmp/nanofuse/sockets"
  default_vcpus: 2
  default_memory_mb: 512

registry:
  default_registry: "ghcr.io"
  cache_dir: "/tmp/nanofuse/registry-cache"
EOF

# Verify config
cat ~/.config/nanofuse/config.yaml
```

### Step 4: Copy Local Image to Storage

Since we haven't set up GHCR authentication yet, let's use the local image:

```bash
# Copy built image to API storage directory
mkdir -p /tmp/nanofuse/images/nanofuse-base/latest
cp images/base/build/rootfs.ext4 /tmp/nanofuse/images/nanofuse-base/latest/
cp images/base/build/vmlinux /tmp/nanofuse/images/nanofuse-base/latest/
cp images/base/build/manifest.json /tmp/nanofuse/images/nanofuse-base/latest/

# Verify
ls -lh /tmp/nanofuse/images/nanofuse-base/latest/
```

### Step 5: Start the API Daemon

**Option A: Foreground (for debugging)**
```bash
./bin/nanofused --config ~/.config/nanofuse/config.yaml
```

**Option B: Background**
```bash
nohup ./bin/nanofused --config ~/.config/nanofuse/config.yaml > /tmp/nanofuse/daemon.log 2>&1 &
echo $! > /tmp/nanofuse/daemon.pid

# View logs
tail -f /tmp/nanofuse/daemon.log
```

**Expected output:**
```
INFO[0000] Starting NanoFuse API daemon
INFO[0000] API server listening on 127.0.0.1:8080
INFO[0000] Storage initialized: /tmp/nanofuse
INFO[0000] Ready to accept requests
```

### Step 6: Test with CLI

In a new terminal:

```bash
# Set API endpoint (if different from default)
export NANOFUSE_API_URL="http://127.0.0.1:8080"

# Check daemon health
./bin/nanofuse health

# List images
./bin/nanofuse image list

# Show image details
./bin/nanofuse image inspect nanofuse-base:latest

# List VMs (should be empty)
./bin/nanofuse vm list

# Create a VM
./bin/nanofuse vm create my-first-vm \
  --image nanofuse-base:latest \
  --vcpus 2 \
  --memory 512

# List VMs again
./bin/nanofuse vm list

# Get VM details
./bin/nanofuse vm inspect my-first-vm

# Start the VM (if Firecracker is installed)
./bin/nanofuse vm start my-first-vm

# Check VM status
./bin/nanofuse vm status my-first-vm

# View VM console output
./bin/nanofuse vm logs my-first-vm

# Stop the VM
./bin/nanofuse vm stop my-first-vm

# Delete the VM
./bin/nanofuse vm delete my-first-vm
```

### Step 7: Advanced Operations

#### Create a Snapshot
```bash
./bin/nanofuse vm create test-vm --image nanofuse-base:latest
./bin/nanofuse vm start test-vm

# Create snapshot
./bin/nanofuse snapshot create test-vm my-snapshot

# List snapshots
./bin/nanofuse snapshot list test-vm

# Stop VM and resume from snapshot
./bin/nanofuse vm stop test-vm
./bin/nanofuse vm resume test-vm --snapshot my-snapshot
```

#### Pull Image from Registry (requires GHCR auth)
```bash
# Set GitHub token
export GITHUB_TOKEN="your_token_here"

# Pull image
./bin/nanofuse image pull ghcr.io/jpoley/nanofuse-base:latest

# List pulled images
./bin/nanofuse image list
```

---

## CI/CD Build (GitHub Actions)

The CI pipeline automatically builds everything on push:

```bash
# Push to trigger CI
git add .
git commit -m "feat: update image build"
git push origin main

# CI will:
# 1. Build Go binaries
# 2. Run tests
# 3. Build Docker image
# 4. Push to GHCR
# 5. Create GitHub release (on tags)
```

View workflow runs:
```bash
gh run list
gh run watch
```

---

## Testing the Build

### Test 1: Validate Artifacts

```bash
cd images/base
./validate-build.sh ./build

# Should show:
# ✓ Rootfs file exists
# ✓ Kernel file exists
# ✓ Manifest file exists
# ✓ Docker image exists
```

### Test 2: Unit Tests

```bash
# Run all Go tests
mage test

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test 3: Integration Tests

```bash
# Run integration tests (requires daemon running)
go test -tags=integration ./test/integration/...
```

### Test 4: Boot Test (requires Firecracker)

```bash
cd images/base
sudo ./test-boot.sh build/vmlinux build/rootfs.ext4

# Should boot VM and verify:
# - Console output appears
# - systemd reaches multi-user.target
# - SSH daemon is running
```

---

## Troubleshooting

### Build Issues

**"permission denied" when building image**
```bash
# Need sudo for mount operations
sudo ./images/base/build.sh
```

**"Docker image exists"**
```bash
# Clean and rebuild
cd images/base
sudo rm -rf build/
docker rmi nanofuse-base:latest
sudo ./build.sh
```

### Daemon Issues

**"connection refused"**
```bash
# Check daemon is running
curl http://127.0.0.1:8080/health

# Check logs
cat /tmp/nanofuse/daemon.log

# Restart daemon
pkill nanofused
./bin/nanofused --config ~/.config/nanofuse/config.yaml
```

**"firecracker binary not found"**
```bash
# Install Firecracker
curl -LO https://github.com/firecracker-microvm/firecracker/releases/download/v1.7.0/firecracker-v1.7.0-x86_64.tgz
tar xzf firecracker-v1.7.0-x86_64.tgz
sudo cp release-v1.7.0-x86_64/firecracker-v1.7.0-x86_64 /usr/bin/firecracker
sudo chmod +x /usr/bin/firecracker

# Update config with correct path
vim ~/.config/nanofuse/config.yaml
```

### CLI Issues

**"command not found"**
```bash
# Use relative path
./bin/nanofuse version

# Or add to PATH
export PATH=$PATH:$(pwd)/bin
nanofuse version
```

**"API connection failed"**
```bash
# Set API URL explicitly
export NANOFUSE_API_URL="http://127.0.0.1:8080"
./bin/nanofuse health
```

---

## Project Structure

```
nanofuse/
├── bin/                    # Built binaries
│   ├── nanofuse           # CLI tool
│   └── nanofused          # API daemon
├── cmd/                   # Go entry points
│   ├── nanofuse/         # CLI source
│   └── nanofused/        # Daemon source
├── images/               # Image definitions
│   └── base/            # Base Ubuntu image
│       ├── Dockerfile
│       ├── build.sh     # ← Build script
│       └── build/       # Build artifacts
│           ├── rootfs.ext4
│           ├── vmlinux
│           └── manifest.json
├── internal/            # Go packages
├── test/               # Tests
└── RUN-ME.md          # ← This file
```

---

## Development Workflow

```bash
# 1. Make changes to code
vim cmd/nanofuse/main.go

# 2. Rebuild
mage all

# 3. Test
mage test

# 4. Run locally
./bin/nanofused --config config.yaml

# 5. Test with CLI
./bin/nanofuse health

# 6. Commit and push
git add .
git commit -m "feat: add new feature"
git push
```

---

## What Works Right Now

✅ Building Go binaries (CLI + daemon)
✅ Building Firecracker base image
✅ CLI commands (help, version, health)
✅ API endpoints (health, images, VMs)
✅ Unit tests
✅ Docker image build
✅ CI/CD pipeline

## What Needs Firecracker Installed

⚠️ Starting VMs
⚠️ VM snapshots
⚠️ VM resume
⚠️ Boot testing

## What Needs GHCR Auth

⚠️ Pulling images from registry
⚠️ Pushing images to registry

---

## Next Steps

1. Install Firecracker: https://github.com/firecracker-microvm/firecracker/releases
2. Set up GHCR auth: `echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin`
3. Test full VM lifecycle: create → start → stop → delete
4. Build custom images extending base image

---

**Questions? Issues?**
- Check logs: `/tmp/nanofuse/daemon.log`
- Read docs: `docs/`
- File issue: https://github.com/jpoley/nanofuse/issues
