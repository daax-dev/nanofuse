# NanoFuse API Quick Start Guide

**Get the NanoFuse API daemon running in 5 minutes**

This guide shows you how to start the NanoFuse API daemon and make your first API calls. Perfect for API users, automation, and integrations.

---

## Prerequisites

- Linux host with KVM support (`/dev/kvm`)
- Root access (required for networking and Firecracker)
- Go 1.21+ (if building from source)

---

## Step 1: Install NanoFuse

### Option A: Download Pre-built Binaries (Fastest)

```bash
# Download latest release
VERSION=v0.1.0  # Check https://github.com/daax-dev/nanofuse/releases for latest
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofused

# Make executable and install
chmod +x nanofused
sudo mv nanofused /usr/local/bin/
```

### Option B: Build from Source

```bash
# Clone repository
git clone https://github.com/daax-dev/nanofuse.git
cd nanofuse

# Install build tool
./scripts/ensure-mage.sh

# Build daemon
mage daemon

# Binary created at: ./bin/nanofused
```

---

## Step 2: Start the API Daemon

### Quick Start (Manual)

```bash
# Start daemon in foreground (for testing)
sudo nanofused

# Output:
# time="2025-11-09T16:30:00Z" level=info msg="Starting NanoFuse API daemon"
# time="2025-11-09T16:30:00Z" level=info msg="Listening on Unix socket" path=/var/run/nanofused.sock
# time="2025-11-09T16:30:00Z" level=info msg="API server ready"
```

The daemon is now running and listening on Unix socket: `/var/run/nanofused.sock`

### Production (systemd Service)

```bash
# Install systemd service
sudo cp systemd/nanofused.service /etc/systemd/system/
sudo systemctl daemon-reload

# Start and enable on boot
sudo systemctl enable --now nanofused

# Check status
sudo systemctl status nanofused

# View logs
sudo journalctl -u nanofused -f
```

---

## Step 3: Test the API

### Health Check

```bash
# Using curl with Unix socket
curl --unix-socket /var/run/nanofused.sock http://localhost/health

# Expected response:
# {"status":"ok","timestamp":"2025-11-09T16:30:05Z"}
```

**Success!** 🎉 Your API is running.

---

## Basic API Operations

### 1. List Images

```bash
# List all pulled images
curl --unix-socket /var/run/nanofused.sock \
  http://localhost/images

# Response (empty initially):
# {"images":[]}
```

### 2. Pull an Image

```bash
# Pull the default base image
curl --unix-socket /var/run/nanofused.sock \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "image": "ghcr.io/daax-dev/nanofuse/base:latest",
    "auth": {
      "username": "YOUR_GITHUB_USERNAME",
      "password": "YOUR_GITHUB_TOKEN"
    }
  }' \
  http://localhost/images/pull

# Response:
# {
#   "job_id": "pull-550e8400-e29b-41d4-a716-446655440000",
#   "status": "started",
#   "message": "Image pull started"
# }
```

**Note**: GitHub Container Registry requires authentication. Create a token at:
https://github.com/settings/tokens/new?scopes=read:packages

### 3. Check Pull Status

```bash
# Check image pull job status
curl --unix-socket /var/run/nanofused.sock \
  http://localhost/images/jobs/pull-550e8400-e29b-41d4-a716-446655440000

# Response (in progress):
# {
#   "job_id": "pull-550e8400-e29b-41d4-a716-446655440000",
#   "status": "running",
#   "progress": 45,
#   "message": "Pulling layer 3/7"
# }

# Response (completed):
# {
#   "job_id": "pull-550e8400-e29b-41d4-a716-446655440000",
#   "status": "completed",
#   "digest": "sha256:abc123...",
#   "message": "Image pulled successfully"
# }
```

### 4. Create a VM

```bash
# Create a VM from the pulled image
curl --unix-socket /var/run/nanofused.sock \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-first-vm",
    "image": "ghcr.io/daax-dev/nanofuse/base:latest",
    "vcpus": 2,
    "memory_mib": 512
  }' \
  http://localhost/vms

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "name": "my-first-vm",
#   "state": "created",
#   "image": "ghcr.io/daax-dev/nanofuse/base:latest",
#   "created_at": "2025-11-09T16:35:00Z"
# }
```

### 5. Start the VM

