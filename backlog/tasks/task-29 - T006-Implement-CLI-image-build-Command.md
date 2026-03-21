---
id: task-29
title: 'T006: Implement CLI image build Command'
status: Done
assignee:
  - '@devops-engineer'
created_date: '2025-12-22 23:15'
updated_date: '2026-01-08 02:40'
labels:
  - phase-1
  - cli
  - build
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add `nanofuse image build --manifest` command for layer-based image building.

**Context**: Part of Phase 1 - depends on T005 (composer).
**Dependency**: T005 (composer)

**Files to Create/Modify**:
- `cmd/nanofuse/main.go` (extend with image subcommand)
- `cmd/nanofuse/image_build.go` (new)

**Command Syntax**:
```bash
nanofuse image build --manifest image.manifest.yaml --output ./build/
nanofuse image build -m image.manifest.yaml -o ./build/ --verbose
nanofuse image build -m image.manifest.yaml --dry-run
```

**Flags**:
- `--manifest, -m` (required): Path to manifest file
- `--output, -o` (default: ./build): Output directory
- `--verbose, -v`: Detailed logging
- `--dry-run`: Validate manifest without building
- `--no-cache`: Skip layer cache
- `--parallel`: Parallel layer fetching (default: 4)

**Exit Codes**: 0=success, 1=build error, 2=validation error
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 --manifest flag accepts path to manifest file
- [x] #2 --output flag specifies output directory (default ./build/)
- [x] #3 --verbose flag enables detailed logging with layer progress
- [x] #4 --dry-run validates manifest without building image
- [x] #5 Progress output shows layer fetch and composition steps
- [x] #6 Exit code 0 on success, 1 on build error, 2 on validation error
- [x] #7 Integration test: CLI -> build -> verify artifacts exist
<!-- AC:END -->
