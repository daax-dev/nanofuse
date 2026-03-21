# PRIORITY #1: Fix Systemd Services Not Starting in Todo-App VM

**Status**: BLOCKING - VM boots but HTTP services don't respond
**Created**: 2025-11-14
**Session Context**: Todo-app image built and registered, VM created and running, but services not accessible

---

## Problem Statement

The todo-app microVM boots successfully and is network-reachable (ping works), but the systemd services (nginx on port 80, todo-backend on port 8080) are not starting. All ports show as "closed" when scanned with nmap.

**Current State**:
- ✅ VM created with digest: `sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628`
- ✅ VM running with IP: `172.16.0.11`
- ✅ VM responds to ping
- ❌ Port 80 (nginx) closed
- ❌ Port 8080 (todo-backend) closed
- ❌ HTTP endpoints not responding

**Expected Behavior**:
- `curl http://172.16.0.11/health` should return JSON health status
- `curl http://172.16.0.11/api/todos` should return todos list
- Both nginx and todo-backend services should be running

---

## Environment & File Locations

### Image Files
- **Rootfs**: `/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4` (2GB ext4 filesystem)
- **Kernel**: `/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/vmlinux` (21M Firecracker kernel)
- **Dockerfile**: `/home/jpoley/ps/nanofuse/examples/todo-app/docker/Dockerfile`

### Database & Config
- **Database**: `/var/lib/nanofuse/nanofuse.db` (SQLite)
- **Daemon Config**: `/etc/nanofuse/nanofused.yaml`
- **Daemon Status**: `systemctl status nanofused` (running on host)

### Current VM Details
- **VM ID**: `a49cbbe2-fd75-4c30-b2e6-68d1bf956375`
- **VM Name**: `my-todo-app`
- **IP Address**: `172.16.0.11`
- **VCPUs**: 2
- **Memory**: 1024 MiB
- **Kernel Args**: `console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd`

### Useful Commands
```bash
# VM management
nanofuse --api-url http://localhost:8080 vm list
nanofuse --api-url http://localhost:8080 vm start my-todo-app
nanofuse --api-url http://localhost:8080 vm stop my-todo-app

# VM creation (using digest to avoid tag lookup bug)
nanofuse --api-url http://localhost:8080 vm create \
  sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628 \
  my-todo-app --vcpus 2 --memory 1024 \
  --kernel-args "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd"

# Testing
ping 172.16.0.11
nmap -p 22,80,8080 172.16.0.11
curl http://172.16.0.11/health
curl http://172.16.0.11/api/todos
```

---

## What We Know

### Image Build Process
1. **Docker Image Built** from `/home/jpoley/ps/nanofuse/examples/todo-app/docker/Dockerfile`
   - Multi-stage build: Go backend + Node frontend + Ubuntu 24.04 base
   - Systemd installed: `apt-get install systemd systemd-sysv nginx`
   - Services created and enabled in Dockerfile (lines 61-90):
     ```dockerfile
     RUN systemctl enable todo-backend.service nginx.service
     ```

2. **Services Configuration**:
   - **todo-backend.service**: Systemd unit at `/etc/systemd/system/todo-backend.service`
     - Runs: `/usr/local/bin/todo-server -db-path /data/todos.db -http-port 8080 -grpc-port 9090`
     - Type: `simple`
     - WantedBy: `multi-user.target`

   - **nginx**: Configured in `/etc/nginx/sites-available/default`
     - Listens on port 80
     - Proxies to backend on localhost:8080

3. **Dockerfile CMD**: `CMD ["/lib/systemd/systemd"]` (line 121)
   - This is NOT used by Firecracker (which boots kernel directly)
   - Must pass `init=/lib/systemd/systemd` in kernel args instead

4. **Rootfs Creation**:
   - Docker image → `docker export` → tarball
   - Extracted to `/tmp/todo-app-rootfs/`
   - Copied into 2GB ext4 filesystem using `mount -o loop` and `rsync`
   - Registered in NanoFuse database

### What We've Verified

✅ **VM Network Stack Works**:
- VM gets IP from IPAM pool (172.16.0.11)
- Responds to ping (0.2-0.4ms RTT)
- Bridge network configured correctly (nanofuse0)
- NAT rules in place

✅ **VM Boots**:
- Firecracker starts successfully
- Kernel loads
- Rootfs mounts (root=/dev/vda1)
- VM enters "running" state
- No errors in daemon logs

