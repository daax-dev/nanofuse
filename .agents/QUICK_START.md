# Quick Start: Using the MicroVM Kernel Expert Agent

This guide shows you how to quickly leverage the MicroVM Kernel Expert for common tasks.

## 5-Minute Start

### Task 1: Build Your First Minimal Kernel

**Goal**: Build a minimal 6.1 kernel for x86_64 Firecracker microVMs

**With Claude Code**:
```
@microvm-kernel-expert I need to build a minimal kernel for Firecracker on x86_64.
Target size: <8MB. Walk me through it step by step.
```

**Manual (following agent guidance)**:
```bash
# 1. Download kernel
wget https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz
tar xf linux-6.1.tar.xz && cd linux-6.1

# 2. Minimal config
make tinyconfig

# 3. Enable essentials (from agent's checklist)
scripts/config --enable SERIAL_8250_CONSOLE
scripts/config --enable VIRTIO_BLK
scripts/config --enable VIRTIO_NET
scripts/config --enable EXT4_FS
scripts/config --enable KVM_GUEST

# 4. Build
make -j$(nproc)

# 5. Check size
ls -lh arch/x86/boot/bzImage
```

**Expected Result**: Kernel image at `arch/x86/boot/bzImage`, size 5-8MB

---

### Task 2: Debug "No Console Output" Issue

**Symptom**: Firecracker starts but you see no console output

**With Claude Code**:
```
@microvm-kernel-expert My Firecracker VM boots but I get no console output.
How do I debug this?
```

