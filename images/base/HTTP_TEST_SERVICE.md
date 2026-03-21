# HTTP Test Server Service

## Overview

The NanoFuse base image includes a built-in HTTP test server (`http-test-server.service`) that enables testing port forwarding and network connectivity features.

## Service Details

**File:** `units/http-test-server.service`

**Purpose:** Provides a simple HTTP endpoint on port 8080 for testing:
- Port forwarding from host to microVM
- Network connectivity between host and guest
- Guest accessibility via HTTP

## Configuration

### Service Definition

The HTTP test server is implemented as a simple Python HTTP server:

```ini
[Unit]
Description=NanoFuse HTTP Test Server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Restart=always
RestartSec=5
ExecStartPre=/bin/bash -c 'mkdir -p /var/www && echo "{\"message\":\"Hello from NanoFuse VM\",\"service\":\"http-test-server\",\"version\":\"1.0\"}" > /var/www/index.html'
ExecStart=/bin/bash -c 'cd /var/www && python3 -m http.server 8080'
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### Service Features

- **Type:** Simple (runs continuously)
- **Auto-restart:** Yes (restarts every 5 seconds if it crashes)
- **Port:** 8080
- **Response:** JSON object with version and service name
- **Dependencies:** Requires network to be online

## Testing

### 1. Inside the microVM (direct test)

Once the microVM boots, test the HTTP server directly:

```bash
# Boot the microVM
nanofuse vm create test-http --image default --vcpus 2 --memory 512
nanofuse vm start test-http

# SSH into the VM
ssh root@<vm-ip>

# Test the HTTP server inside the VM
curl http://localhost:8080/
# Output: {"message":"Hello from NanoFuse VM","service":"http-test-server","version":"1.0"}
```

### 2. Host to microVM (port forwarding test)

Using NanoFuse port forwarding features:

```bash
# Create VM with port forwarding
nanofuse vm create test-http --image default --vcpus 2 --memory 512
nanofuse port-forward add test-http --host 8080 --guest 8080

# Start the VM
nanofuse vm start test-http

# Test from host
curl http://localhost:8080/
# Output: {"message":"Hello from NanoFuse VM","service":"http-test-server","version":"1.0"}
```

### 3. Kernel and Image Validation

The HTTP test server is enabled in the base image but its successful startup depends on:

- Kernel booting correctly
- systemd initializing the service manager
- Python 3 being available (included in base image)
- Network becoming ready

If the microVM fails to boot, the HTTP server will never start.

## Integration in GitHub Actions

The HTTP test server is automatically included in the base image when:

1. The Dockerfile copies `units/http-test-server.service`
2. systemctl enables the service for multi-user.target
3. The Docker image is exported and converted to ext4 rootfs

No additional steps needed - it's baked into the image.

## Troubleshooting

### Service not starting

```bash
# Check service status
systemctl status http-test-server

# Check logs
journalctl -u http-test-server -n 20

# Manually test (if service is stuck)
cd /var/www
python3 -m http.server 8080

# Verify index.html exists
cat /var/www/index.html
```

### Cannot connect on port 8080

1. Verify service is running: `systemctl is-active http-test-server`
2. Verify port is listening: `ss -tlnp | grep 8080`
3. Verify network is ready: `ip route`
4. Check firewall rules: `iptables -L` (should be empty in Firecracker VM)

## Related Files

- `Dockerfile` - Lines 121-124: Service installation
- `units/http-test-server.service` - Service definition
- `TEST_BOOT_VERBOSE.sh` - Boot test script (can be extended to test HTTP)
- `.github/workflows/ci.yaml` - CI pipeline (no explicit HTTP test yet)

## Future Enhancements

Possible improvements:

1. Add HTTP connectivity test to `TEST_BOOT_VERBOSE.sh`
2. Add automated HTTP port forwarding test to GitHub Actions
3. Support custom port configuration via environment variables
4. Add TLS/HTTPS support with self-signed certificates
5. Extend response with system metrics (CPU, memory, etc.)

## Security Considerations

- The HTTP server runs as root (via systemd)
- No authentication or TLS by default
- Should only be used for testing in isolated environments
- Production deployments should replace with application-specific services
- Consider disabling in production images if not needed
