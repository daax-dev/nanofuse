# Firecracker Kernel Selection Guide

> **Last Updated**: 2025-12-28
> **Status**: Production-ready configuration documented

## TL;DR - What Works

| Component | Version | Source | EOL |
|-----------|---------|--------|-----|
| **Kernel** | `vmlinux-5.10.245-no-acpi` | Firecracker CI S3 | Sept 2026 |
| **Rootfs** | Ubuntu 24.04 ext4 | Firecracker CI S3 | 2029 |

```bash
# Download working fixtures
./scripts/download-fixtures.sh
```

## The Problem We Solved

### Symptom
```
VFS: Cannot open root device "vda" or unknown-block(0,0): error -6
Kernel panic - not syncing: VFS: Unable to mount root fs on unknown-block(0,0)
```

### Root Cause

Firecracker CI kernels (5.10.245, 6.1.155) are built with **ACPI enabled but CONFIG_PCI disabled**. This configuration only works with Amazon Linux microvm-patched kernels.

When using these kernels with mainline Linux or standard ext4 rootfs:
1. ACPI initializes but can't parse tables properly without PCI
2. virtio-mmio block devices fail to register
3. Kernel can't find root filesystem → panic

### Why It's Confusing

The CI kernels DO have `CONFIG_VIRTIO_BLK=y` enabled. The issue is the ACPI/PCI interaction, not missing virtio support.

## Kernel Variants Explained

| Kernel | ACPI | Boot Method | Works with ext4 | Notes |
|--------|------|-------------|-----------------|-------|
| `vmlinux-6.1.155` | Yes | ACPI | **No** | Requires Amazon Linux patches |
| `vmlinux-5.10.245` | Yes | ACPI | **No** | Same issue |
| `vmlinux-5.10.245-no-acpi` | No | MPTable | **Yes** | Legacy boot, works everywhere |

## Solution

Use the `-no-acpi` kernel variant which boots using legacy MPTable instead of ACPI. This bypasses the CONFIG_PCI requirement entirely.

### Download URLs

```bash
# Kernel (5.10.245-no-acpi)
https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/20251218-f2f293f67e5f-0/x86_64/vmlinux-5.10.245-no-acpi

# Ubuntu 24.04 rootfs (squashfs, needs conversion to ext4)
https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/20251218-f2f293f67e5f-0/x86_64/ubuntu-24.04.squashfs
```

## Boot Arguments

### Working Configuration
```json
{
  "boot-source": {
    "kernel_image_path": "/path/to/vmlinux-5.10.245-no-acpi",
    "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
  }
}
```

### For nanofused (with networking)
```
console=ttyS0 root=/dev/vda rw init=/lib/systemd/systemd ip=<IP>::<GW>:255.255.255.0::eth0:off net.ifnames=0 biosdevname=0
```

**Key points:**
- Use `root=/dev/vda` (not `vda1`) - rootfs is unpartitioned
- `pci=off` recommended for no-acpi kernels
- `net.ifnames=0 biosdevname=0` ensures `eth0` naming

## References

- [Firecracker Issue #4881](https://github.com/firecracker-microvm/firecracker/issues/4881) - Kernel panic with 6.1 config
- [Firecracker Issue #4816](https://github.com/firecracker-microvm/firecracker/issues/4816) - Unable to boot with newer kernels
- [Firecracker Kernel Policy](https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel-policy.md)

## Decision Log

All decisions tracked in `.logs/decisions/2025-12-28-debug-session.jsonl`

| ID | Decision |
|----|----------|
| DEC-001 | Identified Ubuntu 18.04 as stale (EOL 2023) |
| DEC-002 | Upgrade to Ubuntu 24.04 CI images |
| DEC-003 | Selected kernel 6.1.155 (later found broken) |
| DEC-009 | Fixed root=/dev/vda1 → root=/dev/vda |
| DEC-010 | Discovered CI kernel missing virtio_blk behavior |
| DEC-011 | Selected 5.10.245-no-acpi as working kernel |
| DEC-012 | Documented root cause (ACPI+PCI issue) |
| DEC-013 | Finalized kernel choice with rationale |
| DEC-014 | Ratified complete stack |

## Future Considerations

1. **When 5.10 EOLs (Sept 2026)**: Check if Firecracker releases a `6.x-no-acpi` variant
2. **Amazon Linux option**: Could use AL2023 microvm-tagged kernels with ACPI support
3. **Build custom kernel**: Compile 6.1 with `CONFIG_PCI=y` for ACPI boot

## Quick Validation

```bash
# Test kernel + rootfs directly
./scripts/debug-boot.sh

# Should boot to Ubuntu 24.04 login prompt
# Exit with: reboot or Ctrl+C
```
