# Nanofuse Resume Goal

Use this file as the single handoff document for resuming the Nanofuse objective on a Windows machine.

Slash-command form:

```text
/goal Complete this entirely from goal.md in the Nanofuse repo root. Treat goal.md as the authoritative resume brief and update it when material status changes.
```

## Active Goal

Complete the Windows resume leg of the Nanofuse sandbox objective:

- Package and validate a Windows operator/client path.
- Install or run `nanofuse.exe` and `nanofuse-tray.exe` on Windows.
- Configure the Windows client to reach a running `nanofused` daemon.
- Prove Windows can list and manage sandboxes through the API.
- Prove the operator can inspect sandbox status, open ingress ports, egress policy intent, mounts, and secret references.
- Record exact smoke-test evidence or the exact blocker.

Canonical next backlog task: `TASK-53`.

## Current Status

Completed prerequisite: `TASK-52`.

Known-good compatibility runtime path:

- macOS local validation currently uses `runtime.driver=apple_container`.
- API-created VMs run behind Apple `container` plus Virtualization.framework on supported Apple silicon.
- `nanofuse vm ports` and `nanofuse vm exec` are validated on the macOS compatibility path.
- Firecracker-on-macOS remains a future backend task, not the current validated path.

Current Windows status:

- Windows is currently a client/tray host target, not a validated local runtime host.
- `dist/nanofuse-windows-amd64.zip` and `scripts/install-windows.ps1` now exist as the first package slice.
- The Windows CLI and tray now default to `http://127.0.0.1:18080` when no endpoint is provided.
- Real Windows smoke execution is still blocked on the absence of a Windows desktop session in this workspace.
- MSI, winget, signing, and native Windows local runtime can follow after the client path works.

No remote push has been performed for this handoff state.

## Source Files Already Updated

These files contain the expanded source details, but this `goal.md` file is enough to begin:

- `docs/WINDOWS_RESUME.md`
- `docs/GOALS.md`
- `backlog/tasks/task-53 - Package-Windows-client-and-runtime-follow-up.md`
- `docs/building/sandbox-objective-validation.md`

## Start On Windows

Clone or open the Nanofuse repo on Windows.

Set the API URL to a reachable daemon:

```powershell
$env:NANOFUSE_API_URL = "http://127.0.0.1:18080"
```

If the daemon is running on a Mac or Linux host and only listens locally, create a tunnel:

```powershell
ssh -N -L 18080:127.0.0.1:8080 user@mac-or-linux-runtime-host
```

Use the packaged installer or build the Windows binaries:

```powershell
.\install-windows.ps1 -ApiUrl "http://127.0.0.1:18080"

go build -o bin\nanofuse.exe .\cmd\nanofuse
go build -ldflags "-H=windowsgui" -o bin\nanofuse-tray.exe .\cmd\nanofuse-tray
```

Run client smoke checks:

```powershell
.\bin\nanofuse.exe health
Invoke-RestMethod "$env:NANOFUSE_API_URL/capabilities"
\bin\nanofuse.exe vm list
\bin\nanofuse.exe vm ports
\bin\nanofuse.exe vm status <vm-id>
.\bin\nanofuse-tray.exe --smoke --api-url "$env:NANOFUSE_API_URL"
```

Current blockers to record during Windows smoke:

```powershell
\bin\nanofuse.exe vm status <vm-id>
Invoke-RestMethod "$env:NANOFUSE_API_URL/vms"
```

Mount visibility is not exposed as a first-class CLI/API query surface today. Secret reference visibility is not exposed as a first-class Windows CLI surface today. Record both as blockers unless the repo changes.

## Required Evidence

Record the following before considering `TASK-53` done:

- Windows version and architecture.
- Go version.
- Exact build commands and output.
- `nanofuse.exe health` output.
- `/capabilities` output.
- `vm list` output.
- `vm ports` output or exact missing-feature blocker.
- mount visibility output or exact missing-feature blocker.
- egress policy visibility output or exact missing-feature blocker.
- secret reference visibility output or exact missing-feature blocker.
- tray smoke result.
- installer or ZIP artifact path.
- uninstall instructions.

