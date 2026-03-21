# NanoFuse Steel Thread Plan

## Status: Phase 1 Incomplete

The nanofuse codebase has **solid architecture** (~7000+ lines of Go) but is blocked by a fundamental issue: **VMs boot but systemd services don't start**.

---

## What's Complete

| Component | Status | Notes |
|-----------|--------|-------|
| API Server | ✅ Working | HTTP REST, Unix socket & TCP modes |
| CLI | ✅ Working | Full cobra-based CLI, vm/image/snapshot commands |
| Firecracker Integration | ✅ Working | VMs boot, kernel loads, network comes up |
| Network Stack | ✅ Working | TAP devices, bridge, NAT, port forwarding, IPAM |
| Database | ✅ Working | SQLite for VMs, images, snapshots |
| Image Pull (API) | ✅ Working | GHCR integration, layer caching |
| Base Image Build | ✅ Working | Ubuntu 24.04, Dockerfile, build scripts |
| Design Documentation | ✅ Complete | 2600+ line design spec |

## What's Broken

| Issue | Severity | Impact |
|-------|----------|--------|
| **Systemd not starting services** | 🔴 CRITICAL | VMs boot but can't run workloads |
| Console log visibility | 🟠 HIGH | Can't debug what's happening in guest |
| CLI image pull | 🟠 HIGH | "Job ID is required" error during polling |
| Unix socket vs TCP conflict | 🟡 MEDIUM | Workaround: use `--api-url` flag |
| GetImageByTag bug | 🟡 MEDIUM | Returns wrong image; workaround: use digest |

---

## The Blocking Issue

From `PRIORITY_TODO.md`:

- VM gets IP (172.16.0.x), responds to ping
- All TCP ports show "closed"
- Systemd services (nginx, backend) never start
- No visibility into boot process

**This is THE blocker.** Until systemd works in the guest, you can't run any workloads.

---

## Steel Thread Definition

A minimal vertical slice proving end-to-end functionality:

```
1. Start daemon       ─→ nanofused running
2. Pull/load image    ─→ Image in database
3. Create VM          ─→ VM record created
4. Start VM           ─→ Firecracker running, network up
5. Systemd starts     ─→ ❌ BLOCKED HERE ❌
6. Services run       ─→ nginx/app listening
7. HTTP responds      ─→ curl http://vm-ip/ works
8. Stop/delete VM     ─→ Clean shutdown
```

---

## Priority-Ordered Tasks

### P0: Get Console Visibility

**Why:** Can't debug anything without seeing boot output.

**Tasks:**
- [ ] Verify `nanofuse vm logs <id>` returns console output
- [ ] Check `internal/firecracker/manager.go` → `GetConsoleLogs()`
- [ ] Verify Firecracker config has serial console enabled
- [ ] Test: see kernel messages and systemd output

**Files:**
- `internal/firecracker/manager.go`
- `internal/api/vm_handlers.go:717-774`

**Success:** Can see "Starting XXXX.service" messages in logs.

---

### P1: Debug & Fix Systemd Init

**Why:** This is THE blocker for everything.

**Tasks:**
- [ ] Capture console output on boot
- [ ] Identify where boot stops (kernel panic? systemd not found? target not reached?)
- [ ] Mount rootfs.ext4 on host and inspect:
  ```bash
  sudo mount -o loop /var/lib/nanofuse/images/<digest>/rootfs.ext4 /mnt/rootfs
  ls -la /mnt/rootfs/lib/systemd/systemd
  ls -la /mnt/rootfs/sbin/init
  cat /mnt/rootfs/etc/systemd/system/default.target
  ```
- [ ] Apply fix (likely one of):
  - Create `/sbin/init → /lib/systemd/systemd` symlink
  - Set default.target to multi-user.target
  - Add kernel args: `systemd.unit=multi-user.target systemd.log_level=debug`
  - Mount /proc, /sys, /dev in init script

**Files:**
- `images/base/Dockerfile`
- `internal/api/vm_handlers.go:99,219-222` (kernel args)

**Success:** Console shows "Reached target Multi-User System."

---

### P2: Verify Service Startup

**Why:** Prove services actually run.

**Tasks:**
- [ ] Boot VM with nginx enabled
- [ ] Check console for nginx startup
- [ ] `curl http://<vm-ip>/` from host
- [ ] Verify response received

**Success:** HTTP response from nginx inside VM.

---

### P3: Fix CLI Image Pull

**Why:** Need CLI to work for full flow.

