# Implementation Plan: Sandbox Objective Closed-Loop Validation

**Branch**: `codex-goal` | **Date**: 2026-05-30 | **Spec**: `spec.md`
**Input**: Feature specification from `.specify/features/codex-goal/spec.md`

## Summary

Deliver a truthful, testable slice of the sandbox objective: per-VM writable root disks, typed egress policy enforcement hooks, corrected platform/security documentation, and Vagrant closed-loop validation. The platform remains Firecracker-on-Linux/KVM at runtime; macOS and Windows are operator/developer hosts only when they can reach a Linux/KVM execution environment.

## Technical Context

**Language/Version**: Go 1.24.3
**Primary Dependencies**: Go stdlib, cobra, SQLite, Firecracker process/API integration, iptables
**Storage**: Local filesystem under `storage.data_dir` plus SQLite metadata
**Testing**: `go test`, `go test -race`, `mage ci`, Vagrant closed-loop validation
**Target Platform**: Linux with KVM for runtime; macOS/Windows through compatible Linux/KVM guest or remote runner
**Project Type**: Single Go CLI/daemon project
**Performance Goals**: Preserve current VM lifecycle path; no new synchronous network calls in VM start path
**Constraints**: No secrets in repo; no edits outside this checkout; no direct commit to `main`; document unvalidated claims
**Scale/Scope**: One focused PR covering current-branch objective validation and first security/lifecycle enforcement slice

## Constitution Check

- Spec exists before implementation: PASS.
- Plan exists before implementation: PASS.
- Backlog task exists and is in progress: PASS (`TASK-47`).
- Tests before implementation: REQUIRED for persistent root disk and egress policy behavior.
- Decision logging: REQUIRED for platform model, root disk materialization, and egress enforcement scope.
- Human approval: operator explicitly approved autonomous plan approval on 2026-05-30.

## Project Structure

### Documentation

```text
.specify/features/codex-goal/
├── spec.md
├── plan.md
├── tasks.md
└── quickstart.md

docs/
├── GOALS.md
└── building/
    └── sandbox-objective-validation.md
```

### Source Code

```text
internal/api/
├── vm_handlers.go
└── vm_handlers_test.go

internal/network/
├── egress.go
└── egress_test.go

internal/types/
└── vm.go

dev/vagrant/
├── README.md
├── Vagrantfile
├── setup.sh
├── verify.sh
└── closed-loop.sh
```

**Structure Decision**: VM lifecycle and root disk materialization belong in the daemon API layer because the daemon owns storage and privileged lifecycle. Firewall policy belongs in `internal/network`. Public request/response shape belongs in `internal/types` and `api/openapi.yaml`.

## Implementation Notes

1. Add tests for root disk materialization before changing `handleCreateVM`.
2. Implement per-VM rootfs copy with atomic temp-file rename and restrictive permissions.
3. Add tests for egress rule generation and cleanup with a fake runner.
4. Implement egress policy setup/cleanup with idempotent iptables chain operations and explicit proxy-only validation.
5. Wire egress policy into VM create/delete cleanup without changing default legacy behavior unless a policy is requested.
6. Update OpenAPI and docs to expose the new policy shape.
7. Improve Vagrant validation so unsupported providers fail with exact diagnostics instead of ambiguous provisioning errors.
8. Run targeted tests after each behavior slice, then `mage ci`, then Vagrant.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Platform support documented as constrained rather than native on every OS | Firecracker requires Linux KVM; current host is macOS arm64 without `/dev/kvm` | Claiming native macOS/Windows support would be false and violates validation rules |
| First egress slice uses L3/L4/proxy-only instead of full L7 inspection | The proxy implementation lives outside this repo and is not implemented here | Implementing a full TLS proxy here would duplicate the planned `daax-egress` boundary and expand scope unsafely |
