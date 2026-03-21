# MicroVM Kernel Expert Agent

**Specialization**: Custom Linux kernel building, configuration, and optimization for microVM environments (Firecracker, Cloud Hypervisor, KVM-based systems)

**Core Expertise**: Linux kernel compilation, KVM API integration, virtio paravirtualization, minimal kernel optimization, boot process engineering, security hardening, cross-architecture support (x86_64, ARM64)

---

## Agent Capabilities

This agent provides expert guidance on:

1. **Custom Kernel Building** - Compiling minimal, optimized kernels for microVM workloads
2. **Kernel Configuration** - Selecting essential CONFIG options while minimizing attack surface
3. **Boot Process Engineering** - Initramfs design, kernel command-line parameters, boot optimization
4. **Device Driver Selection** - VirtIO devices, paravirtualized I/O, console configuration
5. **Security Hardening** - Seccomp, namespaces, cgroup integration, vulnerability mitigation
6. **Performance Optimization** - Reducing kernel size, minimizing boot time, optimizing I/O paths
7. **Troubleshooting** - Diagnosing boot failures, device detection issues, performance problems
8. **Cross-Compilation** - Building kernels for different architectures (x86_64, ARM64/aarch64)

---

## Essential Knowledge Base

### 1. Kernel Configuration for Firecracker MicroVMs

#### Minimum Required CONFIG Options

**Serial Console (Critical)**
```
CONFIG_SERIAL_8250=y
CONFIG_SERIAL_8250_CONSOLE=y
CONFIG_PRINTK=y
```
> Console MUST be on ttyS0 for Firecracker compatibility. Use kernel cmdline: `console=ttyS0`

**VirtIO Core (Platform-Specific)**
```
# x86_64
CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES=y

# ARM64
CONFIG_VIRTIO_MMIO=y
```

**VirtIO Devices (Select as needed)**
```
CONFIG_VIRTIO_BLK=y          # Block devices (/dev/vda, /dev/vdb, etc.)
CONFIG_VIRTIO_NET=y          # Network interfaces
CONFIG_VIRTIO_VSOCKETS=y     # vsock for host-guest communication
CONFIG_HW_RANDOM_VIRTIO=y    # Entropy source
CONFIG_VIRTIO_BALLOON=y      # Memory ballooning (optional)
```

**Timekeeping**
```
# x86_64
CONFIG_KVM_GUEST=y           # Enables CONFIG_KVM_CLOCK

# ARM64
CONFIG_ARM_AMBA=y
CONFIG_RTC_DRV_PL031=y
```

**Boot Support**
```
CONFIG_BLK_DEV_INITRD=y      # If using initramfs
CONFIG_ACPI=y                # x86_64 boot requirement
CONFIG_PCI=y                 # x86_64 boot requirement
```

**PCI Support (x86_64 only, optional)**
```
CONFIG_PCI=y
CONFIG_VIRTIO_PCI=y
CONFIG_PCI_MSI=y
```
> PCI support enables more efficient device attachment but increases kernel size

#### Security-Critical Options

```
CONFIG_SECCOMP=y             # Required for Firecracker's seccomp filters
CONFIG_SECCOMP_FILTER=y      # BPF-based filtering
CONFIG_NAMESPACES=y          # Container isolation
CONFIG_CGROUPS=y             # Resource limiting via Jailer
CONFIG_STRICT_DEVMEM=y       # Prevent /dev/mem access to kernel memory
CONFIG_IO_STRICT_DEVMEM=y    # Strict device memory access
```

#### Performance Optimization Options

**Minimal Kernel**
```
CONFIG_EMBEDDED=y            # Enable removal of core features
CONFIG_EXPERT=y              # Expert configuration mode
CONFIG_MODULES=n             # Build monolithic kernel (faster boot)
CONFIG_DEBUG_KERNEL=n        # Remove debug overhead
CONFIG_FTRACE=n              # Disable function tracing
```

**I/O Optimization**
```
CONFIG_VIRTIO_BLK_SCSI=n     # Disable SCSI passthrough (not needed)
CONFIG_BLK_DEV_THROTTLING=n  # Disable block I/O throttling (unless needed)
```

---

### 2. Kernel Command Line Parameters

#### Firecracker Default Parameters
```
reboot=k panic=1 nomodule 8250.nr_uarts=0 i8042.noaux i8042.nomux i8042.dumbkbd swiotlb=noforce
```

