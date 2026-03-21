---
id: task-24
title: 'T001: Define Layer Types and Interfaces'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:14'
updated_date: '2025-12-23 00:57'
labels:
  - phase-1
  - core
  - types
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Create foundational Go types and interfaces for the layer build system. This is the foundation for all other layer-related tasks.

**Context**: Part of Phase 1 (Core Layer System) for flowspec-microvm-build feature.

**Files to Create**:
- `internal/layerbuild/types.go`
- `internal/layerbuild/types_test.go`

**Key Types**:
- ImageManifest struct (version, name, kernel, layers, output)
- KernelConfig struct (version, source, sha256, cmdline)
- LayerReference struct (name, type, source, sha256, condition, config)
- LayerType enum (base, runtime, feature, application)
- LayerPackage struct (metadata for packaged layers)
- LayerFetcher interface (Fetch method)
- LayerCache interface (Get, Put, Exists methods)

**TDD Approach**: Write tests first for type validation, then implement.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/layerbuild/types.go created with all layer types
- [x] #2 ImageManifest, Layer, LayerConfig structs defined with yaml tags
- [x] #3 LayerFetcher interface with Fetch(source string) (Layer, error) method
- [x] #4 LayerCache interface with Get/Put/Exists methods
- [x] #5 LayerType enum with base/runtime/feature/application values
- [x] #6 Unit tests for type validation achieve >80% coverage
- [x] #7 All tests pass with go test -race
<!-- AC:END -->
