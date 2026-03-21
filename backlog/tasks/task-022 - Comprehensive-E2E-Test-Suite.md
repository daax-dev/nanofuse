---
id: task-022
title: 'Task 5.2: Comprehensive E2E Test Suite'
status: Done
assignee: []
created_date: '2025-11-27'
labels:
  - Testing
  - P0
  - E2E
  - Critical
dependencies:
  - task-020
  - task-021
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Outcome: Comprehensive end-to-end test suite that validates the full nanofuse lifecycle after build and redeploy, including SSH and HTTP connectivity testing.

## Background and Rationale

### E2E Test Scope

The E2E test validates the complete user journey:
1. Stop the daemon (clean slate)
2. Reinstall/rebuild nanofused
3. Reload/register a fresh image
4. Start the daemon
5. Create and start a VM
6. Verify SSH connectivity
7. Verify HTTP connectivity (curl)
8. Clean up resources

### Why E2E Testing?

| Benefit | Rationale |
|---------|-----------|
| Catches integration bugs | Unit tests miss component interactions |
| Validates user workflows | Tests real usage patterns |
| Ensures deploy readiness | Proves the system works end-to-end |
| Documents behavior | Tests serve as executable documentation |

### Test Environment Requirements

**Local Requirements:**
- Linux with KVM support (`/dev/kvm`)
- Root/sudo access for daemon operations
- Network namespace capability
- At least 512MB RAM available for VM
- Firecracker binary installed

**CI Requirements (evaluated below):**
- Self-hosted runners with KVM support, OR
- Nested virtualization on cloud VMs, OR
- Alternative: Mock-based integration tests for CI

## GitHub Actions E2E Feasibility Evaluation

### Executive Summary

**Verdict: Feasible with self-hosted runners; not feasible on GitHub-hosted runners**

### Analysis

#### GitHub-Hosted Runners

| Factor | Status | Notes |
|--------|--------|-------|
| KVM access | NO | GitHub runners don't expose /dev/kvm |
| Nested virtualization | NO | Not available on standard runners |
| Root access | LIMITED | Passwordless sudo available |
| Firecracker support | NO | Requires KVM |

