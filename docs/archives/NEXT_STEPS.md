# What to Do Next

## Current Status

✅ **build.sh has been fixed** - reverted to working S3 URL
❌ **Build failed before** - need to retry
❓ **Hang cause unknown** - need to investigate boot sequence

## Quick Path Forward

### Step 1: Clean Rebuild (Run with sudo)

```bash
cd images/base
sudo make clean
sudo make build
```

**Expected output**:
```
✓ Docker image built: nanofuse-base:latest
✓ Rootfs extracted
✓ Kernel downloaded
✓ Manifest generated
```

**Time**: ~3-4 minutes

### Step 2: Verify Kernel Downloaded

```bash
ls -lh ./build/vmlinux
# Should show a file ~21-30MB

strings ./build/vmlinux | grep "Linux version"
# Should show the actual kernel version
```

### Step 3: Start Test VM

```bash
./bin/nanofuse vm create default test-vm
./bin/nanofuse vm start test-vm
```

### Step 4: Monitor Boot Output

```bash
# Watch console in real-time
tail -f /tmp/nanofuse/vms/*/console.log
```

**What to look for**:
- Kernel boot messages (should see within 2-3 seconds)
- systemd startup (within 5-10 seconds)
- Service startup messages
- Error messages (FAILED units)

### Step 5: Let It Boot (Wait 30-60 seconds)

Don't interrupt - old kernels can be slow. Watch for:
- Kernel loading ✓
- systemd starting ✓
- Services coming online ✓
- "Reached target Multi-User System" ✓
- Error messages ✗

### Step 6: Share Console Output

**If it works**: Great! Move to SSH testing
**If it hangs**: Share the last 50 lines of console output

```bash
tail -50 /tmp/nanofuse/vms/*/console.log
```

This shows exactly where it's getting stuck.

---

## Debugging Matrix

| Symptom | Likely Cause | Check |
|---------|--------------|-------|
| Empty console.log | Kernel not loading | `file ./build/vmlinux` |
| Kernel messages then nothing | Systemd hung | Console output after boot messages |
| FAILED systemd units | Service issue | Look for `FAILED` in console |
| Network timeouts | DHCP issue | Check bridge: `ip addr show nanofuse0` |
| VM seems to boot but can't SSH | SSH key missing | Add key to Dockerfile |

---

## Files to Check

```bash
# Build artifacts
./images/base/build/vmlinux          # Kernel binary
./images/base/build/rootfs.ext4      # Filesystem
./images/base/build/manifest.json    # Metadata

# VM console
/tmp/nanofuse/vms/*/console.log      # Boot output

# Configuration
/tmp/nanofuse/vms/*/config.json      # Firecracker config
```

---

## Expected Timeline

- **0-2 sec**: Kernel loads, boot messages appear
- **2-5 sec**: systemd starts initializing
- **5-10 sec**: Services coming online
- **10-30 sec**: All services start, multi-user target reached
- **30+ sec**: Idle (ready for SSH or commands)

If nothing appears after 5 seconds, something is wrong.

---

## Common Issues & Quick Fixes

**"build/vmlinux: Permission denied"**
```bash
# Needs sudo for all make commands
cd images/base
sudo make clean
sudo make build
```

**"curl: 404 error"**
✅ Already fixed - build.sh reverted to working URL

**Empty console.log**
```bash
# Kernel didn't load - check these:
file ./build/vmlinux
ls -la /tmp/nanofuse/vms/*/firecracker.sock
```

**Kernel boots but hangs**
```bash
# Send the last 50 lines of console for analysis
tail -50 /tmp/nanofuse/vms/*/console.log
```

**Can't find VM files**
```bash
# Make sure VM is running
./bin/nanofuse vm status test-vm
# Should show: running

# List all VMs
./bin/nanofuse vm list
```

---

## What Happens After Boot Works

Once the VM boots successfully:

1. **Add SSH key to Dockerfile**
   - Edit `images/base/Dockerfile`
   - Add your SSH public key to authorized_keys
   - Rebuild image

2. **Test SSH access**
   ```bash
   VM_IP=$(./bin/nanofuse vm inspect test-vm | grep IP)
   ssh root@$VM_IP
   ```

3. **Run integration tests**
   ```bash
   ./bin/nanofuse vm exec test-vm whoami
   ./bin/nanofuse vm exec test-vm hostname
   ```

4. **Build custom images** (for your apps)
   - Create new Dockerfile extending base image
   - Install your dependencies
   - Build and test

---

## TL;DR

```bash
# 1. Build
cd images/base && sudo make clean && sudo make build

# 2. Test (watch output for 30 seconds)
./bin/nanofuse vm create default test-vm
./bin/nanofuse vm start test-vm
tail -f /tmp/nanofuse/vms/*/console.log

# 3. If stuck, share last 50 lines
tail -50 /tmp/nanofuse/vms/*/console.log

# 4. If works, add SSH key and test
# (see SSH_ACCESS_QUICK_START.md)
```

That's it!
