---
id: task-42
title: 'T019: Implement CLI layer create/validate Commands'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:17'
updated_date: '2026-01-08 02:33'
labels:
  - phase-4
  - cli
  - layer
  - authoring
  - flowspec-microvm
  - implement
dependencies: []
priority: low
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add layer scaffolding and validation commands.

**Context**: Part of Phase 4 - layer authoring tooling.
**Dependency**: T006 (CLI structure)

**Files to Create**:
- `cmd/nanofuse/layer_create.go` (new)
- `cmd/nanofuse/layer_validate.go` (new)

**layer create Command**:
```bash
nanofuse layer create my-feature --type feature --output ./layers/my-feature
nanofuse layer create python-app --type application
```

Generates:
- layer.yaml (template with required fields)
- rootfs/ directory
- hooks/ directory with template scripts
- tests/ directory

**layer validate Command**:
```bash
nanofuse layer validate ./layers/my-feature
nanofuse layer validate ./layers/my-feature --strict
```

Checks:
- layer.yaml valid and complete
- Required fields present
- rootfs/ exists and not empty
- Hooks are executable
- No invalid file permissions
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 layer create generates layer directory structure
- [x] #2 Template layer.yaml with all required fields
- [x] #3 Template hooks/post-install.sh
- [x] #4 layer validate checks layer.yaml syntax
- [x] #5 layer validate checks rootfs/ exists
- [x] #6 layer validate checks hooks are executable
- [x] #7 Suggest fixes for common issues
- [x] #8 Exit code reflects validation result
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented layer create/validate CLI commands with scaffolding and validation. PR #88 created: https://github.com/peregrinesummit/nanofuse/pull/88
<!-- SECTION:NOTES:END -->
