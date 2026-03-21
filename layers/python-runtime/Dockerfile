# Python Runtime Layer - Minimal Python 3.12 installation
# This Dockerfile creates a minimal Python environment that can be
# extracted as a layer for NanoFuse microVM images.
#
# Build: docker build -t nanofuse-layer-python:3.12 .
# Extract: ./build-layer.sh python-runtime
#
# Target size: ~80MB (compressed)

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse Python Runtime Layer"
LABEL org.opencontainers.image.description="Python 3.12 runtime for NanoFuse microVMs"
LABEL org.opencontainers.image.version="3.12"
LABEL com.nanofuse.layer.type="runtime"
LABEL com.nanofuse.layer.name="python-runtime"

ENV DEBIAN_FRONTEND=noninteractive
ENV PYTHON_VERSION=3.12

# Install Python 3.12 and essential packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3.12 \
    python3.12-venv \
    python3-pip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create symlinks for python/pip
RUN update-alternatives --install /usr/bin/python python /usr/bin/python3.12 1 \
    && update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.12 1 \
    && ln -sf /usr/bin/pip3 /usr/local/bin/pip

# Upgrade pip and install common packages
RUN python3 -m pip install --no-cache-dir --break-system-packages \
    pip \
    wheel \
    setuptools \
    requests \
    pyyaml \
    httpx

# Create version marker for layer identification
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: python-runtime" > /etc/nanofuse/layers/python-runtime.yaml \
    && echo "version: ${PYTHON_VERSION}" >> /etc/nanofuse/layers/python-runtime.yaml \
    && echo "type: runtime" >> /etc/nanofuse/layers/python-runtime.yaml

# Verify installation
RUN python --version && pip --version

CMD ["/bin/bash"]
