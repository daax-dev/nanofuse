# MicroVM Systems: Technical Comparison

**NanoFuse vs K7 (Katakate) vs SlicerVM**

*A comprehensive technical analysis of three Firecracker-based microVM platforms*

---

## Executive Summary

This document provides an in-depth comparison of three microVM management systems: **NanoFuse**, **K7 (Katakate)**, and **SlicerVM**. While all three leverage Firecracker for lightweight virtualization, they differ significantly in architecture, complexity, target use cases, and philosophical approach.

| System | Type | Maturity | License | Primary Focus |
|--------|------|----------|---------|---------------|
| **NanoFuse** | Open Source | 0.x (Phase 1 Complete) | MIT | Minimalist, snapshot/resume, learning |
| **K7** | Open Source | Beta | Apache 2.0 | Kubernetes-based, AI agents, security |
| **SlicerVM** | Commercial | Production | Proprietary | Turnkey, GPU support, enterprise |

**Key Insight**: These systems occupy distinct niches and are **not direct competitors**. They solve different problems for different audiences using different architectural philosophies.

---

## Table of Contents

1. [Architecture Comparison](#architecture-comparison)
2. [Technology Stack](#technology-stack)
3. [Installation & Deployment](#installation--deployment)
4. [Networking](#networking)
5. [Storage & Persistence](#storage--persistence)
6. [Security Model](#security-model)
7. [Snapshot & Resume](#snapshot--resume)
8. [Image Management](#image-management)
9. [API & Interfaces](#api--interfaces)
10. [Performance Characteristics](#performance-characteristics)
11. [Scalability](#scalability)
12. [GPU & Hardware Support](#gpu--hardware-support)
13. [Operational Complexity](#operational-complexity)
14. [Cost Model](#cost-model)
15. [Documentation & Learning](#documentation--learning)
16. [Use Cases & Target Audience](#use-cases--target-audience)
17. [Pros & Cons](#pros--cons)
18. [Decision Framework](#decision-framework)
19. [Future Roadmap](#future-roadmap)
20. [Conclusion](#conclusion)

---

## Architecture Comparison

### NanoFuse: Direct Control Architecture

```
┌──────────────┐         ┌─────────────────────────────┐
│ nanofuse CLI │────────▶│   nanofused API Daemon      │
│   (Go)       │  HTTP   │        (Go)                 │
│              │  Unix   │ - REST API server           │
│              │  Socket │ - VM lifecycle management   │
│              │         │ - SQLite state persistence  │
└──────────────┘         │ - Firecracker orchestration │
                         └────────────┬────────────────┘
                                      │
                         ┌────────────▼────────────────┐
                         │  Firecracker Processes      │
                         │  (direct process mgmt)      │
                         └─────────────────────────────┘
```

**Philosophy**: Minimalist - direct Firecracker process management with just what's needed, nothing more.

**Layers**: 2 (API daemon → Firecracker)

**Components**:
- `nanofuse` CLI (Go binary)
- `nanofused` daemon (Go binary)
- SQLite database (state)
- Firecracker processes (VMs)

### K7: Kubernetes Platform Architecture

```
┌─────────┐     ┌──────────────┐     ┌────────────────┐
│ k7 CLI  │────▶│  k7 API      │────▶│  Kubernetes    │
│(Python) │     │ (Python/     │     │   (K3s)        │
└─────────┘     │  Docker)     │     └────────┬───────┘
                └──────────────┘              │
                                              │
                                    ┌─────────▼──────────┐
                                    │ Kata Containers    │
                                    │ (container→VM)     │
                                    └─────────┬──────────┘
                                              │
                                    ┌─────────▼──────────┐
                                    │  Firecracker VMM   │
                                    │  (actual VMs)      │
                                    └────────────────────┘
```

**Philosophy**: Platform leverage - use Kubernetes to orchestrate Kata-wrapped Firecracker VMs.

**Layers**: 4 (API → Kubernetes → Kata → Firecracker)

**Components**:
- `k7` CLI (Python)
- K7 API (Python, containerized)
- Kubernetes (K3s distribution)
- Kata Containers runtime
- Firecracker VMM
- Devmapper snapshotter + thin-pool LVM
- containerd with Kata runtime
- Jailer (security, noted as not working correctly)

### SlicerVM: Commercial Black Box Architecture

```
┌──────────────┐         ┌─────────────────────────────┐
│ SlicerVM CLI │────────▶│   SlicerVM Platform         │
│   (?)        │  REST   │   (Proprietary)             │
│              │   API   │                              │
└──────────────┘         └────────────┬────────────────┘
                                      │
                         ┌────────────▼────────────────┐
                         │  Firecracker +              │
                         │  Cloud Hypervisor           │
                         │  (multi-VMM support)        │
                         └─────────────────────────────┘
```

**Philosophy**: Commercial polish - hide complexity, provide turnkey solution.

**Layers**: Unknown (proprietary)

**Components**: Not fully disclosed, but includes:
- REST API
- Multiple VMM support (Firecracker, Cloud Hypervisor)
- Multiple storage backends (ZFS, Devmapper, NVMe)

### Architecture Analysis

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Complexity** | Low | High | Hidden |
| **Abstraction Layers** | Minimal | Heavy | Unknown |
| **Dependencies** | Few | Many | Unknown |
| **Control Level** | Direct | Platform-mediated | Vendor-controlled |
| **Extensibility** | High (source access) | High (K8s ecosystem) | Low (proprietary) |
| **Transparency** | Complete | Complete | Limited |

---

## Technology Stack

### NanoFuse Stack

**Core Language**: Go (100%)

**Dependencies**:
- Firecracker microVM
- SQLite (state persistence)
- Linux kernel (KVM support)
- Standard Go libraries (net/http, database/sql)

**Networking**:
- Custom TAP device management
- Custom NAT implementation
- Custom IPAM (IP Address Management)
- Bridge mode (planned Phase 2)

**Storage**:
- SQLite database (/var/lib/nanofuse/nanofuse.db)
- OCI image format (rootfs.ext4 + vmlinux + manifest.json)
- Filesystem-based snapshots

**Build Tools**:
- Go toolchain
- Make
- GitHub Actions (CI/CD)

### K7 Stack

**Core Languages**: Python (CLI, API, core logic) + Bash (Ansible playbooks)

**Major Dependencies**:
- Kubernetes (K3s distribution)
- Kata Containers 3.x
- Firecracker VMM
- containerd with Kata runtime
- Devmapper snapshotter
- LVM (thin-pool provisioning)
- Ansible (installation automation)
- Docker (API deployment)

**Networking**:
- Kubernetes CNI
- NetworkPolicies (egress control)
- Future: Cilium (FQDN-based whitelisting)

**Storage**:
- Kubernetes etcd (state)
- Devmapper + thin-pool LVM (VM disks)
- Container images (standard OCI)

**Build Tools**:
- Python packaging (PyPI)
- Debian packaging (.deb)
- Ansible playbooks
- Docker

### SlicerVM Stack

**Core Languages**: Unknown (proprietary)

**Known Dependencies**:
- Firecracker VMM
- Cloud Hypervisor (for GPU)
- ZFS or Devmapper (storage backends)

**Networking**: Unknown (proprietary)

**Storage**:
- ZFS (optional)
- Devmapper (optional)
- NVMe optimization

### Stack Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Primary Language** | Go | Python | Unknown |
| **Runtime Dependencies** | Minimal | Heavy | Unknown |
| **Package Manager** | Binary distribution | apt (PPA), pip (PyPI) | Proprietary installer |
| **Container Runtime** | None (direct Firecracker) | containerd + Kata | Unknown |
| **Orchestration** | Custom daemon | Kubernetes (K3s) | Unknown |

---

## Installation & Deployment

### NanoFuse Installation

**Prerequisites**:
- Linux host (Ubuntu, Debian, etc.)
- KVM support (`/dev/kvm` accessible)
- x86_64 architecture (ARM64 coming soon)
- Root access for networking setup

**Installation Methods**:

1. **Pre-built binaries** (Recommended):
```bash
VERSION=v0.1.0
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofuse
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofused
chmod +x nanofuse nanofused
sudo mv nanofuse nanofused /usr/local/bin/
```

2. **Build from source**:
```bash
git clone https://github.com/daax-dev/nanofuse.git
cd nanofuse
make build
sudo make install
```

**Setup Time**: ~5 minutes (download + install)

**Configuration**: Minimal (daemon starts with defaults)

**Post-Install**:
```bash
# Start daemon (manual or systemd)
sudo nanofused
# OR
sudo systemctl enable --now nanofused
```

### K7 Installation

**Prerequisites**:
- Ubuntu (amd64 or arm64)
- Hardware virtualization (`/dev/kvm`)
- **Raw disk** (unformatted, unpartitioned) for thin-pool
- Ansible installed
- Docker and Docker Compose

**Installation**:
```bash
# Add K7 repository
sudo add-apt-repository ppa:katakate.org/k7
sudo apt update
sudo apt install k7

# Install all components (automated via Ansible)
k7 install
# OR specify disk explicitly
k7 install --disk /dev/nvme2n1
```

**Setup Time**: ~1-2 minutes (automated but installs: K3s, Kata, Firecracker, Jailer, devmapper)

**Configuration**: Automated by Ansible playbook

**Post-Install**:
- May require logout/login for group membership changes
- Kubernetes cluster ready
- Kata runtime configured

### SlicerVM Installation

**Prerequisites**: (Based on documentation)
- Broad hardware compatibility (Raspberry Pi to servers)
- Various platforms supported

**Installation**: Not fully disclosed (commercial product, likely turnkey)

**Setup Time**: "Minutes" (claimed)

### Installation Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Complexity** | Very Low | Medium-High | Low (automated) |
| **Time Required** | ~5 min | ~1-2 min (automated) | ~minutes |
| **Components Installed** | 2 binaries | 10+ components | Unknown |
| **Hardware Requirements** | Standard Linux | Dedicated raw disk | Flexible |
| **Automation Level** | Manual | High (Ansible) | High |
| **Rollback Difficulty** | Easy | Moderate | Unknown |

---

## Networking

### NanoFuse Networking

**Architecture**: Custom implementation in Go

**Modes**:
1. **NAT Mode** (Default, Phase 1 complete)
   - Each VM gets TAP device (tap0, tap1, ...)
   - Host NATs traffic to internet
   - VMs isolated from each other
   - IP range: 172.16.0.0/24 (configurable)
   - IPAM (IP Address Management) custom-built

2. **Bridge Mode** (Phase 2, planned)
   - VMs connect to Linux bridge
   - Inter-VM communication possible
   - Requires pre-configured bridge on host

**Performance**:
- Sub-millisecond host-to-VM latency (achieved)

**Implementation**:
- Custom TAP device management (`internal/network/tap.go`)
- Custom bridge setup (`internal/network/bridge.go`)
- Custom NAT setup (`internal/network/nat.go`)
- Custom IPAM (`internal/network/ipam.go`)

**Configuration**:
```json
{
  "network": {
    "mode": "nat",
    "bridge_name": null,
    "tap_device": null,  // auto-assigned
    "mac_address": null  // auto-generated
  }
}
```

### K7 Networking

**Architecture**: Leverages Kubernetes NetworkPolicies

**Security Model**:
- **Default-deny**: Complete isolation between VMs
- **Ingress**: All inter-VM communication blocked
- **Egress**: CIDR-based whitelisting

**Configuration**:
```yaml
egress_whitelist:
  - "10.0.0.5/32"     # Your private proxy/gateway
```

**Limitations**:
- DNS blocked when egress locked down (only IP/CIDR reachable)
- No domain-based whitelisting (yet)

**Roadmap**:
- Cilium integration for FQDN-based DNS whitelisting

**Administrative Access**:
- `kubectl exec` works (uses K8s API, not pod networking)
- `k7 shell` preserved

**Implementation**: Standard Kubernetes NetworkPolicy resources

### SlicerVM Networking

**Architecture**: Unknown (proprietary)

**Features** (documented):
- NFS file sharing between VMs
- Nested virtualization support
- SSH and Serial Over SSH (SOS)
- Large-scale scenarios (7000+ pods)

### Networking Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Implementation** | Custom (Go) | Kubernetes CNI | Proprietary |
| **Inter-VM Communication** | Phase 2 (bridge mode) | Blocked by default | Supported (NFS) |
| **Egress Control** | Via iptables (planned) | NetworkPolicies | Unknown |
| **DNS Support** | Standard | Blocked w/ egress control | Unknown |
| **Latency** | Sub-millisecond | Depends on CNI | Unknown |
| **Complexity** | Low (direct control) | Medium (K8s networking) | Hidden |

---

## Storage & Persistence

### NanoFuse Storage

**State Database**: SQLite (`/var/lib/nanofuse/nanofuse.db`)

**Schema**:
- `vms` table - VM state and configuration
- `snapshots` table - Snapshot metadata
- `images` table - Pulled image metadata
- `image_pull_jobs` table - Async pull job tracking

**Benefits**:
- ACID transactions (prevents race conditions)
- Single file (easy backup: copy .db)
- Queryable (SQL for complex queries)
- Excellent Go support

**Limitations**:
- Single-node only (no distributed state)

**Images**:
- Location: `/var/lib/nanofuse/images/sha256:<hash>/`
- Format: OCI-compatible
  - `rootfs.ext4` - Block device image
  - `boot/vmlinux` - Uncompressed kernel
  - `manifest.json` - Metadata

**Snapshots**:
- Location: `/var/lib/nanofuse/snapshots/{vm-id}/{snapshot-id}/`
- Files: `mem.snap` (memory), `vm.snap` (device state)

**Disk Efficiency**: Standard filesystem (no thin provisioning)

### K7 Storage

**State Database**: Kubernetes etcd (distributed key-value store)

**VM Disks**: Devmapper snapshotter + thin-pool LVM

**Thin-Pool Provisioning**:
- Requires dedicated raw disk
- LVM manages logical volumes
- Copy-on-write snapshots
- Efficient disk usage across dozens of VMs

**Images**:
- Standard container images (Docker/OCI)
- Any registry (Docker Hub, GHCR, etc.)

**Disk Efficiency**: Excellent (thin provisioning, snapshots)

**Setup**:
```bash
k7 install --disk /dev/nvme2n1
```

### SlicerVM Storage

**Options**:
- ZFS (snapshots, compression, deduplication)
- Devmapper
- Local NVMe optimization

**Flexibility**: Multiple backend choices

### Storage Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **State Store** | SQLite | Kubernetes etcd | Unknown |
| **Distributed State** | No | Yes (etcd) | Unknown |
| **Disk Provisioning** | Standard FS | Thin-pool LVM | ZFS or Devmapper |
| **Disk Efficiency** | Standard | Excellent | Excellent (ZFS) |
| **Backup Simplicity** | Very easy (.db file) | Complex (etcd backup) | Unknown |
| **Hardware Requirements** | None | Dedicated raw disk | Flexible |

---

## Security Model

### NanoFuse Security

**Isolation Layers**:
1. **Firecracker VM isolation** (KVM hardware virtualization)
2. **Filesystem permissions** (Unix socket: `root:nanofuse`, mode `0660`)
3. **Resource limits** (prevent exhaustion)

**Access Control**:
- Unix socket permissions (Phase 1)
- No authentication on socket
- Future: TCP + Bearer token or mTLS (Phase 2+)

**Concurrency**:
- Pessimistic locking prevents race conditions
- Lock timeout: 5 minutes

**Resource Limits** (configurable):
- `max_vms`: 50 (default)
- `max_total_memory_mib`: 32768 MB
- `max_vcpus_per_vm`: 8
- `max_memory_per_vm_mib`: 8192 MB
- `max_snapshot_storage_gib`: 100 GB

**Error Handling**:
- Expose real constraints (not security through obscurity)
- Clear error messages with remediation hints

**Network Isolation**:
- NAT mode isolates VMs by default
- Bridge mode allows controlled inter-VM communication

### K7 Security

**Isolation Layers** (Defense-in-Depth):
1. **Firecracker VM isolation** (KVM)
2. **Kata Containers** (container → VM boundary)
3. **Jailer** (chroot jail for Firecracker) - *Note: Currently not working correctly*
4. **Kubernetes security features**

**Kata Security**:
- Seccomp restrictions enabled
- VM-level isolation for containers

**Linux Capabilities**:
- All capabilities dropped by default (`drop: ALL`)
- Explicitly add back only needed capabilities via `cap_add`
- `allow_privilege_escalation: false` (always)
- Seccomp profile: `RuntimeDefault`

**Non-Root Execution**:
- `container_non_root`: Run container as UID 65532
- `pod_non_root`: Run entire pod as non-root (UID/GID/FSGroup 65532)

**Network Security**:
- Complete network isolation by default
- Ingress: Blocked between VMs
- Egress: CIDR-based whitelisting
- DNS blocked when egress locked down

**API Security**:
- API keys required
- SHA256 hashed storage
- Timing-attack-resistant comparison
- Expiry enforcement
- Last-used timestamp tracking
- File permissions: `/etc/k7/api_keys.json` mode `600`

**Known Issues**:
- Jailer functionality not working correctly (under investigation)
- Kubernetes secrets may conflict with Jailer

### SlicerVM Security

**Isolation**: Firecracker VM isolation (hardware-level)

**Additional Features**: Unknown (proprietary)

**Commercial Advantage**: Presumably enterprise-grade security with professional review

### Security Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Isolation Layers** | 1 (Firecracker) | 4 (K8s+Kata+Jailer+Firecracker) | Unknown |
| **Authentication** | None (Unix perms) | API keys (SHA256) | Unknown |
| **Network Isolation** | NAT (default) | NetworkPolicies | Unknown |
| **Capability Dropping** | N/A | Yes (all dropped) | Unknown |
| **Non-Root Execution** | N/A | Supported | Unknown |
| **Security Maturity** | Basic | Advanced | Commercial-grade |
| **Known Issues** | None | Jailer not working | Unknown |

**Verdict**: K7 has the most sophisticated security model, but with caveats (Jailer issue). NanoFuse is simple but effective. SlicerVM likely enterprise-grade but opaque.

---

## Snapshot & Resume

### NanoFuse Snapshot/Resume

**Strategy**: Manual, user-controlled snapshots

**Primary Goal**: Sub-2-second cold starts (core requirement)

**Status**: Phase 2 (implementation in progress)

**API**:
```bash
# Create snapshot
POST /vms/{id}/snapshots
{
  "name": "after-boot"  // optional
}

# Resume from snapshot
POST /vms/{id}/resume
{
  "snapshot_id": "snapshot-20251030-100530"  // optional
}
```

**Storage**:
- Location: `/var/lib/nanofuse/snapshots/{vm-id}/{snapshot-id}/`
- Files: `mem.snap`, `vm.snap`
- Metadata in SQLite

**Retention**: Manual (user manages deletion)

**Philosophy**:
- User control over when to snapshot (not automatic)
- Defers automatic policies (following "selling options" principle)
- Snapshots survive API restarts, host reboots

**Future Enhancements** (roadmap):
- Automatic retention policies (keep last N, delete older than X days)
- Snapshot-on-shutdown option
- Cron-style snapshot schedules

### K7 Snapshot/Resume

**Status**: Not yet implemented (on roadmap)

**Roadmap Item**: "Add pause/resume/fork/clone support for sandboxes"

**Planned Features**:
- Pause/resume (basic)
- Fork (create multiple instances from one snapshot)
- Clone (copy VM state)

**Advantages**:
- Fork/clone enable interesting multi-instance scenarios
- Leverage Kata/Firecracker snapshot capabilities

**Challenges**:
- Kubernetes doesn't have native snapshot/resume for pods
- May require custom CRDs or controllers

### SlicerVM Snapshot/Resume

**Status**: Not explicitly documented

**Focus**: Rapid deployment rather than snapshot/resume

**Observation**: Kubernetes cluster boot "in ~1 minute" suggests starting from scratch, not snapshots

### Snapshot Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Status** | Phase 2 (in progress) | Roadmap | Unknown |
| **Strategy** | Manual, user-controlled | Planned: fork/clone | Not emphasized |
| **Cold Start Goal** | <2 seconds | Not stated | Not primary focus |
| **Retention** | Manual | TBD | Unknown |
| **Persistence** | Files on disk | TBD | Unknown |
| **Priority** | High (core feature) | Medium (roadmap) | Low (not mentioned) |

**Verdict**: NanoFuse prioritizes snapshots most heavily. K7 has ambitious plans (fork/clone). SlicerVM focuses elsewhere.

---

## Image Management

### NanoFuse Images

**Format**: Custom OCI-compatible format

**Structure**:
```
/rootfs.ext4         # Block device with Ubuntu 24.04 + systemd
/boot/vmlinux        # Uncompressed kernel (Slicer 5.10.240)
/manifest.json       # Metadata (kernel cmdline, arch, etc.)
```

**Build Process**:
```dockerfile
FROM ubuntu:24.04
# Install systemd, openssh-server, networking
# Copy proven Slicer 5.10.240 kernel
# Configure systemd units (enable, not start)
# Generate rootfs.ext4
```

**Distribution**:
- Push to GHCR (GitHub Container Registry)
- Tags: `:latest`, `:v1.0.0`, `:sha-abc123`
- Pull via CLI: `nanofuse image pull ghcr.io/owner/image:tag`

**Kernel Strategy**:
- Bundle kernel with rootfs (atomic versioning)
- Use proven Slicer 5.10.240 kernel initially
- Defer custom kernel compilation to future (Phase 5+)

**Pull Process**:
- Asynchronous with job tracking
- Job states: pending, in_progress, completed, failed
- Progress reporting (bytes, percentage)

**CLI**:
```bash
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:latest
nanofuse image list
```

### K7 Images

**Format**: Standard container images (Docker/OCI)

**Source**: Any container registry
- Docker Hub
- GHCR
- Private registries

**Example**:
```yaml
image: alpine:latest
# OR
image: ubuntu:24.04
# OR
image: ghcr.io/owner/custom:tag
```

**Customization**:
```yaml
before_script: |
  apk add --no-cache git curl
  git clone https://github.com/owner/repo
```

**Environment**:
```yaml
env_file: path/to/.env
```

**Advantages**:
- No custom image format needed
- Leverage existing container ecosystem
- Easy to use familiar images

**Disadvantages**:
- Container → VM translation via Kata adds complexity
- Kernel managed separately by Kata

### SlicerVM Images

**Format**: Proprietary (not disclosed)

**Availability**: Likely pre-built images for common use cases

**Details**: Insufficient information

### Image Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Format** | Custom OCI + kernel | Standard containers | Proprietary |
| **Kernel** | Bundled in image | Kata-managed | Unknown |
| **Registry** | Any OCI (GHCR) | Any container registry | Unknown |
| **Build Process** | Dockerfile → OCI | Standard Docker build | Unknown |
| **Customization** | Build new image | before_script + env | Unknown |
| **Size Overhead** | +30MB (kernel) | Standard container | Unknown |

**Verdict**: NanoFuse has atomic kernel+rootfs versioning. K7 leverages standard containers. SlicerVM proprietary.

---

## API & Interfaces

### NanoFuse API

**Protocol**: RESTful HTTP/1.1 with JSON

**Transport**: Unix socket (`/var/run/nanofused.sock`)

**Future**: TCP binding with Bearer token or mTLS (Phase 2+)

**Versioning**: None in 0.x (will add `/v1/` at 1.0.0)

**Endpoints**:
- `GET /vms` - List VMs
- `POST /vms` - Create VM
- `GET /vms/{id}` - Get VM status
- `DELETE /vms/{id}` - Delete VM
- `POST /vms/{id}/start` - Start VM
- `POST /vms/{id}/stop` - Stop VM
- `POST /vms/{id}/snapshots` - Create snapshot
- `GET /images` - List images
- `POST /images/pull` - Pull image (async)
- `GET /images/jobs/{id}` - Check pull job status

**Error Responses**:
```json
{
  "error": {
    "code": "RESOURCE_LIMIT_EXCEEDED",
    "message": "Cannot create VM: Would exceed max_total_memory_mib limit",
    "details": {
      "requested_memory_mib": 512,
      "current_memory_mib": 32500,
      "limit_memory_mib": 32768
    }
  }
}
```

**Version Checking**:
- CLI sends: `User-Agent: nanofuse-cli/0.1.0`
- API responds: `X-API-Version: 0.1.0`
- Mismatch warnings

**CLI**:
```bash
nanofuse vm up my-vm --image ghcr.io/owner/base:latest
nanofuse vm status my-vm
nanofuse vm stop my-vm
```

### K7 API

**Protocol**: RESTful (likely FastAPI/Flask)

**Transport**: HTTPS via Cloudflared (SSL support)

**Authentication**: API keys (SHA256 hashed, timing-attack resistant)

**Deployment**: Docker container

**Management**:
```bash
k7 start-api
k7 generate-api-key my-key1
```

**Python SDK** (Synchronous):
```python
from katakate import Client

k7 = Client(
    endpoint='https://<your-endpoint>',
    api_key='your-key'
)

# Create sandbox
sb = k7.create({
    "name": "my-sandbox",
    "image": "alpine:latest"
})

# Execute code
result = sb.exec('echo "Hello World"')
print(result['stdout'])

# List, delete
sandboxes = k7.list()
sb.delete()
```

**Python SDK** (Asynchronous):
```python
from katakate import AsyncClient

async def main():
    k7 = AsyncClient(
        endpoint='https://<your-endpoint>',
        api_key='your-key'
    )
    print(await k7.list())
    await k7.aclose()
```

**CLI**:
```bash
k7 create -f k7.yaml
k7 list
k7 delete my-sandbox
k7 delete-all
```

### SlicerVM API

**Protocol**: REST (details not disclosed)

**Features**: Programmatic VM management

**CLI**: Available but details limited

### API Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Protocol** | REST/HTTP/1.1 | REST/HTTPS | REST |
| **Transport** | Unix socket (TCP future) | HTTPS | Unknown |
| **Authentication** | None (Unix perms) | API keys | Unknown |
| **Versioning** | None (0.x), /v1/ at 1.0 | Not stated | Unknown |
| **SDK** | None (HTTP client) | Python (sync+async) | Unknown |
| **Error Design** | Excellent (detailed) | Unknown | Unknown |
| **SSL Support** | Phase 2+ | Yes (Cloudflared) | Likely |

**Verdict**: NanoFuse has clean REST design. K7 has mature API with Python SDK. SlicerVM likely polished but opaque.

---

## Performance Characteristics

### NanoFuse Performance

**Cold Start Target**: Sub-2-second (via snapshots)

**Network Latency**: Sub-millisecond host-to-VM (achieved)

**Boot Time**: Firecracker typical (~125ms, not including OS boot)

**Memory Overhead**: Minimal
- Direct Firecracker processes
- No container runtime overhead
- Go daemon is lightweight

**Disk I/O**: Standard filesystem (no thin provisioning overhead)

**CPU Overhead**: Minimal
- Direct process management
- No scheduler overhead

**Optimization Focus**: Snapshot/resume for cold start performance

### K7 Performance

**Boot Time**: Firecracker fast, but Kubernetes + Kata adds overhead
- Pod scheduling latency
- Kata container setup
- containerd operations

**Memory Overhead**: Higher
- Kubernetes components (K3s is lighter than full K8s)
- Kata runtime
- containerd

**Disk I/O**: Excellent via thin-pool provisioning
- Copy-on-write performance
- Devmapper snapshotter efficiency

**CPU Overhead**: Moderate
- Kubernetes scheduler
- Kata runtime
- NetworkPolicy evaluation

**Optimization Focus**: Disk efficiency and isolation at scale (dozens of VMs per node)

### SlicerVM Performance

**Kubernetes Boot**: ~1 minute for full cluster

**CPU Performance**: Desktop CPUs up to 5.5GHz bursts with NVMe

**Scale**: 7000+ pods reproducible in minutes

**GPU**: Supported via Cloud Hypervisor + VFIO

**Optimization Focus**: Large-scale scenario reproduction and GPU workloads

### Performance Comparison

| Metric | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Cold Start** | <2s (snapshot goal) | Not optimized for | Not primary focus |
| **Network Latency** | Sub-millisecond | CNI-dependent | Unknown |
| **Boot Overhead** | Minimal | Moderate (K8s) | Unknown |
| **Memory Efficiency** | High (direct) | Lower (K8s overhead) | Unknown |
| **Disk Efficiency** | Standard | Excellent (thin-pool) | Excellent (ZFS) |
| **CPU Overhead** | Minimal | Moderate | Unknown |
| **GPU Support** | No (Phase 5+) | No (roadmap) | Yes |

**Verdict**: NanoFuse optimizes for cold starts. K7 optimizes for disk efficiency and security. SlicerVM optimizes for scale and GPU.

---

## Scalability

### NanoFuse Scalability

**Architecture**: Single-node (deliberately)

**Resource Limits** (defaults, configurable):
- Max VMs: 50
- Max total memory: 32GB
- Max vCPUs per VM: 8
- Max memory per VM: 8GB

**Vertical Scaling**: Increase limits on larger hosts

**Horizontal Scaling**: Not supported (no cluster mode)

**State**: SQLite (single-node only)

**Future** (Phase 5+):
- Distributed orchestration (if demand emerges)
- Would require etcd/Consul for state
- Trade-off accepted: Simplicity now vs. scalability later

**Philosophy**: Start simple, add distribution only if needed

### K7 Scalability

**Architecture**: Single K3s node currently, multi-node on roadmap

**Per-Node Capacity**: Dozens of VMs via thin-pool provisioning

**Multi-Node** (roadmap):
- Kubernetes naturally supports clustering
- K8s scheduler handles cross-node placement
- NetworkPolicies work across nodes
- Cross-node snapshot mobility planned

**State**: Kubernetes etcd (distributed, eventually consistent)

**Advantages**:
- Built on K8s = inherently scalable architecture
- Future multi-node is natural extension

**Current Limitation**: Single-node only (as of October 2025)

### SlicerVM Scalability

**Deployment Targets**: Raspberry Pi to servers

**Large-Scale Scenarios**: 7000+ pods demonstrated

**Multi-Machine**: Appears supported (details unclear)

**Commercial Grade**: Likely has enterprise scalability features

### Scalability Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Single-Node Capacity** | 50 VMs (default) | Dozens (thin-pool) | Unknown |
| **Multi-Node Support** | No (Phase 5+ maybe) | Roadmap | Unknown |
| **State Distribution** | No (SQLite) | Yes (etcd, when multi-node) | Unknown |
| **Scaling Strategy** | Vertical only | Horizontal (future) | Likely horizontal |
| **Scalability Path** | Uncertain | Clear (K8s) | Commercial-grade |

**Verdict**: K7 has clearest scalability path. NanoFuse deliberately simple. SlicerVM likely enterprise-scale.

---

## GPU & Hardware Support

### NanoFuse GPU Support

**Current**: None (Firecracker lacks PCI passthrough)

**Roadmap**: Cloud Hypervisor for GPU (Phase 5+)

**Challenge**: Would require VMM change (Firecracker → Cloud Hypervisor)

**Priority**: Low (not needed for Trigger.dev use case)

### K7 GPU Support

**Current**: Not implemented

**Roadmap**: "Support other VMM such as Qemu for GPU workloads"

**Approach**: Add QEMU as alternative to Firecracker

**Advantage**: Kata supports multiple VMMs
- Firecracker for CPU-only workloads
- QEMU for GPU workloads
- In same cluster

**Challenge**: Managing multiple VMM types

### SlicerVM GPU Support

**Current**: Supported via Cloud Hypervisor + VFIO passthrough

**Use Cases**: AI/LLM workloads (Ollama mentioned)

**Maturity**: Production feature

**Advantage**: Works today

### GPU Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Current Support** | No | No | Yes |
| **Roadmap** | Phase 5+ (Cloud Hypervisor) | Planned (QEMU) | N/A |
| **Implementation** | Not specified | Multiple VMMs | Cloud Hypervisor + VFIO |
| **Priority** | Low | Medium | High (feature complete) |
| **Use Cases** | Future consideration | AI/ML (planned) | AI/LLM (active) |

**Verdict**: SlicerVM only system with GPU support today. K7 and NanoFuse have roadmap plans.

---

## Operational Complexity

### NanoFuse Operations

**Day-to-Day Commands**:
```bash
# VM management
nanofuse vm up my-vm --image ghcr.io/owner/base:latest
nanofuse vm status my-vm
nanofuse vm stop my-vm

# Image management
nanofuse image pull ghcr.io/owner/base:latest
nanofuse image list

# Daemon management
sudo systemctl status nanofused
sudo systemctl restart nanofused
```

**Logging**:
```bash
# View daemon logs
journalctl -u nanofused

# Filter errors
journalctl -u nanofused -p err

# Query JSON logs
journalctl -u nanofused | jq 'select(.vm_id=="550e8400-...")'
```

**State Inspection**:
```bash
# SQLite queries (if needed)
sqlite3 /var/lib/nanofuse/nanofuse.db "SELECT * FROM vms;"
```

**Debugging**:
- Console logs: `/var/lib/nanofuse/vms/{id}/console.log`
- Clear error messages with remediation hints
- Structured JSON logs

**Required Skills**:
- Linux system administration
- systemd
- journalctl
- Basic SQLite (optional)

**Complexity**: Low (standard Unix tooling)

### K7 Operations

**Day-to-Day Commands**:
```bash
# Sandbox management
k7 create -f k7.yaml
k7 list
k7 delete my-sandbox

# API management
k7 start-api
k7 generate-api-key my-key1

# Kubernetes access
kubectl get pods
kubectl describe pod my-sandbox
kubectl exec -it my-sandbox -- /bin/bash
kubectl logs my-sandbox
```

**Monitoring**:
- Kubernetes metrics (kubectl top)
- Potential Prometheus integration
- K8s dashboard (if deployed)

**Debugging**:
```bash
# Pod inspection
kubectl describe pod my-sandbox
kubectl logs my-sandbox
kubectl exec my-sandbox -- command

# Node inspection
kubectl describe node
kubectl top node

# Events
kubectl get events
```

**Required Skills**:
- Kubernetes (kubectl, resources, debugging)
- Linux system administration
- Ansible (for installation)
- Docker (for API)

**Complexity**: Medium-High (requires K8s knowledge)

### SlicerVM Operations

**Interface**: Likely polished CLI/UI (commercial product)

**Support**: Professional support available

**Complexity**: Low for users (complexity hidden)

### Operations Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **CLI Simplicity** | Very simple | Moderate | Likely polished |
| **Logging** | journalctl + JSON | kubectl logs | Unknown |
| **Debugging Tools** | Unix standard | Kubernetes tools | Vendor tools |
| **Required Expertise** | Linux admin | K8s + Linux | Minimal (support) |
| **Learning Curve** | Moderate | Steep | Gentle |
| **Operational Overhead** | Low | Medium | Low (vendor-managed) |

**Verdict**: NanoFuse simple Unix ops. K7 requires K8s expertise. SlicerVM likely easiest for end-users.

---

## Cost Model

### NanoFuse Costs

**Software License**: Free (MIT)

**Hardware**:
- Linux host with KVM
- Standard storage (no special requirements)

**Cloud Hosting**:
- Any provider with nested virtualization
- Example: AWS `.metal` instances, Hetzner dedicated, GCP with flag

**Operational Costs**:
- Self-managed (no vendor fees)
- Requires Linux admin skills

**Support**: Community only (GitHub issues/discussions)

**Development**:
- MIT license = fork, modify, commercialize freely
- No attribution requirement (permissive)

**TCO Estimate**: $0 software + infrastructure + self-support time

**Best For**:
- Developers who can self-support
- Small deployments
- Learning and experimentation

### K7 Costs

**Software License**: Free (Apache 2.0)

**Hardware**:
- Linux host with KVM
- **Dedicated raw disk required** (thin-pool)
- More demanding than NanoFuse

**Cloud Hosting**:
- Bare metal or dedicated instances
- Nested virtualization required
- Example: Hetzner Robot instances

**Operational Costs**:
- Higher complexity (K8s knowledge needed)
- Infrastructure overhead (K8s components consume resources)

**Support**:
- Community (GitHub)
- Email: hi@katakate.org
- No commercial support (yet)

**Development**:
- Apache 2.0 = fork, modify, commercialize with patent grant
- Must include license notice

**TCO Estimate**: $0 software + higher infrastructure + K8s expertise

**Best For**:
- Teams with Kubernetes expertise
- AI/ML workloads needing security
- Organizations comfortable with beta software

### SlicerVM Costs

**Software License**:
- **Home Edition**: Low cost (pricing not disclosed)
- **Commercial**: $250+ USD per seat

**Hardware**: Flexible (Pi to servers)

**Cloud Hosting**: Supported on various platforms

**Operational Costs**: Lower (polished UX reduces admin burden)

**Support**: Professional support included with commercial license

**Updates**: Vendor-managed (presumably included)

**TCO Estimate**: $250/seat + infrastructure + reduced operational overhead

**Best For**:
- Enterprises valuing professional support
- Teams wanting turnkey solutions
- Hobbyists (Home Edition)

### Cost Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Software Cost** | $0 | $0 | $250+/seat |
| **License** | MIT | Apache 2.0 | Proprietary |
| **Hardware Cost** | Standard | Higher (raw disk) | Flexible |
| **Operational Cost** | Low-Medium | Medium-High | Low (supported) |
| **Support Cost** | $0 (community) | $0 (community) | Included |
| **Vendor Lock-In** | None | None | Yes |
| **Total 5-Year TCO** | Infrastructure + time | Infrastructure + K8s expertise | $1250+ + infrastructure |

**Verdict**: NanoFuse cheapest for DIY. K7 free but higher complexity cost. SlicerVM paid but lower operational burden.

---

## Documentation & Learning

### NanoFuse Documentation

**Quality**: Exceptional for 0.x project

**Key Documents**:
- `README.md` - Project overview, quick start
- `ARCHITECTURE_DECISIONS.md` - Comprehensive architectural rationale (920 lines!)
- `API_CONTRACT.md` - REST API specification
- `CLI_SPEC.md` - CLI interface specification
- `CONTRIBUTING.md` - Contribution guidelines
- Multiple implementation guides and reports

**Architectural Rigor**:
- Every decision documented with rationale
- Trade-offs explicitly acknowledged
- Gregor Hohpe principles applied throughout
- Option theory, abstraction design, golden path

**Learning Value**:
- Outstanding for understanding microVM systems
- Shows "why" not just "what"
- Architectural education

**End-User Docs**: Minimal (focused on developers/contributors)

**Target Audience**: Developers wanting to understand deeply

### K7 Documentation

**Quality**: Good for getting started

**Key Documents**:
- `README.md` - Comprehensive quick start
- `ROADMAP.md` - Future plans and priorities
- Tutorial: LangChain ReAct agent integration
- Hetzner setup guide (PDF)
- API usage examples

**Focus**:
- Task-oriented (how to use)
- Production-focused
- Less architectural depth

**Learning Value**: Good for learning AI agent integration

**External Docs**: docs.katakate.org

**Target Audience**: AI/ML practitioners, DevOps engineers

### SlicerVM Documentation

**Quality**: Professional

**Location**: docs.slicervm.com

**Focus**: Use cases and features

**Transparency**: Limited (proprietary implementation)

**Learning Value**: Good for using SlicerVM, not for understanding internals

**Target Audience**: End-users, infrastructure teams

### Documentation Comparison

| Aspect | NanoFuse | K7 | SlicerVM |
|--------|----------|-----|----------|
| **Quality** | Exceptional | Good | Professional |
| **Architectural Depth** | Very High | Low-Medium | Low (proprietary) |
| **Practical Guides** | Minimal | Good | Good |
| **Learning Value** | Outstanding (theory) | Good (practice) | Limited (usage) |
| **Target Audience** | Developers/learners | Practitioners | End-users |
| **Transparency** | Complete | Complete | Limited |

**Verdict**: NanoFuse best for learning architecture. K7 best for practical AI integration. SlicerVM best for usage.

---

## Use Cases & Target Audience

### NanoFuse Use Cases

**Primary Use Case**: Trigger.dev self-hosting
- Dual-environment workloads (web + worker VMs)
- Fast cold starts for ephemeral jobs
- Isolated execution environments

**Secondary Use Cases**:
- Learning Firecracker and microVM internals
- Fast-booting development environments
- Snapshot-based testing (restore to known state)
- Edge computing (future)
- Multi-tenant platforms (future)

**Target Audience**:
- Individual developers
- Small teams
- Technical learners
- Architects studying microVM systems
- Open source contributors

**Ideal User Profile**:
- Comfortable with Go and Linux
- Values simplicity over features
- Wants to understand how things work
- Needs specific use case (like Trigger.dev)
- Comfortable with early-stage software

**Not Ideal For**:
- Enterprise production (yet - 0.x phase)
- GPU workloads
- Multi-node clustering
- Users needing commercial support

### K7 Use Cases

**Primary Use Case**: AI agents executing untrusted code
- LangChain ReAct agents
- Code execution sandboxes
- AI-powered automation
- Safe arbitrary code execution at scale

**Secondary Use Cases**:
- Custom serverless platforms (self-hosted Fargate alternative)
- Hardened CI/CD runners (avoid Docker-in-Docker risks)
- Blockchain execution layers for AI dApps
- Secure development environments

**Target Audience**:
- AI/ML practitioners
- Platform engineers
- Security-conscious teams
- DevOps teams with Kubernetes expertise
- Web3 + AI developers

**Ideal User Profile**:
- Knows Kubernetes (kubectl, resources, debugging)
- Needs strong isolation for untrusted code
- Building AI agent platforms
- Comfortable with beta software
- Values security over simplicity

**Not Ideal For**:
- Kubernetes beginners
- Production workloads requiring maturity (Jailer issue)
- GPU workloads (not yet supported)
- Simple use cases (too much complexity)

### SlicerVM Use Cases

**Primary Use Cases**:
- Kubernetes testing and development (7000+ pod scenarios)
- Chaos engineering and network isolation testing
- Customer support (rapid environment reproduction)
- AI/LLM workloads with GPU acceleration
- Cost optimization (replace expensive cloud VMs)

**Secondary Use Cases**:
- Home labs and learning environments
- Rapid prototyping
- Development/testing infrastructure

**Target Audience**:
- Kubernetes developers
- Infrastructure teams
- DevOps professionals
- AI/ML teams (GPU workloads)
- Enterprises
- Hobbyists (Home Edition)

**Ideal User Profile**:
- Values turnkey solutions
- Needs GPU support today
- Wants professional support
- Budget for commercial licensing
- Focus on use, not implementation

**Not Ideal For**:
- Open source purists
- Budget-constrained projects
- Users wanting to modify internals
- Learning-focused projects (less transparency)

### Use Case Comparison Matrix

| Use Case | NanoFuse | K7 | SlicerVM |
|----------|----------|-----|----------|
| **AI Agents (Untrusted Code)** | ⚠️ Basic isolation | ✅ Purpose-built | ⚠️ Possible |
| **Fast Cold Starts (<2s)** | ✅ Core feature | ❌ Not optimized | ❌ Not focus |
| **GPU Workloads** | ❌ Not yet | ❌ Not yet | ✅ Supported |
| **Kubernetes Testing** | ❌ Not designed for | ⚠️ Could work | ✅ Designed for |
| **Learning/Education** | ✅ Exceptional docs | ⚠️ K8s complexity | ❌ Limited info |
| **Production Enterprise** | ⚠️ 0.x early | ⚠️ Beta (Jailer issue) | ✅ Commercial-grade |
| **CI/CD Runners** | ⚠️ Possible | ✅ Hardened | ⚠️ Possible |
| **Edge Computing** | ⚠️ Future | ❌ Too heavy | ⚠️ Possible |
| **Multi-Tenant SaaS** | ⚠️ Future | ✅ Security model | ✅ Likely |

---

## Pros & Cons

### NanoFuse

**Pros:**
✅ **Exceptional architectural documentation** - Learn microVM systems deeply
✅ **Maximum simplicity** - Just two binaries, minimal dependencies
✅ **Direct control** - No abstraction layers, full transparency
✅ **Sub-2s cold starts** - Snapshot/resume focus (Phase 2)
✅ **Sub-millisecond network latency** - Demonstrates technical excellence
✅ **MIT license** - Maximum freedom (fork, modify, commercialize)
✅ **Clean REST API** - Standard, extensible, well-designed
✅ **Low operational overhead** - Standard Unix tools (systemd, journalctl)
✅ **Easy backup** - Single SQLite file
✅ **Clear error messages** - Actionable, detailed
✅ **Lightweight** - Go binaries, minimal memory footprint

**Cons:**
❌ **Early maturity (0.x)** - Not production-ready yet
❌ **Limited features** - Snapshot/resume in Phase 2, bridge mode in Phase 2
❌ **No authentication** - Unix socket permissions only (TCP + auth in Phase 2)
❌ **Single-node only** - No clustering support (may never add)
❌ **No GPU support** - Phase 5+ (distant future)
❌ **No SDK** - Must use raw HTTP client
❌ **Community support only** - No commercial backing
❌ **Learning curve** - Need Linux admin skills
❌ **Custom image format** - Not standard containers

**Best For**: Developers who value simplicity, need fast cold starts, want to learn deeply, and can self-support.

### K7

**Pros:**
✅ **Strong security model** - Defense-in-depth (Kata + Firecracker + K8s)
✅ **Purpose-built for AI agents** - Untrusted code execution
✅ **Kubernetes integration** - Full ecosystem access
✅ **Thin-pool disk efficiency** - Dozens of VMs per node
✅ **Python SDK** - Sync + async, modern patterns
✅ **LangChain integration** - Tutorial and examples
✅ **Automated installation** - Ansible handles complexity
✅ **Network isolation** - NetworkPolicies, egress control
✅ **Apache 2.0 license** - Permissive with patent grant
✅ **Active community** - Show HN #1, growing traction
✅ **Standard container images** - Use existing Docker images
✅ **API key authentication** - Secure from start

**Cons:**
❌ **High complexity** - Kubernetes, Kata, Firecracker, containerd, devmapper
❌ **Steep learning curve** - Requires K8s knowledge
❌ **Beta maturity** - Known issues (Jailer not working)
❌ **Hardware requirements** - Dedicated raw disk for thin-pool
❌ **Higher resource overhead** - K8s components consume resources
❌ **Single-node currently** - Multi-node on roadmap
❌ **No snapshot/resume yet** - Roadmap feature
❌ **No GPU support yet** - Roadmap feature
❌ **DNS blocked with egress control** - Until Cilium integration
❌ **Complex troubleshooting** - Many components, many failure points

**Best For**: Teams with K8s expertise, AI agent platforms, security-focused deployments, comfort with beta software.

### SlicerVM

**Pros:**
✅ **Production maturity** - Commercial product, battle-tested
✅ **GPU support today** - Cloud Hypervisor + VFIO
✅ **Professional support** - Commercial backing
✅ **Turnkey installation** - Polished user experience
✅ **Broad compatibility** - Raspberry Pi to servers
✅ **Multiple storage backends** - ZFS, Devmapper, NVMe
✅ **Proven at scale** - 7000+ pod scenarios
✅ **Kubernetes-inside-VM** - Full K8s for testing
✅ **NFS file sharing** - Inter-VM communication
✅ **Professional documentation** - Polished, use-case focused
✅ **Nested virtualization** - Cloud provider support
✅ **Home Edition** - Affordable for hobbyists

**Cons:**
❌ **Commercial licensing** - $250+/seat for production
❌ **Proprietary** - Cannot inspect or modify internals
❌ **Vendor lock-in** - Dependency on vendor
❌ **Limited transparency** - Implementation details hidden
❌ **Less hackable** - Cannot contribute or customize
❌ **No open source community** - Vendor-controlled
❌ **Unknown architecture** - Less learning value
❌ **Licensing costs** - Ongoing expense

**Best For**: Enterprises, teams needing GPU support, production maturity, professional support, and turnkey solutions.

---

## Decision Framework

### Choose **NanoFuse** if:

✅ You need **maximum simplicity** and minimal dependencies
✅ **Sub-2-second cold starts** are critical (snapshot/resume)
✅ You're building a **specific application** (like Trigger.dev)
✅ You want to **learn Firecracker** and microVM internals deeply
✅ You prefer **direct control** over abstractions
✅ You have **Go expertise** or want to contribute to a Go project
✅ **Single-node deployment** is sufficient for your needs
✅ You value **architectural clarity** and exceptional documentation
✅ **MIT license** is important (maximum freedom)
✅ You're comfortable with **early-stage software** (0.x)
✅ You can **self-support** (community only)
✅ **Fast snapshots** matter more than features

❌ **Don't choose NanoFuse if:**
- You need production maturity today
- GPU support is required
- Kubernetes integration is needed
- Commercial support is important
- Multi-node clustering is required

### Choose **K7** if:

✅ You need to **execute untrusted code at scale** (AI agents, sandboxes)
✅ You already have **Kubernetes expertise**
✅ **Security is paramount** (defense-in-depth required)
✅ You need **network isolation** with egress controls
✅ **Disk efficiency** is important (many VMs, thin-pool provisioning)
✅ **Python ecosystem** is primary (Python SDK, LangChain)
✅ You want **automated installation** despite complexity
✅ Future **multi-node clustering** is likely needed
✅ You're comfortable with **beta software** and known issues
✅ **Apache 2.0 license** fits your requirements
✅ **AI agent security** is the primary use case
✅ You value **battle-tested stack** (K8s + Kata + Firecracker)

❌ **Don't choose K7 if:**
- You don't know Kubernetes
- Production maturity is critical (Jailer issue)
- Simplicity is more important than features
- GPU support is needed now
- Hardware constraints (no dedicated disk)

### Choose **SlicerVM** if:

✅ You need **GPU support today** (AI/LLM workloads)
✅ **Professional support** is required
✅ **Production maturity** is critical (can't use beta)
✅ **Budget allows** for licensing costs ($250+/seat)
✅ **Kubernetes testing at scale** is the use case
✅ **Polished UX** and turnkey installation are important
✅ You prefer **proven commercial solutions** over open source experiments
✅ **Hardware flexibility** is needed (Pi to servers)
✅ You don't need to **modify internals**
✅ **Vendor support** adds value to your team
✅ **Time-to-value** matters more than cost
✅ **Risk aversion** (commercial backing reduces risk)

❌ **Don't choose SlicerVM if:**
- Budget is constrained (open source required)
- You need to inspect/modify code
- Vendor lock-in is unacceptable
- Learning internals is important
- Open source community contribution is desired

### Quick Decision Tree

```
START
  |
  ├─ Need GPU now? ──YES──> SlicerVM
  |
  ├─ Need commercial support? ──YES──> SlicerVM
  |
  ├─ Know Kubernetes?
  |    ├─ YES ──> Running AI agents? ──YES──> K7
  |    |                               └NO───> Consider use case
  |    └─ NO  ──> Learn K8s or choose simpler option
  |
  ├─ Need sub-2s cold starts? ──YES──> NanoFuse
  |
  ├─ Want to learn microVMs deeply? ──YES──> NanoFuse
  |
  ├─ Value simplicity over features? ──YES──> NanoFuse
  |
  ├─ Need production maturity now? ──YES──> SlicerVM
  |
  └─ Security-focused AI workloads? ──YES──> K7
```

---

## Future Roadmap

### NanoFuse Roadmap

**Current Status**: Phase 1 Complete (Core Components)

**Phase 2** (Next):
- ✅ Snapshot/resume implementation (in progress)
- ✅ Bridge networking mode
- ✅ Advanced logging and metrics (Prometheus)

**Phase 3** (Production Hardening):
- Auto-restart policies (systemd units per VM)
- Snapshot retention policies
- Health checks and monitoring
- VM crash recovery automation

**Phase 4** (Multi-tenancy):
- Namespace isolation
- Per-tenant quotas
- Image signing and verification

**Phase 5+** (Advanced):
- Custom kernel compilation (learning exercise)
- GPU support via Cloud Hypervisor
- Distributed orchestration (if needed)
- TCP API with authentication
- Web UI

**Philosophy**: Each phase builds on previous, no backtracking. Keep options open.

### K7 Roadmap

**Current Focus**: Core stability and foundational runtime improvements

**Completed**:
- ✅ `--disk` argument for explicit thin-pool disk
- ✅ DNS blocking for security
- ✅ ARM support (amd64 + arm64)

**In Progress**:
- ⏳ Pause/resume/fork/clone support for sandboxes
- ⏳ Fix Jailer functionality (known issue)
- ⏳ Multi-node support

**Next Goals**:
- Docker build/run/compose inside VM sandboxes (major feature!)
- Cilium networking integration
- Docker pull deny/whitelist

**Future Work**:
- QEMU support (macOS ARM, GPU)
- Cross-node snapshot mobility
- AppArmor integration
- CI/CD and deployment tests

**Advanced Features**:
- TEE (Trusted Execution Environment) support
- Custom rootfs support (lighter images)

### SlicerVM Roadmap

**Status**: Production-ready (commercial product)

**Roadmap**: Not publicly disclosed (proprietary)

**Likely Focus**:
- Continued feature enhancements
- Enterprise feature requests
- Security updates
- Performance optimizations

### Roadmap Comparison

| Feature | NanoFuse | K7 | SlicerVM |
|---------|----------|-----|----------|
| **Snapshot/Resume** | Phase 2 (active) | Roadmap | Unknown |
| **Multi-Node** | Phase 5+ (maybe) | Roadmap (active) | Unknown |
| **GPU Support** | Phase 5+ | Roadmap (QEMU) | ✅ Available |
| **Docker-in-VM** | Future | Roadmap (major) | Unknown |
| **Web UI** | Phase 5+ | Not mentioned | Likely has |
| **Kubernetes** | Phase 5+ (maybe) | ✅ Core | Works inside VM |
| **TEE Support** | Not planned | Roadmap | Unknown |
| **Transparency** | Complete | Complete | None |

---

## Conclusion

### Summary of Findings

NanoFuse, K7, and SlicerVM represent **three distinct approaches** to the same problem: running isolated workloads efficiently using Firecracker microVMs.

**NanoFuse** embodies the **minimalist philosophy**: Build exactly what's needed, document every decision, maintain simplicity. It's an architectural gem for developers who want to understand microVM systems deeply while having a production-capable tool.

**K7** embodies the **platform leverage philosophy**: Combine battle-tested components (Kubernetes + Kata + Firecracker) to create a secure, scalable foundation for AI agent workloads. Accept complexity in exchange for enterprise features and security depth.

**SlicerVM** embodies the **commercial polish philosophy**: Hide complexity behind a turnkey interface, provide professional support, and deliver production-ready features (like GPU support) today. Trade transparency for maturity and support.

### Key Differentiators

| Differentiator | NanoFuse | K7 | SlicerVM |
|----------------|----------|-----|----------|
| **Architecture** | Direct Firecracker | Kubernetes → Kata → Firecracker | Proprietary |
| **Complexity** | Low | High | Hidden |
| **Maturity** | 0.x (early) | Beta | Production |
| **Target Use** | Fast cold starts | AI agent safety | GPU workloads, K8s testing |
| **License** | MIT | Apache 2.0 | Proprietary ($) |
| **Learning Value** | Exceptional | Good | Limited |
| **Production Ready** | Not yet | Soon (Jailer issue) | Yes |

### Philosophical Positioning

```
Simplicity ◄────────────────────────────────► Features
   │                                              │
NanoFuse                                     SlicerVM
   │                                              │
   └─────────────── K7 ─────────────────────────┘
                    │
             Security-Focused
```

### Recommendations by Scenario

**For Individual Developers / Small Teams:**
- **Choose NanoFuse** if you value simplicity, need fast cold starts, and want to learn
- **Choose K7** if you have K8s expertise and need AI agent security
- **Choose SlicerVM Home Edition** if you want turnkey experience

**For Startups:**
- **Choose NanoFuse** for MVP if use case aligns (snapshot-based workloads)
- **Choose K7** if building AI agent platform (security focus)
- **Choose SlicerVM** if need GPU and budget allows

**For Enterprises:**
- **Choose K7** if have K8s expertise and need defense-in-depth security
- **Choose SlicerVM** if need production maturity, GPU, and professional support
- **Consider NanoFuse** only if willing to contribute/support yourself

**For Learning / Education:**
- **Choose NanoFuse** (exceptional architectural documentation)
- K7 good for learning Kubernetes + Kata integration
- SlicerVM not ideal (limited transparency)

### The Ecosystem Perspective

The microVM ecosystem is **healthier with all three approaches**:

- **NanoFuse** pushes for simplicity and architectural rigor
- **K7** demonstrates platform leverage and security-first design
- **SlicerVM** sets the bar for production maturity and features

They're not competing for the same users - they serve different needs. A developer learning Firecracker uses NanoFuse. A team building an AI code execution platform uses K7. An enterprise running GPU workloads uses SlicerVM.

### Final Thought

All three systems prove that **Firecracker is the foundation** for the next generation of lightweight virtualization. The divergence in approach above Firecracker shows the richness of the problem space:

- **Direct control** (NanoFuse) offers clarity and simplicity
- **Platform leverage** (K7) offers ecosystem and scale
- **Commercial polish** (SlicerVM) offers maturity and support

Choose based on your priorities: **Simplicity vs. Features vs. Support**.

---

## Appendix: Technical Specifications

### System Requirements Comparison

| Requirement | NanoFuse | K7 | SlicerVM |
|-------------|----------|-----|----------|
| **OS** | Linux (Ubuntu, Debian, etc.) | Ubuntu (amd64/arm64) | Linux (broad) |
| **CPU Architecture** | x86_64 (ARM64 planned) | x86_64, ARM64 | x86_64, ARM64 |
| **Virtualization** | KVM required (`/dev/kvm`) | KVM required | KVM required |
| **Memory** | Host-dependent | Host-dependent + K8s overhead | Host-dependent |
| **Storage** | Standard filesystem | Dedicated raw disk (thin-pool) | Flexible |
| **Network** | Standard NIC | Standard NIC | Standard NIC |
| **Special Hardware** | None | Raw disk | None (GPU optional) |

### Version Information (as of October 2025)

| System | Current Version | Status | Last Update |
|--------|----------------|--------|-------------|
| **NanoFuse** | 0.x (Phase 1 complete) | Active development | October 2025 |
| **K7** | Beta | Active development | October 2025 |
| **SlicerVM** | Unknown (commercial) | Production | Unknown |

### Links & Resources

**NanoFuse**:
- Repository: `https://github.com/daax-dev/nanofuse`
- Documentation: `https://github.com/daax-dev/nanofuse/docs`
- Issues: `https://github.com/daax-dev/nanofuse/issues`
- License: MIT

**K7 (Katakate)**:
- Repository: `https://github.com/Katakate/k7`
- Documentation: `https://docs.katakate.org`
- Website: `https://katakate.org`
- PyPI: `https://pypi.org/project/katakate/`
- Contact: `hi@katakate.org`
- License: Apache 2.0

**SlicerVM**:
- Documentation: `https://docs.slicervm.com`
- Pricing: Home Edition (affordable), Commercial ($250+ per seat)
- License: Proprietary

### Acknowledgments

This comparison was conducted in October 2025 based on:
- NanoFuse repository and documentation (October 2025)
- K7 repository, README, and docs (October 2025)
- SlicerVM public documentation (October 2025)

All three projects leverage:
- **Firecracker** by AWS: `https://github.com/firecracker-microvm/firecracker`
- **Kata Containers** (K7): `https://katacontainers.io`
- **Cloud Hypervisor** (SlicerVM GPU): `https://www.cloudhypervisor.org`

---

**Document Version**: 1.0
**Date**: October 31, 2025
**Author**: Technical analysis based on project documentation and source code
**Status**: Comprehensive comparison complete

*This document represents a snapshot in time. All three projects are under active development and features/status may change.*