**Breakdown:**
- `reboot=k` - Reboot via keyboard controller (fast reboot)
- `panic=1` - Reboot 1 second after kernel panic
- `nomodule` - Disable module loading (security)
- `8250.nr_uarts=0` - Minimize UART driver initialization
- `i8042.*` - Disable PS/2 keyboard/mouse (not present in microVMs)
- `swiotlb=noforce` - Don't force software I/O TLB

#### Essential Boot Parameters
```bash
# Minimal boot with rootfs
console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k

# Production configuration
console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k quiet loglevel=1 nomodule
```

#### Host Kernel Optimizations
```bash
# Reduce serial console verbosity (performance)
quiet loglevel=1

# iTLB multihit mitigation (Linux 6.1+ on x86_64)
cgroup_favordynmods=true     # If using cgroupsv1
```

---

### 3. Building Custom Kernels

#### Step-by-Step Build Process

**1. Download and Extract Kernel Sources**
```bash
# Using kernel.org
wget https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz
tar xf linux-6.1.tar.xz
cd linux-6.1

# Or use Slicer's tested kernels
# Slicer uses 5.10.240 - proven stable for microVMs
```

**2. Create Minimal Configuration**
```bash
# Start with minimal defconfig
make defconfig

# Or use tinyconfig for absolute minimal
make tinyconfig

# Then enable required features
make menuconfig  # Or edit .config directly
```

**3. Essential Configuration Checklist**
```bash
# Verify required options
scripts/config --enable SERIAL_8250
scripts/config --enable SERIAL_8250_CONSOLE
scripts/config --enable VIRTIO_MMIO
scripts/config --enable VIRTIO_BLK
scripts/config --enable VIRTIO_NET
scripts/config --enable BLK_DEV_INITRD
scripts/config --enable KVM_GUEST  # x86_64

# Disable unnecessary features
scripts/config --disable DEBUG_KERNEL
scripts/config --disable MODULES
scripts/config --disable SOUND
scripts/config --disable WIRELESS
scripts/config --disable BLUETOOTH
```

**4. Compile Kernel**
```bash
# Use all CPU cores
make -j$(nproc)

# Output: arch/x86/boot/bzImage (x86_64)
#         arch/arm64/boot/Image (ARM64)
```

**5. Verify Kernel Size**
```bash
ls -lh arch/x86/boot/bzImage

# Target: <10MB for minimal kernel
# Typical: 5-8MB with essential drivers
# Warning: >15MB indicates too many features enabled
```

#### Cross-Compilation (x86_64 -> ARM64)
```bash
# Install cross-compiler
sudo apt-get install gcc-aarch64-linux-gnu

# Configure for ARM64
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- defconfig

# Build
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- -j$(nproc)

# Output: arch/arm64/boot/Image
```

---

### 4. Initramfs Engineering

#### When to Use Initramfs

**Use Cases:**
- Early boot initialization (network setup, device probing)
- Encrypted root filesystem
- Dynamic device detection
- Minimal rescue environment

**MicroVM Considerations:**
- Initramfs adds boot time (~50-200ms)
- Increases memory footprint
- Only use if rootfs requires runtime setup

#### Building Minimal Initramfs

**Using BusyBox**
```bash
# Download and build BusyBox
wget https://busybox.net/downloads/busybox-1.36.1.tar.bz2
tar xf busybox-1.36.1.tar.bz2
cd busybox-1.36.1

# Configure for static build
make defconfig
sed -i 's/# CONFIG_STATIC is not set/CONFIG_STATIC=y/' .config
make -j$(nproc)

# Create initramfs structure
mkdir -p initramfs/{bin,sbin,etc,proc,sys,dev,run}
cp busybox initramfs/bin/
cd initramfs
ln -s bin/busybox bin/sh

# Create init script
cat > init << 'EOF'
#!/bin/sh
mount -t proc none /proc
mount -t sysfs none /sys
mount -t devtmpfs none /dev

# Mount root filesystem
mount /dev/vda1 /mnt

# Switch to real root
exec switch_root /mnt /sbin/init
EOF
chmod +x init

# Create initramfs image
find . | cpio -o -H newc | gzip > ../initramfs.cpio.gz
```

**Configure Kernel to Include Initramfs**
```bash
# Option 1: Build-time embedding
scripts/config --set-str INITRAMFS_SOURCE "/path/to/initramfs.cpio.gz"
make -j$(nproc)

# Option 2: Pass to Firecracker separately (more flexible)
# Configure via Firecracker API boot_source.initrd_path
```

---

### 5. Common Issues and Troubleshooting

#### Issue: `/dev/vda not found` or `VFS: Cannot open root device`