```bash
# Start the VM
VM_ID="550e8400-e29b-41d4-a716-446655440000"

curl --unix-socket /var/run/nanofused.sock \
  -X POST \
  http://localhost/vms/${VM_ID}/start

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "state": "running",
#   "network": {
#     "ip": "172.16.0.10",
#     "gateway": "172.16.0.1",
#     "tap_device": "tap0"
#   }
# }
```

### 6. Check VM Status

```bash
# Get VM details
curl --unix-socket /var/run/nanofused.sock \
  http://localhost/vms/${VM_ID}

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "name": "my-first-vm",
#   "state": "running",
#   "image": "ghcr.io/daax-dev/nanofuse/base:latest",
#   "config": {
#     "vcpus": 2,
#     "memory_mib": 512
#   },
#   "network": {
#     "ip": "172.16.0.10",
#     "gateway": "172.16.0.1"
#   },
#   "runtime": {
#     "pid": 12345,
#     "uptime_seconds": 42
#   }
# }
```

### 7. View VM Console Logs

```bash
# Get all console output
curl --unix-socket /var/run/nanofused.sock \
  http://localhost/vms/${VM_ID}/logs

# Get last 50 lines
curl --unix-socket /var/run/nanofused.sock \
  "http://localhost/vms/${VM_ID}/logs?tail=50"

# Response:
# {
#   "logs": "[    0.000000] Linux version 5.10.204...\n[    0.001234] Boot successful..."
# }
```

### 8. Stop the VM

```bash
# Gracefully stop VM
curl --unix-socket /var/run/nanofused.sock \
  -X POST \
  http://localhost/vms/${VM_ID}/stop

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "state": "stopped"
# }
```

### 9. Delete the VM

```bash
# Delete VM and cleanup resources
curl --unix-socket /var/run/nanofused.sock \
  -X DELETE \
  http://localhost/vms/${VM_ID}

# Response:
# {
#   "message": "VM deleted successfully"
# }
```

### 10. List All VMs

```bash
# List all VMs
curl --unix-socket /var/run/nanofused.sock \
  http://localhost/vms

# Response:
# {
#   "vms": [
#     {
#       "id": "550e8400-e29b-41d4-a716-446655440000",
#       "name": "my-first-vm",
#       "state": "running",
#       "created_at": "2025-11-09T16:35:00Z"
#     }
#   ]
# }
```

---

## Complete Workflow Example

Here's a complete script that creates, starts, checks, and stops a VM:

```bash
#!/bin/bash
set -e

SOCKET="/var/run/nanofused.sock"
API="http://localhost"

# Function to call API
api_call() {
  curl --silent --unix-socket "$SOCKET" "$@"
}

echo "1. Checking API health..."
api_call "$API/health" | jq .

echo -e "\n2. Creating VM..."
VM_RESPONSE=$(api_call -X POST -H "Content-Type: application/json" \
  -d '{"name":"test-vm","image":"ghcr.io/daax-dev/nanofuse/base:latest","vcpus":2,"memory_mib":512}' \
  "$API/vms")
VM_ID=$(echo "$VM_RESPONSE" | jq -r .id)
echo "VM ID: $VM_ID"

echo -e "\n3. Starting VM..."
api_call -X POST "$API/vms/$VM_ID/start" | jq .

echo -e "\n4. Waiting 5 seconds for boot..."
sleep 5

echo -e "\n5. Checking VM status..."
api_call "$API/vms/$VM_ID" | jq .

echo -e "\n6. Getting VM logs (last 10 lines)..."
api_call "$API/vms/$VM_ID/logs?tail=10" | jq -r .logs

echo -e "\n7. Stopping VM..."
api_call -X POST "$API/vms/$VM_ID/stop" | jq .

echo -e "\n8. Deleting VM..."
api_call -X DELETE "$API/vms/$VM_ID" | jq .

echo -e "\n✅ Complete workflow finished successfully!"
```

Save this as `test-api.sh`, make it executable, and run:

```bash
chmod +x test-api.sh
sudo ./test-api.sh
```

---

## Using TCP Instead of Unix Socket (Optional)

By default, NanoFuse uses a Unix socket for security. To enable TCP:

1. **Edit daemon configuration** (`/etc/nanofuse/config.yaml`):

```yaml
api:
  bind_address: "127.0.0.1:8080"  # TCP binding
  # socket_path: "/var/run/nanofused.sock"  # Comment out Unix socket
```

2. **Restart daemon**:

```bash
sudo systemctl restart nanofused
```

