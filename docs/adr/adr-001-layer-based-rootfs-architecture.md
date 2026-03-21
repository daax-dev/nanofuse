# ADR-001: Layer-Based RootFS Architecture

## Status

Proposed

## Context

The current NanoFuse image build system creates monolithic ext4 rootfs images via Docker export. While functional, this approach has significant limitations:

1. **No layer reuse**: Every image rebuild creates a complete 2GB ext4 filesystem
2. **No incremental updates**: Minor changes require full image rebuild
3. **No feature composition**: Adding features (recording, debugging tools) requires Dockerfile modification
4. **Poor cache utilization**: Docker layer cache helps, but final ext4 has no layer concept
5. **Testing complexity**: Cannot test individual layers in isolation

### Current Build Flow

```
Dockerfile → docker build → docker export → tar → ext4 filesystem
```

This produces a single, opaque ext4 image with no way to:
- Add features post-build
- Share base layers across different image variants
- Update individual components without full rebuild
- Dynamically inject capabilities at build time

### Driving Forces

1. **Recording Integration**: Falconweb recording capabilities need to be optionally baked in
2. **Feature Modularity**: Different use cases need different capabilities
3. **Build Efficiency**: Faster builds through layer caching
4. **Testability**: Ability to test layers in isolation
5. **Reproducibility**: Manifest-driven builds for consistent outputs

## Decision

We will implement a **layer-based rootfs architecture** with the following design:

### Layer Types

```
┌─────────────────────────────────────────────────────────────────┐
│                    LAYER ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                    APPLICATION LAYER                     │   │
│   │     (Custom user code, configs, startup scripts)        │   │
│   └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                    FEATURE LAYERS                        │   │
│   │   ┌───────────┐ ┌───────────┐ ┌───────────────────┐     │   │
│   │   │ Recording │ │ Debugging │ │ Monitoring Agent  │     │   │
│   │   │  Agent    │ │   Tools   │ │ (Prometheus, etc) │     │   │
│   │   └───────────┘ └───────────┘ └───────────────────────┘     │   │
│   └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                    RUNTIME LAYER                         │   │
│   │  (Python, Node.js, Go runtime, language-specific deps)  │   │
│   └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                    BASE OS LAYER                         │   │
│   │      (Ubuntu 24.04, systemd, SSH, networking)           │   │
│   └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                    KERNEL ARTIFACT                       │   │
│   │           (vmlinux - separate, versioned)               │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Manifest Format (YAML)

Each image is defined by a manifest that specifies layer composition:

```yaml
# image.manifest.yaml
version: "1.0"
name: "nanofuse-flowspec"
description: "MicroVM image with recording capabilities"

kernel:
  version: "6.1.90"
  source: "local://build/vmlinux"
  sha256: "abc123..."
  cmdline: "console=ttyS0 root=/dev/vda1 rw"

layers:
  - name: "base-os"
    type: "base"
    source: "docker://nanofuse-base:latest"
    sha256: "sha256:def456..."
    required: true

  - name: "python-runtime"
    type: "runtime"
    source: "registry://ghcr.io/nanofuse/layers/python:3.12"
    sha256: "sha256:789abc..."
    condition: "${INCLUDE_PYTHON:-false}"

  - name: "recording-agent"
    type: "feature"
    source: "local://layers/recording-agent"
    sha256: "sha256:uvw123..."
    condition: "${INCLUDE_RECORDING:-false}"
    config:
      vsock_port: 52
      buffer_size_mb: 16

  - name: "debug-tools"
    type: "feature"
    source: "registry://ghcr.io/nanofuse/layers/debug:latest"
    sha256: "sha256:xyz789..."
    condition: "${DEBUG_MODE:-false}"

output:
  format: "ext4"
  size_mb: 2048
  compression: "none"
```

### Build Process

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         NEW BUILD FLOW                                    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│   1. PARSE MANIFEST                                                       │
│      ↓                                                                    │
│   ┌───────────────────────────────────────────────────────────────────┐  │
│   │ Read image.manifest.yaml                                          │  │
│   │ Evaluate conditions (env vars, build args)                        │  │
│   │ Determine active layers                                           │  │
│   └───────────────────────────────────────────────────────────────────┘  │
│      ↓                                                                    │
│   2. FETCH LAYERS                                                         │
│      ↓                                                                    │
│   ┌───────────────────────────────────────────────────────────────────┐  │
│   │ For each active layer:                                            │  │
│   │   - Check cache (by SHA256)                                       │  │
│   │   - If miss: fetch from source (docker/registry/local/url)        │  │
│   │   - Verify SHA256 digest                                          │  │
│   │   - Store in layer cache                                          │  │
│   └───────────────────────────────────────────────────────────────────┘  │
│      ↓                                                                    │
│   3. EXTRACT & COMPOSE                                                    │
│      ↓                                                                    │
│   ┌───────────────────────────────────────────────────────────────────┐  │
│   │ Create empty ext4 filesystem                                      │  │
│   │ For each layer (in order):                                        │  │
│   │   - Extract layer tarball to staging                              │  │
│   │   - Apply layer to rootfs (copy/merge)                            │  │
│   │   - Run layer post-install hooks                                  │  │
│   │   - Record layer metadata in /etc/nanofuse/layers/                │  │
│   └───────────────────────────────────────────────────────────────────┘  │
│      ↓                                                                    │
│   4. FINALIZE                                                             │
│      ↓                                                                    │
│   ┌───────────────────────────────────────────────────────────────────┐  │
│   │ Generate /etc/nanofuse/build-manifest.json                        │  │
│   │ Set permissions and ownership                                     │  │
│   │ Unmount and finalize ext4                                         │  │
│   │ Generate output digest                                            │  │
│   └───────────────────────────────────────────────────────────────────┘  │
│      ↓                                                                    │
│   OUTPUT: rootfs.ext4 + vmlinux + manifest.json                          │
│                                                                           │
└──────────────────────────────────────────────────────────────────────────┘
```

