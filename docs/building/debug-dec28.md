# Debug Session - December 28, 2025

## Problem

VM creation failed with:
```
Boot source error: The kernel file cannot be opened: No such file or directory (os error 2)
```

The bridge showed `NO-CARRIER` because the VM never started - Firecracker failed during initialization.

## Root Cause

The registry client (`internal/registry/client.go`) had a TODO at lines 168-175:
```go
// TODO: Extract layers and build rootfs
// For now, create placeholder paths
rootfsPath := filepath.Join(imageDir, "rootfs.ext4")
kernelPath := filepath.Join(imageDir, "vmlinux")
c.logger.Warn("Layer extraction not yet implemented - using placeholder paths")
```

It fetched OCI metadata but never extracted the actual layers, setting placeholder paths that didn't exist.

## Solution Implemented

### 1. Debug Kernel for Testing

Downloaded official Firecracker quickstart kernel/rootfs:
```
test/fixtures/debug-kernel/
├── vmlinux.bin   (21MB - Firecracker 5.10 kernel)
└── rootfs.ext4   (300MB - Ubuntu bionic)
```

### 2. DockerBuilder Package

Created `internal/builder/` package for OCI image extraction:

| File | Purpose |
|------|---------|
| `interface.go` | Builder interface definition |
| `docker.go` | DockerBuilder using Docker/Podman |
| `docker_test.go` | Unit tests |

The DockerBuilder:
1. Pulls image with `docker pull`
2. Creates temp container and exports filesystem
3. Extracts kernel from `/boot/vmlinux*`
4. Creates ext4 rootfs from exported tar

### 3. Registry Client Integration

Updated `internal/registry/client.go` to use DockerBuilder:
- Auto-detects Docker/Podman availability
- Falls back to metadata-only if unavailable
- Reports extraction progress

## Files Created/Modified

**New files:**
- `test/fixtures/debug-kernel/vmlinux.bin`
- `test/fixtures/debug-kernel/rootfs.ext4`
- `scripts/debug-boot.sh`
- `scripts/register-debug-kernel.sh`
- `internal/builder/interface.go`
- `internal/builder/docker.go`
- `internal/builder/docker_test.go`
- `docs/building/IMAGE_EXTRACTION_ARCHITECTURE.md`

**Modified files:**
- `internal/registry/client.go`

## Testing Steps

### Step 1: Verify Firecracker Works Independently

```bash
cd /home/jpoley/prj/ps/nanofuse
./scripts/debug-boot.sh
```

This boots Firecracker directly (bypasses nanofuse). You should see:
- Kernel boot messages
- Login prompt (root, no password)
- Press Ctrl+C to exit

### Step 2: Register Debug Kernel for Nanofuse

```bash
sudo ./scripts/register-debug-kernel.sh
```

This registers the debug kernel in nanofuse's database as `debug:latest`.

### Step 3: Test VM with Debug Kernel

```bash
nanofuse vm create debug:latest test-debug
nanofuse vm start test-debug
nanofuse vm logs test-debug
```

### Step 4: Test Image Pull with Docker Extraction

```bash
# Rebuild daemon with new code
cd /home/jpoley/prj/ps/nanofuse
go build -o bin/nanofused ./cmd/nanofused

# Restart daemon (pick one)
sudo systemctl restart nanofused
# OR manually: sudo ./bin/nanofused

# Test pull with extraction
nanofuse image pull ghcr.io/jpoley/nanofuse/base:latest

# Check if kernel exists
ls -la /var/lib/nanofuse/images/*/vmlinux

# Create and start VM
nanofuse vm create ghcr.io/jpoley/nanofuse/base:latest test-pull
nanofuse vm start test-pull
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Image Pull Flow                              │
├─────────────────────────────────────────────────────────────────┤
│  registry/client.go                                              │
│       │                                                          │
│       ├── Fetch OCI metadata (go-containerregistry)              │
│       │                                                          │
│       └── builder.Extract()                                      │
│                │                                                 │
│                ▼                                                 │
│  builder/docker.go                                               │
│       │                                                          │
│       ├── docker pull <image>                                    │
│       ├── docker create --name temp <image>                      │
│       ├── docker export temp > rootfs.tar                        │
│       ├── Extract kernel from tar (/boot/vmlinux*)               │
│       ├── mkfs.ext4 rootfs.ext4                                  │
│       ├── mount + tar extract (requires root or fuse2fs)         │
│       └── docker rm temp                                         │
│                │                                                 │
│                ▼                                                 │
│       ┌─────────────────────┐                                    │
│       │ <digest>/           │                                    │
│       │   ├── vmlinux       │                                    │
│       │   └── rootfs.ext4   │                                    │
│       └─────────────────────┘                                    │
└─────────────────────────────────────────────────────────────────┘
```

## Known Limitations

1. **Root required for ext4 creation**: The mount-based rootfs creation requires root. Alternative: install `fuse2fs` for unprivileged extraction.

2. **Kernel must exist in image**: The builder looks for kernel at `/boot/vmlinux*`. If the container image doesn't include a kernel, extraction will fail.

