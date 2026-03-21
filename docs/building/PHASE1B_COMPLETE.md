# Phase 1B Complete - VM State Management

**Date**: 2025-11-19
**Status**: PARTIALLY COMPLETE - Config bug fixed, but discovered new issues

---

## What We Fixed ✅

### Bug: VM Config Returns Empty (FIXED)

**Problem**: VM list showed VCPUS=0, MEMORY=0M even though database had correct values

**Root Cause**: Type mismatch between server `VMListItem` (no config field) and client `VM` (has config field)

**Fix Applied**:
1. Added `Config`, `ImageDigest`, `Architecture` fields to `VMListItem` type
2. Updated API handler to copy those fields from VM to VMListItem
3. Fixed CLI default socket path (`/tmp/nanofused.sock` → `/run/nanofused.sock`)

**Files Modified**:
- `internal/types/vm.go` (lines 128-138)
- `internal/api/vm_handlers.go` (lines 40-48)
- `cmd/nanofuse/main.go` (line 90)

**Verification**:
```bash
$ nanofuse vm list
  ID        NAME       STATE    VCPUS  MEMORY
  4bf000b4  test-base  created  2      512M     ← CORRECT!
```

**Status**: ✅ FIXED AND VERIFIED

---

## New Issues Discovered ❌

### Issue #1: Image Path Mismatch

**Problem**: Database stores incorrect image paths

**Evidence**:
```sql
sqlite> SELECT digest, rootfs_path FROM images WHERE digest LIKE 'sha256:b3acbe%';
sha256:b3acbe.../tmp/nanofuse/images/.../rootfs.ext4
```

But actual files are in:
```
/var/lib/nanofuse/images/sha256:b3acbe.../
```

**Impact**: Base image (sha256:b3acbe...) has NO actual files, making it unusable

**Next Steps**:
- Investigate where image paths are set during pull/registration
- Fix to use consistent `/var/lib/nanofuse/images/` path
- Re-pull base image with correct paths

### Issue #2: VM Start Fails Silently

**Problem**: When starting VM with missing image files:
- Firecracker process becomes zombie (`<defunct>`)
- VM state shows "running" but nothing is actually running
- No errors returned to user
- Ping fails, TAP device shows NO-CARRIER

**Evidence**:
```bash
$ nanofuse vm start test-base
VM started successfully!  ← LIE
State: running

$ ps aux | grep firecracker
root   4089487  [firecracker] <defunct>  ← ZOMBIE

$ ping 172.16.0.10
100% packet loss  ← NOT RUNNING
```

**Root Cause**: Missing image files cause Firecracker to crash immediately, but daemon doesn't detect this

**Next Steps**:
- **Phase 1C**: Implement console access to see WHY Firecracker crashes
- Add process monitoring to detect crashes
- Update VM state when Firecracker exits unexpectedly
- Return errors when image files don't exist

### Issue #3: Image Delete Doesn't Remove from Database

**Problem**: Removed image still shows in list

```bash
$ nanofuse image remove sha256:b3acbe0b3
Removed image: sha256:b3acbe0b3

$ nanofuse image list
sha256:b3acbe0b3...  ← STILL THERE!
```

**Impact**: LOW - cosmetic issue, doesn't block usage

---

## Phase 1B Summary

### Completed ✅
- [x] Investigated config persistence bug
- [x] Fixed VMListItem type definition
- [x] Updated API handler to include config
- [x] Fixed CLI socket path
- [x] Deployed and tested fixes
- [x] Verified config displays correctly

### Discovered But Not Fixed ❌
- [ ] Image path mismatch (/tmp vs /var/lib)
- [ ] VM start failures not detected
- [ ] No console access to debug crashes
- [ ] Image delete doesn't clean database

### Time Spent
- Investigation: 30 minutes
- Planning: 30 minutes
- Implementation: 15 minutes
- Testing: 15 minutes
- **Total**: 90 minutes

---

## What's Working Now

1. ✅ VM config persists correctly in database
2. ✅ VM list shows correct vcpus/memory
3. ✅ Unix socket works (`/run/nanofused.sock`)
4. ✅ CLI and daemon communicate properly
5. ✅ VM creation stores full config
6. ✅ TAP devices are created with correct names

## What's Still Broken

1. ❌ Base image has no actual files (path mismatch)
2. ❌ VM starts "succeed" but Firecracker crashes
3. ❌ No way to see console output (blind debugging)
4. ❌ No process monitoring / crash detection
5. ❌ Can't run actual workloads yet

---

## Next Steps: Phase 1C

**Goal**: Get console access to see VM boot logs

**Why Critical**: Without console output, we're blind to:
- Kernel boot messages
- Firecracker errors
- Systemd startup issues
- Application failures

**Approach**:
1. Configure Firecracker to capture serial console output
2. Store console logs in `/var/lib/nanofuse/vms/<id>/console.log`
3. Add `nanofuse vm logs <name>` command
4. Use logs to debug why VMs crash

**Estimated Time**: 3-4 hours

---

## Lessons Learned

1. **Type mismatches are subtle** - Server and client had different type structures, causing silent data loss

2. **Always verify end-to-end** - Config persisted but wasn't returned = useless

3. **Error handling is critical** - VM "started successfully" when it actually crashed immediately

4. **Observability first** - Should have implemented console logging before trying to start VMs

5. **Test with real files** - Would have caught path mismatch sooner

---

## Files Changed in Phase 1B

```
internal/types/vm.go           | 3 +++
internal/api/vm_handlers.go    | 3 +++
cmd/nanofuse/main.go           | 2 +-
```

**Total**: 8 lines changed across 3 files

---

## Testing Status

### Passing ✅
- [x] VM config saves to database correctly
- [x] VM list returns config via API
- [x] CLI displays vcpus/memory
- [x] Unix socket connectivity

### Failing ❌
- [ ] VM actually starts and runs
- [ ] VM is reachable by ping
- [ ] Firecracker process stays alive
- [ ] Console logs accessible

---

**Phase 1B Status**: Config bug fixed ✅, but cannot run VMs until Phase 1C complete
