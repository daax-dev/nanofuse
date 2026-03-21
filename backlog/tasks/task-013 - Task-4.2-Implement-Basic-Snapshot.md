---
id: task-013
title: 'Task 4.2: Implement Basic Snapshot'
status: To Do
assignee: []
created_date: '2025-11-24 01:25'
labels:
  - Task
  - P1
  - Phase2
  - Feature
dependencies:
  - task-012
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Epic: EPIC 4 - Snapshot/Resume Validation

Objective: Minimal viable snapshot/resume functionality

## Acceptance Criteria

### AC1: Pause Command Implemented
**Given** a running VM
**When** `nanofuse vm pause <name>` is executed
**Then** the VM is paused and snapshot is created

**Verification:**
```bash
# Help shows pause command
nanofuse vm pause --help 2>&1 | grep -qi "pause\|snapshot"
# Expected: exit code 0

# Create and start test VM
sudo nanofuse vm create pause-test --image base
sudo nanofuse vm start pause-test
sleep 10

# Execute pause
sudo nanofuse vm pause pause-test
# Expected: exit code 0

# VM state shows paused
sudo nanofuse vm inspect pause-test --format '{{.State}}' | grep -qiE "paused|suspended"
# Expected: exit code 0
```

### AC2: Resume Command Implemented
**Given** a paused VM
**When** `nanofuse vm resume <name>` is executed
**Then** the VM resumes from snapshot

**Verification:**
```bash
# Help shows resume command
nanofuse vm resume --help 2>&1 | grep -qi "resume\|restore"
# Expected: exit code 0

# Execute resume
sudo nanofuse vm resume pause-test
# Expected: exit code 0

# VM state shows running
sudo nanofuse vm inspect pause-test --format '{{.State}}' | grep -qi "running"
# Expected: exit code 0
```

### AC3: Process State Preserved
**Given** a VM with specific processes running
**When** paused and resumed
**Then** the same processes are running after resume

**Verification:**
```bash
# Check nginx is running before
sudo nanofuse vm exec pause-test -- systemctl is-active nginx | grep -q "active"
# Expected: exit code 0

# Pause and resume
sudo nanofuse vm pause pause-test
sudo nanofuse vm resume pause-test

# Check nginx still running after
sudo nanofuse vm exec pause-test -- systemctl is-active nginx | grep -q "active"
# Expected: exit code 0
```

### AC4: Network State Preserved
**Given** a VM with network connectivity
**When** paused and resumed
**Then** network connectivity is restored

**Verification:**
```bash
VM_IP=$(sudo nanofuse vm inspect pause-test --format '{{.NetworkConfig.IPAddress}}')

# Verify network before
curl -sf --max-time 5 http://${VM_IP}:80/ > /dev/null
# Expected: exit code 0

# Pause and resume
sudo nanofuse vm pause pause-test
sleep 1
sudo nanofuse vm resume pause-test
sleep 2

# Verify network after
curl -sf --max-time 5 http://${VM_IP}:80/ > /dev/null
# Expected: exit code 0
```

### AC5: File System State Preserved
**Given** a VM with files created before pause
**When** paused and resumed
**Then** files still exist after resume

**Verification:**
```bash
# Create a test file
sudo nanofuse vm exec pause-test -- sh -c 'echo "test-data-12345" > /tmp/snapshot-test.txt'

# Verify file exists
sudo nanofuse vm exec pause-test -- cat /tmp/snapshot-test.txt | grep -q "test-data-12345"
# Expected: exit code 0

# Pause and resume
sudo nanofuse vm pause pause-test
sudo nanofuse vm resume pause-test

# Verify file still exists with same content
sudo nanofuse vm exec pause-test -- cat /tmp/snapshot-test.txt | grep -q "test-data-12345"
# Expected: exit code 0
```

### AC6: Resume Time Under 500ms
**Given** a paused VM
**When** resume is timed
**Then** the command completes in under 500ms

**Verification:**
```bash
# Ensure VM is paused
sudo nanofuse vm pause pause-test

# Time the resume
START=$(date +%s%3N)
sudo nanofuse vm resume pause-test
END=$(date +%s%3N)

ELAPSED=$((END - START))
echo "Resume completed in ${ELAPSED}ms"
[ $ELAPSED -lt 500 ]
# Expected: exit code 0
```

### AC7: Reliability 9/10 Success
**Given** a healthy VM
**When** pause/resume is performed 10 times
**Then** at least 9 complete successfully

**Verification:**
```bash
PASS=0
for i in {1..10}; do
  echo "=== Cycle $i/10 ==="

  # Pause
  if ! sudo nanofuse vm pause pause-test; then
    echo "  Pause FAILED"
    continue
  fi

  # Resume
  if ! sudo nanofuse vm resume pause-test; then
    echo "  Resume FAILED"
    continue
  fi

  # Verify services
  sleep 2
  VM_IP=$(sudo nanofuse vm inspect pause-test --format '{{.NetworkConfig.IPAddress}}')
  if curl -sf --max-time 5 http://${VM_IP}:80/ > /dev/null; then
    ((PASS++))
    echo "  PASSED"
  else
    echo "  Service check FAILED"
  fi
done

echo ""
echo "Result: $PASS/10 successful"
[ $PASS -ge 9 ]
# Expected: exit code 0
```

### AC8: Error Handling for Invalid States
**Given** a VM in various states
**When** invalid pause/resume operations are attempted
**Then** clear error messages are displayed

**Verification:**
```bash
# Cannot pause a stopped VM
sudo nanofuse vm stop pause-test
sudo nanofuse vm pause pause-test 2>&1 | grep -qiE "not running|cannot pause|invalid state"
# Expected: exit code 0

# Cannot resume a running VM
sudo nanofuse vm start pause-test
sleep 5
sudo nanofuse vm resume pause-test 2>&1 | grep -qiE "not paused|already running|invalid state"
# Expected: exit code 0
```

## Definition of Done
- [ ] All 8 acceptance criteria pass
- [ ] Pause/resume commands documented in CLI help
- [ ] Unit tests added for snapshot logic
- [ ] Edge cases documented (disk I/O during snapshot, long-running processes, etc.)

Estimated Effort: 3 days
Priority: P1
Prerequisites: Task 4.1 complete with GO decision
<!-- SECTION:DESCRIPTION:END -->
