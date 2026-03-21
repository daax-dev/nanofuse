# Nanofuse Project Goals

## Mission Statement

Nanofuse is a secure, microVM-based platform for running untrusted and semi-trusted workloads with granular access controls. The platform enables safe execution of AI coding agents and other potentially risky workloads by providing hardware-level isolation while maintaining container-like performance and developer experience.

## Problem Statement

Modern AI coding agents (Claude Code, GitHub Copilot, Cursor, etc.) require execution environments that balance two competing needs:

1. **Productivity**: Developers need agents with sufficient access to read code, execute commands, and modify files
2. **Security**: Organizations need protection against agents that may exfiltrate data, execute malicious code, or consume excessive resources

Traditional solutions fall short:

| Approach | Limitation |
|----------|------------|
| Containers | Share the host kernel; insufficient isolation for untrusted code ([Amazon Science][1]) |
| Traditional VMs | High overhead (seconds to boot, GBs of memory) |
| Language sandboxes | Bypassable; insufficient for defense in depth ([NVIDIA Technical Blog][2]) |

Nanofuse addresses this gap by providing microVM-based isolation with sub-second boot times and minimal resource overhead.

## Goals

### Primary Goals

1. **Multi-Platform Support**
   - Run on Linux (native KVM), macOS (via Virtualization.framework), and Windows (WSL2)
   - Deploy in self-hosted, AWS, GCP, and Azure environments
   - Provide consistent security guarantees across all platforms

2. **Granular Access Control**
   - **Filesystem**: Controlled read/write access to specific paths
   - **Network**: Allowlist-based outbound connectivity with L3/L4/L7 filtering
   - **Secrets**: Scoped credential injection with automatic rotation support

3. **AI Agent Isolation**
   - Isolate each coding agent session in its own microVM
   - Log and audit all LLM API traffic for prompt inspection
   - Support future AI guardrails for content filtering

4. **Developer Experience**
   - TTY/CLI access via SSH or web terminal
   - Sub-second cold start for interactive use
   - Seamless integration with existing development workflows

### Success Criteria

| Goal | Metric | Target |
|------|--------|--------|
| Boot time | Cold start to shell | < 500ms |
| Memory overhead | Per-microVM footprint | < 10 MiB |
| Density | Concurrent microVMs per host | > 100 |
| Security | Kernel attack surface | No shared kernel with host |

## Design Principles

### Security First

Treat all guest code as untrusted. The guest has full access to its own microVM kernel, but that kernel is explicitly isolated from the host via hardware virtualization ([AWS Lambda Firecracker Design][3]).

**Defense in Depth layers:**
1. Hardware virtualization (KVM/Hypervisor)
2. Minimal VMM with reduced attack surface
3. Jailer component with seccomp-bpf, cgroups, and chroot
4. Network proxy with traffic inspection

### Ephemeral by Default

Treat microVM filesystems as disposable scratch space:
- Persist only explicitly designated artifacts (git commits, build outputs)
- Assume short-lived secrets may exist; minimize blast radius
- Support snapshotting for reproducible environments

### Observable and Auditable

All security-relevant events must be logged:
- Network connections and data transfer volumes
- LLM API calls with prompt/response logging
- Filesystem access patterns
- Resource consumption metrics

## Architecture Overview

```
+-----------------------------------------------------------------+
|                         Host System                              |
+-----------------------------------------------------------------+
|  +-------------+  +-------------+  +-------------+               |
|  |  microVM 1  |  |  microVM 2  |  |  microVM N  |               |
|  | (Agent A)   |  | (Agent B)   |  |    ...      |               |
|  +------+------+  +------+------+  +------+------+               |
|         |                |                |                      |
|  +------+----------------+----------------+------+               |
|  |              Network Proxy / Firewall         |               |
|  |         (L3/L4 rules + L7 AI guardrails)      |               |
|  +------------------------+----------------------+               |
|                           |                                      |
|  +------------------------+----------------------+               |
|  |              Nanofuse Control Plane           |               |
|  |  - microVM lifecycle management (start/stop)  |               |
|  |  - API Gateway (REST + MCP)                   |               |
|  |  - Telemetry aggregation                      |               |
|  |  - Secret injection                           |               |
|  +-----------------------------------------------+               |
+-----------------------------------------------------------------+
```

### Control Plane Components

| Component | Responsibility |
|-----------|----------------|
| **nanofused** | microVM lifecycle daemon; manages start/stop/snapshot |
| **Firewall** | iptables/nftables rules over tun/tap interfaces |
| **API Gateway** | REST API for orchestration; MCP protocol support |
| **Proxy** | L7 reverse proxy for LLM traffic inspection |
| **Telemetry** | Log collection and metrics aggregation |

