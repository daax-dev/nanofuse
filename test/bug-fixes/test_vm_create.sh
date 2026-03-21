#!/bin/bash
# Now that image pull works, test creating a VM

echo "=== Testing VM Creation ==="
echo ""

# Check if image exists, pull if not
IMAGE_COUNT=$(curl -s http://localhost:8080/images | jq '.images | length')
if [ "$IMAGE_COUNT" -eq 0 ]; then
    echo "No image found, pulling default image..."
    ./bin/nanofuse --api-url http://localhost:8080 image pull --default &
    PULL_PID=$!
    echo "Pull started in background (PID: $PULL_PID)"
    echo "Waiting 30 seconds for pull to complete..."
    sleep 30
    # Kill the CLI if still running (it might be stuck polling)
    kill $PULL_PID 2>/dev/null || true

    # Verify image exists now
    IMAGE_COUNT=$(curl -s http://localhost:8080/images | jq '.images | length')
    if [ "$IMAGE_COUNT" -eq 0 ]; then
        echo "ERROR: Image pull failed"
        exit 1
    fi
    echo "✓ Image pull completed"
fi

echo ""
echo "Current images:"
./bin/nanofuse --api-url http://localhost:8080 image list

echo ""
echo "Cleaning up any existing test VM..."
./bin/nanofuse --api-url http://localhost:8080 vm delete test-vm-1 2>/dev/null || true

echo ""
echo "Creating a test VM..."
./bin/nanofuse --api-url http://localhost:8080 vm create test-vm-1 default --vcpus 2 --memory 512

echo ""
echo "Listing VMs..."
./bin/nanofuse --api-url http://localhost:8080 vm list

echo ""
echo "Starting VM..."
./bin/nanofuse --api-url http://localhost:8080 vm start test-vm-1

echo ""
echo "Waiting 10 seconds for boot..."
sleep 10

echo ""
echo "Checking VM status..."
./bin/nanofuse --api-url http://localhost:8080 vm status test-vm-1

echo ""
echo "Getting VM logs..."
./bin/nanofuse --api-url http://localhost:8080 vm logs test-vm-1 --tail 20

echo ""
echo "=== VM Lifecycle Test Complete ==="
echo ""
echo "To cleanup:"
echo "  ./bin/nanofuse --api-url http://localhost:8080 vm stop test-vm-1"
echo "  ./bin/nanofuse --api-url http://localhost:8080 vm delete test-vm-1"
