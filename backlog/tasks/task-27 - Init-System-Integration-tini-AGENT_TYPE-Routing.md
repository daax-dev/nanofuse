---
id: task-27
title: Init System Integration (tini + AGENT_TYPE Routing)
status: To Do
assignee:
  - '@pm-planner'
created_date: '2025-12-23 00:33'
labels:
  - implement
  - rootfs-pipeline
  - init-system
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Configure tini as lightweight init system with entrypoint script for agent type selection. Handle AGENT_TYPE environment variable routing to correct flowspec-agent. Support signal handling and zombie process reaping. Must work for both base and container variants.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 tini configured as /sbin/init
- [ ] #2 Entrypoint script handles AGENT_TYPE environment variable
- [ ] #3 Routes to correct flowspec-agent based on AGENT_TYPE
- [ ] #4 Signal handling works correctly (SIGTERM, SIGINT)
- [ ] #5 Zombie process reaping verified
- [ ] #6 Works in both base and container variants
<!-- AC:END -->
