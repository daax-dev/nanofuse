---
id: task-40
title: 'T017: Implement CLI image validate Command'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:17'
updated_date: '2026-01-08 02:33'
labels:
  - phase-4
  - cli
  - validate
  - quality
  - flowspec-microvm
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add `nanofuse image validate` command for image verification.

**Context**: Part of Phase 4 - validation tooling.
**Dependency**: T006 (CLI structure)

**Files to Create**:
- `cmd/nanofuse/image_validate.go` (new)

**Command Syntax**:
```bash
nanofuse image validate ./build/rootfs.ext4
nanofuse image validate ./build/ --report validation.json
nanofuse image validate ./build/ --strict
```

**Validation Checks**:
1. Filesystem integrity (fsck)
2. Required directories exist (/etc, /usr, /var, etc.)
3. Layer metadata in /etc/nanofuse/
4. Systemd configuration valid
5. SSH configuration valid
6. Kernel cmdline compatible

**Output**: Text report or JSON (--report flag)
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Mount ext4 and verify filesystem integrity
- [x] #2 Check required directories exist
- [x] #3 Verify /etc/nanofuse/build-manifest.json present
- [x] #4 Validate systemd services are properly configured
- [x] #5 Validate SSH configuration is correct
- [x] #6 Output text report by default
- [x] #7 Output JSON with --report flag
- [x] #8 Exit code reflects validation result (0=pass, 1=fail)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented image validate command with filesystem, metadata, systemd, and SSH checks. All unit tests passing.
<!-- SECTION:NOTES:END -->
