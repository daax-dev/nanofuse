# Phase 1 Testing Guide

**Status**: Ready to test!
**Prerequisites**: ✅ Firecracker installed, ✅ Base image built, ✅ KVM available
**Time Required**: ~5-10 minutes

## Quick Status Check

Your system is ready:
- ✅ Firecracker: `/usr/local/bin/firecracker`
- ✅ KVM device: `/dev/kvm` (accessible)
- ✅ Base image: Built (rootfs.ext4 + vmlinux kernel)
- ✅ Test script: `scripts/test-network-e2e.sh`

⚠️ **Note**: You're not in the `kvm` group, so you'll need `sudo` for tests.

## Option 1: Comprehensive End-to-End Test (Recommended)

This tests the entire Phase 1 feature set:

```bash
# Run the full end-to-end network test
sudo ./scripts/test-network-e2e.sh
```

**What it tests:**
- ✅ Daemon startup and initialization
- ✅ Network infrastructure (bridge, NAT, IPAM)
- ✅ VM creation with dynamic IP allocation
- ✅ TAP device creation and cleanup
- ✅ Firecracker VM boot
- ✅ Host-to-VM network connectivity
- ✅ VM console logs

**Expected time**: ~1 minute
**Expected result**: All checks pass with ✓ marks

### Expected Output

```
========================================
NanoFuse Network End-to-End Test
========================================

[1/8] Checking base image files...
✓ Base image files found

[2/8] Setting up test environment...
✓ Test environment ready

[3/8] Starting daemon...
✓ Daemon is running

[4/8] Registering test image in database...
✓ Image registered

[5/8] Verifying network infrastructure...
✓ Bridge nanofuse0 exists

[6/8] Creating VM with networking...
✓ VM created:
  ID:  <uuid>
  IP:  172.16.0.10
  TAP: tap-<vmid>
  MAC: AA:FC:00:xx:xx:xx

[7/8] Verifying TAP device...
✓ TAP device exists
✓ TAP device attached to bridge

[8/8] Starting VM...
✓ VM started successfully!

Testing network connectivity...
  Testing host -> VM connectivity (ping 172.16.0.10)... ✓
  Checking VM network initialization in console... ✓

========================================
Test completed successfully!
========================================
```

## Option 2: Manual Step-by-Step Testing

If you want to manually test each feature:

### 1. Build the Binaries

```bash
# Build both CLI and daemon
go build -o ./bin/nanofuse ./cmd/nanofuse
go build -o ./bin/nanofused ./cmd/nanofused

# Verify builds
./bin/nanofuse version
```

### 2. Start the Daemon

```bash
# Start daemon (requires sudo for networking)
sudo ./bin/nanofused --config-dir /tmp/nanofuse-test

# In another terminal, verify it's running:
curl http://127.0.0.1:8080/health
```

Expected: `{"status":"healthy","timestamp":"..."}`

### 3. Register the Base Image

```bash
# Register the locally built base image
go run register-local-image.go \
  --manifest images/base/build/manifest.json \
  --rootfs images/base/build/rootfs.ext4 \
  --kernel images/base/build/vmlinux \
  --tag nanofuse-base:latest \
  --api-url http://127.0.0.1:8080
```

Expected: `Image registered successfully`

### 4. List Images

```bash
./bin/nanofuse --api-url http://127.0.0.1:8080 image list
```

Expected output:
```
DIGEST                                                  TAGS                    ARCH    SIZE      PULLED
sha256:...                                              nanofuse-base:latest    x86_64  2.0 GB    just now
```

### 5. Test Image Labels (NEW in Phase 1 Polish!)

```bash
./bin/nanofuse --api-url http://127.0.0.1:8080 image labels nanofuse-base:latest
```

Expected: Display of OCI and NanoFuse labels

### 6. Create a VM with Dynamic IP

```bash
./bin/nanofuse --api-url http://127.0.0.1:8080 vm create test-vm \
  --image nanofuse-base:latest \
  --vcpus 1 \
  --memory 512
```

Expected output:
```
VM created successfully
ID:           <uuid>
Name:         test-vm
Image:        nanofuse-base:latest
Architecture: x86_64    # <-- NEW: Dynamic architecture detection!
vCPUs:        1
Memory:       512 MB
State:        created
IP Address:   172.16.0.10    # <-- NEW: Dynamic IP allocation!
Gateway:      172.16.0.1
Netmask:      255.255.255.0
MAC Address:  AA:FC:00:xx:xx:xx
```

### 7. Start the VM

```bash
./bin/nanofuse --api-url http://127.0.0.1:8080 vm start test-vm
```

Expected: `VM started successfully`

### 8. Check VM Status

```bash
./bin/nanofuse --api-url http://127.0.0.1:8080 vm status test-vm
```

Expected:
```
Name:          test-vm
State:         running
IP Address:    172.16.0.10
vCPUs:         1
Memory:        512 MB
Uptime:        <seconds>
```

### 9. Test Logs Tail (NEW in Phase 1 Polish!)

```bash
# Get all logs
./bin/nanofuse --api-url http://127.0.0.1:8080 vm logs test-vm

# Get last 20 lines (NEW feature!)
./bin/nanofuse --api-url http://127.0.0.1:8080 vm logs test-vm --tail 20

# Get last 50 lines
./bin/nanofuse --api-url http://127.0.0.1:8080 vm logs test-vm --tail 50
```

