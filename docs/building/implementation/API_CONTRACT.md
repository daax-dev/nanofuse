# NanoFuse API Contract Specification

**Version:** 0.1.0 (Pre-release)
**Date:** 2025-10-30
**Status:** Phase 0 - Architecture Definition

## Overview

The NanoFuse API is a RESTful HTTP service that manages the complete lifecycle of Firecracker-based microVMs. It runs as a systemd daemon (`nanofused`) and exposes operations for VM creation, lifecycle management, snapshot/resume, and image management.

## Architectural Principles

This API design follows Gregor Hohpe's architectural principles from *The Software Architect Elevator*:

1. **Build Abstractions Not Illusions**: The API exposes real failure modes and constraints (VM locking, resource limits) rather than hiding them
2. **Golden Path**: Sensible defaults enable simple use cases while maintaining full configurability for advanced scenarios
3. **Selling Options**: API design defers decisions (e.g., authentication, advanced networking) while keeping options open for future enhancements

## Transport and Protocol

### Phase 1 (MVP)
- **Protocol**: HTTP/1.1
- **Transport**: Unix domain socket at `/var/run/nanofused.sock`
- **Authentication**: Filesystem permissions (socket owned by `root:nanofuse`, mode `0660`)
- **Content-Type**: `application/json`
- **Encoding**: UTF-8

### Future (Phase 2+)
- **Transport**: TCP binding (e.g., `http://localhost:8080`)
- **Authentication**: Bearer token or mTLS for remote access

### Base URL
- Unix socket: `http://unix:/var/run/nanofused.sock`
- TCP (future): `http://localhost:8080`

## API Versioning

- **0.x releases**: No explicit version prefix in URLs (e.g., `/vms`, not `/v1/vms`)
- **Breaking changes allowed** during 0.x phase (experimental/learning phase)
- **1.0.0 release**: Commit to API stability, introduce `/v1/` prefix, maintain backwards compatibility

## Core Resources

### 1. VM (Virtual Machine)

Represents a Firecracker microVM instance.

**Resource Model:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "web-vm",
  "state": "running",
  "image": "ghcr.io/owner/nanofuse-base:latest",
  "image_digest": "sha256:abc123...",
  "config": {
    "vcpus": 2,
    "memory_mib": 512,
    "kernel_args": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k",
    "network": {
      "mode": "nat",
      "tap_device": "tap0",
      "mac_address": "AA:FC:00:00:00:01",
      "bridge_name": null
    },
    "disks": [
      {
        "drive_id": "rootfs",
        "path_on_host": "/var/lib/nanofuse/images/sha256:abc123.../rootfs.ext4",
        "is_read_only": false,
        "is_root_device": true
      }
    ]
  },
  "runtime": {
    "pid": 12345,
    "socket_path": "/var/lib/nanofuse/vms/550e8400-e29b-41d4-a716-446655440000/firecracker.sock",
    "console_path": "/var/lib/nanofuse/vms/550e8400-e29b-41d4-a716-446655440000/console.log",
    "network_info": {
      "tap_device": "tap0",
      "host_ip": "172.16.0.1",
      "guest_ip": "172.16.0.2",
      "gateway": "172.16.0.1"
    }
  },
  "created_at": "2025-10-30T10:00:00Z",
  "updated_at": "2025-10-30T10:05:30Z",
  "locked_by": null,
  "locked_at": null
}
```

**State Machine:**
```
[created] → [starting] → [running] → [stopping] → [stopped]
                ↓            ↓            ↓
              [failed]    [pausing]   [failed]
                             ↓
                         [paused] → [resuming] → [running]
                             ↓
                         [failed]
