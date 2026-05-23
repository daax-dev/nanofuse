# Phase 1 Critical Fixes - Execution Plan

**Date**: 2025-11-14
**Status**: PLANNING → EXECUTION
**Goal**: Fix all blocking issues preventing basic functionality

---

## Executive Summary

Phase 1 was declared "complete" but critical functionality is broken. This document outlines the plan to fix 4 blocking issues and establish proper testing to verify functionality.

**Timeline**: 2-3 days
**Approach**: Parallel execution where possible, with human-verifiable tests for each fix

---

## Critical Issues Overview

| Issue | Priority | Blocking | Est. Time | Dependencies |
|-------|----------|----------|-----------|--------------|
| #1: CLI Image Pull Broken | 🔴 CRITICAL | YES | 2-3h | None |
| #2: Unix Socket Not Created | 🔴 CRITICAL | YES | 2-3h | None |
| #3: VM Lifecycle Untested | 🔴 CRITICAL | YES | 4-6h | Issue #1 |
| #4: Demo Script Incomplete | 🔴 CRITICAL | YES | 2h | Issues #1, #2, #3 |

**Total Estimated Time**: 10-14 hours (2 working days)

---

## Issue #1: CLI Image Pull Command Broken

### Current State
```bash
$ nanofuse --api-url http://localhost:8080 image pull --default
Error: Failed to get pull status: Job ID is required
```

### Chain of Thought Analysis

**What we know**:
1. Direct API call works: `curl -X POST -d '{"image_ref":"..."}' /images/pull` → Returns job_id
2. CLI call fails with "Job ID is required"
3. Error appears AFTER "Pulling..." message (order is wrong)
4. API expects field `image_ref`, not `image`

**Investigation needed**:
1. Where does CLI call the API? → `cmd/nanofuse/image.go` or `internal/client/client.go`
2. How does CLI handle the pull response?
3. Where is the job polling logic?
4. Why is error checking happening before job creation?

**Hypothesis**:
- CLI might be calling wrong endpoint
- CLI might be sending wrong JSON field name
- Job polling might be called before job creation completes
- Error handling is in wrong order

### Files to Investigate

Priority order for investigation:
1. `cmd/nanofuse/image.go` - CLI pull command entrypoint
2. `internal/client/client.go` - API client implementation
3. `internal/client/types.go` - Request/response types
4. `internal/api/image_handlers.go:122-180` - API handler (for reference)

### Execution Plan

**Step 1**: Investigate CLI code (use general-purpose agent)
- Find pull command implementation
- Trace the API call path
- Identify the bug

**Step 2**: Fix the identified issue
- Correct field names if wrong
- Fix execution order if reversed
- Ensure proper job polling

**Step 3**: Test the fix
- Manual CLI test with `--default` flag
- Verify job creation
- Verify job polling
- Verify completion/failure handling

**Step 4**: Document the fix
- Update CRITICAL_ISSUES.md with root cause
- Add test case to prevent regression

### Test Plan

**Test Case 1.1: Pull Default Image (Success Path)**
```bash
# Prerequisites:
# - daemon running
# - docker login ghcr.io completed (authentication)

# Test:
nanofuse --api-url http://localhost:8080 image pull --default

# Expected:
# 1. "Pulling ghcr.io/daax-dev/nanofuse/base:latest..." message
# 2. Progress updates or job status
# 3. Success message when complete
# 4. Image appears in `nanofuse image list`

# Verify:
nanofuse --api-url http://localhost:8080 image list
# Should show: ghcr.io/daax-dev/nanofuse/base with latest tag
```

**Test Case 1.2: Pull with Authentication Failure**
```bash
# Prerequisites: NOT logged into ghcr.io

# Test:
nanofuse --api-url http://localhost:8080 image pull --default

# Expected:
# Clear error message about authentication
# Error should mention: "authentication required" or similar
# Should NOT show "Job ID is required"
```

**Test Case 1.3: Pull Custom Image**
```bash
# Test:
nanofuse --api-url http://localhost:8080 image pull ghcr.io/daax-dev/nanofuse/base:v1.0.0

# Expected:
# Pulls specific version
# Shows progress
# Completes successfully
```

