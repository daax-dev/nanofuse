# NanoFuse Architecture Decisions

**Version:** 0.1.0 (Phase 0)
**Date:** 2025-10-30
**Status:** Definitive architectural specification for Phase 1 implementation

## Executive Summary

This document records the core architectural decisions for NanoFuse, a Firecracker-based microVM system. All decisions are grounded in **Gregor Hohpe's architectural principles** from *The Software Architect Elevator*, *Enterprise Integration Patterns*, *Cloud Strategy*, and *Platform Strategy*.

### Strategic Context (Penthouse View)

**Business Objectives:**
- Enable fast, isolated workloads for Trigger.dev self-hosting (dual web + worker environments)
- Sub-2-second cold start times via snapshot/resume
- Slicer-like operational simplicity (pull and run)
- Foundation for learning Firecracker and microVM orchestration

**Value Proposition:**
- Transition from manual VM management to automated orchestration
- Demonstrate technical capability through working implementation
- Enable future commercial applications (multi-tenant platforms, edge computing)

## Core Architectural Philosophy

### The Architect Elevator Model

NanoFuse architecture traverses multiple organizational levels:

1. **Penthouse (Strategic)**: Fast, isolated workloads enable new deployment patterns
2. **Mezzanine (Transformation)**: Simple API/CLI abstracts Firecracker complexity
3. **Engine Room (Technical)**: Firecracker processes, kernel management, networking

### Selling Options Framework

Every architectural decision is evaluated through **option theory**:
- **High volatility** (early project, learning phase) → Invest more in architecture
- **Defer decisions** until maximum information available
- **Keep options open** for future enhancements without over-engineering now

### Build Abstractions Not Illusions

- Expose real failure modes (VM crashes, lock conflicts, resource limits)
- Provide clear error messages with actionable remediation
- Don't hide complexity - abstract it intelligently

### Golden Path Principle

- Sensible defaults for common use cases (NAT networking, 2 vCPUs, 512MB)
- All defaults are configurable for advanced scenarios
- Progressive disclosure: simple commands for beginners, flags for power users

---

## ADR-001: CLI/API Separation via REST over Unix Socket

### Context

Need a daemon to manage long-running Firecracker processes and a user-facing CLI for operations.

### Decision

**Separate CLI (user tool) and API daemon (system service) communicating via REST over Unix socket.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Monolithic CLI** | Simple, no daemon | Cannot manage VMs when CLI exits | ❌ Rejected |
| **REST over Unix socket** | Process isolation, remote-ready protocol, standard semantics | Requires daemon | ✅ **Selected** |
| **gRPC** | Better performance, streaming | Complexity, less debuggable (binary protocol) | ❌ Rejected |
| **Direct Firecracker API** | No intermediary | CLI becomes complex, no state management | ❌ Rejected |

### Rationale (Option Theory)

**Option Purchased:** Ability to add remote management (TCP transport) without changing API contract

**Cost Deferred:** Authentication/authorization (not needed for MVP Unix socket use case)

**Volatility Management:** HTTP protocol is stable and well-understood, reduces risk

### Implementation Details

- **Primary transport:** Unix socket at `/var/run/nanofused.sock`
- **Protocol:** HTTP/1.1 with JSON payloads
- **Security:** Filesystem permissions (socket owned by `root:nanofuse`, mode `0660`)
- **Future:** TCP binding with Bearer token or mTLS (Phase 2+)

### Consequences

✅ **Positive:**
- CLI can exit, VMs keep running (daemon manages lifecycle)
- Multiple clients can connect (future web UI, automation tools)
- Clear separation of concerns (CLI=interface, API=orchestration)
- HTTP semantics well-understood (GET=read, POST=action, DELETE=remove)

⚠️ **Negative:**
- Requires daemon to be running (acceptable - systemd manages this)
- Cannot be accessed remotely without TCP binding (deferred to Phase 2)

### Related Patterns (Enterprise Integration Patterns)

- **Message Channel**: Unix socket is our communication channel
- **Request-Reply**: Synchronous HTTP request-response for most operations
- **Message Endpoint**: CLI and API are endpoints in the messaging system

---

## ADR-002: SQLite for State Persistence

### Context

API daemon must persist VM state across restarts (cannot rely on in-memory state).

### Decision