### Layer Source Types

| Source Type | Format | Example |
|-------------|--------|---------|
| `docker://` | Docker image export | `docker://nanofuse-base:latest` |
| `registry://` | OCI registry pull | `registry://ghcr.io/nanofuse/layers/python:3.12` |
| `local://` | Local directory/tarball | `local://layers/recording-agent` |
| `url://` | HTTP(S) tarball download | `url://https://releases.example.com/layer.tar.gz` |

### Layer Structure

Each layer follows a standard structure:

```
layer/
├── layer.yaml           # Layer metadata
├── rootfs/              # Files to add/overlay to rootfs
│   ├── usr/
│   ├── etc/
│   └── ...
├── hooks/
│   ├── pre-install.sh   # Run before extraction
│   └── post-install.sh  # Run after extraction
└── tests/
    └── validate.sh      # Layer-specific validation
```

### Layer Metadata (layer.yaml)

```yaml
name: "recording-agent"
version: "1.0.0"
description: "Falconweb recording agent for session capture"
type: "feature"

dependencies:
  - "base-os>=1.0.0"

provides:
  - "recording"
  - "vsock-communication"

files:
  - path: "/usr/local/bin/record-agent"
    mode: "0755"
  - path: "/etc/systemd/system/record-agent.service"
    mode: "0644"

systemd:
  enable:
    - "record-agent.service"

config_schema:
  vsock_port:
    type: integer
    default: 52
    description: "Virtio-vsock port for host communication"
  buffer_size_mb:
    type: integer
    default: 16
    description: "Recording buffer size in megabytes"
```

## Consequences

### Positive

1. **Modular composition**: Build images with exactly the features needed
2. **Build efficiency**: Layer caching reduces rebuild time significantly
3. **Testability**: Each layer can be tested in isolation
4. **Reproducibility**: SHA256-pinned manifests ensure identical builds
5. **Flexibility**: Easy to add/remove features without Dockerfile changes
6. **Version control**: Each layer can be versioned independently
7. **Reusability**: Base layers shared across multiple image variants

### Negative

1. **Added complexity**: More moving parts than monolithic build
2. **Build tool development**: Need to implement layer composition logic
3. **Cache management**: Need layer cache cleanup strategy
4. **Learning curve**: Team needs to understand layer concepts
5. **Migration effort**: Existing images need layer decomposition

### Neutral

1. **Final output unchanged**: Still produces ext4 + vmlinux artifacts
2. **Firecracker compatibility**: No changes to VM runtime behavior
3. **Docker still used**: Base layers can still be Docker-based

## Alternatives Considered

### Alternative 1: Continue Monolithic Ext4

- **Pros:** Simple, proven, no changes required
- **Cons:** No feature modularity, slow rebuilds, no layer reuse
- **Why rejected:** Does not address recording integration or feature composition needs

### Alternative 2: OCI Image Layer Format

- **Pros:** Industry standard, existing tooling, registry support
- **Cons:** OCI layers designed for containers not VMs, overlayfs complexity
- **Why rejected:** Firecracker expects flat ext4, OCI layers add runtime overhead

### Alternative 3: Runtime OverlayFS

- **Pros:** True layer separation at runtime, smaller base images
- **Cons:** Kernel OverlayFS support needed, performance overhead, complexity
- **Why rejected:** Adds runtime complexity, Firecracker snapshot behavior unclear with overlayfs

### Alternative 4: Dockerfile Multi-Stage Only

- **Pros:** Simple, Docker-native, no new tooling
- **Cons:** Limited flexibility, still monolithic output, hard to conditionally include features
- **Why rejected:** Cannot dynamically compose features at build time

## References

- [Current build.sh implementation](../../images/base/build.sh)
- [Current Dockerfile](../../images/base/Dockerfile)
- [OCI Image Specification](https://github.com/opencontainers/image-spec)
- [Firecracker rootfs documentation](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernel-setup.md)
- Decision log: [decisions.jsonl](../../.specify/features/flowspec-microvm-build/decisions.jsonl)

---

*This ADR follows the [Michael Nygard format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).*
