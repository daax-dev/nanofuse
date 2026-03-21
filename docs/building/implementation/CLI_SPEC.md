# NanoFuse CLI Specification

**Version:** 0.1.0 (Pre-release)
**Date:** 2025-10-30
**Status:** Phase 0 - Architecture Definition

## Overview

The NanoFuse CLI (`nanofuse`) is a Go-based command-line tool for managing Firecracker-based microVMs. It provides a user-friendly interface to the NanoFuse API daemon, supporting image management, VM lifecycle operations, and snapshot/resume functionality.

## Design Principles

This CLI design follows the **Golden Path** principle from Gregor Hohpe's architectural framework:

1. **Sensible Defaults**: Common operations require minimal flags
2. **Progressive Disclosure**: Simple commands for basic use, advanced flags for power users
3. **Clear Feedback**: Human-readable output by default, machine-readable on demand
4. **Build Abstractions Not Illusions**: Expose real failure modes with actionable error messages

## Installation and Distribution

- **Binary Name**: `nanofuse`
- **Distribution**: Single static Go binary (no external dependencies)
- **Target Platforms**: Linux x86_64 (arm64 future)
- **Installation Locations**:
  - User install: `~/.local/bin/nanofuse`
  - System install: `/usr/local/bin/nanofuse`

## Architecture

```
┌─────────────────┐
│  nanofuse CLI   │ (user-facing, static binary)
└────────┬────────┘
         │ HTTP over Unix socket
         │ /var/run/nanofused.sock
         ↓
┌─────────────────┐
│ nanofused API   │ (daemon, systemd service)
└────────┬────────┘
         │
         ↓
    Firecracker VMs
```

**Communication:**
- CLI communicates with API daemon via HTTP over Unix socket
- Default socket: `/var/run/nanofused.sock`
- Configurable via `--api-socket` flag or config file

## Global Options

These flags apply to all commands:

```
--config <path>         Path to config file (default: ~/.config/nanofuse/config.yaml)
--api-socket <path>     API Unix socket path (default: /var/run/nanofused.sock)
--api-url <url>         API URL for remote access (e.g., http://localhost:8080)
--debug                 Enable debug output (verbose logging)
--json                  Output in JSON format (for scripting)
--no-color              Disable colored output
--timeout <duration>    API request timeout (default: 30s, format: 10s, 1m, etc.)
--help, -h              Show help
--version, -v           Show version
```

**Precedence** (highest to lowest):
1. Command-line flags
2. Environment variables
3. Config file
4. Built-in defaults

## Configuration File

**Default Location:** `~/.config/nanofuse/config.yaml`

**Format:**
```yaml
# API Connection
api:
  socket: /var/run/nanofused.sock
  timeout: 30s
  # Alternative: TCP for remote access
  # url: http://localhost:8080
  # auth:
  #   bearer_token: xxx

# Registry Authentication
registry:
  # Default registry for unqualified image references
  default: ghcr.io

  # Authentication (uses Docker credential store by default)
  # Alternatively, specify explicit credentials:
  auth:
    ghcr.io:
      username: ${GITHUB_USER}
      token: ${GITHUB_TOKEN}

# Default VM Configuration
defaults:
  vcpus: 2
  memory_mib: 512
  network_mode: nat

# Output Preferences
output:
  format: table  # table, json, yaml
  color: auto    # auto, always, never
  timezone: local  # local, utc

# Limits (informational, actual limits enforced by API)
limits:
  warn_if_memory_exceeds_mib: 4096
  warn_if_total_vms_exceeds: 20
```

**Environment Variable Substitution:**
- Config file supports `${VAR}` syntax for environment variables
- Example: `token: ${GITHUB_TOKEN}` expands to value of `$GITHUB_TOKEN`

## Command Structure

```
nanofuse <command> <subcommand> [arguments] [flags]
```

### Command Categories