**Use SQLite as the embedded state database.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **In-memory only** | Simple | Lose state on restart | ❌ Rejected |
| **JSON files** | Human-readable | No transactions, poor query performance | ❌ Rejected |
| **SQLite** | Embedded, ACID, queryable, single file | Single-node only | ✅ **Selected** |
| **PostgreSQL** | Full SQL, multi-client | Requires separate process, overkill | ❌ Rejected |
| **etcd** | Distributed, consistent | Complexity, separate process | ❌ Rejected |

### Rationale (Option Theory)

**Option Purchased:** Ability to add advanced queries (filter VMs, retention policies) without changing persistence layer

**Cost Deferred:** Distributed state management (not needed for single-node MVP)

**Volatility Management:** SQLite is rock-solid and widely deployed, low risk

### Implementation Details

**Database Location:** `/var/lib/nanofuse/nanofuse.db`

**Schema Overview:**
- `vms` table: Core VM state and configuration
- `snapshots` table: Snapshot metadata (foreign key to VMs)
- `images` table: Cached image metadata
- `image_pull_jobs` table: Async pull job tracking

**Key Design Decisions:**
- Use `config_json` TEXT column for flexible VM config (avoid schema changes)
- Include `locked_by` and `locked_at` for pessimistic locking
- Foreign key cascade: Deleting VM cascades to snapshots
- Indexes on frequently queried columns (`state`, `name`)

### Consequences

✅ **Positive:**
- State survives daemon restarts
- ACID transactions prevent race conditions
- Queryable (list VMs by state, filter snapshots by date)
- Single file (easy backup: just copy `nanofuse.db`)
- Excellent Go library support (`mattn/go-sqlite3`, `modernc.org/sqlite`)

⚠️ **Negative:**
- Single-node only (no distributed deployments)
- Write contention possible (mitigated by short transactions)

⚠️ **Trade-offs Accepted:**
- Simplicity over scalability (appropriate for MVP and initial production)
- Can re-evaluate if distributed requirements emerge (would need etcd/Consul)

---

## ADR-003: Pessimistic Locking for VM Operations

### Context

Multiple CLI instances (or future web UI) may attempt concurrent operations on same VM.

### Decision

**Use pessimistic locking: Acquire exclusive lock on VM before state-changing operations.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **No locking** | Simple | Race conditions, corrupted state | ❌ Rejected |
| **Optimistic locking** | Better concurrency | Retry complexity, confusing errors | ❌ Rejected |
| **Pessimistic locking** | Simple, predictable | Operations wait if VM locked | ✅ **Selected** |

### Rationale (Build Abstractions Not Illusions)

We **expose the real constraint** (one operation at a time per VM) rather than hiding it:
- Clear error message: `"VM locked by another operation"`
- Error includes lock holder and timestamp
- Lock timeout prevents deadlock

### Implementation Details

**Locking Mechanism:**
1. Before state-changing operation, update `locked_by` (operation ID) and `locked_at` (timestamp)
2. If lock already held, return `409 Conflict` with error `VM_LOCKED`
3. On operation completion (success or failure), clear lock
4. Lock timeout: 5 minutes (prevent deadlock if operation crashes)

**Locked Operations:**
- Start, stop, kill, pause, resume
- Snapshot creation
- VM deletion

**Non-locked Operations:**
- Get status (read-only)
- List VMs (read-only)
- Get logs (read-only)

### Consequences

✅ **Positive:**
- Prevents race conditions (e.g., start and delete simultaneously)
- Simple to implement and reason about
- Clear error messages for conflicts

⚠️ **Negative:**
- Operations must wait if VM locked (acceptable - operations are fast)
- Lock timeout adds complexity (but necessary for resilience)

---

## ADR-004: NAT + Bridge Networking Modes

### Context

Trigger.dev requires inter-VM communication (web VM ↔ worker VM). Single VMs need internet access.

### Decision

**Support both NAT (default) and bridged networking modes.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **NAT only** | Simple, internet access | No inter-VM communication | ❌ Insufficient |
| **Bridge only** | Inter-VM communication | Complex setup, requires host config | ❌ Too complex as default |
| **Both NAT + Bridge** | Flexibility | Slightly more implementation | ✅ **Selected** |
| **User-mode networking** | Very simple | Limited, no inter-VM | ❌ Rejected |