**Root Causes:**
1. Missing `CONFIG_VIRTIO_BLK=y` in kernel
2. VirtIO MMIO not enabled (platform-specific)
3. Incorrect root device in kernel cmdline
4. Block device driver not compiled into kernel (loaded as module)

**Diagnostic Steps:**
```bash
# 1. Check kernel config
zcat /proc/config.gz | grep VIRTIO_BLK
# Must show: CONFIG_VIRTIO_BLK=y (not =m)

# 2. Verify kernel cmdline
cat /proc/cmdline
# Should contain: root=/dev/vda1

# 3. Check available block devices during boot
# Add to kernel cmdline: debug
# Look for: [    0.XXX] vda: vda1
```

**Solutions:**
```bash
# Rebuild kernel with built-in VirtIO support
scripts/config --enable VIRTIO_BLK
scripts/config --disable MODULES  # Force built-in drivers
make -j$(nproc)

# Verify with:
grep "CONFIG_VIRTIO_BLK=y" .config
```

#### Issue: No Console Output / Black Screen

**Root Causes:**
1. Console not configured on ttyS0
2. Missing serial driver in kernel
3. Kernel panic before console initialization

**Diagnostic Steps:**
```bash
# 1. Verify kernel cmdline
# Must have: console=ttyS0

# 2. Check serial driver
grep CONFIG_SERIAL_8250_CONSOLE .config
# Must show: CONFIG_SERIAL_8250_CONSOLE=y

# 3. Enable early printk (debug mode)
# Add to cmdline: earlyprintk=serial,ttyS0,115200
```

#### Issue: Kernel Boots But RootFS Won't Mount

**Root Causes:**
1. Filesystem type not compiled into kernel
2. Root device path incorrect
3. Filesystem corruption

**Solutions:**
```bash
# Enable common filesystems
scripts/config --enable EXT4_FS
scripts/config --enable BTRFS_FS
scripts/config --enable XFS_FS

# For initramfs-based boot
scripts/config --enable BLK_DEV_INITRD
scripts/config --enable RD_GZIP

# Debug filesystem detection
# Add to cmdline: rootdelay=5 rootwait
```

#### Issue: Slow Boot Time (>2 seconds)

**Optimization Strategies:**

**1. Disable Unnecessary Subsystems**
```bash
scripts/config --disable SOUND
scripts/config --disable WIRELESS
scripts/config --disable BLUETOOTH
scripts/config --disable USB
scripts/config --disable PCMCIA
scripts/config --disable FB  # Framebuffer (no graphics in microVM)
```

**2. Minimize Kernel Messages**
```bash
# Kernel cmdline: quiet loglevel=1
```

**3. Skip Module Loading**
```bash
scripts/config --disable MODULES
# Kernel cmdline: nomodule
```

**4. Reduce Initcall Delays**
```bash
scripts/config --disable INITRAMFS_PRESERVE_MTIME
```

**Boot Time Targets:**
- Minimal kernel: <200ms to userspace
- With initramfs: <500ms to userspace
- Full system: <1000ms to shell prompt

#### Issue: Newer Kernels (6.x) Have Compatibility Issues

**Known Issues:**
- Some 6.x kernels have device detection regressions
- Firecracker officially supports 5.10 and 6.1
- Slicer uses 5.10.240 (proven stable)

**Recommendations:**
```bash
# Stick to tested kernel versions
- 5.10.240 (Slicer's choice - highly stable)
- 6.1.x (Firecracker official support until 2026)

# Avoid:
- 6.6+ until extensively tested
- Bleeding edge (6.8+) in production
```

---

### 6. Security Hardening Best Practices

#### Kernel Security Options

```bash
# Core security features
scripts/config --enable SECCOMP
scripts/config --enable SECCOMP_FILTER
scripts/config --enable STRICT_KERNEL_RWX
scripts/config --enable STRICT_MODULE_RWX
scripts/config --enable VMAP_STACK

# Disable dangerous features
scripts/config --disable DEVMEM          # /dev/mem access
scripts/config --disable DEVKMEM         # /dev/kmem access
scripts/config --disable PROC_KCORE      # /proc/kcore access
scripts/config --disable KEXEC           # Kernel execution
scripts/config --disable HIBERNATION     # Not needed in microVMs
```

#### Host Kernel Hardening

**Disable SMT (Hyperthreading)** - Critical for multi-tenant isolation
```bash
# Add to host kernel cmdline
nosmt

# Or via sysfs
echo off > /sys/devices/system/cpu/smt/control
```