```

Valid state transitions:
- `created` → `starting`, `failed`
- `starting` → `running`, `failed`
- `running` → `stopping`, `pausing`, `failed`
- `stopping` → `stopped`, `failed`
- `pausing` → `paused`, `failed`
- `paused` → `resuming`, `stopped`, `failed`
- `resuming` → `running`, `failed`
- Any state → `failed` (on error)

### 2. Image

Represents a pulled OCI image containing rootfs and kernel.

**Resource Model:**
```json
{
  "digest": "sha256:abc123...",
  "tags": ["ghcr.io/owner/nanofuse-base:latest"],
  "architecture": "x86_64",
  "size_bytes": 524288000,
  "kernel_version": "5.10.240",
  "rootfs_path": "/var/lib/nanofuse/images/sha256:abc123.../rootfs.ext4",
  "kernel_path": "/var/lib/nanofuse/images/sha256:abc123.../vmlinux",
  "pulled_at": "2025-10-30T09:00:00Z"
}
```

### 3. Snapshot

Represents saved VM state for fast resume.

**Resource Model:**
```json
{
  "id": "snapshot-20251030-100530",
  "vm_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "after-boot",
  "memory_file_path": "/var/lib/nanofuse/snapshots/550e8400.../snapshot-20251030-100530/mem.snap",
  "snapshot_file_path": "/var/lib/nanofuse/snapshots/550e8400.../snapshot-20251030-100530/vm.snap",
  "size_bytes": 536870912,
  "created_at": "2025-10-30T10:05:30Z"
}
```

### 4. ImagePullJob

Represents an asynchronous image pull operation.

**Resource Model:**
```json
{
  "id": "job-550e8400-e29b-41d4-a716-446655440001",
  "image_ref": "ghcr.io/owner/nanofuse-base:latest",
  "state": "in_progress",
  "progress": {
    "current_bytes": 262144000,
    "total_bytes": 524288000,
    "percentage": 50
  },
  "error": null,
  "created_at": "2025-10-30T09:00:00Z",
  "completed_at": null,
  "result_digest": null
}
```

States: `pending`, `in_progress`, `completed`, `failed`

## API Endpoints

### Health and Status

#### `GET /health`

Health check endpoint.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "uptime_seconds": 3600
}
```

---

### Virtual Machine Operations

#### `POST /vms`

Create a new VM. **Idempotent by name** - if a VM with the given name exists, returns the existing VM.

**Request Body:**
```json
{
  "name": "web-vm",  // optional, auto-generated UUID if omitted
  "image": "ghcr.io/owner/nanofuse-base:latest",  // or sha256 digest for immutability
  "config": {
    "vcpus": 2,  // optional, default: 2
    "memory_mib": 512,  // optional, default: 512
    "kernel_args": "console=ttyS0 root=/dev/vda1 rw",  // optional, uses image default
    "network": {
      "mode": "nat",  // optional: "nat" | "bridged" | "none", default: "nat"
      "bridge_name": null,  // required if mode="bridged"
      "mac_address": null  // optional, auto-generated if omitted
    }
  }
}
```

**Response (201 Created or 200 OK if exists):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "web-vm",
  "state": "created",
  ...
}
```

**Errors:**
- `400 Bad Request` - Invalid input (malformed JSON, invalid config values)
- `404 Not Found` - Image not found locally (must pull first)
- `422 Unprocessable Entity` - Validation errors (e.g., exceeds resource limits)
- `500 Internal Server Error` - Unexpected failure

---

#### `GET /vms`

List all VMs.

**Query Parameters:**
- `state` (optional): Filter by state (e.g., `?state=running`)
- `name` (optional): Filter by name prefix (e.g., `?name=web`)

**Response (200 OK):**
```json
{
  "vms": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "web-vm",
      "state": "running",
      "image": "ghcr.io/owner/nanofuse-base:latest",
      "created_at": "2025-10-30T10:00:00Z",
      "uptime_seconds": 330
    },
    ...
  ],
  "total": 2
}
```

---

#### `GET /vms/{id}`

Get detailed VM information.

**Path Parameters:**
- `id`: VM ID (UUID or name)

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "web-vm",
  "state": "running",
  ...
}
```

**Errors:**
- `404 Not Found` - VM does not exist

---

#### `DELETE /vms/{id}`

Destroy a VM. Stops the VM if running and removes all associated state (but preserves snapshots).

**Response (204 No Content)**

**Errors:**
- `404 Not Found` - VM does not exist
- `409 Conflict` - VM locked by another operation

---

#### `POST /vms/{id}/start`

