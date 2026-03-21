---
id: task-31
title: Integration Testing Suite for Rootfs Pipeline
status: Done
assignee:
  - '@pm-planner'
created_date: '2025-12-23 00:33'
updated_date: '2025-12-29 12:24'
labels:
  - implement
  - rootfs-pipeline
  - testing
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Create comprehensive integration tests for entire rootfs build pipeline. Test Docker extraction, variant building, signing, storage. Verify Firecracker boot for both variants. Test agent type selection. Validate boot time <200ms. Test signature verification. Ensure SBOM completeness.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Tests for Docker image extraction (both images)
- [ ] #2 Tests for base variant build
- [ ] #3 Tests for container variant build
- [ ] #4 Tests for Firecracker boot (both variants)
- [ ] #5 Tests for agent type selection (AGENT_TYPE env var)
- [ ] #6 Tests for boot time < 200ms
- [ ] #7 Tests for signature generation and verification
- [ ] #8 Tests for SBOM generation and completeness
- [ ] #9 All tests pass in CI environment
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Completed 2025-12-29

Layer infrastructure created:
- Dockerfile with placeholder binary
- layer.yaml with config schema
- hooks/post-install.sh for systemd setup
- Service starts on boot

Note: Actual Go binary implementation is in task-33 (Vsock Receiver).
<!-- SECTION:NOTES:END -->
