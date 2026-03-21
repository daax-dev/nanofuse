---
id: task-003
title: 'Task 1.2: Fix Service Startup'
status: Done
assignee: []
created_date: '2025-11-24 01:25'
labels:
  - Task
  - P0
  - Phase1
  - Implementation
dependencies:
  - task-002
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Epic: EPIC 1 - Core Functionality Validation

Objective: Implement fix to get todo-backend running (unified chi server on port 80)

**Note:** Architecture changed via task-18 (Replace nginx with chi). The backend now serves
both static files and API on port 80 using go-chi. No separate nginx service.

## Acceptance Criteria

### AC1: Fix Implemented
**Given** the root cause identified in Task 1.1
**When** the fix is applied to the codebase or image configuration
**Then** the changes are committed with a descriptive message

**Verification:**
```bash
# Check git log for fix commit (task-18 replaced nginx with chi)
git log --oneline -20 | grep -i "replace.*nginx\|chi\|static"
# Expected: at least one matching commit

# Verify Dockerfile has no nginx
! grep -q "nginx" examples/todo-app/docker/Dockerfile
# Expected: exit code 0
```

### AC2: VM Rebuilt with Fix
**Given** the fix is committed
**When** a new todo-app image is built
**Then** the image builds successfully and is tagged

**Verification:**
```bash
# Build new image (via dev-rebuild.sh or manually)
./scripts/dev-rebuild.sh
# Expected: exit code 0

# Verify image exists
sudo nanofuse image list | grep -q "todo"
# Expected: exit code 0
```

### AC3: Backend Service Starts and Serves Static Files
**Given** a VM created from the fixed image
**When** the VM reaches "running" state
**Then** go-chi backend serves static files on port 80

**Verification:**
```bash
# Create test VM
sudo nanofuse vm create test-backend todo-app:latest
sudo nanofuse vm start test-backend

# Wait for boot (max 30 seconds)
for i in {1..30}; do
  VM_IP=$(sudo nanofuse vm inspect test-backend --json | jq -r '.config.network.ip_address')
  [ -n "$VM_IP" ] && [ "$VM_IP" != "null" ] && curl -sf --max-time 2 http://${VM_IP}:80/ > /dev/null && break
  sleep 1
done

# Verify static files served
curl -sf --max-time 5 http://${VM_IP}:80/ | grep -q '<html>'
# Expected: exit code 0
```

### AC4: Health Endpoint Responds
**Given** a VM created from the fixed image
**When** the VM reaches "running" state
**Then** health endpoint responds on port 80

**Verification:**
```bash
# Using same VM from AC3
curl -sf --max-time 5 http://${VM_IP}:80/health
# Expected: {"status":"ok"} or similar JSON, exit code 0
```

### AC5: Deployment Reliability (5/5)
**Given** a clean system with the fixed image
**When** 5 VMs are created sequentially
**Then** all 5 pass service health checks on first attempt

**Verification:**
```bash
#!/bin/bash
PASS=0
for i in {1..5}; do
  VM="reliability-test-$i"
  sudo nanofuse vm create $VM todo-app:latest
  sudo nanofuse vm start $VM
  sleep 10  # Wait for boot

  VM_IP=$(sudo nanofuse vm inspect $VM --json | jq -r '.config.network.ip_address')
  if curl -sf --max-time 5 http://${VM_IP}:80/ > /dev/null && \
     curl -sf --max-time 5 http://${VM_IP}:80/health > /dev/null; then
    ((PASS++))
  fi

  sudo nanofuse vm stop $VM
  sudo nanofuse vm delete $VM
done

echo "Result: $PASS/5 deployments successful"
[ $PASS -eq 5 ]
# Expected: "5/5 deployments successful", exit code 0
```

## Definition of Done
- [x] All 5 acceptance criteria pass
- [x] Fix committed with descriptive message (task-18 merged)
- [x] No regressions in existing functionality
- [x] Architecture simplified (nginx removed, single-port chi server)

Estimated Effort: 2-4 hours
Priority: P0
Prerequisites: Task 1.1 complete
<!-- SECTION:DESCRIPTION:END -->
