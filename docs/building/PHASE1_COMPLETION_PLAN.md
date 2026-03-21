# Phase 1 Completion Plan - Systematic Approach

**Created**: 2025-11-19
**Goal**: Complete Phase 1 with 100% reliability - simple install + run works every time
**Status**: IN PROGRESS

---

## Executive Summary

Phase 1 is approximately 60-70% complete. The infrastructure is built but **basic functionality is broken or untested**. This document outlines a systematic, evidence-based approach to reach 100% completion with full reliability.

### Current Reality Check

**What Works** ✅:
- Daemon starts and runs as systemd service
- HTTP API server accepts connections
- Database schema and migrations
- Network bridge (nanofuse0) configured
- NAT rules for internet access
- Base image builds in CI/CD
- Binary releases automated

**What's Broken** ❌:
- VM state inconsistency (DB says "running" but no Firecracker process exists)
- VM config fields all showing as 0/empty in database
- VMs not reachable by ping (network not working)
- CLI image pull command fails
- Systemd services don't start in VMs
- No console/serial access to see what's happening
- End-to-end workflow never tested successfully

**Confidence Level**: Can build and start daemon, but cannot reliably run applications in VMs.

---

## Phase 1 Success Criteria (Non-Negotiable)

Phase 1 is NOT complete until ALL of these work reliably, every time:

### 1. Installation & Setup
- [ ] Fresh install from binaries works
- [ ] `sudo systemctl enable --now nanofused` succeeds
- [ ] Daemon starts without errors
- [ ] Network infrastructure auto-configures
- [ ] Database initializes correctly

### 2. Image Management
- [ ] `nanofuse image pull --default` works with authentication
- [ ] Images are stored in `/var/lib/nanofuse/images/`
- [ ] `nanofuse image list` shows pulled images
- [ ] Both tag-based and digest-based pulls work

### 3. VM Lifecycle (Base Image)
- [ ] `nanofuse vm create default test-vm` succeeds
- [ ] `nanofuse vm start test-vm` boots VM successfully
- [ ] VM responds to ping within 5 seconds
- [ ] `nanofuse vm stop test-vm` stops gracefully
- [ ] `nanofuse vm delete test-vm` cleans up all resources
- [ ] `nanofuse vm list` shows accurate state

### 4. Application Workloads (Todo-App)
- [ ] Can build custom application image
- [ ] Can pull/register custom image
- [ ] Can create VM from custom image
- [ ] VM boots with systemd as PID 1
- [ ] Systemd services start automatically
- [ ] HTTP services are accessible from host
- [ ] `curl http://<VM_IP>/health` returns 200 OK
- [ ] Application functions correctly

### 5. Network & Connectivity
- [ ] VMs get IP addresses from IPAM pool
- [ ] Host can ping VMs (sub-millisecond latency)
- [ ] VMs can reach internet
- [ ] Multiple VMs can run simultaneously
- [ ] TAP devices are created/cleaned up correctly

### 6. State Management
- [ ] VM state in database matches reality
- [ ] Config (vcpus, memory, kernel_args) persists correctly
- [ ] Daemon restart doesn't corrupt state
- [ ] Orphaned Firecracker processes don't exist

### 7. Observability
- [ ] Console/serial logs accessible
- [ ] Can see VM boot sequence
- [ ] Can see systemd startup messages
- [ ] Can see application logs
- [ ] Daemon logs are useful for debugging

### 8. Error Handling
- [ ] Clear error messages when things fail
- [ ] No silent failures
- [ ] Timeouts are reasonable
- [ ] Resources cleaned up on error

### 9. Documentation
- [ ] README instructions actually work
- [ ] No "coming soon" for Phase 1 features
- [ ] Known issues documented
- [ ] Troubleshooting guide exists

### 10. Repeatability
- [ ] Test suite validates all above
- [ ] Can run `make test-e2e` and it passes
- [ ] Same steps work on clean system
- [ ] No manual workarounds required

**Current Score**: 4/10 criteria fully met (40%)

---

## Root Cause Analysis (Evidence-Based)

### Finding #1: VM State Corruption
**Evidence**:
```bash
$ nanofuse vm list --json
"state": "running"
"config": {"vcpus": 0, "memory_mib": 0, "kernel_args": ""}

$ ps aux | grep firecracker
# No processes found

$ ping 172.16.0.11
Destination Host Unreachable
```

**Analysis**: The database shows VM as "running" with empty config, but no Firecracker process exists and VM is not reachable. This indicates:
1. VM was never actually started, OR
2. VM crashed and state wasn't updated, OR
3. VM config wasn't saved when created

