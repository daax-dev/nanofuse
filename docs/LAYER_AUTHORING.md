# Layer Authoring Guide

This guide explains how to create, test, and publish custom layers for NanoFuse microVM images.

## Overview

NanoFuse uses a layer-based architecture where rootfs images are composed from multiple reusable layers. Each layer provides specific capabilities that can be combined to create customized microVM images.

### Layer Types

| Type | Purpose | Example |
|------|---------|---------|
| **base** | Core OS with systemd, networking, SSH | `base-os` |
| **runtime** | Language runtimes (Python, Node.js, Go) | `python-runtime`, `node-runtime` |
| **feature** | Optional capabilities | `recording-agent` |
| **application** | Application-specific tools | `agent-tools` |

### Layer Composition Order

Layers are applied in dependency order:

```
base-os (foundation)
    |
    +-- runtime layers (python-runtime, node-runtime)
    |       |
    |       +-- application layers (agent-tools)
    |
    +-- feature layers (recording-agent)
```

## Creating a New Layer

### Step 1: Create Directory Structure

```bash
mkdir -p layers/my-layer/hooks
cd layers/my-layer
```

### Step 2: Create layer.yaml

The `layer.yaml` file defines metadata and configuration for your layer:

```yaml
# Layer Metadata
name: "my-layer"
version: "1.0.0"
sha256: ""  # Auto-generated on build
size_mb: 0  # Auto-generated on build
description: "Description of what this layer provides"
type: "runtime"  # One of: base, runtime, feature, application

# Dependencies
dependencies:
  - "base-os>=1.0.0"

# Capabilities this layer provides
provides:
  - "my-capability"
  - "another-capability"

# Key files installed by this layer
files:
  - path: "/usr/bin/my-binary"
    mode: "0755"
  - path: "/etc/my-config.yaml"
    mode: "0644"

# Systemd service management
systemd:
  enable:
    - "my-service.service"
  mask: []

# Configuration schema (optional)
config_schema:
  option_name:
    type: string
    default: "default_value"
    description: "Description of this option"
```

### Step 3: Create Dockerfile

The Dockerfile defines how to build the layer filesystem:

```dockerfile
FROM ubuntu:24.04

# Standard labels
LABEL org.opencontainers.image.title="NanoFuse My Layer"
LABEL com.nanofuse.layer.type="runtime"
LABEL com.nanofuse.layer.name="my-layer"

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    my-package \
    another-package \
    && rm -rf /var/lib/apt/lists/*

# Copy configuration files
COPY config/ /etc/my-layer/

# Create layer marker
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: my-layer" > /etc/nanofuse/layers/my-layer.yaml \
    && echo "version: 1.0.0" >> /etc/nanofuse/layers/my-layer.yaml

# Default command (for testing)
CMD ["/bin/bash"]
```

### Step 4: Create Lifecycle Hooks (Optional)

Hooks are shell scripts executed during image composition.

**hooks/post-install.sh**

```bash
#!/bin/bash
# Post-install hook for my-layer
# Runs after layer rootfs is extracted to the target image

set -euo pipefail

echo "[my-layer] Running post-install hook..."

# Environment variables available:
# LAYER_NAME - Name of this layer
# LAYER_VERSION - Version of this layer
# ROOTFS_PATH - Path to the mounted rootfs
# CONFIG_* - Layer config values from manifest

# Example: Set timezone from config
TIMEZONE="${CONFIG_TIMEZONE:-UTC}"
if [[ -f "${ROOTFS_PATH}/usr/share/zoneinfo/${TIMEZONE}" ]]; then
    ln -sf "/usr/share/zoneinfo/${TIMEZONE}" "${ROOTFS_PATH}/etc/localtime"
    echo "${TIMEZONE}" > "${ROOTFS_PATH}/etc/timezone"
fi

# Example: Configure service based on config
if [[ "${CONFIG_ENABLED:-true}" == "true" ]]; then
    # Enable the service
    chroot "${ROOTFS_PATH}" systemctl enable my-service.service
fi

echo "[my-layer] Post-install complete"
```

