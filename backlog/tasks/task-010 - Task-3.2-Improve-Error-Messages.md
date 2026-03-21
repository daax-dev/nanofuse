---
id: task-010
title: 'Task 3.2: Improve Error Messages'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-11-24 01:25'
updated_date: '2025-12-30 18:38'
labels:
  - Task
  - P1
  - Phase2
  - UX
dependencies:
  - task-005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Epic: EPIC 3 - Usability Improvements

Objective: Actionable error messages with next steps

## Acceptance Criteria
<!-- AC:BEGIN -->
### AC1: All Errors Include Suggested Fix
**Given** any CLI operation that fails
**When** the error is displayed
**Then** it includes at least one actionable suggestion

**Verification:**
```bash
# Test 5 common error scenarios and verify actionable suggestions

# 1. VM not found
nanofuse vm start fake-vm 2>&1 | grep -qiE "try|check|ensure|see|run|list"
# Expected: exit code 0

# 2. Image not found
sudo nanofuse vm create test --image fake-image 2>&1 | grep -qiE "pull|available|list"
# Expected: exit code 0

# 3. Daemon not running (stop daemon first)
sudo systemctl stop nanofused
nanofuse vm list 2>&1 | grep -qiE "start|systemctl|daemon"
sudo systemctl start nanofused
# Expected: exit code 0

# 4. Invalid arguments
nanofuse vm create 2>&1 | grep -qiE "usage|required|example"
# Expected: exit code 0

# 5. Permission denied (if applicable)
# Test depends on system config
```

### AC2: Errors Reference Documentation
**Given** a complex error scenario
**When** the error is displayed
**Then** it references relevant documentation or help commands

**Verification:**
```bash
# Check that at least one error type references docs
nanofuse vm create 2>&1 | grep -qiE "docs/|--help|nanofuse help" || \
nanofuse image pull 2>&1 | grep -qiE "docs/|--help|nanofuse help"
# Expected: exit code 0
```

### AC3: Errors Include Relevant Context
**Given** an operation fails
**When** the error is displayed
**Then** it includes: operation attempted, resource identifier, and failure reason

**Verification:**
```bash
# Error should show what we tried to do and what failed
OUTPUT=$(nanofuse vm start nonexistent-vm-test123 2>&1)

# Should contain the VM name
echo "$OUTPUT" | grep -q "nonexistent-vm-test123"
# Expected: exit code 0

# Should indicate what operation failed
echo "$OUTPUT" | grep -qiE "start|not found|does not exist"
# Expected: exit code 0
```

### AC4: Errors Are Formatted for Readability
**Given** an error occurs
**When** it is displayed
**Then** it uses consistent formatting (no raw stack traces, clear structure)

**Verification:**
```bash
# Errors should not contain Go stack traces in normal mode
nanofuse vm start fake-vm 2>&1 | grep -qv "goroutine\|panic\|runtime"
# Expected: exit code 0 (no stack traces)

# Error should start with "Error:" or similar prefix
nanofuse vm start fake-vm 2>&1 | grep -qiE "^error:|^Error|failed:"
# Expected: exit code 0
```

### AC5: Errors Include Correlation ID (When Applicable)
**Given** an error occurs during an API operation
**When** the error is displayed
**Then** it includes a request ID or correlation ID for debugging

**Verification:**
```bash
# For API errors, check for request/correlation ID
sudo nanofuse vm create test-correlation --image fake 2>&1 | \
  grep -qiE "request.id|correlation|trace" || true
# Note: This may not apply to all errors; verify for API-related ones

# At minimum, timestamp should be available in daemon logs
journalctl -u nanofused --since "1 minute ago" | grep -qE "[0-9]{4}-[0-9]{2}-[0-9]{2}"
# Expected: exit code 0
```

### AC6: Five Common Error Scenarios Improved
**Given** the error improvement work is complete
**When** testing the following scenarios
**Then** each provides better guidance than before

Scenarios to test:
1. `nanofuse vm start <nonexistent>` - suggests checking VM list
2. `nanofuse vm create <name> --image <nonexistent>` - suggests pulling image
3. `nanofuse vm list` (daemon down) - suggests starting daemon
4. `nanofuse image pull <invalid-ref>` - suggests correct format
5. `nanofuse vm create` (missing args) - shows usage example

**Verification:**
```bash
# Run validation script that checks all 5 scenarios
cat > /tmp/test-errors.sh << 'EOF'
#!/bin/bash
PASS=0

# Test 1: VM not found suggests list
if nanofuse vm start fake-vm-xyz 2>&1 | grep -qiE "list|available"; then
  ((PASS++)); echo "Test 1: PASS"
else
  echo "Test 1: FAIL"
fi

# Test 2: Image not found suggests pull
if sudo nanofuse vm create t --image fake 2>&1 | grep -qiE "pull|available"; then
  ((PASS++)); echo "Test 2: PASS"
else
  echo "Test 2: FAIL"
fi

# Test 3: Daemon down suggests start
sudo systemctl stop nanofused
if nanofuse vm list 2>&1 | grep -qiE "start|running|systemctl"; then
  ((PASS++)); echo "Test 3: PASS"
else
  echo "Test 3: FAIL"
fi
sudo systemctl start nanofused

# Test 4: Invalid image ref shows format
if nanofuse image pull ":::invalid" 2>&1 | grep -qiE "format|example|invalid"; then
  ((PASS++)); echo "Test 4: PASS"
else
  echo "Test 4: FAIL"
fi

# Test 5: Missing args shows usage
if nanofuse vm create 2>&1 | grep -qiE "usage|required|name"; then
  ((PASS++)); echo "Test 5: PASS"
else
  echo "Test 5: FAIL"
fi

echo "Result: $PASS/5 tests passed"
[ $PASS -ge 4 ]  # Allow 1 failure for edge cases
EOF
chmod +x /tmp/test-errors.sh
/tmp/test-errors.sh
# Expected: "4/5 tests passed" or better, exit code 0
```

## Definition of Done
- [x] #1 #1 #1 All 6 acceptance criteria pass
- [x] #2 #2 #2 Error message improvements documented in commit message
- [x] #3 #3 #3 No regressions in existing functionality
- [x] #4 #4 #4 Error formatting is consistent across all commands

Estimated Effort: 4 hours
Priority: P1
Prerequisites: EPICs 1 & 2 complete
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Create internal/clierrors package with structured error types
2. Implement error wrapping with suggestions and context
3. Update CLI commands to use new error handling
4. Add tests for error messages
5. Run validation tests from acceptance criteria
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented internal/clierrors package with:
- CLIError struct with Message, Suggestion, Context, DocRef, ExitCode
- Error wrapping for 5 common scenarios (VM not found, image not found, daemon down, invalid image ref, missing args)
- Detection functions for error types
- Consistent formatting without stack traces
- 27 test cases with full coverage

Updated cmd/nanofuse/main.go to use new error handling throughout.

## 2025-12-30: Task Completed

PR #76 merged to main. All acceptance criteria verified and passing.
<!-- SECTION:NOTES:END -->

<!-- AC:END -->

<!-- AC:END -->

<!-- AC:END -->
