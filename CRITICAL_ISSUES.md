# Critical Issues Found - 2025-11-14

## Status: Phase 1 NOT Production Ready

The testing revealed several **critical bugs** that prevent basic functionality from working. The priority should be fixing these issues before ANY work on Phase 2.

---

## Critical Issue #1: Unix Socket Not Created 🔴

**Priority**: CRITICAL
**Impact**: API demo script fails, CLI doesn't work without --api-url flag

**Problem**:
- Config has both `socket` and `tcp_bind` set
- Server code uses if-else logic (line 136-143 in `internal/api/server.go`)
- Only creates TCP listener OR Unix socket, not both
- TCP takes priority, so Unix socket never gets created
- Demo script expects Unix socket at `/tmp/nanofused.sock`

**Root Cause**:
```go
// Line 136 in internal/api/server.go
if cfg.API.TCPBind != "" {
    // Creates TCP listener
} else {
    // Creates Unix socket - NEVER REACHED
}
```

**Fix Applied**:
- Changed priority: Unix socket first, TCP fallback
- But this doesn't support BOTH listeners simultaneously

**Better Fix Needed**:
- Support dual listeners (both TCP AND Unix socket)
- OR clarify in docs/config which mode to use
- OR make CLI auto-detect TCP when socket unavailable

**Files Affected**:
- `internal/api/server.go:136-173`
- `/etc/nanofuse/nanofused.yaml`
- `examples/api-demo.sh`

---

## Critical Issue #2: CLI Image Pull Broken 🔴

**Priority**: CRITICAL
**Impact**: Cannot pull images using CLI command

**Problem**:
```bash
$ nanofuse --api-url http://localhost:8080 image pull --default
Error: Failed to get pull status: Job ID is required
```

**Root Cause**: Unknown - needs investigation
- API endpoint `/images/pull` works fine via curl
- CLI command fails with confusing error about job ID
- Error message appears AFTER "Pulling..." message (reversed order)

**Direct API Test** (WORKS):
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"image_ref":"ghcr.io/jpoley/nanofuse/base:latest"}' \
  http://localhost:8080/images/pull
# Returns: {"job_id":"...","state":"pending",...}
```

**Files to Investigate**:
- `cmd/nanofuse/image.go` (CLI pull command implementation)
- `internal/client/client.go` (client pull implementation)

---

## Critical Issue #3: API Demo Script Doesn't Work 🔴

**Priority**: CRITICAL
**Impact**: Main demo script fails, users cannot test the system

**Problem**:
1. Health check expected `status: "ok"` but API returns `status: "healthy"` ✅ FIXED
2. Unix socket not created (see Issue #1) ✅ WORKAROUND (TCP fallback)
3. No images available to test with
4. Full lifecycle untested

**Current Status**:
- ✅ Health check fixed
- ✅ TCP fallback implemented in demo script
- ❌ Cannot pull images (needs authentication)
- ❌ Full VM lifecycle never tested end-to-end

**Files Affected**:
- `examples/api-demo.sh`
- Needs working image to proceed

---

## Medium Priority Issues

### Issue #4: Database Permissions 🟡

**Problem**:
- Daemon runs as root, owns `/tmp/nanofuse/nanofuse.db`
- Non-root users can't start new daemon (readonly database error)
- Needs proper permission handling

**Error**:
```
failed to run migrations: failed to add architecture column:
attempt to write a readonly database
```

### Issue #5: Documentation Mismatch 🟡

**Problem**:
- README.md claims "Phase 1 Complete ✅" and "Production-ready"
- ROADMAP.md says "Production-ready for basic VM management use cases"
- Reality: Basic functionality doesn't work (can't pull images, CLI broken, demo fails)

**Files**:
- `README.md:307` - "Phase 1 Completed Features ✅"
- `ROADMAP.md:27` - "Production-ready"

---

## Impact Assessment

### What's Actually Working ✅
- Daemon starts and runs
- HTTP API server (TCP mode)
- Health endpoint
- List VMs/images (empty)
- API handlers exist and respond

### What's Broken ❌
1. CLI image pull command
2. Unix socket creation (when TCP also configured)
3. End-to-end VM lifecycle (untested, no images)
4. Demo script (can't complete without images)
5. Image authentication flow

---

## Recommended Action Plan

### Immediate (This Week)

1. **Fix CLI image pull** (2-3 hours)
   - Debug why CLI pull fails
   - Ensure job polling works
   - Test with authenticated registry

2. **Fix Unix socket creation** (2-3 hours)
   - Option A: Support dual listeners
   - Option B: Make CLI detect and use TCP fallback automatically
   - Option C: Document one-mode-only clearly

3. **Test end-to-end workflow** (4-6 hours)
   - Set up authentication properly
   - Pull an image successfully
   - Create, start, stop, delete VM
   - Verify demo script works completely

4. **Update documentation** (1-2 hours)
   - Remove "production-ready" claims
   - Add "Known Issues" section
   - Document current limitations
   - Fix API demo script instructions

### Before ANY Phase 2 Work

- ✅ All critical issues (#1-#3) resolved
- ✅ End-to-end test passing
- ✅ Demo script working
- ✅ CLI commands functional
- ✅ Documentation accurate

---

## Priority Adjustment

**OLD Priority** (from ROADMAP.md):
1. Snapshot/Resume - CRITICAL 🔴
2. S3 Backup - MEDIUM 🟡
3. Security - HIGH 🟠

**NEW Priority** (Reality):
1. **Fix basic functionality** - CRITICAL 🔴 (0-1 week)
2. **Test and verify Phase 1 works** - CRITICAL 🔴 (3-5 days)
3. **Update docs** - HIGH 🟠 (1-2 days)
4. Snapshot/Resume - MEDIUM 🟡 (2-3 weeks)

---

## Success Criteria for "Phase 1 Complete"

Phase 1 should NOT be marked complete until:

- ✅ CLI commands all work (list, pull, create, start, stop, delete)
- ✅ Image pull works with authentication
- ✅ VM lifecycle tested end-to-end (create → start → stop → delete)
- ✅ Demo script completes successfully
- ✅ Both Unix socket AND TCP modes work
- ✅ Network connectivity verified (VM can reach internet)
- ✅ Logs accessible and useful
- ✅ No critical bugs in issue tracker

**Current Status**: 3/8 criteria met (37.5%)

---

## Conclusion

The project has good architecture and code structure, but **basic functionality is broken**. Before adding advanced features like snapshot/resume, we must:

1. Fix the critical bugs blocking basic usage
2. Test the entire system end-to-end
3. Ensure the "happy path" works reliably
4. Update documentation to match reality

**Estimated time to truly complete Phase 1**: 2-3 weeks