Expected: Console boot logs from the VM

### 10. Test Host-to-VM Connectivity

```bash
# Wait ~30 seconds for VM to fully boot, then:
ping -c 3 172.16.0.10
```

Expected: Successful ping responses (< 1ms latency)

### 11. Create Multiple VMs (Test Dynamic IP Allocation)

```bash
# Create VM 2
./bin/nanofuse --api-url http://127.0.0.1:8080 vm create test-vm-2 \
  --image nanofuse-base:latest --vcpus 1 --memory 512

# Create VM 3
./bin/nanofuse --api-url http://127.0.0.1:8080 vm create test-vm-3 \
  --image nanofuse-base:latest --vcpus 1 --memory 512
```

Expected: Each VM gets a unique IP (172.16.0.11, 172.16.0.12, etc.)

### 12. Test Graceful Shutdown (NEW in Phase 1 Polish!)

```bash
# Stop VM with graceful shutdown (10s timeout, then SIGKILL)
time ./bin/nanofuse --api-url http://127.0.0.1:8080 vm stop test-vm
```

Expected: VM stops within timeout, logs show graceful or forced shutdown

### 13. Test TAP Cleanup (NEW in Phase 1 Polish!)

```bash
# Before deletion, note TAP device
ip link show | grep tap-

# Delete VM
./bin/nanofuse --api-url http://127.0.0.1:8080 vm delete test-vm

# Verify TAP device is gone
ip link show | grep tap-
```

Expected: TAP device is automatically removed

### 14. List All VMs

```bash
./bin/nanofuse --api-url http://127.0.0.1:8080 vm list
```

Expected: Table showing all VMs with IPs, state, architecture

## Phase 1 Features to Test

Here's what you can verify from today's work:

### ✅ Dynamic IP Allocation
- Create multiple VMs
- Each should get unique IP from pool (172.16.0.10-254)
- IPs should be released when VMs are deleted
- New VMs should reuse freed IPs

### ✅ Logs Tail
- `vm logs <name>` - Get all logs
- `vm logs <name> --tail 10` - Get last 10 lines
- `vm logs <name> --tail 100` - Get last 100 lines

### ✅ TAP Device Cleanup
- Create and delete VMs
- TAP devices should be automatically removed
- No leaked `tap-*` devices in `ip link show`

### ✅ Architecture Detection
- Image should show correct architecture (x86_64)
- VMs inherit architecture from image
- Display in `vm create`, `vm status`, `image list`

### ✅ Graceful Shutdown
- VM stops should complete within timeout
- Check logs for "Stopping VM" and "VM stopped gracefully" or "Force-killing VM"
- Fallback to SIGKILL if hung (simulated by killing daemon mid-shutdown)

### ✅ Image Labels
- `image labels <name>` shows OCI standard labels
- Shows NanoFuse-specific labels
- JSON output available with `--json`

## What CAN'T Be Tested Yet

These are Phase 2 features:

❌ **Snapshot/Resume** - Not implemented yet
❌ **S3 Backup/Restore** - Stubbed but not functional
❌ **Log Streaming** - Only tail supported, no follow
❌ **ARM64** - Foundation ready but not tested
❌ **TLS for TCP** - Unix socket only (secure by default)

## Troubleshooting

### "Permission denied" errors
**Solution**: Use `sudo` (networking requires root)

### VM doesn't boot
**Checks**:
1. Verify Firecracker installed: `which firecracker`
2. Verify KVM access: `ls -la /dev/kvm`
3. Check daemon logs: `sudo journalctl -u nanofused -f`
4. Check VM console: `./bin/nanofuse vm logs <name>`

### Network not working
**Checks**:
1. Bridge exists: `ip link show nanofuse0`
2. TAP device exists: `ip link show | grep tap-`
3. IP forwarding enabled: `sysctl net.ipv4.ip_forward`
4. Wait 30-60s for VM to fully boot

### "Image not found"
**Solution**: Register the local image first (step 3 above)

## Cleanup After Testing

```bash
# Stop all VMs
./bin/nanofuse --api-url http://127.0.0.1:8080 vm list
./bin/nanofuse --api-url http://127.0.0.1:8080 vm delete <each-vm>

# Stop daemon
sudo pkill nanofused

# Remove test data
sudo rm -rf /tmp/nanofuse-test
```

## Success Criteria

Phase 1 is working if:

✅ Daemon starts and creates network infrastructure
✅ VMs can be created with unique dynamic IPs
✅ VMs boot successfully in Firecracker
✅ Host can ping VMs (< 1ms latency)
✅ VM logs are accessible with tail support
✅ VMs stop gracefully within timeout
✅ TAP devices are cleaned up on VM deletion
✅ Architecture is detected from images
✅ Image labels are displayed correctly

## Next: Phase 2 Testing

Once Phase 2 (snapshot/resume) is implemented, you'll be able to test:
- VM snapshot creation
- Resume from snapshot
- Sub-2-second cold starts (including first boot)

---

**Ready to test?** Start with Option 1 (comprehensive test) for the fastest validation!
