#!/bin/bash
# NanoFuse API Demo Script
# Demonstrates a complete VM lifecycle using the REST API

set -e

# Auto-detect if daemon is on Unix socket or TCP
SOCKET="/tmp/nanofused.sock"
TCP_ENDPOINT="http://localhost:8080"
API="http://localhost"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to call API (auto-detect Unix socket vs TCP)
api_call() {
  if [ -S "$SOCKET" ]; then
    # Unix socket exists, use it
    curl --silent --unix-socket "$SOCKET" "$@"
  else
    # Fall back to TCP
    local url="$1"
    shift
    curl --silent "$@" "${TCP_ENDPOINT}${url#http://localhost}"
  fi
}

# Function to print section headers
section() {
  echo -e "\n${YELLOW}===> $1${NC}"
}

# Function to print success
success() {
  echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error
error() {
  echo -e "${RED}✗ $1${NC}"
  exit 1
}

# Check if daemon is running
section "1. Checking API health..."
if ! api_call "$API/health" > /dev/null 2>&1; then
  error "Daemon not running. Start it with: sudo nanofused"
fi

HEALTH=$(api_call "$API/health" | jq -r .status)
if [ "$HEALTH" = "healthy" ]; then
  success "API is healthy"
else
  error "API health check failed (got: $HEALTH)"
fi

# List current VMs
section "2. Listing current VMs..."
VM_COUNT=$(api_call "$API/vms" | jq '.vms | length')
success "Found $VM_COUNT existing VMs"

# Check if we have images
section "3. Checking for images..."
IMAGE_COUNT=$(api_call "$API/images" | jq '.images | length')
if [ "$IMAGE_COUNT" -eq 0 ]; then
  echo "No images found. You need to pull an image first."
  echo "See docs/API_QUICK_START.md for instructions on pulling images."
  exit 0
fi
success "Found $IMAGE_COUNT images"

# Get first available image
IMAGE=$(api_call "$API/images" | jq -r '.images[0].tags[0]')
success "Using image: $IMAGE"

# Create VM
section "4. Creating VM..."
VM_RESPONSE=$(api_call -X POST -H "Content-Type: application/json" \
  -d "{\"name\":\"demo-vm-$(date +%s)\",\"image\":\"$IMAGE\",\"vcpus\":2,\"memory_mib\":512}" \
  "$API/vms")

VM_ID=$(echo "$VM_RESPONSE" | jq -r .id)
VM_NAME=$(echo "$VM_RESPONSE" | jq -r .name)
success "Created VM: $VM_NAME (ID: $VM_ID)"

# Start VM
section "5. Starting VM..."
api_call -X POST "$API/vms/$VM_ID/start" > /dev/null
success "VM started"

# Wait for boot
section "6. Waiting for VM to boot (5 seconds)..."
sleep 5

# Check VM status
section "7. Checking VM status..."
VM_STATUS=$(api_call "$API/vms/$VM_ID" | jq -r .state)
VM_IP=$(api_call "$API/vms/$VM_ID" | jq -r .network.ip)
success "VM state: $VM_STATUS"
success "VM IP: $VM_IP"

# Get logs
section "8. Getting VM console logs (last 10 lines)..."
api_call "$API/vms/$VM_ID/logs?tail=10" | jq -r .logs | head -10
success "Logs retrieved"

# Wait a bit
echo -e "\n${YELLOW}VM is running. Waiting 10 seconds before cleanup...${NC}"
sleep 10

# Stop VM
section "9. Stopping VM..."
api_call -X POST "$API/vms/$VM_ID/stop" > /dev/null
success "VM stopped"

# Delete VM
section "10. Deleting VM..."
api_call -X DELETE "$API/vms/$VM_ID" > /dev/null
success "VM deleted"

# Final status
section "Demo Complete!"
echo -e "${GREEN}Successfully demonstrated complete VM lifecycle:${NC}"
echo "  ✓ Health check"
echo "  ✓ VM creation"
echo "  ✓ VM start"
echo "  ✓ Status check"
echo "  ✓ Log retrieval"
echo "  ✓ VM stop"
echo "  ✓ VM deletion"
echo ""
echo "For more examples, see: docs/API_QUICK_START.md"
