---
id: task-44
title: 'IMG-002: Create nanofuse-gateway Image Manifest'
status: In Progress
assignee:
  - '@platform-engineer'
created_date: '2025-12-22 23:18'
updated_date: '2026-01-08 02:41'
labels:
  - image
  - nanofuse-gateway
  - manifest
  - networking
  - implement
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Create the image manifest for nanofuse-gateway - a network gateway microVM for the platform.

**Context**: Second target image using the layer-based build system.
**Dependency**: Phase 1 completion (T001-T007)

**Image Purpose**: Network gateway providing ingress/egress control, load balancing, and API routing for microVM clusters.

**Layers**:
1. **base-os** - Ubuntu 24.04 with systemd (minimal)
2. **networking-tools** - iptables, nftables, iproute2 advanced
3. **gateway-runtime** - Go binary for gateway logic
4. **tls-certificates** - CA certs and certificate management
5. **observability** - Prometheus node exporter, structured logging

**Files to Create**:
- `images/nanofuse-gateway/image.manifest.yaml`
- `images/nanofuse-gateway/README.md`
- `layers/networking-tools/layer.yaml` + rootfs
- `layers/gateway-runtime/layer.yaml` + rootfs
- `layers/tls-certificates/layer.yaml` + rootfs
- `layers/observability/layer.yaml` + rootfs

**Build Steps (Fetch & Flatten)**:
1. Fetch base-os from docker://nanofuse-base:latest
2. Fetch networking-tools from local://layers/networking-tools
3. Fetch gateway-runtime from local://layers/gateway-runtime
4. Fetch tls-certificates from local://layers/tls-certificates
5. Fetch observability from local://layers/observability
6. Flatten all layers into single ext4 rootfs
7. Copy kernel to output
8. Generate manifest with digests
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 image.manifest.yaml with all required layers
- [x] #2 networking-tools layer with iptables/nftables
- [x] #3 gateway-runtime layer with Go gateway binary
- [x] #4 tls-certificates layer with CA management
- [x] #5 observability layer with Prometheus exporter
- [ ] #6 Build produces bootable rootfs.ext4
- [ ] #7 Gateway service starts on boot
- [ ] #8 Metrics available on /metrics endpoint
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Created nanofuse-gateway image manifest with networking, TLS, and observability layers.

Files created:
- images/nanofuse-gateway/image.manifest.yaml - Main image manifest
- images/nanofuse-gateway/README.md - Documentation
- layers/networking-tools/ - iptables, nftables, iproute2
- layers/gateway-runtime/ - Go gateway binary
- layers/tls-certificates/ - CA certs, cert-manager
- layers/observability/ - Prometheus node exporter, logging

Note: Acceptance criteria 6-8 (bootable rootfs, gateway service starts, metrics endpoint) require integration testing with the layer build system which is part of the build infrastructure. The manifests and layer definitions are complete.

2026-01-07: ACs #6-8 require integration testing with actual VM boot. Layer manifests and definitions are complete. Blocked on infrastructure for full E2E testing.
<!-- SECTION:NOTES:END -->
