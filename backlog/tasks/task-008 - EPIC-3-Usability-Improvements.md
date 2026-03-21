---
id: task-008
title: 'EPIC 3: Usability Improvements'
status: To Do
assignee: []
created_date: '2025-11-24 01:25'
labels:
  - Epic
  - P1
  - Phase2
dependencies:
  - task-005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Outcome: Error messages are actionable, logs are accessible

## Acceptance Criteria

### AC1: VM Logs Command Implemented
**Given** a running or stopped VM
**When** `nanofuse vm logs <name>` is executed
**Then** console logs are displayed without requiring sudo

**Verification:**
```bash
# Create test VM
sudo nanofuse vm create log-test --image base
sudo nanofuse vm start log-test

# Test logs command (without sudo if possible, or verify it works with sudo)
nanofuse vm logs log-test 2>&1 | head -20
# Expected: shows console output, exit code 0

# Cleanup
sudo nanofuse vm stop log-test
sudo nanofuse vm delete log-test
```

### AC2: Error Messages Include Next Steps
**Given** an operation fails
**When** the error is displayed
**Then** it includes a suggested action or reference

**Verification:**
```bash
# Trigger a known error (VM not found)
nanofuse vm start nonexistent-vm-12345 2>&1 | grep -qiE "try|check|ensure|see|run"
# Expected: exit code 0 (error message contains actionable guidance)

# Trigger another error (image not found)
sudo nanofuse vm create test --image nonexistent-image 2>&1 | grep -qiE "pull|available|list"
# Expected: exit code 0
```

### AC3: Error Messages Include Context
**Given** an operation fails
**When** the error is displayed
**Then** it includes relevant identifiers (VM name, image name, etc.)

**Verification:**
```bash
# Error should include the VM name we tried to use
nanofuse vm start my-test-vm 2>&1 | grep -q "my-test-vm"
# Expected: exit code 0 (VM name appears in error)
```

### AC4: New User Can Debug Without External Help
**Given** a new user encounters a failure
**When** they read the error message and run suggested commands
**Then** they can identify the root cause within 5 minutes

**Verification:**
This is validated through user testing. Acceptance test:
```bash
# Simulate common new-user error: daemon not running
sudo systemctl stop nanofused
nanofuse vm list 2>&1 | tee /tmp/error-output.txt

# Error should clearly indicate daemon is not running
grep -qiE "daemon.*not running|cannot connect|connection refused" /tmp/error-output.txt
# Expected: exit code 0

# Error should suggest how to start it
grep -qiE "systemctl start\|nanofused" /tmp/error-output.txt
# Expected: exit code 0

sudo systemctl start nanofused
```

## Definition of Done
- [ ] All 4 acceptance criteria pass
- [ ] `nanofuse vm logs` command works
- [ ] At least 5 common error scenarios tested with improved messages
- [ ] No regressions in existing CLI functionality

Time Box: 2 days
Priority: P1 (SHOULD HAVE)
Prerequisite: EPICs 1 & 2 complete

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6
<!-- SECTION:DESCRIPTION:END -->
