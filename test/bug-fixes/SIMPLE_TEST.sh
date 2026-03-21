#!/bin/bash
# Simplest possible VM test - no fancy stuff

echo "=== Simple VM Test ==="
echo ""

# Step 1: Kill everything and restart clean
echo "1. Restarting daemon..."
sudo pkill -9 nanofused 2>/dev/null || true
sleep 2
sudo rm -f /tmp/nanofused.log /tmp/nanofused.sock
sudo /home/jpoley/ps/nanofuse/bin/nanofused &
sleep 3

# Step 2: Check it's running
echo ""
echo "2. Testing API..."
if curl -s http://localhost:8080/health | grep -q healthy; then
    echo "✓ Daemon is running"
else
    echo "✗ Daemon failed"
    exit 1
fi

# Step 3: Check image
echo ""
echo "3. Checking images..."
./bin/nanofuse --api-url http://localhost:8080 image list

# Step 4: Create VM (skip delete, just try create)
echo ""
echo "4. Creating VM test-vm-simple..."
./bin/nanofuse --api-url http://localhost:8080 vm create default test-vm-simple --vcpus 2 --memory 512

# Step 5: List VMs
echo ""
echo "5. Listing VMs..."
./bin/nanofuse --api-url http://localhost:8080 vm list

# Step 6: Start VM
echo ""
echo "6. Starting VM..."
./bin/nanofuse --api-url http://localhost:8080 vm start test-vm-simple

# Step 7: Wait and check status
echo ""
echo "7. Waiting 10 seconds..."
sleep 10

echo ""
echo "8. VM Status:"
./bin/nanofuse --api-url http://localhost:8080 vm status test-vm-simple

# Step 8: Get logs
echo ""
echo "9. VM Logs (last 15 lines):"
./bin/nanofuse --api-url http://localhost:8080 vm logs test-vm-simple --tail 15

echo ""
echo "=== Test Complete ==="
echo ""
echo "VM is running. To cleanup manually:"
echo "  ./bin/nanofuse --api-url http://localhost:8080 vm stop test-vm-simple"
echo "  ./bin/nanofuse --api-url http://localhost:8080 vm delete test-vm-simple --force"
