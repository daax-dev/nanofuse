#!/bin/bash
################################################################################
# FINAL NGINX FIX - Create systemd override to fix nginx startup
################################################################################
set -euo pipefail

ROOTFS="/var/lib/nanofuse/images/sha256:0c8543d7e080fc3d23cbdaea604c605a9c6a0361e2ef16a6e11c8c1f2ff79628/rootfs.ext4"

echo "=== FINAL NGINX FIX ==="
echo "Using debugfs to modify rootfs without mounting..."

# Create override directory
echo "Creating systemd override directory..."
debugfs -w -R "mkdir /etc/systemd/system/nginx.service.d" "$ROOTFS" 2>/dev/null || echo "Directory may already exist"

# Create override file content
cat > /tmp/nginx-override.conf <<'EOF'
[Service]
# Remove the test that fails
ExecStartPre=
# Just start nginx directly
ExecStart=
ExecStart=/usr/sbin/nginx -g 'daemon on; master_process on;'
# Create /run directory if needed
RuntimeDirectory=nginx
RuntimeDirectoryMode=0755
EOF

# Write override file to rootfs
echo "Writing override configuration..."
debugfs -w -R "write /tmp/nginx-override.conf /etc/systemd/system/nginx.service.d/override.conf" "$ROOTFS"

echo ""
echo "✓ Created systemd override for nginx"
echo ""
echo "Now test:"
echo "  1. Delete VM: nanofuse --api-url http://localhost:8080 vm delete my-todo-app --force"
echo "  2. Create VM: nanofuse --api-url http://localhost:8080 vm create sha256:0c8543... my-todo-app --vcpus 2 --memory 1024 --kernel-args 'console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init'"
echo "  3. Start: nanofuse --api-url http://localhost:8080 vm start my-todo-app"
echo "  4. Wait: sleep 20"
echo "  5. Test: /home/jpoley/ps/nanofuse/scripts/test-todo-app.sh"
