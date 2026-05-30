# Tasks: Sandbox Objective Closed-Loop Validation

**Input**: `.specify/features/codex-goal/spec.md`, `.specify/features/codex-goal/plan.md`
**Prerequisites**: Backlog task `TASK-47`

## Phase 1: Specification and Governance

- [ ] T001 Create `.specify/features/codex-goal/spec.md`
- [ ] T002 Create `.specify/features/codex-goal/plan.md`
- [ ] T003 Create `.specify/features/codex-goal/tasks.md`
- [ ] T004 Log architecture and validation decisions under `.logs/decisions/`
- [ ] T005 Log primary references under `.logs/references/`

## Phase 2: Persistent Per-VM Filesystem

- [ ] T006 Add failing unit tests for writable rootfs materialization and cleanup in `internal/api/vm_handlers_test.go`
- [ ] T007 Implement per-VM root disk copy and cleanup in `internal/api/vm_handlers.go`
- [ ] T008 Verify image rootfs paths are not mutated across VM creation

## Phase 3: Egress Enforcement

- [ ] T009 Add failing unit tests for default-deny, proxy-only, and cleanup iptables behavior in `internal/network/egress_test.go`
- [ ] T010 Add egress policy types in `internal/types/vm.go`
- [ ] T011 Implement egress policy setup/cleanup in `internal/network/egress.go`
- [ ] T012 Wire egress policy into VM create/delete lifecycle in `internal/api/vm_handlers.go`
- [ ] T013 Update `api/openapi.yaml` for request/response policy schema

## Phase 4: Docs and Vagrant Closed Loop

- [ ] T014 Update `docs/GOALS.md` to match validated current/target behavior
- [ ] T015 Add `docs/building/sandbox-objective-validation.md`
- [ ] T016 Add or update `dev/vagrant/closed-loop.sh` and provider preflight diagnostics
- [ ] T017 Update `dev/vagrant/README.md` with exact Linux/KVM, macOS, and Windows paths

## Phase 5: Validation and PR

- [ ] T018 Run `go fmt ./...`
- [ ] T019 Run targeted Go tests for changed packages
- [ ] T020 Run `mage ci`
- [ ] T021 Run Vagrant closed-loop validation and record output
- [ ] T022 Update `TASK-47` acceptance criteria and final summary
- [ ] T023 Commit, push, and create PR

## Dependencies

- T001-T005 block implementation.
- T006-T008 are independent of T009-T013 after shared types are agreed.
- T014-T017 depend on implementation decisions and validation results.
- T018-T023 run after all code and docs are complete.
