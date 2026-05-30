---
id: TASK-50
title: Refresh PR30 auth branch against current main
status: Done
assignee: []
created_date: '2026-05-30 23:18'
updated_date: '2026-05-30 23:18'
labels:
  - pr30
  - auth
  - branch-refresh
dependencies: []
references:
  - /Users/jasonpoley/prj/dx/src/nanofuse-pr30
  - 'https://github.com/daax-dev/nanofuse/pull/30'
documentation:
  - api/README.md
  - config.dev.yaml
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
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Merge origin/main non-destructively into fix/issues-2-3-4-v2.
2. Resolve stale conflicts toward current main for guidance, CI/lint, dependencies, API, config, SPIRE service, and PR46 egress/capabilities.
3. Remove the unsafe policy and SVID rotation packages from the effective PR diff.
4. Replace the trusted-header auth middleware with TCP mTLS configuration and middleware that requires verified client certificates with SPIFFE URI SANs.
5. Update API docs, dev config comments, and decision logs.
6. Run focused tests and mage ci; do not push until parent review.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Completed. Effective branch-specific diff against origin/main is limited to TCP mTLS SPIFFE identity extraction, docs/config updates, this task, and decision logging. The branch no longer trusts X-SPIFFE-ID or policy headers and does not include Aembit policy or SVID rotation claims. Unix socket access remains local/plain and relies on filesystem permissions.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Refreshed PR30 against origin/main fb56363. Removed unsafe policy and rotation code, added TCP listener mTLS setup with client certificate verification, extracted SPIFFE URI SANs only from verified TLS chains, documented the narrowed scope, and logged the auth decision. Validation: go test ./internal/api ./internal/config; go test ./internal/spire; mage ci all passed in the PR30 worktree. Parent review reran git diff --check, JSONL parsing, and focused API/config tests.
<!-- SECTION:FINAL_SUMMARY:END -->
