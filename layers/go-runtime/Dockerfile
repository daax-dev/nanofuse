# Go Runtime Layer - Minimal Go runtime environment
# This Dockerfile creates a minimal environment for running Go binaries.
# Go binaries are typically statically linked, so we mainly need:
# - CA certificates for HTTPS
# - CGO support libraries (if CGO is used)
# - Timezone data
#
# Build: docker build -t nanofuse-layer-go:1.22 .
# Extract: ./build-layer.sh go-runtime
#
# Target size: ~20MB (compressed)

FROM ubuntu:24.04

LABEL org.opencontainers.image.title="NanoFuse Go Runtime Layer"
LABEL org.opencontainers.image.description="Go runtime libraries for NanoFuse microVMs"
LABEL org.opencontainers.image.version="1.22"
LABEL com.nanofuse.layer.type="runtime"
LABEL com.nanofuse.layer.name="go-runtime"

ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.22

# Install minimal dependencies for Go binaries
# - ca-certificates: for HTTPS
# - libc6: for CGO-enabled binaries
# - tzdata: for timezone support
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libc6 \
    tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create standard directories for Go binaries
RUN mkdir -p /usr/local/bin \
    && mkdir -p /var/log \
    && mkdir -p /tmp

# Create version marker for layer identification
RUN mkdir -p /etc/nanofuse/layers \
    && echo "name: go-runtime" > /etc/nanofuse/layers/go-runtime.yaml \
    && echo "version: ${GO_VERSION}" >> /etc/nanofuse/layers/go-runtime.yaml \
    && echo "type: runtime" >> /etc/nanofuse/layers/go-runtime.yaml

CMD ["/bin/bash"]
