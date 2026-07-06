---
id: TASK-55
title: 'Implement snapshot-load/resume path (issue #227 AC1)'
status: In Progress
assignee:
  - '@claude'
created_date: '2026-07-06 22:49'
updated_date: '2026-07-06 23:14'
labels:
  - issue-227
  - snapshot
  - firecracker
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Wire Firecracker PUT /snapshot/load into the resume path so 'vm resume --from-snapshot <id>' boots a VM from a previously created LOCAL snapshot. Spec: .flowspec/features/issue-227-snapshot-load/spec.md. AC2 (cross-node, depends on #130/#250) is out of scope.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Runtime LoadSnapshot sends PUT /snapshot/load with snapshot_path, mem_backend{backend_type:File,backend_path}, resume_vm:true on a fresh Firecracker process
- [x] #2 resume --from-snapshot restores a non-running VM and leaves it Running
- [x] #3 Snapshot resume validates existence, ownership, and backing-file presence with actionable errors
- [x] #4 Running VM rejected with conflict; unsupported runtime mapped to 501
- [x] #5 Unit tests cover request schema, validation, state machine, and error mapping; e2e documented
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. TDD: firecracker unit tests (load request schema, path/file validation, waitForSocketReady) + api handler tests (happy path, running-conflict, not-found, wrong-owner, missing-files, unsupported->501)\n2. Implement firecracker Manager.LoadSnapshot + sendSnapshotLoad + waitForSocketReady; refactor startFirecrackerProcess to skip --config-file when empty\n3. Add LoadSnapshot to vmm.Manager interface; applecontainer returns unsupported; extend api test stub\n4. Wire handleVMResumeByPath -> resumeVMFromSnapshot + ensureVMTapForResume\n5. Full gate: go test -race, vet, gofmt, golangci-lint\n6. Gemini adversarial review x3 + premortem
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Problem
nanofuse could create Firecracker snapshots but had no restore path: Manager.Resume and the resume handler ignored the snapshot; nothing called Firecracker PUT /snapshot/load. Snapshots were write-only.

## Approach (AC1, local)
- firecracker.Manager.LoadSnapshot: starts a FRESH Firecracker process with no --config-file (snapshot supplies machine state), waits for the API socket, then PUT /snapshot/load with mem_backend{backend_type:File},resume_vm:true. Reaps the process (no zombies); kills+reaps on any post-spawn failure.
- Schema verified against Firecracker v1.15 swagger (SnapshotLoadParams/MemoryBackend); mem_backend preferred over deprecated mem_file_path; resume_vm avoids a second PATCH /vm.
- vmm.Manager gains LoadSnapshot; applecontainer returns ErrUnsupportedOperation.
- handleVMResumeByPath branches on snapshot_id -> resumeVMFromSnapshot: rejects running/resuming/starting VMs (409), validates snapshot existence/ownership/in-root-path/backing-files, re-establishes host tap idempotently, sets Resuming -> LoadSnapshot -> Running with state restore/kill on failure, maps unsupported runtime to 501.
- CLI/client already wired (--from-snapshot).

## Security/robustness (from 3 rounds gemini-2.5-pro review + premortem)
- Path-traversal guard on stored snapshot paths (mirrors handleDeleteSnapshot).
- 0750 vmDir perms. Tap cleanup on attach failure. Mark-failed on state-restore double-failure; kill orphaned runtime if final persist fails.

## Tests
- firecracker: sendSnapshotLoad schema (100%), waitForSocketReady (100%), LoadSnapshot validation, missing-files.
- api: happy path + running-conflict + not-found + wrong-owner + missing-files + path-outside-root + unsupported->501.
- Gate: gofmt clean, go vet clean, golangci-lint v2.12.2 (CI-pinned) 0 issues, go test -race ./... green, gosec at baseline parity (no new findings).

## e2e
Real KVM resume NOT run (agent user lacks /dev/kvm + sudo; host firecracker v1.7.0 != v1.15 target). Exact procedure documented in .flowspec/features/issue-227-snapshot-load/e2e.md. Unit + primary-source verified; fresh-process spawn + live /snapshot/load resume remain for KVM.

## Out of scope (documented follow-ups)
AC2 cross-node resume (depends on #130/#250). Snapshot arch-match + config-drift validation (need snapshot-record schema change; not reachable in AC1 happy path).

Validation: producer claude-opus-4-8, validator gemini-2.5-pro (cross-provider), verdict: no outstanding correctness issues.
<!-- SECTION:NOTES:END -->
