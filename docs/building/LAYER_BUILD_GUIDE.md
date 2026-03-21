# Layer Build Guide

This guide explains how to build and use NanoFuse runtime layers for microVM images.

## Overview

NanoFuse uses a layer-based architecture where rootfs images are composed from multiple reusable layers:

| Layer Type | Purpose | Example |
|------------|---------|---------|
| **base** | Core OS (Ubuntu minimal) | base-os |
| **runtime** | Language runtimes | python-runtime, node-runtime, go-runtime |
| **feature** | Optional capabilities | recording-agent |
| **application** | Application-specific | agent-tools |

## Prerequisites

- Docker installed and running
- Bash shell
- ~10GB disk space for built layers

## Building Layers

### Build a Single Layer

```bash
./scripts/build-layer.sh python-runtime
```

### Build All Layers

```bash
./scripts/build-layer.sh all
```

### List Available Layers

```bash
./scripts/build-layer.sh --list
```

## Available Runtime Layers

### python-runtime (159MB)

Python 3.12 with pip, venv, and common packages.

**Provides:**
- python3.12
- pip3
- venv
- Common packages: requests, pyyaml, httpx

**Use case:** Python-based AI agents, data processing

### node-runtime (466MB)

Node.js 22 LTS with npm, pnpm, and bun.

**Provides:**
- Node.js 22.x
- npm 10.x
- pnpm (latest)
- bun 1.x

**Use case:** JavaScript/TypeScript AI agents, Claude Code CLI

### go-runtime (87MB)

Minimal runtime for compiled Go binaries.

**Provides:**
- CA certificates for HTTPS
- CGO support (libc6)
- Timezone data

**Use case:** Running compiled Go binaries (nanofuse-envd, gateways)

## Layer Structure

Each layer directory contains:

```
layers/<layer-name>/
├── Dockerfile        # Build definition
├── layer.yaml        # Layer metadata (auto-updated on build)
├── rootfs/           # Extracted filesystem (created on build)
├── <layer-name>.tar.gz  # Compressed layer (created on build)
└── hooks/            # Optional lifecycle hooks
    └── post-install.sh
```

### layer.yaml Fields

```yaml
name: "python-runtime"
version: "1.0.0"
sha256: "abc123..."     # Auto-generated on build
size_mb: 159            # Auto-generated on build
description: "Python 3.12 runtime"
type: "runtime"         # base|runtime|feature|application

dependencies:
  - "base-os>=1.0.0"

provides:
  - "python"
  - "pip"

files:
  - path: "/usr/bin/python3"
    mode: "0755"

systemd:
  enable: []
  mask: []

config_schema:
  python_version:
    type: string
    default: "3.12"
```

## Build Process

The `build-layer.sh` script performs these steps:

1. **Docker Build**: Builds the layer image from Dockerfile
2. **Container Create**: Creates a container from the image
3. **Export**: Exports container filesystem via `docker export`
4. **Cleanup**: Removes unnecessary files (apt cache, logs, docs)
5. **Tarball**: Creates compressed tarball
6. **Digest**: Generates SHA256 and updates layer.yaml

## Creating a New Layer

1. **Create directory structure:**

```bash
mkdir -p layers/my-layer/hooks
```

2. **Create Dockerfile:**

```dockerfile
FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse My Layer"
LABEL com.nanofuse.layer.type="runtime"
LABEL com.nanofuse.layer.name="my-layer"

ENV DEBIAN_FRONTEND=noninteractive

# Install your packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    my-package \
    && rm -rf /var/lib/apt/lists/*

# Create layer marker
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: my-layer" > /etc/nanofuse/layers/my-layer.yaml

CMD ["/bin/bash"]
```

3. **Create layer.yaml:**

```yaml
name: "my-layer"
version: "1.0.0"
description: "My custom layer"
type: "runtime"

dependencies:
  - "base-os>=1.0.0"

provides:
  - "my-capability"
```

4. **Build:**

```bash
./scripts/build-layer.sh my-layer
```

## Lifecycle Hooks

Hooks are shell scripts executed during image composition:

### post-install.sh

Runs after layer rootfs is extracted to the target image.

**Environment variables available:**
- `LAYER_NAME` - Name of this layer
- `LAYER_VERSION` - Version of this layer
- `ROOTFS_PATH` - Path to the mounted rootfs
- `CONFIG_*` - Layer config values from manifest

Example:

```bash
#!/bin/bash
set -euo pipefail

echo "[my-layer] Running post-install hook..."

# Set timezone
TIMEZONE="${CONFIG_TIMEZONE:-UTC}"
if [[ -f "${ROOTFS_PATH}/usr/share/zoneinfo/${TIMEZONE}" ]]; then
    ln -sf "/usr/share/zoneinfo/${TIMEZONE}" "${ROOTFS_PATH}/etc/localtime"
fi
```

## Testing Layers

Run layer validation tests:

```bash
go test -v ./internal/layerbuild/... -run "Layer"
```

Tests verify:
- Layer spec consistency (name matches directory)
- Required files exist in rootfs
- SHA256 digests present
- Tarball integrity
- Naming conventions

## Best Practices

1. **Minimize size**: Remove unnecessary files, use `--no-install-recommends`
2. **Clean apt cache**: Always `rm -rf /var/lib/apt/lists/*`
3. **Use multi-stage builds**: For compiled languages
4. **Pin versions**: Use specific package versions for reproducibility
5. **Add layer marker**: Create `/etc/nanofuse/layers/<name>.yaml`
6. **Document provides**: List all capabilities the layer provides

## Troubleshooting

### Build fails with missing dependency

Check the Dockerfile has all required packages. Common issues:
- Missing `unzip` for installers (bun, etc.)
- Missing `curl` or `ca-certificates` for downloads

### Layer too large

- Use Ubuntu minimal base
- Remove docs, man pages, locales
- Use `--no-install-recommends` with apt-get
- Consider multi-stage builds

### SHA256 mismatch after rebuild

This is expected - the tarball content changes each build. The build script automatically updates layer.yaml.

## Integration with Image Manifests

Reference layers in image.manifest.yaml:

```yaml
layers:
  - name: "python-runtime"
    type: "runtime"
    source: "local://layers/python-runtime"
    required: true
    dependencies: ["base-os"]
```

See `images/falcondev-agents/image.manifest.yaml` for a complete example.
