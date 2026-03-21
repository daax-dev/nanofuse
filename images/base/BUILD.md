# Building the NanoFuse Base Image

This guide covers building the NanoFuse Firecracker base image.

## Quick Build

```bash
cd /home/jpoley/src/_mine/nanofuse/images/base

# Build the image (requires sudo for ext4 operations)
sudo ./build.sh
```

## Build Artifacts

After successful build, you'll have:

```
build/
├── rootfs.ext4       # 2GB ext4 root filesystem
├── vmlinux           # Linux kernel (6.1.90)
└── manifest.json     # Build metadata
```

## What Gets Built

### 1. Docker Image (ubuntu:24.04)
- Base: Ubuntu 24.04 LTS
- Includes: systemd, SSH, networking
- Optimized: minimal, no package cache

### 2. Root Filesystem (rootfs.ext4)
- Format: ext4 filesystem (2GB)
- Bootable with Firecracker
- Includes all services pre-configured

### 3. Kernel (vmlinux)
- Version: Linux 6.1.90
- Format: Uncompressed ELF binary (required for Firecracker)
- Config: Firecracker's official microVM configuration
- Size: ~39MB

### 4. Manifest (manifest.json)
- Build metadata in JSON format
- Contains kernel info, filesystem metadata
- Used by API for VM configuration

## Build Process Steps

```
1. Docker build      → Ubuntu 24.04 image
2. Export container  → Extract filesystem
3. Create ext4       → 2GB filesystem image
4. Copy files        → Write to ext4 image
5. Build/get kernel  → Linux 6.1.90
6. Generate manifest → Build metadata JSON
```

## Build Options

```bash
# Custom image name/tag
IMAGE_NAME=custom IMAGE_TAG=v1.0 sudo ./build.sh

# Custom rootfs size (MB)
ROOTFS_SIZE=4096 sudo ./build.sh

# Custom architecture
ARCHITECTURE=arm64 sudo ./build.sh
```

## Using the Artifacts

Once built, use with the NanoFuse API:

```bash
# The API will reference these paths:
# Rootfs: ./build/rootfs.ext4
# Kernel: ./build/vmlinux
# Manifest: ./build/manifest.json
```

## Troubleshooting

### Build fails with "Permission denied"
**Solution:** Run with sudo
```bash
sudo ./build.sh
```

### Docker build is very slow
**Solution:** First build downloads packages. Subsequent builds use Docker cache.
```bash
# Force fresh build
sudo ./clean-sudo.sh
sudo ./build.sh
```

### Kernel build fails
**Solution:** Clear Docker cache and rebuild
```bash
docker system prune -f --all
sudo ./build.sh
```

## Cleaning Up

Remove all build artifacts:

```bash
# Remove user-writable files
./clean.sh

# Or remove everything including root-owned files
sudo ./clean-sudo.sh
```

## Related Documentation

- **README.md** - Full project documentation
- **TEST.md** - Boot testing with Firecracker
- **docs/IMPLEMENTATION_NOTES.md** - Design decisions
- **docs/QUICKSTART.md** - 5-minute quick start
- **docs/CLEANUP_GUIDE.md** - Detailed cleanup information
