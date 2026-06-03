# Nanofuse on Windows — Quick Start

Run untrusted code in Firecracker microVMs from Windows. Windows is the
**client** (CLI + tray); the microVMs run in a Linux **daemon**. On a Windows
workstation that daemon runs inside WSL2 (which provides `/dev/kvm`).

## TL;DR — one command

From the repo root in PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\start-windows.ps1
```

That builds the binaries (if needed), starts the Firecracker daemon in WSL2
(running first-time setup the first time — a few minutes), points the session
at it, and launches the tray. When it finishes it prints the CLI commands to try.

> Prerequisites: WSL2 with an `Ubuntu` distro, virtualization enabled in the
> BIOS and in WSL (`/dev/kvm` present), and Go 1.25.x on `PATH`.

## What you get

- `nanofuse.exe` — the CLI.
- `nanofuse-tray.exe` — a notification-area tray (blue hexagon icon) to watch
  the daemon and start/stop VMs.
- A real Firecracker daemon in WSL2 serving the API on `:18080`.

## Manual steps (if you prefer)

### 1. Start the daemon (inside WSL, as root)

```powershell
wsl -d Ubuntu -u root -e bash -lc "cd /mnt/c/path/to/nanofuse && bash scripts/wsl-firecracker-daemon.sh setup"   # first time only
wsl -d Ubuntu -u root -e bash -lc "cd /mnt/c/path/to/nanofuse && bash scripts/wsl-firecracker-daemon.sh run"     # leave running
```

### 2. Point the client at the daemon

WSL2 localhost mirroring is not always on, so use the WSL IP:

```powershell
$wsl = (wsl hostname -I).Trim().Split(" ")[0]
$env:NANOFUSE_API_URL = "http://$wsl`:18080"
```

(Persist it for all sessions: `[Environment]::SetEnvironmentVariable("NANOFUSE_API_URL", $env:NANOFUSE_API_URL, "User")`.)

### 3. Use it

```powershell
.\bin\nanofuse.exe health
.\bin\nanofuse.exe vm run nanofuse-base:latest demo --memory 512 --vcpus 1 -p 8080:80
.\bin\nanofuse.exe vm list
.\bin\nanofuse.exe vm exec demo -- uname -a
.\bin\nanofuse.exe vm mounts demo
.\bin\nanofuse.exe vm secrets demo
.\bin\nanofuse.exe vm stop demo
.\bin\nanofuse.exe vm delete demo --force
```

Declare mounts and secret references at launch:

```powershell
.\bin\nanofuse.exe vm run nanofuse-base:latest app `
  --mount "src=/srv/data,dst=/data,type=bind,ro" `
  --secret "name=API_TOKEN,source=vault://kv/token"
```

## Tray

Launch `nanofuse-tray.exe` (or let `start-windows.ps1` do it). The tray shows
daemon health and runtime status, lists VMs and cached images, and can launch,
start, stop, kill, and delete VMs through the API. A headless self-check:

```powershell
.\bin\nanofuse-tray.exe --smoke --api-url "$env:NANOFUSE_API_URL"
```

## Connecting to a remote daemon instead

If `nanofused` runs on a separate Linux/macOS host, skip WSL and just set the
endpoint (direct, or via an SSH tunnel):

```powershell
$env:NANOFUSE_API_URL = "http://linux-or-mac-host:8080"
# or:  ssh -N -L 18080:127.0.0.1:8080 user@host   ; then use http://127.0.0.1:18080
```

## More

- Operator handoff and validation details: [`WINDOWS_RESUME.md`](WINDOWS_RESUME.md).
- Packaging a ZIP for distribution: `scripts\package-windows.ps1`.
- Uninstall: remove `%LOCALAPPDATA%\Nanofuse\bin`, clear the `NANOFUSE_API_URL`
  user variable, and drop that bin directory from your user `PATH`.
