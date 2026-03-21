---
id: task-001
title: 'EPIC 1: Core Functionality Validation'
status: Done
assignee: []
created_date: '2025-11-24 01:25'
labels:
  - Epic
  - P0
  - Phase1
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Outcome: Developer can deploy a working VM reliably

## Acceptance Criteria

### AC1: Services Auto-Start on Boot
**Given** a freshly created VM from the base image
**When** the VM reaches the "running" state
**Then** nginx and todo-backend services are active

**Verification:**
```bash
# After VM creation, verify services (requires SSH or console access)
sudo nanofuse vm exec <name> -- systemctl is-active nginx
# Expected: "active"

sudo nanofuse vm exec <name> -- systemctl is-active todo-backend
# Expected: "active"
```

### AC2: Services Accessible from Host
**Given** a running VM with services started
**When** HTTP requests are made to the VM IP
**Then** both services respond correctly

**Verification:**
```bash
# Get VM IP address
VM_IP=$(sudo nanofuse vm inspect <name> --format '{{.NetworkConfig.IPAddress}}')

# Test nginx (port 80)
curl -sf --max-time 5 http://${VM_IP}:80/ | grep -q '<html>'
# Expected: exit code 0

# Test todo-backend (port 8080)
curl -sf --max-time 5 http://${VM_IP}:8080/health | jq -e '.status == "healthy"'
# Expected: exit code 0
```

### AC3: Deployment Reliability
**Given** a clean system with nanofused running
**When** 5 consecutive VMs are created and started
**Then** all 5 VMs pass service health checks

**Verification:**
```bash
# Run reliability test script
./scripts/test-deployment-reliability.sh --count 5
# Expected: "5/5 deployments successful", exit code 0
```

## Definition of Done
- [x] All 3 acceptance criteria pass
- [x] No manual intervention required during testing
- [x] Results documented in test report

**Note:** Architecture simplified via task-18 (nginx replaced with go-chi).
Services now run on single port (80) instead of nginx:80 + backend:8080.

Time Box: 3 days maximum
Priority: P0 (MUST HAVE)

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6
<!-- SECTION:DESCRIPTION:END -->
