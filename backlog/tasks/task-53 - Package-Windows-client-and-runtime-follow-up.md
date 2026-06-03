---
id: TASK-53
title: 'Package Windows client and record smoke validation'
status: Done
assignee: []
created_date: '2026-06-02 15:20'
updated_date: '2026-06-02 20:30'
labels:
  - codex-goal
  - windows
  - packaging
  - validation
dependencies:
  - TASK-51
references:
  - .flowspec/features/codex-goal/spec.md
  - .flowspec/features/codex-goal/plan.md
  - goal.md
priority: high
---

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A Windows client package exists as either `dist/nanofuse-windows-amd64.zip` or `scripts/install-windows.ps1`, with `nanofuse.exe`, `nanofuse-tray.exe`, setup instructions, unsigned-package warning, and uninstall instructions.
- [x] #2 Windows operator commands are documented and verified against the current CLI/API surface: health, capabilities, VM listing, port visibility, mount visibility or exact blocker, egress policy visibility or exact blocker, and secret reference visibility or exact blocker.
- [x] #3 Evidence is recorded with Windows version, architecture, Go version, exact build commands, exact smoke outputs, tray smoke result, artifact path, and uninstall instructions.
- [x] #4 `goal.md` and related Windows runbooks are updated to match the current command names and actual validation status.
- [x] #5 Local validation includes targeted tests plus the full repo gate `mage ci`, or the exact blocker is recorded if the environment cannot run it.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Audit the current Windows client surface: CLI command names, tray flags, Windows scripts, and docs referenced by `goal.md`.
2. Make the minimum code or script changes needed for a first Windows client package. Prefer ZIP or PowerShell installer packaging over MSI, winget, signing, or local Windows runtime work.
3. Add or update tests for any changed Windows-facing behavior and run targeted test coverage before broad validation.
4. Build Windows client artifacts and assemble the package contents.
5. Run local repo validation (`mage ci`) and any packaging-specific checks available in this environment.
6. If a real Windows session is unavailable, record the exact blocker and reduce the gap as far as possible with buildable artifacts, test coverage, and updated instructions; if a Windows session is available, record full smoke evidence.
7. Update `goal.md`, Windows docs, and the validation record with exact evidence and any remaining blockers.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Validated on 2026-06-02:
- `go version` on the build host: `go version go1.25.10 linux/amd64`.
- Targeted tests passed:
  `PATH=/tmp/go1.25.10/go/bin:$PATH GOCACHE=/tmp/nanofuse-go/cache GOPATH=/tmp/nanofuse-go/path GOMODCACHE=/tmp/nanofuse-go/mod /tmp/go1.25.10/go/bin/go test ./cmd/nanofuse ./internal/trayapp -count=1`.
- Windows CLI cross-build passed:
  `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 /tmp/go1.25.10/go/bin/go build -o /tmp/nanofuse.exe ./cmd/nanofuse`.
- Windows tray cross-build passed:
  `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 /tmp/go1.25.10/go/bin/go build -ldflags='-H=windowsgui' -o /tmp/nanofuse-tray.exe ./cmd/nanofuse-tray`.
- Package assembly passed:
  `GO_BIN=/tmp/go1.25.10/go/bin/go ./scripts/package-windows.sh`.
- Produced artifact:
  `dist/nanofuse-windows-amd64.zip`.
- ZIP contents verified:
  `python3 -m zipfile -l dist/nanofuse-windows-amd64.zip`.
- Full gate passed:
  `PATH=/tmp/go1.25.10/go/bin:/tmp/nanofuse-go/path/bin:$PATH HOME=/tmp/nanofuse-go/home GOCACHE=/tmp/nanofuse-go/cache GOPATH=/tmp/nanofuse-go/path GOMODCACHE=/tmp/nanofuse-go/mod CC='/tmp/zig-x86_64-linux-0.16.0/zig cc' mage ci`.

Closed-loop validation completed on 2026-06-02 (real Windows session + WSL2 Firecracker daemon):
- Client host: Windows 11 Pro `10.0.26200`, AMD64, `go version go1.25.0 windows/amd64`.
- Daemon: real Linux Firecracker `nanofused` in WSL2 (`/dev/kvm` r/w), Firecracker `v1.15.1`, via `scripts/wsl-firecracker-daemon.sh run` on `0.0.0.0:18080`.
- Native Windows builds: `go build -o bin\nanofuse.exe .\cmd\nanofuse` and `go build -ldflags "-H=windowsgui" -o bin\nanofuse-tray.exe .\cmd\nanofuse-tray` (CGO disabled).
- Package: `pwsh scripts/package-windows.ps1` → `dist/nanofuse-windows-amd64.zip` (`nanofuse.exe`, `nanofuse-tray.exe`, `install-windows.ps1`, `WINDOWS_RESUME.md`).
- Windows CLI smoke (endpoint `http://<wsl-ip>:18080`): `health` healthy; `/capabilities` `driver=firecracker native_runtime=true`; `image list` shows `nanofuse-base:latest`.
- Full lifecycle from Windows: `vm run` (with `--mount`, `--secret`, `-p 8081:80`) → running, guest IP `172.16.0.10`; `vm list`/`vm status`/`vm ports`/`vm mounts`/`vm secrets`/`vm logs` correct; `vm stop` → stopped; `vm delete --force` → empty list.
- `nanofuse-tray.exe --smoke` exit 0 with health + capabilities + image JSON.
- `vm exec` on Firecracker returns the documented "Runtime does not support VM exec" (apple_container-only capability).
- Full repo gate: `mage ci` exit 0 ("All CI checks passed!") in WSL2; golangci-lint `v2.12.2` (rebuilt with go1.25); `go test -race ./...` green; `gosec` absent and reported.
- Uninstall: see `docs/WINDOWS_RESUME.md` (remove `%LOCALAPPDATA%\Nanofuse\bin`, clear `NANOFUSE_API_URL`, drop PATH entry).

Mount visibility and secret-reference visibility blockers are RESOLVED: `vm mounts`, `vm secrets`, `--mount`/`--secret`, and `vm status`/`/vms` JSON expose them as first-class surfaces.

Remaining (out of scope for this task / tracked elsewhere):
- Mount runtime enforcement (virtio-fs/block) and scoped secret value delivery on the Firecracker backend.
- Firecracker backend API `vm exec` (apple_container supports it; Firecracker uses SSH).
<!-- SECTION:NOTES:END -->