✅ **Image Registered Correctly**:
- Database has proper sha256 digest
- Rootfs and kernel paths are accessible
- Files exist and are readable by daemon

### What We've Tried

1. **Kernel Args Iteration**:
   - ❌ First attempt: No `init=` parameter (services didn't start)
   - ❌ Second attempt: Added `init=/lib/systemd/systemd` (services still didn't start)

2. **Network Testing**:
   - ✅ Ping works (ICMP)
   - ❌ TCP connections fail (ports closed)
   - ❌ Both port 80 and 8080 unresponsive

3. **Wait Time**:
   - Tried waiting 10s, 15s, 20s after boot
   - No change - services never start

---

## Root Cause Hypotheses

### Hypothesis 1: Systemd Not Running as PID 1 ⭐ MOST LIKELY
**Evidence**:
- Kernel arg `init=/lib/systemd/systemd` was added
- But we haven't verified systemd is actually running
- If systemd fails to start, nothing else will start

**Why this might happen**:
- `/lib/systemd/systemd` binary might not exist at that path in rootfs
- Systemd might be expecting different path (e.g., `/sbin/init` symlink)
- Missing systemd dependencies in rootfs
- Permissions issues (systemd needs specific perms)

**How to verify**:
- Need console/serial access to see boot logs
- Check if systemd is even attempting to start
- Look for kernel panic or init failures

### Hypothesis 2: Services Enabled But Target Not Reached
**Evidence**:
- Dockerfile runs `systemctl enable todo-backend.service nginx.service`
- This creates symlinks in `/etc/systemd/system/multi-user.target.wants/`
- But `multi-user.target` might not be the default target

**Why this might happen**:
- Default target might be `rescue.target` or `basic.target`
- Systemd might be waiting for something that never completes
- Network target dependency might be blocking

**How to verify**:
- Check `/etc/systemd/system/default.target` symlink in rootfs
- May need to add `systemd.unit=multi-user.target` to kernel args
- Check service dependencies

### Hypothesis 3: Rootfs Permissions/Ownership Issues
**Evidence**:
- Rootfs created with rsync from Docker export
- Files might have wrong ownership (all owned by user who ran docker export)

**Why this might happen**:
- `docker export` preserves UIDs/GIDs from container
- `rsync -a` preserves permissions
- Systemd services might fail if files owned incorrectly
- `/usr/local/bin/todo-server` might not be executable for systemd

**How to verify**:
- Mount rootfs.ext4 and check permissions
- Verify `/usr/local/bin/todo-server` is executable
- Check systemd unit file ownership

### Hypothesis 4: Missing Systemd Dependencies
**Evidence**:
- Ubuntu 24.04 base image
- Systemd installed via apt-get
- But might be missing dbus, cgroups, or other dependencies

**Why this might happen**:
- Minimal container image doesn't include everything systemd needs
- Missing `/sys`, `/proc`, `/run` mounts
- Missing cgroup filesystem
- Missing dbus daemon

**How to verify**:
- Check rootfs for `/lib/systemd/systemd` existence
- Verify systemd dependencies: `ldd /lib/systemd/systemd`
- Check for dbus socket

### Hypothesis 5: Console/Serial Output Not Configured
**Evidence**:
- Kernel arg has `console=ttyS0`
- But we have no way to see what's being printed to console

**Why this matters**:
- Boot errors might be happening but invisible
- Systemd might be logging to console/journal
- Can't see what's actually happening during boot

**How to verify**:
- Need to configure Firecracker to capture serial console output
- Check NanoFuse code for console logging capabilities

---

## Debugging Strategy (Ordered by Priority)

### STEP 1: Get Console Access 🔴 CRITICAL
**Why first**: We're flying blind without seeing what happens during boot

**Actions**:
1. Check if NanoFuse has console/serial logging capability
   - Search codebase: `grep -r "serial\|console\|stdout" internal/`
   - Look for Firecracker config: `socket_path` for metrics/logs

2. If console logging exists:
   - Enable it for the VM
   - Recreate VM and check logs
   - Look for systemd startup messages, errors, kernel panic

3. If console logging doesn't exist:
   - Check Firecracker's socket API for log retrieval
   - May need to add console logging feature
   - Alternative: Use kernel args to increase verbosity: `systemd.log_level=debug`

**Expected Outcomes**:
- See boot sequence
- Identify where boot process stops or what fails
- Get systemd startup messages

**Files to check**:
- `/home/jpoley/ps/nanofuse/internal/vmm/firecracker.go` (Firecracker config)
- `/home/jpoley/ps/nanofuse/internal/vmm/vm.go` (VM management)

### STEP 2: Verify Systemd Binary in Rootfs
**Why**: Need to confirm `/lib/systemd/systemd` exists and is executable

**Actions**:
```bash
# Mount the rootfs and inspect
sudo mkdir -p /mnt/todo-rootfs
sudo mount -o loop /var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4 /mnt/todo-rootfs

# Check systemd binary
ls -la /mnt/todo-rootfs/lib/systemd/systemd
ls -la /mnt/todo-rootfs/sbin/init
file /mnt/todo-rootfs/lib/systemd/systemd
ldd /mnt/todo-rootfs/lib/systemd/systemd

# Check service files
ls -la /mnt/todo-rootfs/etc/systemd/system/todo-backend.service
ls -la /mnt/todo-rootfs/etc/systemd/system/nginx.service
ls -la /mnt/todo-rootfs/etc/systemd/system/multi-user.target.wants/

# Check default target
ls -la /mnt/todo-rootfs/etc/systemd/system/default.target

# Check backend binary
ls -la /mnt/todo-rootfs/usr/local/bin/todo-server
file /mnt/todo-rootfs/usr/local/bin/todo-server

# Unmount when done
sudo umount /mnt/todo-rootfs
```

**Expected Outcomes**:
- `/lib/systemd/systemd` exists and is ELF64 executable
- All dependencies (libraries) are present
- Service files exist with correct symlinks
- Backend binary exists and is executable

**If Missing**:
- May need to use different init path
- May need to create `/sbin/init` symlink
- May need to rebuild image with proper systemd installation

### STEP 3: Try Alternative Init Methods
**Why**: If systemd is too complex, try simpler approaches first

**Option A: Direct Service Startup (No Systemd)**
```bash
# Create simple init script
cat > /tmp/simple-init.sh <<'EOF'
#!/bin/bash
# Mount necessary filesystems
mount -t proc proc /proc
mount -t sysfs sys /sys
mount -t devtmpfs dev /dev

# Configure network (IP already set by kernel)
ip link set eth0 up

# Start services directly
/usr/sbin/nginx &
/usr/local/bin/todo-server -db-path /data/todos.db -http-port 8080 -grpc-port 9090 &

# Keep init running
while true; do sleep 3600; done
EOF
chmod +x /tmp/simple-init.sh

# Copy to rootfs
sudo mount -o loop /var/lib/nanofuse/images/.../rootfs.ext4 /mnt/todo-rootfs
sudo cp /tmp/simple-init.sh /mnt/todo-rootfs/init
sudo umount /mnt/todo-rootfs

# Create VM with new init
nanofuse vm create sha256:0c8543... my-todo-app \
  --kernel-args "console=ttyS0 root=/dev/vda1 rw init=/init"
```

**Option B: Update Kernel Args for Systemd**
```bash
# Try additional systemd parameters
nanofuse vm create sha256:0c8543... my-todo-app --kernel-args \
  "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd systemd.unit=multi-user.target systemd.log_level=debug systemd.log_target=console"
```

### STEP 4: Check Firecracker Configuration
**Why**: Firecracker might need specific config for systemd to work

**Actions**:
1. Review Firecracker config generation:
   - File: `/home/jpoley/ps/nanofuse/internal/vmm/firecracker.go`
   - Check boot-source configuration
   - Verify boot_args are passed correctly

2. Check for missing Firecracker features:
   - vsock configuration
   - Additional devices needed
   - Memory/CPU constraints

3. Compare with working Firecracker+systemd examples:
   - Firecracker quickstart guide uses systemd
   - Check what kernel args they use
   - Compare VM config structure

### STEP 5: Rebuild Image with Fixes
**Why**: If rootfs issues found, need to rebuild properly

**Possible Fixes**:
1. **Add init script wrapper**:
   - Create `/sbin/init` that ensures mounts then execs systemd

2. **Fix permissions in Dockerfile**:
   ```dockerfile
   RUN chmod 755 /usr/local/bin/todo-server && \
       chown root:root /usr/local/bin/todo-server
   ```

3. **Set default target**:
   ```dockerfile
   RUN systemctl set-default multi-user.target
   ```

4. **Add debugging**:
   ```dockerfile
   RUN echo "SYSTEMD_LOG_LEVEL=debug" >> /etc/environment
   ```

### STEP 6: Verify Service Dependencies
**Actions**:
```bash
# Mount rootfs and check service files
sudo mount -o loop /var/lib/nanofuse/images/.../rootfs.ext4 /mnt/todo-rootfs

# Check backend service
cat /mnt/todo-rootfs/etc/systemd/system/todo-backend.service

# Check nginx service
systemctl cat nginx  # from within chroot

# Verify backend binary works
sudo chroot /mnt/todo-rootfs /usr/local/bin/todo-server --help

# Verify nginx config
sudo chroot /mnt/todo-rootfs nginx -t

sudo umount /mnt/todo-rootfs
```

---

## Quick Wins to Try First

### Quick Win 1: Simple Init Script (15 min)
Create a minimal init script that starts services directly without systemd. This will prove:
- Network works
- Services can run
- Isolates problem to systemd specifically

### Quick Win 2: Console Logging (10 min)
Add console output capture to see boot messages. Even if we can't fix systemd, we'll see why it's failing.

### Quick Win 3: Kernel Args (5 min)
Try different kernel arg combinations that are known to work with systemd in Firecracker:
- Add `systemd.unit=multi-user.target`
- Add `systemd.log_level=debug`
- Try `init=/sbin/init` instead

---

## Success Criteria

When this is SOLVED, you should be able to:

```bash
# Create and start VM
nanofuse --api-url http://localhost:8080 vm create \
  ghcr.io/peregrinesummit/nanofuse/todo-app:latest \
  my-todo-app --vcpus 2 --memory 1024

nanofuse --api-url http://localhost:8080 vm start my-todo-app

# Wait for boot
sleep 15

# Get IP
VM_IP=$(nanofuse --api-url http://localhost:8080 vm list --json | \
  jq -r '.vms[] | select(.name=="my-todo-app") | .config.network.ip_address')

# Test endpoints - BOTH should work
curl http://$VM_IP/health
# Expected: {"status":"ok","timestamp":"2025-11-14T23:30:00Z"}

curl http://$VM_IP/api/todos
# Expected: {"todos":[]}

# Verify services running
nmap -p 80,8080 $VM_IP
# Expected: Both ports open
```

---

## Related Issues

1. **GetImageByTag Bug**: Tag lookup returns wrong image (returns base instead of todo-app)
   - Workaround: Use digest directly
   - Should be fixed in `internal/storage/db.go:434-475`

2. **Database Path Issues**: Had to move from `/tmp/` to `/var/lib/nanofuse/` due to PrivateTmp
   - Fixed but worth documenting

3. **Image File Paths**: Need to be accessible to daemon (not in private /tmp)
   - Fixed for todo-app
   - Base image still has issues but not critical

---

## Additional Context

### Why This Matters
This is the ENTIRE POINT of NanoFuse - to run real applications in microVMs. The todo-app is the first real application being tested. Without services working, we only have infrastructure that can boot VMs but can't run actual workloads.

### What Works So Far
- ✅ Image pull from GHCR
- ✅ Image registration (local)
- ✅ VM creation
- ✅ VM lifecycle (create/start/stop/delete)
- ✅ Network configuration (NAT mode)
- ✅ IP allocation (IPAM)
- ✅ Firecracker integration
- ✅ Database storage
- ✅ API daemon

### Blocking Impact
- Can't demo real application workloads
- Can't test application-level features
- Can't prove the value proposition
- Can't move to Phase 2 (snapshots/resume)

---

## Reference Links

- **Firecracker Documentation**: https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md
- **Systemd in Containers**: https://systemd.io/CONTAINER_INTERFACE/
- **Firecracker Network Setup**: https://github.com/firecracker-microvm/firecracker/blob/main/docs/network-setup.md

---

## Session Handoff Checklist

When picking this up in next session:

- [ ] Read this entire document
- [ ] Verify VM is still running: `nanofuse vm list`
- [ ] Verify current IP: `nanofuse vm list | grep my-todo-app`
- [ ] Try ping test: `ping 172.16.0.11` (or current IP)
- [ ] Start with STEP 1 (console access) above
- [ ] Update this document with findings
- [ ] Document solution when found

**Good luck! The VM boots, network works, we just need services to start. You're 90% there.**
