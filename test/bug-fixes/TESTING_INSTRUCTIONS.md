# Testing Instructions for Critical Fixes

**Date**: 2025-11-14
**Status**: FIXES APPLIED - READY FOR TESTING

## What Was Fixed

We've applied fixes for 2 critical bugs that were preventing basic functionality:

1. **Issue #1**: CLI image pull command failing with "Job ID is required"
   - **Root Cause**: Type mismatch between API response (`job_id`) and client expectations (`id`)
   - **Fix**: Added `PullImageResponse` type and proper field mapping
   - **Files Changed**: `internal/client/types.go`, `internal/client/client.go`

2. **Issue #2**: Unix socket not created when TCP also configured
   - **Root Cause**: if-else logic only created ONE listener
   - **Fix**: Support both Unix socket AND TCP simultaneously
   - **Files Changed**: `internal/api/server.go`

## Prerequisites for Testing

### 1. Install Updated Binaries

The binaries have been built and are in `/home/jpoley/ps/nanofuse/bin/`:

```bash
# Install updated binaries (requires sudo for system-wide)
sudo cp /home/jpoley/ps/nanofuse/bin/nanofused /usr/local/bin/
sudo cp /home/jpoley/ps/nanofuse/bin/nanofuse /usr/local/bin/

# Verify versions
nanofuse version
nanofused --help
```

### 2. Stop Old Daemon

**CRITICAL**: The old daemon is still running. You must restart it:

```bash
# Find the running daemon
ps aux | grep nanofused

# Stop it (requires sudo - it's running as root)
sudo pkill -9 nanofused

# Verify it's stopped
ps aux | grep nanofused
# Should show nothing
```

### 3. Verify Configuration

Check that your config has both listeners configured:

```bash
cat /etc/nanofuse/nanofused.yaml | grep -A2 "api:"
```

Expected:
```yaml
api:
  socket: /tmp/nanofused.sock
  tcp_bind: "127.0.0.1:8080"
```

### 4. Setup GHCR Authentication

To test image pull, you need to authenticate to GitHub Container Registry:

```bash
# Get a GitHub Personal Access Token with read:packages scope
# Go to: https://github.com/settings/tokens/new?scopes=read:packages

# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# Verify authentication
docker login ghcr.io
# Should show: Login Succeeded
```

## Test Execution Plan

Run these tests **in order**:

### Test 1: Dual Listener Support

This tests that both Unix socket AND TCP listeners are created:

```bash
cd /home/jpoley/ps/nanofuse/test/manual

# Run as root (needs to start/stop daemon)
sudo ./test_listeners.sh
```

**Expected Output**:
```
=== Test Case 2: Dual Listener Support ===

Test 2.1: Verify config has both socket and TCP configured
  Socket path: /tmp/nanofused.sock
  TCP bind: 127.0.0.1:8080

Test 2.2: Starting daemon with dual listener config
  Daemon PID: 12345
✓ Daemon started

Test 2.3: Verify Unix socket created
✓ Unix socket exists at /tmp/nanofused.sock
srwxrwxrwx ... /tmp/nanofused.sock

Test 2.4: Verify TCP listener active
✓ TCP listener on port 8080
nanofused  12345 ...

Test 2.5: Test API via Unix socket
✓ Unix socket API working

Test 2.6: Test API via TCP
✓ TCP API working

Test 2.7: Test CLI auto-detection
✓ CLI works (used Unix socket by default)

Test 2.8: Test CLI with explicit TCP endpoint
✓ CLI works with explicit TCP

=== All Listener Tests PASSED ===
```

**If Test Fails**:
- Check daemon logs: `sudo journalctl -u nanofused -n 50`
- Verify config syntax: `cat /etc/nanofuse/nanofused.yaml`
- Check permissions on `/tmp/` directory

---

### Test 2: CLI Image Pull

This tests the "Job ID is required" fix:

**IMPORTANT**: Start the daemon FIRST (either from Test 1, or manually):

