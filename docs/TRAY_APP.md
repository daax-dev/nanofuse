# Nanofuse Tray App

`nanofuse-tray` is a macOS menu bar and Windows tray API client. It does not run Firecracker, edit Nanofuse storage, shell into the runtime host, or manipulate TAP/KVM directly. Every status check, image launch, and VM action goes through `nanofused`.

## Requirements

- A reachable `nanofused` API.
- For Linux execution, that daemon uses Firecracker with read/write `/dev/kvm`.
- For local macOS execution in this branch, the one-liner starts the current Apple `container` compatibility driver. The product runtime target remains a Nanofuse-owned Apple Virtualization.framework backend.
- macOS or Windows with Go installed to build from this checkout.

The default local API URL is `http://127.0.0.1:18080`. `NANOFUSE_TRAY_API_URL` overrides `NANOFUSE_API_URL` for the tray launcher.

## macOS One-Liner

```bash
./scripts/run-tray-macos.sh --start-api --restart
```

The script builds `bin/nanofuse-tray` and `bin/nanofused`, writes a local macOS daemon config under `${NANOFUSE_DATA_DIR:-/tmp/nanofuse-macos}`, starts `nanofused` through launchd with `runtime.driver=apple_container`, then starts the menu bar app through launchd label `com.daax.nanofuse.tray`. That daemon path is a compatibility runtime for this tray/API PR, not the final native Apple VZ backend. Daemon logs go to `${NANOFUSE_API_LOG:-/tmp/nanofused-macos.log}`. Tray logs go to `${NANOFUSE_TRAY_LOG:-/tmp/nanofuse-tray.log}`. Stop the tray with `launchctl bootout gui/$(id -u)/com.daax.nanofuse.tray`.

Useful variants:

```bash
./scripts/run-tray-macos.sh --restart
./scripts/run-tray-macos.sh --start-api --smoke --timeout 5s
./scripts/run-tray-macos.sh --start-api --launch-image docker.io/library/alpine:3.20 --timeout 30s
./scripts/run-tray-macos.sh --foreground
./scripts/run-tray-macos.sh --smoke --timeout 2s
./scripts/run-tray-macos.sh --api-url http://127.0.0.1:18080 --restart
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
./scripts/run-tray-macos.sh --smoke --timeout 2s --api-url "${NANOFUSE_API_URL:-http://127.0.0.1:18080}"
```

Headless launch mode uses the same create/start helper as the menu item and exits after printing the created VM:

```bash
./scripts/run-tray-macos.sh --start-api --launch-image docker.io/library/alpine:3.20 --timeout 30s
```

Windows:

```powershell
.\bin\nanofuse-tray.exe --smoke --api-url $env:NANOFUSE_API_URL
```

## Implemented Controls

The current tray app shows the configured endpoint, health, runtime capability summary, up to 25 VMs, and up to 25 cached images. VM rows include a readable VM name, state, image leaf name, and configured port forwards.

Use `New MicroVM From Container...` to enter an OCI image reference and launch it directly. Use `Add Image to List...` to start an API image pull/resolution job, then refresh the tray. Selecting a cached image enables `Launch Selected Cached Image`, which creates and starts a VM from that image through the API. Tray-created VMs use short generated names such as `nf-alpine-3-20-a1b2c3`.

Tray-created VMs request `host_port: 0` for guest `8080`. `nanofused` allocates the concrete localhost port on the daemon host and returns it in `/vms`, the root status page, the tray row, and `nanofuse vm ports`. Multiple tray launches can therefore coexist while still exposing a host-reachable service port.

Each VM row has its own submenu actions: `Start`, `Stop`, `Kill`, and `Delete`. Actions are enabled only when valid for that VM state: start for `created` or `stopped`, stop for `running` or `paused`, kill for active runtime handles, and delete for known VM rows. VM actions stay disabled when the daemon is unreachable or `/capabilities` reports `native_runtime=false`. Kill and delete require a second click within 10 seconds on the same VM row.

The app uses these endpoints:

- `GET /health`
- `GET /capabilities`
- `GET /vms`
- `GET /images`
- `POST /images/pull`
- `POST /vms`
- `POST /vms/{id}/exec`
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
- `bash -n scripts/run-tray-macos.sh`
- `./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 5s`
- `./scripts/run-tray-macos.sh --help`
- API lifecycle on macOS Apple `container` compatibility runtime:
  `POST /vms` with `alpine:3.20`, `POST /vms/{id}/start`, `nanofuse vm ports`, `nanofuse vm exec <vm> -- uname -a`, `POST /vms/{id}/stop`, `DELETE /vms/{id}`. The guest reported Linux `6.12.28` on `aarch64`; no test VM metadata remained after delete.