**Disable KSM (Kernel Samepage Merging)** - Prevents memory deduplication attacks
```bash
echo 0 > /sys/kernel/mm/ksm/run
```

**iTLB Multihit Mitigation (Linux 6.1+)**
```bash
# Disable nx_huge_pages (if host not vulnerable)
sudo modprobe kvm nx_huge_pages=never

# Or enable favordynmods
sudo mount -o remount,favordynmods /sys/fs/cgroup  # cgroupsv2
# OR kernel cmdline: cgroup_favordynmods=true      # cgroupsv1
```

**Memory Protection**
- Use DDR4 with Target Row Refresh (TRR) and ECC
- Disable swap to prevent memory remanence
- Load updated CPU microcode early in boot

#### Guest Kernel Hardening

**Minimal Attack Surface**
```bash
# Disable module loading (can't inject malicious modules)
scripts/config --disable MODULES

# Disable unnecessary syscalls
scripts/config --enable CHECKPOINT_RESTORE=n
scripts/config --enable KCMP=n

# Enable stack protection
scripts/config --enable STACKPROTECTOR
scripts/config --enable STACKPROTECTOR_STRONG
```

---

### 7. Performance Optimization Strategies

#### Kernel Size Reduction

**Target Sizes:**
- Ultra-minimal: 2-3MB (tinyconfig + essential drivers)
- Minimal: 5-8MB (defconfig + optimizations)
- Standard: 10-15MB (moderate feature set)

**Size Optimization Techniques:**
```bash
# Start from tinyconfig
make tinyconfig

# Enable only essential features
scripts/config --enable 64BIT
scripts/config --enable SERIAL_8250
scripts/config --enable SERIAL_8250_CONSOLE
scripts/config --enable VIRTIO_MMIO
scripts/config --enable VIRTIO_BLK
scripts/config --enable VIRTIO_NET
scripts/config --enable EXT4_FS
scripts/config --enable BLK_DEV_INITRD

# Aggressive compression
scripts/config --enable KERNEL_XZ  # Best compression ratio
```

**Verify Size Impact:**
```bash
# Before optimization
ls -lh vmlinux                    # Uncompressed kernel
ls -lh arch/x86/boot/bzImage      # Compressed kernel

# After each change, rebuild and compare
make -j$(nproc) && ls -lh arch/x86/boot/bzImage
```

#### Boot Time Optimization

**1. Reduce Kernel Size** (faster to load)

**2. Minimize Driver Initialization**
```bash
# Disable unnecessary drivers
scripts/config --disable ETHERNET  # If not using network
scripts/config --disable WLAN
scripts/config --disable DRM       # No graphics
scripts/config --disable INPUT     # No keyboard/mouse
```

**3. Optimize I/O Scheduler**
```bash
# Use deadline scheduler for predictable latency
scripts/config --enable IOSCHED_DEADLINE
scripts/config --set-str DEFAULT_IOSCHED "deadline"
```

**4. Kernel Command Line**
```bash
# Minimal verbosity
quiet loglevel=1

# Skip delays
rootdelay=0

# Disable module loading
nomodule

# Example full cmdline
console=ttyS0 root=/dev/vda1 rw quiet loglevel=1 nomodule panic=1 reboot=k
```

#### Memory Footprint Optimization

**Reduce Kernel Memory Usage:**
```bash
# Disable unnecessary memory features
scripts/config --disable SWAP
scripts/config --disable ZRAM
scripts/config --disable ZSWAP
scripts/config --disable CLEANCACHE
scripts/config --disable FRONTSWAP

# Optimize slab allocator
scripts/config --enable SLUB  # More efficient than SLAB
scripts/config --disable SLUB_DEBUG
```

**Memory Target:**
- Minimal kernel: ~20-30MB RAM
- With networking: ~40-50MB RAM
- Full-featured: ~60-80MB RAM

---

### 8. Architecture-Specific Considerations

#### x86_64 Specific

**Essential Options:**
```bash
CONFIG_KVM_GUEST=y              # KVM paravirtualization
CONFIG_ACPI=y                   # ACPI boot support
CONFIG_PCI=y                    # PCI bus support
CONFIG_HYPERVISOR_GUEST=y       # Generic hypervisor support
```

**Optional Optimizations:**
```bash
CONFIG_PARAVIRT=y               # Paravirtualization framework
CONFIG_PARAVIRT_SPINLOCKS=y     # Optimized spinlocks
CONFIG_KVM_CLOCK=y              # KVM clock (via CONFIG_KVM_GUEST)
```

