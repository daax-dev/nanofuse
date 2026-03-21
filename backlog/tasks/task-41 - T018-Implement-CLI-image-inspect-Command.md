---
id: task-41
title: 'T018: Implement CLI image inspect Command'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:17'
updated_date: '2026-01-08 02:33'
labels:
  - phase-4
  - cli
  - inspect
  - debug
  - flowspec-microvm
  - implement
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add `nanofuse image inspect` command for layer introspection.

**Context**: Part of Phase 4 - debugging tooling.
**Dependency**: T006 (CLI structure)

**Files to Create**:
- `cmd/nanofuse/image_inspect.go` (new)

**Command Syntax**:
```bash
nanofuse image inspect ./build/rootfs.ext4
nanofuse image inspect ./build/rootfs.ext4 --json
nanofuse image inspect ./build/rootfs.ext4 --layers
```

**Output Information**:
- Image name and version
- Build timestamp
- Kernel version and cmdline
- Layer list with names, versions, digests
- Total size and per-layer sizes
- Capabilities provided

**JSON Schema**:
```json
{
  "name": "nanofuse-flowspec",
  "built_at": "2025-12-22T10:00:00Z",
  "kernel": {"version": "6.1.90", "cmdline": "..."},
  "layers": [
    {"name": "base-os", "version": "1.0.0", "digest": "sha256:...", "size_bytes": 500000000}
  ],
  "total_size_bytes": 600000000
}
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Read /etc/nanofuse/build-manifest.json from image
- [x] #2 Display layer information (name, version, digest)
- [x] #3 Display kernel version and cmdline
- [x] #4 Display build timestamp
- [x] #5 JSON output with --json flag
- [x] #6 Handle images without layer metadata gracefully
- [x] #7 Show per-layer and total sizes
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented image inspect-file command with layer introspection and size calculation. Uses debugfs for non-root access or mount for root.
<!-- SECTION:NOTES:END -->
