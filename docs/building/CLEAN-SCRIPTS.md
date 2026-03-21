# Complete Clean Scripts Guide

Comprehensive guide to all clean and check scripts for the NanoFuse build system.

## Overview

Four cleanup scripts handle different levels of cleaning:

| Script | Location | Scope | Sudo | Purpose |
|--------|----------|-------|------|---------|
| `clean-all.sh` | Root (`./`) | Everything | Sometimes* | Master cleanup for complete reset |
| `clean-sudo.sh` | `images/base/` | Image build | Yes | Complete clean with root files |
| `clean.sh` | `images/base/` | Image build | No | Standard clean (user-writable) |
| `check-clean.sh` | `images/base/` | Image build | No | Check what needs cleaning |

*`clean-all.sh` auto-detects if sudo is needed

## Quick Reference

### Just Check
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
./check-clean.sh
```

### Standard Clean (No Sudo)
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
./clean.sh
```

### Complete Clean (All Files)
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
sudo ./clean-sudo.sh
```

### Master Clean from Anywhere
```bash
sudo /home/jpoley/src/_mine/nanofuse/clean-all.sh
```

## Detailed Scripts

### 1. `check-clean.sh` (images/base/)

**Purpose**: Check what artifacts exist without removing anything

**Usage**:
```bash
./check-clean.sh
```

**Checks**:
- Docker images (nanofuse)
- Docker containers
- Build directory
- /tmp/nanofuse data
- Running processes
- Temporary kernel binaries

**Output**: Color-coded list with counts

**Example**:
```
✓ No nanofuse Docker images
⚠ Build directory exists: 2.0G
⚠ /tmp/nanofuse directory exists: 392K
✓ No running processes
⚠ System has 3 artifact(s) to clean
```

### 2. `clean.sh` (images/base/)

**Purpose**: Standard build directory clean (user-writable files only)

**Usage**:
```bash
./clean.sh
```

**Cleans**:
- All nanofuse Docker images
- All nanofuse Docker containers
- Docker buildkit cache
- Build logs from /tmp
- Old kernel binaries from /tmp
- Firecracker processes

**Does NOT Clean** (requires sudo):
- `./build/` directory (root-owned)
- `/tmp/nanofuse/` directory (root-owned)

**Output**: Step-by-step with status messages

**When to Use**:
- Daily development
- Before testing changes
- When Docker images get stale
- Quick cleanup without sudo

### 3. `clean-sudo.sh` (images/base/)

**Purpose**: Complete build directory clean including root-owned files

**Usage**:
```bash
sudo ./clean-sudo.sh
```

**Requires**: Root privileges (via sudo)

**Cleans**: Everything `clean.sh` does, plus:
- `./build/` directory (root-owned from make)
- `/tmp/nanofuse/` directory (root-owned from daemon)
- Firecracker test directories

**Output**: Comprehensive cleanup summary

**When to Use**:
- Before major build changes
- Complete system reset
- Troubleshooting build issues
- CI/CD pipelines
- Guaranteed clean state

### 4. `clean-all.sh` (Root directory)

**Purpose**: Master cleanup script for complete NanoFuse reset

**Usage**:
```bash
# Auto-detects if sudo needed
/home/jpoley/src/_mine/nanofuse/clean-all.sh

# Or force with sudo
sudo /home/jpoley/src/_mine/nanofuse/clean-all.sh
```

**Scope**: Entire NanoFuse repository

**Cleans**:
- All daemon processes (nanofused, firecracker)
- systemd service (if running)
- Docker images (all nanofuse)
- Docker containers (all nanofuse)
- Docker buildkit cache
- Build artifacts (entire `images/base/build/`)
- Temporary data (`/tmp/nanofuse/`, `/tmp/vmlinux-*`)
- Build logs
- Test scripts

**Auto-Detects**:
- Root-owned files that need sudo
- Requests sudo if necessary

**When to Use**:
- Complete system reset
- Before CI/CD runs
- Troubleshooting complex issues
- After major changes
- Cleanup before committing

## Cleanup Levels

### Level 1: Quick Clean (2 seconds)
```bash
./check-clean.sh    # Just see what's dirty
```

### Level 2: Standard Clean (10-30 seconds)
```bash
./clean.sh          # Remove Docker artifacts and logs
```

### Level 3: Complete Build Clean (30-60 seconds)
```bash
sudo ./clean-sudo.sh # Remove all build files
```

### Level 4: Full Reset (60-120 seconds)
```bash
sudo clean-all.sh   # Remove everything NanoFuse-related
```

## Common Workflows

### Development Build Cycle
```bash
# Quick check
./check-clean.sh

# Light clean
./clean.sh

# Build
sudo ./build.sh

# Test
./test-build-and-boot.sh
```

### Before Major Changes
```bash
# Complete clean
sudo ./clean-sudo.sh

# Verify clean
./check-clean.sh

# Rebuild from scratch
sudo ./build.sh
```

### Troubleshooting Build Issues
```bash
# Full system reset
sudo /home/jpoley/src/_mine/nanofuse/clean-all.sh

