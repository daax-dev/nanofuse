#!/bin/bash
# Test the image lookup fix

echo "=== Testing Image Lookup Fix ==="
echo ""

# Kill old daemon
echo "1. Restarting daemon with fixed binary..."
sudo pkill -9 nanofused
sleep 2

# Start new daemon
sudo /home/jpoley/ps/nanofuse/bin/nanofused &
sleep 3

# Test API
echo ""
echo "2. Testing API..."
curl -s http://localhost:8080/health | jq '.status'

# Check image
echo ""
echo "3. Checking image exists..."
./bin/nanofuse --api-url http://localhost:8080 image list

# Try to create VM (THIS SHOULD NOW WORK!)
echo ""
echo "4. Creating VM with tag reference..."
./bin/nanofuse --api-url http://localhost:8080 vm create ghcr.io/daax-dev/nanofuse/base:latest test-vm --vcpus 2 --memory 512

echo ""
echo "5. Listing VMs..."
./bin/nanofuse --api-url http://localhost:8080 vm list

echo ""
echo "6. Starting VM..."
./bin/nanofuse --api-url http://localhost:8080 vm start test-vm

echo ""
echo "7. Waiting 15 seconds for VM to boot..."
sleep 15

echo ""
echo "8. Getting VM IP address..."
VM_IP=$(./bin/nanofuse --api-url http://localhost:8080 vm status test-vm --json | jq -r '.network.ip')
echo "   VM IP: $VM_IP"

echo ""
echo "9. Checking VM status..."
./bin/nanofuse --api-url http://localhost:8080 vm status test-vm

echo ""
echo "10. Testing network connectivity to VM..."
if ping -c 3 $VM_IP; then
    echo "✓ VM is reachable!"
else
    echo "✗ Cannot ping VM"
fi

echo ""
echo "11. Checking if any services are running in VM..."
echo "    (Note: Base image might not have HTTP server running)"
echo "    But you can SSH if configured, or curl will show connection attempt"
curl -v --connect-timeout 2 http://$VM_IP 2>&1 | head -20

echo ""
echo "=== VM Lifecycle Complete! ==="
echo ""
echo "The VM is running at: $VM_IP"
echo "You can:"
echo "  - Ping it: ping $VM_IP"
echo "  - Check logs: ./bin/nanofuse --api-url http://localhost:8080 vm logs test-vm --tail 30"
echo "  - Stop it: ./bin/nanofuse --api-url http://localhost:8080 vm stop test-vm"
echo "  - Delete it: ./bin/nanofuse --api-url http://localhost:8080 vm delete test-vm --force"