3. **Use regular curl**:

```bash
# Now use TCP
curl http://localhost:8080/health

# All API calls work the same way
curl http://localhost:8080/vms
```

**Warning**: TCP binding without TLS is insecure. Only use on localhost or with a reverse proxy (nginx, caddy) that provides TLS.

---

## Authentication & Security

### Unix Socket (Default - Recommended)

- **Security**: Filesystem permissions control access
- **Socket location**: `/var/run/nanofused.sock`
- **Permissions**: `0660` (owner: root, group: nanofuse)

**To grant a user access**:

```bash
# Add user to nanofuse group
sudo usermod -aG nanofuse username

# User must log out and back in for group to take effect
```

### TCP (Future - Phase 2)

Phase 2 will add:
- Bearer token authentication
- mTLS (mutual TLS)
- API keys
- Role-based access control (RBAC)

---

## API Reference Summary

### Health & Status

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/version` | API version info |

### VM Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/vms` | List all VMs |
| POST | `/vms` | Create VM |
| GET | `/vms/{id}` | Get VM details |
| POST | `/vms/{id}/start` | Start VM |
| POST | `/vms/{id}/stop` | Stop VM |
| DELETE | `/vms/{id}` | Delete VM |
| GET | `/vms/{id}/logs` | Get console logs |

### Image Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/images` | List pulled images |
| POST | `/images/pull` | Pull OCI image |
| GET | `/images/jobs/{id}` | Check pull status |
| GET | `/images/{digest}/labels` | Get image labels |
| DELETE | `/images/{digest}` | Delete image |

### Snapshots (Phase 2)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/vms/{id}/snapshots` | Create snapshot |
| GET | `/vms/{id}/snapshots` | List snapshots |
| POST | `/vms/{id}/resume` | Resume from snapshot |
| DELETE | `/vms/{id}/snapshots/{snapshot_id}` | Delete snapshot |

---

## Troubleshooting

### Daemon won't start

**Error**: `failed to create socket: permission denied`

**Solution**: Run with sudo:
```bash
sudo nanofused
```

---

**Error**: `failed to bind socket: address already in use`

**Solution**: Another instance is running or socket file exists:
```bash
sudo pkill nanofused
sudo rm -f /var/run/nanofused.sock
sudo nanofused
```

---

### API calls fail

**Error**: `curl: (7) Couldn't connect to server`

**Solution**: Check daemon is running:
```bash
ps aux | grep nanofused
sudo systemctl status nanofused
```

---

**Error**: `Permission denied`

**Solution**: Use sudo or add user to nanofuse group:
```bash
sudo usermod -aG nanofuse $(whoami)
# Log out and back in
```

---

### VM creation fails

**Error**: `image not found`

**Solution**: Pull the image first:
```bash
curl --unix-socket /var/run/nanofused.sock \
  -X POST -H "Content-Type: application/json" \
  -d '{"image":"ghcr.io/daax-dev/nanofuse/base:latest","auth":{...}}' \
  http://localhost/images/pull
```

---

**Error**: `failed to start VM: no KVM support`

**Solution**: Ensure KVM is available:
```bash
# Check KVM device exists
ls -l /dev/kvm

# Check KVM kernel module loaded
lsmod | grep kvm

# Check user has access
sudo usermod -aG kvm $(whoami)
```

---

## Next Steps

- **Full API Documentation**: See [docs/building/implementation/API_CONTRACT.md](../docs/building/implementation/API_CONTRACT.md)
- **CLI Usage**: Most API operations have CLI equivalents via `nanofuse` command
- **Production Deployment**: See [CONTRIBUTING.md](CONTRIBUTING.md) for systemd setup
- **Development**: See [DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md) for building and debugging

---

## Examples Repository

Find more examples in the repository:

- **Integration Tests**: `test/integration/api_integration_test.go` - Go examples
- **Network Testing**: `scripts/test-network-e2e.sh` - Full E2E test
- **CLI Scripts**: `cmd/nanofuse/` - CLI implementation showing API usage

---

## Getting Help

- **Issues**: https://github.com/daax-dev/nanofuse/issues
- **Discussions**: https://github.com/daax-dev/nanofuse/discussions
- **Documentation**: https://github.com/daax-dev/nanofuse/tree/main/docs

---

**Last Updated**: 2025-11-09
**API Version**: 0.1.0
**Status**: Phase 1 Complete, Production Ready
