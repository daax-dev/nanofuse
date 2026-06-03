param(
    [string]$ApiUrl = "http://127.0.0.1:18080",
    [string]$InstallDir = "$env:LOCALAPPDATA\Nanofuse\bin"
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$candidateRoots = @(
    $scriptDir,
    (Join-Path $scriptDir ".."),
    (Join-Path $scriptDir "..\bin")
)

function Resolve-PackageFile {
    param([string]$Name)

    foreach ($root in $candidateRoots) {
        $candidate = Join-Path $root $Name
        if (Test-Path $candidate) {
            return (Resolve-Path $candidate).Path
        }
    }

    throw "Required package file not found: $Name"
}

$nanofuseExe = Resolve-PackageFile "nanofuse.exe"
$trayExe = Resolve-PackageFile "nanofuse-tray.exe"
$resumeDoc = Resolve-PackageFile "WINDOWS_RESUME.md"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item -Force $nanofuseExe (Join-Path $InstallDir "nanofuse.exe")
Copy-Item -Force $trayExe (Join-Path $InstallDir "nanofuse-tray.exe")
Copy-Item -Force $resumeDoc (Join-Path $InstallDir "WINDOWS_RESUME.md")

# QUICKSTART-WINDOWS.md ships in the package; copy it too when present so the
# "start here" guide stays with the installed client.
$quickstartDoc = $null
try { $quickstartDoc = Resolve-PackageFile "QUICKSTART-WINDOWS.md" } catch { $quickstartDoc = $null }
if ($quickstartDoc) {
    Copy-Item -Force $quickstartDoc (Join-Path $InstallDir "QUICKSTART-WINDOWS.md")
}

[Environment]::SetEnvironmentVariable("NANOFUSE_API_URL", $ApiUrl, "User")

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathEntries = @()
if ($userPath) {
    $pathEntries = $userPath.Split(";") | Where-Object { $_ }
}
if ($pathEntries -notcontains $InstallDir) {
    $newPath = ($pathEntries + $InstallDir) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
}

Write-Host "Nanofuse Windows client installed to $InstallDir"
Write-Host "Default API URL set to $ApiUrl"
Write-Host ""
Write-Host "Unsigned package warning:"
Write-Host "  These binaries are not code signed. Windows SmartScreen may warn before launch."
Write-Host ""
Write-Host "Smoke checks:"
Write-Host "  nanofuse.exe health"
Write-Host "  nanofuse.exe vm list"
Write-Host "  nanofuse.exe vm ports"
Write-Host "  nanofuse-tray.exe --smoke --api-url $ApiUrl"
Write-Host ""
Write-Host "Uninstall:"
Write-Host "  Remove-Item -Recurse -Force '$InstallDir'"
Write-Host "  [Environment]::SetEnvironmentVariable('NANOFUSE_API_URL', \$null, 'User')"
Write-Host "  Remove '$InstallDir' from the user PATH if it was added"
