---
id: task-39
title: 'T016: Create Debug Tools Layer'
status: In Progress
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:17'
updated_date: '2026-01-08 02:41'
labels:
  - phase-3
  - layer
  - debug
  - tools
  - flowspec-microvm
  - implement
dependencies: []
priority: low
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Package debugging tools as a feature layer.

**Context**: Part of Phase 3 - additional feature layer.
**Dependency**: T005 (composer)

**Files to Create**:
- `layers/debug-tools/layer.yaml`
- `layers/debug-tools/rootfs/...` (tool binaries)
- `layers/debug-tools/hooks/post-install.sh`

**Tools to Include**:
- strace (syscall tracing)
- ltrace (library call tracing)
- gdb (GNU debugger)
- tcpdump (network capture)
- htop (process viewer)
- curl, wget (network tools)

**Layer Configuration**:
```yaml
name: debug-tools
version: 1.0.0
type: feature
description: Debugging and diagnostic tools
provides:
  - debugging
  - network-tools
  - tracing
```

**Size Target**: <50MB (tools only, no documentation)
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 layers/debug-tools/layer.yaml with metadata
- [x] #2 Include strace, ltrace, gdb, tcpdump, htop
- [x] #3 Include curl and wget for network testing
- [ ] #4 Layer size under 50MB
- [x] #5 All tools executable in VM
- [ ] #6 Layer validates with nanofuse layer validate
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Created debug-tools layer with strace, ltrace, gdb, tcpdump, htop, curl, wget. Docker image builds successfully and all tools are verified executable. Layer size is 62MB (slightly over 50MB target due to GDB which is 12.5MB alone). PR #82 created: https://daax-dev/nanofuse/pull/82

2026-01-07: AC #6 blocked - `layer validate` expects rootfs/ directory but all layers use Dockerfile approach. Either validation needs to support Dockerfile layers, or a build step should extract rootfs. AC #4 (size <50MB) trade-off documented - GDB adds 12.5MB but provides critical debugging capability.
<!-- SECTION:NOTES:END -->