1. **Image Management**: `image <subcommand>`
2. **VM Lifecycle**: `vm <subcommand>`
3. **Snapshot Operations**: `vm snapshot <subcommand>`
4. **System**: `version`, `completion`, `config`

## Commands

### Health Check

#### `nanofuse health`

Check API daemon health and connectivity.

**Usage:**
```bash
nanofuse health
```

**Output (table):**
```
Status:   Healthy
Version:  0.1.0
Uptime:   1h 30m
```

**Output (--json):**
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "uptime_seconds": 5400
}
```

**Exit Codes:**
- `0` - API healthy
- `1` - API unhealthy or unreachable

---

### Image Commands

#### `nanofuse image pull <image-ref>`

Pull an image from a registry.

**Usage:**
```bash
nanofuse image pull ghcr.io/owner/nanofuse-base:latest
nanofuse image pull ghcr.io/owner/nanofuse-base@sha256:abc123...
```

**Flags:**
```
--platform <arch>    Target architecture (default: auto-detect, options: x86_64, arm64)
--no-progress        Disable progress bar
```

**Output (table, with progress):**
```
Pulling ghcr.io/owner/nanofuse-base:latest...
Downloading rootfs... ████████████████████ 100% (500 MB / 500 MB)
Extracting...
Pull complete!

Digest: sha256:abc123...
Size:   500 MB
```

**Output (--json):**
```json
{
  "image_ref": "ghcr.io/owner/nanofuse-base:latest",
  "digest": "sha256:abc123...",
  "size_bytes": 524288000,
  "architecture": "x86_64"
}
```

**Exit Codes:**
- `0` - Success
- `1` - Pull failed (network error, image not found, auth failure)

**Error Examples:**
```
Error: Failed to authenticate with ghcr.io
Hint: Run 'docker login ghcr.io' or set GITHUB_TOKEN in config

Error: Image not found: ghcr.io/owner/nonexistent:latest
```

---

#### `nanofuse image list`

List locally cached images.

**Usage:**
```bash
nanofuse image list
nanofuse image ls  # alias
```

**Flags:**
```
--filter <key=value>   Filter images (e.g., --filter architecture=x86_64)
--format <template>    Custom output format (Go template)
--quiet, -q            Only show digests
```

**Output (table):**
```
DIGEST           TAGS                                    SIZE     PULLED
sha256:abc123... ghcr.io/owner/nanofuse-base:latest     500 MB   2 hours ago
sha256:def456... ghcr.io/owner/nanofuse-worker:v1.0     450 MB   1 day ago
```

**Output (--json):**
```json
{
  "images": [
    {
      "digest": "sha256:abc123...",
      "tags": ["ghcr.io/owner/nanofuse-base:latest"],
      "architecture": "x86_64",
      "size_bytes": 524288000,
      "pulled_at": "2025-10-30T09:00:00Z"
    }
  ],
  "total": 1
}
```

**Output (--quiet):**
```
sha256:abc123...
sha256:def456...
```

---

#### `nanofuse image inspect <image-ref>`

Show detailed image information.

**Usage:**
```bash
nanofuse image inspect ghcr.io/owner/nanofuse-base:latest
nanofuse image inspect sha256:abc123...
```

**Output (table):**
```
Digest:        sha256:abc123...
Tags:          ghcr.io/owner/nanofuse-base:latest
Architecture:  x86_64
Size:          500 MB
Kernel:        5.10.240
Rootfs:        /var/lib/nanofuse/images/sha256:abc123.../rootfs.ext4
Kernel Path:   /var/lib/nanofuse/images/sha256:abc123.../vmlinux
Pulled:        2025-10-30 09:00:00 PST
```

**Output (--json):**
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

---

#### `nanofuse image remove <image-ref>`

Remove a cached image.

**Usage:**
```bash
nanofuse image remove ghcr.io/owner/nanofuse-base:latest
nanofuse image rm sha256:abc123...  # alias
```

**Flags:**
```
--force, -f    Force removal even if image in use (stops referencing VMs)
```

**Output:**
```
Removed image: ghcr.io/owner/nanofuse-base:latest (sha256:abc123...)
Freed: 500 MB
```

**Error (image in use):**
```
Error: Image in use by 2 VMs: web-vm, worker-vm
Hint: Stop VMs first, or use --force to stop and remove
```

---

### VM Commands

#### `nanofuse vm create <image-ref> [name]`

Create a new VM (does not start).

**Usage:**
```bash
nanofuse vm create ghcr.io/owner/nanofuse-base:latest
nanofuse vm create ghcr.io/owner/nanofuse-base:latest web-vm
nanofuse vm create ghcr.io/owner/nanofuse-base:latest --name web-vm  # equivalent
```

**Flags:**
```
--name <name>              VM name (default: auto-generated UUID)
--vcpus <n>                Number of vCPUs (default: 2)
--memory <size>            Memory size (e.g., 512M, 1G, default: 512M)
--network <mode>           Network mode: nat|bridged|none (default: nat)
--bridge <name>            Bridge name (required if --network=bridged)
--kernel-args <args>       Override kernel arguments
--config <path>            Load VM config from file (YAML/JSON)
```

**Output:**
```
Created VM: web-vm
ID:    550e8400-e29b-41d4-a716-446655440000
State: created
Image: ghcr.io/owner/nanofuse-base:latest

