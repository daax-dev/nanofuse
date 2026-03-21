---
id: task-009
title: 'Task 3.1: Add VM Logs Command'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-11-24 01:25'
updated_date: '2025-12-30 18:43'
labels:
  - Task
  - P1
  - Phase2
  - Feature
dependencies:
  - task-005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Epic: EPIC 3 - Usability Improvements

Objective: Make console logs accessible without sudo

## Acceptance Criteria
<!-- AC:BEGIN -->
### AC1: Command Exists
**Given** nanofuse CLI is built
**When** `nanofuse vm logs --help` is executed
**Then** it displays usage information

**Verification:**
```bash
nanofuse vm logs --help 2>&1 | grep -qi "usage\|logs"
# Expected: exit code 0
```

### AC2: Basic Log Display Works
**Given** a running VM with console output
**When** `nanofuse vm logs <name>` is executed
**Then** the console log content is displayed

**Verification:**
```bash
# Setup
sudo nanofuse vm create logs-test --image base
sudo nanofuse vm start logs-test
sleep 5  # Allow some boot output

# Test
nanofuse vm logs logs-test | head -10
# Expected: shows boot messages, exit code 0

# Verify content looks like console output
nanofuse vm logs logs-test | grep -qE "Linux|kernel|systemd|boot"
# Expected: exit code 0
```

### AC3: Follow Mode Works (--follow / -f)
**Given** a running VM
**When** `nanofuse vm logs --follow <name>` is executed
**Then** new log lines are streamed in real-time (like tail -f)

**Verification:**
```bash
# Start following in background
timeout 5 nanofuse vm logs --follow logs-test > /tmp/follow-test.log 2>&1 &
FOLLOW_PID=$!

# Wait a moment for log capture
sleep 3

# Kill the follow process
kill $FOLLOW_PID 2>/dev/null

# Verify we captured some output
[ -s /tmp/follow-test.log ]
# Expected: exit code 0 (file is non-empty)
```

### AC4: Line Limit Works (--lines / -n)
**Given** a VM with extensive console output
**When** `nanofuse vm logs --lines 20 <name>` is executed
**Then** only the last 20 lines are displayed

**Verification:**
```bash
# Count lines returned with --lines 20
LINES=$(nanofuse vm logs --lines 20 logs-test | wc -l)
[ "$LINES" -le 20 ]
# Expected: exit code 0
```

### AC5: Works with Stopped VMs
**Given** a VM that has been stopped
**When** `nanofuse vm logs <name>` is executed
**Then** the console log from the last run is displayed

**Verification:**
```bash
sudo nanofuse vm stop logs-test

# Should still be able to read logs
nanofuse vm logs logs-test | grep -qE "Linux|kernel|systemd"
# Expected: exit code 0
```

### AC6: Handles Missing VM Gracefully
**Given** a VM name that doesn't exist
**When** `nanofuse vm logs <name>` is executed
**Then** a clear error message is displayed

**Verification:**
```bash
nanofuse vm logs nonexistent-vm-xyz123 2>&1 | grep -qiE "not found|does not exist"
# Expected: exit code 0 (error message is clear)

# Command should return non-zero
! nanofuse vm logs nonexistent-vm-xyz123
# Expected: exit code 0 (the ! inverts, so non-zero becomes 0)
```

### AC7: Handles Permission Errors Gracefully
**Given** a user without access to VM logs
**When** `nanofuse vm logs <name>` is executed
**Then** a helpful error message explains how to gain access

**Verification:**
```bash
# This depends on implementation - if logs require API access:
# Verify error message is helpful if permissions are missing
nanofuse vm logs logs-test 2>&1 | grep -qiE "permission\|access\|denied" || \
nanofuse vm logs logs-test > /dev/null
# Expected: either shows permission error OR succeeds (exit code 0)
```

## Definition of Done
- [x] #1 All 7 acceptance criteria pass
- [x] #2 Command documented in `nanofuse vm --help`
- [x] #3 Command matches Docker CLI conventions where applicable
- [x] #4 Unit tests added for logs command

Estimated Effort: 3 hours
Priority: P1
Prerequisites: EPICs 1 & 2 complete
<!-- SECTION:DESCRIPTION:END -->

<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation Complete (PR #67 Merged)

The vm logs command has been fully implemented and merged in PR #67.

### Features Implemented:
- `nanofuse vm logs <vm-id>` - Basic log display
- `nanofuse vm logs --follow/-f <vm-id>` - Real-time streaming
- `nanofuse vm logs --tail/--lines/-n <N> <vm-id>` - Last N lines

### Code Locations:
- CLI Command: `/cmd/nanofuse/main.go` (lines 878-922)
- Stream Function: `/cmd/nanofuse/main.go` (lines 806-876)
- Client API: `/internal/client/client.go` (GetVMLogs method, lines 142-170)
- Server Handler: `/internal/api/vm_handlers.go` (handleVMLogsByPath, lines 766-819)
- API Route: GET /vms/{id}/logs

### Tests:
- Unit tests: TestVMLogsCommandExists, TestVMLogsCommandFlags
- Tests pass: `go test ./cmd/nanofuse/ -v`

### PR URL:
https://github.com/peregrinesummit/nanofuse/pull/67
<!-- SECTION:NOTES:END -->