**Impact**: CRITICAL - Cannot trust VM state
**Root Cause**: Unknown - needs investigation in `internal/storage` and `internal/api`

### Finding #2: Config Not Persisted
**Evidence**: All config fields (vcpus, memory, kernel_args) show as 0 or empty when listing VMs.

**Analysis**: Either:
1. Config not being saved during VM creation
2. Config not being loaded during VM list
3. Database schema issue

**Impact**: CRITICAL - Cannot configure VMs
**Root Cause**: Likely in `internal/storage/db.go` VM CRUD operations

### Finding #3: Systemd Services Not Starting
**Evidence**: From PRIORITY_TODO.md - todo-app VM boots but ports 80 and 8080 closed.

**Analysis**: Without console access, we're blind to what's happening. Possible causes:
1. Systemd not running as PID 1
2. `/lib/systemd/systemd` doesn't exist or isn't executable
3. Services enabled but target not reached
4. Missing dependencies (dbus, cgroups, etc.)

**Impact**: CRITICAL - Cannot run applications
**Root Cause**: Unknown - need console access first

### Finding #4: No Console Access
**Evidence**: No way to see VM boot logs or systemd output.

**Analysis**: Firecracker supports serial console, but NanoFuse doesn't expose it.

**Impact**: HIGH - Cannot debug VM issues
**Root Cause**: Feature not implemented

### Finding #5: CLI Image Pull Broken
**Evidence**: From CRITICAL_ISSUES.md - CLI returns "Job ID is required" error.

**Analysis**: API endpoint works, CLI client doesn't. Likely polling logic or error handling bug.

**Impact**: HIGH - Cannot use CLI to pull images
**Root Cause**: Likely in `internal/client/client.go` or `cmd/nanofuse/image.go`

---

## Systematic Approach (Chain of Thought)

### Principle: Fix Foundation Before Building Higher

We must fix issues in this order:
1. **Observability** - Can't fix what we can't see
2. **State Management** - Can't trust system without accurate state
3. **Basic VM Lifecycle** - Must work with simple base image first
4. **Application Workloads** - Only after basics work

### Why This Order?

**Step 1: Observability (Console Access)**
- Without console logs, we're guessing
- Systemd debugging is impossible without output
- Must see VM boot sequence to understand failures
- **Decision**: Implement console logging FIRST

**Step 2: State Management**
- Can't debug VMs if state is wrong
- Must fix config persistence
- Must fix state synchronization
- **Decision**: Fix database CRUD operations SECOND

**Step 3: Clean Slate**
- Current VM state is corrupted
- Need to delete and recreate from scratch
- **Decision**: Clean up, start fresh THIRD

**Step 4: Basic VM Lifecycle**
- Use simple base image (no systemd complexity)
- Prove VM can boot, be pinged, stopped, deleted
- **Decision**: Validate basic workflow FOURTH

**Step 5: Systemd & Applications**
- Only tackle systemd after basics work
- Use console logs to debug systemd issues
- **Decision**: Fix application workloads FIFTH

---

## Detailed Execution Plan

### Phase 1A: Fix Critical CLI/API Bugs (2-3 hours)

**Objective**: Fix known bugs blocking basic usage

**Tasks**:
1. **Fix CLI image pull command** (1-2 hours)
   - Read: `cmd/nanofuse/image.go` and `internal/client/client.go`
   - Identify: Where job polling fails
   - Fix: Job status polling logic
   - Test: `nanofuse image pull --default`
   - Document: Decision and fix in this file

2. **Fix Unix socket + TCP listener issue** (1 hour)
   - Read: `internal/api/server.go`
   - Identify: if-else logic preventing dual listeners
   - Fix: Support both simultaneously OR document limitation
   - Test: Both socket and TCP work
   - Document: Decision and configuration guidance

**Exit Criteria**:
- [ ] `nanofuse image pull --default` succeeds
- [ ] Both Unix socket and TCP listeners work
- [ ] No blocking CLI bugs

### Phase 1B: Fix VM State Management (4-6 hours)

**Objective**: Ensure VM state in database matches reality

**Tasks**:
1. **Investigate config persistence** (2 hours)
   - Read: `internal/storage/db.go` VM CRUD operations
   - Read: `internal/api/handlers.go` VM creation handler
   - Identify: Where config is saved and loaded
   - Test: Create VM and check if config persists
   - Document: Findings

2. **Fix config saving** (1-2 hours)
   - Fix: Database INSERT/UPDATE for VM config
   - Test: Create VM, list VMs, verify config matches
   - Document: Fix applied