Use 'nanofuse vm start web-vm' to start the VM
```

**Output (--json):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "web-vm",
  "state": "created",
  "image": "ghcr.io/owner/nanofuse-base:latest",
  "config": {
    "vcpus": 2,
    "memory_mib": 512,
    "network": {"mode": "nat"}
  }
}
```

**Config File Format** (`--config vm-config.yaml`):
```yaml
name: web-vm
image: ghcr.io/owner/nanofuse-base:latest
config:
  vcpus: 4
  memory_mib: 2048
  network:
    mode: bridged
    bridge_name: br0
  kernel_args: "console=ttyS0 root=/dev/vda1 rw"
```

---

#### `nanofuse vm run <image-ref> [name]`

Create and start a VM in one command (convenience wrapper for `create` + `start`).

**Usage:**
```bash
nanofuse vm run ghcr.io/owner/nanofuse-base:latest web-vm
```

**Flags:** Same as `vm create`

**Output:**
```
Created VM: web-vm
Starting VM...
VM started successfully!

ID:     550e8400-e29b-41d4-a716-446655440000
State:  running
IP:     172.16.0.2

SSH: ssh root@172.16.0.2
```

---

#### `nanofuse vm list`

List all VMs.

**Usage:**
```bash
nanofuse vm list
nanofuse vm ls  # alias
```

**Flags:**
```
--all, -a                Show all VMs (default: only running)
--filter <key=value>     Filter VMs (e.g., --filter state=running)
--quiet, -q              Only show IDs
--format <template>      Custom output format
```

**Output (table):**
```
ID       NAME       STATE     IMAGE                           VCPUS  MEMORY  UPTIME
abc123   web-vm     running   ghcr.io/owner/base:latest       2      512M    2h 30m
def456   worker     stopped   ghcr.io/owner/worker:v1.0       4      1G      -
ghi789   test-vm    failed    ghcr.io/owner/base:latest       2      512M    -
```

**Output (--json):**
```json
{
  "vms": [
    {
      "id": "abc123",
      "name": "web-vm",
      "state": "running",
      "image": "ghcr.io/owner/base:latest",
      "config": {"vcpus": 2, "memory_mib": 512},
      "uptime_seconds": 9000
    }
  ],
  "total": 3
}
```

**Color Coding (table output):**
- `running`: Green
- `paused`: Yellow
- `stopped`: Gray
- `failed`: Red

---

#### `nanofuse vm status <vm-id>`

Show detailed VM status.

