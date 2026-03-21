# Firecracker Hang - Debug Checklist

Based on debug analysis from `debug.info.txt`

---

## Immediate Checks (Do These First)

```bash
# 1. Check if nanofused daemon is still running
ps aux | grep nanofused

# 2. Check console output (most important!)
cat /tmp/nanofuse/vms/*/console.log | head -100

# 3. Check if Firecracker config was generated
cat /tmp/nanofuse/vms/*/config.json | jq .

# 4. Check TAP device
ip tuntap show
ip addr show nanofuse0

# 5. Manually try the API socket
echo '{}' | socat - UNIX-CONNECT:/tmp/nanofuse/vms/*/firecracker.sock
```

---

## What the Debug Data Shows

✅ **Firecracker process is healthy**
- Running with 5 threads
- 1060+ context switches (not deadlocked)
- 90MB memory (reasonable)
- All FDs open (kernel, rootfs, TAP, KVM)

❌ **Problem is NOT in the process**
- Not CPU-bound
- Not out of memory
- Not I/O blocked
- Not KVM misconfiguration

✅ **Problem IS likely in**
- Config not sent from daemon
- Console output (kernel/systemd)
- Network/TAP setup
- Image build (rootfs/kernel)

---

## Diagnostic Steps

### Step 1: Console Output

```bash
VM_DIR=/tmp/nanofuse/vms/*/
cat "$VM_DIR/console.log"
```

**If empty**:
- Kernel never booted
- Check kernel path exists
- Check config.json has correct kernel_image_path

**If has kernel messages**:
- Check where it stops
- Look for error messages
- Check systemd startup

**If has systemd but then nothing**:
- VM booted but may be stuck
- Check "Testing Inside VMs" commands

### Step 2: Check Config

```bash
VM_DIR=/tmp/nanofuse/vms/*/
cat "$VM_DIR/config.json" | jq .
```

**Required fields**:
- `boot-source.kernel_image_path` - must exist
- `boot-source.boot_args` - should have `console=ttyS0`
- `drives[0].path_on_host` - rootfs must exist
- `machine-config.vcpu_count` - should be 2+
- `machine-config.mem_size_mib` - should be 512+

### Step 3: Check Files Exist

```bash
# Kernel
file /tmp/nanofuse-debug/images/nanofuse-base/latest/vmlinux
# Should say: "Linux kernel"

# Rootfs
file /tmp/nanofuse-debug/images/nanofuse-base/latest/rootfs.ext4
# Should say: "ext4 filesystem data"

# Check sizes
ls -lh /tmp/nanofuse-debug/images/nanofuse-base/latest/
```

### Step 4: Check Network

```bash
# Bridge exists?
ip addr show nanofuse0

# TAP device?
ip tuntap show

# Routes?
ip route | grep nanofuse0

# Can reach VM IP?
ping -c 1 172.16.0.10
```

### Step 5: Daemon Logs

```bash
# Check daemon logs
journalctl -u nanofused -n 50

# Or if running in foreground
ps aux | grep nanofused
```

---

## Common Issues & Fixes

### Issue: Console Log is Empty

**Likely cause**: Kernel not starting

**Fix**:
1. Verify kernel path in config.json
2. Verify kernel file exists and is readable
3. Verify kernel is uncompressed (file should show "Linux kernel")
4. Check boot_args has `console=ttyS0`

### Issue: Kernel boots but systemd hangs

**Likely cause**: Image issue or network timeout

**Fix**:
1. Rebuild image: `cd images/base && sudo make build`
2. Check console for error messages
3. Verify rootfs.ext4 is valid: `sudo fsck.ext4 -n rootfs.ext4`
4. Check network config in image

### Issue: systemd reaches multi-user but no SSH

**Likely cause**: SSH keys or SSH config

**Fix**:
1. Check SSH daemon running: `systemctl status ssh`
2. Check authorized_keys in image
3. Verify your public key is in Dockerfile
4. Rebuild image if you changed SSH config

### Issue: TAP device not working

**Likely cause**: Bridge misconfigured or missing

**Fix**:
1. Check bridge exists: `ip addr show nanofuse0`
2. Check bridge has IP: `ip addr show nanofuse0 | grep inet`
3. Restart network: `sudo systemctl restart systemd-networkd`
4. Check iptables rules: `sudo iptables -L -n | grep nanofuse`

---

## Advanced Debugging

### Use strace to See What Firecracker is Doing

```bash
# Find Firecracker PID
FCPID=$(pgrep firecracker | head -1)

# Start tracing (in background)
sudo strace -f -s 200 -p $FCPID -o /tmp/strace.log &

# Wait a few seconds
sleep 5

# Check what it's waiting on
tail -100 /tmp/strace.log | grep -E "epoll_wait|futex|read|write|socket"
```

**Common patterns**:
- `epoll_wait()` = Waiting for events (normal, not stuck)
- `futex()` = Mutex wait (might indicate contention)
- `read(6,` = Reading from socket (waiting for config)

### Check System Resources

```bash
# CPU usage
top -p $(pgrep firecracker | head -1)

# Memory mapping
pmap -x $(pgrep firecracker | head -1)

# Open files
lsof -p $(pgrep firecracker | head -1) | head -50

# All threads
ps -eLf | grep firecracker
```

### Manual Firecracker Test

```bash
# Create simple config
mkdir -p /tmp/fc-test
cat > /tmp/fc-test/config.json << 'EOF'
{
  "boot-source": {
    "kernel_image_path": "/tmp/nanofuse-debug/images/nanofuse-base/latest/vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k"
  },
  "drives": [{
    "drive_id": "rootfs",
    "path_on_host": "/tmp/nanofuse-debug/images/nanofuse-base/latest/rootfs.ext4",
    "is_root_device": true,
    "is_read_only": false
  }],
  "machine-config": {
    "vcpu_count": 2,
    "mem_size_mib": 512
  }
}
EOF

# Run Firecracker manually
sudo firecracker --api-sock /tmp/test.sock --config-file /tmp/fc-test/config.json

# Watch output (in another terminal)
tail -f /tmp/fc-test/console.log
```

---

## Decision Tree

```
START
  ↓
[1] Is Firecracker process running?
  NO → Start it: sudo ./bin/nanofused
  YES ↓
[2] Is console.log file created?
  NO → Check /tmp/nanofuse/vms/ dir exists
  YES ↓
[3] Is console.log empty?
  YES → Kernel issue (path, format, permissions)
  NO ↓
[4] Does it show kernel boot messages?
  NO → Console not configured (should have ttyS0)
  YES ↓
[5] Does it say "Multi-User System reached"?
  NO → systemd/services failing (check errors in log)
  YES ↓
[6] Can you SSH in?
  NO → SSH key issue (add key to Dockerfile)
  YES ↓
[7] VM is working! ✅
```

---

## Next Steps

1. **Run console check**:
   ```bash
   cat /tmp/nanofuse/vms/*/console.log
   ```

2. **Share the output** - that's the key diagnostic

3. **If console is empty**:
   - Kernel path issue
   - Try manual firecracker test

4. **If console has errors**:
   - Systemd/service issue
   - Rebuild image

5. **If VM boots but no SSH**:
   - SSH key not in image
   - Add key to Dockerfile
   - Rebuild

---

See: `docs/FIRECRACKER_DEBUG_ANALYSIS.md` for detailed analysis
