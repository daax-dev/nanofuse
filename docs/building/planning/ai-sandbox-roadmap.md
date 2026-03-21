# NanoFuse AI Sandbox Roadmap

**Date**: 2025-11-26
**Status**: Draft
**Based on**: E2B architecture analysis (`e2b-learnings.md`)

---

## Vision

NanoFuse is an open-source, self-hosted alternative to E2B for running untrusted code in secure Firecracker microVM sandboxes. Primary focus on AI code execution with sub-200ms boot times.

## Use Cases (Priority Order)

1. **AI Code Execution Sandbox** - Run LLM-generated code securely
2. **General Isolated Workloads** - Ephemeral compute for multi-tenant workloads
3. **Dev Environment VMs** - Fast-spinning development sandboxes

---

## Phase 1: Core Infrastructure (Current - 60% Complete)

**Goal**: Working VM lifecycle management

### Completed
- [x] CLI implementation (`nanofuse`)
- [x] API daemon implementation (`nanofused`)
- [x] Basic Firecracker integration
- [x] TAP networking with IPAM
- [x] Base image (Ubuntu 24.04 + systemd)
- [x] CI/CD pipeline
- [x] GHCR image distribution

### Remaining
- [ ] Fix any blocking issues in VM lifecycle
- [ ] Verify end-to-end workflow
- [ ] Stabilize for Phase 2

### Success Criteria
- Can start, stop, and manage VMs via CLI and API
- Networking functional (host ↔ VM connectivity)
- Clean resource cleanup on VM termination

---

## Phase 2: Fast Boot (Next Priority)

**Goal**: Achieve <200ms sandbox boot times

### 2.1 OverlayFS Implementation

**Why**: Critical for scaling. Without it, each VM requires full rootfs copy.

**Tasks**:
- [ ] Create overlay-init script for base image
- [ ] Configure dual-drive Firecracker setup (RO rootfs + RW overlay)
- [ ] Implement sparse ext4 overlay creation per instance
- [ ] Update kernel args: `init=/sbin/overlay-init overlay_root=/dev/vdb`
- [ ] Add `overlay_root=ram` option for non-persistent instances
- [ ] Test with 10+ concurrent instances

**Expected Outcome**: 10-100x storage efficiency, faster instance creation

### 2.2 Snapshot/Resume System

**Why**: Eliminates cold start. Boot from snapshot instead of fresh boot.

**Tasks**:
- [ ] Implement Firecracker snapshot creation API
- [ ] Implement snapshot restore API
- [ ] Template metadata storage (what snapshot goes with what template)
- [ ] Snapshot storage management (local filesystem initially)
- [ ] Add `nanofuse template build` command
- [ ] Add `nanofuse template list` command

**Expected Outcome**: Boot time from ~2s to <200ms

### 2.3 Template System

**Why**: Allow custom environments built from Dockerfiles.

**Tasks**:
- [ ] Dockerfile → rootfs extraction pipeline
- [ ] Template build workflow:
  1. Build Docker image
  2. Extract filesystem
  3. Boot VM and let it settle
  4. Create snapshot
  5. Store as template
- [ ] Template versioning
- [ ] Template registry (GHCR-based)

**Expected Outcome**: Users can create custom sandbox templates

### Success Criteria (Phase 2)
- Sandbox boot time <200ms (from snapshot)
- 100+ sandboxes can share same base rootfs via OverlayFS
- Custom templates can be built from Dockerfiles

---

## Phase 3: SDK & In-VM Daemon

**Goal**: Enable programmatic sandbox interaction for AI agents

### 3.1 In-VM Daemon (nanofuse-envd)

**Why**: Provides consistent interface for SDK regardless of workload.

**Tasks**:
- [ ] Design envd API (HTTP + gRPC)
- [ ] Implement filesystem service (read, write, list, delete)
- [ ] Implement process service (exec, kill, stream stdout/stderr)
- [ ] Implement terminal service (PTY allocation)
- [ ] Health check endpoint
- [ ] Include envd in base template
- [ ] Auto-start envd via systemd

**API Design** (based on E2B):
```
POST   /filesystem/read     - Read file contents
POST   /filesystem/write    - Write file contents
POST   /filesystem/list     - List directory
DELETE /filesystem/delete   - Delete file/directory

POST   /process/start       - Start process
POST   /process/kill        - Kill process
GET    /process/stream      - Stream stdout/stderr (WebSocket/gRPC stream)

POST   /terminal/create     - Create PTY session
POST   /terminal/resize     - Resize terminal
GET    /terminal/stream     - Terminal I/O (WebSocket/gRPC stream)
```

### 3.2 Python SDK

**Why**: Primary integration path for AI agents (LangChain, etc.)

**Tasks**:
- [ ] Design SDK API surface
- [ ] Implement `Sandbox.create()` / `Sandbox.close()`
- [ ] Implement `sandbox.run_code(code, language)`
- [ ] Implement `sandbox.filesystem.read/write/list`
- [ ] Implement `sandbox.process.start/kill`
- [ ] Context manager support (`with Sandbox() as sbx:`)
- [ ] Async support
- [ ] Publish to PyPI

**Example Usage**:
```python
from nanofuse import Sandbox

with Sandbox.create() as sandbox:
    result = sandbox.run_code("print('Hello from sandbox!')")
    print(result.stdout)  # "Hello from sandbox!"

    sandbox.filesystem.write("/tmp/data.txt", "some data")
    content = sandbox.filesystem.read("/tmp/data.txt")
```