**Usage:**
```bash
nanofuse vm status web-vm
nanofuse vm status 550e8400-e29b-41d4-a716-446655440000
```

**Aliases:** `inspect`, `info`

**Output (table):**
```
ID:              550e8400-e29b-41d4-a716-446655440000
Name:            web-vm
State:           running
Image:           ghcr.io/owner/nanofuse-base:latest
Architecture:    x86_64

Configuration:
  vCPUs:         2
  Memory:        512 MB
  Network:       nat
  Kernel Args:   console=ttyS0 root=/dev/vda1 rw

Runtime:
  PID:           12345
  Uptime:        2h 30m
  IP Address:    172.16.0.2
  Gateway:       172.16.0.1
  TAP Device:    tap0
  Console:       /var/lib/nanofuse/vms/550e8400.../console.log

Timestamps:
  Created:       2025-10-30 10:00:00 PST
  Started:       2025-10-30 10:00:05 PST
  Last Updated:  2025-10-30 12:30:00 PST
```

---

#### `nanofuse vm start <vm-id>`

Start a created or stopped VM.

**Usage:**
```bash
nanofuse vm start web-vm
```

**Flags:**
```
--wait, -w       Wait for VM to fully boot (ready state)
--timeout <dur>  Timeout for --wait (default: 60s)
```

**Output:**
```
Starting VM: web-vm
VM started successfully!
State: running
IP:    172.16.0.2
```

**Output (with --wait):**
```
Starting VM: web-vm
Waiting for VM to boot...
✓ VM is ready!
Boot time: 1.8s
IP: 172.16.0.2
```

---

#### `nanofuse vm stop <vm-id>`

Stop a running VM gracefully (ACPI shutdown).

**Usage:**
```bash
nanofuse vm stop web-vm
```

**Flags:**
```
--timeout <dur>    Graceful shutdown timeout (default: 30s, max: 300s)
--force, -f        Force kill if graceful shutdown fails
```

**Output:**
```
Stopping VM: web-vm
Sending shutdown signal...
VM stopped successfully!
```

**Output (with --force after timeout):**
```
Stopping VM: web-vm
Sending shutdown signal...
Graceful shutdown timed out (30s)
Force killing VM...
VM stopped!
```

---

#### `nanofuse vm kill <vm-id>`

Force kill a VM (SIGKILL).

**Usage:**
```bash
nanofuse vm kill web-vm
```

**Output:**
```
Force killing VM: web-vm
VM killed!
```

---

#### `nanofuse vm restart <vm-id>`

Restart a VM (stop + start).

**Usage:**
```bash
nanofuse vm restart web-vm
```

**Flags:**
```
--timeout <dur>    Stop timeout (default: 30s)
--force, -f        Force kill if graceful stop fails
```

**Output:**
```
Restarting VM: web-vm
Stopping VM...
VM stopped!
Starting VM...
VM started successfully!
```

---

#### `nanofuse vm pause <vm-id>`

Pause a running VM (freeze execution).

**Usage:**
```bash
nanofuse vm pause web-vm
```

**Output:**
```
Pausing VM: web-vm
VM paused successfully!
```

---

#### `nanofuse vm resume <vm-id>`

Resume a paused VM or resume from snapshot.

**Usage:**
```bash
nanofuse vm resume web-vm
nanofuse vm resume web-vm --from-snapshot snapshot-123
```

**Flags:**
```
--from-snapshot <id>    Resume from specific snapshot (default: resume from paused state)
```

**Output:**
```
Resuming VM: web-vm
VM resumed successfully!
```

**Output (from snapshot):**
```
Resuming VM: web-vm from snapshot: snapshot-20251030-100530
VM resumed successfully!
Boot time: 0.3s (from snapshot)
```

---

#### `nanofuse vm delete <vm-id>`

Delete a VM (stops if running, removes all state).

**Usage:**
```bash
nanofuse vm delete web-vm
nanofuse vm rm web-vm  # alias
```