```bash
# If not already running from Test 1:
sudo nanofused &

# Wait for startup
sleep 3

# Check it's running
curl http://localhost:8080/health
# Should return: {"status":"healthy",...}
```

Then run the test:

```bash
cd /home/jpoley/ps/nanofuse/test/manual
./test_image_pull.sh
```

**Expected Output**:
```
=== Test Case 1: CLI Image Pull ===

Test 1.1: Pull default image
Command: nanofuse --api-url http://localhost:8080 image pull --default

Using default image: ghcr.io/jpoley/nanofuse/base:latest
Pulling ghcr.io/jpoley/nanofuse/base:latest...
Job ID: job-abc123-xyz...
Polling job status...
[Progress updates]
✓ Pull command succeeded

Waiting for pull to complete...

Test 1.2: Verify image appeared in list
Images found: 1
✓ Image successfully pulled and listed
  DIGEST                                                                    TAGS       SIZE     PULLED
  sha256:abc123...                                                          latest     1.2 GB   2025-11-14

=== All Image Pull Tests PASSED ===
```

**If Test Fails**:
- **Authentication Error**: Run `docker login ghcr.io` (see Prerequisites #4)
- **"Job ID is required" error**: The fix didn't apply - check you installed updated binaries
- **Connection refused**: Daemon not running - start with `sudo nanofused`

---

## After Tests Pass

Once both tests pass, you can proceed to:

1. **Test 3**: Full VM lifecycle (create → start → stop → delete)
2. **Test 4**: Complete API demo script

These are documented in `PHASE1_CRITICAL_FIXES.md`.

---

## Troubleshooting

### Problem: Old daemon still running

```bash
sudo systemctl stop nanofused  # If using systemd
# OR
sudo pkill -9 nanofused
```

### Problem: Permission denied on socket

```bash
ls -la /tmp/nanofused.sock
# Should show: srwxrwxrwx (0666 permissions)

# If not, check daemon logs for socket creation errors
sudo journalctl -u nanofused -n 50
```

### Problem: Build errors

```bash
cd /home/jpoley/ps/nanofuse
mage clean
mage all

# Check for errors
echo $?
# Should be 0 for success
```

### Problem: Can't find nanofuse command

```bash
which nanofuse
# Should show: /usr/local/bin/nanofuse or /home/jpoley/bin/nanofuse

# Add to PATH if needed
export PATH="/home/jpoley/bin:$PATH"
```

---

## Quick Reference

### Start Daemon (Manual)

```bash
sudo nanofused &

# Or with explicit config
sudo nanofused --config /etc/nanofuse/nanofused.yaml
```

### Check Daemon Status

```bash
# Check process
ps aux | grep nanofused

# Check listeners
sudo lsof -i:8080
ls -la /tmp/nanofused.sock

# Test API
curl http://localhost:8080/health
curl --unix-socket /tmp/nanofused.sock http://localhost/health
```

### CLI Commands

```bash
# List images (auto-detect socket or use TCP)
nanofuse image list
nanofuse --api-url http://localhost:8080 image list

# Pull image
nanofuse --api-url http://localhost:8080 image pull --default

# List VMs
nanofuse --api-url http://localhost:8080 vm list
```

---

## Next Steps

1. **Run Test 1**: Dual listener test → `sudo ./test_listeners.sh`
2. **Run Test 2**: Image pull test → `./test_image_pull.sh`
3. **If both pass**: Continue to VM lifecycle testing (see `PHASE1_CRITICAL_FIXES.md`)
4. **Document results**: Update `CRITICAL_ISSUES.md` with test outcomes
5. **Report issues**: If tests fail, capture error logs and report details

---

## References

- **Detailed Plan**: `PHASE1_CRITICAL_FIXES.md`
- **Issues Document**: `CRITICAL_ISSUES.md`
- **Test Scripts**: `test/manual/test_*.sh`
- **Code Changes**:
  - CLI fix: `internal/client/client.go:228-241`
  - Socket fix: `internal/api/server.go:131-177, 273-304`
