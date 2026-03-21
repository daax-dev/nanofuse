# Phase 1B Bug Report - VM Config Returns Empty

**Date**: 2025-11-19
**Status**: ROOT CAUSE IDENTIFIED - Not Fixed
**Severity**: CRITICAL - Blocks all VM operations

---

## Bug Summary

VM config is stored correctly in database but API returns empty/zero values.

**Symptom**:
```bash
$ nanofuse vm list
  ID        NAME       STATE    IMAGE           VCPUS  MEMORY  UPTIME
  4bf000b4  test-base  created  sha256:b3acbe... 0      0M      -
```

**Expected**: Should show VCPUS=2, MEMORY=512M

---

## Evidence

### 1. Database Has Correct Data ✅

```bash
$ sqlite3 /var/lib/nanofuse/nanofuse.db \
  "SELECT id, name, config_json FROM vms WHERE name='test-base'"

4bf000b4-fa0e-4df0-9012-1dc6b8aaf7e2|test-base|{"vcpus":2,"memory_mib":512,...}
```

**Full config_json**:
```json
{
  "vcpus": 2,
  "memory_mib": 512,
  "kernel_args": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k ip=172.16.0.10::172.16.0.1:255.255.255.0::eth0:off net.ifnames=0 biosdevname=0",
  "network": {
    "mode": "nat",
    "tap_device": "tap-4bf000b4",
    "mac_address": "AA:FC:00:F1:CD:05",
    "ip_address": "172.16.0.10",
    "gateway": "172.16.0.1",
    "netmask": "255.255.255.0"
  },
  "disks": [
    {
      "drive_id": "rootfs",
      "path_on_host": "/tmp/nanofuse/images/sha256:b3acbe0b35131cc840aabecfa33890dddad86fe76afd5d6e01c224a4e7b2ccbc/rootfs.ext4",
      "is_read_only": false,
      "is_root_device": true
    }
  ]
}
```

### 2. API Returns Empty Config ❌

```bash
$ nanofuse vm list --json | jq '.vms[0].config'

{
  "vcpus": 0,
  "memory_mib": 0,
  "kernel_args": "",
  "network": {
    "mode": ""
  }
}
```

---

## Root Cause Analysis

**The bug is in the API response layer, NOT the database.**

### Data Flow

1. **VM Create** (Working ✅)
   - CLI sends request with vcpus=2, memory=512
   - API handler `handleCreateVM` builds config correctly (vm_handlers.go:273-274)
   - Config saved to database correctly (storage/db.go:84-107)
   - Database confirmed has correct JSON

2. **VM List** (Broken ❌)
   - API handler `handleListVMs` queries database (api/vm_handlers.go:28-62)
   - Database returns VM with correct config (storage/db.go:158-217)
   - **SOMEWHERE** between database and API response, config becomes empty
   - CLI receives empty config

### Code Path for VM List

**File**: `internal/api/vm_handlers.go`

```go
// Line 28-62: handleListVMs
func (s *Server) handleListVMs(w http.ResponseWriter, r *http.Request) {
    stateFilter := r.URL.Query().Get("state")

    // Get VMs from database - returns []*types.VM
    vms, err := s.db.ListVMs(stateFilter)
    if err != nil {
        // error handling
    }

    // Transform to VMListItem (LINE 38-55)
    items := make([]types.VMListItem, 0, len(vms))
    for _, vm := range vms {
        item := types.VMListItem{
            ID:        vm.ID,
            Name:      vm.Name,
            State:     vm.State,
            Image:     vm.Image,
            CreatedAt: vm.CreatedAt,
        }

        // Calculate uptime if running
        if vm.State == types.StateRunning && vm.Runtime != nil {
            uptime := int(time.Since(vm.UpdatedAt).Seconds())
            item.UptimeSeconds = &uptime
        }

        items = append(items, item)  // ← BUG: Config not copied!
    }

    response := types.ListVMsResponse{
        VMs:   items,
        Total: len(items),
    }

    writeJSON(w, http.StatusOK, response)
}
```

**PROBLEM IDENTIFIED**: Lines 40-55 create `VMListItem` but **never copy vm.Config**!

### Type Definitions

Need to check `internal/types/vm.go` to see VMListItem structure:

```bash
$ grep -A 15 "type VMListItem" internal/types/vm.go
```

**Hypothesis**: `VMListItem` either:
1. Doesn't have a `Config` field at all
2. Has a `Config` field but it's not being populated

---

## The Actual Bug

**Location**: `internal/api/vm_handlers.go` lines 40-55