### Success Criteria
- ✅ CLI pull command completes without "Job ID" error
- ✅ Job is created and polled correctly
- ✅ Authentication errors are clear and helpful
- ✅ Image appears in image list after successful pull
- ✅ All 3 test cases pass

### Deliverables
1. Fixed code in `cmd/nanofuse/image.go` and/or `internal/client/client.go`
2. Test script: `test/manual/test_image_pull.sh`
3. Documentation: `docs/testing/IMAGE_PULL_TEST.md`

---

## Issue #2: Unix Socket Not Created

### Current State

**Config** (`/etc/nanofuse/nanofused.yaml`):
```yaml
api:
  socket: /tmp/nanofused.sock
  tcp_bind: "127.0.0.1:8080"
```

**Actual behavior**:
- Only TCP listener created on 127.0.0.1:8080
- Unix socket `/tmp/nanofused.sock` never created
- CLI expects socket at `/var/run/nanofused.sock` by default

**Code issue** (`internal/api/server.go:136-173`):
```go
if cfg.API.TCPBind != "" {
    // Creates TCP
} else {
    // Creates Unix socket - NEVER REACHED when TCP is set
}
```

### Chain of Thought Analysis

**Root cause**: if-else logic means only ONE listener is created

**Options to fix**:

**Option A**: Support both listeners simultaneously
- Pros: Flexible, supports both modes at once
- Cons: More complex, potential security implications
- Complexity: Medium (need to manage two listeners)

**Option B**: Priority-based (socket first, TCP fallback)
- Pros: Simple, already partially implemented
- Cons: Can't use both, users must choose
- Complexity: Low (already done)

**Option C**: Make CLI auto-detect and fallback to TCP
- Pros: User-friendly, works with current setup
- Cons: Doesn't solve root cause
- Complexity: Medium (CLI changes)

**Recommended approach**: **Option A + Option C**
- Support both listeners in daemon (Option A)
- Make CLI try socket first, fallback to TCP (Option C)
- Document which mode to use when

**Why both?**
- Development: TCP easier (curl, browser access)
- Production: Unix socket more secure (filesystem permissions)
- Flexibility: Let users choose based on needs

### Execution Plan

**Step 1**: Implement dual listener support in daemon
- Modify `setupListener()` to return array of listeners
- Create both TCP and Unix socket if both configured
- Update server startup to handle multiple listeners
- Update logging to show both addresses

**Step 2**: Update CLI to auto-detect
- Try Unix socket first (if configured)
- Fallback to TCP if socket doesn't exist
- Add clear logging about which transport is used
- Respect explicit `--api-url` or `--api-socket` flags

**Step 3**: Update configuration docs
- Clarify that both can be enabled
- Document security implications
- Provide examples for different scenarios

**Step 4**: Test all configurations
- Socket only
- TCP only
- Both enabled
- CLI auto-detection

### Test Plan

**Test Case 2.1: Unix Socket Only**
```bash
# Config: socket set, tcp_bind empty
# /etc/nanofuse/nanofused.yaml:
# api:
#   socket: /tmp/nanofused.sock
#   tcp_bind: ""

# Restart daemon
sudo systemctl restart nanofused

# Verify socket exists
ls -la /tmp/nanofused.sock
# Expected: srwxrwxrwx ... /tmp/nanofused.sock

# Verify TCP not listening
sudo lsof -i:8080
# Expected: (empty - no process)

# Test CLI (should auto-detect socket)
nanofuse image list
# Expected: Works without --api-url flag

# Test curl via socket
curl --unix-socket /tmp/nanofused.sock http://localhost/health
# Expected: {"status":"healthy",...}
```

**Test Case 2.2: TCP Only**
```bash
# Config: socket empty, tcp_bind set
# api:
#   socket: ""
#   tcp_bind: "127.0.0.1:8080"

# Restart daemon
sudo systemctl restart nanofused

# Verify TCP listening
curl http://localhost:8080/health
# Expected: {"status":"healthy",...}

# Verify socket doesn't exist
ls /tmp/nanofused.sock
# Expected: No such file

# Test CLI (should auto-detect TCP fallback)
nanofuse image list
# Expected: May fail OR auto-detect TCP (depending on implementation)

# Test with explicit flag
nanofuse --api-url http://localhost:8080 image list
# Expected: Works
```

