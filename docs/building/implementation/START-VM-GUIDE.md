# How to Start and Keep a VM Running

**Complete step-by-step guide to running VMs with NanoFuse**

---

## Prerequisites

1. ✅ Base image built (already done from earlier steps)
2. ✅ Binaries built (`mage all`)
3. ⚠️ Firecracker installed
4. ⚠️ KVM access (your user in `kvm` group or run with sudo)

---

## Step 1: Setup Configuration

### Create config directory and files

```bash
# Create config directory
mkdir -p ~/.config/nanofuse
mkdir -p /tmp/nanofuse/{images,vms,snapshots,sockets}

# Create daemon config
cat > ~/.config/nanofuse/config.yaml << 'EOF'
api:
  address: "127.0.0.1:8080"
  socket: "/tmp/nanofuse/nanofused.sock"
  log_level: "info"

storage:
  data_dir: "/tmp/nanofuse"
  images_dir: "/tmp/nanofuse/images"
  vms_dir: "/tmp/nanofuse/vms"
  snapshots_dir: "/tmp/nanofuse/snapshots"
  database_path: "/tmp/nanofuse/nanofuse.db"

firecracker:
  binary_path: "/usr/bin/firecracker"
  socket_dir: "/tmp/nanofuse/sockets"
  default_vcpus: 2
  default_memory_mb: 512
  default_kernel_args: "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k"

registry:
  default_registry: "ghcr.io"
  cache_dir: "/tmp/nanofuse/registry-cache"
EOF
```

---

## Step 2: Copy Image to Storage

Since we built the image locally, copy it to where the API expects it:

```bash
# Create image directory structure
mkdir -p /tmp/nanofuse/images/nanofuse-base/latest

# Copy built artifacts
cp images/base/build/rootfs.ext4 /tmp/nanofuse/images/nanofuse-base/latest/
cp images/base/build/vmlinux /tmp/nanofuse/images/nanofuse-base/latest/
cp images/base/build/manifest.json /tmp/nanofuse/images/nanofuse-base/latest/

# Verify
ls -lh /tmp/nanofuse/images/nanofuse-base/latest/
```

---

## Step 3: Start the API Daemon

### Option A: Run in Foreground (recommended for testing)

```bash
./bin/nanofused --config ~/.config/nanofuse/config.yaml
```

**Expected output:**
```
INFO[0000] Starting NanoFuse API daemon version=dev
INFO[0000] Loaded configuration file=~/.config/nanofuse/config.yaml
INFO[0000] Initializing storage dir=/tmp/nanofuse
INFO[0000] Starting HTTP server addr=127.0.0.1:8080
INFO[0000] Starting Unix socket server socket=/tmp/nanofuse/nanofused.sock
```

### Option B: Run in Background

```bash
# Start daemon in background
nohup ./bin/nanofused --config ~/.config/nanofuse/config.yaml > /tmp/nanofuse/daemon.log 2>&1 &
DAEMON_PID=$!
echo $DAEMON_PID > /tmp/nanofuse/daemon.pid

# Check it started
ps -p $DAEMON_PID

# View logs
tail -f /tmp/nanofuse/daemon.log
```

### Verify Daemon is Running

```bash
# Check health via HTTP
curl http://127.0.0.1:8080/health

# Or via CLI
./bin/nanofuse health --api-url http://127.0.0.1:8080
```

**Expected response:**
```json
{
  "status": "healthy",
  "version": "dev",
  "uptime": "5s"
}
```

---

## Step 4: Create and Start a VM

### Using CLI (Recommended)

Open a new terminal (keep daemon running in first terminal):

```bash
# Set API endpoint (or use --api-url flag)
export NANOFUSE_API_URL="http://127.0.0.1:8080"

# Create a VM
./bin/nanofuse vm create my-vm \
  --image nanofuse-base:latest \
  --vcpus 2 \
  --memory 512

# Expected output:
# ✓ VM created successfully
# ID: vm-abc123def456
# Name: my-vm
# Image: nanofuse-base:latest
# State: stopped

# Start the VM
./bin/nanofuse vm start my-vm

# Expected output:
# ✓ VM started successfully
# PID: 12345
# Console: /tmp/nanofuse/vms/vm-abc123def456/console.log
```

