param(
    [string]$ApiUrl = $env:NANOFUSE_API_URL
)

if (-not $ApiUrl) {
    $ApiUrl = "http://127.0.0.1:18080"
}

New-Item -ItemType Directory -Force -Path "bin" | Out-Null
go build -ldflags "-H=windowsgui" -o "bin\nanofuse-tray.exe" ".\cmd\nanofuse-tray"
Start-Process -FilePath ".\bin\nanofuse-tray.exe" -ArgumentList @("--api-url", $ApiUrl)
