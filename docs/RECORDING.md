# Session Recording Guide

This guide explains how to set up and use NanoFuse's session recording capabilities for capturing AI agent interactions and terminal sessions.

## Overview

NanoFuse provides session recording to capture terminal I/O, file operations, and other events from microVM sessions. Recordings are useful for:

- **Auditing**: Track what AI agents do during execution
- **Debugging**: Replay sessions to understand failures
- **Training**: Collect interaction data for model improvements
- **Compliance**: Maintain records of automated operations

### Architecture

```
+------------------+          virtio-vsock         +------------------+
|   Guest VM       | --------------------------->  |   Host Daemon    |
|                  |         port 52               |                  |
| +-------------+  |                               | +-------------+  |
| | record-agent|  |  binary event stream          | | receiver    |  |
| +-------------+  |                               | +-------------+  |
|       |          |                               |       |          |
|   captures:      |                               |   stores:        |
|   - terminal I/O |                               |   - events.bin   |
|   - file ops     |                               |   - metadata.json|
|   - network      |                               |                  |
+------------------+                               +------------------+
                                                          |
                                                   +------v------+
                                                   | REST API    |
                                                   | /recordings |
                                                   +-------------+
```

### Components

| Component | Location | Purpose |
|-----------|----------|---------|
| **Recording Agent** | Guest VM | Captures events and streams via vsock |
| **Receiver** | Host daemon | Receives events and writes to storage |
| **Local Storage** | `/var/lib/nanofuse/recordings/` | Persists recording sessions |
| **API Handlers** | Daemon API | REST endpoints for querying recordings |

## Enabling Recording

### Step 1: Include Recording Layer in Manifest

Add the recording-agent layer to your image manifest:

```yaml
layers:
  # ... other layers ...

  - name: "recording-agent"
    type: "feature"
    source: "local://layers/recording-agent"
    condition: "${INCLUDE_RECORDING:-true}"
    dependencies:
      - "base-os"
    config:
      vsock_port: 52
      buffer_size_mb: 16
      capture_modes:
        - "terminal"
        - "file_io"
```

### Step 2: Build Image with Recording

```bash
# Include recording layer
INCLUDE_RECORDING=true nanofuse build -m image.manifest.yaml

# Or exclude recording for production
INCLUDE_RECORDING=false nanofuse build -m image.manifest.yaml
```

### Step 3: Start Daemon with Recording Enabled

```bash
# Start daemon with recording support
nanofused --recording-enabled --recording-path /var/lib/nanofuse/recordings
```

Configuration in `config.yaml`:

```yaml
recording:
  enabled: true
  storage_path: /var/lib/nanofuse/recordings
  retention_days: 30
  vsock_port: 52
```

### Step 4: Start VM

```bash
# Start VM - recording begins automatically
nanofuse vm run my-image my-vm

# Recording session is created when guest agent connects
```

## Recording Configuration

### Guest Configuration (layer.yaml)

```yaml
config_schema:
  enabled:
    type: boolean
    default: true
    description: "Enable recording agent"

  vsock_port:
    type: integer
    default: 52
    description: "Virtio-vsock port for host communication"

  buffer_size_mb:
    type: integer
    default: 16
    description: "Recording buffer size in megabytes"

  capture_modes:
    type: array
    default:
      - "terminal"
    description: "Capture modes: terminal, file_io, network, syscall"
```

### Host Configuration

| Option | Default | Description |
|--------|---------|-------------|
| `recording.enabled` | `true` | Enable recording receiver |
| `recording.storage_path` | `/var/lib/nanofuse/recordings` | Storage directory |
| `recording.retention_days` | `30` | Days to keep recordings |
| `recording.vsock_port` | `52` | Vsock port to listen on |
| `recording.buffer_size` | `1000` | Event buffer per connection |

## Event Types

The recording system captures these event types:

