# Mac and Windows Client Runbook

macOS is both a local runtime host and an API/tray client:

- Local runtime: `nanofused` uses Apple `container` plus Virtualization.framework with `runtime.driver=apple_container`.
- Remote client: macOS can still manage a Linux/KVM Firecracker daemon over the API.

Windows is currently an API/tray client. It manages a reachable Linux or macOS `nanofused` daemon; local Windows runtime execution is not implemented in this repo.

## Supported Topologies

```text
macOS local runtime
  -> nanofuse CLI, curl, or nanofuse-tray
  -> HTTP API on 127.0.0.1:18080
  -> nanofused runtime.driver=apple_container
  -> Apple container / Virtualization.framework Linux microVMs
```

```text
Linux runtime
  -> nanofuse CLI, curl, or nanofuse-tray
  -> HTTP or Unix socket API
  -> nanofused runtime.driver=firecracker
  -> Firecracker microVMs on Linux/KVM
```

```text
Windows client
  -> nanofuse CLI, PowerShell, or nanofuse-tray
  -> HTTP API
  -> reachable macOS or Linux nanofused daemon
```

## macOS Local Runtime

Start the daemon and menu bar app:

```bash
./scripts/run-tray-macos.sh --start-api --restart
```

Smoke check:

```bash
./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 5s
```

CLI and API checks:

```bash
export NANOFUSE_API_URL="http://127.0.0.1:18080"
nanofuse health
curl "$NANOFUSE_API_URL/capabilities"
nanofuse vm list
```

The expected macOS capability signal is `driver=apple_container`, `native_runtime=true`, `apple_container_available=true`, and `virtualization_framework_supported=true`.

## macOS Remote Client

Use this path when managing a Linux/KVM Firecracker daemon:

```bash
ssh -L 18080:127.0.0.1:8080 user@linux-kvm-host
export NANOFUSE_API_URL="http://127.0.0.1:18080"

nanofuse health
curl "$NANOFUSE_API_URL/capabilities"
nanofuse vm list
```

## Linux/KVM Host

Start the daemon bound to localhost and use an SSH tunnel:

```bash
sudo ./bin/nanofused -config config.dev.yaml -tcp 127.0.0.1:8080
```

Or bind to a management interface:

```bash
sudo ./bin/nanofused -config config.dev.yaml -tcp 0.0.0.0:8080
```

Raw TCP has no built-in Nanofuse auth/TLS. Restrict it with host firewall rules, SSH, WireGuard, or an authenticated reverse proxy.

## Windows PowerShell

```powershell
ssh -L 18080:127.0.0.1:8080 user@linux-kvm-host
$env:NANOFUSE_API_URL = "http://127.0.0.1:18080"

.\nanofuse.exe health
Invoke-RestMethod "$env:NANOFUSE_API_URL/capabilities"
.\nanofuse.exe vm list
```

If the daemon is reachable directly on a trusted management network:

```powershell
$env:NANOFUSE_API_URL = "http://linux-or-mac-runtime-host:8080"
.\nanofuse.exe health
.\nanofuse.exe vm list
```

Tray app:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-tray-windows.ps1 -ApiUrl "$env:NANOFUSE_API_URL"
```

## Daily Operations

For the exact commands to see published ports, execute commands inside running VMs through the API, launch multiple VMs, and enable more launchable OCI images, see [Operating Local MicroVMs](OPERATING_LOCAL_MICROVMS.md).

## Vagrant From a Client Host

The Vagrant development environment forwards host `127.0.0.1:18080` to guest `8080`:

```bash
cd dev/vagrant
NANOFUSE_API_HOST_PORT=18080 vagrant up
vagrant ssh -c "sudo systemctl start nanofused"
curl http://127.0.0.1:18080/health
```

This reaches Firecracker execution only when the provider exposes Linux KVM in the guest. On the local Apple Silicon Parallels VM, `/dev/kvm` is absent, so the VM remains useful for Linux build/test validation but not Firecracker boot validation.

## Client Configuration

The CLI reads these environment variables:

| Variable | Purpose |
|----------|---------|
| `NANOFUSE_API_URL` | TCP API base URL, such as `http://127.0.0.1:18080` |
| `NANOFUSE_API_SOCKET` | Unix socket path when running on the Linux host |
| `NANOFUSE_TIMEOUT` | Request timeout, such as `30s` |
| `NANOFUSE_OUTPUT=json` | JSON CLI output |
| `NANOFUSE_DEBUG=true` | Debug request logging |
| `NANOFUSE_NO_COLOR=true` | Disable color output |

`--api-url` and `--api-socket` still work and take precedence over environment values.

## Tray App

`nanofuse-tray` is implemented as an API client. It must not call Firecracker directly, manipulate TAP devices, edit Nanofuse storage outside the daemon, or shell into a runtime host. The current app shows daemon health/capabilities, VM list, image list, create/start from selected image, and VM start/stop/kill/delete actions backed by `api/openapi.yaml`.

See [Tray App](TRAY_APP.md).
