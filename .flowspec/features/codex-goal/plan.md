# Implementation Plan: Sandbox Objective Closed-Loop Validation

**Branch**: `codex-goal` | **Date**: 2026-05-30 | **Spec**: `spec.md`
**Input**: Feature specification from `.flowspec/features/codex-goal/spec.md`

## Summary

Deliver a truthful, testable slice of the sandbox objective: per-VM writable root disks, typed egress policy enforcement hooks, corrected platform/security documentation, API-driven Mac/Windows client paths, a minimal API-only tray/menu app, and Vagrant closed-loop validation through `daax-dev/vagrant-skill`. The platform remains Firecracker-on-Linux/KVM at runtime; macOS and Windows are operator/developer hosts only when they can reach a Linux/KVM execution environment.

## Technical Context

**Language/Version**: Go 1.24.3
**Primary Dependencies**: Go stdlib, cobra, SQLite, Firecracker process/API integration, iptables, getlantern/systray for the macOS/Windows menu app
**Storage**: Local filesystem under `storage.data_dir` plus SQLite metadata
**Testing**: `go test`, `go test -race`, `mage ci`, tray smoke tests, `daax-dev/vagrant-skill` closed-loop validation
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
.flowspec/features/codex-goal/
├── spec.md
├── plan.md
├── tasks.md
└── quickstart.md

docs/
├── GOALS.md
├── API_QUICK_START.md
├── MAC_WINDOWS_CLIENTS.md
└── building/
    ├── sandbox-objective-validation.md
    ├── sandbox-api-comparison.md
    └── nanofuse-tray-app.md
```

### Source Code

```text
api/openapi.yaml

cmd/nanofuse/
├── main.go
└── main_test.go

cmd/nanofuse-tray/
└── main.go

internal/api/
├── handlers.go
├── server.go
├── server_test.go
├── vm_handlers.go
└── vm_handlers_test.go

internal/client/
├── client.go
├── client_test.go
└── types.go

internal/trayapp/
├── app.go
└── app_test.go

internal/network/
├── egress.go
└── egress_test.go

internal/types/
└── vm.go

dev/vagrant/
└── existing local harness docs, kept secondary to vagrant-skill for this objective
```

**Structure Decision**: VM lifecycle and root disk materialization belong in the daemon API layer because the daemon owns storage and privileged lifecycle. Firewall policy belongs in `internal/network`. Runtime capability reporting belongs in the daemon API because clients and tray apps must gate VM actions without touching host internals. Public request/response shape belongs in `internal/types`, `internal/client`, and `api/openapi.yaml`. Tray status/action logic belongs in a testable `internal/trayapp` package; the OS tray loop belongs only in `cmd/nanofuse-tray`.

## Implementation Notes

1. Add tests for root disk materialization before changing `handleCreateVM`.
2. Implement per-VM rootfs copy with atomic temp-file rename and restrictive permissions.
3. Add tests for egress rule generation and cleanup with a fake runner.
4. Implement egress policy setup/cleanup with idempotent iptables chain operations and explicit proxy-only validation.
5. Wire egress policy into VM create/delete cleanup without changing default legacy behavior unless a policy is requested.
6. Update OpenAPI and docs to expose the new policy shape.
7. Add API capabilities reporting and CLI environment configuration for remote API clients.
8. Use `daax-dev/vagrant-skill` as the required Vagrant harness for this branch; keep `dev/vagrant` as a secondary local harness until it is explicitly replaced.
9. Implement a minimal Go tray/menu app for macOS and Windows with smoke mode, health/capability refresh, VM/image lists, and basic VM lifecycle API actions.
10. Document sandbox API differences, tray/menu app run/build instructions, and current validation evidence.
11. Run targeted tests after each behavior slice, then `mage ci`, then `vagrant-skill` validation.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Platform support documented as constrained rather than native on every OS | Firecracker requires Linux KVM; current host is macOS arm64 without `/dev/kvm` | Claiming native macOS/Windows support would be false and violates validation rules |
| First egress slice uses L3/L4/proxy-only instead of full L7 inspection | The proxy implementation lives outside this repo and is not implemented here | Implementing a full TLS proxy here would duplicate the planned `daax-egress` boundary and expand scope unsafely |
| Tray app uses a minimal Go systray library instead of Electron, Tauri, Wails, Swift, or WinUI | The operator requires a runnable tray/menu API client now on the current branch | Electron/Tauri/Wails add larger repo-wide toolchains; separate native apps create two codebases before the API client behavior is proven |