**Tasks:**
- [ ] Debug job polling in `cmd/nanofuse/main.go:207-250`
- [ ] Find why "Job ID is required" occurs
- [ ] Fix the polling loop
- [ ] Test: `nanofuse image pull --default` works

**Files:**
- `cmd/nanofuse/main.go:178-252`
- `internal/client/client.go`

**Success:** `nanofuse image pull` completes without error.

---

### P4: End-to-End Test Script

**Why:** Prove the steel thread works, enable regression testing.

**Tasks:**
- [ ] Create `scripts/test-steel-thread.sh`
- [ ] Test full flow: pull → create → start → curl → stop → delete
- [ ] Add to CI/testing

**Success:** Script passes on clean system.

---

### P5: Add Container Runtime to Guest

**Why:** Your actual goal - run Docker containers.

**Tasks:**
- [ ] Add containerd or Docker to base image
- [ ] Enable containerd.service
- [ ] Test container run inside VM
- [ ] Add exec/attach API endpoints (if needed)

**Files:**
- `images/base/Dockerfile`
- New: guest agent for container operations (see design doc section 5.3)

**Success:** Can run `docker run nginx` inside VM.

---

## What to Defer

These are in the design doc but NOT needed for steel thread:

| Feature | Why Defer |
|---------|-----------|
| Network policies (nftables) | MVP doesn't need per-VM isolation |
| Jailer/seccomp | Security hardening, not core function |
| mTLS/RBAC | Auth can be added later |
| Snapshots | Nice-to-have, not steel thread |
| API Gateway (L7) | Design doc Phase 2+ |
| Tetragon/eBPF | Observability enhancement |
| gRPC API | HTTP works fine |
| AI introspection | Far future |

---

## Daily Iteration Pattern

```
Day 1: Console logs working?
       ├─ Yes → Day 2
       └─ No  → Fix console capture

Day 2: Can see boot output?
       ├─ Yes → Identify systemd problem
       └─ No  → Debug console path/config

Day 3: Understand systemd failure?
       ├─ Yes → Apply fix, rebuild image
       └─ No  → Mount rootfs, inspect manually

Day 4: Services start?
       ├─ Yes → Test HTTP response
       └─ No  → Debug specific service failure

Day 5: HTTP works?
       ├─ Yes → Fix CLI pull, create test script
       └─ No  → Check networking/firewall

Day 6+: Container runtime in guest
```

---

## Quick Debug Commands

```bash
# Start daemon
sudo nanofused --config /etc/nanofuse/nanofused.yaml --tcp :8080

# Create VM (using base image digest)
nanofuse --api-url http://localhost:8080 vm create \
  sha256:DIGEST my-vm --vcpus 2 --memory 512

# Start VM
nanofuse --api-url http://localhost:8080 vm start my-vm

# Get VM IP
nanofuse --api-url http://localhost:8080 vm status my-vm | grep ip

# Check console logs (THE KEY DEBUG TOOL)
nanofuse --api-url http://localhost:8080 vm logs my-vm

# Test connectivity
ping <vm-ip>
curl http://<vm-ip>/

# Mount rootfs to inspect
sudo mount -o loop /var/lib/nanofuse/images/<digest>/rootfs.ext4 /mnt/rootfs
ls -la /mnt/rootfs/lib/systemd/
cat /mnt/rootfs/etc/systemd/system/default.target
sudo umount /mnt/rootfs

# Clean up
nanofuse --api-url http://localhost:8080 vm delete my-vm
```

---

## Success Criteria

Steel thread is DONE when:

- [ ] `nanofuse image pull --default` works
- [ ] `nanofuse vm run default test-vm` boots with services
- [ ] `curl http://<vm-ip>/` returns response
- [ ] Full lifecycle (create→start→stop→delete) works
- [ ] Documented in working test script

---

## Files to Focus On

| File | What It Does | Why Important |
|------|--------------|---------------|
| `internal/firecracker/manager.go` | Starts VMs | Console log capture |
| `internal/api/vm_handlers.go` | API endpoints | Kernel args, lifecycle |
| `images/base/Dockerfile` | Base image | Systemd setup |
| `images/base/build.sh` | Build script | Rootfs creation |
| `cmd/nanofuse/main.go` | CLI | Image pull bug |

---

## Reference

- `CRITICAL_ISSUES.md` - Known bugs from Nov 2024
- `PRIORITY_TODO.md` - Detailed systemd debugging guide
- `docs/firecracker-runner-design.md` - Full architecture (aspirational)
- `images/base/README.md` - Base image documentation
