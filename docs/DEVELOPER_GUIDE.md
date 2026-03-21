# Nanofuse Developer Guide

This guide covers building, running, debugging, and troubleshooting Nanofuse for developers who want to contribute or modify the codebase.

## Table of Contents

1. [Development Setup](#development-setup)
2. [Building Nanofuse](#building-nanofuse)
3. [Running Locally](#running-locally)
4. [Testing](#testing)
5. [Debugging](#debugging)
6. [Building microVM Images](#building-microvm-images)
7. [Troubleshooting](#troubleshooting)

---

## Development Setup

### Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Linux with KVM | - | Hardware virtualization |
| Go | 1.21+ | Building Go binaries |
| Docker | 20.10+ | Building microVM images |
| Mage | Latest | Build automation |

### Clone and Install Build Tools

```bash
# Clone the repository
git clone https://github.com/peregrinesummit/nanofuse.git
cd nanofuse

# Install Mage build tool
./scripts/ensure-mage.sh

# Verify Mage is installed
mage -l
```

### Directory Structure

```
nanofuse/
├── cmd/                 # CLI and daemon entry points
│   ├── nanofuse/        # CLI tool
│   └── nanofused/       # API daemon
├── internal/            # Core packages
│   ├── api/             # REST API handlers
│   ├── config/          # Configuration management
│   ├── firecracker/     # Firecracker VM management
│   ├── registry/        # OCI registry client
│   └── storage/         # Database and storage
├── images/              # microVM image definitions
│   └── base/            # Ubuntu 24.04 base image
├── docs/                # Documentation
├── scripts/             # Helper scripts
└── test/                # Integration tests
```

---

## Building Nanofuse

### Build All Components

```bash
# Build CLI and daemon
mage all

# Output binaries
ls -lh bin/
# bin/nanofuse      (CLI, ~10MB)
# bin/nanofused     (daemon, ~15MB)
```

### Build Individual Components

```bash
# Build CLI only
mage cli

# Build daemon only
mage daemon
```

### Build with Debug Symbols

```bash
# Build with debug info for VS Code debugging
CGO_ENABLED=1 go build -gcflags="all=-N -l" -o ./bin/nanofused ./cmd/nanofused
```

### Install Binaries

```bash
# Install to ~/bin (recommended for development)
mage installUser

# Install system-wide
sudo mage install
```

---

## Running Locally

### Start the Daemon

```bash
# Start daemon in foreground (for development)
sudo ./bin/nanofused

# Or with debug logging
LOG_LEVEL=debug sudo ./bin/nanofused
```

### Use the CLI

In a separate terminal:

```bash
# List VMs
./bin/nanofuse vm list

# Pull an image
./bin/nanofuse image pull --default

# Create and start a VM
./bin/nanofuse vm run default test-vm

# Check status
./bin/nanofuse vm status test-vm

# View logs
./bin/nanofuse vm logs test-vm

# Stop and delete
./bin/nanofuse vm stop test-vm
./bin/nanofuse vm delete test-vm
```

### Development Configuration

Create a development config file:

```bash
cp config.dev.yaml ~/.nanofuse/config.yaml
```

Key configuration options:

```yaml
api:
  socket_path: "/var/run/nanofused.sock"
  # bind_address: "127.0.0.1:8080"  # Alternative: TCP binding

storage:
  data_dir: "/var/lib/nanofuse"

network:
  bridge_name: "nanofuse0"
  subnet: "172.16.0.0/24"
  gateway: "172.16.0.1"

logging:
  level: "debug"
  format: "text"
```

---

## Testing

### Run Unit Tests

```bash
# Run all unit tests
mage test

# With verbose output
mage testVerbose

# With race detector (slower but catches concurrency bugs)
go test -race ./...

# Generate coverage report
mage testCoverage
```

### Run Integration Tests

Integration tests require a running daemon with KVM access:

```bash
# Start daemon first
sudo ./bin/nanofused &

# Run integration tests
mage testIntegration

# Run specific test
go test -v -run TestVMLifecycle ./test/integration/...
```

### Run Linters

```bash
# Run all linters
mage lint

# Run security checks
mage securityCheck
```

### Pre-commit Checks

Before committing, run the full CI suite locally:

```bash
mage ci
```

---

## Debugging

### VS Code Setup

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug CLI",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/nanofuse",
      "args": ["vm", "list"],
      "cwd": "${workspaceFolder}"
    },
    {
      "name": "Debug Daemon",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/nanofused",
      "cwd": "${workspaceFolder}",
      "env": {
        "LOG_LEVEL": "debug"
      }
    }
  ]
}
```

### Debug with Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug daemon
sudo dlv debug ./cmd/nanofused

# Debug with arguments
dlv debug ./cmd/nanofuse -- vm list
```

### View Logs

```bash
# Daemon logs (if running via systemd)
sudo journalctl -u nanofused -f

# VM console logs
sudo cat /var/lib/nanofuse/vms/<vm-id>/console.log
```

---

## Building microVM Images

### Build the Base Image

```bash
cd images/base

# Build image (requires sudo for mounting)
sudo make build

# Build with custom rootfs size
sudo make build ROOTFS_SIZE=4G

# Validate build artifacts
make validate
```

### Build Artifacts

```
images/base/build/
├── rootfs.ext4      # 2GB ext4 filesystem
├── vmlinux          # Linux 5.10.204 kernel
└── manifest.json    # Image metadata
```

### Test the Image

```bash
# Run automated boot test
cd images/base
make test
```

### Publish to GHCR

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Push image
make push REGISTRY=ghcr.io/peregrinesummit
```

### Build Custom Images

Create a custom image by extending the base:

```dockerfile
# images/myapp/Dockerfile
FROM ghcr.io/peregrinesummit/nanofuse/base:latest

# Install your application
RUN apt-get update && apt-get install -y python3

# Add your service
COPY myapp.service /etc/systemd/system/
RUN systemctl enable myapp.service
```

Build the custom image:

```bash
cd images/myapp
sudo bash build.sh
```

---

## Troubleshooting

### Build Issues

**Mage not found:**

```bash
# Reinstall Mage
./scripts/ensure-mage.sh

# Or install manually
go install github.com/magefile/mage@latest
```

**CGO errors:**

```bash
# Ensure CGO is enabled (required for SQLite)
export CGO_ENABLED=1
mage all
```

### Runtime Issues

**Daemon fails to start:**

```bash
# Check if socket already exists
sudo rm -f /var/run/nanofused.sock

# Check for port conflicts
sudo ss -tlnp | grep 8080

# Check permissions
ls -la /dev/kvm
```

**VM fails to boot:**

```bash
# Check console output
./bin/nanofuse vm logs <vm-name>

# Check daemon logs
sudo journalctl -u nanofused -n 100

# Verify image exists
./bin/nanofuse image list
```

**Network not working:**

```bash
# Check bridge interface
ip addr show nanofuse0

# Check NAT rules
sudo iptables -t nat -L -n

# Restart daemon to recreate network
sudo systemctl restart nanofused
```

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `connection refused` | Daemon not running | Start `nanofused` |
| `image not found` | Image not pulled | Run `nanofuse image pull` |
| `permission denied` | Missing root access | Use `sudo` |
| `KVM not available` | No virtualization | Check `/dev/kvm` exists |
| `bridge creation failed` | Network conflict | Check existing bridges |

### Log Locations

| Log | Location |
|-----|----------|
| Daemon logs | `journalctl -u nanofused` |
| VM console | `/var/lib/nanofuse/vms/<vm-id>/console.log` |
| VM config | `/var/lib/nanofuse/vms/<vm-id>/vm.json` |
| Database | `/var/lib/nanofuse/nanofuse.db` |

### Getting Help

1. Check the [FAQ](FAQ.md) for common questions
2. Review the [Troubleshooting Guide](TROUBLESHOOTING.md) for detailed diagnostics
3. Search [existing issues](https://github.com/peregrinesummit/nanofuse/issues)
4. Open a new issue with logs and reproduction steps

---

## Related Documentation

- [Quick Start Guide](QUICKSTART.md) - Get running quickly
- [API Quick Start](API_QUICK_START.md) - REST API usage
- [Contributing](CONTRIBUTING.md) - How to contribute
- [Architecture Design](firecracker-runner-design.md) - System design details

---

*See also: [README](../README.md) | [GOALS](GOALS.md) | [Troubleshooting](TROUBLESHOOTING.md)*