**Flags:**
```
--force, -f             Force delete without confirmation
--keep-snapshots        Keep snapshots (only delete VM state)
```

**Output (with confirmation):**
```
Delete VM 'web-vm' (550e8400-e29b-41d4-a716-446655440000)?
This will stop the VM and remove all state. Snapshots will be preserved.
Continue? [y/N]: y

Stopping VM...
Removing VM state...
VM deleted successfully!
```

**Output (--force):**
```
Deleted VM: web-vm
```

---

#### `nanofuse vm logs <vm-id>`

Show VM console logs.

**Usage:**
```bash
nanofuse vm logs web-vm
nanofuse vm logs web-vm --follow
nanofuse vm logs web-vm --tail 100
```

**Flags:**
```
--follow, -f      Stream logs in real-time
--tail <n>        Show last N lines (default: all)
--timestamps, -t  Show timestamps
```

**Output:**
```
[    0.000000] Linux version 5.10.240 ...
[    0.100000] Booting kernel...
[    0.500000] systemd[1]: Starting system...
[    1.234567] systemd[1]: Reached target Multi-User System
```

**Output (with --timestamps):**
```
2025-10-30T10:00:05.000Z [    0.000000] Linux version 5.10.240 ...
2025-10-30T10:00:05.100Z [    0.100000] Booting kernel...
```

---

#### `nanofuse vm exec <vm-id> <command>`

Execute a command inside the VM (requires SSH configured).

**Usage:**
```bash
nanofuse vm exec web-vm ps aux
nanofuse vm exec web-vm "systemctl status nginx"
```

**Flags:**
```
--user <name>    SSH user (default: root)
--timeout <dur>  Command timeout (default: 30s)
```

**Output:**
```
USER  PID  %CPU  %MEM  COMMAND
root  1    0.1   0.5   /lib/systemd/systemd
...
```

**Note:** This is a convenience wrapper around SSH. VM must have SSH running and accessible.

---

### Snapshot Commands

#### `nanofuse vm snapshot create <vm-id> [name]`

Create a snapshot of a running or paused VM.

**Usage:**
```bash
nanofuse vm snapshot create web-vm
nanofuse vm snapshot create web-vm after-boot
```

**Flags:**
```
--name <name>    Snapshot name (default: auto-generated timestamp)
```

**Output:**
```
Creating snapshot for VM: web-vm
Snapshot created successfully!

ID:      snapshot-20251030-100530
Name:    after-boot
Size:    512 MB
Created: 2025-10-30 10:05:30 PST

Use 'nanofuse vm resume web-vm --from-snapshot snapshot-20251030-100530' to resume from this snapshot
```

---

#### `nanofuse vm snapshot list <vm-id>`

List snapshots for a VM.

**Usage:**
```bash
nanofuse vm snapshot list web-vm
nanofuse vm snapshot ls web-vm  # alias
```

**Output (table):**
```
ID                        NAME         SIZE    CREATED
snapshot-20251030-100530  after-boot   512 MB  2 hours ago
snapshot-20251030-120000  pre-deploy   512 MB  30 minutes ago
```

**Output (--json):**
```json
{
  "snapshots": [
    {
      "id": "snapshot-20251030-100530",
      "vm_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "after-boot",
      "size_bytes": 536870912,
      "created_at": "2025-10-30T10:05:30Z"
    }
  ],
  "total": 2
}
```

---

#### `nanofuse vm snapshot inspect <snapshot-id>`

Show detailed snapshot information.

**Usage:**
```bash
nanofuse vm snapshot inspect snapshot-20251030-100530
```

**Output:**
```
ID:           snapshot-20251030-100530
VM ID:        550e8400-e29b-41d4-a716-446655440000
VM Name:      web-vm
Name:         after-boot
Size:         512 MB
Created:      2025-10-30 10:05:30 PST

Files:
  Memory:     /var/lib/nanofuse/snapshots/550e8400.../snapshot-20251030-100530/mem.snap
  Device:     /var/lib/nanofuse/snapshots/550e8400.../snapshot-20251030-100530/vm.snap
```

