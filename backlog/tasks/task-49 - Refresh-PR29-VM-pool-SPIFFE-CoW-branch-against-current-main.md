---
id: TASK-49
title: Refresh PR29 VM pool/SPIFFE/CoW branch against current main
status: Done
assignee: []
created_date: '2026-05-30 23:01'
updated_date: '2026-05-30 23:41'
labels:
  - pr29
  - branch-refresh
  - code-quality
  - replacement-pr
dependencies: []
references:
  - /Users/jasonpoley/prj/dx/src/nanofuse-pr29
  - 'https://github.com/daax-dev/nanofuse/pull/29'
  - 'https://github.com/daax-dev/nanofuse/pull/47'
documentation:
  - api/openapi.yaml
  - api/README.md
  - docs/building/pr29-refresh-audit.md
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Audit the PR29 branch fix/issues-11-12-13-v2 against origin/main after PR #46. Identify stale or unfit implementation claims around VM pool cold starts, per-VM SPIFFE SVIDs, and CoW filesystem layers. Make branch-local fixes only after plan approval. Do not push.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Branch diff against origin/main is audited and stale/unfit pieces are identified.
- [x] #2 Branch-local fixes preserve PR #46/current-main changes and avoid broad unrelated rewrites.
- [x] #3 Documentation and tests accurately describe implemented and verified behavior without claiming hardware Firecracker validation unless run.
- [x] #4 Non-trivial decisions are logged under .logs/decisions/*.jsonl.
- [x] #5 Focused formatter/linter/tests are run where possible, with exact commands and results reported.
- [x] #6 Closed PR #29 Copilot comments are audited against the current branch head and stale/outdated threads are documented rather than blindly reintroduced.
- [x] #7 Current-head snapshot/API defects found during the audit are fixed on fix/issues-11-12-13-v2 without reverting current-main behavior.
- [x] #8 The updated branch is pushed and a replacement PR is opened from the same branch.
- [x] #9 Existing pause/resume API stubs are wired to Firecracker VM state calls or explicitly documented if validation blocks that wiring.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Use gh review-thread/REST data for closed PR #29 to separate outdated Copilot comments from comments still represented in the current diff.
2. Inspect the current branch against origin/main for snapshot/API behavior that is inconsistent with Firecracker snapshot requirements or project docs.
3. Fix branch-local current-head defects: require API-created snapshots to target paused VMs; wire existing pause/resume handlers to Firecracker's pinned VM state API so callers can reach the valid snapshot state; align OpenAPI/user docs/audit notes; and add regression coverage for the state gate plus Firecracker pause/resume requests.
4. Append JSONL decisions explaining the state-gate hardening, pause/resume wiring, and why stale Copilot comments are not reintroduced.
5. Run gofmt, focused Go tests, diff checks, JSONL parsing, and mage ci.
6. Commit, push fix/issues-11-12-13-v2, and open a replacement PR from the same branch with review audit and validation evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Completed. Effective branch-specific diff against origin/main is limited to Firecracker snapshot API implementation/tests plus PR29 audit docs/logs. Old VMPool/CoW/SVID code is removed by merge resolution because those files do not exist on current main. Current-main PR46 API/egress/capabilities, Flowspec guidance, golangci config, go.mod/go.sum, and daax-dev module identity are preserved. Firecracker request fields target repo-pinned v1.7.0 swagger; future Firecracker upgrades require API-field review. No hardware Firecracker validation was run and no latency/pool claim is made.

Reopened on 2026-05-30 for replacement PR hardening after closed PR #29 retained outdated Copilot threads but no open current-head review threads.

Replacement hardening completed: closed PR #29 Copilot threads were stale/outdated against old commit a74d209; current-head fixes require paused snapshots and wire pause/resume to Firecracker PATCH /vm. Validation: go test ./internal/api; go test ./internal/firecracker; jq -c . .logs/decisions/pr29-refresh.jsonl; git diff --check; mage ci.

Fresh Copilot review on replacement PR #47 identified that LoadSnapshotWithResume cannot restore into a real Firecracker process because /snapshot/load must run before microVM configuration. Scope revised to remove the unusable load helper and document restore as future restore-specific launch work.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replacement PR open: https://github.com/daax-dev/nanofuse/pull/47. Closed PR #29 Copilot threads were audited and found stale/outdated against old commit a74d209. Fresh PR #47 Copilot feedback was addressed by removing the attempted LoadSnapshotWithResume helper because Firecracker requires /snapshot/load before microVM configuration; restore now remains explicitly future restore-specific launch work. Current branch hardening requires paused VMs for snapshot creation, wires pause/resume handlers to Firecracker v1.7.0 PATCH /vm state transitions, updates OpenAPI/API docs/audit docs, and records JSONL decisions pr29-refresh-003 through pr29-refresh-005. Validation passed after the fresh fix: go test ./internal/api ./internal/firecracker; jq -c . .logs/decisions/pr29-refresh.jsonl .logs/references/pr29-refresh.jsonl; git diff --check; mage ci. mage ci printed existing non-fatal gosec-not-found and macOS linker warnings.
<!-- SECTION:FINAL_SUMMARY:END -->