**Test Case 2.3: Both Enabled (Dual Listener)**
```bash
# Config: both set
# api:
#   socket: /tmp/nanofused.sock
#   tcp_bind: "127.0.0.1:8080"

# Restart daemon
sudo systemctl restart nanofused

# Check daemon logs
sudo journalctl -u nanofused -n 20
# Expected:
# "Listening on Unix socket: /tmp/nanofused.sock"
# "Listening on TCP: 127.0.0.1:8080"

# Verify both work
curl --unix-socket /tmp/nanofused.sock http://localhost/health
curl http://localhost:8080/health
# Both should return: {"status":"healthy",...}

# Verify socket exists
ls -la /tmp/nanofused.sock
# Expected: Socket file exists

# Verify TCP listening
sudo lsof -i:8080
# Expected: nanofused listening

# Test CLI auto-detection
nanofuse image list
# Expected: Works using socket (preferred)

# Test CLI TCP override
nanofuse --api-url http://localhost:8080 image list
# Expected: Works using TCP
```

### Success Criteria
- ✅ Can configure Unix socket only
- ✅ Can configure TCP only
- ✅ Can configure both simultaneously
- ✅ Both listeners work independently when both enabled
- ✅ CLI auto-detects socket and falls back to TCP
- ✅ All 3 test cases pass
- ✅ Daemon logs clearly show which listeners are active

### Deliverables
1. Updated `internal/api/server.go` with dual listener support
2. Updated CLI client with auto-detection in `internal/client/client.go`
3. Test script: `test/manual/test_listeners.sh`
4. Documentation: `docs/testing/LISTENER_TEST.md`
5. Updated config examples in `config/nanofused.yaml.example`

---

## Issue #3: VM Lifecycle End-to-End Testing

### Current State

**Never tested**:
- Pull image → Create VM → Start VM → Check status → Get logs → Stop VM → Delete VM

**Dependencies**:
- Requires Issue #1 fixed (image pull working)
- Requires Issue #2 fixed (CLI working reliably)
- Requires proper authentication setup

### Chain of Thought Analysis

**Why this is critical**:
- Core value proposition of NanoFuse
- Every component must work together
- Network, storage, Firecracker, database all involved
- Real-world usage simulation

