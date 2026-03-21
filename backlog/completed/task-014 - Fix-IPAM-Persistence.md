---
id: task-014
title: Fix IPAM State Loss on Restart
status: Done
assignee: []
created_date: '2025-11-25'
labels:
  - Bug
  - Critical
  - Networking
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Objective: Ensure IP allocations persist across daemon restarts to prevent IP conflicts.

The `IPAM` struct has a `LoadAllocations` method, but it is never called. When the `nanofused` daemon restarts, it initializes a fresh pool and has no knowledge of IPs assigned to currently running VMs. This causes it to re-issue IPs that are already in use.

## Acceptance Criteria

### AC1: Daemon Loads Existing Allocations on Startup
**Given** VMs exist in the database with assigned IPs
**When** the nanofused daemon starts
**Then** those IPs are marked as allocated in the IPAM pool

**Verification:**
```bash
# Setup: Create VM and note its IP
sudo nanofuse vm create ipam-test --image base
sudo nanofuse vm start ipam-test
VM_IP=$(sudo nanofuse vm inspect ipam-test --format '{{.NetworkConfig.IPAddress}}')
echo "VM IP: $VM_IP"

# Restart daemon
sudo systemctl restart nanofused
sleep 3

# Verify IP is still the same after restart
NEW_IP=$(sudo nanofuse vm inspect ipam-test --format '{{.NetworkConfig.IPAddress}}')
[ "$VM_IP" = "$NEW_IP" ]
# Expected: exit code 0 (IP unchanged)
```

### AC2: No IP Conflicts After Restart
**Given** a running VM with an assigned IP
**When** daemon restarts and a new VM is created
**Then** the new VM gets a different IP

**Verification:**
```bash
# Get existing VM's IP
VM1_IP=$(sudo nanofuse vm inspect ipam-test --format '{{.NetworkConfig.IPAddress}}')

# Restart daemon
sudo systemctl restart nanofused
sleep 3

# Create second VM
sudo nanofuse vm create ipam-test-2 --image base
sudo nanofuse vm start ipam-test-2
VM2_IP=$(sudo nanofuse vm inspect ipam-test-2 --format '{{.NetworkConfig.IPAddress}}')

echo "VM1 IP: $VM1_IP"
echo "VM2 IP: $VM2_IP"

# IPs must be different
[ "$VM1_IP" != "$VM2_IP" ]
# Expected: exit code 0 (different IPs)

# Both IPs should be pingable (no conflict)
ping -c 1 -W 2 $VM1_IP && ping -c 1 -W 2 $VM2_IP
# Expected: exit code 0
```

### AC3: IPAM Pool Correctly Reflects Running VMs
**Given** multiple VMs running
**When** querying IPAM state (via logs or API)
**Then** all running VM IPs are shown as allocated

**Verification:**
```bash
# Check daemon logs for IPAM loading
journalctl -u nanofused --since "1 minute ago" | grep -qiE "load.*allocation|ipam.*loaded|restored.*ip"
# Expected: exit code 0 (logs show allocations were loaded)

# Alternative: verify via VM count matching
VM_COUNT=$(sudo nanofuse vm list --format '{{.Name}}' | wc -l)
# IPAM should have at least this many allocations
```

### AC4: Code Changes Implemented
**Given** the fix is implemented
**When** reviewing the code
**Then** `LoadAllocations` is called during server startup

**Verification:**
```bash
# Check that LoadAllocations is called in server.go
grep -rn "LoadAllocations" internal/api/server.go
# Expected: exit code 0, shows the call

# Check it happens during NewServer or init
grep -A5 -B5 "LoadAllocations" internal/api/server.go | grep -qiE "NewServer\|init\|Start"
# Expected: exit code 0
```

### AC5: Unit Test Exists
**Given** the fix is implemented
**When** running tests
**Then** there is a test that verifies IPAM persistence

**Verification:**
```bash
# Test file exists
test -f internal/network/ipam_test.go
# Expected: exit code 0

# Test covers persistence
grep -q "LoadAllocations\|persist\|restart" internal/network/ipam_test.go
# Expected: exit code 0

# Tests pass
go test ./internal/network/... -v 2>&1 | grep -q "PASS"
# Expected: exit code 0
```

## Definition of Done
- [ ] All 5 acceptance criteria pass
- [ ] No IP conflicts observed in testing (10 restart cycles)
- [ ] Code review completed
- [ ] Related tests passing

Priority: Critical
Implementation Location: `internal/api/server.go`
<!-- SECTION:DESCRIPTION:END -->
