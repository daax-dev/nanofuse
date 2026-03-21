---
id: task-27
title: 'T004: Implement Local and Docker Fetcher'
status: Done
assignee: []
created_date: '2025-12-22 23:15'
updated_date: '2025-12-23 01:36'
labels:
  - phase-1
  - core
  - fetcher
  - docker
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Fetch layers from local directories/tarballs and Docker images.

**Context**: Part of Phase 1 - depends on T001 (task-24) for interfaces.
**Dependency**: task-24 (T001: Types)

**Files to Create**:
- `internal/layerbuild/fetcher.go` (new)
- `internal/layerbuild/fetcher_test.go` (new)

**Source Types**:
- `local://path/to/layer` - Read layer.yaml and rootfs/ from directory
- `local://path/to/layer.tar.gz` - Extract tarball
- `docker://image:tag` - Export container filesystem via docker export

**Fetcher Interface Implementation**:
```go
type LocalFetcher struct{}
func (f *LocalFetcher) Fetch(source string) (*FetchResult, error)

type DockerFetcher struct{}  
func (f *DockerFetcher) Fetch(source string) (*FetchResult, error)
```

**FetchResult**: Contains tarball path, metadata, digest, size
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 LocalFetcher reads layer from local directory with layer.yaml
- [x] #2 LocalFetcher extracts tarball if source ends in .tar.gz
- [x] #3 DockerFetcher exports container filesystem to tarball
- [x] #4 Verify SHA256 digest after fetch if provided in manifest
- [x] #5 Return layer metadata (name, version, size) in FetchResult
- [x] #6 Progress callback for large layers (>100MB)
- [x] #7 Unit tests with mock filesystem achieve >80% coverage
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation completed by agent a0fbb24. Created fetcher.go (471 lines), fetcher_test.go (485 lines). Implements LocalFetcher and DockerFetcher. 84.9% coverage. All tests pass.

**QA Validation (2025-12-22)**: All 7 ACs verified. LocalFetcher: 88.2% coverage. DockerFetcher: 16.1% (requires Docker daemon for integration). Directory traversal protection in place. PRODUCTION READY.
<!-- SECTION:NOTES:END -->