| Event Type | Description | Payload |
|------------|-------------|---------|
| `SESSION_START` | Session begins | Session metadata |
| `SESSION_END` | Session ends | Final statistics |
| `TERMINAL_INPUT` | User/agent input | Raw bytes |
| `TERMINAL_OUTPUT` | Terminal output | Raw bytes |
| `FILE_READ` | File read operation | Path, offset, size |
| `FILE_WRITE` | File write operation | Path, offset, data |
| `NETWORK_REQUEST` | Outbound network | Request metadata |
| `NETWORK_RESPONSE` | Network response | Response metadata |
| `CHECKPOINT` | Manual checkpoint | Custom payload |

### Event Wire Format

Events are streamed as binary messages (little-endian):

```
[4 bytes] total length
[36 bytes] VM ID (UUID, null-padded)
[36 bytes] Session ID (UUID, null-padded)
[8 bytes] timestamp (nanoseconds since epoch)
[1 byte] event type
[4 bytes] payload length
[N bytes] payload
[4 bytes] metadata count
[...] metadata entries
```

## Recording API

### List All Recordings

```bash
curl http://localhost:8080/api/v1/recordings
```

Response:

```json
{
  "recordings": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "vm_id": "test-vm",
      "started_at": "2025-12-30T10:00:00Z",
      "ended_at": "2025-12-30T10:30:00Z",
      "event_count": 1523,
      "size_bytes": 524288,
      "status": "completed",
      "compressed": true
    }
  ],
  "total": 1
}
```

### List Recordings for a VM

```bash
curl http://localhost:8080/api/v1/vms/{vm_id}/recordings
```

### Get Recording Details

```bash
curl http://localhost:8080/api/v1/recordings/{session_id}
```

Response:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "vm_id": "test-vm",
  "started_at": "2025-12-30T10:00:00Z",
  "ended_at": "2025-12-30T10:30:00Z",
  "event_count": 1523,
  "size_bytes": 524288,
  "status": "completed",
  "compressed": true
}
```

### Get Recording Events

```bash
# Get events with pagination
curl "http://localhost:8080/api/v1/recordings/{session_id}/events?offset=0&limit=100"
```

Response:

```json
{
  "events": [
    {
      "timestamp": "2025-12-30T10:00:01Z",
      "type": "SESSION_START",
      "payload": "",
      "metadata": {
        "agent": "claude-code",
        "version": "1.0.0"
      }
    },
    {
      "timestamp": "2025-12-30T10:00:02Z",
      "type": "TERMINAL_INPUT",
      "payload": "ls -la\n",
      "metadata": {}
    }
  ],
  "offset": 0,
  "limit": 100,
  "total": 1523
}
```

### Finalize Recording

```bash
curl -X POST http://localhost:8080/api/v1/recordings/{session_id}/finalize
```

### Delete Recording

```bash
curl -X DELETE http://localhost:8080/api/v1/recordings/{session_id}
```

## Storage Structure

Recordings are stored in the configured storage path:

```
/var/lib/nanofuse/recordings/
    |
    +-- {session_id}/
    |       |
    |       +-- metadata.json    # Session metadata
    |       +-- events.bin       # Raw events (active)
    |       +-- events.bin.zst   # Compressed (finalized)
    |
    +-- {another_session_id}/
            ...
```

### metadata.json

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "vm_id": "test-vm",
  "started_at": "2025-12-30T10:00:00Z",
  "ended_at": "2025-12-30T10:30:00Z",
  "event_count": 1523,
  "size_bytes": 524288,
  "status": "completed",
  "compressed": true
}
```

### Session States

| State | Description |
|-------|-------------|
| `active` | Recording in progress |
| `completed` | Recording finalized and compressed |
| `failed` | Recording failed (compression error, etc.) |

## Using Recordings

### Playback

Currently, recordings can be read programmatically:

```go
import "github.com/daax-dev/nanofuse/internal/recording"

// Read events from a file
file, _ := os.Open("/var/lib/nanofuse/recordings/{id}/events.bin")
defer file.Close()

for {
    event, err := recording.ReadEvent(file)
    if err == io.EOF {
        break
    }

    fmt.Printf("[%s] %s: %s\n",
        event.Timestamp.Format(time.RFC3339),
        event.Type.String(),
        string(event.Payload))
}
```