Make the hook executable:

```bash
chmod +x hooks/post-install.sh
```

## Building Layers

### Build a Single Layer

```bash
./scripts/build-layer.sh my-layer
```

### Build All Layers

```bash
./scripts/build-layer.sh all
```

### Build with Verbose Output

```bash
./scripts/build-layer.sh my-layer --verbose
```

### What the Build Does

1. **Docker Build**: Builds the layer image from Dockerfile
2. **Container Export**: Exports container filesystem via `docker export`
3. **Cleanup**: Removes unnecessary files (apt cache, logs, docs)
4. **Compression**: Creates compressed tarball (`.tar.gz`)
5. **Digest**: Generates SHA256 and updates `layer.yaml`

### Build Output

After building, your layer directory contains:

```
layers/my-layer/
├── Dockerfile
├── layer.yaml        # Updated with sha256 and size_mb
├── rootfs/           # Extracted filesystem
├── my-layer.tar.gz   # Compressed layer
└── hooks/
    └── post-install.sh
```

## Using Layers in Image Manifests

Reference your layer in an `image.manifest.yaml`:

```yaml
version: "1.0"
name: "my-image"
description: "My custom microVM image"

kernel:
  version: "6.1.102"
  source: "local://test/fixtures/kernel/vmlinux"
  cmdline: "console=ttyS0 root=/dev/vda rw init=/sbin/init"

layers:
  # Base OS (required)
  - name: "base-os"
    type: "base"
    source: "local://layers/base-os"
    required: true

  # Your custom layer
  - name: "my-layer"
    type: "runtime"
    source: "local://layers/my-layer"
    required: true
    dependencies:
      - "base-os"
    config:
      option_name: "custom_value"

output:
  path: "./build/my-image"
  format: "ext4"
  size_mb: 2048
```

### Layer Sources

| Source Type | Format | Description |
|-------------|--------|-------------|
| `local://` | `local://layers/my-layer` | Local filesystem path |
| `docker://` | `docker://my-image:tag` | Docker image |
| `registry://` | `registry://ghcr.io/org/layer:v1` | OCI registry (requires sha256) |
| `url://` | `url://https://example.com/layer.tar.gz` | HTTP(S) URL (requires sha256) |

### Conditional Layers

Layers can be conditionally included based on environment variables:

```yaml
layers:
  - name: "recording-agent"
    type: "feature"
    source: "local://layers/recording-agent"
    condition: "${INCLUDE_RECORDING:-true}"
    dependencies:
      - "base-os"
```

## Testing Layers

### Run Layer Validation Tests

```bash
go test -v ./internal/layerbuild/... -run "Layer"
```

### What Tests Verify

- Layer spec consistency (name matches directory)
- Required files exist in rootfs
- SHA256 digests are present
- Tarball integrity
- Naming conventions

### Manual Testing

Build and inspect the layer:

```bash
# Build the layer
./scripts/build-layer.sh my-layer

# Inspect the tarball
tar -tzf layers/my-layer/my-layer.tar.gz | head -20

# Check layer size
du -sh layers/my-layer/rootfs/

# Verify layer marker
cat layers/my-layer/rootfs/etc/nanofuse/layers/my-layer.yaml
```

### Test in a Complete Image

1. Add layer to a manifest
2. Build the image: `nanofuse build -m image.manifest.yaml`
3. Boot a VM: `nanofuse vm run my-image test-vm`
4. Verify layer components are present

## Best Practices

### 1. Minimize Layer Size

```dockerfile
# Use --no-install-recommends
RUN apt-get install -y --no-install-recommends package-name

# Clean up apt cache
RUN rm -rf /var/lib/apt/lists/*

# Remove docs and man pages
RUN rm -rf /usr/share/doc /usr/share/man

# Use multi-stage builds for compiled languages
FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app

FROM ubuntu:24.04
COPY --from=builder /app /usr/local/bin/
```

### 2. Use Proper Base Images