3. **Fix state synchronization** (1-2 hours)
   - Ensure: VM state updates when Firecracker process exits
   - Add: Process monitoring if needed
   - Test: Start VM, kill firecracker process, check state
   - Document: Solution

**Exit Criteria**:
- [ ] VM config (vcpus, memory, kernel_args) persists correctly
- [ ] VM state reflects actual Firecracker process state
- [ ] Database queries return accurate information

### Phase 1C: Implement Console Access (3-4 hours)

**Objective**: Add ability to see VM boot logs and console output

**Tasks**:
1. **Research Firecracker serial console** (30 min)
   - Read: Firecracker documentation on serial console
   - Identify: How to capture output to file or socket

2. **Implement console logging** (2-3 hours)
   - Modify: `internal/firecracker/vm.go` to configure serial console
   - Add: Log file path to VM config
   - Store: Console logs in `/var/lib/nanofuse/vms/<id>/console.log`
   - Test: Start VM and see boot messages in log file

3. **Add CLI command to view logs** (30 min)
   - Add: `nanofuse vm logs <name>` command
   - Implement: Tail functionality
   - Test: Can see real-time logs

**Exit Criteria**:
- [ ] VM console output captured to log file
- [ ] `nanofuse vm logs <name>` shows boot messages
- [ ] Can see kernel, systemd, and application output

### Phase 1D: Clean Slate & Basic VM Lifecycle (2-3 hours)

**Objective**: Prove basic VM workflow with simple base image

**Tasks**:
1. **Clean up corrupted state** (30 min)
   - Stop: All VMs
   - Delete: All VMs from database
   - Clean: Any orphaned Firecracker processes
   - Document: Clean state procedure

2. **Pull fresh base image** (30 min)
   - Use: Fixed CLI pull command
   - Verify: Image stored correctly
   - Document: Image digest and paths

3. **Test basic VM lifecycle** (1-2 hours)
   - Create: Simple VM from base image
   - Start: VM and wait for boot
   - Verify: VM responds to ping
   - Test: Can connect to VM (if SSH enabled)
   - Stop: VM gracefully
   - Verify: State updates correctly
   - Delete: VM and verify cleanup
   - Document: Every step and result

**Exit Criteria**:
- [ ] Can create VM from base image
- [ ] VM boots and is reachable via network
- [ ] VM can be stopped and deleted
- [ ] State remains consistent throughout lifecycle
- [ ] No orphaned resources

### Phase 1E: Fix Systemd in Todo-App (4-6 hours)

**Objective**: Get systemd services running in custom application images

**Tasks**:
1. **Inspect todo-app rootfs** (1 hour)
   - Mount: `/var/lib/nanofuse/images/sha256:0c8543.../rootfs.ext4`
   - Check: `/lib/systemd/systemd` exists and is executable
   - Check: Service files in `/etc/systemd/system/`
   - Check: Default target symlink
   - Check: Application binary `/usr/local/bin/todo-server`
   - Document: Findings

2. **Create fresh todo-app VM with console** (30 min)
   - Delete: Old corrupted VM
   - Create: New VM with proper config
   - Start: VM and watch console logs
   - Document: What we see during boot

3. **Debug systemd startup** (2-3 hours)
   - Analyze: Console logs for errors
   - Try: Different kernel args if needed
   - Try: Simplified init script to test
   - Fix: Issues identified
   - Rebuild: Image if necessary
   - Document: Root cause and solution

4. **Verify application works** (1 hour)
   - Test: Services start automatically
   - Test: `curl http://<VM_IP>/health`
   - Test: Full CRUD operations
   - Document: Success

**Exit Criteria**:
- [ ] Todo-app VM boots with systemd as PID 1
- [ ] Nginx and todo-backend services start
- [ ] HTTP endpoints respond correctly
- [ ] Application functions as expected

### Phase 1F: Create End-to-End Test Suite (4-6 hours)

**Objective**: Automated tests that validate everything

**Tasks**:
1. **Design test structure** (1 hour)
   - Define: Test scenarios
   - Choose: Testing framework (Go test or bash)
   - Plan: Test data and cleanup

2. **Implement tests** (2-3 hours)
   - Test: Image pull
   - Test: VM creation
   - Test: VM start and network
   - Test: VM stop and delete
   - Test: Application workload
   - Test: Error cases
   - Test: Resource cleanup

3. **Create test automation** (1-2 hours)
   - Add: `make test-e2e` target
   - Add: CI integration (optional)
   - Document: How to run tests

