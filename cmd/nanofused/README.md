# nanofused - NanoFuse API Daemon

The NanoFuse API daemon manages the lifecycle of Firecracker-based microVMs, providing a RESTful HTTP API for VM operations, image management, and snapshot/resume functionality.

## Overview

`nanofused` is a systemd service that:
- Manages Firecracker microVM processes
- Provides REST API over Unix socket (or TCP)
- Persists state in SQLite database
- Handles VM lifecycle (create, start, stop, pause, resume)
- Manages OCI image pulls from registries
- Supports VM snapshots for fast cold starts

## Installation

### Build from source

```bash
make build-daemon
sudo install -m 755 cmd/nanofused/nanofused /usr/local/bin/nanofused
```

### Install systemd service

```bash
sudo cp systemd/nanofused.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable nanofused
```

### Configuration

Create configuration directory and file:

```bash
sudo mkdir -p /etc/nanofuse
sudo cp config/nanofused.yaml.example /etc/nanofuse/nanofused.yaml
sudo nano /etc/nanofuse/nanofused.yaml
```

Create data directory:

```bash
sudo mkdir -p /var/lib/nanofuse
```

## Configuration

The daemon reads configuration from `/etc/nanofuse/nanofused.yaml`:

```yaml
api:
  socket: /var/run/nanofused.sock
  socket_mode: "0660"
  socket_group: nanofuse

storage:
  data_dir: /var/lib/nanofuse
  database: /var/lib/nanofuse/nanofuse.db

firecracker:
  binary_path: /usr/local/bin/firecracker

limits:
  max_vms: 50
  max_total_memory_mib: 32768
  max_vcpus_per_vm: 8
  max_memory_per_vm_mib: 8192

logging:
  level: info
  format: json
```

See `config/nanofused.yaml.example` for full configuration options.

## Usage

### Start the daemon

```bash
sudo systemctl start nanofused
```

### Check status

```bash
sudo systemctl status nanofused
```

### View logs

```bash
sudo journalctl -u nanofused -f
```

### Test API connectivity

```bash
curl --unix-socket /var/run/nanofused.sock http://localhost/health
```

## API Endpoints

The daemon implements the full NanoFuse API specification. See `docs/API_CONTRACT.md` for complete documentation.

### Health Check

```bash
GET /health
```

### VM Operations

- `POST /vms` - Create VM
- `GET /vms` - List VMs
- `GET /vms/{id}` - Get VM details
- `DELETE /vms/{id}` - Delete VM
- `POST /vms/{id}/start` - Start VM
- `POST /vms/{id}/stop` - Stop VM
- `POST /vms/{id}/kill` - Force kill VM
- `POST /vms/{id}/pause` - Pause VM
- `POST /vms/{id}/resume` - Resume VM
- `GET /vms/{id}/logs` - Get console logs

### Image Operations

- `GET /images` - List cached images
- `GET /images/{digest}` - Get image details
- `DELETE /images/{digest}` - Delete image
- `POST /images/pull` - Pull image (async)
- `GET /images/jobs/{id}` - Get pull job status

### Snapshot Operations

- `POST /vms/{id}/snapshots` - Create snapshot
- `GET /vms/{id}/snapshots` - List snapshots
- `GET /snapshots/{id}` - Get snapshot details
- `DELETE /snapshots/{id}` - Delete snapshot

## Database Schema

The daemon uses SQLite for state persistence:

- **vms** - VM state and configuration
- **snapshots** - Snapshot metadata
- **images** - Cached OCI images
- **image_pull_jobs** - Async pull operations

Database location: `/var/lib/nanofuse/nanofuse.db`

## Concurrency and Locking

The API implements pessimistic locking to prevent race conditions:

- State-changing operations acquire an exclusive lock
- Lock timeout: 5 minutes
- Returns `409 Conflict` if VM is locked
- Lock automatically released on operation completion

## Failure Recovery

On startup, the daemon reconciles state:

1. Loads all VMs from database
2. Checks which Firecracker processes are running
3. Updates state for crashed/orphaned VMs
4. Continues managing running VMs

## Security Considerations

### Unix Socket Mode

- Default: Unix socket with filesystem permissions
- Socket owned by `root:nanofuse`, mode `0660`
- Add users to `nanofuse` group for access

### TCP Mode (Remote Access)

- Uncomment `tcp_bind` in config
- **No authentication in MVP** - bind to localhost only
- For remote access, use SSH tunneling or reverse proxy with auth

## Troubleshooting

### Daemon won't start

Check logs:
```bash
journalctl -u nanofused -n 50
```

Common issues:
- Firecracker binary not found
- Database directory not writable
- Socket path not accessible

### VMs failing to start

1. Check if Firecracker is installed:
   ```bash
   which firecracker
   ```

2. Check VM logs in database:
   ```bash
   GET /vms/{id}/logs
   ```

3. Verify image exists:
   ```bash
   GET /images
   ```

### Database locked errors

- Only one daemon instance should run
- Check for stale processes:
  ```bash
  ps aux | grep nanofused
  ```

## Development

### Run in development mode

```bash
go run ./cmd/nanofused --config /tmp/nanofused.yaml
```

### Build with debug symbols

```bash
go build -gcflags="all=-N -l" -o nanofused ./cmd/nanofused
```

### Enable debug logging

Set `logging.level: debug` in config file.

## Testing

### Unit tests

```bash
make test
```

### Integration tests

```bash
make test-integration
```

### Manual API testing

```bash
# Start daemon
./nanofused --config test.yaml

# In another terminal
curl --unix-socket /var/run/nanofused.sock http://localhost/health
```

## Architecture

The daemon consists of:

- **API Server** - HTTP handlers, routing, middleware
- **Storage Layer** - SQLite database operations
- **Firecracker Manager** - VM lifecycle management
- **Registry Client** - OCI image pulls
- **Configuration** - YAML config loading

See `docs/ARCHITECTURE_DECISIONS.md` for design rationale.

## Performance

Expected performance characteristics:

- **VM Start Time**: 1-2 seconds (cold start)
- **VM Resume Time**: <500ms (from snapshot)
- **API Latency**: <10ms (local Unix socket)
- **Concurrent VMs**: 50+ per host (configurable)

## Limitations (MVP)

- Single-node only (no distributed state)
- No authentication (Unix socket permissions only)
- Snapshot/resume not fully implemented (stub)
- Basic network setup (NAT mode only in MVP)
- No automatic VM restart policies

## Future Enhancements

- TCP API with authentication (Bearer token, mTLS)
- Full snapshot/resume implementation
- Bridge networking mode
- Auto-restart policies
- Health checks and monitoring
- Metrics endpoint (Prometheus)
- Distributed state (etcd/Consul)

## Related Documentation

- API Contract: `docs/API_CONTRACT.md`
- CLI Specification: `docs/CLI_SPEC.md`
- Architecture Decisions: `docs/ARCHITECTURE_DECISIONS.md`

## License

See LICENSE file in repository root.

## Support

For issues and questions, see the main project README.
