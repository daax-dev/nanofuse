# NanoFuse Base MicroVM Image

This is the base Ubuntu 24.04 microVM image for NanoFuse, designed to run in Firecracker with systemd, SSH, and networking support.

## Quick Start

- **Building**: See [BUILD.md](BUILD.md) for build instructions
- **Testing**: See [TEST.md](TEST.md) for boot testing with Firecracker
- **Full Quickstart**: See [docs/QUICKSTART.md](docs/QUICKSTART.md) for 5-minute setup
- **Details**: See [docs/IMPLEMENTATION_NOTES.md](docs/IMPLEMENTATION_NOTES.md) for design decisions

## Overview

The base image provides:

- **Base OS**: Ubuntu 24.04 LTS
- **Init System**: systemd (multi-user.target)
- **Kernel**: 5.10.204 (official Firecracker CI kernel)
- **Networking**: systemd-networkd with DHCP
- **Remote Access**: OpenSSH server
- **Console**: Serial console on ttyS0 (Firecracker compatible)

## Architecture

The image is built using a hybrid approach that combines learning and practicality:

1. Start FROM ubuntu:24.04 (learning: build from scratch)
2. Install systemd, openssh-server, networking packages
3. Bundle official Firecracker CI kernel 5.10.204 (practical: use proven kernel)
4. Follow Slicer's Dockerfile best practices (systemctl enable, not start)
5. Export to rootfs.ext4 block device for Firecracker

## Build Requirements

- Docker (for building container image)
- Linux host with KVM support (for testing)
- Firecracker binary (for testing)
- Root access (for creating ext4 images)

## Building the Image

### Quick Build

```bash
# Using Make (calls build.sh)
make build

# Or directly with build.sh
sudo ./build.sh
```

This will:
1. Build Docker image from Dockerfile
2. Export container filesystem
3. Create ext4 filesystem image (2GB)
4. Mount and copy files to ext4
5. Download Firecracker kernel
6. Generate manifest.json with metadata

### Build Artifacts

After successful build, you'll have:

```
build/
├── rootfs.ext4       # Root filesystem (2GB ext4)
├── vmlinux           # Uncompressed kernel (5.10.240)
└── manifest.json     # Image metadata
```

### Customizing Build

```bash
# Specify custom image name and tag
make build IMAGE_NAME=my-base IMAGE_TAG=v1.0.0

# Change rootfs size (default: 2G)
make build ROOTFS_SIZE=4G
```

## Testing the Image

### Boot Test in Firecracker

```bash
# Run automated boot test
make test
```

This script tests that the image:
- ✅ Builds without interactive prompts
- ✅ Boots in Firecracker in <2 seconds
- ✅ Console output visible on ttyS0
- ✅ systemd reaches multi-user.target
- ✅ SSH daemon running and accessible
- ✅ Network configured via systemd-networkd (DHCP)
- ✅ All systemd units started successfully (no failed units)

### Manual Testing

```bash
# Start Firecracker manually
sudo firecracker \
  --api-sock /tmp/firecracker.sock \
  --config-file vm-config.json

# vm-config.json example:
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
```

## Image Structure

### Filesystem Layout

```
/
├── boot/
│   ├── vmlinux-5.10.204    # Firecracker CI kernel
│   └── vmlinux -> vmlinux-5.10.204
├── etc/
│   ├── systemd/
│   │   ├── system/
│   │   │   └── firstboot.service    # First-boot initialization
│   │   └── network/
│   │       └── 20-wired.network     # DHCP network config
│   └── ssh/
│       └── sshd_config              # SSH configuration
├── var/
│   └── log/
│       └── nanofuse/
│           └── firstboot.log        # First boot logs
└── manifest.json                    # Image metadata
```

### Enabled Services

The following systemd services are enabled (not started during build):

- `serial-getty@ttyS0.service` - Serial console
- `ssh.service` - OpenSSH server
- `systemd-networkd.service` - Network management
- `firstboot.service` - First-boot initialization

### Kernel Configuration

