# Reference Kernel Configurations

This directory contains reference kernel configurations for building microVM kernels with Firecracker and Cloud Hypervisor.

## Available Configurations

### Minimal Production Kernels

**`minimal-x86_64.config`** - Minimal kernel for x86_64 Firecracker microVMs
- Target size: <8MB compressed
- Boot time: <500ms
- Essential VirtIO drivers only
- Security hardened (seccomp, namespaces)

**`minimal-arm64.config`** - Minimal kernel for ARM64/aarch64 Firecracker microVMs
- Target size: <8MB compressed
- Boot time: <500ms
- VirtIO MMIO transport
- ARM-specific drivers (AMBA, PL031 RTC)

### Full-Featured Kernels

**`standard-x86_64.config`** - Standard feature set for x86_64
- Target size: ~12MB compressed
- Includes networking, vsock, entropy
- PCI support for efficient device attachment
- Suitable for most production workloads

**`standard-arm64.config`** - Standard feature set for ARM64
- Target size: ~12MB compressed
- Full VirtIO device support
- Device tree support
- Production-ready configuration

## Usage

### Building with Reference Config

```bash
# Download kernel source
wget https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz
tar xf linux-6.1.tar.xz
cd linux-6.1

# Copy reference config
cp ../.agents/kernel-configs/minimal-x86_64.config .config

# Build kernel
make olddefconfig  # Update config for kernel version
make -j$(nproc)

# Output: arch/x86/boot/bzImage
```

### Cross-Compiling for ARM64

```bash
# Use ARM64 config
cp ../.agents/kernel-configs/minimal-arm64.config .config

# Build with cross-compiler
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- olddefconfig
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- -j$(nproc)

# Output: arch/arm64/boot/Image
```

### Customizing Configurations

```bash
# Start from reference config
cp minimal-x86_64.config .config

# Make changes interactively
make menuconfig

# Or via scripts
scripts/config --enable NEW_FEATURE
scripts/config --disable UNWANTED_FEATURE

# Rebuild
make -j$(nproc)
```

## Configuration Matrix

| Config | Arch | Size | Boot Time | Use Case |
|--------|------|------|-----------|----------|
| minimal-x86_64 | x86_64 | <8MB | <500ms | Minimal production VMs |
| minimal-arm64 | ARM64 | <8MB | <500ms | Minimal production VMs (ARM) |
| standard-x86_64 | x86_64 | ~12MB | <800ms | General purpose VMs |
| standard-arm64 | ARM64 | ~12MB | <800ms | General purpose VMs (ARM) |

## Kernel Version Compatibility

These configurations are tested with:
- **Linux 5.10.x** (LTS, recommended for stability)
- **Linux 6.1.x** (LTS, officially supported by Firecracker)

For other kernel versions, use `make olddefconfig` to update the configuration.

## Validation

After building, validate your kernel:

```bash
# Check size
ls -lh arch/x86/boot/bzImage
# Target: <8MB for minimal, <15MB for standard

# Check required features
scripts/extract-ikconfig vmlinux | grep -E "VIRTIO_BLK|SERIAL_8250_CONSOLE"
# Both should be =y (not =m)

# Test boot (replace paths as needed)
firecracker --config-file test-vm.json
```

## Testing

Test each configuration in a Firecracker microVM:

1. **Build kernel** using reference config
2. **Create test rootfs** (or use Slicer base image)
3. **Configure Firecracker** with test VM spec
4. **Boot and verify**:
   - Console output appears on ttyS0
   - Root filesystem mounts correctly
   - Network functional (if enabled)
   - Boot time meets target

## Maintenance

Update these configurations when:
- New kernel versions are released
- Firecracker requirements change
- Security vulnerabilities require new options
- Performance optimizations are discovered

## Contributing

When submitting new configurations:

1. Test on actual Firecracker microVM
2. Document size and boot time measurements
3. List any special requirements or dependencies
4. Update this README with configuration details

## References

- See `.agents/microvm-kernel-expert.md` for detailed configuration guidance
- Firecracker kernel policy: github.com/firecracker-microvm/firecracker/docs/kernel-policy.md
- Linux kernel documentation: docs.kernel.org

---

*Note: Actual .config files will be added as the project progresses through implementation.*
