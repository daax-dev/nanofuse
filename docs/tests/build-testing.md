# Build Testing

## Overview

Build testing validates the kernel and rootfs build process to ensure Firecracker-compatible artifacts are produced correctly.

## Why Test Builds?

Building Firecracker-compatible kernels and rootfs images is complex:
- Kernel must have specific options enabled
- Rootfs must have correct structure and init
- Image format must be compatible
- Cross-compilation adds complexity

## Test Location

```
test/
├── build/
│   ├── kernel_test.go     # Kernel build validation
│   ├── rootfs_test.go     # Rootfs structure validation
│   └── image_test.go      # Combined image validation
├── gdt/
│   └── build/
│       ├── kernel.yaml    # Kernel build tests
│       └── rootfs.yaml    # Rootfs build tests
```

## Running Build Tests

```bash
# Run all build tests
mage TestBuild

# Run kernel tests only
go test -v ./test/build/... -run Kernel

# Run rootfs tests only
go test -v ./test/build/... -run Rootfs
```

## Kernel Configuration Requirements

Based on [Firecracker kernel policy](https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel-policy.md):

### Required Options

```
# Virtio (required for Firecracker)
CONFIG_VIRTIO_BLK=y
CONFIG_VIRTIO_NET=y
CONFIG_VIRTIO_MMIO=y

# Filesystem
CONFIG_EXT4_FS=y
CONFIG_TMPFS=y

# Device management
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

### Verification Script

```bash
#!/bin/bash
# scripts/verify-kernel-config.sh

CONFIG=${1:-images/kernel/config}

REQUIRED=(
    "CONFIG_VIRTIO_BLK=y"
    "CONFIG_VIRTIO_NET=y"
    "CONFIG_VIRTIO_MMIO=y"
    "CONFIG_EXT4_FS=y"
)

for opt in "${REQUIRED[@]}"; do
    if ! grep -q "^$opt$" "$CONFIG"; then
        echo "MISSING: $opt"
        exit 1
    fi
done

echo "All required kernel options present"
```

## Rootfs Structure Requirements

### Required Directories

```
/
├── bin/           # Essential binaries
├── dev/           # Device files (populated by devtmpfs)
├── etc/           # Configuration
│   ├── passwd     # User database
│   ├── group      # Group database
│   └── inittab    # Init configuration (if using busybox)
├── lib/           # Libraries (optional for static builds)
├── proc/          # Proc filesystem (mount point)
├── sbin/
│   └── init       # Init process (required!)
├── sys/           # Sysfs (mount point)
└── tmp/           # Temp files
```

### Required Init Script

```bash
#!/bin/sh
# /sbin/init

# Mount essential filesystems
mount -t proc none /proc
mount -t sysfs none /sys
mount -t devtmpfs none /dev

# Setup networking
ip link set lo up
ip link set eth0 up
udhcpc -i eth0 -s /etc/udhcpc.script

# Start SSH if available
if [ -x /usr/sbin/sshd ]; then
    mkdir -p /run/sshd
    /usr/sbin/sshd
fi

# Drop to shell
exec /bin/sh
```

## Example Build Tests (gdt format)

```yaml
# test/gdt/build/kernel.yaml
name: Kernel Build Validation
description: Validates kernel build output

tests:
  - name: kernel-config-exists
    exec:
      command: test -f images/kernel/config
      assert:
        exit_code: 0

  - name: kernel-has-virtio-blk
    exec:
      command: grep -q "CONFIG_VIRTIO_BLK=y" images/kernel/config
      assert:
        exit_code: 0

  - name: kernel-has-virtio-net
    exec:
      command: grep -q "CONFIG_VIRTIO_NET=y" images/kernel/config
      assert:
        exit_code: 0

  - name: kernel-binary-exists-after-build
    skip_if:
      - no_kernel_build
    exec:
      command: test -f images/kernel/vmlinux
      assert:
        exit_code: 0
```

```yaml
# test/gdt/build/rootfs.yaml
name: Rootfs Build Validation
description: Validates rootfs structure

tests:
  - name: rootfs-image-exists
    exec:
      command: ls images/base/*.ext4 2>/dev/null || ls images/base/rootfs.img
      assert:
        exit_code: 0

  - name: dockerfile-exists
    exec:
      command: test -f images/base/Dockerfile
      assert:
        exit_code: 0
```

## Mage Integration

```go
// In magefile.go

// TestBuild validates kernel and rootfs build outputs
func TestBuild() error {
    fmt.Println("Running build validation tests...")
    return sh.RunV("go", "test", "-v", "./test/build/...", "./test/gdt/build/...")
}
```

## References

- [Firecracker Kernel Policy](https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel-policy.md)
- [Building Custom Kernels](https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel-building.md)
- [Alpine Linux Rootfs](https://wiki.alpinelinux.org/wiki/Creating_a_Virtual_Machine)
