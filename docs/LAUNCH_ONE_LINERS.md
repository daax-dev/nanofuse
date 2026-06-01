# Nanofuse Launch One-Liners

Nanofuse now has two local runtime paths:

- Linux: `nanofused` runs Firecracker on Linux/KVM.
- macOS: `nanofused` can run OCI images through Apple's `container` CLI, which uses Virtualization.framework on Apple silicon.

Windows is currently a client/tray host. It manages a reachable `nanofused` API; it is not a local runtime host in this repo.

`nanofuse-tray` is a macOS menu bar and Windows tray API client. On macOS, `scripts/run-tray-macos.sh --start-api` starts a local Apple-container-backed daemon through launchd, then starts the menu bar app.

## Tested Status

| Path | Status |
|------|--------|
| macOS local Apple-container daemon | Tested on this Mac with `runtime.driver=apple_container`; `/capabilities` reports `native_runtime=true`. |
| macOS API-created VM from OCI image | Tested with `alpine:3.20`; `container exec` inside the API-created VM returned Linux `6.12.28` on `aarch64`. |
| macOS tray app against local daemon | Tested with `./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 5s`. |
| Linux/KVM direct Firecracker boot | Previously tested on a Linux/KVM host with Firecracker `1.7.0`, kernel `6.1.77`, and Ubuntu rootfs. |
| Linux/KVM rootless API daemon with no VM networking | Tested with `network.setup=false` and VM `network=none`. |
| Vagrant Linux validation | `daax-dev/vagrant-skill` `mage ci` passes in the local Parallels Ubuntu guest; that guest does not expose `/dev/kvm`, so it is not claimed as a Firecracker boot host. |
| Windows tray executable build | Cross-built from this Mac with `GOOS=windows GOARCH=amd64` and `-H=windowsgui`; runtime click testing still requires a Windows desktop session. |
| Windows one-liner | Syntax only in this workspace. |

## macOS Local Runtime and Tray

One-line local startup from the repo root:

```bash
./scripts/run-tray-macos.sh --start-api --restart
```

That command builds `bin/nanofused` and `bin/nanofuse-tray`, writes a macOS daemon config under `${NANOFUSE_DATA_DIR:-/tmp/nanofuse-macos}`, starts `nanofused` through launchd with `runtime.driver=apple_container`, then starts the menu bar app through launchd label `com.daax.nanofuse.tray`. Daemon logs go to `${NANOFUSE_API_LOG:-/tmp/nanofused-macos.log}`. Tray logs go to `${NANOFUSE_TRAY_LOG:-/tmp/nanofuse-tray.log}`. Stop the tray with `launchctl bootout gui/$(id -u)/com.daax.nanofuse.tray`.

Smoke test without opening the menu bar UI:

```bash
./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 5s
```

Launch one VM through the same tray create/start path without opening the menu:

```bash
./scripts/run-tray-macos.sh --start-api --launch-image docker.io/library/alpine:3.20 --timeout 30s
```

Confirm API readiness:

```bash
open http://127.0.0.1:18080/
curl http://127.0.0.1:18080/health
curl http://127.0.0.1:18080/capabilities
```

The root URL is a browser status page. It shows runtime readiness, VMs, images, and host-to-VM port forwards.

Create, start, inspect, stop, and delete a local macOS-backed Linux VM:

```bash
API=http://127.0.0.1:18080
VM_NAME="mac-api-alpine-$(date +%s)"
VM_ID="$(curl -fsS -X POST "$API/vms" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${VM_NAME}\",\"image\":\"alpine:3.20\",\"config\":{\"vcpus\":1,\"memory_mib\":256}}" \
  | jq -r '.id')"
curl -fsS -X POST "$API/vms/$VM_ID/start" | jq '.runtime'
curl -fsS "$API/vms" | jq '.vms[] | {id,name,state,ports:.config.network.port_forwards}'
CONTAINER_ID="$(curl -fsS "$API/vms/$VM_ID" | jq -r '.runtime.external_id')"
container exec "$CONTAINER_ID" uname -a
curl -fsS -X POST "$API/vms/$VM_ID/stop" -H "Content-Type: application/json" -d '{"timeout_seconds":10}' | jq '.state'
curl -fsS -X DELETE "$API/vms/$VM_ID" -o /dev/null -w "%{http_code}\n"
```