# Verify
cd /home/jpoley/src/_mine/nanofuse/images/base
./check-clean.sh

# Rebuild
sudo ./build.sh
```

### CI/CD Pipeline
```bash
# Start clean
sudo /home/jpoley/src/_mine/nanofuse/clean-all.sh

# Verify clean state
test ! -d /tmp/nanofuse
test ! -f /tmp/vmlinux-*
! docker images | grep nanofuse

# Build
cd /home/jpoley/src/_mine/nanofuse/images/base
sudo ./build.sh

# Test
cd /home/jpoley/src/_mine/nanofuse
./test-build-and-boot.sh
```

## What Gets Cleaned (Details)

### Docker Artifacts
- Images matching: `nanofuse*`, `kernel-builder`
- Containers matching: `nanofuse*`, `kernel-builder`
- BuildKit cache
- Dangling volumes

### Build Directory (`./images/base/build/`)
- `vmlinux` - Kernel binary
- `rootfs.ext4` - Filesystem image
- `manifest.json` - Build metadata
- `.mnt/` - Mount point

### Temporary Files
- `/tmp/vmlinux-*` - Kernel builds
- `/tmp/nanofuse/` - Daemon data
- `/tmp/fc-test-*` - Test directories
- `/tmp/*build*.log` - Build logs
- `/tmp/*test*.sh` - Test scripts

### Processes
- `nanofused` - NanoFuse daemon
- `firecracker` - VM runtime

### systemd
- `nanofused.service` - Stopped if running

## Troubleshooting Clean Scripts

### "Permission denied" errors

**Cause**: Trying to remove root-owned files without sudo

**Solution**:
```bash
# Use sudo version
sudo ./clean-sudo.sh

# Or full master clean
sudo /home/jpoley/src/_mine/nanofuse/clean-all.sh
```

### Script says "Some artifacts remain"

**Cause**: Normal when not using sudo

**Solution**: Either ignore (if not building) or:
```bash
sudo ./clean-sudo.sh
```

### Clean runs but files still exist

**Cause**: File permissions or Docker in use

**Solution**:
```bash
# Force kill everything
sudo pkill -9 -f nanofused
sudo pkill -9 -f firecracker
docker system prune -f --volumes

# Manual removal
sudo rm -rf /tmp/nanofuse
sudo rm -rf /home/jpoley/src/_mine/nanofuse/images/base/build
```

## Manual Cleanup Commands

If scripts don't work, use these commands:

```bash
# Stop processes
sudo pkill -9 -f nanofused
sudo pkill -9 -f firecracker

# Remove Docker artifacts
docker rmi -f $(docker images | grep nanofuse | awk '{print $3}')
docker rm -f $(docker ps -a | grep nanofuse | awk '{print $1}')
docker system prune -f --volumes

# Remove build artifacts
sudo rm -rf /home/jpoley/src/_mine/nanofuse/images/base/build
sudo rm -rf /tmp/nanofuse
rm -f /tmp/vmlinux-*
rm -f /tmp/*build*.log

# Verify clean
docker images | grep nanofuse
ls /tmp/nanofuse 2>&1 | head -1
```

## Pre-Build Checklist

After cleaning, verify:

```bash
# No Docker images
docker images | grep nanofuse
# Expected: (empty - no output)

# No Docker containers
docker ps -a | grep nanofuse
# Expected: (empty - no output)

# No build directory
ls -la /home/jpoley/src/_mine/nanofuse/images/base/build
# Expected: No such file or directory

# No temporary data
ls /tmp/nanofuse
# Expected: No such file or directory

# No running daemons
pgrep -f nanofused
# Expected: (empty - no output)
```

## Files Reference

### Scripts (images/base/)
- `clean.sh` (6.5KB) - Standard clean
- `clean-sudo.sh` (6.9KB) - Complete clean
- `check-clean.sh` (2.7KB) - Check status

### Scripts (root)
- `clean-all.sh` (6.1KB) - Master cleanup

### Documentation
- `CLEAN-SCRIPTS.md` - This file
- `CLEAN.md` - Detailed clean guide
- `README-BUILD.md` - Build process

## Environment Variables

None currently used by clean scripts.

## Performance

| Script | Speed | Notes |
|--------|-------|-------|
| `check-clean.sh` | <1s | Just checks, doesn't clean |
| `clean.sh` | 10-30s | Docker operations take time |
| `clean-sudo.sh` | 30-60s | Same as clean.sh + disk I/O |
| `clean-all.sh` | 60-120s | Entire repo cleanup |

Slowest operations:
- Docker buildx prune (can be 30+ seconds)
- Large filesystem removal (if /tmp/nanofuse has many VMs)

## See Also

- `build.sh` - Main build script
- `build-kernel-docker.sh` - Kernel builder
- `README-BUILD.md` - Build process guide
- `/tmp` - Where temporary build artifacts go

## Questions?

Check these docs in order:
1. `CLEAN-SCRIPTS.md` (this file) - Overview and workflows
2. `CLEAN.md` - Detailed clean guide
3. `README-BUILD.md` - Build process
4. Run `./check-clean.sh` to see current state
