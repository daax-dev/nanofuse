# Phase 1C/1D: Comprehensive Services Fix Plan

**Date**: 2025-11-19
**Status**: Investigation & Planning Phase

---

## Current Situation

### What Works ✅
- VM boots successfully
- Network configured correctly (172.16.0.10)
- VM is pingable from host (0.4ms latency)
- Firecracker/TAP/bridge networking functional

### What's Broken ❌
- Services (nginx, todo-backend) NOT running
- Ports 80 and 8080 CLOSED
- Console shows: `[FAILED] Failed to start nginx.service`

### Known Facts
- **Image**: `ghcr.io/peregrinesummit/nanofuse/todo-app:latest`
- **Image Digest**: `sha256:0c8543d7...`
- **Architecture**: x86_64
- **Kernel**: 5.10.240 (bundled with image)
- **Current Kernel Args**: `console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k ip=172.16.0.10::172.16.0.1:255.255.255.0::eth0:off net.ifnames=0 biosdevname=0`
- **Missing**: `init=/lib/systemd/systemd`

---

## What the Image SHOULD Contain

From `examples/todo-app/docker/Dockerfile` analysis:

### Services Expected
1. **nginx.service**
   - Port 80
   - Serves static frontend from `/var/www/html`
   - Reverse proxies `/api` → `localhost:8080`
   - Enabled via `systemctl enable nginx.service` (Dockerfile line 81)

2. **todo-backend.service**
   - Port 8080
   - Go backend binary at `/usr/local/bin/todo-server`
   - DuckDB database at `/data/todos.db`
   - Enabled via `systemctl enable todo-backend.service` (Dockerfile line 81)

### Init Configuration
- **Dockerfile CMD**: `["/lib/systemd/systemd"]` (line 121)
- **CRITICAL**: Docker CMD is ignored when running as Firecracker VM
- **VM requires**: Explicit `init=/lib/systemd/systemd` in kernel args

---

## Root Cause Analysis Framework

### Hypothesis 1: Systemd NOT Starting as PID 1 ⚠️ MOST LIKELY
**Evidence**:
- Kernel args MISSING `init=/lib/systemd/systemd`
- Docker CMD is not used by Firecracker
- Without explicit init, kernel uses default (may be broken/wrong)

**If True**:
- Wrong init runs as PID 1
- Systemd never starts
- Services never start
- BUT: Console shows "[FAILED] Failed to start nginx.service" (systemd message!)

**Contradiction**:
- If systemd not running, how do we see systemd error messages?
- Possible: systemd IS starting somehow (symlink?) but dying early?

**Verification Needed**:
- Read full console log
- Check what process is PID 1
- Check if systemd output appears in console

### Hypothesis 2: Systemd Starting BUT Services Failing
**Evidence**:
- Console shows systemd error: `[FAILED] Failed to start nginx.service`
- This is a systemd-formatted message

**If True**:
- Systemd IS running as PID 1
- Services are being attempted
- Something else is wrong:
  - Service files broken?
  - Binaries missing?
  - Dependencies not met?
  - Permissions wrong?

**Verification Needed**:
- Read full console log for systemd startup messages
- Check journalctl output in console
- See actual error messages for service failures

### Hypothesis 3: Services Starting But Timing Out
**Evidence**:
- Services might start but fail health checks
- Race condition at boot time

**If True**:
- Network not ready when services start
- Services exit before becoming functional

**Verification Needed**:
- Check service dependencies (After=network.target)
- Look for timeout messages

---

## Investigation Plan (Evidence Gathering)

### Step 1: Read Console Log 📋
**Purpose**: Understand what actually happens during boot

**User Action Required**:
```bash
sudo tail -200 /var/lib/nanofuse/vms/4803490f-2d7e-4825-8b02-607636b9962d/console.log
```

**What to Look For**:
1. **Kernel boot messages**:
   - `[    0.000000] Linux version 5.10.240`
   - `[    0.000000] Command line: console=ttyS0 root=/dev/vda1...`

2. **Init detection**:
   - `Run /sbin/init as init process`
   - `Run /lib/systemd/systemd as init process`
   - `Failed to execute /sbin/init`

3. **Systemd startup**:
   - `systemd[1]: System time before build time`
   - `systemd[1]: Detected virtualization kvm`
   - `systemd[1]: Reached target Basic System`

4. **Service attempts**:
   - `Starting nginx...`
   - `Starting Todo App Backend Service...`
   - `[FAILED] Failed to start nginx.service`
   - `[FAILED] Failed to start todo-backend.service`

5. **Actual error messages**:
   - `nginx.service: Failed with result 'exit-code'`
   - Specific error text explaining WHY

**Decision Tree Based on Console Log**:
```
IF: "Run /lib/systemd/systemd as init process"
  → Systemd IS starting as PID 1
  → Hypothesis 2 (services failing)
  → Go to Step 2B

ELSE IF: "Run /sbin/init as init process"
  → Check if /sbin/init → systemd symlink works
  → If symlink broken: Hypothesis 1
  → Go to Step 2A

ELSE IF: No init messages or kernel panic
  → Critical boot failure
  → Image/kernel mismatch issue
  → Go to Step 3
```