Start a created or stopped VM. **Idempotent** - returns success if already running.

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "starting",
  ...
}
```

**Errors:**
- `404 Not Found` - VM does not exist
- `409 Conflict` - Invalid state transition or VM locked
- `500 Internal Server Error` - Failed to start Firecracker

---

#### `POST /vms/{id}/stop`

Stop a running VM gracefully (sends ACPI shutdown signal). **Idempotent** - returns success if already stopped.

**Request Body (optional):**
```json
{
  "timeout_seconds": 30  // optional, default: 30, max: 300
}
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "stopping",
  ...
}
```

**Errors:**
- `404 Not Found` - VM does not exist
- `409 Conflict` - Invalid state transition or VM locked

---

#### `POST /vms/{id}/kill`

Force kill a VM (SIGKILL to Firecracker process). Use when graceful stop fails.

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "stopped",
  ...
}
```

**Errors:**
- `404 Not Found` - VM does not exist

---

#### `POST /vms/{id}/pause`

Pause a running VM (freezes execution, saves memory state).

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "pausing",
  ...
}
```

**Errors:**
- `404 Not Found` - VM does not exist
- `409 Conflict` - Invalid state transition (VM not running) or VM locked

---

#### `POST /vms/{id}/resume`

Resume a paused VM or resume from a snapshot.

**Request Body (optional):**
```json
{
  "snapshot_id": "snapshot-20251030-100530"  // optional, if omitted resumes from paused state
}
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "resuming",
  ...
}
```

**Errors:**
- `404 Not Found` - VM or snapshot does not exist
- `409 Conflict` - Invalid state transition or VM locked
- `422 Unprocessable Entity` - Snapshot incompatible with VM config

---

#### `GET /vms/{id}/logs`

Get VM console output.

**Query Parameters:**
- `follow` (optional): Stream logs (boolean, default: false)
- `tail` (optional): Number of lines from end (integer, default: all)

**Response (200 OK):**
```
[    0.000000] Linux version 5.10.240 ...
[    0.100000] Booting kernel...
...
```

**Content-Type:** `text/plain` (or `application/x-ndjson` if following)

**Errors:**
- `404 Not Found` - VM does not exist

---

### Snapshot Operations

#### `POST /vms/{id}/snapshots`

Create a snapshot of a running or paused VM.

**Request Body:**
```json
{
  "name": "after-boot"  // optional, auto-generated timestamp-based ID if omitted
}
```

**Response (201 Created):**
```json
{
  "id": "snapshot-20251030-100530",
  "vm_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "after-boot",
  ...
}
```

**Errors:**
- `404 Not Found` - VM does not exist
- `409 Conflict` - Invalid state (VM not running/paused) or VM locked
- `507 Insufficient Storage` - Not enough disk space for snapshot

---

#### `GET /vms/{id}/snapshots`

List all snapshots for a VM.

**Response (200 OK):**
```json
{
  "snapshots": [
    {
      "id": "snapshot-20251030-100530",
      "vm_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "after-boot",
      "size_bytes": 536870912,
      "created_at": "2025-10-30T10:05:30Z"
    },
    ...
  ],
  "total": 3
}
```

**Errors:**
- `404 Not Found` - VM does not exist

---

#### `GET /snapshots/{snapshot_id}`

Get snapshot metadata.

**Response (200 OK):**
```json
{
  "id": "snapshot-20251030-100530",
  "vm_id": "550e8400-e29b-41d4-a716-446655440000",
  ...
}
```

**Errors:**
- `404 Not Found` - Snapshot does not exist

---

#### `DELETE /snapshots/{snapshot_id}`

Delete a snapshot.

**Response (204 No Content)**

**Errors:**
- `404 Not Found` - Snapshot does not exist
- `409 Conflict` - Snapshot in use

---

### Image Operations

#### `GET /images`

List locally cached images.

**Response (200 OK):**
```json
{
  "images": [
    {
      "digest": "sha256:abc123...",
      "tags": ["ghcr.io/owner/nanofuse-base:latest"],
      "architecture": "x86_64",
      "size_bytes": 524288000,
      "pulled_at": "2025-10-30T09:00:00Z"
    },
    ...
  ],
  "total": 5
}
```

---

#### `GET /images/{digest}`

Get image metadata by digest.

**Path Parameters:**
- `digest`: Image digest (e.g., `sha256:abc123...`)

**Response (200 OK):**
```json
{
  "digest": "sha256:abc123...",
  "tags": ["ghcr.io/owner/nanofuse-base:latest"],
  ...
}
```

**Errors:**
- `404 Not Found` - Image not found locally

---

#### `DELETE /images/{digest}`

Remove a cached image. Fails if any VMs reference this image.

**Response (204 No Content)**

**Errors:**
- `404 Not Found` - Image does not exist
- `409 Conflict` - Image in use by VMs

---

#### `POST /images/pull`

Pull an image from a registry. **Asynchronous operation** - returns job ID immediately.

**Request Body:**
```json
{
  "image_ref": "ghcr.io/owner/nanofuse-base:latest"
}
```

**Response (202 Accepted):**
```json
{
  "job_id": "job-550e8400-e29b-41d4-a716-446655440001",
  "image_ref": "ghcr.io/owner/nanofuse-base:latest",
  "state": "pending",
  "status_url": "/images/jobs/job-550e8400-e29b-41d4-a716-446655440001"
}
```

**Errors:**
- `400 Bad Request` - Invalid image reference format
- `401 Unauthorized` - Registry authentication failed (missing or invalid credentials)
- `404 Not Found` - Image not found in registry

---

#### `GET /images/jobs/{job_id}`

Get image pull job status.

**Response (200 OK):**
```json
{
  "id": "job-550e8400-e29b-41d4-a716-446655440001",
  "image_ref": "ghcr.io/owner/nanofuse-base:latest",
  "state": "completed",
  "progress": {
    "current_bytes": 524288000,
    "total_bytes": 524288000,
    "percentage": 100
  },
  "error": null,
  "result_digest": "sha256:abc123...",
  "completed_at": "2025-10-30T09:05:00Z"
}
```

**Errors:**
- `404 Not Found` - Job does not exist

---

## Error Handling

### Error Response Format

All errors return JSON with the following structure:

```json
{
  "error": {
    "code": "VM_NOT_FOUND",
    "message": "Virtual machine with ID '550e8400-e29b-41d4-a716-446655440000' does not exist",
    "details": {
      "vm_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

### Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `INVALID_REQUEST` | Malformed request (invalid JSON, missing required fields) |
| 400 | `INVALID_CONFIG` | Invalid configuration values |
| 401 | `REGISTRY_AUTH_FAILED` | Registry authentication failed |
| 404 | `VM_NOT_FOUND` | VM does not exist |
| 404 | `IMAGE_NOT_FOUND` | Image does not exist (locally or in registry) |
| 404 | `SNAPSHOT_NOT_FOUND` | Snapshot does not exist |
| 409 | `INVALID_STATE_TRANSITION` | Operation not valid for current VM state |
| 409 | `VM_LOCKED` | VM locked by another operation |
| 409 | `RESOURCE_IN_USE` | Resource (image, snapshot) in use, cannot delete |
| 422 | `VALIDATION_ERROR` | Input validation failed |
| 422 | `RESOURCE_LIMIT_EXCEEDED` | Requested resources exceed configured limits |
| 422 | `SNAPSHOT_INCOMPATIBLE` | Snapshot incompatible with VM configuration |
| 500 | `INTERNAL_ERROR` | Unexpected server error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |
| 507 | `INSUFFICIENT_STORAGE` | Not enough disk space |

### HTTP Status Codes

- `200 OK` - Request succeeded
- `201 Created` - Resource created
- `202 Accepted` - Async operation accepted
- `204 No Content` - Request succeeded, no content to return
- `400 Bad Request` - Client error (invalid input)
- `401 Unauthorized` - Authentication required or failed
- `404 Not Found` - Resource not found
- `409 Conflict` - State conflict or resource locked
- `422 Unprocessable Entity` - Validation error
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Temporary unavailability
- `507 Insufficient Storage` - Storage exhausted

## Concurrency and Locking

### VM Operation Locking

To prevent race conditions, the API implements pessimistic locking:

1. Before any state-changing operation, the API acquires an exclusive lock on the VM
2. Lock is stored in database: `locked_by` (operation ID), `locked_at` (timestamp)
3. If lock acquisition fails, returns `409 Conflict` with error `VM_LOCKED`
4. Lock is released on operation completion (success or failure)
5. Lock timeout: 5 minutes (prevents deadlock from crashed operations)

**Locked operations:**
- Start, stop, kill, pause, resume
- Snapshot creation
- VM deletion

**Non-locked operations:**
- Get VM status (read-only)
- List VMs (read-only)
- Get logs (read-only)

### Idempotency

The following operations are **idempotent**:

- `POST /vms` with name specified (returns existing VM if name exists)
- `POST /vms/{id}/start` (success if already running)
- `POST /vms/{id}/stop` (success if already stopped)
- `GET` requests (by definition)
- `DELETE` requests (404 if already deleted, but same outcome)

**Non-idempotent operations:**
- `POST /vms` without name (always creates new VM)
- `POST /vms/{id}/snapshots` (each call creates new snapshot)

## Resource Limits

The API enforces configurable resource limits to prevent exhaustion:

### Per-VM Limits
- `max_vcpus`: Maximum vCPUs per VM (default: 8)
- `max_memory_mib`: Maximum memory per VM (default: 8192)

### Global Limits
- `max_vms`: Maximum total VMs (default: 50)
- `max_total_memory_mib`: Maximum total memory across all VMs (default: 32768)
- `max_snapshot_storage_gib`: Maximum snapshot storage (default: 100)

**Validation:**
- Limits checked during VM creation
- Returns `422 RESOURCE_LIMIT_EXCEEDED` if limits violated
- Error includes current usage and limit values

## State Persistence

The API uses SQLite for state persistence:

**Database Location:** `/var/lib/nanofuse/nanofuse.db`

**Schema:**
```sql
CREATE TABLE vms (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE,
  state TEXT NOT NULL,
  image_ref TEXT NOT NULL,
  image_digest TEXT NOT NULL,
  config_json TEXT NOT NULL,
  runtime_json TEXT,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  locked_by TEXT,
  locked_at TIMESTAMP
);

CREATE TABLE snapshots (
  id TEXT PRIMARY KEY,
  vm_id TEXT NOT NULL,
  name TEXT,
  memory_file_path TEXT NOT NULL,
  snapshot_file_path TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  created_at TIMESTAMP NOT NULL,
  FOREIGN KEY (vm_id) REFERENCES vms(id) ON DELETE CASCADE
);

CREATE TABLE images (
  digest TEXT PRIMARY KEY,
  tags_json TEXT NOT NULL,
  architecture TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  kernel_version TEXT,
  rootfs_path TEXT NOT NULL,
  kernel_path TEXT NOT NULL,
  pulled_at TIMESTAMP NOT NULL
);

CREATE TABLE image_pull_jobs (
  id TEXT PRIMARY KEY,
  image_ref TEXT NOT NULL,
  state TEXT NOT NULL,
  progress_json TEXT,
  error TEXT,
  result_digest TEXT,
  created_at TIMESTAMP NOT NULL,
  completed_at TIMESTAMP
);
```

## Failure Recovery

### API Daemon Restart

On API daemon startup:

1. Load all VMs from database
2. Check which Firecracker processes are actually running (by PID)
3. Reconcile state:
   - **Orphaned processes** (running but not in DB): Log warning, optionally adopt or terminate
   - **Zombie entries** (in DB but process dead): Update state to `failed`
   - **Running VMs** (in DB and process alive): Keep as-is

### VM Crashes

When Firecracker process exits unexpectedly:

1. API detects via process monitoring (waitpid)
2. Update VM state to `failed`
3. Log exit code and last console output
4. Preserve logs and state for debugging

## Configuration

API daemon reads configuration from `/etc/nanofuse/nanofused.yaml`:

```yaml
api:
  socket: /var/run/nanofused.sock
  socket_mode: "0660"
  socket_group: nanofuse
  # tcp_bind: "127.0.0.1:8080"  # optional, for remote access

storage:
  data_dir: /var/lib/nanofuse
  database: /var/lib/nanofuse/nanofuse.db

firecracker:
  binary_path: /usr/local/bin/firecracker
  jailer_path: /usr/local/bin/jailer  # optional

limits:
  max_vms: 50
  max_total_memory_mib: 32768
  max_vcpus_per_vm: 8
  max_memory_per_vm_mib: 8192
  max_snapshot_storage_gib: 100

registry:
  auth_config_path: /root/.docker/config.json
  # Alternative: explicit credentials
  # auth:
  #   ghcr.io:
  #     username: user
  #     token: ghp_xxx

logging:
  level: info  # debug, info, warn, error
  format: json  # json, text
  console_log_max_size_mb: 10
  console_log_max_backups: 3
```

## Security Considerations

### Phase 1 (MVP)
1. **Isolation**: Firecracker provides hardware-level virtualization
2. **Access Control**: Unix socket with filesystem permissions
3. **Input Validation**: All API inputs validated (no injection attacks)
4. **Resource Limits**: Prevent exhaustion attacks

### Future Phases
5. **Authentication**: Bearer tokens or mTLS for TCP access
6. **Image Verification**: Signature verification (cosign)
7. **Image Scanning**: Vulnerability scanning (Trivy)
8. **Audit Logging**: Comprehensive audit trail

## Future Enhancements

The following features are explicitly deferred to maintain MVP simplicity while keeping options open:

1. **API Versioning** (at 1.0.0): Add `/v1/` prefix
2. **Authentication** (Phase 2+): Bearer tokens, mTLS
3. **Advanced Networking** (Phase 2+): CNI plugins, custom VPC
4. **Storage Backends** (Phase 3+): NFS, S3, Ceph
5. **Event Webhooks** (Phase 3+): Notify external systems of VM state changes
6. **Metrics** (Phase 3+): Prometheus `/metrics` endpoint
7. **Multi-tenancy** (Phase 4+): Namespace isolation, quotas per tenant
8. **Image Building** (Phase 4+): Build images directly via API

## Architectural Decision Records

### ADR-001: REST over Unix Socket

**Context:** Need local API for CLI communication with daemon.

**Decision:** Use REST over Unix domain socket as primary transport.

**Rationale:**
- Lowest latency (no TCP overhead)
- Simple security model (filesystem permissions)
- Keeps option open for remote management (just change transport, HTTP protocol unchanged)
- Industry standard (Docker, Podman use same approach)

**Consequences:**
- Cannot be accessed remotely without TCP binding (deferred to Phase 2)
- Requires local client (appropriate for MVP use case)

---

### ADR-002: SQLite for State Persistence

**Context:** Need to persist VM state across daemon restarts.

**Decision:** Use SQLite as state database.

**Rationale:**
- Embedded (no separate process)
- ACID transactions (critical for consistency)
- Queryable (filtering, sorting)
- Single file (easy backup)
- Proven in production (used by many production systems)

**Consequences:**
- Not suitable for distributed deployments (single-node only)
- Sufficient for MVP and initial production use

---

### ADR-003: Pessimistic Locking for Concurrency

**Context:** Multiple clients may access API simultaneously.

**Decision:** Use pessimistic locking (acquire lock before operation).

**Rationale:**
- Prevents race conditions on state transitions
- Simple to implement and reason about
- Clear error message when conflict occurs (409 with lock info)

**Consequences:**
- Operations must wait if VM locked
- Lock timeout prevents deadlock but adds complexity

---

### ADR-004: Manual Snapshot Management

**Context:** Need snapshot/resume for fast cold starts.

**Decision:** Snapshots created manually via API, user manages retention.

**Rationale:**
- Gives users full control
- Simpler than automatic policies
- Keeps option open for automation (can add policies later)

**Consequences:**
- Users must explicitly create snapshots
- No automatic cleanup (users manage disk space)

---

### ADR-005: No Explicit Versioning in 0.x

**Context:** API versioning strategy.

**Decision:** No `/v1/` prefix during 0.x phase, add at 1.0.0.

**Rationale:**
- High volatility phase (learning, experimentation)
- Premature versioning adds complexity without value
- At 1.0, commit to stability and add versioning

**Consequences:**
- Breaking changes allowed in 0.x (experimental phase)
- Must add versioning at 1.0 (breaking change itself)

---

## Related Documents

- **CLI Specification**: See `CLI_SPEC.md` for CLI commands that interact with this API
- **Execution Plan**: See `EXECUTION_PLAN.md` for implementation phases
- **Project Overview**: See `../CLAUDE.md` for project context

---

*This specification follows the architectural principles from Gregor Hohpe's works: The Software Architect Elevator, Enterprise Integration Patterns, Cloud Strategy, and Platform Strategy.*
