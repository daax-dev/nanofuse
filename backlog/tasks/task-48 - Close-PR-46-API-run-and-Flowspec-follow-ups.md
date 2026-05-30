---
id: TASK-48
title: Close PR 46 API run and Flowspec follow-ups
status: Done
assignee: []
created_date: '2026-05-30 20:58'
updated_date: '2026-05-30 21:10'
labels:
  - api
  - docs
  - flowspec
  - sandbox
dependencies: []
references:
  - 'https://github.com/daax-dev/nanofuse/pull/46'
documentation:
  - docs/API_QUICK_START.md
  - docs/MAC_WINDOWS_CLIENTS.md
  - docs/GOALS.md
  - docs/building/sandbox-objective-validation.md
  - docs/building/sandbox-api-comparison.md
  - docs/building/nanofuse-tray-app.md
  - api/openapi.yaml
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Update the existing PR #46 branch so the sandbox objective includes a real documented API run path, current Flowspec naming/artifacts, Mac and Windows client instructions, a current comparison against other sandbox APIs, and a scoped tray/menu-app plan. Keep the work on the existing branch and PR.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 PR-added spec artifacts use .flowspec paths and visible repo guidance describes the current workflow as Flowspec.
- [x] #2 Nanofuse API docs provide a runnable Linux/KVM daemon path and explicit Mac and Windows client instructions over the API.
- [x] #3 API contract and examples match implemented request shapes for image pull and VM create.
- [x] #4 Sandbox API comparison covers relevant peer tools using current official sources and states how Nanofuse differs.
- [x] #5 Tray/menu app requirements are captured for macOS and Windows without introducing an unapproved desktop runtime stack.
- [x] #6 Decision and reference logs are updated in .logs/ as JSONL.
- [x] #7 Relevant tests, syntax checks, and mage ci are run or exact blockers are recorded.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Move PR-added feature artifacts from .flowspec/features/codex-goal to the current Flowspec path and update current workflow references.
2. Add or repair the API run path: API contract, daemon/client docs, Vagrant TCP forwarding, and client configuration where needed.
3. Compare Nanofuse API shape against current official sandbox API references.
4. Capture tray/menu app requirements and decision logs without adding a desktop framework in this PR.
5. Run targeted tests, syntax checks, mage ci, then push updates to PR #46.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Updated PR #46 on the existing codex-goal branch for the API-run, Flowspec, comparison, and tray/menu-app follow-ups. Moved PR-added feature artifacts to .flowspec, corrected current workflow guidance, added GET /capabilities plus client support, added CLI NANOFUSE_* environment configuration, fixed OpenAPI/API examples, added Mac/Windows client runbooks, documented sandbox API differences using official sources, captured tray app requirements without selecting a desktop stack, forwarded Vagrant guest API port 8080 to host 18080, and logged decisions/references/validation in JSONL. Validation: gofmt, git diff --check, focused go tests, go test ./..., YAML/OpenAPI parse, JSONL parse, bash -n, ruby -c, vagrant validate, and mage ci passed. Vagrant closed-loop on local Parallels remains blocked at /dev/kvm not found; VM was halted after the run.
<!-- SECTION:FINAL_SUMMARY:END -->
