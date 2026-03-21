# Script Inventory and Consolidation Plan

**Note**: Some scripts listed below have been archived to `scripts/archives/` as they were used during initial development. See the archives directory for historical scripts.

## Current Active Scripts (5 total)

### BUILD Scripts (2)
1. **build.sh** - Main build script, creates rootfs + kernel
2. **download-fc-kernel.sh** - Downloads pre-built Firecracker kernel

### TEST Scripts (2)
1. **test-boot.sh** - Comprehensive boot test with all checks
2. **validate-build.sh** - Validates build artifacts

### CLEAN Scripts (1)
1. **clean.sh** - Complete clean (handles non-root cleanup)

## Archived Scripts (in scripts/archives/)

### BUILD Scripts
1. **BUILD_AND_TEST.sh** - Build kernel and immediately test (archived)
2. **BUILD_CLEAN.sh** - Nuclear clean Docker build from scratch (archived)
3. **BUILD_KERNEL_ONLY.sh** - Builds only kernel (archived)
4. **build-kernel-docker.sh** - Builds only kernel using Docker (archived)
5. **test-kernel-fix.sh** - Validates VIRTIO_MMIO_CMDLINE_DEVICES fix (archived)

### CLEAN Scripts (4) ⚠️ DUPLICATION
1. **clean.sh** - Complete clean (handles non-root cleanup)
2. **clean-sudo.sh** - Complete clean with sudo (handles root-owned files)
3. **check-clean.sh** - Check what needs cleaning (dry-run)
4. **BUILD_CLEAN.sh** - Docker-only clean + rebuild kernel

### TEST Scripts (4) ⚠️ DUPLICATION
1. **test-boot.sh** - Comprehensive boot test with all checks
2. **TEST_BOOT.sh** - Basic kernel boot test
3. **TEST_BOOT_VERBOSE.sh** - Detailed boot test with verbose output
4. **test-kernel-fix.sh** - Validates VIRTIO_MMIO_CMDLINE_DEVICES fix

### UTILITY Scripts (1)
1. **download-fc-kernel.sh** - Downloads pre-built Firecracker kernel
2. **validate-build.sh** - Validates build artifacts

---

## Detailed Analysis

### CLEAN Scripts - Overlapping Functionality

| Script | Docker | Build Dir | /tmp files | Processes | Needs Sudo | Docker Cache |
|--------|--------|-----------|------------|-----------|------------|--------------|
| clean.sh | ✓ | ✓ (warn) | ✓ | ✓ | No (warns) | ✓ |
| clean-sudo.sh | ✓ | ✓ (force) | ✓ | ✓ | Yes (required) | ✓ |
| check-clean.sh | Check only | Check only | Check only | Check only | No | No |
| BUILD_CLEAN.sh | ✓ (prune -af) | No | No | No | No | ✓ |

**Analysis**:
- `clean.sh` and `clean-sudo.sh` are 95% identical - differ only in sudo check and force cleanup
- `check-clean.sh` is a dry-run version of the other two
- `BUILD_CLEAN.sh` is Docker-specific and also rebuilds

**Recommendation**: Consolidate into ONE script with intelligent sudo detection

### TEST Scripts - Overlapping Functionality

| Script | Purpose | Output Style | Checks VIRTIO | Comprehensive | Use Case |
|--------|---------|--------------|---------------|---------------|----------|
| test-boot.sh | Full validation | Structured | No | Yes (6 tests) | Production |
| TEST_BOOT.sh | Basic boot | Simple | Yes | Limited | Quick check |
| TEST_BOOT_VERBOSE.sh | Detailed analysis | Verbose | Yes | Yes (7 tests) | Debugging |
| test-kernel-fix.sh | VIRTIO validation | Detailed | Yes | Specific | Development |

**Analysis**:
- `TEST_BOOT.sh` and `TEST_BOOT_VERBOSE.sh` are 80% identical - differ only in output verbosity
- `test-boot.sh` is most comprehensive but lacks VIRTIO-specific checks
- `test-kernel-fix.sh` is specialized for one thing

**Recommendation**: Consolidate into ONE script with verbosity levels

---

## Consolidation Plan

### Phase 1: Consolidate CLEAN scripts → `clean.sh`

