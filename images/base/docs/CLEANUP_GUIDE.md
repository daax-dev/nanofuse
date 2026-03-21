# Complete Clean Guide

This document explains the clean scripts and when to use them.

## Overview

Two clean scripts are provided to reset the build environment:

1. **`clean.sh`** - Standard clean (no sudo required)
2. **`clean-sudo.sh`** - Complete clean with sudo privileges

## What Gets Cleaned

### Both Scripts Clean:
- All nanofuse Docker images
- All nanofuse Docker containers
- All nanofuse kernel builder images
- Docker buildkit cache
- Build logs from `/tmp`
- Old kernel binaries from `/tmp`
- Build system temporary files
- All running nanofused/firecracker processes

### `clean-sudo.sh` Also Cleans (Root-Owned Files):
- `./build/` directory (owned by root from make)
- `/tmp/nanofuse/` directory (owned by root from daemon)
- Firecracker test directories

## Usage

### Standard Clean (Recommended for Development)
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
./clean.sh
```

This removes all Docker artifacts and logs. Some root-owned files may remain with a warning.

### Complete Clean (Root-Owned Files Removed)
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
sudo ./clean-sudo.sh
```

This removes everything including root-owned files and directories. Complete reset.

## What Remains After `clean.sh`

If you see warnings about these after running `clean.sh`, they're root-owned:

```
⚠ Build directory still exists after clean, removing...
  To remove: sudo rm -rf /path/to/build

⚠ Could not remove /tmp/nanofuse (permission denied)
  To remove: sudo rm -rf /tmp/nanofuse
```

These can be safely removed with sudo or run `clean-sudo.sh` instead.

## Complete Clean Workflow

For a guaranteed clean build from scratch:

```bash
#!/bin/bash
set -euo pipefail

IMAGES_BASE="/home/jpoley/src/_mine/nanofuse/images/base"

echo "Complete clean with sudo..."
sudo "$IMAGES_BASE/clean-sudo.sh"

echo ""
echo "Verifying clean state..."
docker images | grep nanofuse && echo "ERROR: Images still exist" || echo "✓ No nanofuse images"
[ -d /tmp/nanofuse ] && echo "ERROR: /tmp/nanofuse still exists" || echo "✓ /tmp/nanofuse cleaned"
[ -d "$IMAGES_BASE/build" ] && echo "ERROR: build dir still exists" || echo "✓ build dir cleaned"
pgrep -f nanofused && echo "ERROR: nanofused still running" || echo "✓ No nanofused running"

echo ""
echo "✓ System is clean and ready for fresh build!"
```

## Clean Script Features

Both scripts include:
- Color-coded logging (✓ for success, ⚠ for warning, ✗ for error)
- Process verification (kills any running daemons/VMs)
- Multiple Docker cleanup methods
- Build log removal
- Temporary file cleanup
- Final verification of clean state
- Comprehensive error handling

## Common Issues

### "Permission denied" when removing build directory
**Cause**: Directory owned by root from previous `sudo make` or daemon run
**Solution**: Run `clean-sudo.sh` with sudo, or manually clean:
```bash
sudo rm -rf /home/jpoley/src/_mine/nanofuse/images/base/build
```

### Docker images/containers persist
**Cause**: Docker daemon or containers still in use
**Solution**: The clean scripts handle this, but if issues persist:
```bash
docker ps -a                                    # See all containers
docker ps -a | grep nanofuse | awk '{print $1}' # Get nanofuse IDs
docker rm -f <container-id>                    # Remove containers
docker rmi -f <image-id>                       # Remove images
```

### /tmp/nanofuse directory won't delete
**Cause**: Root-owned by daemon process
**Solution**: Run with sudo or use `clean-sudo.sh`:
```bash
sudo rm -rf /tmp/nanofuse
```

## Pre-Build Checklist

After running clean, verify:

```bash
# Should show nothing
docker images | grep nanofuse
ls -la /home/jpoley/src/_mine/nanofuse/images/base/build 2>&1 | head -2
ls -la /tmp/nanofuse 2>&1 | head -2

# Should show no results
pgrep -f nanofused
pgrep -f firecracker
```

## Quick Commands

```bash
# Full clean with sudo
sudo /home/jpoley/src/_mine/nanofuse/images/base/clean-sudo.sh

# Just remove root-owned build directory
sudo rm -rf /home/jpoley/src/_mine/nanofuse/images/base/build

# Just remove root-owned /tmp/nanofuse
sudo rm -rf /tmp/nanofuse

# List all nanofuse Docker artifacts
docker images | grep nanofuse
docker ps -a | grep nanofuse

# Kill all nanofused processes
pkill -9 -f nanofused

# Remove all Docker images
docker rmi -f $(docker images -q)

# Complete Docker cleanup
docker system prune -f --volumes
docker buildx prune -f -a
```

## See Also

- `build.sh` - Main build script
- `Dockerfile` - Base image definition
- `Dockerfile.kernel` - Kernel builder
- `build-kernel-docker.sh` - Kernel build script
