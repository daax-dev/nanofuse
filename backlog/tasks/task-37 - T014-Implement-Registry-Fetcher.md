---
id: task-37
title: 'T014: Implement Registry Fetcher'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:17'
updated_date: '2026-01-08 02:33'
labels:
  - phase-3
  - registry
  - fetcher
  - oci
  - flowspec-microvm
  - implement
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Fetch layers from OCI registries (ghcr.io, Docker Hub, etc.).

**Context**: Part of Phase 3 - extends fetcher with registry support.
**Dependency**: T004 (base fetcher)

**Files to Modify**:
- `internal/layerbuild/fetcher.go` (extend)
- `internal/layerbuild/fetcher_test.go` (extend)
- `go.mod` (add go-containerregistry dependency)

**Source Format**: `registry://ghcr.io/nanofuse/layers/python:3.12`

**Implementation**:
```go
type RegistryFetcher struct {
  auth authn.Authenticator  // For private registries
}

func (f *RegistryFetcher) Fetch(source string) (*FetchResult, error)
```

**Features**:
- Parse registry URL into image reference
- Support authenticated and anonymous pulls
- Download layer blob by digest
- Verify digest matches manifest
- Retry with exponential backoff for rate limits

**Registry Support**: ghcr.io, Docker Hub, any OCI-compliant registry
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Parse registry:// source into image reference
- [x] #2 Support authenticated pulls with docker config
- [x] #3 Support anonymous pulls for public images
- [x] #4 Download layer blob and verify SHA256 digest
- [x] #5 Retry with exponential backoff on rate limit (429)
- [x] #6 Integration test with ghcr.io public image
- [x] #7 Integration test with Docker Hub public image
- [x] #8 Handle registry errors with clear messages
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented RegistryFetcher with go-containerregistry library

Supports authenticated pulls via docker config keychain

Supports anonymous pulls for public images

Exponential backoff on HTTP 429 rate limit errors

Clear error messages for auth, network, and registry errors

Integration tests pass for Docker Hub and ghcr.io
<!-- SECTION:NOTES:END -->
