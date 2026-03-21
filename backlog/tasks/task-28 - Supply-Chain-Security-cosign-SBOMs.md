---
id: task-28
title: Supply Chain Security (cosign + SBOMs)
status: To Do
assignee:
  - '@pm-planner'
created_date: '2025-12-23 00:33'
labels:
  - implement
  - rootfs-pipeline
  - security
  - slsa
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement SLSA 1.2 supply chain security with cosign signing and SBOM generation. Use cosign/sigstore for image hash signatures. Generate SBOMs with syft or trivy documenting all components. Publish signatures and SBOMs alongside rootfs images. Enable signature verification before VM launch.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Images signed with cosign/sigstore
- [ ] #2 SBOMs generated with syft or trivy
- [ ] #3 Signatures published to registry alongside images
- [ ] #4 SBOMs published in SPDX or CycloneDX format
- [ ] #5 Verification script can validate signatures
- [ ] #6 Achieves SLSA Level 1.2 compliance
<!-- AC:END -->