### Step 2A: Fix Init Parameter (If systemd not starting)
**Action**: Add `init=/lib/systemd/systemd` to kernel args

**Files to Modify**:
1. `internal/api/vm_handlers.go` line 99:
```go
KernelArgs: "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd",
```

2. `internal/api/vm_handlers.go` line 217:
```go
config.KernelArgs = fmt.Sprintf(
    "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd ip=%s::%s:255.255.255.0::eth0:off net.ifnames=0 biosdevname=0",
    ip, network.BridgeGateway,
)
```

**Deployment**:
- Rebuild daemon
- Stop daemon
- Deploy new binary
- Create fresh VM
- Test

### Step 2B: Debug Service Failures (If systemd IS starting)
**Investigation Questions**:
1. Are service files present and correct?
2. Are binaries present and executable?
3. What are the ACTUAL error messages?
4. Are dependencies met?

**Possible Fixes**:
- Service files might need `After=` dependencies adjusted
- Binaries might be missing execute permission
- Configuration files might be missing
- DuckDB might fail to initialize `/data` directory

### Step 3: Handle Critical Boot Failure (Worst case)
**If**: Kernel panic or no init messages at all

**Actions**:
1. Verify image integrity
2. Check kernel compatibility
3. Verify rootfs filesystem integrity
4. Check if image was properly extracted

---

## Tips.md Wisdom Applied

From `docs/building/tips.md`, we know that **ping working ≠ everything working**.

### Relevant Failure Categories

**Category 2: Guest OS Networking** (If services are starting but can't bind):
- Guest firewall blocking ports?
- IP forwarding issues?

**Category 3: Guest Kernel Limitations** (If systemd crashes):
- Missing cgroup controllers?
- Wrong cgroup version?
- Missing kernel modules?

**Category 11: Container Behavior vs VM Lifecycle** (Race conditions):
- Services starting before network ready?
- Ephemeral root causing issues?
- Race conditions at boot?

**NOT Relevant**:
- Categories 4-10 (CNI/Docker/containers inside VM) - this is NOT container-in-VM
- Category 1 (TAP/bridge) - already proven working by ping

---

## Fix Strategy Decision Matrix

| Console Log Shows | Root Cause | Fix Required | Priority |
|-------------------|------------|--------------|----------|
| No init messages | Kernel panic | Image/kernel mismatch | P0 |
| "Run /sbin/init" but broken symlink | Init not found | Add `init=` param | P0 |
| "Run /lib/systemd/systemd" SUCCESS | Systemd starts | Debug services | P1 |
| Systemd starts, services fail with specific errors | Service config | Fix service files | P1 |
| Systemd starts, services timeout | Race condition | Fix dependencies | P2 |

---

## Deployment Plan (After Investigation)

### Option A: If Init Parameter Missing
**Script**: `scripts/building/deploy-systemd-init-fix.sh`

**Steps**:
1. Build new daemon binary with init parameter
2. Stop nanofused
3. Deploy binary
4. Delete old VMs
5. Create fresh VM with fixed kernel args
6. Test services

**Validation**:
- `curl http://172.16.0.10:8080/health` → 200 OK
- `curl http://172.16.0.10/` → HTML response

### Option B: If Services Need Fixing
**Requires**: Mount rootfs, fix service files, rebuild image

**Steps**:
1. Extract rootfs from image
2. Mount ext4 filesystem
3. Fix service files
4. Rebuild image
5. Push to GHCR
6. Pull new image
7. Create VM
8. Test

**More Complex**: Requires image rebuild, not just daemon fix

---

## Success Criteria

### Phase 1C: Console Access ✅
- [x] Can read console.log
- [ ] Console log provides debugging info
- [ ] Understand boot sequence

### Phase 1D: Services Working
- [ ] Systemd starts as PID 1
- [ ] `nginx.service` starts successfully
- [ ] `todo-backend.service` starts successfully
- [ ] Port 80 accessible from host
- [ ] Port 8080 accessible from host
- [ ] Health checks pass:
  - `curl http://172.16.0.10:8080/health` → `{"status":"healthy"}`
  - `curl http://172.16.0.10/` → HTML response

---

## Next Actions

### IMMEDIATE (User Action Required):
```bash
# Read console log
sudo tail -200 /var/lib/nanofuse/vms/4803490f-2d7e-4825-8b02-607636b9962d/console.log | tee /tmp/vm-console.txt
```

**Then**:
- Share console log output
- Analyze based on decision tree above
- Choose appropriate fix strategy
- Implement fix
- Deploy and test

### DO NOT:
- ❌ Deploy anything before reading console log
- ❌ Assume fix without evidence
- ❌ Go in circles without new data
- ❌ Move to next phase before services work

---

## Files

- **This Plan**: `docs/building/PHASE1CD_COMPREHENSIVE_PLAN.md`
- **Investigation Data**: `docs/building/PHASE1CD_CONSOLE_ANALYSIS.md` (to be created)
- **Fix Script**: `scripts/building/deploy-systemd-fix.sh` (if init fix needed)
- **Test Script**: `scripts/building/test-services.sh` (to be created)

---

**Status**: Waiting for console log analysis before proceeding.
**Blocker**: Need sudo access to read `/var/lib/nanofuse/vms/*/console.log`
