# Tasks: Sandbox Objective Closed-Loop Validation

**Input**: `.flowspec/features/codex-goal/spec.md`, `.flowspec/features/codex-goal/plan.md`
**Prerequisites**: Backlog task `TASK-47`

## Phase 1: Specification and Governance

- [x] T001 Create `.flowspec/features/codex-goal/spec.md`
- [x] T002 Create `.flowspec/features/codex-goal/plan.md`
- [x] T003 Create `.flowspec/features/codex-goal/tasks.md`
- [x] T004 Log architecture and validation decisions under `.logs/decisions/`
- [x] T005 Log primary references under `.logs/references/`

## Phase 2: Persistent Per-VM Filesystem

- [x] T006 Add failing unit tests for writable rootfs materialization and cleanup in `internal/api/vm_handlers_test.go`
- [x] T007 Implement per-VM root disk copy and cleanup in `internal/api/vm_handlers.go`
- [x] T008 Verify image rootfs paths are not mutated across VM creation

## Phase 3: Egress Enforcement

- [x] T009 Add failing unit tests for default-deny, proxy-only, and cleanup iptables behavior in `internal/network/egress_test.go`
- [x] T010 Add egress policy types in `internal/types/vm.go`
- [x] T011 Implement egress policy setup/cleanup in `internal/network/egress.go`
- [x] T012 Wire egress policy into VM create/delete lifecycle in `internal/api/vm_handlers.go`
- [x] T013 Update `api/openapi.yaml` for request/response policy schema

## Phase 4: Docs and Vagrant Closed Loop

- [x] T014 Update `docs/GOALS.md` to match validated current/target behavior
- [x] T015 Add `docs/building/sandbox-objective-validation.md`
- [x] T016 Add or update `dev/vagrant/closed-loop.sh` and provider preflight diagnostics
- [x] T017 Update `dev/vagrant/README.md` with exact Linux/KVM, macOS, and Windows paths
- [x] T018 Add API capability reporting and remote client configuration support
- [x] T019 Add Mac/Windows API client runbook and fix API examples
- [x] T020 Add sandbox API comparison and tray/menu app requirements

## Phase 5: Tray/Menu App

- [x] T021 Add testable tray API status/action package under `internal/trayapp`
- [x] T022 Add `cmd/nanofuse-tray` macOS/Windows menu app using only the Nanofuse API
- [x] T023 Add non-GUI tray smoke mode and tests for health, capabilities, VM list, and image list calls
- [x] T024 Add Mac and Windows one-line launch/build instructions for the tray app
- [x] T025 Verify the macOS tray binary builds locally and the smoke mode runs against a test API

## Phase 6: macOS Native Runtime

- [x] T026 Add a runtime manager interface shared by Firecracker and Apple-container backends
- [x] T027 Add Apple-container image resolution, list, start, stop, kill, delete, and log lifecycle support
- [x] T028 Add `runtime.driver=apple_container` config and capability reporting
- [x] T029 Start local macOS `nanofused` through `scripts/run-tray-macos.sh --start-api`
- [x] T030 Validate an API-created macOS VM runs a Linux guest kernel through Apple container and is cleaned up

## Phase 7: Validation and PR

- [x] T031 Run `go fmt ./...`
- [x] T032 Run targeted Go tests for changed packages
- [x] T033 Run `mage ci`
- [x] T034 Run `daax-dev/vagrant-skill` validation and record output
- [x] T035 Update Backlog acceptance criteria and final summaries
- [x] T036 Commit, push, and update PR

## Dependencies

- T001-T005 block implementation.
- T006-T008 are independent of T009-T013 after shared types are agreed.
- T014-T017 depend on implementation decisions and validation results.
- T018-T020 depend on the API boundary decision.
- T021-T025 depend on T018 because the tray app uses the API client boundary.
- T026-T030 depend on T018 and T021 because the macOS runtime is managed through the same API and tray launch path.
- T031-T036 run after all code and docs are complete.
