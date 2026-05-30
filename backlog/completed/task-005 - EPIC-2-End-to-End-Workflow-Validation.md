---
id: task-005
title: 'EPIC 2: End-to-End Workflow Validation'
status: Done
assignee: []
created_date: '2025-11-24 01:25'
labels:
  - Epic
  - P0
  - Phase1
dependencies:
  - task-001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Outcome: Complete pull-to-running cycle works reliably

## Acceptance Criteria

### AC1: E2E Test Script Exists and Passes
**Given** EPIC 1 is complete
**When** the E2E test script is executed
**Then** it completes the full workflow: pull → run → verify → cleanup

**Verification:**
```bash
test -x test/e2e/full-workflow-test.sh
# Expected: exit code 0

./test/e2e/full-workflow-test.sh
# Expected: exit code 0, "All tests passed"
```

### AC2: Default Base Image Works
**Given** a clean system with nanofused running
**When** the E2E test runs with the default base image
**Then** the test passes all verification steps

**Verification:**
```bash
./test/e2e/full-workflow-test.sh --image ghcr.io/daax-dev/nanofuse/base:latest
# Expected: exit code 0
```

### AC3: Todo-App Example Works
**Given** a clean system with nanofused running
**When** the E2E test runs with the todo-app example image
**Then** the test passes all verification steps

**Verification:**
```bash
./test/e2e/full-workflow-test.sh --image ghcr.io/daax-dev/nanofuse/todo-app:latest
# Expected: exit code 0
```

### AC4: 10/10 Reliability on Clean System
**Given** a freshly rebooted system with no VMs
**When** the E2E test is run 10 consecutive times
**Then** all 10 runs pass without intervention

**Verification:**
```bash
#!/bin/bash
# Clean system definition: no existing VMs, daemon freshly started
sudo nanofuse vm list | grep -c "running\|stopped" | grep -q "^0$"

PASS=0
for i in {1..10}; do
  echo "Run $i/10..."
  if ./test/e2e/full-workflow-test.sh > /tmp/e2e-run-$i.log 2>&1; then
    ((PASS++))
    echo "  PASSED"
  else
    echo "  FAILED - see /tmp/e2e-run-$i.log"
  fi
done

echo "Result: $PASS/10 runs successful"
[ $PASS -eq 10 ]
# Expected: "10/10 runs successful", exit code 0
```

## Definition of Done
- [x] All 4 acceptance criteria pass
- [x] Test script is committed to test/e2e/
- [x] Troubleshooting documentation created (Task 2.2)
- [x] CI integration ready (script can run headlessly)

Time Box: 2 days
Priority: P0 (MUST HAVE)
Prerequisite: EPIC 1 complete

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6
<!-- SECTION:DESCRIPTION:END -->
