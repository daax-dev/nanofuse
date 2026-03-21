#!/bin/bash
# Check what's actually stored in the database

echo "=== VMs in database ==="
sqlite3 /tmp/nanofuse/nanofuse.db "SELECT id, name, state FROM vms;"

echo ""
echo "=== Network config JSON for first VM ==="
sqlite3 /tmp/nanofuse/nanofuse.db "SELECT config_json FROM vms LIMIT 1;" | jq '.network'
