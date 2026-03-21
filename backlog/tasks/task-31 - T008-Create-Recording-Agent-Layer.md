---
id: task-31
title: 'T008: Create Recording Agent Layer'
status: To Do
assignee: []
created_date: '2025-12-22 23:16'
labels:
  - phase-2
  - recording
  - layer
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Package recording agent as a feature layer for session capture.

**Context**: Part of Phase 2 (Recording Integration) - depends on Phase 1 completion.
**Dependency**: T005 (composer) for layer packaging

**Files to Create**:
- `layers/recording-agent/layer.yaml`
- `layers/recording-agent/rootfs/usr/local/bin/record-agent`
- `layers/recording-agent/rootfs/etc/systemd/system/record-agent.service`
- `layers/recording-agent/hooks/post-install.sh`
- `layers/recording-agent/tests/validate.sh`

**Recording Agent Features**:
- Captures terminal I/O via PTY interception
- Streams events to host via virtio-vsock (port 52)
- Local ring buffer (16MB default) for reliability
- Configurable capture modes: terminal, file_io, network

**Layer Configuration Schema**:
```yaml
config_schema:
  vsock_port:
    type: integer
    default: 52
  buffer_size_mb:
    type: integer
    default: 16
  capture_modes:
    type: array
    default: ["terminal"]
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 layers/recording-agent/layer.yaml with metadata and config schema
- [ ] #2 Recording agent binary in rootfs/usr/local/bin/
- [ ] #3 Systemd service file for auto-start on boot
- [ ] #4 Post-install hook enables systemd service
- [ ] #5 Layer validates with nanofuse layer validate
- [ ] #6 Agent starts and connects to vsock on VM boot
- [ ] #7 Configurable vsock port and buffer size via layer config
<!-- AC:END -->
