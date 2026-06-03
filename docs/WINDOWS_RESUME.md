# Windows Client Package and Resume

This document is the Windows operator handoff for the current Nanofuse client path.

## Scope

- Windows is a CLI and tray client host.
- Windows is not a local Firecracker runtime host: `nanofused` rejects a Windows host OS by design.
- The daemon must run on a reachable Linux or macOS host. On a Windows workstation, WSL2 (which exposes `/dev/kvm`) is a supported local Linux backend — see "Closed-Loop Backend (WSL2 Firecracker)" below.

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
.\nanofuse.exe vm mounts
.\nanofuse.exe vm secrets
.\nanofuse-tray.exe --smoke --api-url "$env:NANOFUSE_API_URL"
```

Full lifecycle, including the mount and secret-reference surfaces:

```powershell
.\nanofuse.exe vm run nanofuse-base:latest demo --memory 512 --vcpus 1 `
  -p 8081:80 `
  --mount "src=/srv/data,dst=/data,type=bind,ro" `
  --mount "type=tmpfs,dst=/scratch" `
  --secret "name=API_TOKEN,source=vault://kv/token" `
  --secret "name=tls,type=file,target=/etc/tls/key.pem,source=spire://"
.\nanofuse.exe vm status demo
.\nanofuse.exe vm mounts demo
.\nanofuse.exe vm secrets demo
.\nanofuse.exe vm logs demo --tail 20
.\nanofuse.exe vm stop demo
.\nanofuse.exe vm delete demo --force
```

Inspect one VM in more detail:

```powershell
.\nanofuse.exe vm status <vm-id>
```

## Visibility Status

- Ingress ports: visible through `nanofuse.exe vm ports` and `vm status`.
- Egress policy intent: visible through `vm status` or `/vms` JSON under `config.network.egress_policy`.
- Mount visibility: working. Declared with `--mount` on `vm create`/`vm run`, queried with `nanofuse.exe vm mounts`, and shown in `vm status` and `/vms` JSON under `config.mounts`.
- Secret reference visibility: working. Declared with `--secret` on `vm create`/`vm run`, queried with `nanofuse.exe vm secrets`, and shown in `vm status` and `/vms` JSON under `config.secrets`. Secret references never carry values — only the name, source reference, delivery type, and in-guest target.

### Backend capability note

`nanofuse.exe vm exec` is an apple_container (macOS) runtime capability. The
Firecracker backend is a bare microVM and returns a clear "Runtime does not
support VM exec" error; in-guest command execution on Firecracker is via SSH to
the guest. The CLI command surface is identical across platforms; only the
daemon backend differs.

Mount and secret references are operator-visibility/inventory surfaces:
declared, validated, persisted, and queryable on every backend. Runtime
enforcement (virtio-fs/block attachment, scoped secret value delivery) is
performed by the backend where supported and remains tracked runtime work.

## Closed-Loop Backend (WSL2 Firecracker)

On a Windows workstation with WSL2, a real Linux Firecracker `nanofused` can run
inside WSL2 (which exposes `/dev/kvm`) and serve the Windows client over TCP.
This gives a full closed loop on a single machine, equivalent to the macOS
`apple_container` path.

Bring-up (run inside WSL2 as root — `wsl -d Ubuntu -u root`):

```bash
cd /mnt/c/path/to/nanofuse
bash scripts/wsl-firecracker-daemon.sh setup   # Go + Firecracker + CI fixtures + build + register image
bash scripts/wsl-firecracker-daemon.sh run     # network + daemon on 0.0.0.0:18080
```

From Windows, point the client at the WSL endpoint. WSL2 mirrored-localhost is
not always available, so the WSL IP (the "direct management network" pattern)
is the reliable default:

```powershell
$wsl = (wsl hostname -I).Trim().Split(" ")[0]
$env:NANOFUSE_API_URL = "http://$wsl`:18080"
.\nanofuse.exe health
```

To keep the documented `http://127.0.0.1:18080` default, add a port proxy
(elevated PowerShell) or an SSH tunnel:

```powershell
netsh interface portproxy add v4tov4 listenaddress=127.0.0.1 listenport=18080 connectaddress=$wsl connectport=18080
```

## Uninstall

```powershell
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\Nanofuse\bin"
[Environment]::SetEnvironmentVariable("NANOFUSE_API_URL", $null, "User")
```

If `%LOCALAPPDATA%\Nanofuse\bin` was added to the user `PATH`, remove that entry as well.