3. **Docker/Podman required**: Without Docker or Podman, falls back to metadata-only (VMs won't boot).

## Future Work

- Move `internal/builder/` and `internal/layerbuild/` to `provenance` project
- Add native Go extraction (no Docker dependency)
- Support fuse2fs for unprivileged rootfs creation
- Add kernel download from external source if not in image

## Quick Reference

```bash
# Debug boot (raw Firecracker)
./scripts/debug-boot.sh

# Register debug kernel
sudo ./scripts/register-debug-kernel.sh

# Run builder tests
go test ./internal/builder/... -v

# Run integration test (requires Docker)
INTEGRATION_TEST=1 go test ./internal/builder/... -v -run Integration
```

---

## Session Update: 2025-12-28 Evening

### Issue Discovered

The debug-boot.sh successfully boots Firecracker, but uses **Ubuntu 18.04.5 LTS (bionic)** from 2018:
- EOL: April 2023 (7 years old, 2+ years past end-of-life)
- apt-daily services fail (no network + stale repos)
- Kernel 4.14.174 (old)

### How to Exit the VM

Multiple methods:
1. **Ctrl+C** - Sends SIGINT to foreground firecracker process
2. **`reboot`** in guest - Firecracker will terminate with `reboot=k panic=1` boot args
3. **`poweroff`** in guest - Clean shutdown
4. **From another terminal**: `pkill firecracker` or `pkill -9 firecracker`

### Upgrade Plan

**Decision**: Use official Firecracker CI images (December 2025):
- **Ubuntu 24.04** squashfs → convert to ext4
- **Kernel 6.1.155** (LTS, support until 2028)

Source: `s3://spec.ccfc.min/firecracker-ci/20251218-f2f293f67e5f-0/x86_64/`

### New Scripts

```bash
# Download modern fixtures (Ubuntu 24.04 + kernel 6.1.155)
./scripts/download-fixtures.sh

# This will:
# 1. Download kernel 6.1.155
# 2. Download Ubuntu 24.04 squashfs
# 3. Convert to ext4 (requires sudo for mount)
# 4. Create symlinks for compatibility
```

### Next Steps

1. **Kill stuck VM**: `Ctrl+C` or `pkill firecracker`
2. **Download modern images**: `./scripts/download-fixtures.sh`
3. **Test boot**: `./scripts/debug-boot.sh`
4. **Verify Ubuntu 24.04**: Should see `Ubuntu 24.04 LTS` at login
5. **Continue with Step 2**: Register debug kernel with nanofuse

### Decisions Logged

See `.logs/decisions/2025-12-28-debug-session.jsonl` for full decision record.

---

## Session Update: 2025-12-28 Night - Script Resilience

### Issue Discovered

When re-registering the debug kernel, `register-local-image` failed:
```
UNIQUE constraint failed: images.digest
```

### Solution Implemented

Made `register-local-image.go` resilient and idempotent:

1. **Added `UpsertImage` method** to `internal/storage/db.go`:
   - Uses SQLite `ON CONFLICT DO UPDATE` for idempotent operations
   - Can be run multiple times without failure

2. **Symlink resolution** with `filepath.EvalSymlinks`:
   - Resolves `vmlinux.bin` → actual `vmlinux-5.10.245-no-acpi`
   - Stores resolved paths in database

3. **Kernel version extraction** from filename:
   - Extracts `5.10.245` from `vmlinux-5.10.245-no-acpi`
   - Uses regex `(\d+\.\d+\.\d+)`

### Verification

```bash
# Tool is now idempotent - can run multiple times
$ ./bin/register-local-image /tmp/test.db "debug:latest" ... (runs 2x)
✓ Registered image: debug:latest  # No errors on second run

# Symlinks properly resolved
$ sqlite3 /tmp/test.db "SELECT kernel_version, kernel_path FROM images;"
5.10.245|test/fixtures/debug-kernel/vmlinux-5.10.245-no-acpi
```

### Next Steps

1. **Register with nanofuse** (requires sudo):
   ```bash
   sudo ./scripts/register-debug-kernel.sh
   ```

2. **Test VM creation**:
   ```bash
   nanofuse vm create debug:latest test-debug
   nanofuse vm start test-debug
   nanofuse vm logs test-debug
   ```

3. **Continue to Step 4**: Test Image Pull with Docker Extraction

---

## Session Update: 2025-12-28 Evening (Part 2) - Kernel Research

### Problem Discovered

The Firecracker CI kernel 6.1.155 (and 5.10.245) failed with:
```
VFS: Cannot open root device "vda" or unknown-block(0,0): error -6
```

### Deep Research Findings

**Root Cause**: Firecracker CI kernels are built with ACPI enabled but `CONFIG_PCI=n`. This configuration ONLY works with Amazon Linux microvm-patched kernels. Standard kernels fail because:
1. ACPI needs PCI to parse tables properly
2. Without proper ACPI init, virtio-mmio devices don't register
3. No block device → kernel panic

**References**:
- [Issue #4881](https://github.com/firecracker-microvm/firecracker/issues/4881) - Known bug
- [Issue #4816](https://github.com/firecracker-microvm/firecracker/issues/4816) - Related discussion

### Solution

Use the `-no-acpi` kernel variant:
- `vmlinux-5.10.245-no-acpi` (Dec 2025, LTS until Sept 2026)
- Uses legacy MPTable boot instead of ACPI
- Works with standard ext4 rootfs

### Final Approved Stack

| Component | Version | EOL |
|-----------|---------|-----|
| Kernel | 5.10.245-no-acpi | Sept 2026 |
| Rootfs | Ubuntu 24.04 | 2029 |

### Documentation Created

- `docs/building/FIRECRACKER_KERNEL_GUIDE.md` - Complete kernel selection guide
- `.logs/decisions/2025-12-28-debug-session.jsonl` - 14 decisions logged
