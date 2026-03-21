# NanoFuse CLI

The `nanofuse` CLI is a command-line tool for managing Firecracker-based microVMs through the NanoFuse API daemon.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/jpoley/nanofuse.git
cd nanofuse

# Build the CLI
make build-cli

# Install to /usr/local/bin
sudo make install
```

### Binary Download

Download the latest release from [GitHub Releases](https://github.com/jpoley/nanofuse/releases).

## Quick Start

```bash
# Check API daemon health
nanofuse health

# Pull an image
nanofuse image pull ghcr.io/owner/nanofuse-base:latest

# Create and start a VM
nanofuse vm run ghcr.io/owner/nanofuse-base:latest web-vm --vcpus 2 --memory 1024

# List VMs
nanofuse vm list

# Check VM status
nanofuse vm status web-vm

# Stop VM
nanofuse vm stop web-vm

# Delete VM
nanofuse vm delete web-vm
```

## Commands

### Image Management

- `nanofuse image pull <ref>` - Pull an image from registry
- `nanofuse image list` - List cached images
- `nanofuse image inspect <ref>` - Show image details
- `nanofuse image remove <ref>` - Remove cached image

### VM Lifecycle

- `nanofuse vm create <image> [name]` - Create a new VM (doesn't start)
- `nanofuse vm run <image> [name]` - Create and start a VM
- `nanofuse vm list` - List all VMs
- `nanofuse vm status <id>` - Show detailed VM status
- `nanofuse vm start <id>` - Start a VM
- `nanofuse vm stop <id>` - Stop a VM gracefully
- `nanofuse vm kill <id>` - Force kill a VM
- `nanofuse vm restart <id>` - Restart a VM
- `nanofuse vm pause <id>` - Pause a running VM
- `nanofuse vm resume <id>` - Resume a paused VM
- `nanofuse vm delete <id>` - Delete a VM
- `nanofuse vm logs <id>` - Show VM console logs

### Snapshot Management

- `nanofuse vm snapshot create <vm-id> [name]` - Create a snapshot
- `nanofuse vm snapshot list <vm-id>` - List VM snapshots
- `nanofuse vm snapshot inspect <snapshot-id>` - Show snapshot details
- `nanofuse vm snapshot delete <snapshot-id>` - Delete a snapshot

### System Commands

- `nanofuse health` - Check API daemon health
- `nanofuse version` - Show version information
- `nanofuse config view` - Show current configuration
- `nanofuse config init` - Initialize configuration file
- `nanofuse completion <shell>` - Generate shell completion

## Global Flags

- `--api-socket <path>` - API Unix socket path (default: `/var/run/nanofused.sock`)
- `--api-url <url>` - API URL for remote access
- `--debug` - Enable debug output
- `--json` - Output in JSON format
- `--no-color` - Disable colored output
- `--timeout <duration>` - API request timeout (default: 30s)

## Configuration

Create a configuration file at `~/.config/nanofuse/config.yaml`:

```yaml
# API Connection
api:
  socket: /var/run/nanofused.sock
  timeout: 30s

# Default VM Configuration
defaults:
  vcpus: 2
  memory_mib: 512
  network_mode: nat

# Output Preferences
output:
  format: table
  color: auto
```

Initialize with:

```bash
nanofuse config init
```

## Examples

### Create and manage a VM

```bash
# Pull base image
nanofuse image pull ghcr.io/owner/nanofuse-base:latest

# Create VM with custom configuration
nanofuse vm create ghcr.io/owner/nanofuse-base:latest web-vm \
  --vcpus 4 \
  --memory 2048 \
  --network nat

# Start VM
nanofuse vm start web-vm

# Check status
nanofuse vm status web-vm

# View logs
nanofuse vm logs web-vm --tail 100
```

### Snapshot workflow

```bash
# Start a VM
nanofuse vm run ghcr.io/owner/base:latest app-vm

# Create snapshot after boot
nanofuse vm snapshot create app-vm after-boot

