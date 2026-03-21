---
id: task-25
title: Base Variant Rootfs Builder (~50MB)
status: To Do
assignee:
  - '@pm-planner'
created_date: '2025-12-23 00:33'
labels:
  - implement
  - rootfs-pipeline
  - base-variant
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Build flowspec-base rootfs variant containing Alpine + Python + flowspec-cli. Target size ~50MB. Supports tech-writer, researcher, product-manager, and business-validator agents (60% of use cases). Includes tini init system for lightweight process management. Must be bootable in Firecracker with <200ms boot time.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Builds base variant with Alpine, Python, and flowspec-cli
- [ ] #2 Total size ≤ 60MB (target 50MB)
- [ ] #3 Includes tini init system
- [ ] #4 Bootable in Firecracker microVM
- [ ] #5 Boot time < 200ms verified
- [ ] #6 Supports all base agent types (tech-writer, researcher, product-manager, business-validator)
<!-- AC:END -->