**Exit Criteria**:
- [ ] Comprehensive test suite exists
- [ ] All tests pass consistently
- [ ] Tests catch regressions

### Phase 1G: Documentation & Validation (2-3 hours)

**Objective**: Documentation reflects reality, system is production-ready

**Tasks**:
1. **Update core docs** (1-2 hours)
   - README.md: Fix any inaccuracies
   - ROADMAP.md: Update Phase 1 status
   - CRITICAL_ISSUES.md: Close fixed issues
   - Add: Troubleshooting guide

2. **Create operational docs** (1 hour)
   - Add: Installation guide (tested)
   - Add: Configuration guide
   - Add: Common issues and solutions

3. **Final validation** (1 hour)
   - Run: Full test suite
   - Test: Fresh install on clean system
   - Verify: All success criteria met
   - Document: Phase 1 COMPLETE

**Exit Criteria**:
- [ ] All documentation accurate
- [ ] No "coming soon" for Phase 1 features
- [ ] Installation guide tested
- [ ] All 10 success criteria met

---

## Testing Strategy

### Test Levels

**1. Unit Tests** (Already exist)
- Go test suite for packages
- Run: `go test ./...`

**2. Integration Tests** (To be created)
- API endpoints with real daemon
- Database operations
- Network setup/teardown

**3. End-to-End Tests** (To be created)
- Full VM lifecycle
- Application workloads
- Multi-VM scenarios

**4. Manual Acceptance Tests** (To be created)
- Follow README exactly
- Document every step
- Verify user experience

### Test Documentation

For every test:
- **What**: What are we testing?
- **Why**: Why does this matter?
- **How**: Exact commands to run
- **Expected**: What should happen
- **Actual**: What did happen
- **Pass/Fail**: Clear verdict

---

## Risk Management

### High Risk Items

1. **Firecracker Version Compatibility**
   - Risk: Kernel or Firecracker version mismatch
   - Mitigation: Document tested versions, add version checks

2. **Systemd in MicroVMs**
   - Risk: Systemd might not work reliably in Firecracker
   - Mitigation: Research existing solutions, consider alternatives

3. **State Corruption**
   - Risk: Database state diverges from reality
   - Mitigation: Add health checks, reconciliation logic

4. **Network Configuration**
   - Risk: TAP devices or bridge config fails
   - Mitigation: Better error messages, automated setup script

### Contingency Plans

**If systemd doesn't work**:
- Option 1: Use simpler init (runit, s6)
- Option 2: Run services directly from init script
- Option 3: Use alpine-based images without systemd

**If state management is too complex**:
- Option 1: Simplify state model
- Option 2: Add reconciliation loop
- Option 3: Make daemon stateless, query Firecracker directly

---

## Decision Log

This section will be updated as we make decisions during execution.

### Decision #1: [To be filled]

**Date**:
**Context**:
**Options Considered**:
**Decision**:
**Rationale**:
**Outcome**:

---

## Progress Tracking

| Phase | Status | Started | Completed | Notes |
|-------|--------|---------|-----------|-------|
| 1A: CLI/API Bugs | ✅ Complete | 2025-11-19 | 2025-11-19 | CLI pull works. Unix socket fixed. See PHASE1A_FINDINGS.md |
| 1B: State Management | ✅ Partial | 2025-11-19 | 2025-11-19 | Config bug fixed. Discovered image path issues. See PHASE1B_COMPLETE.md |
| 1C: Console Access | Not Started | | | BLOCKING: Need logs to debug VM crashes |
| 1D: Basic Lifecycle | Not Started | | | Blocked by 1C |
| 1E: Systemd/Apps | Not Started | | | Blocked by 1C, 1D |
| 1F: Test Suite | Not Started | | | |
| 1G: Documentation | Not Started | | | |

---

## Next Session Checklist

When resuming work:
1. Read this document fully
2. Check current state: `nanofuse vm list`
3. Check daemon: `systemctl status nanofused`
4. Review progress tracking table
5. Continue with next pending phase
6. Update decision log with findings
7. Update progress tracking table

---

## Success Metrics

We'll know Phase 1 is complete when:
- ✅ All 10 success criteria met
- ✅ Test suite passes 100%
- ✅ Can demo full workflow start to finish
- ✅ Documentation tested by fresh user
- ✅ No critical bugs in backlog
- ✅ Confident to recommend to others

**Target Date**: 2-3 weeks from start (2025-12-03 to 2025-12-10)
**Current Confidence**: 60% (have good foundation, need execution)

---

**This is a living document. Update it as we learn and make decisions.**
