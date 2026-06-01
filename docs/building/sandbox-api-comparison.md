# Sandbox API Comparison

**Date:** 2026-05-30

This comparison uses current official documentation for the API shape of other sandbox tools. It is scoped to developer-facing API semantics, not pricing or hosting model.

## Summary

| Tool | API center | Workload boundary | In-sandbox process/file API | Lifecycle model | Nanofuse difference |
|------|------------|-------------------|-----------------------------|-----------------|---------------------|
| Nanofuse | Self-hosted `nanofused` REST API over Unix socket or TCP | Firecracker microVM on Linux/KVM; Apple `container`/Virtualization.framework on macOS | Not yet. Current API manages VM/image/snapshot/log lifecycle. | Create/start/stop/kill/delete VMs; pull or resolve images; snapshots are partially implemented. | Self-hosted microVM control plane with host-owned networking, egress policy, image materialization, local storage, and a macOS-native runtime backend. |
| E2B | Cloud SDK/API centered on `Sandbox` | Cloud sandbox environment | Yes: files, commands, PTY, git, code execution, MCP options. | Create/connect/kill/pause/snapshot/timeouts. | E2B is closer to a ready code-execution SDK. Nanofuse currently exposes lower-level VM lifecycle and must add envd/exec/file APIs to match that ergonomics. |
| Daytona | SDK/API centered on `Sandbox` and toolbox APIs | Dedicated sandbox computer with kernel/filesystem/network resources | Yes: filesystem, process, code interpreter, computer-use, git. | Start/stop/delete/recover/resize/autostop/archive/delete timers. | Daytona has richer developer-environment semantics. Nanofuse is lower-level, self-hosted, and focused on Firecracker VM controls first. |
| Modal Sandbox | Python SDK object similar to process/container execution | Modal-managed sandbox/container | Yes: command exec, filesystem namespace, tunnels, volumes, network controls. | Async create, wait/poll/terminate, exec, snapshots/volumes/tunnels. | Modal is cloud/serverless SDK-first. Nanofuse is local daemon/API-first and does not require Modal's platform. |
| Fly Machines | REST Machines API | Fly-managed fast-launching VM/machine | No general file API in Machines API; machine config/lifecycle is primary. | Create/start/stop/suspend/delete/update/wait/lease/metadata. | Similar lifecycle shape, but Fly is hosted and app-scoped; Nanofuse is self-hosted and will add sandbox-specific controls such as per-VM egress policy. |
| Docker Engine API | REST API to Docker daemon | Host-kernel containers | Yes: container exec and filesystem-related APIs through Docker primitives. | Container create/start/stop/kill/delete/exec/logs/images. | Docker is container isolation. Nanofuse uses containers as image/build inputs but the runtime isolation boundary is a microVM. |
| Firecracker | Per-process VMM API over host endpoint | Single Firecracker microVM | No guest command/file API. It configures VM devices and lifecycle. | Configure boot source, drives, network, machine config, actions, metrics, MMDS. | Nanofuse wraps Firecracker with image registry, storage, network/IPAM, policy, API clients, and multi-VM daemon state. |

## Required Nanofuse API Direction

Nanofuse needs two API layers:

1. **Control-plane API**: current `nanofused` surface for health, capabilities, images, VM lifecycle, snapshots, logs, egress policy, and storage.
2. **Sandbox interaction API**: planned in-VM agent/envd surface for command execution, file transfer, port/tunnel management, process sessions, and tool/credential brokerage.

The current PR strengthens the first layer. It does not claim parity with E2B, Daytona, or Modal for process/file/code execution APIs.

## API Gaps To Track

| Gap | Why it matters | Direction |
|-----|----------------|-----------|
| In-VM command execution | E2B, Daytona, Modal users can run commands/code without SSH. | Add a guest agent over vsock with session, process, stdout/stderr, exit status, and timeout semantics. |
| File transfer API | SDK/tray clients need upload/download without SSH. | Add guest-agent filesystem endpoints with path validation and size limits. |
| Authenticated remote API | Mac/Windows clients and tray apps need safe remote access. | Add API profiles with mTLS or bearer auth; document reverse-proxy option until native auth ships. |
| SDK-generated clients | Desktop/tray and external automation should not hand-roll REST calls. | Generate clients from `api/openapi.yaml` after endpoint shapes stabilize. |
| Port/tunnel management | Developer sandboxes need preview/service access. | Build explicit port-forward/tunnel endpoints above existing network config. |

## Official References

- [E2B Python Sandbox SDK](https://e2b.dev/docs/sdk-reference/python-sdk/v2.15.0/sandbox_sync)
- [Daytona Python Sandbox SDK](https://www.daytona.io/docs/en/python-sdk/sync/sandbox/)
- [Modal `modal.Sandbox`](https://modal.com/docs/reference/modal.Sandbox)
- [Fly Machines resource](https://fly.io/docs/machines/api/machines-resource/)
- [Docker Engine API](https://docs.docker.com/reference/api/engine/)
- [Firecracker repository/API overview](https://github.com/firecracker-microvm/firecracker)