## Technology Choices

### Primary Hypervisor: Firecracker

[Firecracker][4] is the default microVM engine, selected for:

- **Minimal attack surface**: ~50,000 lines of Rust vs. 1.4M+ lines in QEMU ([Amazon Science][1])
- **Fast startup**: 125ms to user code execution
- **Low overhead**: < 5 MiB memory per microVM
- **Production proven**: Powers AWS Lambda and Fargate

**Limitations:**
- No PCI bus (no GPU passthrough)
- Limited block device options (virtio-block only)

### Alternative Hypervisor: Cloud Hypervisor

[Cloud Hypervisor][5] is used when additional capabilities are required:

- **GPU passthrough**: VFIO support for direct GPU access ([Cloud Hypervisor v38.0][6])
- **Multi-GPU workloads**: PCIe P2P for GPUDirect operations
- **Hot-plug support**: CPU, memory, and device hot-plug

**Trade-offs:**
- Larger footprint than Firecracker
- More complex configuration

### Container Runtime

microVMs must support running OCI-compatible containers via:
- [containerd][7] as the container runtime
- [firecracker-containerd][8] for Firecracker integration

### Network Security

Egress filtering implemented via:
- **L3/L4**: iptables/nftables rules on host tun/tap interfaces
- **L7**: Reverse proxy for HTTPS traffic inspection (especially LLM API calls)
- **Future**: AI guardrails for prompt/response filtering

See [Advanced Firewall Capabilities](future-fw.md) for the complete L3-L7 security reference.

## Boot-Time Configuration

microVMs receive configuration at boot via:

| Method | Use Case |
|--------|----------|
| Kernel command line | Basic parameters |
| virtio-vsock | Control plane communication |
| Cloud-init / ignition | User configuration, SSH keys |
| Secrets injection | Short-lived credentials |

Required boot-time parameters:
- SSH public keys for `authorized_keys`
- Network configuration
- Workspace mount points
- Initial secrets (with TTL)

## Related Documentation

| Document | Description |
|----------|-------------|
| [Firecracker Runner Design](firecracker-runner-design.md) | Detailed implementation specification |
| [Networking Extension](firecracker-runner-networking-extension.md) | VM-to-VM communication and overlay networks |
| [Advanced Firewall Capabilities](future-fw.md) | L3-L7 security control reference |

## References

### Core Technologies

- [KVM (Kernel-based Virtual Machine)][9] - Linux hypervisor
- [Linux ABI Guide][10] - Kernel interface stability

### microVM Platforms

- [Firecracker][4] - AWS microVM
- [Cloud Hypervisor][5] - GPU-capable microVM
- [firecracker-containerd][8] - Container runtime integration

### Related Projects

- [SlicerVM][11] - microVM management platform
- [Kata Containers with Firecracker][12] - Kubernetes integration
- [Firecracker in Docker (PoC)][13] - Nested virtualization exploration

### AI Agent Security

- [Security of AI Agents (arXiv)][14] - Academic research on agent security
- [Code Sandboxes for LLM AI Agents][15] - Sandbox comparison
- [Sandboxing for AI Agents][16] - Best practices guide

---

[1]: https://www.amazon.science/blog/how-awss-firecracker-virtual-machines-work
[2]: https://developer.nvidia.com/blog/how-code-execution-drives-key-risks-in-agentic-ai-systems/
[3]: https://aws.amazon.com/blogs/aws/firecracker-lightweight-virtualization-for-serverless-computing/
[4]: https://github.com/firecracker-microvm/firecracker
[5]: https://github.com/cloud-hypervisor/cloud-hypervisor
[6]: https://www.cloudhypervisor.org/blog/cloud-hypervisor-v38.0-released/
[7]: https://containerd.io/
[8]: https://github.com/firecracker-microvm/firecracker-containerd
[9]: https://linux-kvm.org/page/Main_Page
[10]: https://docs.kernel.org/admin-guide/abi.html
[11]: https://docs.slicervm.com/
[12]: https://arun-gupta.github.io/kata-firecracker/
[13]: https://github.com/fadams/firecracker-in-docker
[14]: https://arxiv.org/html/2406.08689v2
[15]: https://amirmalik.net/2025/03/07/code-sandboxes-for-llm-ai-agents
[16]: https://medium.com/@yessine.abdelmaksoud.03/sandboxing-for-ai-agents-2420ac69569e

---

*Last updated: December 2025*
