<#
.SYNOPSIS
  One-command start for the Nanofuse Windows client + WSL2 Firecracker daemon.

.DESCRIPTION
  This is the easy button. From the repo root:

      powershell -ExecutionPolicy Bypass -File scripts\start-windows.ps1

  It will:
    1. Build nanofuse.exe / nanofuse-tray.exe if missing.
    2. Make sure a Linux Firecracker nanofused daemon is running in WSL2
       (running first-time setup if needed, then starting it in its own window).
    3. Point this session's NANOFUSE_API_URL at the WSL daemon.
    4. Launch the tray and print the CLI commands you can run next.

.PARAMETER NoTray
  Skip launching the tray; just start/verify the daemon and set the endpoint.

.PARAMETER Distro
  WSL distro to use (default: Ubuntu).
#>
[CmdletBinding()]
param(
    [switch]$NoTray,
    [string]$Distro = "Ubuntu"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

function Info($m) { Write-Host "[start] $m" -ForegroundColor Cyan }
function Warn($m) { Write-Host "[start] $m" -ForegroundColor Yellow }

# --- 1. Build client binaries if missing -----------------------------------
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go toolchain not found on PATH. Install Go 1.25.x, then re-run."
}
if (-not (Test-Path bin\nanofuse.exe)) {
    Info "Building nanofuse.exe ..."
    $env:CGO_ENABLED = "0"; go build -o bin\nanofuse.exe .\cmd\nanofuse
    if ($LASTEXITCODE -ne 0) { throw "nanofuse.exe build failed" }
}
if (-not (Test-Path bin\nanofuse-tray.exe)) {
    Info "Building nanofuse-tray.exe ..."
    $env:CGO_ENABLED = "0"; go build -ldflags "-H=windowsgui" -o bin\nanofuse-tray.exe .\cmd\nanofuse-tray
    if ($LASTEXITCODE -ne 0) { throw "nanofuse-tray.exe build failed" }
}

# --- 2. Ensure the WSL2 Firecracker daemon is running -----------------------
$wslRepo = (wsl -d $Distro -u root -e wslpath -u "$RepoRoot").Trim()
$wslIp = ((wsl -d $Distro -u root -e hostname -I).Trim() -split '\s+')[0]
if (-not $wslIp) { throw "Could not determine the WSL IP for distro '$Distro'." }
$endpoint = "http://${wslIp}:18080"

function Test-Daemon { try { $null = Invoke-RestMethod "$endpoint/health" -TimeoutSec 3; return $true } catch { return $false } }

if (Test-Daemon) {
    Info "Daemon already healthy at $endpoint"
} else {
    $haveDaemon = (wsl -d $Distro -u root -e bash -lc "test -x '$wslRepo/bin/nanofused' && echo yes || echo no").Trim()
    if ($haveDaemon -ne "yes") {
        Warn "First-time WSL setup (Go + Firecracker + image, several minutes)..."
        wsl -d $Distro -u root -e bash -lc "cd '$wslRepo' && sed -i 's/\r`$//' scripts/wsl-firecracker-daemon.sh && bash scripts/wsl-firecracker-daemon.sh setup"
        if ($LASTEXITCODE -ne 0) { throw "WSL daemon setup failed; see output above." }
    }
    Info "Starting nanofused in a new WSL window ..."
    Start-Process wsl.exe -ArgumentList @(
        "-d", $Distro, "-u", "root", "-e", "bash", "-lc",
        "cd '$wslRepo' && bash scripts/wsl-firecracker-daemon.sh run"
    )
    Info "Waiting for the daemon to come up ..."
    for ($i = 0; $i -lt 30; $i++) {
        Start-Sleep -Seconds 2
        if (Test-Daemon) { break }
    }
    if (-not (Test-Daemon)) { throw "Daemon did not become healthy at $endpoint. Check the WSL window for errors." }
    Info "Daemon healthy at $endpoint"
}

# --- 3. Configure the endpoint for this session -----------------------------
$env:NANOFUSE_API_URL = $endpoint
Info "NANOFUSE_API_URL = $endpoint (this session)"
Info "Tip: persist it with  [Environment]::SetEnvironmentVariable('NANOFUSE_API_URL','$endpoint','User')"

# --- 4. Launch the tray + print next steps ----------------------------------
if (-not $NoTray) {
    Info "Launching the tray ..."
    Start-Process -FilePath (Resolve-Path bin\nanofuse-tray.exe) -ArgumentList "--api-url", $endpoint
}

Write-Host ""
Write-Host "Ready. Try:" -ForegroundColor Green
Write-Host "  .\bin\nanofuse.exe health"
Write-Host "  .\bin\nanofuse.exe vm run nanofuse-base:latest demo --memory 512 --vcpus 1"
Write-Host "  .\bin\nanofuse.exe vm list"
Write-Host "  .\bin\nanofuse.exe vm exec demo -- uname -a"
Write-Host "  .\bin\nanofuse.exe vm delete demo --force"
