---
id: task-021
title: 'Task 5.1: Base Image and Kernel Build Testing'
status: Done
assignee: []
created_date: '2025-11-27'
labels:
  - Testing
  - P0
  - Build
  - Critical
dependencies:
  - task-020
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Outcome: Comprehensive test suite for building the Firecracker base image and kernel from scratch, validating the entire build process.

## Background and Rationale

### Why Test the Build Process?

Building a Firecracker-compatible kernel and rootfs is complex with many potential failure modes:

1. **Kernel Configuration**: Many kernel options required for Firecracker
2. **Rootfs Structure**: Specific init system, networking, and device requirements
3. **Image Format**: ext4 with correct block size and filesystem options
4. **Cross-compilation**: Build tools and dependencies must be correct

### Build Components to Test

| Component | Source | Validation |
|-----------|--------|------------|
| Kernel | Linux kernel source | Boots, provides required features |
| Rootfs | Alpine/BusyBox | Init works, network works, shell works |
| Init System | Custom init or openrc | Services start, networking initializes |
| Image Format | ext4 filesystem | Firecracker can mount and boot |

### Critical Build Boundaries

**Kernel Build:**
- Kernel config has all required options (virtio, ext4, etc.)
- Kernel binary size is within limits (< 50MB recommended)
- Kernel boots without panic
- Required modules/features are builtin or loadable

**Rootfs Build:**
- Root filesystem is valid ext4
- Init process exists and is executable
- /etc/inittab or equivalent configured
- Networking tools present (ip, dhclient or equivalent)
- SSH server present and configured

**Integration:**
- Kernel + rootfs boot together
- Network interface comes up
- SSH accessible after boot
- Console output captured

## Acceptance Criteria

### AC1: Build Scripts Exist and Execute

**Given** the build testing is implemented
**When** examining build scripts
**Then** scripts exist for kernel and rootfs builds

**Verification:**
```bash
# Check build scripts exist
test -f scripts/build-kernel.sh || test -f images/kernel/build.sh
# Expected: exit code 0

test -f scripts/build-rootfs.sh || test -f images/base/build.sh
# Expected: exit code 0

# Scripts are executable
test -x scripts/build-kernel.sh || test -x images/kernel/build.sh 2>/dev/null
test -x scripts/build-rootfs.sh || test -x images/base/build.sh 2>/dev/null
# Expected: at least one is true
```

### AC2: Kernel Build Test Suite Exists

**Given** the kernel build test suite is implemented
**When** running kernel build tests
**Then** tests validate kernel build output

**Verification:**
```bash
# Check test files exist
test -f test/build/kernel_test.go || test -f test/gdt/build/kernel.yaml
# Expected: exit code 0

# Run kernel build tests
go test -v ./test/build/... -run Kernel 2>&1 | tee /tmp/kernel-build.log
# Expected: tests exist (may skip if kernel not built)
```

### AC3: Rootfs Build Test Suite Exists

**Given** the rootfs build test suite is implemented
**When** running rootfs build tests
**Then** tests validate rootfs contents

**Verification:**
```bash
# Check test files exist
test -f test/build/rootfs_test.go || test -f test/gdt/build/rootfs.yaml
# Expected: exit code 0

# Run rootfs build tests
go test -v ./test/build/... -run Rootfs 2>&1 | tee /tmp/rootfs-build.log
# Expected: tests exist (may skip if rootfs not built)
```

### AC4: Kernel Configuration Validation

**Given** a kernel config file exists
**When** validating the configuration
**Then** all required options are enabled

**Verification:**
```bash
# Check kernel config exists
test -f images/kernel/config || test -f kernel/.config
# Expected: exit code 0

# Check required options (if config exists)
KERNEL_CONFIG=$(find . -name ".config" -o -name "kernel-config" | head -1)
if [ -n "$KERNEL_CONFIG" ]; then
  grep -q "CONFIG_VIRTIO_BLK=y" "$KERNEL_CONFIG" || echo "WARN: VIRTIO_BLK not enabled"
  grep -q "CONFIG_VIRTIO_NET=y" "$KERNEL_CONFIG" || echo "WARN: VIRTIO_NET not enabled"
  grep -q "CONFIG_EXT4_FS=y" "$KERNEL_CONFIG" || echo "WARN: EXT4_FS not enabled"
fi
# Expected: all required options present
```

### AC5: Rootfs Structure Validation

**Given** a built rootfs exists
**When** validating the structure
**Then** all required files and directories exist

**Verification:**
```bash
# Find rootfs image
ROOTFS=$(find . -name "*.ext4" -o -name "rootfs.img" | head -1)

if [ -n "$ROOTFS" ]; then
  # Mount and check structure (requires sudo)
  sudo mkdir -p /tmp/rootfs-check
  sudo mount -o loop "$ROOTFS" /tmp/rootfs-check 2>/dev/null || true

  # Check critical paths exist
  test -f /tmp/rootfs-check/sbin/init || test -L /tmp/rootfs-check/sbin/init
  test -d /tmp/rootfs-check/etc
  test -d /tmp/rootfs-check/dev

  sudo umount /tmp/rootfs-check 2>/dev/null || true
fi
# Expected: critical paths exist
```

