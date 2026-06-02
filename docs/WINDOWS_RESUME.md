# Windows Client Package and Resume

This document is the Windows operator handoff for the current Nanofuse client path.

## Scope

- Windows is a CLI and tray client host.
- Windows is not a validated local Firecracker runtime host in this repo.
- The daemon must run on a reachable Linux or macOS host.

## Package Contents

The first package target is `dist/nanofuse-windows-amd64.zip` with:

- `nanofuse.exe`
- `nanofuse-tray.exe`
- `install-windows.ps1`
- `WINDOWS_RESUME.md`

## Install

From PowerShell in the extracted package directory:

```powershell
.\install-windows.ps1 -ApiUrl "http://127.0.0.1:18080"
```

That installer copies the binaries into `%LOCALAPPDATA%\Nanofuse\bin`, adds that directory to the user `PATH`, and sets user `NANOFUSE_API_URL`.

## Unsigned Package Warning

These binaries are not code signed. Windows SmartScreen or Defender reputation checks may warn before first launch. This package is acceptable for the current task only because MSI, winget, and signing are explicitly out of scope.

## Build From Source

From the repo root on Windows with Go 1.25.x installed:

```powershell
go build -o bin\nanofuse.exe .\cmd\nanofuse
go build -ldflags "-H=windowsgui" -o bin\nanofuse-tray.exe .\cmd\nanofuse-tray
```

From this repo on Linux or macOS, the package can be assembled with:

```bash
GO_BIN=/tmp/go1.25.10/go/bin/go scripts/package-windows.sh
```

## Configure the API Endpoint

Local tunnel example:

```powershell
ssh -N -L 18080:127.0.0.1:8080 user@linux-kvm-host
$env:NANOFUSE_API_URL = "http://127.0.0.1:18080"
```

Direct management network example:

```powershell
$env:NANOFUSE_API_URL = "http://linux-or-mac-runtime-host:8080"
```

When no explicit endpoint is provided, the Windows CLI and tray default to `http://127.0.0.1:18080`.

## Smoke Commands

Use the current CLI surface:

```powershell
.\nanofuse.exe health
Invoke-RestMethod "$env:NANOFUSE_API_URL/capabilities"
.\nanofuse.exe vm list
.\nanofuse.exe vm ports
.\nanofuse-tray.exe --smoke --api-url "$env:NANOFUSE_API_URL"
```

Inspect one VM in more detail:

```powershell
.\nanofuse.exe vm status <vm-id>
```

## Visibility Status

- Ingress ports: visible through `nanofuse.exe vm ports` and `vm status`.
- Egress policy intent: visible through `vm status` or `/vms` JSON under `config.network.egress_policy`.
- Mount visibility: blocker. The current CLI and API do not expose mount metadata as a first-class operator query surface.
- Secret reference visibility: blocker. The current CLI has no `secret` command, and the VM query surface does not expose first-class secret reference inventory.

These blockers must be recorded as blockers, not treated as working features.

## Uninstall

```powershell
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\Nanofuse\bin"
[Environment]::SetEnvironmentVariable("NANOFUSE_API_URL", $null, "User")
```

If `%LOCALAPPDATA%\Nanofuse\bin` was added to the user `PATH`, remove that entry as well.
