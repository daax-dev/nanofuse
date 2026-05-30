# Nanofuse

[![CI/CD Pipeline](https://github.com/daax-dev/nanofuse/actions/workflows/ci.yaml/badge.svg)](https://github.com/daax-dev/nanofuse/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/daax-dev/nanofuse)](https://goreportcard.com/report/github.com/daax-dev/nanofuse)

A Firecracker-based microVM platform for running untrusted code in secure, isolated sandboxes. Think [E2B](https://e2b.dev) but self-hosted.

## What is Nanofuse?

Nanofuse provides hardware-level isolation for running untrusted workloads with container-like performance. Each workload runs in its own microVM with a dedicated kernel, providing stronger isolation than containers while maintaining sub-second boot times.

**Primary use cases:**

| Use Case | Description |
|----------|-------------|
| **AI Code Execution** | Run LLM-generated code securely with sub-200ms boot times |
| **Isolated Workloads** | Ephemeral compute for untrusted or multi-tenant workloads |
| **Dev Sandboxes** | Fast-spinning development environments |

## Documentation

### Getting Started

| Document | Description |
|----------|-------------|
| [Quick Start Guide](docs/QUICKSTART.md) | Get Nanofuse running in 5 minutes |
| [API Quick Start](docs/API_QUICK_START.md) | REST API usage with curl examples |
| [Mac and Windows Clients](docs/MAC_WINDOWS_CLIENTS.md) | Manage a Linux/KVM daemon from macOS or Windows |
| [SSH Access Guide](docs/SSH_ACCESS_QUICK_START.md) | SSH into your microVMs |

### For Developers

| Document | Description |
|----------|-------------|
| [Developer Guide](docs/DEVELOPER_GUIDE.md) | Building, running, debugging, and troubleshooting |
| [Contributing](docs/CONTRIBUTING.md) | How to contribute to Nanofuse |
| [Troubleshooting](docs/TROUBLESHOOTING.md) | Common issues and solutions |
| [FAQ](docs/FAQ.md) | Frequently asked questions |

### Architecture and Design

| Document | Description |
|----------|-------------|
| [Project Goals](docs/GOALS.md) | Mission, design principles, and success criteria |
| [Sandbox API Comparison](docs/building/sandbox-api-comparison.md) | How Nanofuse differs from other sandbox APIs |
| [Architecture Design](docs/firecracker-runner-design.md) | Detailed implementation specification |
| [Networking](docs/firecracker-runner-networking-extension.md) | VM-to-VM communication and overlay networks |
| [Firewall Capabilities](docs/future-fw.md) | L3-L7 security control reference |

### Reference

| Document | Description |
|----------|-------------|
| [API Contract](docs/building/implementation/API_CONTRACT.md) | Complete REST API specification |
| [CLI Specification](docs/building/implementation/CLI_SPEC.md) | Command-line interface reference |
| [Image Version Catalog](images/VERSION_CATALOG.md) | Available microVM image versions |

## Quick Install

```bash
# Download latest release
VERSION=v0.1.0
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofuse
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofused
chmod +x nanofuse nanofused
sudo mv nanofuse nanofused /usr/local/bin/
```

For detailed installation instructions, see the [Quick Start Guide](docs/QUICKSTART.md).

## Architecture Overview

```
+-----------------------------------------------------------------+
|                         Host System                              |
+-----------------------------------------------------------------+
|  +-------------+  +-------------+  +-------------+               |
|  |  microVM 1  |  |  microVM 2  |  |  microVM N  |               |
|  | (Workload)  |  | (Workload)  |  |    ...      |               |
|  +------+------+  +------+------+  +------+------+               |
|         |                |                |                      |
|  +------+----------------+----------------+------+               |
|  |              Network Proxy / Firewall         |               |
|  +------------------------+----------------------+               |
|                           |                                      |
|  +------------------------+----------------------+               |
|  |              Nanofuse Control Plane           |               |
|  |  - microVM lifecycle (nanofused daemon)       |               |
|  |  - REST API + CLI (nanofuse)                  |               |
|  |  - Network management                         |               |
|  +-----------------------------------------------+               |
+-----------------------------------------------------------------+
```

**Components:**

| Component | Description |
|-----------|-------------|
| `nanofuse` | CLI tool for VM and image management |
| `nanofused` | API daemon managing Firecracker processes |
| `nanofuse-envd` | In-VM daemon for SDK interaction (planned) |

## Project Status

**Current Phase:** Phase 1 - Core Infrastructure

**Status:** Alpha - Core infrastructure approximately 60% complete. Not production-ready.

| Feature | Status |
|---------|--------|
| CLI implementation | Complete |
| API daemon | Complete |
| TAP networking with IPAM | Complete |
| Base microVM image | Complete |
| CI/CD pipeline | Complete |
| Snapshot/resume | In Progress |
| Python/JS SDKs | Planned |

See [Project Goals](docs/GOALS.md) for the complete roadmap.

## Requirements

- Linux host with KVM support (`/dev/kvm`)
- x86_64 architecture (ARM64 support planned)
- Root access or appropriate capabilities for networking

## Support

- [Issue Tracker](https://github.com/daax-dev/nanofuse/issues)
- [Discussions](https://github.com/daax-dev/nanofuse/discussions)
- [FAQ](docs/FAQ.md)

## Acknowledgments

- [E2B](https://github.com/e2b-dev) - Primary architectural inspiration
- [Firecracker](https://github.com/firecracker-microvm/firecracker) - The microVM technology powering Nanofuse
- [Slicer](https://docs.slicervm.com) - Inspiration for image format and build approach

---

*Last updated: December 2025*