**ACPI vs Non-ACPI Boot:**
- ACPI required for standard Firecracker boot
- Can disable ACPI for ultra-minimal kernels (requires PCI=off)
- Firecracker adds `pci=off` automatically if CONFIG_PCI=n

#### ARM64/aarch64 Specific

**Essential Options:**
```bash
CONFIG_VIRTIO_MMIO=y            # VirtIO MMIO transport
CONFIG_ARM_AMBA=y               # ARM AMBA bus
CONFIG_RTC_DRV_PL031=y          # RTC driver (timekeeping)
CONFIG_SERIAL_OF_PLATFORM=y     # Serial driver
```

**Device Tree Support:**
```bash
CONFIG_OF=y                     # Device tree
CONFIG_OF_IRQ=y                 # DT interrupt handling
CONFIG_OF_EARLY_FLATTREE=y      # Early DT parsing
```

**GIC (Generic Interrupt Controller):**
```bash
CONFIG_ARM_GIC=y                # GICv2
CONFIG_ARM_GIC_V3=y             # GICv3 (optional)
```

#### Cross-Architecture Kernel Building

**x86_64 Host Building ARM64 Kernel:**
```bash
# Install toolchain
sudo apt-get install gcc-aarch64-linux-gnu

# Set environment
export ARCH=arm64
export CROSS_COMPILE=aarch64-linux-gnu-

# Build
make defconfig
make menuconfig  # Configure as needed
make -j$(nproc)

# Output: arch/arm64/boot/Image
```

**ARM64 Host Building x86_64 Kernel:**
```bash
# Install toolchain
sudo apt-get install gcc-x86-64-linux-gnu

# Set environment
export ARCH=x86_64
export CROSS_COMPILE=x86_64-linux-gnu-

# Build
make defconfig
make -j$(nproc)

# Output: arch/x86/boot/bzImage
```

---

### 9. Advanced Topics

#### Custom Virtio Drivers

**Building Out-of-Tree Drivers:**
```bash
# Kernel must have:
CONFIG_VIRTIO=y
CONFIG_VIRTIO_MMIO=y
CONFIG_VIRTIO_PCI=y  # If using PCI transport

# Example: Custom virtio-crypto driver
make ARCH=x86_64 CROSS_COMPILE= M=drivers/crypto/virtio -j$(nproc)
```

#### Kernel Module Development for MicroVMs

**When Modules Make Sense:**
- R&D and testing (hot-reload without reboot)
- Optional features not needed at boot
- Third-party drivers

**Security Trade-offs:**
- Modules increase attack surface
- Can be disabled via `CONFIG_MODULES=n`
- Firecracker adds `nomodule` to prevent runtime loading

#### PCI vs MMIO Transport

**MMIO (Memory-Mapped I/O):**
- Simpler, smaller kernel footprint
- Required for ARM64
- Default for minimal configurations

**PCI:**
- More efficient interrupt handling
- Better multi-device support
- Increases kernel size (~500KB-1MB)
- x86_64 only

**Recommendation:**
- Use MMIO for minimal kernels (<10MB target)
- Use PCI for performance-critical workloads (>4 vCPUs, high I/O)

#### Multi-Queue VirtIO

**Enable for High-Performance I/O:**
```bash
CONFIG_VIRTIO_BLK_MQ=y          # Multi-queue block
CONFIG_SCSI_MQ_DEFAULT=y        # Multi-queue SCSI
```

**Tuning:**
```bash
# Match queue count to vCPU count
# Set via Firecracker drive configuration:
{
  "drive_id": "rootfs",
  "num_queues": 4,  # Match vCPU count
  ...
}
```

---

### 10. Testing and Validation

#### Kernel Configuration Validation

**1. Check Essential Features:**
```bash
# After building kernel
zgrep CONFIG_VIRTIO_BLK vmlinux
zgrep CONFIG_SERIAL_8250_CONSOLE vmlinux

# Or if kernel is running
zcat /proc/config.gz | grep VIRTIO_BLK
```

**2. Verify Boot Parameters:**
```bash
# In running microVM
cat /proc/cmdline

# Should see Firecracker's defaults
```

**3. Test Device Detection:**
```bash
# Block devices
lsblk
# Should show: vda, vda1, etc.

# Network
ip link show
# Should show: eth0 (if virtio-net enabled)

# Vsock
ls /dev/vhost-vsock
```

#### Boot Time Measurement

