# Phase 1B Fix Plan - VM Config in List Response

**Date**: 2025-11-19
**Bug**: VM list returns empty config (vcpus=0, memory=0)
**Status**: READY FOR IMPLEMENTATION

---

## Root Cause (CONFIRMED)

**TYPE MISMATCH** between server and client:

### Server Side (`internal/types/vm.go`)
```go
type VMListItem struct {
    ID            string
    Name          string
    State         VMState
    Image         string
    CreatedAt     time.Time
    UptimeSeconds *int
    // ← NO CONFIG FIELD!
}
```

### Client Side (`internal/client/types.go`)
```go
type VM struct {
    ID            string
    Name          string
    State         string
    Image         string
    ImageDigest   string
    Architecture  string
    Config        VMConfig   // ← HAS CONFIG!
    Runtime       *VMRuntime
    CreatedAt     time.Time
    ...
}
```

### What Happens
1. Server API handler returns `[]VMListItem` (no config)
2. Client expects `[]VM` (with config)
3. JSON unmarshaling fills `VM.Config` with zero values
4. CLI displays: VCPUS=0, MEMORY=0M

---

## Fix Approach: Add Config to VMListItem

**Decision**: Add `Config` field to `VMListItem` struct

**Rationale**:
- Simple, straightforward fix
- Config data is small (~200 bytes)
- Already being loaded from database
- No breaking changes to client
- VM list is not high-traffic endpoint

**Alternative Rejected**: Lightweight list + separate detail endpoint
- Reason: Adds complexity, requires CLI changes, more API calls
- Not worth it for Phase 1

---

## Implementation Plan

### Step 1: Add Config to VMListItem (5 min)

**File**: `internal/types/vm.go`

**Change**:
```go
type VMListItem struct {
    ID            string     `json:"id"`
    Name          string     `json:"name"`
    State         VMState    `json:"state"`
    Image         string     `json:"image"`
    ImageDigest   string     `json:"image_digest"`      // ADD
    Architecture  string     `json:"architecture"`       // ADD
    Config        VMConfig   `json:"config"`             // ADD
    CreatedAt     time.Time  `json:"created_at"`
    UptimeSeconds *int       `json:"uptime_seconds,omitempty"`
}
```

**Why add ImageDigest and Architecture**:
- Client VM type has these fields
- Minimal extra data
- Useful for display/debugging
- Makes VMListItem closer to full VM

### Step 2: Update List Handler (5 min)

**File**: `internal/api/vm_handlers.go`

**Current** (lines 38-55):
```go
for _, vm := range vms {
    item := types.VMListItem{
        ID:        vm.ID,
        Name:      vm.Name,
        State:     vm.State,
        Image:     vm.Image,
        CreatedAt: vm.CreatedAt,
    }

    if vm.State == types.StateRunning && vm.Runtime != nil {
        uptime := int(time.Since(vm.UpdatedAt).Seconds())
        item.UptimeSeconds = &uptime
    }

    items = append(items, item)
}
```

**New**:
```go
for _, vm := range vms {
    item := types.VMListItem{
        ID:           vm.ID,
        Name:         vm.Name,
        State:        vm.State,
        Image:        vm.Image,
        ImageDigest:  vm.ImageDigest,   // ADD
        Architecture: vm.Architecture,  // ADD
        Config:       vm.Config,        // ADD
        CreatedAt:    vm.CreatedAt,
    }

    if vm.State == types.StateRunning && vm.Runtime != nil {
        uptime := int(time.Since(vm.UpdatedAt).Seconds())
        item.UptimeSeconds = &uptime
    }

    items = append(items, item)
}
```

### Step 3: Test Fix (10 min)

**Test 1: List shows config**
```bash
# Current VM
nanofuse vm list
# Expected: VCPUS=2, MEMORY=512M (not 0)

# Via API
curl http://localhost:8080/vms | jq '.vms[0].config'
# Expected: {"vcpus": 2, "memory_mib": 512, ...}
```

**Test 2: Create new VM**
```bash
nanofuse vm delete test-base -y
nanofuse vm create sha256:b3acbe... test2 --vcpus 4 --memory 2048
nanofuse vm list
# Expected: VCPUS=4, MEMORY=2048M
```

**Test 3: Multiple VMs**
```bash
nanofuse vm create sha256:b3acbe... vm1 --vcpus 1 --memory 256
nanofuse vm create sha256:b3acbe... vm2 --vcpus 2 --memory 512
nanofuse vm create sha256:b3acbe... vm3 --vcpus 4 --memory 1024
nanofuse vm list
# Expected: All three show different correct configs
```

### Step 4: Rebuild and Deploy (5 min)

```bash
# Rebuild daemon
cd /home/jpoley/ps/nanofuse
mage daemon

# Stop daemon
sudo systemctl stop nanofused

# Replace binary
sudo cp build/nanofused /usr/local/bin/nanofused

# Start daemon
sudo systemctl start nanofused

# Verify
nanofuse vm list
```

---

## Files to Modify

1. `internal/types/vm.go` - Add fields to VMListItem struct
2. `internal/api/vm_handlers.go` - Copy fields in list handler

**That's it.** Two files, ~10 lines changed.

---

## Validation Checklist

After fix applied:

- [ ] Build succeeds without errors
- [ ] Daemon starts without errors
- [ ] `nanofuse vm list` shows VCPUS > 0
- [ ] `nanofuse vm list` shows MEMORY > 0
- [ ] `nanofuse vm list --json` has full config
- [ ] Create new VM shows correct config immediately
- [ ] Multiple VMs each show their own config
- [ ] Config persists after daemon restart

---

## Risk Assessment

**Risk**: LOW

**Why**:
- Simple additive change (adding fields)
- No breaking changes
- Client already expects these fields
- Database already has data
- Just connecting existing pieces

**Rollback**:
- Keep old binary as backup
- If issues, revert binary and restart daemon

---

## Estimated Time

- Implementation: 10 minutes
- Testing: 10 minutes
- Deployment: 5 minutes
- **Total: 25 minutes**

---

## Next Steps After This Fix

1. Delete test VM: `nanofuse vm delete test-base`
2. Create fresh VM with config
3. **START** the VM (Phase 1B continued)
4. Verify Firecracker process actually runs
5. Verify network works (ping)
6. Move to Phase 1C (console access)

---

## Approval Required

**User**: Please review this plan and approve before I implement.

**Questions for user**:
1. Is adding ~200 bytes to list response acceptable?
2. Should I proceed with this fix?
3. Any concerns about the approach?

---

**Status**: AWAITING APPROVAL
