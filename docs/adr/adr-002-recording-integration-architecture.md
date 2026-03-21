# ADR-002: Recording Integration Architecture

## Status

Proposed

## Context

Falconweb provides recording capabilities for capturing session interactions. We need to integrate these capabilities into NanoFuse microVMs for:

1. **Session recording**: Capture terminal sessions and code execution
2. **Audit trails**: Compliance and security logging
3. **Debugging**: Replay sessions for troubleshooting
4. **Training data**: Capture interactions for AI model training

### Key Challenges

1. **Ephemeral filesystem**: MicroVM rootfs may be read-only or ephemeral
2. **Storage location**: Where do recordings go? (VM has limited/no persistence)
3. **Low latency**: Recording must not impact execution performance
4. **Reliability**: Recordings must survive VM termination
5. **Size management**: Recordings can grow large during long sessions
6. **Security**: Recordings may contain sensitive data

### Recording Data Flow Requirements

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         MICROVM                                           │
│                                                                           │
│   ┌─────────────────┐      ┌─────────────────┐                           │
│   │  User Session   │ ───► │ Recording Agent │                           │
│   │  (Terminal/IDE) │      │                 │                           │
│   └─────────────────┘      └────────┬────────┘                           │
│                                     │                                     │
│                                     │ Events                              │
│                                     ▼                                     │
│   ┌─────────────────────────────────────────────────────────┐            │
│   │               Local Ring Buffer                          │            │
│   │         (Survives short outages, ~16MB)                  │            │
│   └─────────────────────────────┬───────────────────────────┘            │
│                                 │                                         │
└─────────────────────────────────┼─────────────────────────────────────────┘
                                  │ virtio-vsock
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              HOST                                            │
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    Recording Receiver Service                        │   │
│   │                                                                      │   │
│   │   ┌───────────┐    ┌───────────┐    ┌─────────────────────────┐    │   │
│   │   │  Receive  │ ─► │ Validate  │ ─► │   Route to Storage       │    │   │
│   │   │  Events   │    │ & Buffer  │    │   (Local/S3/GCS/etc)     │    │   │
│   │   └───────────┘    └───────────┘    └─────────────────────────┘    │   │
│   │                                                                      │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   Destinations:                                                             │
│   • Local: /var/lib/nanofuse/recordings/{vm-id}/                           │
│   • S3: s3://recordings-bucket/{tenant}/{vm-id}/                           │
│   • Custom: webhook, streaming endpoint                                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Decision

We will implement recording integration using a **baked-in agent with virtio-vsock communication** to the host:

### 1. Recording Agent Layer

The recording agent is packaged as an optional feature layer:

```yaml
# layers/recording-agent/layer.yaml
name: "recording-agent"
version: "1.0.0"
type: "feature"

provides:
  - "session-recording"
  - "vsock-communication"

dependencies:
  - "base-os>=1.0.0"

config_schema:
  enabled:
    type: boolean
    default: true
  vsock_port:
    type: integer
    default: 52
  buffer_size_mb:
    type: integer
    default: 16
  capture_modes:
    type: array
    items: ["terminal", "file_io", "network", "syscall"]
    default: ["terminal"]
```

### 2. Communication Protocol: Virtio-vsock

**Why virtio-vsock over alternatives:**

| Option | Latency | Reliability | Complexity | Security |
|--------|---------|-------------|------------|----------|
| **virtio-vsock** | ~10μs | High (kernel-level) | Low | High (no network) |
| Network upload | ~1ms+ | Medium (retries) | Medium | Medium (TLS) |
| Shared volume | ~100μs | Medium (sync) | High (locking) | Low (file perms) |
| 9p filesystem | ~50μs | Medium | Medium | Low |

Virtio-vsock provides:
- Direct kernel-to-kernel communication (no network stack)
- ~10μs latency per message
- Reliable stream semantics (like TCP)
- No network configuration required
- Firecracker native support

### 3. Recording Event Format

```protobuf
// recording/events.proto
message RecordingEvent {
  string vm_id = 1;
  string session_id = 2;
  uint64 timestamp_ns = 3;
  EventType type = 4;
  bytes payload = 5;
  map<string, string> metadata = 6;
}

enum EventType {
  SESSION_START = 0;
  SESSION_END = 1;
  TERMINAL_INPUT = 2;
  TERMINAL_OUTPUT = 3;
  FILE_READ = 4;
  FILE_WRITE = 5;
  NETWORK_REQUEST = 6;
  NETWORK_RESPONSE = 7;
  CHECKPOINT = 8;
}
```

### 4. Host-Side Recording Receiver

The nanofused daemon includes a recording receiver that:

1. Listens on vsock port for each VM
2. Receives and validates events
3. Buffers events for batch writing
4. Routes to configured storage backend

```go
// internal/recording/receiver.go
type RecordingReceiver struct {
    vmID        string
    vsockPort   uint32
    storage     RecordingStorage
    buffer      *RingBuffer
    compression CompressionType
}

type RecordingStorage interface {
    Write(ctx context.Context, events []RecordingEvent) error
    Finalize(ctx context.Context, sessionID string) error
    GetPlaybackURL(sessionID string) (string, error)
}
```

