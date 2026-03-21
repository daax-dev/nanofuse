#!/bin/bash
VM_ID=$(nanofuse --api-url http://localhost:8080 vm list 2>&1 | grep "my-todo-app" | awk '{print $1}')
echo "VM ID: $VM_ID"
echo ""
echo "=== NGINX ERRORS FROM CONSOLE LOG ==="
cat /var/lib/nanofuse/vms/${VM_ID}*/console.log 2>/dev/null | grep -i nginx
