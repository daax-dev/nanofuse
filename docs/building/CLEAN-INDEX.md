# NanoFuse Clean Scripts Index

Complete index of all clean scripts, documentation, and usage guides.

## Quick Start

**Just want to clean?**
```bash
# Check what needs cleaning (no removal)
cd /home/jpoley/src/_mine/nanofuse/images/base
./check-clean.sh

# Standard clean (no sudo)
./clean.sh

# Complete clean (with sudo)
sudo ./clean-sudo.sh

# Master clean from anywhere
sudo /home/jpoley/src/_mine/nanofuse/clean-all.sh
```

## Scripts Location

### Root Level: `/home/jpoley/src/_mine/nanofuse/`
- **`clean-all.sh`** (6.1KB)
  - Master cleanup script
  - Cleans entire repository
  - Auto-detects if sudo needed
  - Run from any directory

### Image Build Level: `images/base/`
- **`check-clean.sh`** (2.7KB)
  - Check what needs cleaning
  - Non-destructive
  - Shows artifact counts
  - Recommends cleanup commands

- **`clean.sh`** (6.5KB)
  - Standard build clean
  - No sudo required
  - Removes Docker artifacts and logs
  - Warns about root-owned files

- **`clean-sudo.sh`** (6.9KB)
  - Complete build clean
  - Requires sudo
  - Removes all files
  - Removes root-owned artifacts

## Documentation

### Master Clean Guide: `CLEAN-SCRIPTS.md`
**Location:** `/home/jpoley/src/_mine/nanofuse/CLEAN-SCRIPTS.md` (8.6KB)

**Contents:**
- Overview of all 4 scripts
- Quick reference table
- Detailed usage instructions
- 4 cleanup levels
- Common workflows
- Troubleshooting guide
- Manual cleanup commands
- Pre-build checklist

**Best for:** Understanding all clean options

---

### Build Clean Guide: `CLEAN.md`
**Location:** `/home/jpoley/src/_mine/nanofuse/images/base/CLEAN.md` (4.6KB)

**Contents:**
- What each script cleans
- Usage examples
- Common issues and solutions
- Quick command reference
- Pre-build verification

**Best for:** Specific build issues

---

### Build Documentation: `README-BUILD.md`
**Location:** `/home/jpoley/src/_mine/nanofuse/images/base/README-BUILD.md` (7.6KB)

**Contents:**
- Quick start guide
- Build artifacts explanation
- Script descriptions
- Kernel details (6.1.90, ELF format)
- Dockerfile breakdown
- Complete workflow
- Troubleshooting
- File reference

**Best for:** Understanding the build process

## Cleanup Levels

| Level | Script | Time | Scope | Sudo |
|-------|--------|------|-------|------|
| 1 | `check-clean.sh` | 2s | Check only | No |
| 2 | `clean.sh` | 10-30s | User files | No |
| 3 | `clean-sudo.sh` | 30-60s | All files | Yes |
| 4 | `clean-all.sh` | 60-120s | Full repo | Auto |

## Common Workflows

### Daily Development
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
./check-clean.sh        # Quick check
./clean.sh              # Light clean
sudo ./build.sh         # Build
```

### Before Major Changes
```bash
cd /home/jpoley/src/_mine/nanofuse/images/base
sudo ./clean-sudo.sh    # Complete clean
./check-clean.sh        # Verify
sudo ./build.sh         # Fresh build
```

### Troubleshooting
```bash
sudo /home/jpoley/src/_mine/nanofuse/clean-all.sh  # Full reset
cd /home/jpoley/src/_mine/nanofuse/images/base
./check-clean.sh                                    # Verify clean
sudo ./build.sh                                     # Rebuild
```

## What Gets Cleaned

### Docker
- Images: `nanofuse*`, `kernel-builder`
- Containers
- BuildKit cache
- Volumes

### Build Artifacts
- `./images/base/build/vmlinux`
- `./images/base/build/rootfs.ext4`
- `./images/base/build/manifest.json`

### Temporary Files
- `/tmp/vmlinux-*` (kernel builds)
- `/tmp/nanofuse/` (daemon data)
- `/tmp/fc-test-*` (test directories)
- `/tmp/*build*.log` (build logs)

### Processes
- `nanofused` daemon
- `firecracker` VMs
- `systemd nanofused.service`

## File Locations Quick Reference

```
/home/jpoley/src/_mine/nanofuse/
├── clean-all.sh                    # Master clean (run from here)
├── CLEAN-SCRIPTS.md                # Master guide
├── CLEAN-INDEX.md                  # This file
│
└── images/base/
    ├── clean.sh                    # Standard clean
    ├── clean-sudo.sh               # Complete clean
    ├── check-clean.sh              # Check status
    ├── CLEAN.md                    # Detailed clean guide
    ├── README-BUILD.md             # Build documentation
    │
    ├── build.sh                    # Main build script
    ├── build-kernel-docker.sh      # Kernel builder
    ├── Dockerfile                  # Base image
    ├── Dockerfile.kernel           # Kernel builder
    │
    └── build/                      # Build artifacts (created)
        ├── vmlinux                 # ELF kernel
        ├── rootfs.ext4             # Filesystem
        └── manifest.json           # Metadata
```

## Decision Tree

```
Q: Just want to check?
├─ Yes → ./check-clean.sh

Q: Need to build?
├─ Yes, standard clean → ./clean.sh
├─ Yes, major changes  → sudo ./clean-sudo.sh
├─ Yes, troubleshooting → sudo clean-all.sh

Q: Starting fresh build?
├─ From anywhere  → sudo clean-all.sh
├─ From images/base → sudo ./clean-sudo.sh
└─ Quick refresh  → ./clean.sh
```

## Reading Guide

**New to NanoFuse?**
1. Start with: `README-BUILD.md` (understand build process)
2. Then read: `CLEAN-SCRIPTS.md` (understand clean options)
3. Reference: This file (find things quickly)

**Just need to clean?**
1. Run: `./check-clean.sh` (see what's dirty)
2. Pick a level: Standard, Complete, or Full
3. Reference: `CLEAN.md` if issues

**Building regularly?**
1. Know: `./clean.sh` for quick refresh
2. Know: `sudo ./clean-sudo.sh` for major changes
3. Reference: `CLEAN-SCRIPTS.md` for workflows

## Features

All clean scripts include:
- ✓ Color-coded output
- ✓ Error handling
- ✓ Process verification
- ✓ Final verification
- ✓ Informative warnings
- ✓ Manual fallback commands

## Support

If something goes wrong:
1. Check: `CLEAN.md` (Troubleshooting section)
2. Check: `CLEAN-SCRIPTS.md` (Troubleshooting section)
3. Run: `./check-clean.sh` (see current state)
4. Use: Manual commands (in `CLEAN.md`)

## See Also

- `../build.sh` - Main build script
- `../build-kernel-docker.sh` - Kernel builder
- `../test-build-and-boot.sh` - End-to-end test
- `../../CLEAN-SCRIPTS.md` - Main clean guide

---

**Last Updated:** 2025-11-06
**Scripts:** 4 (check, clean, clean-sudo, clean-all)
**Docs:** 3 (this index + 2 guides)
