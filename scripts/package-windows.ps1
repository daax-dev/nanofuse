<#
.SYNOPSIS
  Build the Windows nanofuse client package natively on Windows.

.DESCRIPTION
  Produces dist/nanofuse-windows-amd64.zip containing nanofuse.exe,
  nanofuse-tray.exe, install-windows.ps1, and WINDOWS_RESUME.md.

  This is the Windows-native counterpart to scripts/package-windows.sh
  (which cross-builds from Linux/macOS). Run from the repo root:

    pwsh scripts/package-windows.ps1
    # or
    powershell -ExecutionPolicy Bypass -File scripts\package-windows.ps1
#>
[CmdletBinding()]
param(
    [string]$OutDir = "dist"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

$PackageName = "nanofuse-windows-amd64"
$StageDir = Join-Path $OutDir $PackageName
$ArchivePath = Join-Path $OutDir "$PackageName.zip"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "go toolchain not found on PATH"
}

if (Test-Path $StageDir) { Remove-Item -Recurse -Force $StageDir }
if (Test-Path $ArchivePath) { Remove-Item -Force $ArchivePath }
New-Item -ItemType Directory -Force -Path $StageDir | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "Building Windows CLI..."
go build -o (Join-Path $StageDir "nanofuse.exe") ./cmd/nanofuse
if ($LASTEXITCODE -ne 0) { throw "nanofuse.exe build failed" }

Write-Host "Building Windows tray..."
go build -ldflags "-H=windowsgui" -o (Join-Path $StageDir "nanofuse-tray.exe") ./cmd/nanofuse-tray
if ($LASTEXITCODE -ne 0) { throw "nanofuse-tray.exe build failed" }

Copy-Item scripts\install-windows.ps1 (Join-Path $StageDir "install-windows.ps1")
Copy-Item docs\WINDOWS_RESUME.md (Join-Path $StageDir "WINDOWS_RESUME.md")

Write-Host "Creating $ArchivePath..."
Compress-Archive -Path $StageDir -DestinationPath $ArchivePath -Force

Write-Host "Created $ArchivePath"
Get-ChildItem $StageDir | Select-Object Name, Length | Format-Table -AutoSize