**New unified clean.sh**:
```bash
#!/bin/bash
# Usage: ./clean.sh [--check|--sudo]
#   --check : Dry-run, show what would be cleaned
#   --sudo  : Force sudo mode for root-owned files
#   (default): Auto-detect sudo if needed
```

Features:
- Auto-detect if sudo is needed (check for root-owned files)
- Offer to re-run with sudo if needed
- Support --check flag for dry-run
- Support --sudo flag to force sudo mode
- Clean everything: Docker, build/, /tmp/, processes, cache

**Replaces**: clean.sh, clean-sudo.sh, check-clean.sh
**Keep**: BUILD_CLEAN.sh (specialized for Docker-only + rebuild)

### Phase 2: Consolidate TEST scripts → `test-boot.sh`

**New unified test-boot.sh**:
```bash
#!/bin/bash
# Usage: ./test-boot.sh [--verbose] [--check-virtio] <kernel> <rootfs>
#   --verbose      : Show detailed output
#   --check-virtio : Include VIRTIO-specific checks
#   (default): Standard output with all tests
```

Features:
- All 7 boot validation tests
- VIRTIO-specific checks (opt-in via flag)
- Configurable verbosity
- Support both /tmp/vmlinux-* and build/vmlinux paths

**Replaces**: test-boot.sh, TEST_BOOT.sh, TEST_BOOT_VERBOSE.sh
**Keep**: test-kernel-fix.sh (specialized for VIRTIO validation during development)

---

## Final Script List (7 scripts, down from 13)

### BUILD (4)
1. **build.sh** - Main build (rootfs + kernel)
2. **build-kernel-docker.sh** - Build kernel only
3. **BUILD_AND_TEST.sh** - Build kernel + test
4. **BUILD_CLEAN.sh** - Docker-only clean + rebuild

### CLEAN (1)
1. **clean.sh** - Unified clean with auto-sudo detection

### TEST (1)
1. **test-boot.sh** - Unified boot test with verbosity options

### UTILITY (2)
1. **download-fc-kernel.sh** - Download pre-built kernel
2. **validate-build.sh** - Validate build artifacts
3. **test-kernel-fix.sh** - VIRTIO-specific validation (development only)

---

## Migration Plan

### Step 1: Create new unified scripts
- [ ] Create new clean.sh with all features
- [ ] Create new test-boot.sh with all features

### Step 2: Test new scripts
- [ ] Test clean.sh without sudo
- [ ] Test clean.sh with sudo
- [ ] Test clean.sh --check
- [ ] Test test-boot.sh normal mode
- [ ] Test test-boot.sh --verbose
- [ ] Test test-boot.sh --check-virtio

### Step 3: Remove old scripts
- [ ] Remove clean-sudo.sh
- [ ] Remove check-clean.sh
- [ ] Remove TEST_BOOT.sh
- [ ] Remove TEST_BOOT_VERBOSE.sh

### Step 4: Update documentation
- [ ] Update BUILD_GUIDE.md
- [ ] Update README.md
- [ ] Add migration notes for users

---

## Rationale

### Why consolidate?

1. **Reduces confusion**: Users won't wonder which script to use
2. **Easier maintenance**: One script to fix bugs, not four
3. **Better UX**: Flags are clearer than different filenames
4. **Less duplication**: DRY principle
5. **Consistency**: Standard Unix pattern (flags over multiple scripts)

### Why keep some scripts separate?

1. **BUILD_CLEAN.sh**: Specialized workflow (Docker clean + rebuild)
2. **test-kernel-fix.sh**: Development-specific, not for regular use
3. **build-kernel-docker.sh**: Focused single purpose
4. **validate-build.sh**: Different use case (validation vs testing)
5. **download-fc-kernel.sh**: Utility, not part of main workflow

---

## Examples After Consolidation

```bash
# Clean everything (auto-detects sudo need)
./clean.sh

# Check what would be cleaned
./clean.sh --check

# Force sudo mode
./clean.sh --sudo

# Test kernel boot (standard output)
./test-boot.sh /tmp/vmlinux-test /tmp/rootfs-working.ext4

# Test with verbose output
./test-boot.sh --verbose build/vmlinux build/rootfs.ext4

# Test with VIRTIO checks
./test-boot.sh --check-virtio /tmp/vmlinux-fresh-build

# Full clean Docker rebuild
./BUILD_CLEAN.sh

# Build and test in one command
./BUILD_AND_TEST.sh
```
