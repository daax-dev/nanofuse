# Debug Tools Layer - Debugging and diagnostic tools
# This Dockerfile creates a layer with debugging, tracing, and network tools.
#
# Provides:
# - strace (syscall tracing)
# - ltrace (library call tracing)
# - gdb (GNU debugger)
# - tcpdump (network packet capture)
# - htop (interactive process viewer)
# - curl, wget (network tools)
#
# Build: docker build -t nanofuse-layer-debug-tools:latest .
# Extract: ./scripts/build-layer.sh debug-tools
#
# Approximate compressed layer size: ~62MB (including GDB; uncompressed filesystem will be larger)

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse Debug Tools Layer"
LABEL org.opencontainers.image.description="Debugging and diagnostic tools for microVM troubleshooting"
LABEL org.opencontainers.image.version="1.0.0"
LABEL com.nanofuse.layer.type="feature"
LABEL com.nanofuse.layer.name="debug-tools"

ENV DEBIAN_FRONTEND=noninteractive

# Install debugging and diagnostic tools
# Using --no-install-recommends to minimize layer size
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Tracing tools
    strace \
    ltrace \
    # Debugger
    gdb \
    # Network tools
    tcpdump \
    curl \
    wget \
    # Process monitoring
    htop \
    # Minimal dependencies for tools to work
    libncursesw6 \
    libcap2-bin \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create layer marker
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: debug-tools" > /etc/nanofuse/layers/debug-tools.yaml \
    && echo "version: 1.0.0" >> /etc/nanofuse/layers/debug-tools.yaml \
    && echo "type: feature" >> /etc/nanofuse/layers/debug-tools.yaml

# Verify all tools are installed and executable
RUN strace --version | head -1 \
    && ltrace --version | head -1 \
    && gdb --version | head -1 \
    && tcpdump --version 2>&1 | head -1 \
    && htop --version | head -1 \
    && curl --version | head -1 \
    && wget --version | head -1

# Clean up to reduce layer size
RUN rm -rf /var/log/* /tmp/* \
    && rm -rf /usr/share/doc/* \
    && rm -rf /usr/share/man/* \
    && rm -rf /usr/share/info/* \
    && find /usr/share/locale -mindepth 1 -maxdepth 1 ! \( -name "en" -o -name "en_*" \) -exec rm -rf {} + \
    && rm -rf /var/cache/*

CMD ["/bin/bash"]