### Using API Directly

```bash
# Create VM
curl -X POST http://127.0.0.1:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "image": "nanofuse-base:latest",
    "vcpus": 2,
    "memory_mb": 512
  }'

# Response:
# {
#   "id": "vm-abc123def456",
#   "name": "my-vm",
#   "state": "stopped",
#   ...
# }

# Start VM
curl -X POST http://127.0.0.1:8080/api/v1/vms/my-vm/start

# Response:
# {
#   "id": "vm-abc123def456",
#   "name": "my-vm",
#   "state": "running",
#   "runtime": {
#     "pid": 12345,
#     "socket_path": "/tmp/nanofuse/sockets/vm-abc123def456.sock",
#     "console_path": "/tmp/nanofuse/vms/vm-abc123def456/console.log"
#   }
# }
```

---

## Step 5: Check VM Status and Keep It Running

### Check if VM is Running

```bash
# Via CLI
./bin/nanofuse vm list

# Expected output:
# ID              NAME    STATE     VCPUS  MEMORY  UPTIME
# vm-abc123def456 my-vm   running   2      512MB   30s

# Get detailed status
./bin/nanofuse vm inspect my-vm

# Via API
curl http://127.0.0.1:8080/api/v1/vms/my-vm
```

### View Console Output

```bash
# Via CLI
./bin/nanofuse vm logs my-vm

# Or directly
tail -f /tmp/nanofuse/vms/vm-abc123def456/console.log

# You should see systemd boot messages and login prompt
```

### Check VM Process

```bash
# Find VM process
ps aux | grep firecracker

# Should show something like:
# jpoley  12345  firecracker --api-sock /tmp/nanofuse/sockets/vm-abc123def456.sock ...
```

---

## Step 6: The VM Stays Running

Once started, the VM runs as a **background process** managed by Firecracker:

