---
id: task-015
title: Fix Broken MAC Address Generation
status: Done
assignee: []
created_date: '2025-11-25'
labels:
  - Bug
  - High
  - Security
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Objective: Use crypto/rand for MAC address generation to ensure uniqueness.

The current `randomByte` function is a placeholder returning `0x00`, resulting in identical MAC addresses for all VMs (`AA:FC:00:00:00:00`). This breaks networking if multiple VMs are on the same bridge.

## Acceptance Criteria

### AC1: MAC Addresses Use Cryptographic Randomness
**Given** the fix is implemented
**When** reviewing the code
**Then** `crypto/rand` is used for MAC generation

**Verification:**
```bash
# Check for crypto/rand import
grep -q 'crypto/rand' internal/firecracker/vm.go
# Expected: exit code 0

# Check randomByte uses crypto/rand
grep -A10 "func randomByte" internal/firecracker/vm.go | grep -qE "rand\.Read|crypto/rand"
# Expected: exit code 0

# Verify no placeholder return
! grep -A5 "func randomByte" internal/firecracker/vm.go | grep -q "return 0x00"
# Expected: exit code 0 (placeholder removed)
```

### AC2: Multiple VMs Have Unique MACs
**Given** multiple VMs are created
**When** inspecting their network configuration
**Then** each has a unique MAC address

**Verification:**
```bash
# Create 3 VMs
for i in 1 2 3; do
  sudo nanofuse vm create mac-test-$i --image base
done

# Get MAC addresses
MAC1=$(sudo nanofuse vm inspect mac-test-1 --format '{{.NetworkConfig.MACAddress}}')
MAC2=$(sudo nanofuse vm inspect mac-test-2 --format '{{.NetworkConfig.MACAddress}}')
MAC3=$(sudo nanofuse vm inspect mac-test-3 --format '{{.NetworkConfig.MACAddress}}')

echo "MAC1: $MAC1"
echo "MAC2: $MAC2"
echo "MAC3: $MAC3"

# All must be different
[ "$MAC1" != "$MAC2" ] && [ "$MAC2" != "$MAC3" ] && [ "$MAC1" != "$MAC3" ]
# Expected: exit code 0

# Cleanup
for i in 1 2 3; do
  sudo nanofuse vm delete mac-test-$i
done
```

### AC3: MAC Addresses Are Valid Format
**Given** a VM is created
**When** inspecting its MAC address
**Then** it follows the expected format (AA:FC:XX:XX:XX:XX)

**Verification:**
```bash
sudo nanofuse vm create mac-format-test --image base
MAC=$(sudo nanofuse vm inspect mac-format-test --format '{{.NetworkConfig.MACAddress}}')

echo "MAC: $MAC"

# Validate format: AA:FC:XX:XX:XX:XX (locally administered, unicast)
echo "$MAC" | grep -qE "^[Aa][Aa]:[Ff][Cc]:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}$"
# Expected: exit code 0

sudo nanofuse vm delete mac-format-test
```

### AC4: MAC Address Not All Zeros
**Given** multiple VMs are created
**When** inspecting their MAC addresses
**Then** none have the broken all-zeros pattern

**Verification:**
```bash
# Create and check 5 VMs
for i in {1..5}; do
  sudo nanofuse vm create mac-zero-test-$i --image base
  MAC=$(sudo nanofuse vm inspect mac-zero-test-$i --format '{{.NetworkConfig.MACAddress}}')

  # Should NOT match the broken pattern
  if echo "$MAC" | grep -qi "AA:FC:00:00:00:00"; then
    echo "FAIL: VM $i has broken MAC: $MAC"
    exit 1
  fi
  echo "VM $i MAC: $MAC (OK)"
done

echo "All MACs are unique and valid"
# Expected: exit code 0

# Cleanup
for i in {1..5}; do
  sudo nanofuse vm delete mac-zero-test-$i
done
```

### AC5: Network Works with Multiple VMs
**Given** multiple VMs with unique MACs
**When** all VMs are started on the same bridge
**Then** all VMs have network connectivity

**Verification:**
```bash
# Create and start 3 VMs
for i in 1 2 3; do
  sudo nanofuse vm create net-test-$i --image base
  sudo nanofuse vm start net-test-$i
done

sleep 15  # Wait for boot

# Verify all have network
PASS=0
for i in 1 2 3; do
  VM_IP=$(sudo nanofuse vm inspect net-test-$i --format '{{.NetworkConfig.IPAddress}}')
  if ping -c 1 -W 2 $VM_IP > /dev/null; then
    ((PASS++))
    echo "VM $i ($VM_IP): OK"
  else
    echo "VM $i ($VM_IP): FAILED"
  fi
done

echo "Result: $PASS/3 VMs have network"
[ $PASS -eq 3 ]
# Expected: exit code 0

# Cleanup
for i in 1 2 3; do
  sudo nanofuse vm stop net-test-$i
  sudo nanofuse vm delete net-test-$i
done
```

## Definition of Done
- [ ] All 5 acceptance criteria pass
- [ ] crypto/rand used for MAC generation
- [ ] Unit test added for MAC generation
- [ ] No networking issues observed with 5+ concurrent VMs

Priority: High
Implementation Location: `internal/firecracker/vm.go`
<!-- SECTION:DESCRIPTION:END -->
