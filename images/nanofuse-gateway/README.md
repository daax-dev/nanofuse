# NanoFuse Gateway Image

A network gateway microVM for ingress/egress control, load balancing, and API routing for microVM clusters.

## Overview

This image provides a complete network gateway running in a Firecracker microVM with:

- **Advanced Networking**: iptables, nftables, iproute2 for traffic control
- **Gateway Service**: Go-based reverse proxy and load balancer
- **TLS Support**: CA certificates and self-signed certificate management (ACME not yet implemented)
- **Observability**: Prometheus node exporter and structured logging

## Architecture

Built using the NanoFuse layer-based architecture (ADR-001):

```
+-----------------------------------+
|        gateway-runtime            |  <- Go gateway binary
+-----------------------------------+
|   networking-tools | tls-certs    |  <- iptables, nftables, CA certs
+-----------------------------------+
|         observability             |  <- Prometheus, logging
+-----------------------------------+
|            base-os                |  <- Ubuntu 24.04 + systemd
+-----------------------------------+
```

## Layers

| Layer | Type | Size | Description |
|-------|------|------|-------------|
| `base-os` | base | ~185MB | Ubuntu 24.04 with systemd, SSH, networking |
| `networking-tools` | feature | ~25MB | iptables, nftables, iproute2, traffic control |
| `gateway-runtime` | application | ~30MB | Go gateway binary with reverse proxy |
| `tls-certificates` | feature | ~10MB | CA certificates, cert-manager tool |
| `observability` | feature | ~15MB | Prometheus node exporter, log rotation |

**Total estimated size**: ~265MB uncompressed

## Building

```bash
# Build with default settings
nanofuse image build images/nanofuse-gateway/image.manifest.yaml

# Dry-run to preview build
nanofuse image build --dry-run images/nanofuse-gateway/image.manifest.yaml
```

## Ports

| Port | Service | Description |
|------|---------|-------------|
| 8080 | Gateway | Main traffic ingress port |
| 8081 | Health | Health check endpoint (/health, /ready) |
| 9090 | Metrics | Gateway Prometheus metrics |
| 9100 | Node Exporter | System Prometheus metrics |

## Configuration

Layer configuration can be customized in the manifest:

```yaml
layers:
  - name: "gateway-runtime"
    config:
      listen_port: 8080          # Main listener port
      metrics_port: 9090         # Gateway metrics port
      health_port: 8081          # Health check port
      upstream_timeout: 30s      # Upstream connection timeout
      max_connections: 10000     # Maximum concurrent connections

  - name: "observability"
    config:
      node_exporter_port: 9100   # Node exporter port
      log_format: json           # Log format: json, text
      log_level: info            # Log level: debug, info, warn, error
```

## Usage

### Start a Gateway VM

```bash
# Create a gateway VM
nanofuse vm create gateway-01 --image nanofuse-gateway

# Start with custom ports
nanofuse vm create gateway-01 \
  --image nanofuse-gateway \
  --port-forward 8080:8080 \
  --port-forward 9090:9090
```

### Health Checks

```bash
# Check gateway health
curl http://gateway:8081/health

# Check readiness
curl http://gateway:8081/ready
```

### Metrics Endpoints

```bash
# Gateway metrics
curl http://gateway:9090/metrics

# System metrics (node exporter)
curl http://gateway:9100/metrics
```

## Firewall Configuration

The gateway starts with a permissive nftables ruleset (`default_policy: accept`) since application-level
security is handled by the gateway itself. To customize firewall rules:

```bash
# SSH into the gateway VM
nanofuse vm exec gateway-01 -- bash

# Edit nftables rules
vi /etc/nftables.conf

# Restart service to apply rules (reload may not be supported)
systemctl restart nftables
# Or for immediate testing without systemctl:
nft -f /etc/nftables.conf
```

Default nftables configuration allows:
- Established/related connections
- Loopback traffic
- All forwarding (for gateway routing)
- Service ports: 8080 (gateway), 8081 (health), 9090 (metrics), 9100 (node exporter)

**Note**: These ports are accessible because this image uses `default_policy: accept` in the
networking-tools layer configuration. Images without this setting will default to `drop` policy,
requiring explicit firewall rules to allow traffic.

## TLS Configuration

The gateway supports TLS for secure connections. TLS is disabled by default and requires the
`tls-certificates` layer (included in this image).

### Enable TLS

1. Generate a self-signed certificate:

```bash
nanofuse vm exec gateway-01 -- cert-manager generate-self-signed gateway.example.com 365
```

2. Enable TLS in the gateway configuration by updating the existing `tls` section.

**Option A: Manual edit (simplest)**

Edit `/etc/nanofuse/gateway.yaml` so the `tls` section looks like:

```yaml
tls:
  enabled: true
  cert_file: /etc/nanofuse/certs/server.crt
  key_file: /etc/nanofuse/certs/server.key
```

**Option B: Using sed (no additional tools required)**

```bash
# Matches YAML key-value format with flexible whitespace.
# Pattern breakdown:
#   ^([[:space:]]*)      - Captures leading indentation
#   enabled:             - Matches the key literally
#   [[:space:]]*         - Allows any whitespace after colon
#   false                - Matches boolean false
#   \1enabled: true      - Rewrites value while preserving indentation
nanofuse vm exec gateway-01 -- bash -c "sed -i -E 's/^([[:space:]]*)enabled:[[:space:]]*false/\1enabled: true/' /etc/nanofuse/gateway.yaml"
```

**Option C: Using yq (if installed)**

```bash
# Requires yq: https://github.com/mikefarah/yq - install with: snap install yq
nanofuse vm exec gateway-01 -- bash -c "yq -i '.tls.enabled = true' /etc/nanofuse/gateway.yaml"
```

3. Restart the gateway:

```bash
nanofuse vm exec gateway-01 -- systemctl restart nanofuse-gateway
```

### Verify Certificate

```bash
nanofuse vm exec gateway-01 -- cert-manager verify
nanofuse vm exec gateway-01 -- cert-manager info
```

## Logging

Logs are written to `/var/log/nanofuse/` in JSON format by default:

```bash
# View gateway logs
nanofuse vm exec gateway-01 -- journalctl -u nanofuse-gateway -f

# View node exporter logs
nanofuse vm exec gateway-01 -- journalctl -u node-exporter -f
```

Log rotation is configured via logrotate with 7-day retention.

## Services

The following systemd services are enabled:

| Service | Description |
|---------|-------------|
| `nanofuse-gateway.service` | Main gateway service |
| `node-exporter.service` | Prometheus node exporter |
| `nftables.service` | Firewall rules |
| `cert-renewal.timer` | Certificate renewal timer |

## References

- [ADR-001: Layer-Based RootFS Architecture](../../docs/adr/adr-001-layer-based-rootfs-architecture.md)
- [Networking Tools Layer](../../layers/networking-tools/layer.yaml)
- [Gateway Runtime Layer](../../layers/gateway-runtime/layer.yaml)
- [TLS Certificates Layer](../../layers/tls-certificates/layer.yaml)
- [Observability Layer](../../layers/observability/layer.yaml)