---

#### `nanofuse vm snapshot delete <snapshot-id>`

Delete a snapshot.

**Usage:**
```bash
nanofuse vm snapshot delete snapshot-20251030-100530
nanofuse vm snapshot rm snapshot-20251030-100530  # alias
```

**Flags:**
```
--force, -f    Force delete without confirmation
```

**Output:**
```
Delete snapshot 'after-boot' (snapshot-20251030-100530)?
This will free 512 MB of disk space.
Continue? [y/N]: y

Deleted snapshot: snapshot-20251030-100530
Freed: 512 MB
```

---

### System Commands

#### `nanofuse version`

Show CLI and API version information.

**Usage:**
```bash
nanofuse version
```

**Output:**
```
CLI Version:  0.1.0
API Version:  0.1.0
Git Commit:   abc123
Built:        2025-10-30T10:00:00Z
Go Version:   go1.21.0
Platform:     linux/amd64
```

**Output (--json):**
```json
{
  "cli_version": "0.1.0",
  "api_version": "0.1.0",
  "git_commit": "abc123",
  "built_at": "2025-10-30T10:00:00Z",
  "go_version": "go1.21.0",
  "platform": "linux/amd64"
}
```

---

#### `nanofuse config view`

Show current configuration (merged from all sources).

**Usage:**
```bash
nanofuse config view
```

**Output:**
```
Configuration Sources:
  Config File:  ~/.config/nanofuse/config.yaml (loaded)
  Environment:  NANOFUSE_API_SOCKET (not set)

Merged Configuration:
  API Socket:        /var/run/nanofused.sock
  API Timeout:       30s
  Default vCPUs:     2
  Default Memory:    512 MB
  Network Mode:      nat
  Output Format:     table
  Color:             auto
```

---

#### `nanofuse config init`

Initialize default configuration file.

**Usage:**
```bash
nanofuse config init
```

**Output:**
```
Creating config file: ~/.config/nanofuse/config.yaml
Default configuration created!

Edit the file to customize settings:
  vi ~/.config/nanofuse/config.yaml
```

---

#### `nanofuse completion <shell>`

Generate shell completion script.

**Usage:**
```bash
nanofuse completion bash
nanofuse completion zsh
nanofuse completion fish
```

**Example:**
```bash
# Bash
nanofuse completion bash > /etc/bash_completion.d/nanofuse

# Zsh
nanofuse completion zsh > "${fpath[1]}/_nanofuse"

# Fish
nanofuse completion fish > ~/.config/fish/completions/nanofuse.fish
```

---

## Output Formats

### Table Format (Default)

Human-readable tables with color coding:

```
ID       NAME       STATE     IMAGE                           UPTIME
abc123   web-vm     running   ghcr.io/owner/base:latest       2h 30m
```

**Features:**
- Color-coded states (green=running, red=failed, etc.)
- Automatically sized columns
- Sorted by creation time (newest first)

### JSON Format (--json)

Machine-readable JSON:

```json
{
  "vms": [...],
  "total": 5
}
```

**Features:**
- Always an object (not bare array)
- ISO 8601 timestamps
- Consistent schema across commands

### YAML Format (--yaml)

Machine-readable YAML (future enhancement):

```yaml
vms:
  - id: abc123
    name: web-vm
    state: running
total: 5
```

### Custom Format (--format)

Go template syntax for custom output:

```bash
nanofuse vm list --format '{{.Name}}\t{{.State}}\t{{.Image}}'
```

Output:
```
web-vm    running    ghcr.io/owner/base:latest
worker    stopped    ghcr.io/owner/worker:v1.0
```

---

## Environment Variables

