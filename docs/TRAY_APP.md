# Nanofuse Tray App

`nanofuse-tray` is a macOS menu bar and Windows tray API client. It does not run Firecracker, edit Nanofuse storage, shell into the runtime host, or manipulate TAP/KVM directly. Every status check and VM action goes through `nanofused`.

## Requirements

- A reachable `nanofused` API.
- For VM execution, that daemon must run on Linux with read/write `/dev/kvm`.
- macOS or Windows with Go installed to build from this checkout.

The default local API URL is `http://127.0.0.1:18080`, which matches the documented SSH/Vagrant tunnel examples.

## macOS One-Liner

```bash
NANOFUSE_API_URL="${NANOFUSE_API_URL:-http://127.0.0.1:18080}" ./scripts/run-tray-macos.sh
```

Equivalent explicit build and launch:

```bash
go build -o bin/nanofuse-tray ./cmd/nanofuse-tray && ./bin/nanofuse-tray --api-url "${NANOFUSE_API_URL:-http://127.0.0.1:18080}"
```

## Windows One-Liner

Run from PowerShell in the repository root:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-tray-windows.ps1 -ApiUrl "$env:NANOFUSE_API_URL"
```

Equivalent explicit build and launch:

```powershell
if (-not $env:NANOFUSE_API_URL) { $env:NANOFUSE_API_URL = "http://127.0.0.1:18080" }; go build -ldflags "-H=windowsgui" -o bin\nanofuse-tray.exe .\cmd\nanofuse-tray; Start-Process .\bin\nanofuse-tray.exe -ArgumentList @("--api-url", $env:NANOFUSE_API_URL)
```

## Smoke Test

Smoke mode uses the same Nanofuse API client and exits without starting a desktop tray loop:

```bash
./bin/nanofuse-tray --smoke --api-url "${NANOFUSE_API_URL:-http://127.0.0.1:18080}"
```

Windows:

```powershell
.\bin\nanofuse-tray.exe --smoke --api-url $env:NANOFUSE_API_URL
```

## Implemented Controls

The current tray app shows the configured endpoint, health, runtime capability summary, up to 10 VMs, and up to 10 cached images. Selecting a VM enables start, stop, kill, and delete actions through the REST API. Kill and delete require a second click within 10 seconds.

The app uses these endpoints:

- `GET /health`
- `GET /capabilities`
- `GET /vms`
- `GET /images`
- `POST /vms/{id}/start`
- `POST /vms/{id}/stop`
- `POST /vms/{id}/kill`
- `DELETE /vms/{id}`

## Validation

Validated on 2026-05-31 from this Mac:

- `go test ./internal/trayapp ./cmd/nanofuse-tray`
- `go build -o bin/nanofuse-tray ./cmd/nanofuse-tray`
- `GOOS=windows GOARCH=amd64 go build -ldflags='-H=windowsgui' -o /tmp/nanofuse-tray.exe ./cmd/nanofuse-tray`
- `./bin/nanofuse-tray --smoke --api-url http://127.0.0.1:19080` against a local fake Nanofuse API
- bounded macOS tray launch against the same local fake API
