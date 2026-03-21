---
id: task-29
title: GitHub Actions Build Pipeline
status: To Do
assignee:
  - '@pm-planner'
created_date: '2025-12-23 00:33'
labels:
  - implement
  - rootfs-pipeline
  - github-actions
  - ci-cd
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Create GitHub Actions workflow triggered by flowspec-agents and nanofuse-gateway releases. Support manual workflow_dispatch. Build both base and container variants. Sign with cosign, generate SBOMs. Publish to local filesystem storage initially (/var/lib/nanofuse/rootfs/). Tag with semantic versioning (v1.2.3).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Workflow triggers on flowspec-agents release (tag event)
- [ ] #2 Workflow triggers on nanofuse-gateway release (tag event)
- [ ] #3 Supports manual workflow_dispatch trigger
- [ ] #4 Builds both base and container variants
- [ ] #5 Signs images with cosign
- [ ] #6 Generates and publishes SBOMs
- [ ] #7 Tags with semantic versioning
- [ ] #8 Complete workflow runs end-to-end without errors
<!-- AC:END -->