The CLI respects the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `NANOFUSE_CONFIG` | Config file path | `~/.config/nanofuse/config.yaml` |
| `NANOFUSE_API_SOCKET` | API socket path | `/var/run/nanofused.sock` |
| `NANOFUSE_API_URL` | API URL (TCP) | (none) |
| `NANOFUSE_DEBUG` | Enable debug output | `false` |
| `NANOFUSE_NO_COLOR` | Disable colors | `false` |
| `NO_COLOR` | Disable colors (standard) | `false` |
| `GITHUB_TOKEN` | GitHub PAT for GHCR | (from config or Docker) |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Validation error (invalid input) |
| 3 | API unreachable |
| 4 | Resource not found |
| 5 | Operation conflict (invalid state) |
| 130 | Interrupted (Ctrl+C) |

**Usage in scripts:**
```bash
if ! nanofuse vm start web-vm; then
  echo "Failed to start VM"
  exit 1
fi
```

---

## Error Handling

### Error Message Format

```
Error: <short description>
<detailed error message>

Hint: <actionable suggestion>
```

**Example:**
```
Error: VM not found
Virtual machine 'web-vm' does not exist

Hint: Run 'nanofuse vm list' to see available VMs
```

### Common Error Messages

**API Connection Failed:**
```
Error: Cannot connect to API
Failed to connect to /var/run/nanofused.sock: connection refused

Hint: Ensure nanofused service is running:
  sudo systemctl start nanofused
  sudo systemctl status nanofused
```

**Permission Denied:**
```
Error: Permission denied
Access to /var/run/nanofused.sock denied

Hint: Add your user to the 'nanofuse' group:
  sudo usermod -aG nanofuse $USER
  newgrp nanofuse
```

**Image Not Found:**
```
Error: Image not found locally
Image 'ghcr.io/owner/base:latest' is not cached

Hint: Pull the image first:
  nanofuse image pull ghcr.io/owner/base:latest
```

**Invalid State Transition:**
```
Error: Invalid operation
Cannot start VM 'web-vm': VM is already running

Hint: Check VM status with:
  nanofuse vm status web-vm
```

---

## Logging and Debugging

### Debug Mode

Enable with `--debug` or `NANOFUSE_DEBUG=1`:

```bash
nanofuse --debug vm start web-vm
```

**Debug Output:**
```
DEBUG: Loading config from ~/.config/nanofuse/config.yaml
DEBUG: Connecting to API at unix:///var/run/nanofused.sock
DEBUG: POST /vms/web-vm/start
DEBUG: Request: {"timeout_seconds":30}
DEBUG: Response: 200 OK
DEBUG: Body: {"id":"abc123","state":"starting",...}
Starting VM: web-vm
VM started successfully!
```

### Verbose Mode (Future)

```bash
nanofuse --verbose vm start web-vm
```

Shows:
- Progress for long operations
- Intermediate status updates
- Warnings (non-fatal issues)

---

## Aliases and Shortcuts

**Command Aliases:**
- `ls` → `list`
- `rm` → `remove` / `delete`
- `inspect` → `status`
- `info` → `status`

**Flag Aliases:**
- `-f` → `--force`
- `-a` → `--all`
- `-q` → `--quiet`
- `-h` → `--help`
- `-v` → `--version`

**Example:**
```bash
nanofuse vm ls -a      # Same as: nanofuse vm list --all
nanofuse image rm -f   # Same as: nanofuse image remove --force
```

---

## Examples and Common Workflows

### Quick Start

```bash
# Pull base image
nanofuse image pull ghcr.io/owner/nanofuse-base:latest

# Create and start VM
nanofuse vm run ghcr.io/owner/nanofuse-base:latest web-vm

# Check status
nanofuse vm status web-vm

# View logs
nanofuse vm logs web-vm --follow
```

### Snapshot Workflow

```bash
# Create VM and start
nanofuse vm run ghcr.io/owner/base:latest app-vm

# Wait for app to boot completely
sleep 30

# Create snapshot after boot
nanofuse vm snapshot create app-vm after-boot

# Later: Fast resume from snapshot
nanofuse vm resume app-vm --from-snapshot after-boot
```

