---
id: TASK-50
title: Refresh PR30 auth branch against current main
status: In Progress
assignee: []
created_date: '2026-05-30 23:18'
updated_date: '2026-05-30 23:27'
labels:
  - pr30
  - auth
  - branch-refresh
  - replacement-pr
dependencies: []
references:
  - /Users/jasonpoley/prj/dx/src/nanofuse-pr30
  - 'https://github.com/daax-dev/nanofuse/pull/30'
documentation:
  - api/README.md
  - config.dev.yaml
  - docs/building/pr30-refresh-audit.md
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Audit the PR30 branch fix/issues-2-3-4-v2 against origin/main after PR #46. Replace stale or unsafe SPIFFE header auth, Aembit policy, and SVID rotation claims with a current-main-compatible production slice that does not trust client-controlled identity headers. Preserve current-main API, egress, capabilities, guidance, and module identity.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Branch diff against origin/main is audited and stale/unfit auth, policy, and rotation pieces are identified.
- [x] #2 Branch-local fixes preserve PR #46/current-main changes and avoid broad unrelated rewrites.
- [x] #3 TCP API auth extracts SPIFFE identity only from verified client certificates, not request headers.
- [x] #4 Documentation states the implemented scope and does not claim Aembit policy or SVID rotation.
- [x] #5 Non-trivial decisions are logged under .logs/decisions/*.jsonl.
- [x] #6 Focused tests and mage ci are run, with exact results reported.
- [x] #7 Closed PR #30 Copilot comments are audited against the current branch head and stale/outdated threads are documented rather than blindly reintroduced.
- [x] #8 Current-head mTLS/API auth defects found during the audit are fixed on fix/issues-2-3-4-v2 without reverting current-main behavior.
- [ ] #9 The updated branch is pushed and a replacement PR is opened from the same branch.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Use gh review-thread/REST data for closed PR #30 to separate outdated Copilot comments from comments still represented in the current diff.
2. Inspect the current branch against origin/main for mTLS auth behavior, config validation, and spoofed-header protections that are under-tested or under-documented.
3. Fix only branch-local current-head gaps: add explicit spoofed X-SPIFFE-ID rejection coverage, add config validation tests for TCP mTLS requirements, and tighten trust-boundary documentation/comments.
4. Append a JSONL decision explaining the explicit header-spoofing/config-validation hardening and why stale policy/rotation Copilot comments are not reintroduced.
5. Run gofmt, focused Go tests, diff checks, JSONL parsing, and mage ci.
6. Commit, push fix/issues-2-3-4-v2, and open a replacement PR from the same branch with review audit and validation evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Completed. Effective branch-specific diff against origin/main is limited to TCP mTLS SPIFFE identity extraction, docs/config updates, this task, and decision logging. The branch no longer trusts X-SPIFFE-ID or policy headers and does not include Aembit policy or SVID rotation claims. Unix socket access remains local/plain and relies on filesystem permissions.

Reopened on 2026-05-30 for replacement PR hardening after closed PR #30 retained outdated Copilot threads but no open current-head review threads.

Replacement hardening completed: closed PR #30 Copilot threads were stale/outdated against old commit f21a353; current-head fixes add spoofed X-SPIFFE-ID rejection/ignore coverage and TCP mTLS config validation tests. Validation: go test ./internal/api; go test ./internal/config; jq -c . .logs/decisions/auth.jsonl; git diff --check; mage ci.
<!-- SECTION:NOTES:END -->
