#!/bin/bash
# Check VM console log to see why services are failing

VM_ID=$(nanofuse vm list --json | jq -r '.vms[0].id')
echo "VM ID: $VM_ID"
echo "=================================="
echo "Console log (last 50 lines):"
echo "=================================="
tail -50 /var/lib/nanofuse/vms/$VM_ID/console.log