### AC6: Mage Build Targets Exist

**Given** the build process is automated
**When** listing mage targets
**Then** build targets for kernel and image exist

**Verification:**
```bash
# Check for build-related mage targets
mage -l | grep -qiE "kernel|image|rootfs|build"
# Expected: exit code 0
```

### AC7: Documentation Exists in docs/tests/

**Given** the build test suite is implemented
**When** checking documentation
**Then** comprehensive build docs exist

**Verification:**
```bash
# Check build test docs exist
test -f docs/tests/build-testing.md
# Expected: exit code 0

# Check content covers key areas
grep -qiE "kernel|rootfs|image" docs/tests/build-testing.md
# Expected: exit code 0
```

## Technical Implementation

### Directory Structure

```
test/
├── build/
│   ├── kernel_test.go          # Kernel build validation
│   ├── rootfs_test.go          # Rootfs structure validation
│   ├── image_test.go           # Combined image validation
│   └── testdata/
│       └── required-kernel-opts.txt
├── gdt/
│   └── build/
│       ├── build_test.go       # Go test wrapper
│       ├── kernel.yaml         # Kernel build tests
│       └── rootfs.yaml         # Rootfs build tests
images/
├── kernel/
│   ├── build.sh                # Kernel build script
│   ├── config                  # Kernel configuration
│   └── Dockerfile              # Containerized build
└── base/
    ├── build.sh                # Rootfs build script
    ├── Dockerfile              # Multi-stage build
    └── init/
        └── init.sh             # Init script
docs/
└── tests/
    └── build-testing.md        # Build testing documentation
```

### Kernel Configuration Requirements

Based on Firecracker documentation (https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel-policy.md):

```
# Required for Firecracker
CONFIG_VIRTIO_BLK=y
CONFIG_VIRTIO_NET=y
CONFIG_VIRTIO_MMIO=y
CONFIG_EXT4_FS=y
CONFIG_TMPFS=y
CONFIG_DEVTMPFS=y
CONFIG_DEVTMPFS_MOUNT=y

# Networking
CONFIG_NET=y
CONFIG_INET=y
CONFIG_PACKET=y

# Console
CONFIG_SERIAL_8250=y
CONFIG_SERIAL_8250_CONSOLE=y
```

### Example Build Test (gdt format)

```yaml
# test/gdt/build/kernel.yaml
name: Kernel Build Validation
description: Validates kernel build output

tests:
  - name: kernel-binary-exists
    exec:
      command: test -f images/kernel/vmlinux
      assert:
        exit_code: 0

  - name: kernel-size-reasonable
    exec:
      command: |
        SIZE=$(stat -f%z images/kernel/vmlinux 2>/dev/null || stat -c%s images/kernel/vmlinux)
        test "$SIZE" -lt 52428800  # Less than 50MB
      assert:
        exit_code: 0

  - name: kernel-config-has-virtio
    exec:
      command: grep -q "CONFIG_VIRTIO_BLK=y" images/kernel/config
      assert:
        exit_code: 0
```

### Example Rootfs Validation

```yaml
# test/gdt/build/rootfs.yaml
name: Rootfs Build Validation
description: Validates rootfs structure and contents

tests:
  - name: rootfs-image-exists
    exec:
      command: ls images/base/*.ext4 2>/dev/null || ls images/base/rootfs.img
      assert:
        exit_code: 0

  - name: rootfs-has-init
    needs_sudo: true
    exec:
      command: |
        ROOTFS=$(ls images/base/*.ext4 2>/dev/null | head -1)
        mkdir -p /tmp/rootfs-test
        mount -o loop "$ROOTFS" /tmp/rootfs-test
        test -f /tmp/rootfs-test/sbin/init || test -L /tmp/rootfs-test/sbin/init
        RESULT=$?
        umount /tmp/rootfs-test
        exit $RESULT
      assert:
        exit_code: 0
```

### References

- **Firecracker Kernel Requirements**: https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel-policy.md
- **Building Custom Kernels**: https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel-building.md
- **Alpine Linux as rootfs**: https://wiki.alpinelinux.org/wiki/Creating_a_Virtual_Machine
- **BusyBox init**: https://git.busybox.net/busybox/tree/examples/inittab

## Definition of Done
- [ ] All 7 acceptance criteria pass
- [ ] Kernel build validation tests (config, binary, size)
- [ ] Rootfs structure validation tests (init, /etc, /dev)
- [ ] Mage target for build validation (`mage TestBuild`)
- [ ] Build scripts documented
- [ ] Documentation in docs/tests/build-testing.md

Priority: P0 (MUST HAVE)
Prerequisite: Task-20 (test harness framework)
Output Files:
- `test/build/*.go` or `test/gdt/build/*.yaml`
- `scripts/build-*.sh` or `images/*/build.sh`
- `docs/tests/build-testing.md`
<!-- SECTION:DESCRIPTION:END -->