Expected runtime fields include:

```json
{
  "driver": "apple_container",
  "external_id": "nf-..."
}
```

Expected `uname -a` result on the validated host:

```text
Linux nf-... 6.12.28 #1 SMP Tue May 20 15:19:05 UTC 2025 aarch64 Linux
```

Useful macOS launcher variants:

```bash
./scripts/run-tray-macos.sh --start-api --restart --foreground
./scripts/run-tray-macos.sh --start-api --restart --api-url http://127.0.0.1:18080
NANOFUSE_DATA_DIR=/tmp/nanofuse-macos-dev ./scripts/run-tray-macos.sh --start-api --restart
```

Stop launchd-managed local daemon:

```bash
launchctl bootout "gui/$(id -u)/com.daax.nanofuse.macos-api" 2>/dev/null || true
```

## Linux/KVM Firecracker Host

Run this on a machine that has Linux, `/dev/kvm`, and Firecracker for the normal daemon:

```bash
cd /path/to/nanofuse && ./scripts/ensure-mage.sh && mage build && sudo ./bin/nanofused -config config.dev.yaml -tcp 127.0.0.1:8080
```

For remote clients, keep the daemon bound to `127.0.0.1:8080` and use an SSH tunnel from the client machine.

For an explicit no-network rootless validation daemon, use a config with:

```yaml
network:
  setup: false
```

Then create VMs with `--network none`.

## Windows PowerShell Client

Run this from Windows PowerShell after replacing `user@linux-kvm-host` or pointing at a reachable macOS/Linux daemon:

```powershell
Start-Process ssh -ArgumentList '-N','-L','18080:127.0.0.1:8080','user@linux-kvm-host' -WindowStyle Hidden; $env:NANOFUSE_API_URL='http://127.0.0.1:18080'; .\nanofuse.exe health
```

This assumes `ssh.exe` is installed and the Windows `nanofuse.exe` binary is in the current directory. If the daemon is directly reachable on a trusted management network, set `NANOFUSE_API_URL` directly instead of using SSH.

## Windows Tray App

One-line build and launch from PowerShell in the repo root:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-tray-windows.ps1 -ApiUrl "$env:NANOFUSE_API_URL"
```

Explicit one-liner:

```powershell
if (-not $env:NANOFUSE_API_URL) { $env:NANOFUSE_API_URL = "http://127.0.0.1:18080" }; go build -ldflags "-H=windowsgui" -o bin\nanofuse-tray.exe .\cmd\nanofuse-tray; Start-Process .\bin\nanofuse-tray.exe -ArgumentList @("--api-url", $env:NANOFUSE_API_URL)
```

Smoke test:

```powershell
.\bin\nanofuse-tray.exe --smoke --api-url $env:NANOFUSE_API_URL
```

## Remote TCP Client

Use this only on a trusted management network or behind another authenticated transport:

```bash
NANOFUSE_API_URL=http://linux-or-mac-runtime-host:8080 ./bin/nanofuse health
```

```powershell
$env:NANOFUSE_API_URL='http://linux-or-mac-runtime-host:8080'; .\nanofuse.exe health
```

## Local Parallels/KVM Result

The local Parallels Ubuntu guest remains useful for Linux build/test validation. It is not a local Firecracker runtime host on this Apple Silicon Mac because `/dev/kvm` is not exposed inside the guest.

Host nested virtualization check:

```bash
swift -e 'import Virtualization; print(VZGenericPlatformConfiguration.isNestedVirtualizationSupported)'
```

Result:

```text
false
```

Guest KVM check:

```bash
vagrant ssh -c 'ls -l /dev/kvm 2>/dev/null || echo /dev/kvm-missing'
```

Result:

```text
/dev/kvm-missing
```

This result does not block macOS local VM execution because the macOS runtime uses Apple `container` plus Virtualization.framework instead of Firecracker/KVM.

Sources:

- Apple `container`: https://github.com/apple/container
- Apple container documentation: https://apple.github.io/container/documentation/
- Slicer Mac overview: https://docs.slicervm.com/mac/overview/
- Firecracker Getting Started: https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md
- Apple nested virtualization support API: https://developer.apple.com/documentation/virtualization/vzgenericplatformconfiguration/isnestedvirtualizationsupported
