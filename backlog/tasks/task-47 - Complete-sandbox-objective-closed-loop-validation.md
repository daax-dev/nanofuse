---
id: TASK-47
title: Complete sandbox objective closed-loop validation
status: Done
assignee:
  - codex
created_date: '2026-05-30 20:05'
updated_date: '2026-05-30 20:33'
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
  - objective.md
  - docs/GOALS.md
  - .claude/workflow.md
  - .claude/CLAUDE.md
  - .claude/language.md
  - .claude/architecture.md
  - .claude/stack.md
  - .claude/sourcecontrol.md
  - .claude/history.md
  - .specify/features/codex-goal/spec.md
  - .specify/features/codex-goal/plan.md
  - .specify/features/codex-goal/tasks.md
  - .specify/features/codex-goal/quickstart.md
  - docs/building/sandbox-objective-validation.md
  - dev/vagrant/closed-loop.sh
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
- [x] #11 Formatter, linter, and tests including mage ci are run, or every blocked gate is recorded with exact cause.
- [x] #12 A pull request is opened with problem statement, approach, alternatives, test evidence, and AI producer/validator information.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Operator authorization: the operator explicitly approved autonomous plan approval on 2026-05-30 and asked execution not to stop while away.

Plan:
1. Create repo-local spec artifacts under .specify/features/codex-goal/ for the sandbox objective: spec.md, plan.md, tasks.md, and a validation quickstart. Keep the spec technology-agnostic; put implementation details in plan/tasks.
2. Correct the documented platform model: Firecracker runtime requires Linux + KVM; macOS and Windows support are host/developer paths through a Linux VM/WSL2/remote Linux runner only when /dev/kvm is exposed. Do not claim native macOS/Windows Firecracker support.
3. Fix per-VM filesystem persistence/isolation by materializing a writable per-VM root disk from the registered image rootfs before first boot, preserving it across stop/start and removing it on VM delete. Add unit coverage for copy behavior and cleanup.
4. Add a first implementation slice for forced egress controls: typed VM network egress policy, iptables command generation/application/cleanup, fail-closed default-deny mode, DNS/proxy-only allowances for future LLM/API/MCP interception, and unit tests using a fake command runner. Integrate cleanup into VM delete/stop paths where policy created host rules.
5. Strengthen identity/secrets posture without introducing a full secret broker in this PR: document that raw secrets are not injected into guest-visible config; preserve existing SPIFFE/vsock path; improve/validate SPIRE command construction where practical; clearly mark remaining Vault/credential-broker work as not yet implemented.
6. Update docs/GOALS.md and directly related docs so goals match validated implementation state and current constraints. Replace unsupported performance/security claims with measurable current/target states.
7. Add/repair Vagrant closed-loop validation under dev/vagrant so host capability checks are explicit. Run the Vagrant workflow from dev/vagrant; on this macOS arm64 host with Parallels, record exact KVM/provider constraints if /dev/kvm cannot be exposed. Do not edit any other repo.
8. Run local gates: go fmt, targeted tests while developing, then mage ci. Run Vagrant closed-loop tests. Record every blocked gate with exact command and failure cause.
9. Log non-trivial decisions in .logs/decisions/*.jsonl and cite primary references in .logs/references/*.jsonl.
10. Commit the branch, push to origin, and open a PR with problem statement, approach, alternatives, test evidence, and AI producer/validator statement. Do not merge.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-05-30: Local `mage ci` passed after implementation. `gosec` is not installed; current mage security target records that and continues.

2026-05-30: Vagrant/Parallels guest boot and rsync succeeded, but provisioning failed at `/dev/kvm not found`. A nested virtualization attempt was also tried; Parallels accepted `--nested-virt on` but the VM then failed to start, so this host cannot execute the Firecracker closed-loop locally.

2026-05-30: Parallels VM created for validation is stopped, not destroyed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Completed sandbox objective validation work on branch `codex-goal` and opened ready-for-review PR https://github.com/daax-dev/nanofuse/pull/46. Implemented per-VM writable rootfs materialization and cleanup, typed egress policy support with iptables default-deny/proxy-only enforcement, API/client schema updates, corrected platform/runtime goals, Vagrant closed-loop tooling, and JSONL decision/reference/validation logs. Local `mage ci` passed; shell/Vagrantfile syntax and Vagrant validation passed where possible. Local Parallels validation is blocked for Firecracker execution because `/dev/kvm` is not exposed, and enabling Parallels nested virtualization prevents the VM from starting on this host.
<!-- SECTION:FINAL_SUMMARY:END -->
