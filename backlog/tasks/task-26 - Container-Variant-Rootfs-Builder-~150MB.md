---
id: task-26
title: Container Variant Rootfs Builder (~150MB)
status: To Do
assignee:
  - '@pm-planner'
created_date: '2025-12-23 00:33'
labels:
  - implement
  - rootfs-pipeline
  - container-variant
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Build flowspec-container rootfs variant containing base variant + containerd + nerdctl. Target size ~150MB. Supports backend-engineer, frontend-engineer, platform-engineer, and sre-agent (40% of use cases). Enables rootless container builds. Must maintain <200ms boot time despite larger size.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Builds container variant with base + containerd + nerdctl
- [ ] #2 Total size ≤ 180MB (target 150MB)
- [ ] #3 Supports rootless container builds (buildkit/podman)
- [ ] #4 Bootable in Firecracker microVM
- [ ] #5 Boot time < 200ms verified
- [ ] #6 Supports all container-capable agent types (backend-engineer, frontend-engineer, platform-engineer, sre-agent)
<!-- AC:END -->
