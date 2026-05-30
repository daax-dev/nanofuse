# NanoFuse API Specification

This directory contains the OpenAPI 3.0 specification for the NanoFuse API.

## Overview

NanoFuse is a Firecracker-based microVM management system that provides a simple REST API for:
- **Virtual Machine Management** - Create, start, stop, pause, resume, and delete microVMs
- **Image Management** - Pull and manage container images from OCI registries
- **Snapshot Management** - Create and restore VM snapshots
- **Backup Management** - Backup VMs to S3 and restore from backups

## API Documentation

The complete API specification is available in [`openapi.yaml`](./openapi.yaml).

### Viewing the Specification

You can view the API documentation in several ways:

#### 1. Swagger Editor (Online)
Visit https://editor.swagger.io/ and paste the contents of `openapi.yaml`

#### 2. Swagger UI (Local)
```bash
# Using npx (requires Node.js)
npx @redocly/cli preview-docs openapi.yaml

# Or using Docker
docker run -p 8080:8080 -v $(pwd):/api swaggerapi/swagger-ui
# Then visit http://localhost:8080
```

#### 3. Redoc (Local)
```bash
npx @redocly/cli preview-docs openapi.yaml --port 8080
# Then visit http://localhost:8080
```

## API Endpoints

### Health
- `GET /health` - Health check
- `GET /capabilities` - Runtime host, KVM, Firecracker, and API transport capabilities

### Virtual Machines
- `GET /vms` - List all VMs
- `POST /vms` - Create a new VM
- `GET /vms/{vmId}` - Get VM details
- `DELETE /vms/{vmId}` - Delete a VM
- `POST /vms/{vmId}/start` - Start a VM
- `POST /vms/{vmId}/stop` - Stop a VM
- `POST /vms/{vmId}/kill` - Force kill a VM
- `POST /vms/{vmId}/pause` - Pause a running VM
- `POST /vms/{vmId}/resume` - Resume a paused VM
- `GET /vms/{vmId}/logs` - Get console logs

### Snapshots
- `GET /vms/{vmId}/snapshots` - List VM snapshots
- `POST /vms/{vmId}/snapshots` - Create a snapshot of a paused VM
- `GET /snapshots/{snapshotId}` - Get snapshot details
- `DELETE /snapshots/{snapshotId}` - Delete a snapshot

### Backups (⚠️ not yet implemented)
- `GET /vms/{vmId}/backups` - List VM backups from S3
- `POST /vms/{vmId}/backups` - Create a backup to S3
- `GET /backups/{backupId}` - Get backup details
- `DELETE /backups/{backupId}` - Delete a backup from S3
- `POST /backups/{backupId}/restore` - Restore a VM from backup

### Images
- `GET /images` - List local images
- `POST /images/pull` - Pull an image from a registry
- `GET /images/{digest}` - Get image details
- `DELETE /images/{digest}` - Delete an image
- `GET /images/jobs/{jobId}` - Get pull job status

## Authentication

Currently, the NanoFuse API does not require authentication when accessed via Unix socket (default).

If using TCP binding, consider implementing authentication middleware or using a reverse proxy with authentication.

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "additional": "context"
    }
  }
}
```

### Error Codes

- `INVALID_REQUEST` - Malformed request
- `VM_NOT_FOUND` - VM does not exist
- `IMAGE_NOT_FOUND` - Image not found locally
- `SNAPSHOT_NOT_FOUND` - Snapshot does not exist
- `INVALID_STATE_TRANSITION` - Cannot perform operation in current state
- `VM_LOCKED` - VM is locked by another operation
- `RESOURCE_IN_USE` - Resource cannot be deleted (still in use)
- `RESOURCE_LIMIT_EXCEEDED` - Resource limits exceeded
- `INTERNAL_ERROR` - Internal server error

## Example Usage

### Create and Start a VM

```bash
# Pull an image
curl -X POST http://localhost:8080/images/pull \
  -H "Content-Type: application/json" \
  -d '{
    "image_ref": "ghcr.io/daax-dev/nanofuse/base:latest"
  }'

# Create a VM
curl -X POST http://localhost:8080/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "image": "ghcr.io/daax-dev/nanofuse/base:latest",
    "config": {
      "vcpus": 2,
      "memory_mib": 512,
      "network": {
        "mode": "nat",
        "port_forwards": [
          {
            "host_port": 8080,
            "vm_port": 80,
            "protocol": "tcp"
          }
        ]
      }
    }
  }'

# Start the VM
curl -X POST http://localhost:8080/vms/{vmId}/start

# Get console logs
curl http://localhost:8080/vms/{vmId}/logs
```

### Create a Snapshot

```bash
# Create a snapshot after pausing the VM
curl -X POST http://localhost:8080/vms/{vmId}/snapshots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "before-upgrade"
  }'

# List snapshots
curl http://localhost:8080/vms/{vmId}/snapshots
```

### Backup to S3 (Planned)

```bash
# Create a backup
curl -X POST http://localhost:8080/vms/{vmId}/backups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "weekly-backup",
    "s3_bucket": "my-backups",
    "s3_prefix": "nanofuse/backups",
    "compression": "zstd"
  }'

# Restore from backup
curl -X POST http://localhost:8080/backups/{backupId}/restore \
  -H "Content-Type: application/json" \
  -d '{
    "name": "restored-vm"
  }'
```

## Implementation Status

### ✅ Implemented
- Health check
- VM lifecycle (create, start, stop, kill, pause, resume, delete)
- Image pull and management
- Snapshot create, list, delete
- Console logs
- Port forwarding

### 🚧 Not Yet Implemented
- **S3 Backups** - Complete backup and restore functionality:
  - Create snapshot from paused VM
  - Compress and upload to S3
  - Download and restore from S3
  - Backup job status tracking
  - Restore job status tracking

## Development

### Validating the Spec

```bash
# Install validator
npm install -g @apidevtools/swagger-cli

# Validate
swagger-cli validate api/openapi.yaml
```

### Generating Client Code

You can generate client libraries in various languages:

```bash
# Install generator
npm install -g @openapitools/openapi-generator-cli

# Generate Go client
openapi-generator-cli generate \
  -i api/openapi.yaml \
  -g go \
  -o clients/go

# Generate Python client
openapi-generator-cli generate \
  -i api/openapi.yaml \
  -g python \
  -o clients/python
```

## Contributing

When adding new endpoints:

1. Update `openapi.yaml` with the new endpoint specification
2. Implement the handler in the appropriate `*_handlers.go` file
3. Update the routing in `internal/api/server.go`
4. Add any new types to `internal/types/`
5. Update this README with the new endpoint
6. Add tests for the new functionality

## License

MIT
