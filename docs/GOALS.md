# Nanofuse Project Goals

**Last updated:** 2026-05-31
**Status:** Alpha. Linux Firecracker lifecycle exists; macOS local microVM lifecycle now runs through Apple's container runtime / Virtualization.framework; production sandbox controls are still being hardened.

## Mission

Nanofuse is a self-hosted microVM sandbox platform for running untrusted and semi-trusted code with a smaller security surface than container-only execution. The immediate target is AI coding-agent workloads that need filesystem state, controlled network access, and credential isolation.

## Current Platform Model

| Host family | Current support | Constraint |
|-------------|-----------------|------------|
| Linux | Native runtime target when `/dev/kvm` is present and readable/writable. | Firecracker requires Linux KVM. |
| macOS | Native local runtime target on Apple Silicon through Apple's `container` CLI, which runs Linux containers in lightweight VMs using Virtualization.framework. `nanofused` can run locally with `runtime.driver=apple_container`, and `nanofuse-tray` can start that daemon with `--start-api`. | Uses Apple container / Virtualization.framework, not Firecracker/KVM. Requires macOS 26-era Apple container tooling and arm64 Linux images. |
| Windows | Operator/developer host only. Use the API, CLI, PowerShell, SSH tunnel, or `nanofuse-tray` against a reachable Linux/KVM or macOS Apple-container `nanofused` daemon. WSL2 or local VM paths only work when Linux KVM is exposed. | Native Windows Firecracker runtime is not supported. |

Do not treat host OS sandboxing as equivalent to the Nanofuse security boundary. The runtime security boundary is a guest-kernel VM boundary: Firecracker/KVM on Linux, Apple Virtualization.framework on macOS.

Primary runtime references:

- [Firecracker getting started](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md)
- [Firecracker jailer](https://github.com/firecracker-microvm/firecracker/blob/main/docs/jailer.md)
- [Firecracker seccomp](https://github.com/firecracker-microvm/firecracker/blob/main/docs/seccomp.md)
- [Apple container project](https://github.com/apple/container)
- [Apple container documentation](https://apple.github.io/container/documentation/)
- [Slicer for Mac overview](https://docs.slicervm.com/mac/overview/)

## Required Outcomes

| Objective | Current state | Target |
|-----------|---------------|--------|
| Small security surface | Runtime abstraction supports Firecracker/KVM on Linux and Apple container / Virtualization.framework on macOS. Linux jailer integration is configured but not the default launch path. | Linux Firecracker launched through jailer by default; macOS backend tightened with explicit profile/policy controls; both paths release-gated by escape and lifecycle tests. |
| Container workload support | Linux can extract OCI/container images into rootfs artifacts. macOS can launch OCI images directly through Apple's container runtime, one Linux VM per container. | Any supported container workload can be wrapped into a bootable rootfs or runtime-native VM image and run inside microVM isolation. |
| API-driven control | `nanofused` exposes the REST API over Unix socket or optional TCP. CLI clients can use `--api-url` or `NANOFUSE_API_URL`; `GET /capabilities` reports selected runtime readiness (`firecracker` or `apple_container`). | Authenticated/TLS API profiles, generated SDKs, and packaged tray/menu clients for macOS and Windows. |
| Desktop management UI | Minimal `nanofuse-tray` API client exists for macOS and Windows. On macOS, `scripts/run-tray-macos.sh --start-api` starts a local Apple-container-backed daemon through launchd, then the tray can create/start/stop/delete VMs through the REST API. | Add packaged installers, profile storage, logs, image pull progress, auth-aware profiles, and richer VM/container controls. |
| Persistent filesystem | Writable root disks are materialized per VM under daemon storage; registered image rootfs files remain sources. | Policy-selectable ephemeral, persistent, and snapshot-backed filesystems. |
| Fast short-running sessions | Create/start/stop/kill lifecycle exists. Fast-start targets are not yet proven by current gates. | Measured cold-start and warm-start budgets with regression tests. |
| Long-running sessions | VMs can remain running until stopped or killed. | Lease, quota, idle timeout, and recovery policies for long-running sessions. |
| Secrets and identity away from LLMs | Host-side SPIFFE workload registration and vsock proxy exist. Raw secret injection is not implemented in VM config. | Guest receives identity material only; host-side broker or forced egress proxy injects credentials per request. Raw long-lived secrets never enter LLM-visible filesystem, env, logs, or prompts. |
| Restricted/interceptable egress | Per-VM L3/L4 egress policy supports default deny and proxy-only mode. L7 interception proxy is documented, not embedded. | All external LLM/API/MCP traffic forced through a host-controlled proxy with audit, policy, and credential injection. Direct bypass blocked. |
| Closed-loop kernel/runtime testing | `daax-dev/vagrant-skill` remains the Linux/kernel harness. Local macOS runtime validation now uses Apple container and proved an API-created VM can run Linux `6.12.28` on `aarch64`. | Every kernel/rootfs update is validated through vagrant-skill or equivalent Linux/KVM; every macOS runtime change is validated through local Apple-container API lifecycle tests. |

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
nanofused daemon
├── REST API / CLI / tray clients
├── runtime manager interface
│   ├── Linux: Firecracker + KVM, TAP/IPAM/port-forward/egress policy
│   └── macOS: Apple container + Virtualization.framework, runtime-managed NAT
├── image registry / rootfs build integration
├── VM metadata, logs, and filesystem state
└── identity and egress policy hooks
```

## Success Criteria

| Goal | Current gate | Target gate |
|------|--------------|-------------|
| Build and unit correctness | `mage ci` | Required before PR review |
| Kernel/rootfs compatibility | Vagrant or Linux/KVM closed-loop run | Required for kernel/rootfs changes |
| macOS local runtime | Apple-container API lifecycle run | Required for macOS runtime changes |
| Filesystem isolation | Unit tests for per-VM rootfs materialization | E2E sentinel tests inside a VM |
| Network restriction | Unit tests for egress rule generation | E2E blocked/allowed traffic tests inside a VM |
| Secret isolation | SPIFFE/vsock unit and integration tests | No raw secret matches in guest fs/env/log fixtures |
| Container wrapping | Image/layer build tests plus macOS Apple-container run | Boot a container-derived rootfs inside Firecracker and launch an OCI image through the macOS runtime |

## Related Documentation

| Document | Description |
|----------|-------------|
| [Launch One-Liners](LAUNCH_ONE_LINERS.md) | Tested macOS local runtime, Linux/KVM, Windows client, and local Parallels/KVM result |
| [Tray App](TRAY_APP.md) | macOS/Windows tray app launch, smoke test, and validation evidence |
| [Sandbox Objective Validation](building/sandbox-objective-validation.md) | Current validation plan and evidence |
| [API Quick Start](API_QUICK_START.md) | Runnable Linux/KVM and macOS Apple-container daemon examples |
| [Mac and Windows Clients](MAC_WINDOWS_CLIENTS.md) | macOS local runtime plus Windows client runbook |
| [Sandbox API Comparison](building/sandbox-api-comparison.md) | Comparison against current sandbox APIs |
| [Tray App Plan](building/nanofuse-tray-app.md) | macOS/Windows tray app implementation notes and remaining requirements |
| [Firecracker Runner Design](firecracker-runner-design.md) | Daemon/Firecracker design |
| [Programmable Egress Proxy Integration](prd/programmable-egress-proxy-integration.md) | Forced proxy model for LLM/API/MCP egress |
| [SPIFFE/SPIRE Integration Status](specs/spiffe-integration-status.md) | Current identity implementation status |
| [Advanced Firewall Capabilities](future-fw.md) | L3-L7 security reference |
