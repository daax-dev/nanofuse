#!/bin/bash
# Deploy todo-app to NanoFuse VM and validate

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="ghcr.io/daax-dev/nanofuse/todo-app:latest"
VM_NAME="todo-app-test"
VM_MEMORY=1024
VM_VCPUS=2

echo -e "${GREEN}=== NanoFuse Todo App Deployment ===${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root (for nanofused)${NC}"
    echo "Usage: sudo $0"
    exit 1
fi

# Check if nanofused is installed
if ! command -v nanofused &> /dev/null; then
    echo -e "${RED}Error: nanofused not found${NC}"
    echo "Please install nanofused first"
    exit 1
fi

# Step 1: Check if nanofused is running
echo -e "${YELLOW}[1/6] Checking nanofused daemon...${NC}"
if pgrep -x nanofused > /dev/null; then
    echo -e "${GREEN}✓ nanofused is running${NC}"
else
    echo -e "${YELLOW}Starting nanofused daemon...${NC}"
    nanofused > /var/log/nanofused.log 2>&1 &
    sleep 3

    if pgrep -x nanofused > /dev/null; then
        echo -e "${GREEN}✓ nanofused started successfully${NC}"
    else
        echo -e "${RED}✗ Failed to start nanofused${NC}"
        cat /var/log/nanofused.log
        exit 1
    fi
fi

# Step 2: Check if image exists locally or pull it
echo -e "${YELLOW}[2/6] Checking container image...${NC}"
if docker images | grep -q "nanofuse-todo-app.*test"; then
    echo -e "${GREEN}✓ Image exists locally${NC}"
    IMAGE_NAME="nanofuse-todo-app:test"
else
    echo -e "${YELLOW}Image not found locally, will use: ${IMAGE_NAME}${NC}"
fi

# Step 3: Export image to tar if needed
# Note: Using docker export (container filesystem) instead of docker save (OCI format)
# docker save gives OCI metadata, not the actual Linux filesystem!
echo -e "${YELLOW}[3/6] Preparing image for NanoFuse...${NC}"
IMAGE_TAR="/tmp/todo-app.tar"
if [ -f "$IMAGE_TAR" ]; then
    echo -e "${YELLOW}Removing old image tar...${NC}"
    rm -f "$IMAGE_TAR"
fi

CONTAINER_ID=$(docker create "$IMAGE_NAME")
docker export "$CONTAINER_ID" -o "$IMAGE_TAR"
docker rm "$CONTAINER_ID" > /dev/null
echo -e "${GREEN}✓ Container filesystem exported to ${IMAGE_TAR}${NC}"

# Step 4: Clean up any existing VM with same name
echo -e "${YELLOW}[4/6] Cleaning up existing VM...${NC}"
if nanofuse vm list 2>/dev/null | grep -q "$VM_NAME"; then
    echo -e "${YELLOW}Stopping existing VM...${NC}"
    nanofuse vm stop "$VM_NAME" 2>/dev/null || true
    sleep 2
    nanofuse vm delete "$VM_NAME" 2>/dev/null || true
    sleep 1
fi
echo -e "${GREEN}✓ Ready to create new VM${NC}"

# Step 5: Create VM from image
echo -e "${YELLOW}[5/6] Creating VM from image...${NC}"
echo "VM Name: $VM_NAME"
echo "Memory: ${VM_MEMORY}MB"
echo "vCPUs: $VM_VCPUS"
echo "Image: $IMAGE_NAME"

# Note: nanofuse might need the image in a registry or specific format
# For now, we'll try creating with the image reference
if nanofuse vm create "$VM_NAME" \
    --image "$IMAGE_NAME" \
    --memory "$VM_MEMORY" \
    --vcpus "$VM_VCPUS" \
    2>&1 | tee /tmp/vm-create.log; then
    echo -e "${GREEN}✓ VM created successfully${NC}"
else
    echo -e "${RED}✗ VM creation failed${NC}"
    echo "Check logs: /tmp/vm-create.log"
    cat /tmp/vm-create.log
    exit 1
fi

# Step 6: Wait for VM to start and verify
echo -e "${YELLOW}[6/6] Waiting for VM to start...${NC}"
sleep 5

VM_STATUS=$(nanofuse vm status "$VM_NAME" 2>/dev/null || echo "unknown")
echo "VM Status: $VM_STATUS"

if echo "$VM_STATUS" | grep -iq "running"; then
    echo -e "${GREEN}✓ VM is running!${NC}"
else
    echo -e "${YELLOW}VM status: $VM_STATUS${NC}"
fi

# Get VM info
echo ""
echo -e "${GREEN}=== VM Information ===${NC}"
nanofuse vm info "$VM_NAME" 2>/dev/null || echo "Could not get VM info"

# Try to get IP address
echo ""
echo -e "${YELLOW}Attempting to get VM IP address...${NC}"
VM_IP=$(nanofuse vm info "$VM_NAME" 2>/dev/null | grep -oP 'IP.*:\s*\K[\d.]+' || echo "")

if [ -n "$VM_IP" ]; then
    echo -e "${GREEN}VM IP: $VM_IP${NC}"
    echo ""
    echo -e "${YELLOW}Wait 10 seconds for services to start...${NC}"
    sleep 10

    echo -e "${YELLOW}Testing connectivity...${NC}"
    if curl -s --connect-timeout 5 "http://${VM_IP}:8080/health" > /dev/null; then
        echo -e "${GREEN}✓ Backend is responding!${NC}"
        curl -s "http://${VM_IP}:8080/health" | jq . || cat
    else
        echo -e "${RED}✗ Backend not responding yet${NC}"
        echo "This might be normal - services may still be starting"
    fi
else
    echo -e "${YELLOW}Could not determine VM IP address${NC}"
    echo "Check VM status: nanofuse vm info $VM_NAME"
fi

echo ""
echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo "Next steps:"
echo "  1. Check VM status: nanofuse vm status $VM_NAME"
echo "  2. Get VM info: nanofuse vm info $VM_NAME"
echo "  3. Test API: curl http://\$VM_IP:8080/health"
echo "  4. Run tests: ./scripts/test-vm.sh"
echo "  5. View logs: nanofuse vm logs $VM_NAME"
echo ""