### Multi-VM Setup (Trigger.dev Use Case)

```bash
# Pull images
nanofuse image pull ghcr.io/owner/trigger-web:latest
nanofuse image pull ghcr.io/owner/trigger-worker:latest

# Create web VM on bridge
nanofuse vm create ghcr.io/owner/trigger-web:latest web \
  --vcpus 2 --memory 1G --network bridged --bridge br0

# Create worker VM on same bridge
nanofuse vm create ghcr.io/owner/trigger-worker:latest worker \
  --vcpus 4 --memory 2G --network bridged --bridge br0

# Start both
nanofuse vm start web
nanofuse vm start worker

# Check connectivity
nanofuse vm exec web ping worker.local
```

### Cleanup

```bash
# Stop all VMs
nanofuse vm list --quiet | xargs -I {} nanofuse vm stop {}

# Delete all VMs
nanofuse vm list --quiet | xargs -I {} nanofuse vm delete {} --force

# Remove unused images
nanofuse image list --quiet | xargs -I {} nanofuse image remove {}
```

---

## Shell Integration

### Bash

```bash
# ~/.bashrc
eval "$(nanofuse completion bash)"

# Aliases
alias nf='nanofuse'
alias nfvm='nanofuse vm'
alias nfimg='nanofuse image'
```

### Zsh

```zsh
# ~/.zshrc
eval "$(nanofuse completion zsh)"

# Aliases
alias nf='nanofuse'
alias nfvm='nanofuse vm'
alias nfimg='nanofuse image'
```

### Fish

```fish
# ~/.config/fish/config.fish
nanofuse completion fish | source

# Aliases
alias nf='nanofuse'
alias nfvm='nanofuse vm'
alias nfimg='nanofuse image'
```

---

## CLI Implementation Notes

### HTTP Client Configuration

- **Transport**: HTTP over Unix socket using `net/http` with custom `DialContext`
- **Timeout**: Configurable per-request (default: 30s)
- **Retries**: Exponential backoff for network errors (max 3 retries)
- **User-Agent**: `nanofuse-cli/0.1.0`

### Progress Bars

For long operations (image pull, snapshot creation):
- Use `github.com/schollz/progressbar` or similar
- Update every 100ms
- Show: percentage, current/total bytes, elapsed time, ETA
- Disable with `--no-progress` or when output is not a TTY

### Color Support

- Auto-detect TTY (disable colors if not interactive)
- Respect `NO_COLOR` environment variable
- Colors: green (success), red (error), yellow (warning), blue (info)
- Use `github.com/fatih/color` or ANSI escape codes

### Table Rendering

- Use `github.com/olekukonko/tablewriter` or similar
- Auto-size columns based on content
- Headers in bold
- Borders: ASCII or Unicode (based on locale)

---

## Future Enhancements

The following features are explicitly deferred to maintain MVP simplicity:

1. **Interactive Mode** (Phase 3+): `nanofuse shell` for interactive session
2. **Watch Mode** (Phase 2+): `nanofuse vm watch` to continuously monitor VMs
3. **Batch Operations** (Phase 2+): `nanofuse vm stop --all`
4. **Config Profiles** (Phase 3+): Multiple config profiles (dev, prod)
5. **Plugins** (Phase 4+): Extensibility via plugins
6. **Web UI** (Phase 4+): Browser-based management console

---

## Related Documents

- **API Contract**: See `API_CONTRACT.md` for the API that this CLI consumes
- **Execution Plan**: See `EXECUTION_PLAN.md` for implementation phases
- **Project Overview**: See `../CLAUDE.md` for project context

---

*This specification follows the architectural principles from Gregor Hohpe's works: The Software Architect Elevator, Enterprise Integration Patterns, Cloud Strategy, and Platform Strategy.*
