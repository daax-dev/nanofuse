---
id: task-30
title: 'T007: Phase 1 Integration Tests'
status: In Progress
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:15'
updated_date: '2026-01-08 02:40'
labels:
  - phase-1
  - testing
  - integration
  - e2e
  - flowspec-microvm
  - validate
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
End-to-end integration tests for core layer build system.

**Context**: Part of Phase 1 - validates all Phase 1 components work together.
**Dependency**: T006 (CLI)

**Files to Create**:
- `test/integration/layer_build_test.go` (new)
- `test/fixtures/manifests/` (test manifests)
- `test/fixtures/layers/` (test layers)

**Test Scenarios**:
1. Build with base layer only -> verify boots in Firecracker
2. Build with multiple layers -> verify layer order applied
3. Cached build runs faster than first build
4. Invalid manifest -> clear error message
5. Missing layer source -> clear error message
6. Layer with hooks -> hooks executed in order
7. File conflict between layers -> warning logged

**Performance Benchmark**: Cached build must complete in <30 seconds
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Test: build with base layer only produces bootable rootfs.ext4
- [x] #2 Test: build with 3 layers verifies correct layer application order
- [x] #3 Test: cached build completes faster than initial build
- [x] #4 Test: invalid manifest returns clear error with field path
- [x] #5 Test: missing layer source returns actionable error message
- [ ] #6 Test: layer hooks execute pre-install before copy, post-install after
- [x] #7 Benchmark: cached build completes in under 30 seconds
<!-- AC:END -->
