# Mac and Windows Client Runbook

Nanofuse runtime execution is Linux/KVM. macOS and Windows clients manage a Linux `nanofused` daemon over the API.

## Supported Topology

```text
macOS or Windows client
  -> nanofuse CLI, curl, PowerShell, or tray app
  -> HTTP API
  -> Linux/KVM host running nanofused
  -> Firecracker microVMs
```

This is the supported cross-platform model today. Do not treat native macOS or Windows virtualization as the Nanofuse security boundary.

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

## macOS

```bash
ssh -L 18080:127.0.0.1:8080 user@linux-kvm-host
export NANOFUSE_API_URL="http://127.0.0.1:18080"

nanofuse health
curl "$NANOFUSE_API_URL/capabilities"
nanofuse vm list
```

## Windows PowerShell

```powershell
ssh -L 18080:127.0.0.1:8080 user@linux-kvm-host
$env:NANOFUSE_API_URL = "http://127.0.0.1:18080"

.\nanofuse.exe health
Invoke-RestMethod "$env:NANOFUSE_API_URL/capabilities"
.\nanofuse.exe vm list
```

## Vagrant From a Client Host

The Vagrant development environment forwards host `127.0.0.1:18080` to guest `8080`:

```bash
cd dev/vagrant
NANOFUSE_API_HOST_PORT=18080 vagrant up
vagrant ssh -c "sudo systemctl start nanofused"
curl http://127.0.0.1:18080/health
```

This requires a provider that exposes Linux KVM in the guest. On providers that do not expose `/dev/kvm`, the API may be reachable for health checks only if `nanofused` starts; VM execution must fail because Firecracker cannot run.

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

## Tray App Requirement

A macOS/Windows tray app should be an API client only. It must not call Firecracker, manipulate TAP devices, edit Nanofuse storage, or shell into the runtime host directly. The required first screen is daemon health/capabilities, followed by VM and image lifecycle controls backed by `api/openapi.yaml`.