# List snapshots
nanofuse vm snapshot list app-vm

# Resume from snapshot (fast cold start)
nanofuse vm resume app-vm --from-snapshot snapshot-20251030-100530
```

### Multi-VM setup

```bash
# Pull images
nanofuse image pull ghcr.io/owner/trigger-web:latest
nanofuse image pull ghcr.io/owner/trigger-worker:latest

# Create VMs on bridge network
nanofuse vm create ghcr.io/owner/trigger-web:latest web \
  --vcpus 2 --memory 1024 --network bridged --bridge br0

nanofuse vm create ghcr.io/owner/trigger-worker:latest worker \
  --vcpus 4 --memory 2048 --network bridged --bridge br0

# Start both
nanofuse vm start web
nanofuse vm start worker
```

### Cleanup

```bash
# Stop all VMs
nanofuse vm list --json | jq -r '.vms[].id' | xargs -I {} nanofuse vm stop {}

# Delete all VMs
nanofuse vm list --json | jq -r '.vms[].id' | xargs -I {} nanofuse vm delete {} --force

# Remove unused images
nanofuse image list --json | jq -r '.images[].digest' | xargs -I {} nanofuse image remove {}
```

## Output Formats

### Table format (default)

Human-readable tables with color coding:

```bash
$ nanofuse vm list
ID       NAME       STATE     IMAGE                           VCPUS  MEMORY  UPTIME
abc123   web-vm     running   ghcr.io/owner/base:latest       2      512M    2h 30m
def456   worker     stopped   ghcr.io/owner/worker:v1.0       4      1G      -
```

### JSON format

Machine-readable JSON:

```bash
$ nanofuse vm list --json
{
  "vms": [
    {
      "id": "abc123",
      "name": "web-vm",
      "state": "running",
      ...
    }
  ],
  "total": 2
}
```

## Shell Completion

### Bash

```bash
# Load in current shell
source <(nanofuse completion bash)

# Install system-wide
nanofuse completion bash > /etc/bash_completion.d/nanofuse
```

### Zsh

```bash
# Add to .zshrc
nanofuse completion zsh > "${fpath[1]}/_nanofuse"
```

### Fish

```bash
nanofuse completion fish > ~/.config/fish/completions/nanofuse.fish
```

## Error Handling

The CLI provides clear error messages with hints:

```
Error: Cannot connect to API
Failed to connect to /var/run/nanofused.sock: connection refused

Hint: Ensure nanofused service is running:
  sudo systemctl start nanofused
  sudo systemctl status nanofused
```

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Validation error
- `3` - API unreachable
- `4` - Resource not found
- `5` - Operation conflict

## Troubleshooting

### Cannot connect to API

```bash
# Check if daemon is running
sudo systemctl status nanofused

# Start daemon
sudo systemctl start nanofused

# Check socket permissions
ls -l /var/run/nanofused.sock

# Add user to nanofuse group
sudo usermod -aG nanofuse $USER
newgrp nanofuse
```

### Enable debug output

```bash
nanofuse --debug vm start web-vm
```

This shows:
- HTTP requests and responses
- API connection details
- Request/response bodies

## Development

### Building from source

```bash
# Clone repository
git clone https://github.com/jpoley/nanofuse.git
cd nanofuse

# Download dependencies
make deps

# Build CLI
make build-cli

# Run tests
make test

# Build static binary
CGO_ENABLED=0 make build-cli
```

### Project structure

```
cmd/nanofuse/          # CLI entry point
internal/
  client/              # API client (HTTP over Unix socket)
  output/              # Output formatting (tables, JSON)
```

## License

See [LICENSE](../../LICENSE) file.

## Related Documentation

- [API Specification](../../docs/API_CONTRACT.md)
- [CLI Specification](../../docs/CLI_SPEC.md)
- [Architecture Decisions](../../docs/ARCHITECTURE_DECISIONS.md)
