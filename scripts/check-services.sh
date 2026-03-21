#!/bin/bash
VM_ID=$(nanofuse --api-url http://localhost:8080 vm list 2>&1 | grep "my-todo-app" | awk '{print $1}')
echo "VM ID: $VM_ID"
echo ""
echo "=== SERVICES STATUS FROM CONSOLE ==="
tail -150 /var/lib/nanofuse/vms/${VM_ID}*/console.log 2>/dev/null | tail -50