**Quick Fix (from agent's troubleshooting section)**:
```bash
# 1. Verify kernel has serial support
grep CONFIG_SERIAL_8250_CONSOLE .config
# Must show: CONFIG_SERIAL_8250_CONSOLE=y

# 2. Check kernel cmdline includes console=ttyS0
# In Firecracker config:
{
  "boot_source": {
    "kernel_image_path": "vmlinuz",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw"
  }
}

# 3. Enable debug output
# Temporarily add to boot_args: earlyprintk=serial,ttyS0,115200 debug
```

---

### Task 3: Cross-Compile for ARM64

**Goal**: Build ARM64 kernel from x86_64 host

**With Claude Code**:
```
@microvm-kernel-expert Guide me through cross-compiling a 6.1 kernel for ARM64
from my x86_64 development machine.
```

**Manual (from agent's workflow)**:
```bash
# 1. Install cross-compiler
sudo apt-get install gcc-aarch64-linux-gnu

# 2. Configure for ARM64
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- defconfig

# 3. Enable ARM-specific features
scripts/config --enable VIRTIO_MMIO
scripts/config --enable ARM_AMBA
scripts/config --enable RTC_DRV_PL031

# 4. Build
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- -j$(nproc)

# 5. Output
ls -lh arch/arm64/boot/Image
```

---

## Common Questions & Quick Answers

### Q: What kernel version should I use?

**A** (from agent's recommendations):
- **Production**: Linux 5.10.240 (Slicer's choice - proven stable)
- **Latest LTS**: Linux 6.1.x (Firecracker official support until 2026)
- **Avoid**: 6.6+ until extensively tested

### Q: Why is my kernel 50MB instead of 8MB?

**A** (agent's optimization tips):
```bash
# Check if debug enabled (common culprit)
grep CONFIG_DEBUG .config

# Disable debug
scripts/config --disable DEBUG_KERNEL
scripts/config --disable DEBUG_INFO

# Enable best compression
scripts/config --enable KERNEL_XZ

# Rebuild
make -j$(nproc)
```

### Q: `/dev/vda not found` error - what's wrong?

**A** (agent's troubleshooting):
```bash
# Most common cause: VIRTIO_BLK built as module (=m) instead of built-in (=y)

# Fix:
scripts/config --enable VIRTIO_BLK
scripts/config --disable MODULES  # Force built-in drivers
make -j$(nproc)

# Verify:
grep "CONFIG_VIRTIO_BLK=y" .config
```

### Q: How do I make boot faster?

**A** (agent's optimization workflow):
```bash
# 1. Disable unnecessary drivers (biggest impact)
scripts/config --disable USB
scripts/config --disable SOUND
scripts/config --disable WLAN

# 2. Build monolithic
scripts/config --disable MODULES

# 3. Optimize cmdline
# Use: quiet loglevel=1 nomodule

# 4. Measure with:
# Add to cmdline: initcall_debug
# Then check: dmesg | grep "initcall"
```

---

## Integration Examples

### Example 1: Dockerfile Build

```dockerfile
FROM ubuntu:24.04

# Install build tools (from agent's dependencies)
RUN apt-get update && apt-get install -y \
    build-essential flex bison bc libssl-dev libelf-dev wget

# Download kernel
RUN wget https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz && \
    tar xf linux-6.1.tar.xz

WORKDIR /linux-6.1

# Use minimal config from agent
RUN make tinyconfig && \
    scripts/config --enable SERIAL_8250_CONSOLE && \
    scripts/config --enable VIRTIO_BLK && \
    scripts/config --enable EXT4_FS && \
    scripts/config --enable KVM_GUEST && \
    make -j$(nproc)

# Extract kernel
RUN cp arch/x86/boot/bzImage /vmlinuz
```

### Example 2: GitHub Actions

```yaml
name: Build Kernel
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build Minimal Kernel
        run: |
          # Following .agents/microvm-kernel-expert.md workflow
          wget https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz
          tar xf linux-6.1.tar.xz && cd linux-6.1

          # Minimal config from agent
          make tinyconfig
          scripts/config --enable SERIAL_8250_CONSOLE
          scripts/config --enable VIRTIO_BLK
          scripts/config --enable KVM_GUEST

          make -j$(nproc)
          ls -lh arch/x86/boot/bzImage

      - name: Upload Kernel
        uses: actions/upload-artifact@v4
        with:
          name: vmlinuz
          path: linux-6.1/arch/x86/boot/bzImage
```

### Example 3: Makefile Integration

```makefile
# Reference: .agents/microvm-kernel-expert.md

KERNEL_VERSION = 6.1
KERNEL_URL = https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-$(KERNEL_VERSION).tar.xz

.PHONY: kernel-x86_64 kernel-arm64

kernel-x86_64:
	wget $(KERNEL_URL)
	tar xf linux-$(KERNEL_VERSION).tar.xz
	cd linux-$(KERNEL_VERSION) && \
		make tinyconfig && \
		scripts/config --enable SERIAL_8250_CONSOLE && \
		scripts/config --enable VIRTIO_BLK && \
		scripts/config --enable KVM_GUEST && \
		make -j$$(nproc)
	cp linux-$(KERNEL_VERSION)/arch/x86/boot/bzImage vmlinuz-x86_64

kernel-arm64:
	# Cross-compile following agent's workflow
	cd linux-$(KERNEL_VERSION) && \
		make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- defconfig && \
		scripts/config --enable VIRTIO_MMIO && \
		scripts/config --enable ARM_AMBA && \
		make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- -j$$(nproc)
	cp linux-$(KERNEL_VERSION)/arch/arm64/boot/Image vmlinuz-arm64
```

---

## Cheat Sheet

### Essential Commands (from agent)

```bash
# Configuration
make menuconfig                    # Interactive config
scripts/config --enable OPTION     # Enable option
scripts/config --disable OPTION    # Disable option

# Building
make -j$(nproc)                    # Standard build
make ARCH=arm64 CROSS_COMPILE=...  # Cross-compile

# Verification
grep CONFIG_NAME .config           # Check option
ls -lh arch/x86/boot/bzImage       # Check size
scripts/extract-ikconfig vmlinux   # Extract running kernel config

# Testing
qemu-system-x86_64 -kernel bzImage -nographic  # Quick test
```

### Minimal Config Checklist

For Firecracker x86_64:
```bash
CONFIG_SERIAL_8250_CONSOLE=y
CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y
CONFIG_VIRTIO_BLK=y
CONFIG_EXT4_FS=y
CONFIG_KVM_GUEST=y
CONFIG_ACPI=y
CONFIG_PCI=y
```

For Firecracker ARM64:
```bash
CONFIG_SERIAL_8250_CONSOLE=y
CONFIG_VIRTIO_MMIO=y
CONFIG_VIRTIO_BLK=y
CONFIG_EXT4_FS=y
CONFIG_ARM_AMBA=y
CONFIG_RTC_DRV_PL031=y
```

### Production Checklist

- [ ] Kernel size <15MB (ideally <10MB)
- [ ] Boot time <2 seconds
- [ ] Console on ttyS0 working
- [ ] Root filesystem mounts
- [ ] Security options enabled (SECCOMP, etc.)
- [ ] No debug features enabled
- [ ] Tested on target architecture

---

## Getting Help

**From the Agent**:
- Open `.agents/microvm-kernel-expert.md`
- Search for your specific issue (Ctrl+F)
- Check "Common Issues and Troubleshooting" section
- Review relevant workflow

**With Claude Code**:
```
@microvm-kernel-expert [describe your specific issue with details]
```

**Tips for Better Results**:
1. Specify your architecture (x86_64 or ARM64)
2. Mention kernel version
3. Include error messages or symptoms
4. State your goals (size, boot time, features)
5. Describe your environment (host OS, tools)

---

## Next Steps

After mastering the basics:

1. **Optimize for your workload** - Review agent's performance tuning section
2. **Security harden** - Check agent's security hardening checklist
3. **Multi-architecture** - Learn cross-compilation workflows
4. **CI/CD integration** - Automate kernel builds in your pipeline
5. **Custom drivers** - Explore advanced VirtIO customization

**Full Documentation**: `.agents/microvm-kernel-expert.md`

---

*This quick start is designed to get you productive immediately. For comprehensive guidance, consult the full agent documentation.*
