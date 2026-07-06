---
id: TASK-54
title: 'Codex review fixes for per-VM credential isolation (#19, supersedes PR #137)'
status: Done
assignee:
  - '@jpoley'
created_date: '2026-07-03 04:59'
labels:
  - security
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Address a harsh Codex review of the feat/nanofuse-19-credential-isolation branch after PR #137 was closed. Fix the mount-guard path byte-semantics finding; reject the verifier-default finding as a documented design decision (--strict already covers it).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GuardMounts reasons about raw mount-target bytes (no TrimSpace); leading-whitespace targets fail closed with ErrInvalidMountTarget
- [x] #2 Private-tmpfs exception requires an exactly-empty Source (whitespace no longer widens it)
- [x] #3 Tests cover the tightened byte-semantics and narrow tmpfs exception
- [x] #4 mage ci green with CI-pinned golangci-lint v2.12.2; new PR opened on same branch
<!-- AC:END -->
