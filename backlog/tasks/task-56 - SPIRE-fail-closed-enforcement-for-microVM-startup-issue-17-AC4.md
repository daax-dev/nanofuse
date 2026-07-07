---
id: TASK-56
title: 'SPIRE fail-closed enforcement for microVM startup (issue #17 AC4)'
status: In Progress
assignee:
  - '@claude'
created_date: '2026-07-07 01:41'
updated_date: '2026-07-07 01:56'
labels: []
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Config-gated default-OFF fail-closed: when SPIRE.Enabled && SPIRE.Required, a SPIRE registration failure during VM create fails the request (503, SPIRE-naming error) with full resource cleanup. Spec: .flowspec/features/issue-17-spire-enforcement/spec.md
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Required+SPIRE-unavailable: create fails 503 naming SPIRE, no leaked VM/network/storage/entry
- [x] #2 Required+SPIRE-ok: create succeeds with SpiffeID
- [x] #3 not-Required+SPIRE-unavailable: create still succeeds (WARN only)
- [x] #4 SPIRE.Required defaults to false
- [x] #5 Full local gate green incl golangci-lint v2.12.2
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Config-gated fail-closed SPIRE enforcement for issue #17 DoD AC4.

- Added SPIREConfig.Required (yaml: required, default false). Enforcement active only when Enabled && Required.
- handleCreateVM: on SPIRE registration failure with spireRequired(), aborts with 503 SERVICE_UNAVAILABLE naming SPIRE unreachability and calls cleanupCreatedVMResources (no leaked VM/network/storage/entry). Not-required path unchanged (WARN, proceed).
- Introduced package-private spireRegistrar interface so tests can stub SPIRE (concrete *spire.Service shells to docker exec, cannot succeed in unit tests).
- Config.Validate() rejects Required && !Enabled as fatal startup error (closes silent fail-open gap; gemini premortem).
- Tests: fail-closed 503 + storage-cleanup proof, required+ok success, not-required+fail success, spireRequired gate matrix, default-off, config validation accept/reject.

Gate: gofmt clean, go vet clean, golangci-lint v2.12.2 = 0 issues, go test -race ./... all pass, mage test pass.
Validation: produced by Claude Opus 4.8; adversarially reviewed by gemini-2.5-pro (4 rounds incl premortem).
<!-- SECTION:NOTES:END -->
