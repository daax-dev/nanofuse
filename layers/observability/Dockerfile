# Observability Layer Dockerfile
# Provides Prometheus node exporter and structured logging
#
# Build: docker build -t nanofuse-layer-observability .
# Export: docker export $(docker create nanofuse-layer-observability) | tar -C rootfs -xf -

FROM ubuntu:24.04 AS builder

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Download node_exporter with checksum verification
# Version 1.8.2 released May 2024, verified from official GitHub releases
# Checksums from: https://github.com/prometheus/node_exporter/releases/tag/v1.8.2
# Update this when newer versions are available (check for security advisories)
ARG NODE_EXPORTER_VERSION=1.8.2
ARG NODE_EXPORTER_SHA256=6809dd0b3ec45fd6e992c19071d6b5253aed3ead7bf0686885a51d85c6643c66
RUN curl -fsSL -o /tmp/node_exporter.tar.gz \
    "https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz" \
    && echo "${NODE_EXPORTER_SHA256}  /tmp/node_exporter.tar.gz" | sha256sum -c - \
    && tar -xzf /tmp/node_exporter.tar.gz --strip-components=1 -C /usr/local/bin \
       node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64/node_exporter \
    && rm /tmp/node_exporter.tar.gz

# Runtime image
FROM ubuntu:24.04

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    logrotate \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy node_exporter from builder
COPY --from=builder /usr/local/bin/node_exporter /usr/local/bin/node_exporter

# Create directories
RUN mkdir -p /var/log/nanofuse /etc/nanofuse

# Create default logging configuration
# Note: This is overwritten by the post-install hook with values from CONFIG_ environment
# variables. These defaults ensure the layer works standalone for development/testing.
RUN cat > /etc/nanofuse/logging.yaml << 'EOF'
# NanoFuse Logging Configuration

logging:
  level: info
  format: json
  outputs:
    - type: file
      path: /var/log/nanofuse/gateway.log
    - type: stdout

rotation:
  max_size_mb: 100
  max_age_days: 7
  max_backups: 5
  compress: true
EOF

# Create logrotate configuration
RUN cat > /etc/logrotate.d/nanofuse << 'EOF'
/var/log/nanofuse/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
    sharedscripts
    # No postrotate required - NanoFuse gateway logs directly to files/stdout
}
EOF
