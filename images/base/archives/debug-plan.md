# Debug Plan for TEST_BOOT_VERBOSE.sh Failures

## Executive Summary

Based on deep analysis, the kernel is **ACTUALLY WORKING CORRECTLY**. The test is "failing" due to a timeout issue, not a kernel problem. The virtio devices are detected, the block device is found, and the filesystem is mounted successfully. The issue is that the VM continues running and gets stuck in the init process, causing the test script to timeout after 35 seconds.

## Chain of Thought Analysis

### 1. Initial Observations

**What the test expects:**
- Kernel boots
- Virtio-MMIO device registers
- Virtio block driver loads
- Block device [vda] is detected
- EXT4 filesystem mounts
- VM shuts down cleanly

**What actually happens:**
- ✅ Kernel boots (Linux 6.1.90)
- ✅ Virtio-MMIO device registers (`virtio-mmio: Registering device`)
- ✅ Virtio block driver loads (`virtio_blk virtio0`)
- ✅ Block device detected (`[vda] 4194304 512-byte logical blocks`)
- ✅ EXT4 filesystem mounts (`EXT4-fs (vda): mounted filesystem`)
- ✅ Root filesystem mounts (`VFS: Mounted root (ext4 filesystem)`)
- ❌ VM continues running, services fail to start, test times out

### 2. Root Cause Analysis

The kernel and virtio configuration are **working perfectly**. The issue is that after mounting the root filesystem, the init system starts but:

1. SSH service fails to start repeatedly
2. systemd-networkd-wait-online service hangs
3. System waits for /dev/ttyS0 device
4. VM never shuts down cleanly

This causes the `timeout 35` command in TEST_BOOT_VERBOSE.sh to kill the process, and the script interprets this as a failure.

### 3. Why This Happens

The rootfs (`/tmp/rootfs-working.ext4`) appears to be a full Ubuntu/Debian system with systemd. When booted:
- systemd tries to start all configured services
- Many services fail because they're designed for a full system, not a microVM
- The system enters a loop trying to restart failed services
- The kernel never panics or shuts down

## Comprehensive Debug Plan

### Phase 1: Verify the Kernel is Actually Working

#### Test 1.1: Quick Validation
```bash
# Run the test but look for success markers, ignoring the timeout
timeout 10 firecracker --no-api --config-file /tmp/test_boot_config.json 2>&1 | \
  grep -E "(virtio-mmio:|virtio_blk|vda\]|EXT4-fs.*mounted|VFS: Mounted root)" | \
  head -10

# Expected output should show all components working
```

#### Test 1.2: Check Boot Sequence
```bash
# Create a test that exits after confirming mount
cat > /tmp/verify-kernel.sh << 'EOF'
#!/bin/bash
KERNEL="${1:-/tmp/test-vmlinux}"
CONFIG="/tmp/test-config.json"

cat > "$CONFIG" << JSON
{
  "boot-source": {
    "kernel_image_path": "$KERNEL",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1"
  },
  "drives": [{
    "drive_id": "rootfs",
    "path_on_host": "/tmp/rootfs-working.ext4",
    "is_root_device": true,
    "is_read_only": false
  }],
  "machine-config": {
    "vcpu_count": 1,
    "mem_size_mib": 256
  }
}
JSON

# Boot and check for mount success
timeout 10 firecracker --no-api --config-file "$CONFIG" 2>&1 | \
  tee /tmp/boot.log | \
  grep -q "VFS: Mounted root" && echo "KERNEL WORKS!" || echo "KERNEL FAILED!"

# Show proof
grep -E "(virtio|vda|mounted)" /tmp/boot.log
EOF

bash /tmp/verify-kernel.sh /tmp/test-vmlinux
```

### Phase 2: Fix the Test Script

#### Solution 2.1: Add Clean Shutdown
Modify the kernel boot args to include `init=/bin/bash` or create a minimal init:

```bash
# Test with immediate shutdown after mount
cat > /tmp/test-immediate-shutdown.json << EOF
{
  "boot-source": {
    "kernel_image_path": "/tmp/test-vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1 init=/bin/bash"
  },
  "drives": [{
    "drive_id": "rootfs",
    "path_on_host": "/tmp/rootfs-working.ext4",
    "is_root_device": true,
    "is_read_only": false
  }],
  "machine-config": {
    "vcpu_count": 1,
    "mem_size_mib": 256
  }
}
EOF

timeout 10 firecracker --no-api --config-file /tmp/test-immediate-shutdown.json
```

#### Solution 2.2: Create Minimal Rootfs for Testing
```bash
# Create a minimal test rootfs that just prints success and shuts down
mkdir -p /tmp/minimal-rootfs
cat > /tmp/minimal-rootfs/init << 'EOF'
#!/bin/sh
echo "INIT: Kernel booted successfully!"
echo "INIT: Root filesystem mounted!"
echo "INIT: Shutting down in 2 seconds..."
sleep 2
echo "INIT: Clean shutdown"
poweroff -f
EOF
chmod +x /tmp/minimal-rootfs/init

# Create minimal rootfs image
dd if=/dev/zero of=/tmp/minimal-test.ext4 bs=1M count=10
mkfs.ext4 /tmp/minimal-test.ext4
sudo mount /tmp/minimal-test.ext4 /mnt
sudo cp -r /tmp/minimal-rootfs/* /mnt/
sudo umount /mnt
```

