#!/bin/bash
set -e

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"

# Create the fixed nginx config
cat > /tmp/nginx-fixed.conf << 'EOF'
# Nginx configuration for Todo App

server {
    listen 80 default_server;

    server_name _;

    root /var/www/html;
    index index.html;

    # Frontend static files
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to backend
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Proxy health endpoint
    location /health {
        proxy_pass http://localhost:8080/health;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
    }

    # Proxy metrics endpoint
    location /metrics {
        proxy_pass http://localhost:8080/metrics;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;
}
EOF

echo "Patching nginx config in rootfs..."
sudo guestfish -a "${ROOTFS}" -m /dev/sda <<GUESTFISH
upload /tmp/nginx-fixed.conf /etc/nginx/sites-available/default
GUESTFISH

rm /tmp/nginx-fixed.conf
echo "Nginx config patched successfully!"