### Rationale (Golden Path)

- **Default (NAT)**: Simple, works out of the box, sufficient for single-VM use cases
- **Advanced (Bridge)**: Enables multi-VM scenarios (Trigger.dev), requires explicit configuration

This follows the **Golden Path** principle: Default is simple, advanced users can opt into complexity.

### Implementation Details

**NAT Mode (Default):**
- Each VM gets a TAP device (`tap0`, `tap1`, ...)
- Host NATs traffic to internet
- VMs isolated from each other
- IP range: `172.16.0.0/24` (configurable)

**Bridged Mode:**
- VMs connect to Linux bridge (`br0` by default)
- VMs on same bridge can communicate via local network
- Requires bridge pre-configured on host (documented in deployment guide)

**API Configuration:**
```json
{
  "network": {
    "mode": "nat",  // or "bridged"
    "bridge_name": null,  // required if mode="bridged"
    "tap_device": null,  // auto-assigned if omitted
    "mac_address": null  // auto-generated if omitted
  }
}
```

### Consequences

✅ **Positive:**
- NAT mode works out-of-the-box (no host configuration needed)
- Bridge mode enables Trigger.dev's web+worker architecture
- Clear configuration (mode explicitly specified)

⚠️ **Negative:**
- Bridge mode requires host setup (create bridge, configure routing)
- Complexity in IP allocation for bridged mode

**Implementation Phasing:**
- **Phase 1:** NAT mode fully working
- **Phase 2:** Add bridge mode support

---

## ADR-005: Manual Snapshot Management

### Context

Need snapshot/resume for <2s cold start times (key requirement).

### Decision

**Snapshots created manually via API (`POST /vms/{id}/snapshots`). User manages retention.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Automatic snapshots** | Hands-off | When to snapshot? Complex policies | ❌ Rejected |
| **Manual snapshots** | User control, simple | User must explicitly create | ✅ **Selected** |
| **Snapshot-on-shutdown** | Consistent | Waste space if VM won't restart | ❌ Rejected |

### Rationale (Selling Options)

**Option Purchased:** Ability to add automatic snapshot policies later (cron-style, event-based)

**Cost Deferred:** Policy engine complexity (not needed for MVP)

**User Control:** Users know their workload best (when to snapshot)

### Implementation Details

**Snapshot Storage:**
- Location: `/var/lib/nanofuse/snapshots/{vm-id}/{snapshot-id}/`
- Files: `mem.snap` (memory state), `vm.snap` (device state)
- Metadata in SQLite (`snapshots` table)

**Snapshot Naming:**
- User can provide optional name (e.g., `"after-boot"`)
- Otherwise auto-generated timestamp-based ID (e.g., `snapshot-20251030-100530`)

**Retention:**
- No automatic deletion (user manages disk space)
- Can list snapshots and delete manually
- Future: Add retention policies (keep last N, delete older than X days)

**Resume Options:**
- `POST /vms/{id}/resume` with optional `snapshot_id` parameter
- If omitted, resumes from paused state (not snapshot)
- If specified, loads snapshot and resumes from that point

### Consequences

✅ **Positive:**
- Simple, predictable behavior
- User has full control
- Easy to implement (no policy engine needed)

⚠️ **Negative:**
- User must remember to create snapshots (not automatic)
- User must manage disk space (delete old snapshots)

**Mitigation:**
- Document best practices (snapshot after boot, before workload)
- CLI command to list snapshots with size info
- Warning if snapshot storage exceeds configured limit

---

## ADR-006: Image Format - Bundle Kernel with Rootfs

### Context

Need to distribute VM images (rootfs + kernel) for Firecracker.

### Decision

