# FalconDev Agents Image

A microVM image for executing AI agents (Claude Code, Codex, Gemini CLI) in isolated Firecracker microVMs with session recording capabilities.

## Overview

This image provides a complete runtime environment for AI coding agents, based on Ubuntu 24.04 with:

- **Python 3.12** with pip, venv, wheel
- **Node.js 22 LTS** with npm, pnpm, bun
- **AI Agent CLIs**: Claude Code, Codex, Gemini CLI
- **Session Recording**: Terminal and file I/O capture via virtio-vsock
- **Dev Tools**: git, jq, curl, and common utilities

## Architecture

Built using the NanoFuse layer-based architecture (ADR-001):

```
+-----------------------------------+
|         agent-tools               |  <- Claude Code, Codex, Gemini, git
+-----------------------------------+
|    python-runtime  | node-runtime |  <- Python 3.12, Node.js 22
+-----------------------------------+
|        recording-agent            |  <- Session capture (optional)
+-----------------------------------+
|           base-os                 |  <- Ubuntu 24.04 + systemd + SSH
+-----------------------------------+
```

## Layers

| Layer | Type | Size | Description |
|-------|------|------|-------------|
| `base-os` | base | ~200MB | Ubuntu 24.04 with systemd, SSH, networking |
| `python-runtime` | runtime | ~80MB | Python 3.12, pip, venv, wheel |
| `node-runtime` | runtime | ~120MB | Node.js 22 LTS, npm, pnpm, bun |
| `recording-agent` | feature | ~5MB | Session recording via vsock |
| `agent-tools` | application | ~50MB | AI CLIs, git, dev tools |

**Total estimated size**: ~450MB uncompressed

## Building

```bash
# Build with default settings (recording enabled)
nanofuse image build images/falcondev-agents/image.manifest.yaml

# Build without recording
INCLUDE_RECORDING=false nanofuse image build images/falcondev-agents/image.manifest.yaml

# Dry-run to preview build
nanofuse image build --dry-run images/falcondev-agents/image.manifest.yaml
```

## Environment Variables

The following environment variables are passed to the VM at runtime:

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub authentication token |
| `ANTHROPIC_API_KEY` | Anthropic API key for Claude |
| `OPENAI_API_KEY` | OpenAI API key for Codex |
| `GOOGLE_API_KEY` | Google API key for Gemini |

## Configuration

Layer configuration can be customized in the manifest:

```yaml
layers:
  - name: "recording-agent"
    config:
      vsock_port: 52           # virtio-vsock port
      buffer_size_mb: 16       # recording buffer size
      capture_modes:
        - "terminal"
        - "file_io"
```

## Usage

### Start a VM with this image

```bash
# Start a new VM
nanofuse vm create my-agent --image falcondev-agents

# Run Claude Code inside the VM
nanofuse vm exec my-agent -- claude

# Run with environment variables
nanofuse vm create my-agent \
  --image falcondev-agents \
  --env ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  --env GITHUB_TOKEN=$GITHUB_TOKEN
```

### DevContainer Usage

This image can be used as a dev container base:

```json
{
  "name": "FalconDev Agents Dev Environment",
  "image": "ghcr.io/daax-dev/nanofuse/falcondev-agents:latest"
}
```

## Session Recording

When recording is enabled, the recording agent:

1. Starts automatically via systemd
2. Connects to host via virtio-vsock port 52
3. Captures terminal I/O and file operations
4. Streams events to host recording receiver

Recording data is stored on the host at `/var/lib/nanofuse/recordings/{vm-id}/`.

## References

- [ADR-001: Layer-Based RootFS Architecture](../../docs/adr/adr-001-layer-based-rootfs-architecture.md)
- [ADR-002: Recording Integration Architecture](../../docs/adr/adr-002-recording-integration-architecture.md)
- [FalconDev DevContainer Spec](../../../jp/falcondev/.flowspec/templates/devcontainer/devcontainer.json)