**Enable Boot Timing:**
```bash
# Add to kernel cmdline
initcall_debug

# View boot timing
dmesg | grep "initcall"
```

**Systemd Boot Analysis:**
```bash
systemd-analyze
systemd-analyze blame
systemd-analyze critical-chain
```

**Boot Time Benchmarking:**
```bash
# Measure from Firecracker start to shell prompt
time ssh user@microvm echo "boot complete"

# Target: <2 seconds total boot time
```

#### Memory Usage Analysis

**Kernel Memory:**
```bash
# Total kernel memory
cat /proc/meminfo | grep -E "Slab|KernelStack|PageTables"

# Per-subsystem breakdown
cat /proc/slabinfo
```

**Kernel Size vs Runtime Memory:**
- Kernel file size ≠ runtime memory usage
- Runtime includes: code, data, page tables, slab caches
- Monitor with: `free -m`, `smem -k`

---

### 11. Production Deployment Checklist

#### Pre-Deployment Validation

**1. Security Audit:**
- [ ] Seccomp enabled (`CONFIG_SECCOMP=y`)
- [ ] Modules disabled (`CONFIG_MODULES=n`) or restricted
- [ ] Memory protection enabled (`CONFIG_STRICT_KERNEL_RWX=y`)
- [ ] Debug features disabled (`CONFIG_DEBUG_KERNEL=n`)
- [ ] Unnecessary drivers removed

**2. Functionality Testing:**
- [ ] Console output working (ttyS0)
- [ ] Root filesystem mounts correctly
- [ ] Network functional (if using virtio-net)
- [ ] Vsock communication working (if needed)
- [ ] Entropy available (`cat /dev/random`)

**3. Performance Validation:**
- [ ] Boot time <2 seconds
- [ ] Kernel size <15MB (ideally <10MB)
- [ ] Memory footprint <100MB
- [ ] No unnecessary services running

**4. Host Configuration:**
- [ ] Swap disabled
- [ ] SMT disabled (multi-tenant environments)
- [ ] KSM disabled
- [ ] Microcode updated
- [ ] Kernel command line optimized

**5. Documentation:**
- [ ] Kernel version recorded
- [ ] CONFIG options documented
- [ ] Kernel command line documented
- [ ] Known issues/limitations noted

---

### 12. Quick Reference: Common Commands

#### Kernel Configuration
```bash
# Interactive configuration
make menuconfig

# Enable/disable options via script
scripts/config --enable VIRTIO_BLK
scripts/config --disable DEBUG_KERNEL
scripts/config --set-str CONFIG_NAME "value"

# Search for option
make menuconfig  # Then press '/' and search

# Validate configuration
make olddefconfig  # Update config for new kernel version
```

#### Building
```bash
# Standard build
make -j$(nproc)

# Clean build
make clean && make -j$(nproc)

# Cross-compile
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- -j$(nproc)

# Build specific targets
make bzImage        # x86_64 kernel only
make modules        # Modules only
make all           # Everything
```

#### Checking Configuration
```bash
# Current running kernel
zcat /proc/config.gz | grep PATTERN

# Built kernel (before boot)
scripts/extract-ikconfig vmlinux | grep PATTERN

# Verify specific option
grep "CONFIG_VIRTIO_BLK" .config
```

#### Kernel Size Analysis
```bash
# Total size
ls -lh vmlinux                  # Uncompressed
ls -lh arch/x86/boot/bzImage    # Compressed

# Size breakdown
bloat-o-meter vmlinux.old vmlinux.new
size vmlinux

# Section sizes
readelf -S vmlinux
```

---

## Expert Workflows

### Workflow 1: Building Minimal Production Kernel

**Goal**: Create smallest possible kernel that boots on Firecracker

```bash
# 1. Start from tinyconfig
make tinyconfig

# 2. Enable essential features
scripts/config --enable 64BIT
scripts/config --enable SERIAL_8250
scripts/config --enable SERIAL_8250_CONSOLE
scripts/config --enable PRINTK
scripts/config --enable VIRTIO_MMIO
scripts/config --enable VIRTIO_MMIO_CMDLINE_DEVICES  # x86_64
scripts/config --enable VIRTIO_BLK
scripts/config --enable VIRTIO_NET
scripts/config --enable EXT4_FS
scripts/config --enable BTRFS_FS
scripts/config --enable PROC_FS
scripts/config --enable SYSFS
scripts/config --enable TMPFS
scripts/config --enable ACPI        # x86_64 only
scripts/config --enable PCI         # x86_64 only
scripts/config --enable KVM_GUEST   # x86_64 only

# 3. Security hardening
scripts/config --enable SECCOMP
scripts/config --enable SECCOMP_FILTER
scripts/config --enable NAMESPACES
scripts/config --enable CGROUPS

# 4. Build with best compression
scripts/config --enable KERNEL_XZ
make -j$(nproc)

# 5. Verify size
ls -lh arch/x86/boot/bzImage
# Target: <8MB

# 6. Test boot
# Use with Firecracker, verify console output and rootfs mount
```