### 3.3 JavaScript/TypeScript SDK

**Tasks**:
- [ ] Mirror Python SDK functionality
- [ ] Publish to npm

### Success Criteria (Phase 3)
- Python SDK can create sandbox, execute code, read/write files
- AI agent (e.g., LangChain) can use SDK to run generated code
- envd provides reliable in-VM communication

---

## Phase 4: Code Interpreter

**Goal**: Jupyter-like code execution with context persistence

### 4.1 Jupyter Integration

**Tasks**:
- [ ] Include Jupyter server in code-interpreter template
- [ ] Implement Jupyter kernel protocol (partial, like E2B)
- [ ] Context persistence between executions
- [ ] Support multiple languages (Python, JavaScript, Bash)
- [ ] Rich output support (text, images, plots, HTML)

### 4.2 Multi-Language Support

**Tasks**:
- [ ] Python kernel (ipykernel)
- [ ] JavaScript kernel (Node.js)
- [ ] Bash kernel
- [ ] R kernel (optional)

### 4.3 Code Interpreter SDK

**Tasks**:
- [ ] `CodeInterpreter` class extending base `Sandbox`
- [ ] `interpreter.run_code()` with context preservation
- [ ] Result objects with multiple output types
- [ ] Chart/visualization extraction

**Example Usage**:
```python
from nanofuse import CodeInterpreter

with CodeInterpreter.create() as interp:
    interp.run_code("import pandas as pd")
    interp.run_code("df = pd.DataFrame({'a': [1,2,3]})")
    result = interp.run_code("df.sum()")
    print(result.text)  # Shows sum result
```

### Success Criteria (Phase 4)
- Code interpreter maintains context across executions
- Multiple languages supported
- Rich outputs (charts, images) can be extracted

---

## Phase 5: Production Hardening

**Goal**: Production-ready for real workloads

### 5.1 Security Hardening

**Tasks**:
- [ ] Implement Firecracker jailer integration
- [ ] Seccomp filter configuration
- [ ] Cgroups resource limits
- [ ] Network isolation policies
- [ ] Audit logging

### 5.2 Observability

**Tasks**:
- [ ] Metrics export (Prometheus)
- [ ] Structured logging
- [ ] Distributed tracing
- [ ] Lifecycle events API (created, paused, resumed, killed)

### 5.3 Scaling

**Tasks**:
- [ ] Multi-node orchestration (optional Nomad integration)
- [ ] Template caching/distribution
- [ ] Connection pooling
- [ ] Rate limiting

### 5.4 Pause/Resume

**Tasks**:
- [ ] Sandbox pause API
- [ ] Sandbox resume API
- [ ] State preservation (filesystem + processes)
- [ ] Configurable pause duration (up to 24h like E2B)

---

## Technical Decisions

### Language: Go

All server-side components (nanofused, nanofuse-envd) in Go for:
- Static binaries
- Excellent concurrency
- Firecracker SDK compatibility
- E2B infra is also Go

### API Design: REST + gRPC

- **REST**: Sandbox lifecycle (create, kill, pause, resume)
- **gRPC**: Real-time operations (file streaming, process I/O, terminal)

### Storage: Local First

Start with local filesystem for:
- Snapshots
- Templates
- Overlay images

Later: Object storage (S3/GCS) for multi-node deployments

### Networking: TAP + NAT

Current approach works. Consider CNI plugins for advanced use cases later.

---

## Non-Goals (For Now)

1. **Desktop/GUI sandbox** - Different use case, add later if needed
2. **Multi-cloud orchestration** - Focus on single-node first
3. **Kubernetes integration** - Keep it simple
4. **Windows VMs** - Linux only

---

## Comparison: NanoFuse vs E2B

| Feature | E2B | NanoFuse (Target) |
|---------|-----|-------------------|
| Boot time | ~150ms | <200ms |
| Self-hosted | BYOC option | Primary mode |
| Open source | Partial (infra repo) | Fully open |
| Orchestration | Nomad + Consul | Single-node first |
| SDK | Python + JS | Python + JS |
| Code interpreter | Full Jupyter | Jupyter-compatible |
| Desktop sandbox | Yes | No (not planned) |
| Pricing | Per-second | Self-hosted (free) |

---

## Success Metrics

### Phase 2 (Fast Boot)
- [ ] Boot time <200ms (measured)
- [ ] 100 concurrent sandboxes on single node
- [ ] Storage overhead <50MB per sandbox (with OverlayFS)

### Phase 3 (SDK)
- [ ] Python SDK published to PyPI
- [ ] LangChain integration demo working
- [ ] envd uptime >99.9% during sandbox lifetime

### Phase 4 (Code Interpreter)
- [ ] Context preserved across 100 consecutive executions
- [ ] 3+ languages supported
- [ ] Rich output extraction working

---

## Timeline

No time estimates (per project guidelines). Phases are sequential dependencies:

1. **Phase 1** → Must complete before Phase 2
2. **Phase 2** → Enables Phase 3 (fast boot needed for good SDK UX)
3. **Phase 3** → Enables Phase 4 (SDK needed for code interpreter)
4. **Phase 4** → Can proceed to Phase 5 in parallel

---

## References

- [E2B Learnings](./e2b-learnings.md) - Detailed E2B architecture analysis
- [E2B Documentation](https://e2b.dev/docs)
- [E2B Infra Repository](https://github.com/e2b-dev/infra)
- [Firecracker Documentation](https://github.com/firecracker-microvm/firecracker/tree/main/docs)
