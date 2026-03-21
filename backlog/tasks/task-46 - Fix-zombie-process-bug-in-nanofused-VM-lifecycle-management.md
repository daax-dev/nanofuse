---
id: task-46
title: Fix zombie process bug in nanofused VM lifecycle management
status: Done
assignee: []
created_date: '2025-12-29 01:49'
updated_date: '2025-12-29 12:52'
labels:
  - bug
  - nanofused
  - process-management
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Firecracker child processes become zombies because nanofused doesn't properly wait() on them.

**Symptoms:**
- `nanofuse vm stop` fails with "did not stop even after SIGKILL"
- `ps aux` shows `[firecracker] <defunct>` zombie processes
- Zombies accumulate until daemon restart

**Root Cause:**
nanofused starts firecracker as a child process but doesn't have a background goroutine calling wait() to reap exited children.

**Fix Required:**
Add a goroutine per VM that waits on the firecracker process and updates VM state when it exits. Handle both graceful shutdown and unexpected exits.

**Discovered:** 2025-12-28 debug session (DEC-016)
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 VM processes are properly reaped when they exit
- [ ] #2 No zombie processes accumulate
- [ ] #3 VM state is updated to stopped/failed on process exit
- [ ] #4 Network resources are cleaned up on unexpected exit
- [ ] #5 Unit tests verify handler mechanism
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Fixed 2025-12-29

### Changes Made

1. **internal/firecracker/vm.go**
   - Added `ProcessExitHandler` callback type
   - Added `SetProcessExitHandler()` method
   - Modified `Start()` to spawn goroutine calling `waitForProcessExit()`
   - `waitForProcessExit()` calls `cmd.Wait()` (the key fix) and invokes callback

2. **internal/api/vm_handlers.go**
   - Added `handleVMProcessExit()` - updates VM state on exit
   - Added `cleanupVMNetwork()` - releases IP, removes port forwards, deletes TAP

3. **internal/api/server.go**
   - Wire up `fcManager.SetProcessExitHandler(server.handleVMProcessExit)`

4. **internal/firecracker/vm_test.go**
   - Added 4 unit tests for ProcessExitHandler mechanism

### How It Works

When `fcManager.Start()` is called:
1. Firecracker process starts via `cmd.Start()`
2. Goroutine spawned immediately calls `cmd.Wait()` (blocks until exit)
3. When process exits, `cmd.Wait()` returns and reaps the zombie
4. Exit handler is called with vmID, exit code, and error
5. Handler updates VM state in database and cleans up resources
<!-- SECTION:NOTES:END -->
