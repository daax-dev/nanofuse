# Agent Tools Layer - AI agent CLIs and development utilities
# This Dockerfile creates a layer with common development tools and AI agent CLIs.
#
# Provides:
# - git, curl, wget, jq (common dev tools)
# - Claude Code CLI (via pnpm)
# - Codex CLI (placeholder - requires API key)
# - Gemini CLI (placeholder - requires API key)
#
# Build: docker build -t nanofuse-layer-agent-tools:latest .
# Extract: ./scripts/build-layer.sh agent-tools
#
# Target size: ~100MB

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse Agent Tools Layer"
LABEL org.opencontainers.image.description="AI agent CLIs and development utilities"
LABEL org.opencontainers.image.version="1.0.0"
LABEL com.nanofuse.layer.type="application"
LABEL com.nanofuse.layer.name="agent-tools"

ENV DEBIAN_FRONTEND=noninteractive

# Install development tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    wget \
    jq \
    vim-tiny \
    less \
    openssh-client \
    gnupg \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install Node.js 22 LTS (needed for CLI tools)
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install pnpm globally
RUN npm install -g pnpm@latest

# Install Claude Code CLI
# Note: This installs the CLI globally. Users will need ANTHROPIC_API_KEY at runtime.
RUN pnpm install -g @anthropic-ai/claude-code || echo "Claude Code install skipped (may require auth)"

# Create wrapper scripts for other AI CLIs (placeholders)
RUN cat > /usr/local/bin/codex << 'SCRIPT'
#!/bin/bash
# OpenAI Codex CLI wrapper
# Install: npm install -g @openai/codex
echo "[codex] OpenAI Codex CLI"
echo "To install: npm install -g @openai/codex"
echo "Requires: OPENAI_API_KEY environment variable"
if command -v codex &>/dev/null; then
    exec codex "$@"
else
    echo "Codex not installed. Run: npm install -g @openai/codex"
    exit 1
fi
SCRIPT
RUN chmod +x /usr/local/bin/codex

RUN cat > /usr/local/bin/gemini << 'SCRIPT'
#!/bin/bash
# Google Gemini CLI wrapper
# Install: pip install google-generativeai
echo "[gemini] Google Gemini CLI"
echo "To install: pip install google-generativeai"
echo "Requires: GOOGLE_API_KEY environment variable"
if command -v gemini &>/dev/null; then
    exec gemini "$@"
else
    echo "Gemini CLI not installed. Run: pip install google-generativeai"
    exit 1
fi
SCRIPT
RUN chmod +x /usr/local/bin/gemini

# Configure git defaults
RUN git config --system init.defaultBranch main \
    && git config --system core.editor vim \
    && git config --system credential.helper store

# Create directories for agent workspaces
RUN mkdir -p /workspace \
    && mkdir -p /home/agent \
    && chmod 755 /workspace /home/agent

# Create layer marker
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: agent-tools" > /etc/nanofuse/layers/agent-tools.yaml \
    && echo "version: 1.0.0" >> /etc/nanofuse/layers/agent-tools.yaml \
    && echo "type: application" >> /etc/nanofuse/layers/agent-tools.yaml

# Verify key tools are installed
RUN git --version && curl --version | head -1 && jq --version

# Clean up
RUN rm -rf /var/log/* /tmp/* \
    && rm -rf /usr/share/doc/* /usr/share/man/* \
    && npm cache clean --force

CMD ["/bin/bash"]
