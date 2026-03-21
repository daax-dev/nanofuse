#!/bin/bash
# Debug script for nanofused daemon issues

echo "=== NanoFuse Daemon Debug ==="
echo

echo "1. Binary check:"
echo "   Installed: $(md5sum /usr/local/bin/nanofused 2>/dev/null | cut -d' ' -f1 || echo 'NOT FOUND')"
echo "   Local:     $(md5sum ./bin/nanofused 2>/dev/null | cut -d' ' -f1 || echo 'NOT FOUND')"
[ -f /usr/local/bin/nanofused ] && [ -f ./bin/nanofused ] && \
  [ "$(md5sum /usr/local/bin/nanofused | cut -d' ' -f1)" = "$(md5sum ./bin/nanofused | cut -d' ' -f1)" ] && \
  echo "   Status: MATCH" || echo "   Status: MISMATCH - run: sudo cp bin/nanofused /usr/local/bin/"
echo

echo "2. Daemon process:"
pgrep -a nanofused || echo "   NOT RUNNING"
echo

echo "3. Config file:"
if [ -f /etc/nanofuse/nanofused.yaml ]; then
  echo "   Path: /etc/nanofuse/nanofused.yaml"
  grep -q "file_path" /etc/nanofuse/nanofused.yaml && \
    echo "   Logging: $(grep file_path /etc/nanofuse/nanofused.yaml | awk '{print $2}')" || \
    echo "   Logging: NOT CONFIGURED"
else
  echo "   NOT FOUND"
fi
echo

echo "4. Log directory:"
ls -la /var/log/nanofuse/ 2>/dev/null || echo "   NOT FOUND - run: sudo mkdir -p /var/log/nanofuse"
echo

echo "5. Socket:"
ls -la /run/nanofused.sock 2>/dev/null || ls -la /var/run/nanofused.sock 2>/dev/null || echo "   NOT FOUND"
echo

echo "6. Database:"
if [ -f /var/lib/nanofuse/nanofuse.db ]; then
  echo "   Path: /var/lib/nanofuse/nanofuse.db"
  echo "   VMs: $(sqlite3 /var/lib/nanofuse/nanofuse.db 'SELECT COUNT(*) FROM vms;' 2>/dev/null || echo 'ERROR')"
  echo "   Images: $(sqlite3 /var/lib/nanofuse/nanofuse.db 'SELECT COUNT(*) FROM images;' 2>/dev/null || echo 'ERROR')"
  echo "   Pending jobs: $(sqlite3 /var/lib/nanofuse/nanofuse.db "SELECT COUNT(*) FROM image_pull_jobs WHERE state != 'completed' AND state != 'failed';" 2>/dev/null || echo 'ERROR')"
else
  echo "   NOT FOUND"
fi
echo

echo "7. Recent errors (last 10):"
grep -i "error\|fail" /var/log/nanofuse/nanofused.log 2>/dev/null | tail -10 || echo "   No log file or no errors"
echo

echo "8. Systemd status:"
systemctl is-active nanofused 2>/dev/null || echo "   Not managed by systemd"
echo

echo "=== End Debug ==="