### Export to JSON

Recording export is not yet implemented as a built-in CLI command.
You can implement JSON export by writing a helper program that uses the `recording` package
to read events for a given session and serialize them to JSON (see the Go example above).

### Analysis

```python
# Python example for analyzing recordings
import json
from collections import Counter

with open('recording-export.json') as f:
    recording = json.load(f)

# Count event types
event_types = Counter(e['type'] for e in recording['events'])
print("Event distribution:", event_types)

# Find all file writes
writes = [e for e in recording['events'] if e['type'] == 'FILE_WRITE']
print(f"Total file writes: {len(writes)}")
```

## Monitoring

### Check Recording Status

```bash
# Via API
curl http://localhost:8080/api/v1/recordings | jq '.recordings[] | select(.status == "active")'

# Via daemon logs
journalctl -u nanofused | grep "Recording"
```

### Storage Usage

```bash
# Check recording storage size
du -sh /var/lib/nanofuse/recordings/

# List sessions by size
du -sh /var/lib/nanofuse/recordings/* | sort -h
```

### Cleanup Expired Recordings

Recordings older than `retention_days` are automatically cleaned up. Manual cleanup:

```bash
# Delete specific recording
curl -X DELETE http://localhost:8080/api/v1/recordings/{session_id}

# Delete all recordings for a VM
for id in $(curl -s http://localhost:8080/api/v1/vms/{vm_id}/recordings | jq -r '.recordings[].id'); do
    curl -X DELETE http://localhost:8080/api/v1/recordings/$id
done
```

## Troubleshooting

### Recording Not Starting

1. Check recording layer is included:
   ```bash
   nanofuse image inspect my-image | grep recording-agent
   ```

2. Verify daemon has recording enabled:
   ```bash
   grep recording /etc/nanofuse/config.yaml
   ```

3. Check vsock connectivity:
   ```bash
   # In guest VM
   systemctl status record-agent
   journalctl -u record-agent
   ```

### Events Not Appearing

1. Check receiver is listening:
   ```bash
   journalctl -u nanofused | grep "Recording receiver"
   ```

2. Verify vsock port:
   ```bash
   # Should show listening on port 52
   ss -l | grep vsock
   ```

### Large Recording Files

1. Recordings are compressed on finalization
2. Check compression status:
   ```bash
   ls -la /var/lib/nanofuse/recordings/{session_id}/
   # Should have events.bin.zst, not events.bin
   ```

3. Force finalization:
   ```bash
   curl -X POST http://localhost:8080/api/v1/recordings/{session_id}/finalize
   ```

### Storage Full

1. Check retention policy:
   ```yaml
   recording:
     retention_days: 7  # Reduce retention
   ```

2. Manual cleanup:
   ```bash
   # Remove old recordings (safer pattern with maxdepth/mindepth)
   # Using {} + batches arguments for efficiency and handles special characters
   find "/var/lib/nanofuse/recordings" -maxdepth 1 -mindepth 1 -mtime +7 -type d -exec rm -rf {} +
   ```

## Security Considerations

### Access Control

- Recording storage should have restricted permissions (0700)
- API endpoints may contain sensitive data
- Consider encrypting recordings at rest

### Data Sensitivity

Recordings may contain:
- Credentials entered at terminal
- API keys in environment variables
- Sensitive file contents

Best practices:
- Limit capture modes to what's needed
- Implement log scrubbing for sensitive data
- Set appropriate retention policies
- Encrypt recordings in transit and at rest

### Compliance

For compliance requirements:
- Enable audit logging for API access
- Implement recording integrity verification
- Consider immutable storage backends
- Document data retention policies

## Next Steps

- [LAYER_AUTHORING.md](LAYER_AUTHORING.md) - Create custom layers
- [QUICKSTART.md](QUICKSTART.md) - Get started with NanoFuse
- [API_QUICK_START.md](API_QUICK_START.md) - Full API documentation
- [docs/adr/adr-002-recording-integration-architecture.md](adr/adr-002-recording-integration-architecture.md) - Architecture decision record
