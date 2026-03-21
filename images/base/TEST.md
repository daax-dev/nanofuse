# Testing the NanoFuse Base Image

This guide covers testing the built image by booting it in Firecracker.

## Prerequisites

Before testing, you need:

1. **Firecracker installed**
   ```bash
   # Check if installed
   which firecracker

   # If not installed, see docs/QUICKSTART.md for installation
   ```

2. **Built artifacts**
   ```bash
   # Make sure build succeeded
   ls -lh ./build/
   ```

3. **KVM access**
   ```bash
   # Check KVM permission
   [ -r /dev/kvm ] && [ -w /dev/kvm ] && echo "✓ KVM accessible" || echo "✗ Need KVM access"

   # If not accessible, add user to kvm group:
   sudo usermod -aG kvm $USER
   newgrp kvm
   ```

## Running Boot Test

```bash
# Test boot in Firecracker (requires Firecracker installed)
sudo ./test-boot.sh build/vmlinux build/rootfs.ext4
```

This will:
1. Start a VM in Firecracker
2. Wait for it to boot (timeout: 30 seconds)
3. Verify:
   - VM boots successfully
   - Console output on ttyS0
   - systemd reaches multi-user.target
   - SSH daemon running
   - Network configured (systemd-networkd)
   - No failed systemd units
   - Boot time < 2 seconds

## Expected Output

```
======================================
NanoFuse Base Image Boot Test
======================================

Configuration:
  Kernel:  ./build/vmlinux (39M)
  Rootfs:  ./build/rootfs.ext4 (2.0G)
  Socket:  /tmp/nanofuse-test-XXXX.sock
  Timeout: 30s

Test directory: /tmp/tmp.XXXX

Starting Firecracker VM...
Firecracker PID: 12345

Waiting for VM to boot (timeout: 30s)...

✓ VM booted successfully in 1s

======================================
Test Results
======================================

✓ Test 1: VM boots successfully
✓ Test 2: Console output visible on ttyS0
✓ Test 3: systemd reaches multi-user.target
✓ Test 4: SSH daemon running
✓ Test 5: Network configured (systemd-networkd)
✓ Test 6: No failed systemd units detected
✓ Test 7: Boot time < 2s (1s)

======================================
Overall: PASS
```

## What Gets Tested

### 1. Boot Success
- VM starts without errors
- No immediate crashes

### 2. Console Output
- Serial console on ttyS0 is working
- Kernel messages visible

### 3. systemd Initialization
- systemd starts as PID 1
- Reaches multi-user.target

### 4. SSH Service
- OpenSSH server installed
- Service enabled and started
- Listening on port 22

### 5. Networking
- systemd-networkd running
- Should be configured for DHCP

### 6. Service Health
- No failed systemd units
- All configured services started

### 7. Performance
- Boot time measured
- Target: < 2 seconds to multi-user.target

## Test Artifacts

After test completes, artifacts are preserved:

```
/tmp/nanofuse-test-<TIMESTAMP>/
├── vm-config.json      # Firecracker config used
├── console.log         # Full console output
└── (cleanup happens automatically)
```

View console output:

```bash
# Find latest test directory
LATEST_TEST=$(ls -td /tmp/nanofuse-test-* 2>/dev/null | head -1)

# View console output
cat "$LATEST_TEST/console.log"
```

## Manual Testing

If you want to manually boot and test the VM:

```bash
# Create config file
cat > vm-config.json <<'EOF'
{
  "boot-source": {
    "kernel_image_path": "./build/vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "./build/rootfs.ext4",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 2,
    "mem_size_mib": 512,
    "smt": false
  }
}
EOF

# Start Firecracker
sudo firecracker \
  --api-sock /tmp/firecracker.sock \
  --config-file vm-config.json
```

Watch console output - you'll see kernel boot messages and systemd startup. VM will boot to login prompt.

## Common Issues

### "firecracker binary not found"
**Solution:** Install Firecracker from https://github.com/firecracker-microvm/firecracker/releases

### "/dev/kvm: Permission denied"
**Solution:** Run with sudo or add user to kvm group:
```bash
sudo usermod -aG kvm $USER
newgrp kvm
```

### VM doesn't boot / no console output
**Solution:** Check that artifacts are valid:
```bash
file ./build/vmlinux ./build/rootfs.ext4

# Should show:
# vmlinux:     ELF 64-bit LSB executable (Linux kernel)
# rootfs.ext4: Linux ext4 filesystem
```

### Boot is slow (> 5 seconds)
**Possible causes:**
- First boot generates SSH keys (one-time)
- Slow disk I/O
- Check console log for failed services

## Build + Test Workflow

```bash
#!/bin/bash
set -e

cd /home/jpoley/src/_mine/nanofuse/images/base

echo "Building..."
sudo ./build.sh

echo ""
echo "Validating..."
./validate-build.sh

echo ""
echo "Testing boot..."
sudo ./test-boot.sh build/vmlinux build/rootfs.ext4

echo ""
echo "✓ Build and test complete!"
```

## Related Documentation

- **BUILD.md** - Building the image
- **README.md** - Full project documentation
- **docs/QUICKSTART.md** - Quick start (includes Firecracker install)
- **docs/IMPLEMENTATION_NOTES.md** - Design decisions
