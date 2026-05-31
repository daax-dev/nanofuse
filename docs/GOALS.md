# Nanofuse Project Goals

**Last updated:** 2026-05-31
**Status:** Alpha. Core Firecracker lifecycle exists; production sandbox controls are being added and validated.

## Mission

Nanofuse is a self-hosted microVM sandbox platform for running untrusted and semi-trusted code with a smaller security surface than container-only execution. The immediate target is AI coding-agent workloads that need filesystem state, controlled network access, and credential isolation.

## Current Platform Model

| Host family | Current support | Constraint |
|-------------|-----------------|------------|
| Linux | Native runtime target when `/dev/kvm` is present and readable/writable. | Firecracker requires Linux KVM. |
| macOS | Operator/developer host only. Use the API, CLI, curl, SSH tunnel, or planned tray app against a Linux/KVM `nanofused` daemon. Local Vagrant only works when the provider exposes `/dev/kvm`. | Native macOS Firecracker runtime is not supported. |
| Windows | Operator/developer host only. Use the API, CLI, PowerShell, SSH tunnel, or planned tray app against a Linux/KVM `nanofused` daemon. WSL2 or local VM paths only work when Linux KVM is exposed. | Native Windows Firecracker runtime is not supported. |

Do not treat macOS or Windows local OS sandboxing as equivalent to the Nanofuse security boundary. The runtime security boundary is the Linux/KVM microVM.

Primary Firecracker references:

- [Firecracker getting started](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md)
- [Firecracker jailer](https://github.com/firecracker-microvm/firecracker/blob/main/docs/jailer.md)
- [Firecracker seccomp](https://github.com/firecracker-microvm/firecracker/blob/main/docs/seccomp.md)

## Required Outcomes

| Objective | Current state | Target |
|-----------|---------------|--------|
| Small security surface | Firecracker VMM with TAP networking and optional SPIRE vsock proxy. Jailer integration is configured but not the default launch path. | Firecracker launched through jailer by default with cgroups, chroot, seccomp, least-privilege file layout, and release-gated escape tests. |
| Container workload support | OCI/container images can be extracted into microVM rootfs artifacts through Docker/Podman and layer build paths. | Any supported container workload can be wrapped into a bootable rootfs or container-capable guest image and run inside microVM isolation. |
| API-driven control | `nanofused` exposes the REST API over Unix socket or optional TCP. CLI clients can use `--api-url` or `NANOFUSE_API_URL`; `GET /capabilities` reports runtime readiness. | Authenticated/TLS API profiles, generated SDKs, and tray/menu clients for macOS and Windows. |
| Desktop management UI | No desktop UI is implemented. Requirements are captured for an API-only tray/menu app so it does not bypass the daemon boundary. | macOS and Windows tray apps manage daemon profiles, health, images, VMs, logs, and lifecycle actions through the REST API only. |
| Persistent filesystem | Writable root disks are materialized per VM under daemon storage; registered image rootfs files remain sources. | Policy-selectable ephemeral, persistent, and snapshot-backed filesystems. |
| Fast short-running sessions | Create/start/stop/kill lifecycle exists. Fast-start targets are not yet proven by current gates. | Measured cold-start and warm-start budgets with regression tests. |
| Long-running sessions | VMs can remain running until stopped or killed. | Lease, quota, idle timeout, and recovery policies for long-running sessions. |
| Secrets and identity away from LLMs | Host-side SPIFFE workload registration and vsock proxy exist. Raw secret injection is not implemented in VM config. | Guest receives identity material only; host-side broker or forced egress proxy injects credentials per request. Raw long-lived secrets never enter LLM-visible filesystem, env, logs, or prompts. |
| Restricted/interceptable egress | Per-VM L3/L4 egress policy supports default deny and proxy-only mode. L7 interception proxy is documented, not embedded. | All external LLM/API/MCP traffic forced through a host-controlled proxy with audit, policy, and credential injection. Direct bypass blocked. |
| Closed-loop kernel testing | `dev/vagrant` provides the local hypervisor harness and capability checks. | Every kernel/rootfs update is validated through Vagrant or an equivalent Linux/KVM runner before merge. |

## Design Principles

### Security Boundary First

Untrusted code must run behind a guest-kernel boundary. Containers are accepted as packaging/build inputs or as guest-internal workloads, not as the host isolation boundary for adversarial code.

### Immutable Source Images, Mutable VM State

Registered images are source artifacts. Writable VM disks are copied into VM-specific storage so one VM cannot mutate the base image used by another VM.

### Deny by Policy

Network policy must be explicit for untrusted agent jobs. Proxy-only mode is the required path for credential-injected LLM/API/MCP calls because the guest should not possess raw upstream credentials.

### Validate Before Claiming

Performance, platform support, and security claims must be tied to a test result or a primary source. If a host cannot expose Linux KVM, validation must fail with that exact reason.

## Architecture Direction

```text
Host Linux/KVM
├── nanofused daemon
│   ├── VM lifecycle and Firecracker process control
│   ├── per-VM root disk materialization
│   ├── TAP/IPAM/port-forward/egress policy setup
│   ├── image registry and rootfs build integration
│   └── SPIFFE/SPIRE registration hooks
├── host egress proxy (planned sidecar)
│   ├── LLM/API/MCP allow rules
│   ├── credential injection
│   └── audit events
└── Firecracker microVMs
    ├── dedicated guest kernel
    ├── VM-specific writable root disk
    └── optional guest identity client over vsock
```

## Success Criteria

| Goal | Current gate | Target gate |
|------|--------------|-------------|
| Build and unit correctness | `mage ci` | Required before PR review |
| Kernel/rootfs compatibility | Vagrant or Linux/KVM closed-loop run | Required for kernel/rootfs changes |
| Filesystem isolation | Unit tests for per-VM rootfs materialization | E2E sentinel tests inside a VM |
| Network restriction | Unit tests for egress rule generation | E2E blocked/allowed traffic tests inside a VM |
| Secret isolation | SPIFFE/vsock unit and integration tests | No raw secret matches in guest fs/env/log fixtures |
| Container wrapping | Image/layer build tests | Boot a container-derived rootfs inside Firecracker |

## Related Documentation

| Document | Description |
|----------|-------------|
| [Launch One-Liners](LAUNCH_ONE_LINERS.md) | Tested Mac/Windows client one-liners and local Parallels/KVM result |
| [Sandbox Objective Validation](building/sandbox-objective-validation.md) | Current validation plan and evidence |
| [API Quick Start](API_QUICK_START.md) | Runnable Linux/KVM API daemon and client examples |
| [Mac and Windows Clients](MAC_WINDOWS_CLIENTS.md) | Cross-platform client runbook |
| [Sandbox API Comparison](building/sandbox-api-comparison.md) | Comparison against current sandbox APIs |
| [Tray App Plan](building/nanofuse-tray-app.md) | macOS/Windows tray app requirements |
| [Firecracker Runner Design](firecracker-runner-design.md) | Daemon/Firecracker design |
| [Programmable Egress Proxy Integration](prd/programmable-egress-proxy-integration.md) | Forced proxy model for LLM/API/MCP egress |
| [SPIFFE/SPIRE Integration Status](specs/spiffe-integration-status.md) | Current identity implementation status |
| [Advanced Firewall Capabilities](future-fw.md) | L3-L7 security reference |
