---
id: task-004
title: 'Task 1.3: Create VM Health Check Script'
status: Done
assignee: []
created_date: '2025-11-24 01:25'
labels:
  - Task
  - P0
  - Phase1
  - Automation
dependencies:
  - task-003
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Epic: EPIC 1 - Core Functionality Validation

Objective: Automated validation that VMs are fully functional

**Note:** Architecture changed via task-18. Backend now serves on port 80 only (no nginx, no 8080).
Health check script updated to reflect single-port architecture.

## Acceptance Criteria

### AC1: Script Exists at Specified Path
**Given** the task is complete
**When** checking for the script
**Then** it exists and is executable

**Verification:**
```bash
test -x scripts/health-check.sh
# Expected: exit code 0
```

### AC2: Script Validates VM Boot Status
**Given** a running VM
**When** the health check script is executed
**Then** it verifies the VM state is "running"

**Verification:**
```bash
# Create and start test VM
sudo nanofuse vm create health-test todo-app:latest
sudo nanofuse vm start health-test

# Run health check
./scripts/health-check.sh health-test 2>&1 | grep -q "VM state.*OK\|running"
# Expected: exit code 0
```

### AC3: Script Validates Network Connectivity
**Given** a running VM with network configured
**When** the health check script is executed
**Then** it confirms the VM has an IP and responds to ping

**Verification:**
```bash
./scripts/health-check.sh health-test 2>&1 | grep -q "Network.*OK"
# Expected: exit code 0
```

### AC4: Script Validates Service Endpoints
**Given** a running VM with services started
**When** the health check script is executed
**Then** it confirms HTTP endpoint responds correctly on port 80

**Verification:**
```bash
./scripts/health-check.sh health-test 2>&1 | grep -q "HTTP (port 80): OK"
# Expected: exit code 0
```

### AC5: Script Reports Boot Time
**Given** a VM that just booted
**When** the health check script is executed
**Then** it reports the time from start to services ready

**Verification:**
```bash
./scripts/health-check.sh health-test 2>&1 | grep -E "Boot time: [0-9]+(\.[0-9]+)?s"
# Expected: exit code 0 (boot time is reported)
```

### AC6: Script Completes Within Time Limit
**Given** a healthy VM
**When** the health check script is executed
**Then** it completes in under 10 seconds

**Verification:**
```bash
START=$(date +%s.%N)
./scripts/health-check.sh health-test > /dev/null
END=$(date +%s.%N)
ELAPSED=$(echo "$END - $START" | bc)
echo "Elapsed: ${ELAPSED}s"
[ $(echo "$ELAPSED < 10" | bc) -eq 1 ]
# Expected: exit code 0, elapsed < 10 seconds
```

### AC7: Script Returns Correct Exit Codes
**Given** the script is executed
**When** all checks pass
**Then** exit code is 0
**When** any check fails
**Then** exit code is 1

**Verification:**
```bash
# Test success case
./scripts/health-check.sh health-test
echo "Exit code for healthy VM: $?"
# Expected: 0

# Test failure case (non-existent VM)
./scripts/health-check.sh nonexistent-vm-12345
echo "Exit code for missing VM: $?"
# Expected: 1
```

### AC8: Script Detects Known Failure Modes
**Given** VMs with specific failure conditions
**When** the health check script is executed
**Then** it correctly identifies each failure type

**Verification:**
```bash
# VM not found
./scripts/health-check.sh fake-vm 2>&1 | grep -q "not found\|FAILED"
# Expected: exit code 0
```

## Definition of Done
- [x] All 8 acceptance criteria pass
- [x] Script is well-documented with --help flag
- [x] Script handles edge cases gracefully (no VM name, invalid input)
- [x] Output is machine-parseable (JSON mode via --json flag)

Output: scripts/health-check.sh
Estimated Effort: 2 hours
Priority: P0
Prerequisites: Task 1.2 complete
<!-- SECTION:DESCRIPTION:END -->