- ✅ **Persistent**: Runs until explicitly stopped
- ✅ **Survives**: Keeps running even if CLI exits
- ✅ **Managed**: Daemon tracks PID and state
- ✅ **Accessible**: Can view logs anytime
- ⚠️ **Note**: Daemon restart will lose tracking (VM still runs but daemon doesn't know about it)

### To Keep VM Running Long-Term

1. **Keep daemon running** - Don't stop nanofused
2. **Don't stop the VM** - It runs indefinitely
3. **Monitor it**:
   ```bash
   # Check status periodically
   ./bin/nanofuse vm list

   # Watch logs
   tail -f /tmp/nanofuse/vms/$(./bin/nanofuse vm list --json | jq -r '.[0].id')/console.log
   ```

---

## Step 7: Stop the VM (When Needed)

### Graceful Stop

```bash
# Via CLI (sends SIGTERM, waits for shutdown)
./bin/nanofuse vm stop my-vm

# Via API
curl -X POST http://127.0.0.1:8080/api/v1/vms/my-vm/stop
```

### Force Kill

```bash
# Via CLI (sends SIGKILL, immediate termination)
./bin/nanofuse vm kill my-vm

# Via API
curl -X DELETE http://127.0.0.1:8080/api/v1/vms/my-vm/kill
```

### Delete VM

```bash
# Stops and removes VM (including disk files)
./bin/nanofuse vm delete my-vm

# Via API
curl -X DELETE http://127.0.0.1:8080/api/v1/vms/my-vm
```

---

## Complete Example Session

```bash
# Terminal 1: Start daemon
./bin/nanofused --config ~/.config/nanofuse/config.yaml

# Terminal 2: Manage VMs
export NANOFUSE_API_URL="http://127.0.0.1:8080"

# Check health
./bin/nanofuse health
# ✓ API is healthy

# List images
./bin/nanofuse image list
# NAME              TAG     SIZE    CREATED
# nanofuse-base     latest  2.0GB   2025-10-31

# Create VM
./bin/nanofuse vm create test-vm --image nanofuse-base:latest --vcpus 2 --memory 512
# ✓ VM created: vm-abc123

# Start VM
./bin/nanofuse vm start test-vm
# ✓ VM started (PID: 12345)

# Check status
./bin/nanofuse vm list
# ID         NAME      STATE     VCPUS  MEMORY  UPTIME
# vm-abc123  test-vm   running   2      512MB   5s

# View logs
./bin/nanofuse vm logs test-vm
# [boot messages...]
# Ubuntu 24.04.3 LTS localhost ttyS0
# localhost login:

# Let it run... (VM stays running)
# Hours later...

# Still running
./bin/nanofuse vm list
# ID         NAME      STATE     VCPUS  MEMORY  UPTIME
# vm-abc123  test-vm   running   2      512MB   3h15m

# Stop when done
./bin/nanofuse vm stop test-vm
# ✓ VM stopped

# Clean up
./bin/nanofuse vm delete test-vm
# ✓ VM deleted
```

---

## Troubleshooting

### Daemon won't start

```bash
# Check if port is in use
lsof -i :8080

# Check logs
cat /tmp/nanofuse/daemon.log

# Try different port
vim ~/.config/nanofuse/config.yaml  # Change api.address
```

### VM won't start - "permission denied"

```bash
# Check KVM access
ls -l /dev/kvm
# Should be: crw-rw---- 1 root kvm

# Add yourself to kvm group
sudo usermod -a -G kvm $USER
newgrp kvm  # Or logout/login

# Or run with sudo (not recommended for production)
sudo ./bin/nanofused --config ~/.config/nanofuse/config.yaml
```

### VM won't start - "firecracker binary not found"

```bash
# Install Firecracker
curl -LO https://github.com/firecracker-microvm/firecracker/releases/download/v1.13.0/firecracker-v1.13.0-x86_64.tgz
tar xzf firecracker-v1.13.0-x86_64.tgz
sudo cp release-v1.13.0-x86_64/firecracker-v1.13.0-x86_64 /usr/bin/firecracker
sudo chmod +x /usr/bin/firecracker

# Verify
which firecracker
firecracker --version
```

### VM starts but immediately dies

```bash
# Check console logs
./bin/nanofuse vm logs my-vm

# Check if rootfs is readable
ls -l /tmp/nanofuse/images/nanofuse-base/latest/rootfs.ext4
# Should be: -rw-rw-r-- (readable by your user)

# Fix permissions if needed
chmod 664 /tmp/nanofuse/images/nanofuse-base/latest/rootfs.ext4
```

### Can't connect to API

```bash
# Check daemon is running
ps aux | grep nanofused

# Check socket exists
ls -l /tmp/nanofuse/nanofused.sock

# Try HTTP instead of socket
./bin/nanofuse health --api-url http://127.0.0.1:8080
```

---

## What's Implemented vs Not

### ✅ Fully Working

- VM creation (via CLI/API)
- VM start/stop/kill
- VM listing and status
- Console log viewing
- Image management (local copies)
- Health checks
- Multiple VMs running simultaneously

### ⚠️ Partially Working

- Networking (TAP device setup is stubbed)
  - VMs boot but won't have network connectivity yet
  - TODO: Implement TAP device creation
- Image pull from GHCR (code exists but needs testing with auth)

### ❌ Not Yet Implemented

- VM snapshots (stubbed as "not yet implemented")
- VM pause/resume (stubbed)
- VM migration
- GPU passthrough

---

## Next Steps

1. ✅ **Start daemon** - Keep it running in background
2. ✅ **Create VM** - Using built base image
3. ✅ **Start VM** - It runs persistently
4. ✅ **Monitor** - Check logs and status
5. ⏭️ **Networking** - Implement TAP device setup for VM connectivity
6. ⏭️ **Snapshots** - Implement Firecracker snapshot/resume
7. ⏭️ **Production** - Systemd service, monitoring, auto-restart

---

## Production Deployment

For production use, create a systemd service:

```bash
sudo tee /etc/systemd/system/nanofused.service << 'EOF'
[Unit]
Description=NanoFuse API Daemon
After=network.target

[Service]
Type=simple
User=nanofuse
Group=kvm
ExecStart=/usr/local/bin/nanofused --config /etc/nanofuse/config.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable nanofused
sudo systemctl start nanofused
```

---

**Questions?**
- Daemon logs: `/tmp/nanofuse/daemon.log`
- VM console logs: `/tmp/nanofuse/vms/<vm-id>/console.log`
- Config: `~/.config/nanofuse/config.yaml`