```dockerfile
# Always use the latest LTS Ubuntu version for new layers
FROM ubuntu:24.04

# Avoid older LTS versions when possible (prefer latest for security updates)
# OK but older: FROM ubuntu:22.04  # Supported until 2027
# OK but older: FROM ubuntu:20.04  # Standard support ended April 2025; ESM available until 2030
# AVOID: FROM ubuntu:18.04  # Standard support ended 2023
```

### 3. Pin Package Versions

```dockerfile
# Pin specific versions for reproducibility
RUN apt-get install -y python3.12=3.12.3-1ubuntu2
```

### 4. Add Layer Markers

Always create a layer marker file:

```dockerfile
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: my-layer" > /etc/nanofuse/layers/my-layer.yaml \
    && echo "version: 1.0.0" >> /etc/nanofuse/layers/my-layer.yaml
```

### 5. Document Provides

List all capabilities your layer provides:

```yaml
provides:
  - "python"       # Python interpreter
  - "pip"          # Package manager
  - "venv"         # Virtual environment support
  - "ssl"          # SSL/TLS support
```

### 6. Handle Configuration

Use the config schema for customizable options:

```yaml
config_schema:
  log_level:
    type: string
    default: "info"
    description: "Logging level (debug, info, warn, error)"
    enum: ["debug", "info", "warn", "error"]

  max_connections:
    type: integer
    default: 100
    description: "Maximum concurrent connections"
    minimum: 1
    maximum: 1000
```

## Troubleshooting

### Build Fails with Missing Dependency

Check the Dockerfile has all required packages:

```dockerfile
# Common missing dependencies
RUN apt-get install -y \
    ca-certificates \   # For HTTPS
    curl \              # For downloads
    unzip \             # For extracting archives
    gnupg \             # For key management
```

### Layer Too Large

Analyze what's taking space:

```bash
# Check directory sizes in rootfs
du -h --max-depth=2 layers/my-layer/rootfs/ | sort -h

# Common large directories to clean:
# /usr/share/doc
# /usr/share/man
# /var/cache
# /var/log
```

### SHA256 Mismatch After Rebuild

This is expected - the tarball content changes each build. The build script automatically updates `layer.yaml` with the new digest.

### Layer Not Found During Composition

Verify the source path is correct:

```yaml
# Correct - relative to project root
source: "local://layers/my-layer"

# Wrong - absolute path
source: "local:///home/user/layers/my-layer"
```

### Hook Not Executing

Check hook permissions and path:

```bash
# Verify hook is executable
ls -la layers/my-layer/hooks/post-install.sh

# Hook must be at: hooks/post-install.sh
# Not: scripts/post-install.sh
```

## Available Layers Reference

### base-os

Core Ubuntu 24.04 with systemd, SSH, and networking.

```yaml
provides:
  - systemd
  - ssh
  - networking
  - ca-certificates
```

### python-runtime (159MB)

Python 3.12 with pip, venv, and common packages.

```yaml
provides:
  - python3.12
  - pip3
  - venv
  - httpx
  - pyyaml
```

### node-runtime (466MB)

Node.js 22 LTS with npm, pnpm, and bun.

```yaml
provides:
  - nodejs22
  - npm10
  - pnpm
  - bun
```

### go-runtime (87MB)

Minimal runtime for compiled Go binaries.

```yaml
provides:
  - ca-certificates
  - libc6
  - tzdata
```

### recording-agent (81MB)

Session recording agent for terminal and file I/O capture.

```yaml
provides:
  - session-recording
  - vsock-communication
  - terminal-capture
  - file-io-capture

config_schema:
  vsock_port:
    type: integer
    default: 52
  buffer_size_mb:
    type: integer
    default: 16
```

## Next Steps

- [QUICKSTART.md](QUICKSTART.md) - Get started with NanoFuse
- [RECORDING.md](RECORDING.md) - Set up session recording
- [docs/building/LAYER_BUILD_GUIDE.md](building/LAYER_BUILD_GUIDE.md) - Detailed build guide
- [API_QUICK_START.md](API_QUICK_START.md) - Use the REST API