**What happens**:
1. Database returns full `types.VM` with populated `Config`
2. Handler transforms to `types.VMListItem`
3. **Transformation only copies ID, Name, State, Image, CreatedAt**
4. **Config is NEVER copied to VMListItem**
5. CLI receives VMListItem with empty/zero config

**Why it shows 0/empty**:
- Go zero values: `int` → 0, `string` → ""
- CLI displays zero values as "0" and "0M"

---

## Investigation Plan

### Step 1: Check VMListItem Type Definition
**File**: `internal/types/vm.go`

**Questions**:
- Does `VMListItem` have a `Config` field?
- If yes, what type is it?
- If no, why is CLI expecting config in list response?

### Step 2: Check Client Type Definition
**File**: `internal/client/types.go`

**Questions**:
- What does client expect from `/vms` endpoint?
- Does client's VM list type include config?
- Is there a mismatch between API response and client expectation?

### Step 3: Check CLI Display Logic
**File**: `cmd/nanofuse/main.go` or `internal/output/*.go`

**Questions**:
- How does CLI extract VCPUS and MEMORY for display?
- Is CLI expecting config in list response?
- Does CLI have default/fallback values?

### Step 4: Determine Correct Fix

**Option A**: Add Config to VMListItem
- Modify `types.VMListItem` to include `Config types.VMConfig`
- Update `handleListVMs` to copy config
- Pros: Simple, complete data
- Cons: List response becomes larger

**Option B**: Create separate detailed endpoint
- Keep list lightweight (no config)
- Use `GET /vms/{id}` for full details
- Update CLI to call detail endpoint
- Pros: Efficient list, full details when needed
- Cons: More API calls for CLI

**Option C**: Add query parameter for config
- `GET /vms?include=config` returns config
- Default `/vms` lightweight
- Pros: Flexible
- Cons: More complex

---

## Test Plan (After Fix)

### Test 1: Create VM with Config
```bash
# Clean slate
nanofuse vm list
# Should show: empty

# Create VM with specific config
nanofuse vm create sha256:b3acbe0b3... test-vm --vcpus 4 --memory 1024

# Verify list shows correct config
nanofuse vm list
# Expected: VCPUS=4, MEMORY=1024M

# Verify via API
curl http://localhost:8080/vms | jq '.vms[0].config'
# Expected: {"vcpus": 4, "memory_mib": 1024, ...}
```

### Test 2: Multiple VMs
```bash
# Create VMs with different configs
nanofuse vm create <image> vm1 --vcpus 2 --memory 512
nanofuse vm create <image> vm2 --vcpus 4 --memory 2048

# List all
nanofuse vm list
# Expected: Both show correct different configs

# Verify persistence after daemon restart
sudo systemctl restart nanofused
sleep 2
nanofuse vm list
# Expected: Same configs still correct
```

### Test 3: Start VM and Verify Config Used
```bash
# Create VM
nanofuse vm create <image> test --vcpus 2 --memory 512

# Start VM
nanofuse vm start test

# Check Firecracker process args
ps aux | grep firecracker
# Expected: Should show memory and vcpu flags

# Check if VM actually has resources
# (requires console access - Phase 1C)
```

---

## Files to Investigate

1. `internal/types/vm.go` - Type definitions
2. `internal/api/vm_handlers.go:38-55` - List handler
3. `internal/client/types.go` - Client expectations
4. `internal/output/*.go` - CLI display logic
5. `cmd/nanofuse/main.go` - VM list command

---

## Current Status

- ✅ Bug reproduced consistently
- ✅ Database stores config correctly
- ✅ API transformation loses config
- ❌ Root cause location identified but not confirmed
- ❌ Fix approach not determined
- ❌ No fix applied

---

## Next Steps

1. **Investigate types** (15 min)
   - Read `internal/types/vm.go`
   - Read `internal/client/types.go`
   - Document type structures

2. **Analyze mismatch** (15 min)
   - Compare server types vs client types
   - Identify where data is lost
   - Confirm root cause

3. **Design fix** (30 min)
   - Choose approach (A, B, or C above)
   - Document reasoning
   - Get user approval

4. **Implement fix** (30 min)
   - Modify server types/handlers
   - Update client if needed
   - Test thoroughly

5. **Validate** (15 min)
   - Run test plan above
   - Verify config persists
   - Verify CLI displays correctly

**Total estimated time**: 2 hours

---

## Blocking Issues

None - can proceed with investigation and fix.

---

## Related Issues

This may explain why old VMs showed as "running" with 0 vcpus/memory:
- They may have been created before this bug existed
- Or bug existed and they were always showing 0
- Or state corruption is a separate bug

Will know more after fixing this bug and re-testing VM lifecycle.

---

**Status**: Documented, ready for investigation phase.
