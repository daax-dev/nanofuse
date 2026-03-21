---
id: task-011
title: 'EPIC 4: Snapshot/Resume Validation'
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
Outcome: VMs can be paused and resumed reliably

## Acceptance Criteria

### AC1: Snapshot Command Works
**Given** a running VM
**When** `nanofuse vm pause <name>` is executed
**Then** VM state is saved and VM is paused

**Verification:**
```bash
# Setup
sudo nanofuse vm create snap-test --image base
sudo nanofuse vm start snap-test
sleep 10  # Wait for boot

# Pause
sudo nanofuse vm pause snap-test
# Expected: exit code 0

# Verify state
sudo nanofuse vm inspect snap-test --format '{{.State}}' | grep -qiE "paused|suspended"
# Expected: exit code 0

# Verify snapshot file exists
ls /var/lib/nanofuse/vms/*/snapshot* 2>/dev/null | grep -q snap-test || \
sudo nanofuse vm inspect snap-test --format '{{.SnapshotPath}}' | grep -q "/"
# Expected: exit code 0
```

### AC2: Resume Command Works
**Given** a paused VM with snapshot
**When** `nanofuse vm resume <name>` is executed
**Then** VM resumes from saved state

**Verification:**
```bash
# Resume from paused state
sudo nanofuse vm resume snap-test
# Expected: exit code 0

# Verify running
sudo nanofuse vm inspect snap-test --format '{{.State}}' | grep -qi "running"
# Expected: exit code 0
```

### AC3: Resume Time Under 500ms
**Given** a paused VM
**When** resume is executed
**Then** VM is responsive within 500ms

**Verification:**
```bash
# Measure resume time
START=$(date +%s%3N)
sudo nanofuse vm resume snap-test
END=$(date +%s%3N)

ELAPSED=$((END - START))
echo "Resume time: ${ELAPSED}ms"
[ $ELAPSED -lt 500 ]
# Expected: exit code 0, resume < 500ms
```

### AC4: State Preservation - Processes
**Given** a VM with running processes before pause
**When** resumed
**Then** the same processes are still running

**Verification:**
```bash
# Get process list before pause
sudo nanofuse vm exec snap-test -- ps aux > /tmp/before-pause.txt

# Pause and resume
sudo nanofuse vm pause snap-test
sudo nanofuse vm resume snap-test

# Get process list after resume
sudo nanofuse vm exec snap-test -- ps aux > /tmp/after-resume.txt

# Compare key processes (nginx, systemd, etc.)
diff <(grep nginx /tmp/before-pause.txt) <(grep nginx /tmp/after-resume.txt)
# Expected: exit code 0 (no difference)
```

### AC5: State Preservation - Network
**Given** a VM with active network before pause
**When** resumed
**Then** network connectivity is restored

**Verification:**
```bash
VM_IP=$(sudo nanofuse vm inspect snap-test --format '{{.NetworkConfig.IPAddress}}')

# Pause and resume
sudo nanofuse vm pause snap-test
sudo nanofuse vm resume snap-test

# Test network
ping -c 1 -W 2 $VM_IP
# Expected: exit code 0

# Test service
curl -sf --max-time 5 http://${VM_IP}:80/ > /dev/null
# Expected: exit code 0
```

### AC6: Reliability - 9/10 Success Rate
**Given** a healthy VM
**When** pause/resume cycle is performed 10 times
**Then** at least 9 succeed

**Verification:**
```bash
PASS=0
for i in {1..10}; do
  echo "Cycle $i/10..."
  sudo nanofuse vm pause snap-test
  sudo nanofuse vm resume snap-test

  VM_IP=$(sudo nanofuse vm inspect snap-test --format '{{.NetworkConfig.IPAddress}}')
  if curl -sf --max-time 5 http://${VM_IP}:80/ > /dev/null; then
    ((PASS++))
    echo "  PASS"
  else
    echo "  FAIL"
  fi
  sleep 1
done

echo "Result: $PASS/10 cycles successful"
[ $PASS -ge 9 ]
# Expected: "9/10 cycles successful" or better, exit code 0
```

## Definition of Done
- [ ] All 6 acceptance criteria pass
- [ ] Research document completed (Task 4.1)
- [ ] Implementation tested on multiple VM configurations
- [ ] Edge cases documented (what happens with disk I/O, network sockets, etc.)

Time Box: 5 days (includes research)
Priority: P1 (SHOULD HAVE)
Prerequisite: EPICs 1 & 2 complete

IMPORTANT: Must complete Task 4.1 (research) BEFORE implementation

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6
<!-- SECTION:DESCRIPTION:END -->
