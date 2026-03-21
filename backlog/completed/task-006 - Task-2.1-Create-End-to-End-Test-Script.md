---
id: task-006
title: 'Task 2.1: Create End-to-End Test Script'
status: Done
assignee: []
created_date: '2025-11-24 01:25'
labels:
  - Task
  - P0
  - Phase1
  - Testing
dependencies:
  - task-001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Epic: EPIC 2 - E2E Workflow Validation

Objective: Automated test for complete pull-to-running workflow

## Acceptance Criteria

### AC1: Script Exists at Specified Path
**Given** the task is complete
**When** checking for the script
**Then** it exists and is executable

**Verification:**
```bash
test -x test/e2e/full-workflow-test.sh
# Expected: exit code 0
```

### AC2: Script Authenticates to GHCR
**Given** GITHUB_TOKEN is set in environment
**When** the script runs the pull step
**Then** it successfully authenticates to ghcr.io

**Verification:**
```bash
# Script should include authentication step
grep -q "docker login\|crane auth\|GITHUB_TOKEN" test/e2e/full-workflow-test.sh
# Expected: exit code 0

# Or verify in logs during run
./test/e2e/full-workflow-test.sh --verbose 2>&1 | grep -q "Authenticated\|Login Succeeded"
# Expected: exit code 0
```

### AC3: Script Pulls Image Successfully
**Given** authentication succeeded
**When** the script runs the pull step
**Then** the image is available locally

**Verification:**
```bash
./test/e2e/full-workflow-test.sh --verbose 2>&1 | grep -q "Pull.*complete\|Image pulled"
# Expected: exit code 0

# After script runs, image should be listed
sudo nanofuse image list | grep -q "base\|todo"
# Expected: exit code 0
```

### AC4: Script Creates and Starts VM
**Given** the image is available
**When** the script runs the create/start steps
**Then** a VM is running with the test image

**Verification:**
```bash
# During test, VM should be created
./test/e2e/full-workflow-test.sh --verbose 2>&1 | grep -q "VM.*created\|VM.*started"
# Expected: exit code 0
```

### AC5: Script Validates Services Running
**Given** the VM is started
**When** the script runs verification
**Then** it confirms services respond to health checks

**Verification:**
```bash
./test/e2e/full-workflow-test.sh --verbose 2>&1 | grep -E "nginx.*OK|port 80.*OK"
# Expected: exit code 0

./test/e2e/full-workflow-test.sh --verbose 2>&1 | grep -E "backend.*OK|port 8080.*OK|health.*OK"
# Expected: exit code 0
```

### AC6: Script Cleans Up Resources
**Given** verification is complete
**When** the script runs cleanup
**Then** the test VM and pulled image are removed

**Verification:**
```bash
# Before: note existing VMs
BEFORE=$(sudo nanofuse vm list | wc -l)

# Run test
./test/e2e/full-workflow-test.sh

# After: same count (test VM was cleaned up)
AFTER=$(sudo nanofuse vm list | wc -l)
[ "$BEFORE" -eq "$AFTER" ]
# Expected: exit code 0
```

### AC7: Script Is Idempotent
**Given** the script was run once
**When** the script is run again immediately
**Then** it succeeds without errors about existing resources

**Verification:**
```bash
# Run twice in succession
./test/e2e/full-workflow-test.sh && ./test/e2e/full-workflow-test.sh
# Expected: exit code 0 (both runs succeed)
```

### AC8: Script Completes Within Time Limit
**Given** a system with network access to GHCR
**When** the script runs end-to-end
**Then** it completes in under 90 seconds

**Verification:**
```bash
START=$(date +%s)
./test/e2e/full-workflow-test.sh
END=$(date +%s)
ELAPSED=$((END - START))
echo "Elapsed: ${ELAPSED}s"
[ $ELAPSED -lt 90 ]
# Expected: exit code 0, elapsed < 90 seconds
```

### AC9: Script Supports Required Flags
**Given** the script is invoked
**When** using --help, --image, or --verbose flags
**Then** they work as expected

**Verification:**
```bash
# Help flag
./test/e2e/full-workflow-test.sh --help | grep -q "Usage"
# Expected: exit code 0

# Image flag
./test/e2e/full-workflow-test.sh --image ghcr.io/peregrinesummit/nanofuse-base:latest --verbose
# Expected: exit code 0
```

## Definition of Done
- [x] All 9 acceptance criteria pass
- [x] Script has inline documentation
- [x] Error handling provides actionable messages
- [x] Script can run in CI (no interactive prompts)

**Note:** Script validates port 80 only (single-port architecture per task-18).

Output: test/e2e/full-workflow-test.sh
Estimated Effort: 4 hours
Priority: P0
Prerequisites: EPIC 1 complete
<!-- SECTION:DESCRIPTION:END -->
