---
id: TASK-47
title: Complete sandbox objective closed-loop validation
status: Done
assignee:
  - codex
created_date: '2026-05-30 20:05'
updated_date: '2026-05-31 10:40'
labels:
  - sandbox
  - microvm
  - vagrant
  - docs
  - testing
dependencies: []
references:
  - >-
    /Users/jasonpoley/prj/dx/arch/may-update/deep-research/sandbox/SUM-Sandbox.md
  - >-
    /Users/jasonpoley/prj/dx/arch/may-update/deep-research/identity/SUM-Identity.md
  - .logs/decisions/sandbox-objective.jsonl
  - .logs/references/sandbox-objective.jsonl
  - .logs/validation/sandbox-objective.jsonl
documentation:
  - docs/GOALS.md
  - docs/building/sandbox-objective-validation.md
  - .flowspec/features/codex-goal/spec.md
  - .flowspec/features/codex-goal/plan.md
  - .flowspec/features/codex-goal/tasks.md
  - .flowspec/features/codex-goal/quickstart.md
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Deliver the repository objective from objective.md on the current branch only. Nanofuse must provide a documented and tested path for microVM-level isolation that covers supported host platforms, container workload wrapping, persistent filesystem behavior, short- and long-running lifetimes, secrets/identity isolation from LLMs, and restricted/interceptable egress for API and MCP traffic. Use the local Vagrant/hypervisor workflow for kernel-level validation and update docs/GOALS.md plus any directly related documentation so the stated goals match validated behavior and known gaps.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A written repo-local specification and implementation plan exist for the sandbox objective and reflect the operator-approved scope.
- [x] #2 The implementation or documentation clearly states the supported host-platform model for Linux, macOS, and Windows and distinguishes validated behavior from planned or constrained behavior.
- [x] #3 Container workload execution through microVM isolation is covered by tests or a closed-loop validation path.
- [x] #4 Persistent filesystem behavior is covered by tests or a closed-loop validation path.
- [x] #5 Short-running fast-start and long-running lifecycle behavior are covered by tests or a closed-loop validation path.
- [x] #6 Secrets/identity handling is documented and tested or validated so secrets are kept out of LLM-visible paths where the project can enforce it.
- [x] #7 Network egress and LLM/API/MCP interception or restriction behavior is documented and tested or validated where the project can enforce it.
- [x] #8 Vagrant/hypervisor closed-loop testing is executed and evidence is recorded, including any host capability constraints encountered.
- [x] #9 docs/GOALS.md and directly related docs are corrected to match the validated design and implementation state.
- [x] #10 Decision logs under .logs/ contain JSONL entries for non-trivial decisions.
- [x] #11 Formatter, linter, and tests including mage ci are run, or every unavailable/failed gate is recorded with exact cause.
- [x] #12 A pull request is opened with problem statement, approach, alternatives, test evidence, and AI producer/validator information.
- [x] #13 daax-dev/vagrant-skill is used as the required local Vagrant harness for this branch, and the exact KVM/runtime result is recorded.
- [x] #14 A minimal macOS/Windows tray/menu app exists as an API-only client with a non-GUI smoke mode and documented one-line launch commands.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Create repo-local spec artifacts under .flowspec/features/codex-goal/ for the sandbox objective: spec.md, plan.md, tasks.md, and a validation quickstart. Keep the spec technology-agnostic; put implementation details in plan/tasks.
2. Implement per-VM writable root disk materialization and cleanup.
3. Implement typed L3/L4 egress policy generation and cleanup with proxy-only behavior.
4. Update GOALS.md, API/OpenAPI docs, and Vagrant closed-loop validation docs.
5. Run mage ci and Vagrant closed-loop validation where the host/provider exposes Linux KVM.
6. Push the branch and open PR #46 with evidence.
7. Reopen the task for the operator-requested correction: use daax-dev/vagrant-skill explicitly, ship a minimal API-only tray/menu app, retest, and open/update the next PR with truthful evidence.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-05-30: Local `mage ci` passed after implementation. `gosec` is not installed; current mage security target records that and continues.

