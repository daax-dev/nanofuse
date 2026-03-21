#!/bin/bash
# Show console log for VM
VM_ID=$(nanofuse --api-url http://localhost:8080 vm list 2>&1 | grep "my-todo-app" | awk '{print $1}')
if [[ -z "$VM_ID" ]]; then
    echo "VM not found"
    exit 1
fi
echo "VM ID: $VM_ID"
echo "Console log:"
echo "=========================================="
cat /var/lib/nanofuse/vms/${VM_ID}*/console.log 2>/dev/null || echo "Cannot read (need sudo)"