**Source**: [GitHub Actions Virtual Environments](https://github.com/actions/runner-images)

#### Self-Hosted Runner Options

1. **Fireactions** (Recommended)
   - Purpose-built for Firecracker CI
   - Runs GitHub Actions jobs inside Firecracker VMs
   - Source: https://github.com/hostinger/fireactions
   - Provides KVM access within runner

2. **Actuated**
   - ARM64 and x86 runners with KVM support
   - Commercial service with free tier for open source
   - Source: https://actuated.dev/
   - Native Firecracker support

3. **Custom Self-Hosted Runner**
   - Bare metal or nested-virt cloud VMs
   - AWS metal instances (e.g., m5.metal, c5.metal)
   - GCP with nested virtualization enabled
   - Requires maintenance and security hardening

4. **Mock-Based Alternative**
   - Use mock Firecracker interface for CI
   - Test real Firecracker only on self-hosted
   - Provides coverage without KVM
   - Drawback: Doesn't catch real FC issues

### Recommendation

**Hybrid Approach:**
1. **GitHub-hosted runners**: Run unit tests, lint, build, mock-based integration tests
2. **Self-hosted runner** (optional): Run full E2E tests with real Firecracker
3. **Manual E2E**: Document and script for developers to run locally

### CI Strategy Implementation

```yaml
# .github/workflows/ci.yaml additions

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    # ... existing unit tests

  e2e-tests:
    runs-on: self-hosted
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    needs: [unit-tests]
    steps:
      - uses: actions/checkout@v4
      - name: Run E2E tests
        run: mage TestE2E
```

## Acceptance Criteria

### AC1: E2E Test Script Exists and Documents Steps

**Given** the E2E test suite is implemented
**When** examining the test script
**Then** all lifecycle steps are documented and executable

**Verification:**
```bash
# Check E2E test script exists
test -f test/e2e/e2e_test.go || test -f scripts/e2e-test.sh || test -f test/gdt/e2e/lifecycle.yaml
# Expected: exit code 0

# Check script covers all phases
SCRIPT=$(find . -name "e2e*" -type f | head -1)
grep -qiE "stop|daemon" "$SCRIPT" && \
grep -qiE "install|build" "$SCRIPT" && \
grep -qiE "image|register" "$SCRIPT" && \
grep -qiE "start|boot" "$SCRIPT" && \
grep -qiE "ssh" "$SCRIPT" && \
grep -qiE "curl|http" "$SCRIPT"
# Expected: all greps succeed
```

### AC2: Daemon Lifecycle Management Works

**Given** the E2E test is run
**When** managing the daemon lifecycle
**Then** daemon can be stopped, reinstalled, and started

**Verification:**
```bash
# This is the actual E2E test step - requires sudo
sudo systemctl stop nanofused 2>/dev/null || true
sudo systemctl status nanofused 2>&1 | grep -qiE "inactive|dead|not found"
# Expected: daemon is stopped

# Reinstall (build and install)
mage All && sudo cp bin/nanofused /usr/local/bin/
# Expected: exit code 0

# Start daemon
sudo systemctl start nanofused
sudo systemctl is-active nanofused
# Expected: exit code 0
```

### AC3: Image Registration Works

**Given** the daemon is running
**When** registering/loading a new image
**Then** image is available for VM creation

**Verification:**
```bash
# Register local image
sudo bin/register-local-image images/base/rootfs.ext4 base
# Expected: exit code 0

# Verify image is registered
nanofuse image list | grep -q "base"
# Expected: exit code 0
```

### AC4: VM Creation and Boot Works

**Given** an image is registered
**When** creating and starting a VM
**Then** VM boots successfully

**Verification:**
```bash
# Create VM
sudo nanofuse vm create e2e-test --image base
# Expected: exit code 0

# Start VM
sudo nanofuse vm start e2e-test
# Expected: exit code 0

# Wait for boot (with timeout)
for i in {1..30}; do
  if nanofuse vm status e2e-test | grep -q "running"; then
    break
  fi
  sleep 1
done
nanofuse vm status e2e-test | grep -q "running"
# Expected: exit code 0
```

### AC5: SSH Connectivity Works

**Given** a VM is running
**When** attempting SSH connection
**Then** SSH successfully connects

**Verification:**
```bash
# Get VM IP
VM_IP=$(nanofuse vm show e2e-test --format json | jq -r '.ip_address')

# Test SSH (with timeout)
timeout 30 ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no root@${VM_IP} "echo 'SSH OK'"
# Expected: exit code 0, output contains "SSH OK"
```

### AC6: HTTP Connectivity Works

**Given** a VM is running with network access
**When** making HTTP requests from/to VM
**Then** HTTP connectivity is verified

**Verification:**
```bash
# Get VM IP
VM_IP=$(nanofuse vm show e2e-test --format json | jq -r '.ip_address')

# Test HTTP from host to VM (if VM has web server)
curl -sf --connect-timeout 5 "http://${VM_IP}:80/" 2>/dev/null && echo "HTTP to VM OK"
# Expected: may fail if no web server, that's OK

# Test HTTP from VM to internet (validates NAT/routing)
timeout 30 ssh -o StrictHostKeyChecking=no root@${VM_IP} "wget -q -O- http://example.com | head -1"
# Expected: shows HTML content
```

### AC7: Cleanup Works

**Given** E2E test has run
**When** cleaning up resources
**Then** all resources are removed

**Verification:**
```bash
# Stop and delete VM
sudo nanofuse vm stop e2e-test
sudo nanofuse vm delete e2e-test
# Expected: exit code 0

# Verify VM is gone
! nanofuse vm show e2e-test 2>/dev/null
# Expected: exit code 0 (command fails, VM not found)
```

### AC8: Mage Target Exists

**Given** the E2E test suite is implemented
**When** listing mage targets
**Then** E2E test target exists

**Verification:**
```bash
# Check for E2E mage target
mage -l | grep -qiE "e2e|end.to.end|endtoend"
# Expected: exit code 0
```

### AC9: Documentation Exists in docs/tests/

**Given** the E2E test suite is implemented
**When** checking documentation
**Then** comprehensive E2E docs exist

**Verification:**
```bash
# Check E2E test docs exist
test -f docs/tests/e2e-testing.md
# Expected: exit code 0

# Check content covers key areas
grep -qiE "ssh|curl|lifecycle|github actions" docs/tests/e2e-testing.md
# Expected: exit code 0
```

### AC10: CI Strategy Documented

**Given** GitHub Actions feasibility was evaluated
**When** checking documentation
**Then** CI strategy and recommendations are documented

**Verification:**
```bash
# Check CI strategy is documented
grep -qiE "self-hosted|github actions|kvm" docs/tests/e2e-testing.md
# Expected: exit code 0

# Check recommendations exist
grep -qiE "recommend|feasib|option" docs/tests/e2e-testing.md
# Expected: exit code 0
```

## Technical Implementation

### Directory Structure

```
test/
├── e2e/
│   ├── e2e_test.go             # Go test wrapper
│   ├── lifecycle_test.go       # Daemon lifecycle tests
│   ├── vm_test.go              # VM operations tests
│   └── connectivity_test.go    # SSH/HTTP tests
├── gdt/
│   └── e2e/
│       ├── e2e_test.go         # Go test wrapper for gdt
│       └── full_lifecycle.yaml # Complete E2E YAML test
scripts/
└── e2e-test.sh                 # Standalone E2E script
docs/
└── tests/
    └── e2e-testing.md          # E2E testing documentation
```

### Example E2E Test (gdt format)

```yaml
# test/gdt/e2e/full_lifecycle.yaml
name: Full E2E Lifecycle Test
description: Tests complete nanofuse lifecycle from install to SSH

fixtures:
  - clean_environment

tests:
  - name: stop-daemon
    exec:
      command: sudo systemctl stop nanofused || true
      assert:
        exit_code:
          - 0
          - 5  # Already stopped

  - name: build-and-install
    exec:
      command: mage All && sudo cp bin/nanofused /usr/local/bin/
      assert:
        exit_code: 0
    timeout: 300s

  - name: start-daemon
    exec:
      command: sudo systemctl start nanofused
      assert:
        exit_code: 0
    wait:
      duration: 2s

  - name: register-image
    exec:
      command: sudo bin/register-local-image images/base/rootfs.ext4 base
      assert:
        exit_code: 0

  - name: create-vm
    exec:
      command: sudo nanofuse vm create e2e-test --image base
      assert:
        exit_code: 0

  - name: start-vm
    exec:
      command: sudo nanofuse vm start e2e-test
      assert:
        exit_code: 0
    wait:
      duration: 10s

  - name: verify-vm-running
    exec:
      command: nanofuse vm status e2e-test
      assert:
        exit_code: 0
        stdout:
          contains: "running"
    retry:
      count: 10
      interval: 3s

  - name: test-ssh
    exec:
      command: |
        VM_IP=$(nanofuse vm show e2e-test --format json | jq -r '.ip_address')
        ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no root@${VM_IP} "echo 'SSH_SUCCESS'"
      assert:
        exit_code: 0
        stdout:
          contains: "SSH_SUCCESS"
    retry:
      count: 5
      interval: 5s

  - name: test-http-outbound
    exec:
      command: |
        VM_IP=$(nanofuse vm show e2e-test --format json | jq -r '.ip_address')
        ssh -o StrictHostKeyChecking=no root@${VM_IP} "wget -q -O- http://example.com 2>/dev/null | grep -q 'Example Domain'"
      assert:
        exit_code: 0
    retry:
      count: 3
      interval: 5s

  - name: cleanup-vm
    exec:
      command: |
        sudo nanofuse vm stop e2e-test || true
        sudo nanofuse vm delete e2e-test || true
      assert:
        exit_code: 0
    always_run: true
```

### Standalone E2E Script

```bash
#!/bin/bash
# scripts/e2e-test.sh - Run full E2E test suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() { echo -e "${GREEN}[E2E]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }

# Check prerequisites
check_prereqs() {
    log "Checking prerequisites..."
    command -v mage >/dev/null || fail "mage not found"
    command -v jq >/dev/null || fail "jq not found"
    [ -e /dev/kvm ] || fail "KVM not available (/dev/kvm)"
    [ "$(id -u)" -eq 0 ] || warn "Not running as root, some tests may fail"
}

# Phase 1: Stop daemon
phase_stop() {
    log "Phase 1: Stopping daemon..."
    systemctl stop nanofused 2>/dev/null || true
    sleep 2
}

# Phase 2: Build and install
phase_build() {
    log "Phase 2: Building and installing..."
    mage All
    cp bin/nanofused /usr/local/bin/
    cp bin/nanofuse /usr/local/bin/
}

# Phase 3: Register image
phase_image() {
    log "Phase 3: Registering image..."
    systemctl start nanofused
    sleep 2
    bin/register-local-image images/base/rootfs.ext4 base
}

# Phase 4: Create and start VM
phase_vm() {
    log "Phase 4: Creating and starting VM..."
    nanofuse vm create e2e-test --image base
    nanofuse vm start e2e-test

    log "Waiting for VM to boot..."
    for i in {1..30}; do
        if nanofuse vm status e2e-test 2>/dev/null | grep -q "running"; then
            break
        fi
        sleep 1
    done
}

# Phase 5: Test connectivity
phase_connectivity() {
    log "Phase 5: Testing connectivity..."
    VM_IP=$(nanofuse vm show e2e-test --format json | jq -r '.ip_address')

    log "Testing SSH to $VM_IP..."
    ssh -o ConnectTimeout=30 -o StrictHostKeyChecking=no root@${VM_IP} "echo 'SSH OK'" || fail "SSH failed"

    log "Testing HTTP from VM..."
    ssh -o StrictHostKeyChecking=no root@${VM_IP} "wget -q -O- http://example.com | head -1" || warn "HTTP test failed"
}

# Phase 6: Cleanup
phase_cleanup() {
    log "Phase 6: Cleanup..."
    nanofuse vm stop e2e-test 2>/dev/null || true
    nanofuse vm delete e2e-test 2>/dev/null || true
}

# Main
main() {
    check_prereqs
    trap phase_cleanup EXIT

    phase_stop
    phase_build
    phase_image
    phase_vm
    phase_connectivity

    log "E2E test PASSED!"
}

main "$@"
```

### References

- **Firecracker CI Testing**: https://github.com/firecracker-microvm/firecracker/tree/main/tests
- **GitHub Actions Self-Hosted Runners**: https://docs.github.com/en/actions/hosting-your-own-runners
- **Fireactions (Firecracker CI)**: https://github.com/hostinger/fireactions
- **Actuated (ARM/KVM runners)**: https://actuated.dev/
- **gdt-dev/gdt exec plugin**: https://github.com/gdt-dev/gdt/tree/main/plugin/exec

## Definition of Done
- [ ] All 10 acceptance criteria pass
- [ ] E2E test script covers full lifecycle
- [ ] SSH connectivity test implemented
- [ ] HTTP/curl connectivity test implemented
- [ ] Mage target `mage TestE2E` works
- [ ] CI strategy documented (self-hosted vs hosted)
- [ ] Documentation in docs/tests/e2e-testing.md
- [ ] Cleanup phase removes all resources

Priority: P0 (MUST HAVE)
Prerequisites: Task-20, Task-21
Output Files:
- `test/e2e/*.go` or `test/gdt/e2e/*.yaml`
- `scripts/e2e-test.sh`
- `docs/tests/e2e-testing.md`
<!-- SECTION:DESCRIPTION:END -->