2026-05-30: Vagrant/Parallels guest boot and rsync succeeded, but provisioning failed at `/dev/kvm not found`. A nested virtualization attempt was also tried; Parallels accepted `--nested-virt on` but the VM then failed to start, so this host cannot execute the Firecracker closed-loop locally.

2026-05-30: Parallels VM created for validation is stopped, not destroyed.

2026-05-31: Task reopened because the previous completion record did not use daax-dev/vagrant-skill as the required harness and did not ship the requested tray/menu app. Current correction scope is vagrant-skill validation, tray/menu app implementation, retest, docs, and PR update on the same branch.

2026-05-31: Added `cmd/nanofuse-tray`, `internal/trayapp`, macOS/Windows launch scripts, and `docs/TRAY_APP.md`. Local macOS tests/build/smoke/bounded launch passed. Windows tray executable cross-built. vagrant-skill verify passed with KVM skipped, focused tray smoke passed inside the VM, and synced vagrant-skill `mage ci` passed. The local Parallels guest still reports `KVM_MISSING`, so no local Firecracker boot is claimed.

2026-05-31: Merged current `origin/main` into `codex-goal`, resolved conflicts, reran local `mage ci`, reran `mage ci` inside the synced `daax-dev/vagrant-skill` Parallels VM, reran tray smoke inside the VM, and opened replacement PR https://github.com/daax-dev/nanofuse/pull/55.

2026-05-31: Updated PR #55 after Copilot review and operator SlicerVM guidance so the tray app can select a cached container-derived image and create/start a VM through `nanofused`. Local focused tray tests/build/smoke, macOS bounded launch, local `mage ci`, vagrant-skill focused tray smoke, and vagrant-skill `mage ci` passed. PR #55 was later closed without merge.

2026-05-31: Opened replacement PR https://github.com/daax-dev/nanofuse/pull/56 from the same `codex-goal` branch with the two Copilot review fixes and tray image-launch workflow.

2026-05-31: Fixed the macOS tray launcher so `./scripts/run-tray-macos.sh` no longer appears to do nothing. It now prints build/start status, reports existing `nanofuse-tray` PIDs, supports `--restart`, `--foreground`, `--smoke`, and `--timeout`, and documents the background log path.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Completed sandbox objective validation work on branch `codex-goal` and opened ready-for-review PR https://github.com/daax-dev/nanofuse/pull/46. Implemented per-VM writable rootfs materialization and cleanup, typed egress policy support with iptables default-deny/proxy-only enforcement, API/client schema updates, corrected platform/runtime goals, Vagrant closed-loop tooling, and JSONL decision/reference/validation logs. Local `mage ci` passed; shell/Vagrantfile syntax and Vagrant validation passed where possible. Local Parallels validation cannot execute Firecracker because `/dev/kvm` is not exposed, and enabling Parallels nested virtualization prevents the VM from starting on this host.

Follow-up on 2026-05-30: PR #46 now also includes the explicit API run path required by the operator: GET /capabilities, Mac/Windows API client docs, corrected OpenAPI examples, Vagrant host port forwarding for the guest API, sandbox API comparison, tray/menu-app requirements, and Flowspec artifact path corrections. Local Parallels Vagrant remains KVM-unavailable at /dev/kvm not found.

Follow-up on 2026-05-31: Replacement PR https://github.com/daax-dev/nanofuse/pull/56 contains the real tray/menu app, one-line launch scripts, vagrant-skill validation, post-merge conflict resolution from `origin/main`, and refreshed JSONL validation evidence. Local and Vagrant `mage ci` passed after the merge. The local Apple Silicon Parallels VM remains usable for repo/API/tray validation but not Firecracker execution because `/dev/kvm` is absent.

PR #56 carries the PR #55 Copilot fixes: runtime-capability action gating in the tray app and fail-fast Windows build launcher behavior. The tray app now includes the missing create/start-from-image workflow. The Mac launch point is `scripts/run-tray-macos.sh`; it prints visible status, detects existing tray processes, and supports `--restart`/`--foreground` for predictable operator control. The implementation is `cmd/nanofuse-tray` plus `internal/trayapp`.
<!-- SECTION:FINAL_SUMMARY:END -->
