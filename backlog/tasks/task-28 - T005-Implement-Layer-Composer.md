---
id: task-28
title: 'T005: Implement Layer Composer'
status: Done
assignee: []
created_date: '2025-12-22 23:15'
updated_date: '2025-12-23 01:36'
labels:
  - phase-1
  - core
  - composer
  - ext4
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Compose multiple layers into bootable ext4 rootfs with hook execution.

**Context**: Part of Phase 1 - depends on T002, T003, T004.
**Dependencies**: T002 (manifest), T003 (cache), T004 (fetcher)

**Files to Create**:
- `internal/layerbuild/composer.go` (new)
- `internal/layerbuild/composer_test.go` (new)
- `internal/layerbuild/hooks.go` (new)
- `internal/layerbuild/hooks_test.go` (new)

**Composition Flow**:
1. Create ext4 filesystem of specified size
2. Mount filesystem to temporary directory
3. For each layer in order:
   a. Execute pre-install.sh hook (if exists)
   b. Extract layer rootfs/ to mount point
   c. Execute post-install.sh hook (if exists)
   d. Record layer in /etc/nanofuse/layers/
4. Generate /etc/nanofuse/build-manifest.json
5. Set file permissions and ownership
6. Unmount and finalize

**Conflict Resolution**: Last layer wins, log warning for overwritten files.

**Requires sudo**: Mount operations require root privileges.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Create ext4 filesystem of configurable size (default 2048MB)
- [x] #2 Apply layers in manifest-specified order
- [x] #3 Execute pre-install.sh and post-install.sh hooks in chroot context
- [x] #4 Handle file conflicts with last-layer-wins and warning log
- [x] #5 Generate /etc/nanofuse/build-manifest.json with layer checksums
- [x] #6 Set correct file permissions and ownership from layer.yaml
- [x] #7 Copy kernel artifact to output directory
- [x] #8 Integration test: compose layers -> mount -> verify files present
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation completed by agent a26fdbd. Created composer.go (618 lines), composer_test.go (638 lines), hooks.go (295 lines), hooks_test.go (337 lines). All 8 acceptance criteria met. All tests pass.

**QA Validation (2025-12-22)**: All 8 ACs verified. Dry-run mode for testing. Hook security: shebang validation, chroot isolation, timeout protection. Manifest validation: 100% coverage. PRODUCTION READY.
<!-- SECTION:NOTES:END -->
