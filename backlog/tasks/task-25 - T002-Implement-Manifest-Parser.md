---
id: task-25
title: 'T002: Implement Manifest Parser'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:15'
updated_date: '2025-12-23 01:08'
labels:
  - phase-1
  - core
  - manifest
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Parse and validate YAML image manifests with condition evaluation and dependency resolution.

**Context**: Part of Phase 1 - depends on T001 (task-24) for types.
**Dependency**: task-24 (T001: Types)

**Files to Create**:
- `internal/layerbuild/manifest.go`
- `internal/layerbuild/manifest_test.go`

**Key Functions**:
- ParseManifest(path string) (*ImageManifest, error)
- ValidateManifest(m *ImageManifest) error
- EvaluateConditions(m *ImageManifest, env map[string]string) []LayerReference
- ResolveDependencies(layers []LayerReference) ([]LayerReference, error)

**Condition Syntax**: `${VAR_NAME:-default}` with environment substitution

**Validation Rules**:
- version must be "1.0"
- name must match pattern ^[a-z0-9][a-z0-9-]*[a-z0-9]$
- kernel section required with version, source, cmdline
- at least one layer required
- sha256 required for remote sources
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Parse YAML manifest into ImageManifest struct
- [x] #2 Validate required fields with clear error messages including line numbers
- [x] #3 Evaluate layer conditions with environment variable substitution
- [x] #4 Resolve layer dependencies and detect circular dependencies
- [x] #5 Return validation errors with field path and reason
- [x] #6 Unit tests cover all validation rules with >80% coverage
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Create internal/layerbuild/ package directory
2. Implement types.go with all layer types (T001 dependency)
3. Write manifest_test.go tests first (TDD)
4. Implement manifest.go parser with validation
5. Implement condition evaluation with env var substitution
6. Implement dependency resolution with cycle detection
7. Achieve >80% test coverage
8. Run go vet and staticcheck
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation Complete

### Files Created
- `internal/layerbuild/types.go` - All layer types, interfaces, validation error types
- `internal/layerbuild/manifest.go` - Manifest parser, validator, condition evaluator, dependency resolver
- `internal/layerbuild/manifest_test.go` - Comprehensive test suite

### Key Functions Implemented
- `ParseManifest(path string) (*ImageManifest, error)` - YAML parsing with defaults
- `ValidateManifest(m *ImageManifest) error` - Field validation with clear error messages
- `EvaluateConditions(m *ImageManifest, env map[string]string) []LayerReference` - Environment variable substitution
- `ResolveDependencies(layers []LayerReference) ([]LayerReference, error)` - Topological sort with cycle detection
- `ParseSourceType(source string) (SourceType, bool)` - Source URL scheme parsing

### Test Results
- **All 30 tests pass** (including subtests)
- **Coverage: 92.9%** (exceeds 80% requirement)
- **go vet**: Clean
- **Race detector**: Clean

### Validation Rules Implemented
- version must be "1.0"
- name must match `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
- kernel.version, kernel.source, kernel.cmdline required
- At least one layer required
- Layer type must be: base, runtime, feature, application
- sha256 required for registry:// and url:// sources (not local:// or docker://)

### Condition Syntax
- `${VAR_NAME:-default}` - Use VAR_NAME or default if not set
- Truthy values: "true", "1", "yes", "on" (case-insensitive)
- Required layers always included regardless of condition
<!-- SECTION:NOTES:END -->