Store evidence in either:

- `backlog/tasks/task-53 - Package-Windows-client-and-runtime-follow-up.md`
- `docs/building/sandbox-objective-validation.md`

## Packaging Target

First Windows package can be one of:

- `dist/nanofuse-windows-amd64.zip`
- `scripts/install-windows.ps1`

Minimum package contents:

- `nanofuse.exe`
- `nanofuse-tray.exe`
- default API profile setup instructions
- clear unsigned-package warning
- uninstall instructions

Do not block initial completion on MSI, winget, code signing, notarization equivalent, or native Windows local runtime unless explicitly required.

## Objective Mapping

The Windows work maps to the larger objective as follows:

| Objective area | Current mapping |
| --- | --- |
| Linux, Windows, Mac support | Windows client packaging and smoke validation closes the operator/client part. |
| MicroVM isolation | Runtime daemon remains on validated macOS Lima or Linux Firecracker host. Windows does not need to host local microVMs for TASK-53. |
| Container wrapping | One sandbox VM runs its guest containers through containerd/nerdctl. |
| Sandbox listing | Must work from Windows through the daemon API via `nanofuse vm list` or `/vms`. |
| Ingress ports | Must be visible from Windows through `nanofuse vm ports` or API equivalent. |
| Egress policy | Current intent is visible through VM status or `/vms` JSON. Enforcement is still not fail-closed on the macOS compatibility path. |
| Mounts | Current Windows operator path has no first-class mount metadata query surface. Record this as a blocker. |
| Secrets | Current Windows operator path has no first-class secret reference query surface. Record this as a blocker. |
| Easy installer | First target is ZIP or PowerShell installer. MSI/signing can follow. |

## Current Blockers

The broad objective is not complete until these are resolved:

- Full Windows smoke validation has not been run on a Windows desktop session from this workspace.
- Native Windows local runtime is not implemented.
- Scoped secret broker/handoff delivery is not implemented.
- Mount metadata is not exposed as a first-class Windows operator query surface.
- Secret reference inventory is not exposed as a first-class Windows operator query surface.
- The macOS compatibility path egress implementation is not fail-closed.
- M3/M4/M5 Firecracker-on-macOS path is unvalidated on supported Apple Silicon hardware.
- Linux Firecracker jailer is not yet the default hardened launch path.

## Do Not Do Yet

- Do not push or open a PR until local validation is complete.
- Do not treat Windows as a local runtime host for `TASK-53`.
- Do not require Firecracker-on-macOS validation for `TASK-53`.
- Do not commit secrets, tokens, keys, or live `.env` files.
- Do not bypass repo PR immutability guardrails.

## Validation Already Completed Before Windows Handoff

Local macOS validation completed before this handoff:

- `git diff --check` passed.
- `mage ci` passed.
- `mage ci` reported `gosec` was not installed; that check is non-fatal in the current mage target.

Earlier closed-loop runtime validation completed for the macOS Lima path:

- daemon started with `runtime.driver=apple_container`
- VM created and started from `alpine:3.20`
- ingress mapping worked and surfaced through `nanofuse vm ports`
- `nanofuse vm list`, `nanofuse vm status`, and `nanofuse vm exec` surfaced runtime state
- API exec worked
- stop/delete cleanup worked

Additional packaging validation completed on 2026-06-02 from a Linux amd64 workspace:

- Go toolchain used: `go version go1.25.10 linux/amd64`
- Windows binaries cross-built successfully with `GOOS=windows GOARCH=amd64 CGO_ENABLED=0`
- `dist/nanofuse-windows-amd64.zip` was created with `nanofuse.exe`, `nanofuse-tray.exe`, `install-windows.ps1`, and `WINDOWS_RESUME.md`
- `mage ci` passed locally using Zig as the CGO compiler and a writable temp HOME/cache path
- Real Windows command output, tray smoke, Windows version, and Windows architecture remain blocked pending an actual Windows session
