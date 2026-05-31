# Nanofuse API Quick Start

Nanofuse is controlled through the `nanofused` REST API daemon. The daemon must run on a Linux host that exposes `/dev/kvm`; macOS and Windows machines connect to that daemon as API clients.

## Requirements

- Runtime host: Linux with read/write `/dev/kvm`.
- Runtime privileges: root or equivalent capabilities for TAP, bridge, NAT, and Firecracker.
- Client host: Linux, macOS, or Windows with `curl`, PowerShell, the `nanofuse` CLI, or `nanofuse-tray`.

Native Firecracker execution on macOS or Windows is not supported. Those hosts can still manage VMs through the API when a Linux/KVM daemon is reachable.

## Start a Linux/KVM Daemon

Build from source:

```bash
./scripts/ensure-mage.sh
mage daemon
```

Start with the development config on localhost:

```bash
sudo ./bin/nanofused -config config.dev.yaml -tcp 127.0.0.1:8080
```

Expose the API to another machine only on a trusted management network or tunnel:

```bash
sudo ./bin/nanofused -config config.dev.yaml -tcp 0.0.0.0:8080
```

The TCP API currently has no built-in authentication or TLS. Do not expose it directly to untrusted networks. Use SSH forwarding, WireGuard, or a reverse proxy with authentication until first-party API auth lands.

## Health and Capabilities

Unix socket:

```bash
curl --unix-socket /var/run/nanofused.sock http://localhost/health
curl --unix-socket /var/run/nanofused.sock http://localhost/capabilities
```

TCP:

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/capabilities
```

Expected health shape:

```json
{
  "status": "healthy",
  "version": "0.1.0",
  "uptime_seconds": 12
}
```

`GET /capabilities` reports the daemon OS, architecture, KVM availability, Firecracker binary path, and configured API transports. Tray apps and SDKs should use it before enabling VM actions.

## Mac Client

Point the CLI at a Linux/KVM daemon:

```bash
export NANOFUSE_API_URL="http://linux-kvm-host:8080"
nanofuse health
nanofuse vm list
```

Use an SSH tunnel when the daemon should remain bound to localhost on the Linux host:

```bash
ssh -L 18080:127.0.0.1:8080 user@linux-kvm-host
export NANOFUSE_API_URL="http://127.0.0.1:18080"
nanofuse health
```

## Windows Client

PowerShell against a Linux/KVM daemon:

```powershell
$env:NANOFUSE_API_URL = "http://linux-kvm-host:8080"
Invoke-RestMethod "$env:NANOFUSE_API_URL/health"
Invoke-RestMethod "$env:NANOFUSE_API_URL/capabilities"
.\nanofuse.exe health
.\nanofuse.exe vm list
```

SSH tunnel from Windows:

```powershell
ssh -L 18080:127.0.0.1:8080 user@linux-kvm-host
$env:NANOFUSE_API_URL = "http://127.0.0.1:18080"
.\nanofuse.exe health
```

## Tray Client

macOS:

```bash
NANOFUSE_API_URL="${NANOFUSE_API_URL:-http://127.0.0.1:18080}" ./scripts/run-tray-macos.sh
```

Windows PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-tray-windows.ps1 -ApiUrl "$env:NANOFUSE_API_URL"
```

See [Tray App](TRAY_APP.md) for smoke mode and validation evidence.

In the tray menu, select a cached image, then choose `Create and Start VM From Image`. This calls `POST /vms` followed by `POST /vms/{id}/start` against the configured Linux/KVM daemon.

## Vagrant API Path

The development Vagrant VM forwards guest port `8080` to host port `18080` by default:

```bash
cd dev/vagrant
NANOFUSE_API_HOST_PORT=18080 vagrant up
vagrant ssh -c "sudo systemctl start nanofused"
curl http://127.0.0.1:18080/health
```

This only reaches Firecracker execution when the Vagrant provider exposes Linux KVM inside the guest. If `/dev/kvm` is missing in the guest, validation fails before VM boot.

## Image and VM Workflow

Pull an image:

```bash
curl -X POST http://127.0.0.1:8080/images/pull \
  -H "Content-Type: application/json" \
  -d '{"image_ref":"ghcr.io/daax-dev/nanofuse/base:latest"}'
```

Poll the pull job:

```bash
curl http://127.0.0.1:8080/images/jobs/job-550e8400-e29b-41d4-a716-446655440000
```

Create a VM:

```bash
curl -X POST http://127.0.0.1:8080/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api-test",
    "image": "ghcr.io/daax-dev/nanofuse/base:latest",
    "config": {
      "vcpus": 2,
      "memory_mib": 512,
      "network": {
        "mode": "nat",
        "egress_policy": {
          "enabled": true,
          "default_action": "deny",
          "allow_dns": true
        }
      }
    }
  }'
```

Start, inspect, read logs, stop, and delete:

```bash
VM_ID="550e8400-e29b-41d4-a716-446655440000"
curl -X POST "http://127.0.0.1:8080/vms/${VM_ID}/start"
curl "http://127.0.0.1:8080/vms/${VM_ID}"
curl "http://127.0.0.1:8080/vms/${VM_ID}/logs?tail=50"
curl -X POST "http://127.0.0.1:8080/vms/${VM_ID}/stop" \
  -H "Content-Type: application/json" \
  -d '{"timeout_seconds":30}'
curl -X DELETE "http://127.0.0.1:8080/vms/${VM_ID}"
```

CLI equivalent:

```bash
export NANOFUSE_API_URL="http://127.0.0.1:8080"
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:latest
nanofuse vm create ghcr.io/daax-dev/nanofuse/base:latest api-test --vcpus 2 --memory 512
nanofuse vm list
```

## Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Daemon health |
| GET | `/capabilities` | Host/runtime/API capability details |
| GET | `/vms` | List VMs |
| POST | `/vms` | Create VM |
| GET | `/vms/{id}` | Get VM details |
| DELETE | `/vms/{id}` | Delete VM |
| POST | `/vms/{id}/start` | Start VM |
| POST | `/vms/{id}/stop` | Stop VM |
| POST | `/vms/{id}/kill` | Kill VM |
| POST | `/vms/{id}/pause` | Pause VM |
| POST | `/vms/{id}/resume` | Resume VM |
| GET | `/vms/{id}/logs` | VM console logs |
| GET | `/vms/{id}/snapshots` | List VM snapshots |
| POST | `/vms/{id}/snapshots` | Create VM snapshot |
| GET | `/snapshots/{id}` | Get snapshot |
| DELETE | `/snapshots/{id}` | Delete snapshot |
| GET | `/images` | List local images |
| POST | `/images/pull` | Pull image |
| GET | `/images/jobs/{id}` | Get pull job status |
| GET | `/images/{digest}` | Get image |
| DELETE | `/images/{digest}` | Delete image |

See [`api/openapi.yaml`](../api/openapi.yaml) for the contract.