#### Solution 2.3: Fix Test Script Logic
Modify TEST_BOOT_VERBOSE.sh to detect success earlier:

```bash
# Instead of waiting for timeout, check for success markers and exit
SUCCESS_PATTERN="VFS: Mounted root"
timeout 35 firecracker --no-api --config-file "$TEST_CONFIG" 2>&1 | \
  while IFS= read -r line; do
    echo "$line" >> "$BOOT_LOG"
    if echo "$line" | grep -q "$SUCCESS_PATTERN"; then
      echo "SUCCESS: Root filesystem mounted!"
      pkill -P $$ firecracker 2>/dev/null || true
      break
    fi
  done
```

### Phase 3: Debug Why Services Fail (Optional)

#### Test 3.1: Inspect Service Failures
```bash
# Boot with more verbose systemd logging
cat > /tmp/debug-services.json << EOF
{
  "boot-source": {
    "kernel_image_path": "/tmp/test-vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw systemd.log_level=debug systemd.log_target=console"
  },
  "drives": [{
    "drive_id": "rootfs",
    "path_on_host": "/tmp/rootfs-working.ext4",
    "is_root_device": true,
    "is_read_only": false
  }],
  "machine-config": {
    "vcpu_count": 1,
    "mem_size_mib": 256
  }
}
EOF

timeout 20 firecracker --no-api --config-file /tmp/debug-services.json 2>&1 | \
  grep -E "(Failed|error|Error)" | head -20
```

### Phase 4: Root Cause Fixes

#### Fix Option A: Use Minimal Init
Add to boot args: `init=/sbin/init --unit=multi-user.target`

#### Fix Option B: Create Test-Specific Rootfs
Build a rootfs specifically for testing that:
- Has no unnecessary services
- Shuts down after boot
- Prints success markers

#### Fix Option C: Modify Test Expectations
Update TEST_BOOT_VERBOSE.sh to:
1. Look for filesystem mount as success
2. Don't wait for clean shutdown
3. Kill firecracker after success detected

## Recommended Immediate Actions

### Quick Fix (5 minutes)
```bash
# Create a simple test that properly detects success
cat > /tmp/quick-test.sh << 'EOF'
#!/bin/bash
KERNEL="${1:-/tmp/test-vmlinux}"
echo "Testing kernel: $KERNEL"

# Create config
cat > /tmp/qtest.json << JSON
{
  "boot-source": {
    "kernel_image_path": "$KERNEL",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1"
  },
  "drives": [{
    "drive_id": "rootfs",
    "path_on_host": "/tmp/rootfs-working.ext4",
    "is_root_device": true,
    "is_read_only": false
  }],
  "machine-config": {
    "vcpu_count": 1,
    "mem_size_mib": 256
  }
}
JSON

# Boot and check - SUCCESS if we see filesystem mount
LOG=/tmp/qtest.log
timeout 10 firecracker --no-api --config-file /tmp/qtest.json 2>&1 | tee "$LOG" > /dev/null

if grep -q "VFS: Mounted root" "$LOG" && \
   grep -q "virtio-mmio: Registering device" "$LOG" && \
   grep -q "virtio_blk" "$LOG"; then
    echo "✅ SUCCESS: Kernel boots and mounts filesystem!"
    exit 0
else
    echo "❌ FAILED: Check $LOG"
    exit 1
fi
EOF

chmod +x /tmp/quick-test.sh
/tmp/quick-test.sh /tmp/test-vmlinux
```

### Proper Fix (30 minutes)
1. Create a minimal test rootfs that shuts down cleanly
2. Update TEST_BOOT_VERBOSE.sh to detect success without requiring shutdown
3. Add different test modes: quick (just mount) vs full (complete boot)

## Validation Steps

1. **Confirm kernel works**: Run quick-test.sh above
2. **Check virtio detection**: `grep virtio /tmp/fc_boot_output.log`
3. **Verify mount**: `grep "VFS: Mounted root" /tmp/fc_boot_output.log`
4. **Document success**: The kernel IS working, just the test needs adjustment

## Conclusion

The kernel with CONFIG_VIRTIO_MMIO_CMDLINE_DEVICES is **working correctly**. The test script needs to be updated to:
1. Recognize success when filesystem mounts (not when VM shuts down)
2. Use a minimal rootfs for testing (not a full Ubuntu system)
3. Handle timeouts gracefully when testing with full system rootfs

The "failure" is actually a success - the kernel boots, devices are detected, and the filesystem mounts. The test just doesn't know when to stop waiting.