### 5. Storage Backend Options

```yaml
# config/nanofused.yaml
recording:
  enabled: true
  storage:
    type: "local"  # local | s3 | gcs | azure-blob | webhook
    local:
      path: "/var/lib/nanofuse/recordings"
      retention_days: 30
    # s3:
    #   bucket: "recordings-bucket"
    #   region: "us-west-2"
    #   prefix: "nanofuse/"
  compression: "zstd"
  encryption:
    enabled: true
    key_source: "vault://secret/recording-key"
```

### 6. Recording Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      RECORDING LIFECYCLE                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   VM Start                                                               │
│      │                                                                   │
│      ▼                                                                   │
│   ┌──────────────────────────────────────────────────────────────────┐  │
│   │ 1. Recording agent starts (if enabled in layer config)          │  │
│   │ 2. Agent connects to host via vsock:52                          │  │
│   │ 3. Agent sends SESSION_START event with VM/session metadata     │  │
│   └──────────────────────────────────────────────────────────────────┘  │
│      │                                                                   │
│      ▼                                                                   │
│   Session Active (recording events continuously)                         │
│      │                                                                   │
│      │  Events flow: Terminal I/O, file ops, network (configurable)    │
│      │                                                                   │
│      ▼                                                                   │
│   ┌──────────────────────────────────────────────────────────────────┐  │
│   │ 4. Periodic CHECKPOINT events (every 30s) for resume points     │  │
│   │ 5. Host receiver batches and writes to storage                  │  │
│   └──────────────────────────────────────────────────────────────────┘  │
│      │                                                                   │
│      ▼                                                                   │
│   VM Stop (graceful or forced)                                          │
│      │                                                                   │
│      ▼                                                                   │
│   ┌──────────────────────────────────────────────────────────────────┐  │
│   │ 6. Agent sends SESSION_END event (if graceful)                   │  │
│   │ 7. Host flushes remaining buffer to storage                     │  │
│   │ 8. Host finalizes recording (index, metadata, encryption)       │  │
│   │ 9. Recording available for playback                             │  │
│   └──────────────────────────────────────────────────────────────────┘  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 7. Build Integration

Recording layer is included via image manifest:

```yaml
# images/flowspec/image.manifest.yaml
layers:
  - name: "base-os"
    source: "docker://nanofuse-base:latest"
    required: true

  - name: "recording-agent"
    source: "local://layers/recording-agent"
    condition: "${INCLUDE_RECORDING:-true}"
    config:
      vsock_port: 52
      capture_modes: ["terminal", "file_io"]
```

## Consequences

### Positive

1. **Low latency**: virtio-vsock adds minimal overhead (~10μs)
2. **Reliability**: Kernel-level transport, survives network issues
3. **Security**: No network exposure, host-controlled storage
4. **Flexibility**: Multiple storage backends supported
5. **Modularity**: Recording is optional feature layer
6. **Playback**: Recordings can be replayed for debugging/training

### Negative

1. **Host dependency**: Requires nanofused receiver running
2. **Firecracker-specific**: virtio-vsock is Firecracker feature
3. **Development effort**: Need to implement recording agent and receiver
4. **Storage costs**: Recordings consume storage (mitigated by compression)

### Neutral

1. **Binary size**: Recording agent adds ~2-3MB to image
2. **CPU overhead**: Minimal unless capturing syscalls
3. **Memory**: Ring buffer uses configurable memory (default 16MB)

## Alternatives Considered

### Alternative 1: No Baked-In Recording

- **Pros:** Simpler images, no agent overhead
- **Cons:** Cannot capture sessions without external tooling
- **Why rejected:** Recording is core requirement for flowspec use case

### Alternative 2: Network-Only Upload

- **Pros:** Works with any VM platform, standard protocols
- **Cons:** Higher latency, network configuration required, less reliable
- **Why rejected:** virtio-vsock provides better performance and reliability

### Alternative 3: Shared Volume/9p

- **Pros:** Simple file-based approach
- **Cons:** File locking complexity, performance overhead, less reliable
- **Why rejected:** Streaming via vsock is more robust for real-time capture

### Alternative 4: Runtime Injection

- **Pros:** No image changes, dynamic enablement
- **Cons:** Complexity, may miss early session events, security concerns
- **Why rejected:** Baked-in agent is more reliable and secure

## References

- [Firecracker vsock documentation](https://github.com/firecracker-microvm/firecracker/blob/main/docs/vsock.md)
- [Linux vsock(7) man page](https://man7.org/linux/man-pages/man7/vsock.7.html)
- [ADR-001: Layer-Based RootFS Architecture](./adr-001-layer-based-rootfs-architecture.md)
- Decision log: [decisions.jsonl](../../.specify/features/flowspec-microvm-build/decisions.jsonl)

---

*This ADR follows the [Michael Nygard format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).*
