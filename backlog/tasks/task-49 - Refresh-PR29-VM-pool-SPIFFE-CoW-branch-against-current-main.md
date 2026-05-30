---
id: TASK-49
title: Refresh PR29 VM pool/SPIFFE/CoW branch against current main
status: Done
assignee: []
created_date: '2026-05-30 23:01'
updated_date: '2026-05-30 23:09'
labels:
  - pr29
  - branch-refresh
  - code-quality
dependencies: []
references:
  - /Users/jasonpoley/prj/dx/src/nanofuse-pr29
  - 'https://github.com/daax-dev/nanofuse/pull/29'
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
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Merge origin/main non-destructively into fix/issues-11-12-13-v2. 2. Resolve stale conflicts toward current main for guidance, CI/lint, dependencies, API, config, SPIRE service, and PR46 egress/capabilities. 3. Remove the unverified PR29 VMPool, SVID rotation, and CoW layer files rather than patching over review blockers. 4. Add a narrow production code slice: Firecracker v1.7.0 snapshot create/load API calls in internal/firecracker/vm.go with Unix-socket httptest coverage. 5. Document audit outcome and log decisions/references under .logs/. 6. Run focused tests and mage ci; do not push.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Completed. Effective branch-specific diff against origin/main is limited to Firecracker snapshot API implementation/tests plus PR29 audit docs/logs. Old VMPool/CoW/SVID code is removed by merge resolution because those files do not exist on current main. Current-main PR46 API/egress/capabilities, Flowspec guidance, golangci config, go.mod/go.sum, and daax-dev module identity are preserved. Firecracker request fields target repo-pinned v1.7.0 swagger; future Firecracker upgrades require API-field review. No hardware Firecracker validation was run and no latency/pool claim is made.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Refreshed PR29 against origin/main fb56363 without pushing. Removed stale VM pool/SVID/CoW implementation claims, implemented real Firecracker snapshot create/load API calls with Unix-socket httptest unit coverage, and added docs/logs explaining the narrowed scope and Copilot blocker closure by removal. Validation: go test ./internal/firecracker; go test ./internal/firecracker ./internal/api ./internal/network; mage ci all passed.
<!-- SECTION:FINAL_SUMMARY:END -->