**What could fail**:
1. Image pull authentication
2. VM creation (disk, network, firecracker config)
3. Network setup (TAP, bridge, IPAM)
4. Firecracker boot process
5. VM connectivity (can't reach internet)
6. Graceful shutdown
7. Resource cleanup

**Test coverage needed**:
- Happy path (everything works)
- Error paths (what happens when things fail)
- Resource limits (max VMs, memory)
- Concurrent operations (multiple VMs)

### Execution Plan

**Step 1**: Setup authentication
```bash
# Get GitHub token with read:packages scope
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

**Step 2**: Create comprehensive test script
- Pull image via CLI
- Create VM with specific config
- Start VM
- Wait for boot (check logs)
- Verify VM running (status check)
- Test network connectivity (ping from VM)
- Get console logs
- Stop VM gracefully
- Verify cleanup
- Delete VM
- Verify complete cleanup (TAP device gone, DB entry removed)

**Step 3**: Run test with verbose logging
- Capture all output
- Log timing information
- Check for errors/warnings

**Step 4**: Document every step with expected output
- Create markdown test guide
- Include troubleshooting section
- Add verification commands

### Test Plan

**Test Case 3.1: Complete Happy Path**

```bash
#!/bin/bash
# test/manual/test_vm_lifecycle.sh

set -e

echo "=== VM Lifecycle Test ==="

# Cleanup from previous runs
echo "1. Cleanup any existing test VMs..."
nanofuse vm delete test-vm-lifecycle 2>/dev/null || true

# Verify no test VMs exist
VM_COUNT=$(nanofuse vm list --json | jq '.vms | length')
echo "   Current VMs: $VM_COUNT"

# Pull image
echo "2. Pull base image..."
nanofuse image pull --default
# Expected: Job created, polling, success
# Wait for completion

# Verify image pulled
echo "3. Verify image exists..."
nanofuse image list
# Expected: Shows ghcr.io/daax-dev/nanofuse/base:latest

# Get image digest
IMAGE=$(nanofuse image list --json | jq -r '.images[0].digest')
echo "   Image digest: $IMAGE"

# Create VM
echo "4. Create VM..."
nanofuse vm create test-vm-lifecycle default --vcpus 2 --memory 512
# Expected: VM created with ID

# Get VM ID
VM_ID=$(nanofuse vm list --json | jq -r '.vms[] | select(.name=="test-vm-lifecycle") | .id')
echo "   VM ID: $VM_ID"

# Check initial state
echo "5. Check VM state..."
nanofuse vm status test-vm-lifecycle
# Expected: State = "created" or "stopped"

# Start VM
echo "6. Start VM..."
nanofuse vm start test-vm-lifecycle
# Expected: VM starting...

# Wait for boot
echo "7. Wait for boot (10 seconds)..."
sleep 10

# Check running state
echo "8. Check VM running..."
nanofuse vm status test-vm-lifecycle
# Expected: State = "running"

# Get VM IP
VM_IP=$(nanofuse vm status test-vm-lifecycle --json | jq -r '.network.ip')
echo "   VM IP: $VM_IP"

# Verify IP allocated from pool
if [[ "$VM_IP" =~ ^172\.16\.0\.([0-9]{1,3})$ ]]; then
    echo "   ✓ IP in correct range (172.16.0.0/24)"
else
    echo "   ✗ IP NOT in expected range!"
    exit 1
fi

# Check console logs
echo "9. Check console logs..."
nanofuse vm logs test-vm-lifecycle --tail 20
# Expected: Boot messages, systemd output

# Verify network (from host)
echo "10. Test network connectivity (ping VM from host)..."
ping -c 3 $VM_IP
# Expected: 3 packets transmitted, 3 received

# Test VM can reach internet (optional, requires SSH access)
# echo "11. Test VM internet access..."
# ssh to VM and ping 8.8.8.8

# Stop VM
echo "11. Stop VM gracefully..."
nanofuse vm stop test-vm-lifecycle
# Expected: Stopping... Done

# Wait for shutdown
echo "12. Wait for shutdown (5 seconds)..."
sleep 5

# Verify stopped
echo "13. Verify VM stopped..."
nanofuse vm status test-vm-lifecycle
# Expected: State = "stopped"

# Check TAP device still exists (should be cleaned up on delete, not stop)
TAP_DEVICE=$(ip link show | grep tap-$VM_ID || echo "")
if [ -n "$TAP_DEVICE" ]; then
    echo "   TAP device still exists (expected before delete)"
else
    echo "   ✗ TAP device gone too early!"
fi

# Delete VM
echo "14. Delete VM..."
nanofuse vm delete test-vm-lifecycle
# Expected: VM deleted

# Verify VM gone
echo "15. Verify VM deleted..."
nanofuse vm list
# Expected: Empty list or no test-vm-lifecycle

# Verify TAP device cleaned up
echo "16. Verify TAP device cleaned up..."
TAP_DEVICE=$(ip link show | grep tap-$VM_ID || echo "")
if [ -z "$TAP_DEVICE" ]; then
    echo "   ✓ TAP device cleaned up"
else
    echo "   ✗ TAP device still exists!"
    exit 1
fi

# Verify IPAM released IP
echo "17. Create another VM to verify IP reuse..."
nanofuse vm create test-vm-lifecycle-2 default --vcpus 2 --memory 512
nanofuse vm start test-vm-lifecycle-2
sleep 5
VM_IP_2=$(nanofuse vm status test-vm-lifecycle-2 --json | jq -r '.network.ip')
echo "   New VM IP: $VM_IP_2"
# Expected: Should get same IP (172.16.0.10 - first in pool) since released

# Cleanup
nanofuse vm stop test-vm-lifecycle-2
nanofuse vm delete test-vm-lifecycle-2

echo ""
echo "=== VM Lifecycle Test PASSED ==="
```

**Test Case 3.2: Concurrent VMs**

```bash
#!/bin/bash
# test/manual/test_concurrent_vms.sh

echo "=== Testing Concurrent VMs ==="

# Create 3 VMs
for i in 1 2 3; do
    echo "Creating VM $i..."
    nanofuse vm create test-concurrent-$i default --vcpus 1 --memory 256
    nanofuse vm start test-concurrent-$i
done

# Wait for all to boot
sleep 15

# Check all running
echo "Checking all VMs..."
for i in 1 2 3; do
    STATUS=$(nanofuse vm status test-concurrent-$i --json | jq -r '.state')
    IP=$(nanofuse vm status test-concurrent-$i --json | jq -r '.network.ip')
    echo "VM $i: $STATUS at $IP"
done

# Verify unique IPs
IPS=$(nanofuse vm list --json | jq -r '.vms[].network.ip' | sort | uniq -d)
if [ -z "$IPS" ]; then
    echo "✓ All IPs unique"
else
    echo "✗ Duplicate IPs found: $IPS"
    exit 1
fi

# Cleanup
for i in 1 2 3; do
    nanofuse vm stop test-concurrent-$i
    nanofuse vm delete test-concurrent-$i
done

echo "=== Concurrent Test PASSED ==="
```

**Test Case 3.3: Error Handling**

```bash
#!/bin/bash
# test/manual/test_error_handling.sh

echo "=== Testing Error Handling ==="

# Test 1: Create VM with non-existent image
echo "Test: Create VM with bad image..."
nanofuse vm create test-bad-image nonexistent:tag --vcpus 2 --memory 512 2>&1 | tee /tmp/error.log
# Expected: Clear error message about image not found

# Test 2: Start already running VM
echo "Test: Start running VM..."
nanofuse vm create test-error default --vcpus 2 --memory 512
nanofuse vm start test-error
sleep 5
nanofuse vm start test-error 2>&1 | tee -a /tmp/error.log
# Expected: Error about VM already running

# Test 3: Delete running VM (should fail or force)
echo "Test: Delete running VM..."
nanofuse vm delete test-error 2>&1 | tee -a /tmp/error.log
# Expected: Error or requires --force flag

# Cleanup
nanofuse vm stop test-error
nanofuse vm delete test-error

echo "=== Error Handling Test PASSED ==="
```

### Success Criteria
- ✅ Complete lifecycle test passes (3.1)
- ✅ Can run 3+ concurrent VMs with unique IPs (3.2)
- ✅ Error cases handled gracefully (3.3)
- ✅ Network connectivity verified (ping works)
- ✅ TAP devices cleaned up properly
- ✅ IPAM releases and reuses IPs correctly
- ✅ Console logs accessible
- ✅ No leaked resources (check with `ip link`, `ps aux | grep firecracker`)

### Deliverables
1. Test scripts: `test/manual/test_vm_lifecycle.sh`
2. Test scripts: `test/manual/test_concurrent_vms.sh`
3. Test scripts: `test/manual/test_error_handling.sh`
4. Documentation: `docs/testing/VM_LIFECYCLE_TEST.md`
5. Troubleshooting guide: `docs/testing/TROUBLESHOOTING.md`

---

## Issue #4: Demo Script Complete Workflow

### Current State

**Partially working**:
- ✅ Health check fixed
- ✅ TCP fallback added
- ❌ Cannot complete without images
- ❌ Never tested end-to-end

### Execution Plan

**Step 1**: Update demo script with fixes from Issues #1-3

**Step 2**: Add setup instructions
- Authentication requirements
- How to pull images first
- What to expect at each step

**Step 3**: Add verification at each step
- Check return codes
- Verify state changes
- Confirm cleanup

**Step 4**: Run complete demo
- Document output
- Capture timing
- Screenshot key moments

### Test Plan

**Test Case 4.1: Demo Script End-to-End**

```bash
# Prerequisites:
# 1. Daemon running
# 2. Authenticated to GHCR
# 3. No existing VMs

# Run demo
cd /home/jpoley/ps/nanofuse/examples
sudo ./api-demo.sh

# Expected output:
# ===> 1. Checking API health...
# ✓ API is healthy
#
# ===> 2. Listing current VMs...
# ✓ Found 0 existing VMs
#
# ===> 3. Checking for images...
# ✓ Found 1 images
# ✓ Using image: ghcr.io/daax-dev/nanofuse/base:latest
#
# ===> 4. Creating VM...
# ✓ Created VM: demo-vm-1731612309 (ID: vm-abc123)
#
# ===> 5. Starting VM...
# ✓ VM started
#
# ===> 6. Waiting for VM to boot (5 seconds)...
#
# ===> 7. Checking VM status...
# ✓ VM state: running
# ✓ VM IP: 172.16.0.10
#
# ===> 8. Getting VM console logs (last 10 lines)...
# [    0.000000] Linux version 6.1.0...
# [    0.123456] systemd[1]: Reached target basic.target
# ...
# ✓ Logs retrieved
#
# VM is running. Waiting 10 seconds before cleanup...
#
# ===> 9. Stopping VM...
# ✓ VM stopped
#
# ===> 10. Deleting VM...
# ✓ VM deleted
#
# ===> Demo Complete!
# Successfully demonstrated complete VM lifecycle:
#   ✓ Health check
#   ✓ VM creation
#   ✓ VM start
#   ✓ Status check
#   ✓ Log retrieval
#   ✓ VM stop
#   ✓ VM deletion

# Verify no resources leaked
sudo ip link show | grep tap-
# Expected: (empty)

ps aux | grep firecracker | grep -v grep
# Expected: (empty)
```

### Success Criteria
- ✅ Demo script runs to completion without errors
- ✅ All steps succeed with ✓ checkmarks
- ✅ VM created, started, stopped, deleted successfully
- ✅ Network and resources cleaned up
- ✅ Demo can be run multiple times consecutively
- ✅ Clear error messages if prerequisites missing

### Deliverables
1. Updated `examples/api-demo.sh` with robust error handling
2. Setup guide: `examples/API_DEMO_SETUP.md`
3. Expected output doc: `examples/API_DEMO_OUTPUT.md`

---

## Parallel Execution Strategy

### Timeline (Gantt Chart)

```
Day 1:
Hour 0-3:  [Issue #1: CLI Pull] + [Issue #2: Unix Socket] (PARALLEL)
Hour 3-4:  Test Issue #1 and #2 fixes
Hour 4-5:  Documentation for #1 and #2

Day 2:
Hour 0-4:  [Issue #3: VM Lifecycle] (Sequential - depends on #1)
Hour 4-5:  Test Issue #3
Hour 5-6:  Documentation for #3

Hour 6-8:  [Issue #4: Demo Script] (Sequential - depends on all)
Hour 8:    Final end-to-end verification
```

### Agent Assignment

**Track A (CLI Pull)**:
- Use `general-purpose` agent to investigate CLI code
- Agent task: "Find and debug the CLI image pull command bug"
- Expected output: Root cause analysis and suggested fix

**Track B (Unix Socket)**:
- Direct implementation (well-understood problem)
- No agent needed (straightforward code change)

**Track C (VM Lifecycle)**:
- Manual testing with detailed script
- Use `Explore` agent if needed to find test examples

**Track D (Demo Script)**:
- Manual update and testing

---

## Success Criteria - Overall

### Definition of "Phase 1 Complete"

Phase 1 can ONLY be marked complete when ALL of the following are true:

- ✅ CLI image pull works (`nanofuse image pull --default` succeeds)
- ✅ Unix socket created when configured
- ✅ CLI auto-detects socket/TCP
- ✅ VM lifecycle test passes (create → start → stop → delete)
- ✅ Concurrent VM test passes (3+ VMs with unique IPs)
- ✅ Error handling test passes (graceful failures)
- ✅ Demo script runs to completion
- ✅ Network connectivity verified (ping works)
- ✅ Resource cleanup verified (no leaked TAP devices, processes)
- ✅ All test documentation written
- ✅ No "production-ready" claims in docs
- ✅ CRITICAL_ISSUES.md shows all issues resolved

### Test Coverage Matrix

| Component | Unit Tests | Integration Tests | Manual Tests | E2E Tests |
|-----------|-----------|-------------------|--------------|-----------|
| CLI | Existing | ❌ Missing | ✅ Added | ✅ Demo |
| API | Existing | ❌ Missing | ✅ Added | ✅ Demo |
| Image Pull | Existing | ❌ Missing | ✅ Added | ✅ Demo |
| VM Lifecycle | Existing | ❌ Missing | ✅ Added | ✅ Demo |
| Network | ❌ Partial | ❌ Missing | ✅ Added | ✅ Demo |
| Cleanup | ❌ Missing | ❌ Missing | ✅ Added | ✅ Demo |

**Goal**: Fill all ❌ with ✅ or ⚠️ (documented limitation)

---

## Documentation Updates

### Files to Update

1. **README.md**
   - Remove "Production-ready" claim (line 307, 335)
   - Change to "Alpha - Testing Phase"
   - Add "Known Limitations" section
   - Link to test documentation

2. **ROADMAP.md**
   - Move from "Phase 1 Complete" to "Phase 1 - Testing & Fixes"
   - Update status from 100% to realistic % (current: ~60%)
   - Add "Critical Fixes" section at top

3. **DONE.md**
   - Move to `docs/archive/DONE_PLANNED.md` (what was planned)
   - Create new `DONE.md` with actual verified features

4. **docs/testing/** (NEW)
   - `IMAGE_PULL_TEST.md`
   - `LISTENER_TEST.md`
   - `VM_LIFECYCLE_TEST.md`
   - `TROUBLESHOOTING.md`
   - `TEST_INDEX.md` (links to all tests)

---

## Risk Management

### Potential Blockers

1. **Authentication Issues**
   - Risk: Can't pull images from GHCR
   - Mitigation: Document token setup clearly, use public mirror if needed

2. **Firecracker Permissions**
   - Risk: KVM access issues
   - Mitigation: Verify /dev/kvm permissions, document requirements

3. **Network Issues**
   - Risk: Bridge/NAT setup fails
   - Mitigation: Add network troubleshooting guide

4. **Time Overruns**
   - Risk: Fixes take longer than estimated
   - Mitigation: Parallel execution, prioritize critical path

---

## Rollout Plan

### Phase 1: Fixes (Day 1-2)
- Fix all 4 critical issues
- Write all test scripts
- Initial documentation

### Phase 2: Testing (Day 2-3)
- Run all manual tests
- Document results
- Fix any issues found

### Phase 3: Documentation (Day 3)
- Complete all test docs
- Update README/ROADMAP
- Remove production-ready claims

### Phase 4: Verification (Day 3)
- Fresh clone test (clean environment)
- Run all tests in sequence
- Verify docs match reality

---

## Verification Checklist

Before declaring Phase 1 complete, verify:

- [ ] Fresh clone on clean VM
- [ ] Follow installation docs exactly
- [ ] Run all 4 test cases
- [ ] Run demo script
- [ ] Check all links in docs work
- [ ] Verify no broken references
- [ ] Check no "production" claims remain
- [ ] All TODOs in code addressed or documented
- [ ] All test scripts have accompanying .md docs
- [ ] CRITICAL_ISSUES.md updated with resolutions

---

## Next Steps

1. **Immediate**: Start parallel execution
   - Launch agent for CLI investigation
   - Start Unix socket fix

2. **Within 1 hour**: Complete fixes for #1 and #2

3. **Within 4 hours**: Begin VM lifecycle testing

4. **Within 8 hours**: Complete demo script verification

5. **Within 12 hours**: All documentation complete

6. **Within 24 hours**: Fresh environment test

---

## References

- `CRITICAL_ISSUES.md` - Issue details
- `internal/api/server.go:136-173` - Unix socket bug
- `cmd/nanofuse/image.go` - CLI pull command
- `examples/api-demo.sh` - Demo script
- `README.md:307` - Production-ready claim to remove
- `ROADMAP.md:27` - Status to update
