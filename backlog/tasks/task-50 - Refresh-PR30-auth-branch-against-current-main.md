---
id: TASK-50
title: Refresh PR30 auth branch against current main
status: Done
assignee: []
created_date: '2026-05-30 23:18'
updated_date: '2026-05-31 00:16'
labels:
  - pr30
  - auth
  - branch-refresh
  - replacement-pr
dependencies: []
references:
  - 'https://github.com/daax-dev/nanofuse/pull/49'
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
- [x] #9 The updated branch is pushed and a replacement PR is opened from the same branch.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Merge origin/main into fix/issues-2-3-4-v2 with a normal merge commit to avoid rewriting the pushed branch history.
2. Rework the PR48 auth feedback locations so the JSON 401 response and client CA parse-path behavior are explicit, test-covered, and not anchored to stale review lines.
3. Update API/building docs, Backlog notes, and .logs/decisions/auth.jsonl for the follow-up scope.
4. Run gofmt, focused go tests, JSONL parsing, git diff --check, and mage ci.
5. Commit, push the same branch, and open a new replacement PR with current validation evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Completed. Effective branch-specific diff against origin/main is limited to TCP mTLS SPIFFE identity extraction, docs/config updates, this task, and decision logging. The branch no longer trusts X-SPIFFE-ID or policy headers and does not include Aembit policy or SVID rotation claims. Unix socket access remains local/plain and relies on filesystem permissions.

Reopened on 2026-05-30 for replacement PR hardening after closed PR #30 retained outdated Copilot threads but no open current-head review threads.

Replacement hardening completed: closed PR #30 Copilot threads were stale/outdated against old commit f21a353; current-head fixes add spoofed X-SPIFFE-ID rejection/ignore coverage and TCP mTLS config validation tests. Validation: go test ./internal/api; go test ./internal/config; jq -c . .logs/decisions/auth.jsonl; git diff --check; mage ci.

Fresh Copilot review on replacement PR #48 identified plaintext auth errors, global slog audit logging, and missing CA path context. Scope revised to return standard JSON errors, route audit events through internal/logging.Logger, and include the CA path in parse errors.

Reopened on 2026-05-30 after replacement PR #48 was closed without merge. Current origin/main now includes PR #47, so the same branch must absorb current main without history rewrite, make the PR48 auth fixes unambiguous for a fresh replacement PR, rerun local gates, push, and open a new PR from fix/issues-2-3-4-v2.

Follow-up replacement PR opened as #49 from the same branch after merging current main and adding explicit mTLS denial/client-CA helper coverage.

Fresh Copilot review on PR #49 identified an inaccurate README sentence about no-cert clients receiving JSON 401 after TLS RequireAndVerifyClientCert, and a non-portable absolute worktree path in task references. The README now distinguishes TLS handshake failure from middleware JSON 401 responses, and references use the branch name plus public PR URLs.

TASK-50 references were further shortened to only the replacement PR URL so the stale multi-line Copilot anchor over the old references block no longer tracks an unrelated public URL line.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replacement PR open: https://github.com/daax-dev/nanofuse/pull/49. Closed PR #48 was not merged, so fix/issues-2-3-4-v2 was updated with current origin/main via a normal merge commit, then hardened with explicit mTLS denial and client CA loading helpers. Fresh PR #49 Copilot comments were addressed by documenting that no-cert TCP clients fail the TLS handshake before HTTP JSON handling and by removing the non-portable absolute worktree path from task references. The branch still scopes auth to TCP mTLS SPIFFE URI SAN identity, rejects/ignores spoofed X-SPIFFE-ID, routes audit events through internal/logging.Logger, returns JSON 401 errors for middleware-level auth denials, and includes configured CA paths in read/parse errors. Validation passed: go test ./internal/api ./internal/config ./internal/types; jq -c . .logs/decisions/auth.jsonl; git diff --check; mage ci. mage ci printed existing non-fatal gosec-not-found, no-tag git describe, and macOS linker warnings.
<!-- SECTION:FINAL_SUMMARY:END -->
