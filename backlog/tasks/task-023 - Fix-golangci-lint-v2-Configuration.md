---
id: task-023
title: Fix golangci-lint v2 Configuration
status: Done
assignee: []
created_date: '2025-11-27'
updated_date: '2025-12-30 01:26'
labels:
  - Bug
  - CI
  - P0
  - Blocking
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
**Blocker:** GitHub Actions CI is failing because golangci-lint v2.x requires a version field in the configuration file.

## Error Message

```
Error: can't load config: unsupported version of the configuration: ""
See https://golangci-lint.run/docs/product/migration-guide for migration instructions
Error: golangci-lint exit with code 3
```

## Root Cause

golangci-lint v2.0+ introduced a breaking change requiring a `version` field at the top of `.golangci.yml`. The current config file lacks this field.

## Fix Required

Add `version: "2"` to the top of `.golangci.yml`:

```yaml
version: "2"

run:
  timeout: 5m
  # ... rest of config
```

## Migration Guide

See: https://golangci-lint.run/docs/product/migration-guide

Key changes in v2:
1. `version` field is now required
2. Some linter configurations may have changed
3. Some deprecated linters removed

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Add `version: "2"` to `.golangci.yml`
- [x] #2 Verify CI lint job passes
- [x] #3 Test locally with `golangci-lint run`

## Priority

**P0 - Blocking**: CI is completely broken until this is fixed.
<!-- SECTION:DESCRIPTION:END -->

<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Already fixed - `.golangci.yml` has `version: "2"` at line 1. Lint passes locally with golangci-lint v2.7.2.
<!-- SECTION:NOTES:END -->