- **Version**: 5.10.204
- **Source**: Official Firecracker CI (proven compatibility)
- **cmdline**: `console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k`
- **Format**: Uncompressed (vmlinux, not bzImage)

## Usage with NanoFuse API

Once the NanoFuse API is implemented, this image can be used as:

```bash
# Pull image (future)
nanofuse image pull ghcr.io/jpoley/nanofuse-base:latest

# Create VM from image (future)
nanofuse vm create ghcr.io/jpoley/nanofuse-base:latest my-vm

# Start VM (future)
nanofuse vm start my-vm
```

## Publishing to GHCR

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Push image
make push REGISTRY=ghcr.io/jpoley
```

## Critical Build Constraints

These constraints MUST be followed when modifying the Dockerfile:

1. **NEVER use `systemctl start`** during Docker build
   - Services are not running in container context
   - Use `systemctl enable` to configure for VM boot

2. **ALWAYS use `systemctl enable`** for services
   - This configures systemd to start services when VM boots
   - Example: `RUN systemctl enable ssh.service`

3. **Set `DEBIAN_FRONTEND=noninteractive`** for apt operations
   - Prevents interactive prompts during build
   - Required for automated builds

4. **Console MUST be on ttyS0** for Firecracker
   - Firecracker only supports serial console
   - Enable with: `systemctl enable serial-getty@ttyS0.service`

5. **Keep CMD unchanged** (expects systemd)
   - CMD should be `/sbin/init` or equivalent
   - Firecracker boots kernel directly, not container runtime

## Troubleshooting

### Build Issues

**Error: "Unable to download kernel"**
```bash
# Verify kernel URL is accessible
curl -I https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.7/x86_64/vmlinux-5.10.204
```

**Error: "mkfs.ext4: Permission denied"**
```bash
# Run with sudo
sudo make build
```

### Boot Issues

**VM fails to boot / No console output**
- Check kernel is uncompressed: `file build/vmlinux` should show "Linux kernel"
- Verify rootfs.ext4 is valid: `file build/rootfs.ext4` should show "ext4 filesystem"
- Check Firecracker logs for error messages

**"systemd did not reach multi-user.target"**
- Review console log: `tail -f /tmp/nanofuse-test-*/console.log`
- Check for failed systemd units in console output
- Verify systemd services were enabled, not started during build

**Network not working**
- Ensure TAP device configured on host
- Check systemd-networkd status in console log
- Verify DHCP server available on network

## Development

### Interactive Shell

```bash
# Start shell in Docker image (not VM)
make shell
```

### Inspect Build Artifacts

```bash
# Show detailed information about build artifacts
make inspect
```

### Clean Build

```bash
# Remove all build artifacts
make clean

# Rebuild from scratch
make clean build
```

## Next Steps

1. **Integration with API**: This image will be used by the nanofused API daemon
2. **Testing**: API agent will use this for VM lifecycle testing
3. **Extensions**: Build Trigger.dev web and worker images extending this base
4. **CI/CD**: Automate builds and publishing via GitHub Actions

## References

- [Firecracker Documentation](https://github.com/firecracker-microvm/firecracker)
- [Slicer Project](https://github.com/openfaasltd/slicer)
- [Ubuntu systemd Documentation](https://ubuntu.com/server/docs/service-management-systemd)
- [NanoFuse Architecture Decisions](../../docs/ARCHITECTURE_DECISIONS.md)

## Coordination with Other Agents

This image is being built in **parallel** with:

- **CLI agent** (building nanofuse CLI tool)
- **API agent** (building nanofused daemon)
- **CI/CD agent** (building GitHub Actions workflow)

The API agent will reference this image for local testing at:
- **Rootfs**: `/home/jpoley/src/_mine/nanofuse/images/base/build/rootfs.ext4`
- **Kernel**: `/home/jpoley/src/_mine/nanofuse/images/base/build/vmlinux`
- **Manifest**: `/home/jpoley/src/_mine/nanofuse/images/base/build/manifest.json`
