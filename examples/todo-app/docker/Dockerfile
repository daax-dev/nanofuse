# Multi-stage Dockerfile for NanoFuse Todo App Example
# This creates a complete VM image with backend + frontend + systemd

# Stage 1: Build backend
FROM golang:1.23-bookworm AS backend-builder

WORKDIR /build

# Copy backend source
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o todo-server ./cmd/server

# Stage 2: Build frontend
FROM node:20-bookworm AS frontend-builder

WORKDIR /build

# We'll add this when we create the frontend
# COPY frontend/package*.json ./
# RUN npm ci
# COPY frontend/ ./
# RUN npm run build

# For now, create a placeholder
RUN mkdir -p dist && echo '<html><body><h1>Todo App</h1><p>Backend running. Frontend coming soon!</p></body></html>' > dist/index.html

# Stage 3: Final runtime image with systemd
FROM ubuntu:24.04@sha256:80dd3c3b9c6cecb9f1667e9290b3bc61b78c2678c02cbdae5f0fea92cc6734ab

# Set noninteractive frontend
ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies (no nginx - go-chi serves static files directly)
RUN apt-get update && \
    apt-get install -y \
        systemd \
        systemd-sysv \
        openssh-server \
        ca-certificates \
        curl \
        iproute2 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy backend binary
COPY --from=backend-builder /build/todo-server /usr/local/bin/todo-server
RUN chmod +x /usr/local/bin/todo-server

# Copy frontend static files
COPY --from=frontend-builder /build/dist /var/www/html

# Create data directory for DuckDB
RUN mkdir -p /data && chmod 755 /data

# Create systemd service for backend (serves HTTP on port 80, static files from /var/www/html)
RUN cat > /etc/systemd/system/todo-backend.service <<'EOF'
[Unit]
Description=Todo App Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/usr/local/bin
ExecStart=/usr/local/bin/todo-server -db-path /data/todos.db -http-port 80 -grpc-port 9090 -static-dir /var/www/html
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Configure SSH for root login with key only (no password)
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Create SSH key injection script (reads base64-encoded key from kernel cmdline)
RUN cat > /usr/local/bin/install-ssh-key.sh <<'EOF'
#!/bin/bash
set -e
SSHKEY_B64=$(cat /proc/cmdline | grep -oP 'sshkey=\K[^ ]+' || true)
if [ -z "$SSHKEY_B64" ]; then
    echo "install-ssh-key: No SSH key provided via kernel cmdline"
    exit 0
fi
SSH_KEY=$(echo "$SSHKEY_B64" | base64 -d 2>/dev/null)
if [ -z "$SSH_KEY" ]; then
    echo "install-ssh-key: Failed to decode SSH key"
    exit 1
fi
if ! echo "$SSH_KEY" | grep -qE '^(ssh-|ecdsa-)'; then
    echo "install-ssh-key: Invalid SSH key format"
    exit 1
fi
mkdir -p /root/.ssh
chmod 700 /root/.ssh
echo "$SSH_KEY" >> /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys
echo "install-ssh-key: SSH key installed successfully"
EOF

RUN chmod +x /usr/local/bin/install-ssh-key.sh

# Create systemd service for SSH key injection (runs before sshd starts)
RUN cat > /etc/systemd/system/install-ssh-key.service <<'EOF'
[Unit]
Description=Install SSH key from kernel cmdline
Before=ssh.service sshd.service
After=local-fs.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/install-ssh-key.sh
RemainAfterExit=yes
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable services (no nginx - backend serves static files)
RUN systemctl enable todo-backend.service ssh.service install-ssh-key.service

# Health check script
RUN cat > /usr/local/bin/healthcheck.sh <<'EOF'
#!/bin/bash
# Check if backend is responding (now serves on port 80)
if ! curl -sf http://localhost:80/health > /dev/null; then
    echo "Backend health check failed"
    exit 1
fi

echo "Health check passed"
exit 0
EOF

RUN chmod +x /usr/local/bin/healthcheck.sh

# Expose ports (80 for HTTP/static, 9090 for gRPC)
EXPOSE 80 9090

# Labels for metadata (SLSA compliance)
LABEL org.opencontainers.image.title="NanoFuse Todo App Example"
LABEL org.opencontainers.image.description="Production-quality reference implementation demonstrating long-running VM stability"
LABEL org.opencontainers.image.vendor="NanoFuse Project"
LABEL org.opencontainers.image.source="https://github.com/daax-dev/nanofuse"
LABEL org.opencontainers.image.documentation="https://github.com/daax-dev/nanofuse/blob/main/examples/todo-app/README.md"
LABEL org.opencontainers.image.version="1.0.0"

# Constitutional compliance labels
LABEL netcon.constitutional.principle.iii="SLSA v1.1 Compliance"
LABEL netcon.constitutional.principle.v="Automation-Driven Repeatability"
LABEL netcon.constitutional.principle.vi="Security by Architectural Design"

# CMD expects systemd
CMD ["/lib/systemd/systemd"]
