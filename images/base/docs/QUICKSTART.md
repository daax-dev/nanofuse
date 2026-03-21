# Quick Start Guide - NanoFuse Base Image

This guide will get you from zero to a running microVM in under 5 minutes.

## Prerequisites

```bash
# Check prerequisites
docker --version        # Docker installed
sudo -v                 # Root access (for ext4 operations)
which firecracker       # Firecracker installed (optional, for testing)
```

### Install Firecracker (if not installed)

```bash
# Download latest release
ARCH="$(uname -m)"
release_url="https://github.com/firecracker-microvm/firecracker/releases"
latest=$(curl -fsSL ${release_url}/latest | cut -d'"' -f2 | rev | cut -d'/' -f1 | rev)
curl -L ${release_url}/download/${latest}/firecracker-${latest}-${ARCH}.tgz \
    | tar -xz

# Move to PATH
sudo mv release-${latest}-$(uname -m)/firecracker-${latest}-$(uname -m) \
    /usr/local/bin/firecracker
sudo chmod +x /usr/local/bin/firecracker

# Verify
firecracker --version
```

## Build the Image (3-4 minutes)

```bash
# Navigate to base image directory
cd /home/jpoley/src/_mine/nanofuse/images/base

# Build everything (with sudo for ext4 operations)
sudo make build

# Expected output:
# ✓ Docker image built successfully
# ✓ Rootfs extracted: ./build/rootfs.ext4
# ✓ Kernel extracted: ./build/vmlinux
# ✓ Manifest generated: ./build/manifest.json
```

## Validate Build (30 seconds)

```bash
# Validate artifacts
sudo make validate

# Expected output:
# ✓ Rootfs file exists
# ✓ Rootfs is valid ext4 filesystem
# ✓ Kernel file exists
# ✓ Kernel is valid Linux kernel
# ✓ Manifest is valid JSON
# ✓ All validations passed!
```

## Test Boot in Firecracker (< 2 seconds)

```bash
# Test boot (requires firecracker and /dev/kvm access)
sudo make test

# Expected output:
# ✓ VM booted successfully in 1s
# ✓ Test 1: VM boots successfully
# ✓ Test 2: Console output visible on ttyS0
# ✓ Test 3: systemd reaches multi-user.target
# ✓ Test 4: SSH daemon running
# ✓ Test 5: Network configured (systemd-networkd)
# ✓ Test 6: No failed systemd units detected
# ✓ Test 7: Boot time < 2s
# Overall: PASS
```

## Inspect Build Artifacts

```bash
# Show detailed information
sudo make inspect

# Output shows:
# - Rootfs size and format
# - Kernel size and type
# - Manifest contents (JSON)
```

## Manual Testing (Advanced)

### Start VM Manually

```bash
# Create a simple VM config
cat > vm-config.json <<EOF
{
  "boot-source": {
    "kernel_image_path": "./build/vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "./build/rootfs.ext4",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 2,
    "mem_size_mib": 512
  }
}
EOF

# Start Firecracker
sudo firecracker \
    --api-sock /tmp/firecracker.sock \
    --config-file vm-config.json

# Watch console output (in same terminal)
# You should see Linux kernel boot messages and systemd startup
# VM will boot to login prompt

# To stop: Ctrl+C or send shutdown via API
```

### View Console Log

```bash
# After running 'make test', view full console output
cat /tmp/nanofuse-test-*/console.log | less

# Look for:
# - Linux kernel boot messages
# - systemd startup
# - "Reached target Multi-User System"
# - SSH daemon started
# - Network configured
```

## Common Issues

### Issue: "Permission denied" when building

**Solution**: Run with sudo (required for ext4 operations)
```bash
sudo make build
```

### Issue: "firecracker binary not found"

**Solution**: Install Firecracker (see Prerequisites above) or skip test
```bash
# Build without testing
sudo make build validate
```

### Issue: "/dev/kvm: Permission denied"

**Solution**: Add user to kvm group or run with sudo
```bash
# Add user to kvm group
sudo usermod -aG kvm $USER
newgrp kvm

# Or run with sudo
sudo make test
```

### Issue: Build is very slow

**Solution**: First build downloads packages. Subsequent builds use Docker cache
```bash
# First build: ~3-4 minutes
# Cached builds: ~1-2 minutes

# Force rebuild from scratch
sudo make clean
sudo make build
```

## Next Steps

### For Development

```bash
# Interactive shell in Docker image (before VM conversion)
make shell

# Clean and rebuild
sudo make clean
sudo make build

# Customize rootfs size
sudo make build ROOTFS_SIZE=4G

# Customize image name/tag
sudo make build IMAGE_NAME=my-base IMAGE_TAG=v1.0.0
```

### For Production

```bash
# Push to GitHub Container Registry (GHCR)
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
sudo make push REGISTRY=ghcr.io/USERNAME

# Image will be available as:
# ghcr.io/USERNAME/nanofuse-base:latest
```

### For Integration

The build artifacts are ready to use with the NanoFuse API (once implemented):

**Artifact Locations**:
- Rootfs: `./build/rootfs.ext4`
- Kernel: `./build/vmlinux`
- Manifest: `./build/manifest.json`

**API Integration** (future):
```bash
# When nanofuse CLI is ready
nanofuse image pull ghcr.io/jpoley/nanofuse-base:latest
nanofuse vm create nanofuse-base:latest my-vm
nanofuse vm start my-vm
```

## What's Inside the Image?

```bash
# View rootfs contents (requires root)
sudo make inspect

# Or manually mount and explore
sudo mkdir -p /tmp/rootfs-inspect
sudo mount -o loop,ro ./build/rootfs.ext4 /tmp/rootfs-inspect
ls -la /tmp/rootfs-inspect
sudo umount /tmp/rootfs-inspect
```

**Included**:
- Ubuntu 24.04 base system
- systemd (init system)
- OpenSSH server
- systemd-networkd (DHCP networking)
- Firecracker-compatible kernel (5.10.204)
- Serial console on ttyS0
- First-boot initialization service

**NOT included** (minimal image):
- GUI/desktop environment
- Python, Node.js, etc. (add in extended images)
- Package manager caches (cleaned during build)

## Performance Metrics

Based on testing:

- **Build time**: 3-4 minutes (first), 1-2 minutes (cached)
- **Boot time**: < 2 seconds to multi-user.target
- **Image size**: ~2GB (rootfs.ext4)
- **Memory usage**: ~512MB default (configurable)
- **Disk usage**: ~500-800MB actual (in 2GB sparse file)

## Help and Troubleshooting

```bash
# Show all make targets
make help

# View detailed build output
sudo make build 2>&1 | tee build.log

# Validate without rebuilding
sudo make validate

# Clean all artifacts
sudo make clean
```

**For more information**:
- See `README.md` for complete documentation
- See `NOTES.md` for implementation details
- See `test-boot.sh` for testing logic
- See `validate-build.sh` for validation logic

## Success Criteria

You know the build is successful when:

1. ✅ `make build` completes without errors
2. ✅ `make validate` reports "All validations passed"
3. ✅ `make test` reports "Overall: PASS"
4. ✅ Boot time is < 2 seconds
5. ✅ No failed systemd units in console log

If all criteria pass, your base image is ready for use!