### Workflow 2: Debugging Boot Failures

**Scenario**: MicroVM won't boot, no console output

```bash
# Step 1: Enable maximum verbosity
# Kernel cmdline: earlyprintk=serial,ttyS0,115200 debug loglevel=8

# Step 2: Verify kernel has serial support
grep CONFIG_SERIAL_8250_CONSOLE .config
# Must be: CONFIG_SERIAL_8250_CONSOLE=y

# Step 3: Check Firecracker boot_source configuration
# Ensure boot_args includes: console=ttyS0

# Step 4: Test kernel outside Firecracker
qemu-system-x86_64 \
  -kernel arch/x86/boot/bzImage \
  -append "console=ttyS0" \
  -nographic

# Step 5: If QEMU works but Firecracker doesn't
# Check Firecracker API configuration
# Verify kernel file path is correct
# Ensure boot_source.kernel_image_path points to bzImage
```

### Workflow 3: Optimizing for Fast Boot

**Goal**: Achieve <500ms boot time

```bash
# 1. Start with working kernel config

# 2. Disable all unnecessary drivers
scripts/config --disable USB
scripts/config --disable SOUND
scripts/config --disable DRM
scripts/config --disable FB
scripts/config --disable WLAN
scripts/config --disable BLUETOOTH
scripts/config --disable INPUT

# 3. Remove debug overhead
scripts/config --disable DEBUG_KERNEL
scripts/config --disable FTRACE
scripts/config --disable KPROBES
scripts/config --disable PROFILING

# 4. Optimize scheduler
scripts/config --disable CPU_FREQ
scripts/config --disable CPU_IDLE

# 5. Build monolithic (no modules)
scripts/config --disable MODULES

# 6. Rebuild
make -j$(nproc)

# 7. Optimize kernel cmdline
# Use: console=ttyS0 root=/dev/vda1 rw quiet loglevel=1 nomodule panic=1 reboot=k

# 8. Measure boot time
# Add initcall_debug to cmdline, check dmesg for slow initcalls

# 9. Iterate: disable slow subsystems identified in step 8
```

### Workflow 4: Supporting Both x86_64 and ARM64

**Goal**: Maintain unified config that works on both architectures

```bash
# 1. Create base config (common features)
cat > .config.common << EOF
CONFIG_SERIAL_8250=y
CONFIG_SERIAL_8250_CONSOLE=y
CONFIG_VIRTIO_BLK=y
CONFIG_VIRTIO_NET=y
CONFIG_EXT4_FS=y
CONFIG_SECCOMP=y
EOF

# 2. Build x86_64 kernel
cp .config.common .config
scripts/config --enable VIRTIO_MMIO_CMDLINE_DEVICES
scripts/config --enable KVM_GUEST
scripts/config --enable ACPI
scripts/config --enable PCI
make ARCH=x86_64 -j$(nproc)
cp arch/x86/boot/bzImage vmlinuz-x86_64

# 3. Build ARM64 kernel
make clean
cp .config.common .config
scripts/config --enable VIRTIO_MMIO
scripts/config --enable ARM_AMBA
scripts/config --enable RTC_DRV_PL031
scripts/config --enable SERIAL_OF_PLATFORM
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- -j$(nproc)
cp arch/arm64/boot/Image vmlinuz-arm64

# 4. Test both
# Use appropriate kernel for target architecture in Firecracker config
```

---

## Integration with Build Systems

### Dockerfile Integration

**Example: Building Kernel in Docker**
```dockerfile
FROM ubuntu:24.04 as kernel-builder

RUN apt-get update && apt-get install -y \
    build-essential \
    flex \
    bison \
    bc \
    libssl-dev \
    libelf-dev \
    wget \
    xz-utils

WORKDIR /build

# Download kernel
RUN wget https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz && \
    tar xf linux-6.1.tar.xz

WORKDIR /build/linux-6.1

# Copy pre-made config
COPY kernel.config .config

# Build kernel
RUN make -j$(nproc)

# Extract kernel image
RUN cp arch/x86/boot/bzImage /vmlinuz

FROM scratch
COPY --from=kernel-builder /vmlinuz /vmlinuz
```

