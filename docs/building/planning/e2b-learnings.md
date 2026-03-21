# E2B Technical Analysis and Learnings for NanoFuse

**Date**: 2025-11-26
**Purpose**: Deep technical analysis of E2B repositories to evaluate features for adoption into NanoFuse
**Research Methodology**: Web search, documentation analysis, repository structure analysis

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [E2B Organization Overview](#e2b-organization-overview)
3. [Core Architecture](#core-architecture)
4. [Infrastructure Components](#infrastructure-components)
5. [Template and Image Building](#template-and-image-building)
6. [Snapshot and Resume System](#snapshot-and-resume-system)
7. [Filesystem Architecture](#filesystem-architecture)
8. [Networking](#networking)
9. [Security Model](#security-model)
10. [SDK and API Design](#sdk-and-api-design)
11. [Code Interpreter Architecture](#code-interpreter-architecture)
12. [Desktop Sandbox](#desktop-sandbox)
13. [Self-Hosting Requirements](#self-hosting-requirements)
14. [Feature Comparison with NanoFuse](#feature-comparison-with-nanofuse)
15. [Recommendations for NanoFuse](#recommendations-for-nanofuse)
16. [Sources](#sources)

---

## Executive Summary

E2B is an open-source infrastructure for running AI-generated code in secure, isolated Firecracker microVM sandboxes. Key metrics:

- **Boot time**: ~125-150ms (leveraging Firecracker + snapshots)
- **Memory overhead**: <5MB per microVM
- **Density**: 4,000+ microVMs per server
- **Session duration**: Up to 24 hours with pause/resume
- **Market adoption**: Used by ~50% of Fortune 500, millions of sandboxes per week
- **Funding**: $21M Series A (2025)

### Key Takeaways for NanoFuse

1. **OverlayFS is critical** for scaling - E2B uses it to share read-only rootfs across thousands of instances
2. **Dual-protocol API** (REST + gRPC) provides flexibility for different operation types
3. **Template system** with Dockerfile → snapshot conversion eliminates cold starts
4. **Envd daemon** running inside each VM provides consistent SDK interaction model
5. **Jailer-based security** adds container-like isolation on top of VM isolation

---

## E2B Organization Overview

### Repository Structure

| Repository | Stars | Language | Purpose |
|------------|-------|----------|---------|
| **E2B** | 10k+ | MDX/TS | Main SDK and CLI |
| **infra** | 732 | Go (84.7%) | Core infrastructure powering E2B Cloud |
| **code-interpreter** | 2.1k | MDX | Python/JS SDK for code execution |
| **dashboard** | 110 | TypeScript | Management UI (Next.js 15 + Supabase) |
| **fc-kernels** | - | HCL (64.2%) | Custom kernel build pipeline |
| **desktop** | - | TypeScript | Desktop sandbox for computer use |
| **fragments** | 6k | TypeScript | Reference app for AI-generated UIs |
| **surf** | 644 | TypeScript | Computer use AI agent demo |

### Technology Stack

**Infrastructure:**
- Go for backend services
- Terraform (HCL) for infrastructure-as-code
- Nomad for workload orchestration
- Consul for service discovery
- GCP (primary), AWS (in progress)

**Virtualization:**
- Firecracker microVMs
- KVM for hardware acceleration
- Custom kernel builds (5.10, 6.1)

---

## Core Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────────────┐
│                        E2B Cloud Platform                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┐    ┌────────────────┐    ┌──────────────────┐   │
│  │  REST API    │    │ Edge Controller│    │   Orchestrator   │   │
│  │  (Lifecycle) │◄──►│  (Routing)     │◄──►│  (Scheduling)    │   │
│  └──────────────┘    └────────────────┘    └────────┬─────────┘   │
│                                                      │              │
│  ┌──────────────────────────────────────────────────┼────────────┐ │
│  │                    Nomad Job Scheduler            │            │ │
│  └──────────────────────────────────────────────────┼────────────┘ │
│                                                      │              │
│  ┌──────────────────────────────────────────────────▼────────────┐ │
│  │                   Firecracker MicroVMs                        │ │
│  │  ┌──────────────────────────────────────────────────────────┐ │ │
│  │  │  ┌─────────┐    ┌─────────┐    ┌─────────┐              │ │ │
│  │  │  │  envd   │    │  envd   │    │  envd   │   ...        │ │ │
│  │  │  │ (gRPC)  │    │ (gRPC)  │    │ (gRPC)  │              │ │ │
│  │  │  ├─────────┤    ├─────────┤    ├─────────┤              │ │ │
│  │  │  │ Sandbox │    │ Sandbox │    │ Sandbox │              │ │ │
│  │  │  │  VM 1   │    │  VM 2   │    │  VM 3   │              │ │ │
│  │  │  └─────────┘    └─────────┘    └─────────┘              │ │ │
│  │  └──────────────────────────────────────────────────────────┘ │ │
│  └───────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### Component Details

**1. Orchestrator**
- Responsible for managing sandboxes and their lifecycle
- Schedules VMs on appropriate nodes based on available capacity
- Can optionally run the template builder component
- Written in Go

**2. Edge Controller**
- Routes traffic to sandboxes
- Exposes API for cluster management
- Provides gRPC proxy for orchestrator communication

**3. Envd (Environment Daemon)**
- Runs inside EVERY sandbox (port 49983)
- Provides SDK interaction via gRPC and HTTP
- Manages filesystem, process, and terminal operations
- Key services: filesystem service, process service, logging/exporter
- Source: `github.com/e2b-dev/infra/packages/envd`

**4. Instance Management Service**
- Oversees sandbox creation, monitoring, termination
- Maintains routing catalog (Redis-backed)

---

## Infrastructure Components

### Storage Layer

| Component | Purpose | Technology |
|-----------|---------|------------|
| PostgreSQL | Template metadata, build records, snapshot info | Managed/Self-hosted |
| ClickHouse | Analytics and structured logging | Managed/Self-hosted |
| Redis | Routing catalog mapping sandboxes to nodes | Managed/Self-hosted |
| GCS/S3 | Templates, snapshots, kernels, binaries | Object storage |

### Compute Resources (GCP)

| Instance Type | Purpose |
|---------------|---------|
| n2-standard-8 | Client clusters running Firecracker VMs |
| e2-medium | Server clusters for coordination |
| n2-standard-2 | API clusters for API service |

### Service Discovery

- Consul provides service mesh and discovery
- Nomad handles job scheduling
- Automatic clustering via Consul integration

---

## Template and Image Building

### Build Pipeline

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Dockerfile  │────►│ Docker Build │────►│  FC Snapshot │────►│   Deploy     │
│ e2b.Dockerfile│    │   (Image)    │     │  Conversion  │     │  (Template)  │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

### Build Process Details

1. **Docker Image Creation**
   ```dockerfile
   FROM e2bdev/code-interpreter:latest
   RUN pip install cowsay pandas numpy
   ```

2. **Snapshot Creation**
   - CLI pushes Docker image to E2B cloud
   - Image converted to Firecracker microVM format
   - Optional start command executes (20-second default readiness window)
   - System snapshots the running VM state
   - Snapshot includes: filesystem + running processes

3. **Template Configuration** (e2b.toml)
   ```toml
   [template]
   id = "my-template"
   cpu = 2
   memory_mb = 2048
   ```

### Key CLI Commands

```bash
e2b template init      # Generate basic Dockerfile
e2b template build     # Build and convert to microVM
e2b template build -c "python -m jupyter notebook"  # With start command
```

### Build System 2.0 (Newer Approach)

- No Dockerfiles required
- "Just write code"
- Build step runs automatically on template execution
- Simplifies developer experience

---

## Snapshot and Resume System

### Snapshot Mechanics

**Creation:**
- Full VM state serialized (filesystem + running processes + memory)
- Stored in object storage (GCS/S3)
- Associated with template ID

**Restoration:**
- ~150ms load time
- Entire VM state restored
- Processes continue from snapshotted state

### Pause/Resume Functionality

- Sandboxes can be paused for up to 24 hours
- Preserves:
  - Complete filesystem state
  - Running processes
  - Installed packages
  - Memory state

### Cold Start Elimination

**Traditional approach:**
1. Boot VM (~2-5 seconds)
2. Initialize services
3. Ready for use

**E2B approach:**
1. Restore snapshot (~150ms)
2. Immediately ready

### Known Snapshot Limitations

From Firecracker documentation:
- Network connections may not survive snapshot/resume
- vsock connections closed on snapshot (listen sockets remain)
- Occasional process stuck issues after resume (reported bug)
- Page fault handling options: kernel-based or userfaultfd

### Page Fault Handling Options

1. **Host OS handling** (default): Kernel handles page faults on resume
2. **Userfaultfd**: Dedicated userspace process for page fault handling (better control)

---

## Filesystem Architecture

### OverlayFS Implementation

**Critical for scaling** - Without this, copying rootfs for each VM is prohibitive.

```
┌────────────────────────────────────────────────────────────────┐
│                     Merged View (VM sees this)                 │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌──────────────────────────────┐  ┌───────────────────────┐  │
│  │     Upper Directory          │  │    Lower Directory     │  │
│  │    (Per-instance, RW)        │  │   (Shared, RO)        │  │
│  │                              │  │                        │  │
│  │  - Modified files            │  │  - Base rootfs        │  │
│  │  - New files                 │  │  - Ubuntu 24.04       │  │
│  │  - Deleted file markers      │  │  - System packages    │  │
│  └──────────────────────────────┘  └───────────────────────┘  │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Implementation Steps

**1. Create Read-Only Base (Squashfs)**
```bash
mksquashfs /tmp/rootfs rootfs.ext4 -noappend
```

**2. Create Overlay Init Script**
```bash
#!/bin/bash
# overlay-init script (runs before real init)
rw_root="/overlay/rw"
work_dir="/overlay/work"
mount -o noatime,lowerdir=/,upperdir=${rw_root},workdir=${work_dir} -t overlay overlay /
exec /sbin/init
```

**3. Per-Instance Writable Layer**
```bash
# Create sparse ext4 image (only uses space as needed)
dd if=/dev/zero of=overlay.ext4 bs=1M count=5120
mkfs.ext4 overlay.ext4
```

**4. Firecracker Configuration**
```json
{
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "/path/to/rootfs.ext4",
      "is_read_only": true
    },
    {
      "drive_id": "overlay",
      "path_on_host": "/path/to/overlay.ext4",
      "is_read_only": false
    }
  ],
  "boot-args": "init=/sbin/overlay-init overlay_root=/dev/vdb"
}
```

### Benefits

- **Space savings**: Single rootfs shared across thousands of VMs
- **Speed**: No rootfs copying at VM start
- **Copy-on-write**: Only modified files consume upper layer space
- **RAM option**: `overlay_root=ram` for non-persistent instances

---

## Networking

### TAP Device Setup

Standard Firecracker TAP networking:

```bash
# Create TAP device
sudo ip tuntap add tapA mode tap

# Configure IP
sudo ip addr add 10.0.0.29/30 dev tapA

# Bring up
sudo ip link set tapA up

# Enable IP forwarding
echo 1 > /proc/sys/net/ipv4/ip_forward
```

### Network Isolation

- Each VM gets dedicated TAP device
- Host-level iptables for isolation
- Rate limiting available via Firecracker

### Snapshot/Resume Networking Notes

From Firecracker docs:
- Network packet loss expected on resume in different process
- Network connection state may not survive
- Recommend re-establishing connections after resume

---

## Security Model

### Multi-Layer Security Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Security Layers                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Layer 1: KVM Hardware Virtualization                               │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  - Hardware-level isolation                                 │   │
│  │  - Each VM has own kernel                                   │   │
│  │  - Prevents cross-tenant attacks                            │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  Layer 2: Firecracker Jailer                                        │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  - Cgroups: CPU, memory, I/O limits                         │   │
│  │  - Namespaces: PID, network, mount isolation                │   │
│  │  - Seccomp: 35 allowed syscalls (30 simple, 5 filtered)     │   │
│  │  - Privilege dropping: Root → unprivileged user             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  Layer 3: Firecracker Minimalism                                    │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  - ~50,000 lines of Rust (vs QEMU's 2M lines of C)          │   │
│  │  - Minimal device emulation                                 │   │
│  │  - No legacy hardware support                               │   │
│  │  - No PCIe (no GPU passthrough)                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Jailer Configuration Details

**Seccomp Filters:**
- Configured in `vmm/src/default_syscalls/filters.rs`
- 35 allowed system calls total
- 30 with simple filtering
- 5 with parameter-based filtering

**Cgroups Resource Control:**
- CPU usage limits
- Memory consumption caps
- I/O bandwidth throttling

**Namespace Isolation:**
- Complete isolation from host
- Separation from other VMs
- Independent network stack

### Host-Level Mitigations (Production)

Required for multi-tenant deployments:
- Disable SMT (Simultaneous Multi-Threading)
- Verify KPTI support
- Disable KSM (Kernel Samepage Merging)
- Apply L1TF mitigations
- Apply SSBD mitigations
- Use Rowhammer-resistant memory
- Secure swap configuration

### Why VMs Over Containers

From security analysis:
> "In Linux, namespaces, cgroups and seccomp filters work together to create the appearance of containment. But they all run on top of the same kernel. If that kernel is compromised, those boundaries collapse."

Firecracker provides VM-level isolation with container-like performance.

---

## SDK and API Design

### Dual Protocol Architecture

| Protocol | Use Case | Authentication |
|----------|----------|----------------|
| REST API | Sandbox lifecycle (create, kill, timeout) | API key |
| gRPC | Real-time operations (filesystem, commands, terminals) | Access token |

### REST API Endpoints

**Sandbox Lifecycle:**
```
POST   /sandboxes                    # Create sandbox
DELETE /sandboxes/{id}               # Kill sandbox
PATCH  /sandboxes/{id}               # Update timeout
GET    /events/sandboxes             # Lifecycle events
GET    /events/sandboxes/{id}        # Specific sandbox events
```

**Lifecycle Event Types:**
- `sandbox.lifecycle.created`
- `sandbox.lifecycle.updated`
- `sandbox.lifecycle.paused`
- `sandbox.lifecycle.resumed`
- `sandbox.lifecycle.killed`

### SDK Interfaces

**JavaScript/TypeScript:**
```typescript
import { Sandbox } from '@e2b/code-interpreter'

const sbx = await Sandbox.create()
await sbx.runCode('x = 1')
const execution = await sbx.runCode('x+=1; x')
console.log(execution.text)  // outputs 2
await sbx.close()
```

**Python:**
```python
from e2b_code_interpreter import Sandbox

with Sandbox.create() as sandbox:
    sandbox.run_code("x = 1")
    execution = sandbox.run_code("x+=1; x")
    print(execution.text)  # outputs 2
```

### Available SDK Versions

| SDK | Latest Version |
|-----|----------------|
| CLI | v2.4.2 |
| JavaScript SDK | v2.8.1 |
| Python SDK | v2.8.0 |
| Desktop JS SDK | v2.2.0 |
| Desktop Python SDK | v2.2.0 |
| Code Interpreter JS SDK | v2.3.1 |
| Code Interpreter Python SDK | v2.4.1 |

### Envd Internal API

**HTTP API** (`/packages/envd/api`):
- File operations
- Process management
- Legacy compatibility

**gRPC API** (`/packages/shared/pkg/grpc/envd`):
- `filesystemconnect`: Bidirectional file streaming
- `processconnect`: Process execution and management
- Connect protocol for bidirectional communication

---

## Code Interpreter Architecture

### Core Design

```
┌──────────────────────────────────────────────────────────────┐
│                    Code Interpreter Sandbox                   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   Jupyter Server                        │ │
│  │  ┌──────────────────────────────────────────────────┐  │ │
│  │  │  Python Kernel (ipykernel)                       │  │ │
│  │  │  - Context sharing between executions            │  │ │
│  │  │  - Variable persistence                          │  │ │
│  │  │  - Import caching                                │  │ │
│  │  └──────────────────────────────────────────────────┘  │ │
│  │  ┌──────────────────────────────────────────────────┐  │ │
│  │  │  Additional Kernels                              │  │ │
│  │  │  - JavaScript (nodejs)                           │  │ │
│  │  │  - R                                             │  │ │
│  │  │  - Java                                          │  │ │
│  │  │  - Bash                                          │  │ │
│  │  └──────────────────────────────────────────────────┘  │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   FastAPI Server                        │ │
│  │  - Envd integration                                    │ │
│  │  - SDK interaction                                     │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Jupyter Kernel Protocol

- Implements Jupyter Kernel messaging protocol (partial)
- Result structure similar to ipython kernel
- Supports multiple data types: text, images, plots, HTML

### Context Persistence

Key feature: Variables, imports, and functions persist between executions within a session.

```python
# First execution
sandbox.run_code("import pandas as pd")
sandbox.run_code("df = pd.DataFrame({'a': [1,2,3]})")

# Second execution - context preserved
execution = sandbox.run_code("df.sum()")
print(execution.text)  # Works because df is still defined
```

### Custom Template Setup

```dockerfile
FROM e2bdev/code-interpreter:latest

# Add packages
RUN pip install jupyter-server ipykernel ipython pandas numpy

# Configure Jupyter
RUN jupyter server --generate-config
```

---

## Desktop Sandbox

### Purpose

Enables "computer use" for AI agents - LLMs can interact with full desktop GUI environment.

### Architecture

- Complete Ubuntu desktop environment
- VNC/streaming for visual output
- Mouse/keyboard control via SDK
- Application automation

### Capabilities

**Streaming:**
```typescript
const stream = await desktop.stream.start()
// Returns URL for viewing/controlling desktop
```

**Input Control:**
```typescript
await desktop.click(x, y)           // Mouse click
await desktop.drag(x1, y1, x2, y2)  // Drag operation
await desktop.scroll(delta)          // Scroll
await desktop.type("text", speed)    // Keyboard input
```

**Application Management:**
```typescript
await desktop.launch("chromium-browser")
await desktop.launch("code")  // VS Code
const windowId = await desktop.getWindow("chromium")
```

### Use Cases

- Browser automation
- GUI application testing
- Visual AI agent feedback loops
- Computer use demonstrations (Surf, Open Computer Use)

---

## Self-Hosting Requirements

### Prerequisites

- GCP account (primary), AWS (in progress)
- Go 1.24.7+
- Linux host with KVM support
- Docker
- Terraform
- gcloud CLI

### Infrastructure Requirements

1. **Compute**
   - Client clusters (n2-standard-8) for VMs
   - Server clusters (e2-medium) for coordination
   - API clusters (n2-standard-2) for API service

2. **Storage**
   - PostgreSQL (template metadata)
   - ClickHouse (analytics)
   - Redis (routing)
   - GCS buckets (artifacts)

3. **Networking**
   - VPC configuration
   - Firewall rules
   - Load balancer for API

### Deployment Process

```bash
# Build services
make build-and-upload

# Deploy infrastructure
terraform apply

# Verify deployment
make health-check
```

### Environment Configuration

```bash
# Switch environments
make switch-env ENV=dev

# Environment files
.env.dev
.env.staging
.env.production
```

### Estimated Costs

Self-hosting trade-offs:
- **Pro**: Predictable costs (pay for infrastructure, not usage)
- **Pro**: Data sovereignty
- **Con**: 5-10 person infrastructure team needed
- **Con**: 6-12 months to build equivalent

---

## Feature Comparison with NanoFuse

### Current NanoFuse Status

| Feature | NanoFuse | E2B |
|---------|----------|-----|
| Boot time | Sub-2-second | ~150ms |
| Snapshot/Resume | Planned (Phase 2) | Full support |
| OverlayFS | Not implemented | Production use |
| SDK | CLI only | Python + JS + CLI |
| API Protocol | REST only | REST + gRPC |
| Template System | Docker images | Dockerfile → Snapshot |
| In-VM Daemon | None | envd |
| Desktop/GUI | Not planned | Full support |
| Code Interpreter | Not planned | Full Jupyter support |
| Orchestration | Manual | Nomad + Consul |
| Multi-tenancy | Single user | Enterprise scale |

### Alignment with NanoFuse Goals

**High Alignment:**
1. Firecracker-based microVMs
2. Fast boot times
3. Snapshot/resume for cold start elimination
4. OCI-compatible image management
5. CLI + API architecture

**Different Focus:**
1. E2B targets AI code execution; NanoFuse targets Trigger.dev workloads
2. E2B is multi-tenant SaaS; NanoFuse is self-hosted
3. E2B has extensive SDK; NanoFuse is simpler

---

## Recommendations for NanoFuse

### HIGH PRIORITY - Adopt These

#### 1. OverlayFS Implementation

**Why**: Critical for scaling. Without it, each VM requires full rootfs copy.

**Implementation**:
```bash
# 1. Create overlay-init script in base image
# 2. Configure second drive for overlay
# 3. Update kernel args: init=/sbin/overlay-init overlay_root=/dev/vdb
```

**Expected benefit**: 10-100x storage efficiency at scale

#### 2. Template Snapshot System

**Why**: Eliminates cold start. Boot from snapshot instead of fresh boot.

**Implementation**:
- Build template from Dockerfile
- Boot VM and let it settle
- Create Firecracker snapshot
- On start: restore snapshot instead of boot

**Expected benefit**: Boot time from ~2s to <200ms

#### 3. In-VM Daemon (like envd)

**Why**: Provides consistent interface for SDK/API regardless of workload.

**Implementation**:
- Lightweight Go daemon running in VM
- HTTP/gRPC interface on well-known port
- Handles: file ops, process management, networking

**Expected benefit**: Enables rich SDK, consistent behavior

### MEDIUM PRIORITY - Consider These

#### 4. Dual Protocol API (REST + gRPC)

**Why**: REST for lifecycle, gRPC for real-time streaming.

**Current NanoFuse**: REST only
**Recommendation**: Add gRPC for console streaming, file transfers

#### 5. Lifecycle Events API

**Why**: Enables monitoring, debugging, audit trails.

**E2B Events**: created, updated, paused, resumed, killed
**Recommendation**: Implement similar webhook/polling API

#### 6. Custom Kernel Builds

**Why**: Optimized for Firecracker, security hardening.

**E2B approach**: fc-kernels repo with Terraform + shell build pipeline
**Recommendation**: Consider custom kernel for production

### LOWER PRIORITY - Future Consideration

#### 7. Code Interpreter Support

**Why**: Would expand use cases to AI code execution.

**Complexity**: Requires Jupyter integration, multi-language support
**Recommendation**: Only if Trigger.dev needs it

#### 8. Desktop Sandbox

**Why**: GUI workloads, browser automation.

**Complexity**: VNC, input handling, streaming
**Recommendation**: Out of scope for current use case

#### 9. Multi-Tenant Orchestration

**Why**: Enterprise-scale operations.

**E2B stack**: Nomad + Consul + Terraform
**Recommendation**: Consider when scaling beyond single-machine

### DO NOT ADOPT

1. **Managed SaaS Model**: NanoFuse should remain self-hosted focused
2. **Full Jupyter Protocol**: Over-engineered for Trigger.dev use case
3. **Build System 2.0**: Nice UX but adds complexity for our use case

---

## Key Implementation Details from E2B

### Firecracker Configuration for OverlayFS

```json
{
  "boot-source": {
    "kernel_image_path": "/path/to/vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw init=/sbin/overlay-init overlay_root=/dev/vdb panic=1 reboot=k"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "/path/to/rootfs.squashfs",
      "is_root_device": true,
      "is_read_only": true
    },
    {
      "drive_id": "overlay",
      "path_on_host": "/path/to/overlay-instance-1.ext4",
      "is_read_only": false
    }
  ]
}
```

### Overlay Init Script

```bash
#!/bin/bash
set -e

# Parse kernel args for overlay_root
for arg in $(cat /proc/cmdline); do
    case $arg in
        overlay_root=*)
            OVERLAY_ROOT="${arg#overlay_root=}"
            ;;
    esac
done

if [ "$OVERLAY_ROOT" = "ram" ]; then
    # Use tmpfs for non-persistent
    mount -t tmpfs tmpfs /overlay
else
    # Use block device for persistent
    mount "$OVERLAY_ROOT" /overlay
fi

mkdir -p /overlay/upper /overlay/work

# Mount overlay filesystem
mount -t overlay overlay \
    -o lowerdir=/,upperdir=/overlay/upper,workdir=/overlay/work \
    /mnt

# Move to new root
cd /mnt
pivot_root . mnt
exec chroot . /sbin/init
```

### Jailer Security Configuration

```bash
# Run Firecracker with jailer for production
jailer --id my-vm \
    --exec-file /usr/bin/firecracker \
    --uid 1000 \
    --gid 1000 \
    --chroot-base-dir /srv/jailer \
    --daemonize \
    -- \
    --api-sock /run/firecracker.socket
```

---

## Sources

### Primary Documentation
- [E2B Documentation](https://e2b.dev/docs)
- [E2B SDK Reference](https://e2b.dev/docs/sdk-reference)
- [GitHub - e2b-dev/E2B](https://github.com/e2b-dev/E2B)
- [GitHub - e2b-dev/infra](https://github.com/e2b-dev/infra)

### Technical Deep Dives
- [Firecracker vs QEMU - E2B Blog](https://e2b.dev/blog/firecracker-vs-qemu)
- [Scaling Firecracker: Using OverlayFS - E2B Blog](https://e2b.dev/blog/scaling-firecracker-using-overlayfs-to-save-disk-space)
- [E2B Breakdown - Dwarves Foundation](https://memo.d.foundation/breakdown/e2b)

### Go Package Documentation
- [envd package](https://pkg.go.dev/github.com/e2b-dev/infra/packages/envd)
- [process package](https://pkg.go.dev/github.com/e2b-dev/infra/packages/shared/pkg/grpc/envd/process)
- [integration tests](https://pkg.go.dev/github.com/e2b-dev/infra/tests/integration)

### Self-Hosting
- [DeepWiki Self-Hosting Guide](https://deepwiki.com/e2b-dev/infra/8-self-hosting-guide)
- [GitHub - e2b-dev/fc-kernels](https://github.com/e2b-dev/fc-kernels)

### Firecracker
- [Firecracker Snapshot Support](https://github.com/firecracker-microvm/firecracker/blob/main/docs/snapshotting/snapshot-support.md)
- [Firecracker MicroVM Security](https://securemachinery.com/2019/09/08/firecracker-microvm-security/)
- [Firecracker Rootfs and Kernel Setup](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernel-setup.md)

### Related Projects
- [GitHub - e2b-dev/code-interpreter](https://github.com/e2b-dev/code-interpreter)
- [GitHub - e2b-dev/desktop](https://github.com/e2b-dev/desktop)
- [GitHub - e2b-dev/surf](https://github.com/e2b-dev/surf)

### Market & Business
- [E2B Pricing](https://e2b.dev/pricing)
- [E2B Enterprise](https://e2b.dev/enterprise)
- [SkyPilot LLM Sandbox Comparison](https://blog.skypilot.co/skypilot-llm-sandbox/)
- [VentureBeat - E2B Funding](https://venturebeat.com/ai/how-e2b-became-essential-to-88-of-fortune-100-companies-and-raised-21-million/)

---

## Appendix: E2B Repository File Structure

### infra repository (`e2b-dev/infra`)

```
e2b-dev/infra/
├── iac/
│   └── provider-gcp/          # Terraform for GCP
├── packages/
│   ├── envd/                  # Environment daemon
│   │   ├── api/              # HTTP API (OpenAPI)
│   │   ├── internal/
│   │   │   ├── filesystem/   # File operations
│   │   │   ├── process/      # Process management
│   │   │   ├── host/         # Host interaction
│   │   │   └── permissions/  # Access control
│   │   └── spec/             # API specifications
│   ├── orchestrator/          # VM orchestration
│   │   └── internal/
│   │       └── cfg/          # Configuration
│   └── shared/
│       └── pkg/
│           └── grpc/
│               └── envd/     # gRPC definitions
│                   ├── filesystemconnect/
│                   └── processconnect/
├── scripts/                   # Deployment scripts
├── spec/                      # Specifications
└── tests/
    └── integration/           # Integration tests
```

### fc-kernels repository (`e2b-dev/fc-kernels`)

```
e2b-dev/fc-kernels/
├── configs/
│   ├── 5.10.xxx.config       # Kernel configs
│   └── 6.1.xxx.config
├── terraform/                 # Build infrastructure
├── build.sh                   # Build script
├── kernel_versions.txt        # Supported versions
└── Makefile
```

---

*Document generated by Claude Code research workflow*
*Last updated: 2025-11-26*