**Bundle kernel with rootfs in OCI image (Slicer's approach). Use proven Slicer 5.10.240 kernel initially.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Separate kernel management** | Flexibility | Complexity, version mismatches | ❌ Rejected |
| **Bundle kernel in image** | Atomic, version-locked | Larger images | ✅ **Selected** |
| **Host-provided kernel** | Small images | Configuration drift, incompatibility | ❌ Rejected |

### Rationale (Selling Options)

**Option Purchased:** Ability to support custom kernels per-image (different kernel configs for different workloads)

**Cost Deferred:** Building custom kernels (use proven Slicer kernel initially, defer compilation complexity)

**Consistency:** Kernel and rootfs versions always match (prevents compatibility issues)

### Implementation Details

**Image Structure (OCI format):**
```
/rootfs.ext4         # Block device image with filesystem
/boot/vmlinux        # Uncompressed kernel (Firecracker requires uncompressed)
/manifest.json       # Metadata (kernel cmdline, architecture, etc.)
```

**Build Process (Dockerfile):**
```dockerfile
FROM ubuntu:24.04
# Install systemd, openssh-server, networking
# Copy proven Slicer 5.10.240 kernel to /boot/vmlinux
# Configure systemd units (enable, not start)
# Generate rootfs.ext4 from filesystem
```

**Distribution:**
- Push to GHCR as OCI image
- Tag with version (`:latest`, `:v1.0.0`, `:sha-abc123`)
- Pull via CLI/API using standard OCI registry API

### Consequences

✅ **Positive:**
- Kernel and rootfs always match (atomic versioning)
- Easy distribution (standard OCI registry)
- Can use different kernels for different images (GPU-enabled, real-time, etc.)
- Proven kernel (Slicer 5.10.240) reduces boot issues

⚠️ **Negative:**
- Larger images (kernel adds ~30MB)
- Cannot easily swap kernels without rebuilding image

**Future Enhancement:**
- Phase 5+: Build custom kernel from source (learning exercise)
- Document kernel configuration for specific use cases

---

## ADR-007: Asynchronous Image Pull with Job Tracking

### Context

Image pulls can take minutes (500MB+ downloads). Cannot block API request.

### Decision

**Image pull is asynchronous. `POST /images/pull` returns immediately with job ID. Client polls `GET /images/jobs/{id}` for status.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Synchronous** | Simple | Blocks request, timeout issues | ❌ Rejected |
| **Async with job tracking** | Non-blocking, progress updates | Slightly more complex | ✅ **Selected** |
| **Webhooks** | Event-driven | Requires user to run webhook server | ❌ Too complex |

### Rationale (Enterprise Integration Patterns)

This is a classic **Request-Reply with Correlation ID** pattern:
1. Client: `POST /images/pull` → Server: `202 Accepted` with job ID
2. Client: `GET /images/jobs/{job-id}` → Server: `200 OK` with progress
3. Repeat step 2 until `state=completed` or `state=failed`

### Implementation Details

**Job States:**
- `pending`: Job queued, not started
- `in_progress`: Actively downloading
- `completed`: Pull successful
- `failed`: Pull failed

**Job Response:**
```json
{
  "id": "job-abc123",
  "image_ref": "ghcr.io/owner/base:latest",
  "state": "in_progress",
  "progress": {
    "current_bytes": 262144000,
    "total_bytes": 524288000,
    "percentage": 50
  },
  "error": null,
  "result_digest": null
}
```

**CLI Behavior:**
- Start pull: `nanofuse image pull <ref>`
- Show progress bar (auto-polling job status)
- Complete when state=completed
- Error if state=failed

### Consequences

✅ **Positive:**
- Non-blocking (API can handle other requests)
- Progress updates (better UX)
- Retryable (client can reconnect and check status)

⚠️ **Negative:**
- More complex than synchronous (but standard pattern)
- Client must poll (future: WebSocket for push updates)

---

## ADR-008: No API Versioning in 0.x Phase

### Context

Need strategy for API evolution and backwards compatibility.

### Decision

**No explicit versioning (`/v1/` prefix) during 0.x phase. Add at 1.0.0 release.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Version from start** | Clear stability signal | Premature, adds complexity | ❌ Rejected |
| **No versioning ever** | Simple | Cannot evolve API | ❌ Rejected |
| **Version at 1.0** | Signals maturity | Breaking change to add prefix | ✅ **Selected** |

### Rationale (Selling Options)

**Option Deferred:** Versioning cost (URL prefix, routing complexity) until API proven stable

**Volatility Management:** 0.x phase is high volatility (learning, experimentation). Breaking changes are expected and acceptable.

**Commitment at 1.0:** Adding `/v1/` prefix at 1.0.0 signals commitment to backwards compatibility

### Implementation Details

**0.x Releases (Current):**
- URLs: `/vms`, `/images`, `/snapshots` (no version prefix)
- Breaking changes allowed (document in release notes)
- Semantic versioning: `0.1.0`, `0.2.0`, ... (minor version bump = breaking change)

**1.0.0 Release (Future):**
- Add `/v1/` prefix: `/v1/vms`, `/v1/images`, ...
- Commit to backwards compatibility (no breaking changes in 1.x)
- Major version bump (2.0.0) for next breaking change

**CLI Version Checking:**
- CLI sends `User-Agent: nanofuse-cli/0.1.0`
- API responds with `X-API-Version: 0.1.0`
- CLI warns if versions mismatch (major or minor)

### Consequences

✅ **Positive:**
- Flexibility during 0.x phase (can iterate quickly)
- Clear signal at 1.0 (API is stable, production-ready)
- Simple URLs initially (no versioning overhead)

⚠️ **Negative:**
- Breaking changes in 0.x may disrupt early users (mitigated by clear release notes)
- Adding `/v1/` at 1.0 is itself a breaking change (acceptable as part of 1.0 milestone)

---

## ADR-009: Resource Limits and Quotas

### Context

Need to prevent resource exhaustion (memory, disk, VM count).

### Decision

**Enforce configurable resource limits at API level. Return `422 Unprocessable Entity` if limits exceeded.**

### Options Considered

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **No limits** | Simple | System can be exhausted | ❌ Rejected |
| **Hard-coded limits** | Simple | Not flexible | ❌ Rejected |
| **Configurable limits** | Flexible, safe | Slight complexity | ✅ **Selected** |

### Rationale (Build Abstractions Not Illusions)

**Expose Real Constraints:** We don't pretend resources are unlimited. Error message includes current usage and limit.

**Prevent Exhaustion Attacks:** Even in single-user scenarios, bugs can cause runaway VM creation

### Implementation Details

**Per-VM Limits:**
- `max_vcpus_per_vm`: Max vCPUs per VM (default: 8)
- `max_memory_per_vm_mib`: Max memory per VM (default: 8192 MB)

**Global Limits:**
- `max_vms`: Max total VMs (default: 50)
- `max_total_memory_mib`: Max memory across all VMs (default: 32768 MB = 32 GB)
- `max_snapshot_storage_gib`: Max snapshot storage (default: 100 GB)

**Validation:**
- Check limits during VM creation
- Check snapshot storage before creating snapshot
- Return clear error with current usage:
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

### Consequences

✅ **Positive:**
- Prevents system exhaustion (OOM killer, disk full)
- Clear error messages (user knows why request rejected)
- Configurable (can adjust for larger hosts)

⚠️ **Negative:**
- Adds validation logic (but necessary for robustness)
- Limits may need tuning based on workload

---

## ADR-010: Failure Recovery and Reconciliation

### Context

What happens when things go wrong? (API crashes, VM crashes, host reboots)

### Decision

**API reconciles state on startup. VMs are ephemeral by default (don't survive reboot).**

### Reconciliation Algorithm (API Startup)

1. Load all VMs from SQLite (`vms` table)
2. Check which Firecracker processes are actually running (by PID)
3. Reconcile discrepancies:
   - **Running process, in DB**: Keep as-is (VM survived restart)
   - **Running process, not in DB**: Orphaned process → log warning, optionally terminate
   - **Not running, DB shows "running"**: Zombie entry → update state to `failed`
   - **Not running, DB shows "stopped"**: Consistent → no action

### VM Crash Handling

- API monitors Firecracker processes (`waitpid`)
- On unexpected exit:
  - Update VM state to `failed`
  - Log exit code and last 100 lines of console output
  - Preserve logs at `/var/lib/nanofuse/vms/{id}/console.log`

### Snapshots Survive Everything

- Snapshots are files on disk (persist across API restart, host reboot)
- Can manually resume from snapshot after reboot:
  ```bash
  nanofuse vm create <image> <name>
  nanofuse vm resume <name> --from-snapshot <snapshot-id>
  ```

### Consequences

✅ **Positive:**
- API handles restarts gracefully (no manual cleanup needed)
- Clear failure modes (VM state=failed, logs available)
- Snapshots enable recovery

⚠️ **Negative:**
- VMs don't auto-restart on reboot (acceptable for MVP - can add systemd units later)
- User must manually investigate failed VMs

**Future Enhancement (Phase 3+):**
- Auto-restart policies (systemd unit per VM)
- Health checks (restart VM if unresponsive)

---

## ADR-011: Structured Logging with JSON

### Context

Need observability for debugging and operations.

### Decision

**API logs structured JSON to stderr. Systemd captures to journald.**

### Rationale

- **Structured logs**: Easy to parse, query, and analyze
- **JSON format**: Standard, tooling available (jq, log aggregators)
- **Stderr → journald**: Standard Unix pattern, queryable with `journalctl`

### Implementation Details

**Log Format:**
```json
{
  "timestamp": "2025-10-30T10:00:00Z",
  "level": "info",
  "message": "VM started successfully",
  "vm_id": "550e8400-e29b-41d4-a716-446655440000",
  "vm_name": "web-vm",
  "duration_ms": 1834
}
```

**Log Levels:**
- `debug`: Verbose (HTTP requests, internal state)
- `info`: Normal operations (VM started, stopped)
- `warn`: Non-fatal issues (lock acquisition retry, high memory usage)
- `error`: Failures (VM crash, API errors)

**Query Examples:**
```bash
# All logs for nanofused
journalctl -u nanofused

# Errors only
journalctl -u nanofused -p err

# Logs for specific VM
journalctl -u nanofused | jq 'select(.vm_id=="550e8400-e29b-41d4-a716-446655440000")'

# Recent failures
journalctl -u nanofused --since "1 hour ago" | jq 'select(.level=="error")'
```

### Consequences

✅ **Positive:**
- Easy to query and analyze
- Standard tools (journalctl, jq)
- Can forward to log aggregators (Loki, Elasticsearch)

---

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         User Space                          │
│                                                              │
│  ┌──────────────┐         ┌─────────────────────────────┐  │
│  │ nanofuse CLI │────────▶│   nanofused API Daemon      │  │
│  │   (Go)       │  HTTP   │        (Go)                 │  │
│  │              │  Unix   │                              │  │
│  │ - parse args │  Socket │ - REST API server           │  │
│  │ - format out │         │ - VM lifecycle management   │  │
│  │ - call API   │         │ - State persistence (SQLite)│  │
│  └──────────────┘         │ - Firecracker orchestration │  │
│                           │ - Network management (TAP)  │  │
│                           │ - Snapshot management       │  │
│                           └────────────┬────────────────┘  │
│                                        │                    │
│                                        │ spawn/manage       │
│                           ┌────────────▼────────────────┐  │
│                           │  Firecracker Processes      │  │
│                           │  (multiple instances)       │  │
│                           │                              │  │
│                           │  ┌──────────┐  ┌──────────┐ │  │
│                           │  │ VM 1     │  │ VM 2     │ │  │
│                           │  │ (web)    │  │ (worker) │ │  │
│                           │  └──────────┘  └──────────┘ │  │
│                           └─────────────────────────────┘  │
│                                        │                    │
└────────────────────────────────────────┼────────────────────┘
                                         │ /dev/kvm
                                         │ TAP devices
┌────────────────────────────────────────▼────────────────────┐
│                         Kernel Space                         │
│                                                              │
│  ┌──────────────┐  ┌─────────────┐  ┌───────────────────┐  │
│  │     KVM      │  │  TAP/Bridge │  │  cgroups v2       │  │
│  │ (virtio)     │  │  Networking │  │  (resource limits)│  │
│  └──────────────┘  └─────────────┘  └───────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘

Storage Layout:
/var/run/nanofused.sock                    # Unix socket
/var/lib/nanofuse/
  ├── nanofuse.db                          # SQLite database
  ├── images/                              # Pulled images
  │   └── sha256:abc123.../
  │       ├── rootfs.ext4
  │       ├── vmlinux
  │       └── manifest.json
  ├── vms/                                 # VM runtime data
  │   └── 550e8400-e29b-41d4-a716-446655440000/
  │       ├── console.log
  │       └── firecracker.sock
  └── snapshots/                           # Snapshots
      └── 550e8400-e29b-41d4-a716-446655440000/
          └── snapshot-20251030-100530/
              ├── mem.snap
              └── vm.snap
```

---

## Trade-Off Summary

All architecture involves trade-offs. Here's what we accepted:

| Decision | Benefit | Cost | Justification |
|----------|---------|------|---------------|
| Unix socket (not TCP) | Low latency, simple security | No remote access | Remote access deferred to Phase 2+ |
| SQLite (not distributed) | Simple, ACID, queryable | Single-node only | Sufficient for MVP, distributed later if needed |
| Manual snapshots | User control, simple | User must remember | Can add auto-policies later |
| Bundle kernel in image | Atomic versioning | Larger images | Worth it for consistency |
| No versioning in 0.x | Flexibility | Breaking changes | Acceptable for learning phase |
| NAT default (not bridge) | Works out-of-box | No inter-VM by default | Bridge mode available for advanced users |

---

## Alignment Framework Assessment

Using Hohpe's 4-dimensional alignment model:

### Business Alignment ✅
- **Objective**: Enable Trigger.dev dual-environment workloads
- **Value**: Fast, isolated VMs with snapshot/resume
- **Delivered**: API/CLI abstracts Firecracker complexity

### Organization Alignment ✅
- **Team**: Single developer initially
- **Skills**: Go, Linux, networking (all present)
- **Documentation**: Comprehensive specs enable future contributors

### Technology Alignment ✅
- **Stack**: Go (static binaries), Firecracker (proven), SQLite (reliable)
- **Integration**: Standard protocols (HTTP, OCI registry)
- **Maintainability**: Simple architecture, clear abstractions

### Financial Alignment ✅
- **Cost**: Open source, free GitHub hosting
- **Investment**: Time (reasonable for learning + production value)
- **Return**: Working system for Trigger.dev + transferable knowledge

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Firecracker process crashes | Medium | High | Monitoring, auto-restart (Phase 3+) |
| API daemon crashes | Low | Medium | Systemd restarts, state in SQLite |
| Disk full (snapshots) | Medium | Medium | Resource limits, warnings, retention policies |
| Network misconfiguration | Medium | Medium | Validation, clear error messages, docs |
| Image pull auth failures | Medium | Low | Clear errors, retry logic |
| VM-to-VM routing issues | Low | Medium | Bridge mode documentation, examples |
| Kernel incompatibility | Low | High | Use proven Slicer kernel initially |

---

## Future Architecture Evolution

### Phase 2 (After MVP)
- TCP API binding with authentication
- Bridge networking mode
- Advanced logging and metrics (Prometheus)

### Phase 3 (Production Hardening)
- Auto-restart policies (systemd units per VM)
- Snapshot retention policies
- Health checks and monitoring

### Phase 4 (Multi-tenancy)
- Namespace isolation
- Per-tenant quotas
- Image signing and verification

### Phase 5+ (Advanced Features)
- Custom kernel compilation (learning exercise)
- GPU support via Cloud Hypervisor
- Distributed orchestration (if needed)

**Key Principle:** Each phase builds on previous, no backtracking. Decisions made now keep options open for future enhancements.

---

## Validation Against Hohpe's Principles

### ✅ Architect Elevator
- **Penthouse**: Business value clearly articulated
- **Transformation**: API/CLI simplifies operations
- **Engine Room**: Technical implementation sound

### ✅ Selling Options
- Every decision evaluated through option theory
- Defer costs until needed (auth, versioning, policies)
- Buy options for future (extensibility, networking modes)

### ✅ Build Abstractions Not Illusions
- Expose real failures (locks, limits, crashes)
- Clear error messages with remediation
- Don't hide complexity of VM management

### ✅ Golden Path
- Sensible defaults (NAT, 2 vCPUs, 512MB)
- Progressive disclosure (simple → advanced)
- Common case optimized, edge cases supported

---

## Related Documents

- **API Contract**: `API_CONTRACT.md` - REST API specification
- **CLI Specification**: `CLI_SPEC.md` - Command-line interface
- **Execution Plan**: `EXECUTION_PLAN.md` - Implementation phases
- **Project Overview**: `../CLAUDE.md` - Project context

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0 | 2025-10-30 | Initial architecture definition (Phase 0) |

---

*This architecture follows principles from Gregor Hohpe's seminal works: The Software Architect Elevator, Enterprise Integration Patterns, Cloud Strategy, and Platform Strategy.*

**Architect's Signature**: Definitive specification for Phase 1 implementation.
