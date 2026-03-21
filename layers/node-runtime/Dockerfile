# Node.js Runtime Layer - Node.js 22 LTS installation
# This Dockerfile creates a minimal Node.js environment that can be
# extracted as a layer for NanoFuse microVM images.
#
# Build: docker build -t nanofuse-layer-node:22 .
# Extract: ./build-layer.sh node-runtime
#
# Target size: ~100MB (compressed)

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse Node.js Runtime Layer"
LABEL org.opencontainers.image.description="Node.js 22 LTS runtime for NanoFuse microVMs"
LABEL org.opencontainers.image.version="22"
LABEL com.nanofuse.layer.type="runtime"
LABEL com.nanofuse.layer.name="node-runtime"

ENV DEBIAN_FRONTEND=noninteractive
ENV NODE_VERSION=22

# Install dependencies for Node.js and package managers
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    gnupg \
    unzip \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install Node.js 22 LTS from NodeSource
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install pnpm globally
RUN npm install -g pnpm@latest

# Install bun (BUN_INSTALL sets destination directory)
ENV BUN_INSTALL=/usr/local
RUN curl -fsSL https://bun.sh/install | bash

# Create version marker for layer identification
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: node-runtime" > /etc/nanofuse/layers/node-runtime.yaml \
    && echo "version: ${NODE_VERSION}" >> /etc/nanofuse/layers/node-runtime.yaml \
    && echo "type: runtime" >> /etc/nanofuse/layers/node-runtime.yaml

# Verify installations
RUN node --version && npm --version && pnpm --version && bun --version

CMD ["/bin/bash"]
