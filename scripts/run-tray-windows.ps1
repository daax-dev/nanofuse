param(
    [string]$ApiUrl = $env:NANOFUSE_API_URL
)

$ErrorActionPreference = "Stop"

if (-not $ApiUrl) {
    $ApiUrl = "http://127.0.0.1:18080"
}

New-Item -ItemType Directory -Force -Path "bin" | Out-Null
go build -ldflags "-H=windowsgui" -o "bin\nanofuse-tray.exe" ".\cmd\nanofuse-tray"
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
if (-not (Test-Path ".\bin\nanofuse-tray.exe")) {
    throw "go build did not produce .\bin\nanofuse-tray.exe"
}
Start-Process -FilePath ".\bin\nanofuse-tray.exe" -ArgumentList @("--api-url", $ApiUrl)