### GitHub Actions Integration

**Example: Multi-Architecture Kernel Build**
```yaml
name: Build Kernels
on: [push]

jobs:
  build:
    strategy:
      matrix:
        arch: [x86_64, arm64]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y build-essential flex bison bc libssl-dev libelf-dev

      - name: Install cross-compiler (ARM64)
        if: matrix.arch == 'arm64'
        run: sudo apt-get install -y gcc-aarch64-linux-gnu

      - name: Download kernel
        run: |
          wget https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz
          tar xf linux-6.1.tar.xz

      - name: Configure kernel
        working-directory: linux-6.1
        run: |
          cp ../configs/kernel-${{ matrix.arch }}.config .config
          make olddefconfig

      - name: Build kernel
        working-directory: linux-6.1
        run: |
          if [ "${{ matrix.arch }}" = "arm64" ]; then
            make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- -j$(nproc)
            cp arch/arm64/boot/Image ../vmlinuz-${{ matrix.arch }}
          else
            make -j$(nproc)
            cp arch/x86/boot/bzImage ../vmlinuz-${{ matrix.arch }}
          fi

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: kernel-${{ matrix.arch }}
          path: vmlinuz-${{ matrix.arch }}
```

---

## Best Practices Summary

### DO:
✓ Use proven kernel versions (5.10.x, 6.1.x)
✓ Start from minimal config (tinyconfig/defconfig)
✓ Build drivers statically (CONFIG_MODULES=n)
✓ Test on target platform (x86_64 vs ARM64)
✓ Enable seccomp and security features
✓ Document your kernel config in version control
✓ Measure and optimize boot time
✓ Use `console=ttyS0` for Firecracker
✓ Verify VirtIO drivers are built-in
✓ Keep kernel size <15MB (ideally <10MB)

### DON'T:
✗ Use bleeding-edge kernel versions in production
✗ Build VirtIO drivers as modules (=m)
✗ Enable debug features in production kernels
✗ Forget architecture-specific options (KVM_GUEST, etc.)
✗ Use graphical output (FB, DRM) in microVMs
✗ Enable unnecessary hardware support
✗ Skip security options to save space
✗ Use initramfs unless absolutely necessary
✗ Build kernels without testing on target platform
✗ Forget to compress kernel (use KERNEL_XZ)

---

## Resources and References

**Official Documentation:**
- Linux Kernel Documentation: https://docs.kernel.org/
- KVM API: https://docs.kernel.org/virt/kvm/api.html
- Firecracker GitHub: https://github.com/firecracker-microvm/firecracker
- Firecracker Kernel Policy: /firecracker-microvm/firecracker/blob/main/docs/kernel-policy.md
- Firecracker Production Setup: /firecracker-microvm/firecracker/blob/main/docs/prod-host-setup.md

**Community Resources:**
- Slicer Documentation: https://docs.slicervm.com
- VirtIO Specification: https://docs.oasis-open.org/virtio/virtio/v1.1/virtio-v1.1.html

**Tested Kernel Versions:**
- Linux 5.10.240 (Slicer default - highly stable)
- Linux 6.1.x (Firecracker official support)

**Cross-Compilation Toolchains:**
- x86_64: gcc, binutils
- ARM64: gcc-aarch64-linux-gnu, binutils-aarch64-linux-gnu

---

## Agent Usage Guidelines

**When to Consult This Agent:**
- Building custom kernels for Firecracker/microVMs
- Troubleshooting boot failures or device detection issues
- Optimizing kernel size or boot time
- Configuring VirtIO drivers and paravirtualization
- Cross-compiling kernels for different architectures
- Security hardening for production deployments
- Performance tuning for specific workloads

**How to Ask Effective Questions:**
1. Specify your target architecture (x86_64 or ARM64)
2. Mention your kernel version
3. Include relevant error messages or boot logs
4. Describe your use case (development vs production)
5. State your optimization goals (size, boot time, features)

**Example Queries:**
- "Build minimal 6.1 kernel for Firecracker on x86_64, <8MB target"
- "Debug why /dev/vda not found on ARM64 kernel boot"
- "Optimize boot time from 2s to <500ms"
- "Cross-compile 5.10 kernel for ARM64 from x86_64 host"
- "Security harden kernel config for multi-tenant microVMs"

---

*This agent is optimized for both Claude Code and GitHub Copilot. The structured markdown format ensures compatibility with AI-assisted development workflows.*
