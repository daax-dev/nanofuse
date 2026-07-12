---
id: TASK-55
title: 'Implement snapshot-load/resume path (issue #227 AC1)'
status: Done
assignee:
  - '@claude'
created_date: '2026-07-06 22:49'
updated_date: '2026-07-11 22:36'
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
- [x] #1 resume --from-snapshot restores a non-running VM and leaves it Running
- [x] #2 Snapshot resume validates existence, ownership, and backing-file presence with actionable errors
- [x] #3 Running VM rejected with conflict; unsupported runtime mapped to 501
- [x] #4 Unit tests cover request schema, validation, state machine, and error mapping; e2e documented
- [x] #5 Runtime LoadSnapshot sends PUT /snapshot/load (mem_backend{backend_type:File}, resume_vm:false) on a fresh Firecracker process, loads paused, then resumes vCPUs via PATCH /vm after the vsock proxy is listening
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
- firecracker.Manager.LoadSnapshot: starts a FRESH Firecracker process with no --config-file (snapshot supplies machine state), waits for the API socket, then PUT /snapshot/load with mem_backend{backend_type:File}, resume_vm:FALSE (loads paused). It then starts the SPIRE vsock proxy and resumes vCPUs via PATCH /vm {Resumed} - the proxy is listening before any guest vsock traffic (starting it before the load would race Firecracker vsock restore on the same UDS). setupVMRuntime + reaper run only after a successful resume, so a failed resume never persists a stale runtime. Kills+reaps on any post-spawn failure; backing paths must be regular files (Lstat, rejects symlinks/dirs).
- Schema verified against Firecracker v1.15 swagger (SnapshotLoadParams/MemoryBackend); mem_backend preferred over deprecated mem_file_path.
- vmm.Manager gains LoadSnapshot; applecontainer returns ErrUnsupportedOperation.
- handleVMResumeByPath branches on snapshot_id -> resumeVMFromSnapshot: allows only no-live-runtime states (Stopped/Created/Failed), validates snapshot existence/ownership/in-root-path/backing-files (IsNotExist->404, other stat errors->500), re-establishes host tap idempotently, sets Resuming -> LoadSnapshot -> Running with state restore/kill on failure, maps unsupported runtime to 501.
- CLI/client already wired (--from-snapshot).

## Security/robustness (from 3 rounds gemini-2.5-pro review + premortem, plus Copilot review)
- Path-traversal guard on stored snapshot paths (mirrors handleDeleteSnapshot); backing paths must be regular files.
- 0750 vmDir perms (enforced via Chmod). Tap cleanup on attach failure. Mark-failed on state-restore double-failure; kill orphaned runtime if final persist fails. vsock proxy stale-cleanup before restart; no leak.

## Tests
- firecracker: sendSnapshotLoad schema (resume_vm:false), waitForSocketReady, LoadSnapshot validation.
- api: happy path + live-runtime-state conflict + not-found + wrong-owner + missing-files + not-accessible(500) + non-regular-file(500) + path-outside-root + empty-body-unpause + unsupported->501.
- Gate: gofmt/vet clean, golangci-lint v2.12.2 (CI-pinned) 0 issues, go test -race ./... green.

## e2e
Real-KVM resume validated on the dev/vagrant closed-loop harness (nested KVM, Firecracker 1.16.1): create -> SSH marker -> pause -> snapshot -> stop -> resume-from-snapshot booted a fresh Firecracker process reporting state:Running with the pre-snapshot guest marker intact. Durable procedure documented in .flowspec/features/issue-227-snapshot-load/e2e.md. Also unit + primary-source verified (request schema, socket-readiness polling, handler state machine + error mapping).

## Out of scope (documented follow-ups)
AC2 cross-node resume (depends on #130). Snapshot arch-match + config-drift validation (need snapshot-record schema change; not reachable in AC1 happy path).

Validation: producer claude-opus-4-8, validator gemini-2.5-pro (cross-provider) + Copilot review rounds; verdict: no outstanding correctness issues.
<!-- SECTION:NOTES:END -->
