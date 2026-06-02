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

Known-good runtime path:

- macOS M1/M2 uses `runtime.driver=lima_container`.
- Each sandbox is backed by one Lima VM.
- Containers inside that sandbox run through containerd/nerdctl inside the Lima VM.
- Two to three containers for one sandbox should run in the same sandbox VM, not one VM per container.
- Firecracker-on-macOS for M3/M4/M5 is capability-gated and must be tested on supported hardware later.

Current Windows status:

- Windows is currently a client/tray host target, not a validated local runtime host.
- The first acceptable artifact is an unsigned ZIP or PowerShell installer containing Windows binaries and setup instructions.
- MSI, winget, signing, and native Windows local runtime can follow after the client path works.

No remote push has been performed for this handoff state.

## Source Files Already Updated

These files contain the expanded source details, but this `goal.md` file is enough to begin:

- `objective.md`
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
ssh -N -L 18080:127.0.0.1:18080 user@mac-or-linux-runtime-host
```

Build the Windows binaries:

```powershell
go build -o bin\nanofuse.exe .\cmd\nanofuse
go build -ldflags "-H=windowsgui" -o bin\nanofuse-tray.exe .\cmd\nanofuse-tray
```

Run client smoke checks:

```powershell
.\bin\nanofuse.exe health
Invoke-RestMethod "$env:NANOFUSE_API_URL/capabilities"
.\bin\nanofuse.exe sandbox list
.\bin\nanofuse.exe sandbox ports
.\bin\nanofuse.exe secret list
.\bin\nanofuse-tray.exe --smoke --api-url "$env:NANOFUSE_API_URL"
```

If command names differ from current CLI help, run:

```powershell
.\bin\nanofuse.exe --help
.\bin\nanofuse.exe sandbox --help
.\bin\nanofuse.exe secret --help
```

Then adjust the smoke commands to the actual current CLI surface and update this file.

## Required Evidence

Record the following before considering `TASK-53` done:

- Windows version and architecture.
- Go version.
- Exact build commands and output.
- `nanofuse.exe health` output.
- `/capabilities` output.
- `sandbox list` output.
- `sandbox ports` output or exact missing-feature blocker.
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
| Sandbox listing | Must work from Windows through the daemon API. |
| Ingress ports | Must be visible from Windows through `sandbox ports` or API equivalent. |
| Egress policy | Must expose configured intent/status. Lima enforcement is not fail-closed yet. |
| Mounts | Must expose configured mount metadata. |
| Secrets | Must expose secret references only. Actual scoped secret handoff is a future blocker. |
| Easy installer | First target is ZIP or PowerShell installer. MSI/signing can follow. |

## Current Blockers

The broad objective is not complete until these are resolved:

- Windows client packaging and smoke validation have not been run on Windows.
- Native Windows local runtime is not implemented.
- Scoped secret broker/handoff delivery is not implemented.
- Lima egress enforcement is not fail-closed.
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

- daemon started with `runtime.driver=lima_container`
- sandbox created
- two Alpine containers ran in one Lima VM
- ingress mapping worked with nginx on `127.0.0.1:19190`
- sandbox list/status/ports surfaced runtime state
- mount and secret references surfaced as metadata
- API exec worked
- stop/delete cleanup worked
