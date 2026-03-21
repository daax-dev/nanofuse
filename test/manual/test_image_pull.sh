#!/bin/bash
# Test Case 1: CLI Image Pull
# Tests the fix for "Job ID is required" error
#
# Prerequisites:
# - nanofused daemon running
# - Authenticated to GHCR: docker login ghcr.io
#
# Run: ./test_image_pull.sh

set -e

echo "=== Test Case 1: CLI Image Pull ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

API_URL="http://localhost:8080"

# Test 1.1: Pull default image (success path)
echo "Test 1.1: Pull default image"
echo "Command: nanofuse --api-url $API_URL image pull --default"
echo ""

# Clear any existing images first (optional)
# nanofuse --api-url $API_URL image delete ... 2>/dev/null || true

# Run pull command
if nanofuse --api-url $API_URL image pull --default; then
    echo -e "${GREEN}✓ Pull command succeeded${NC}"
else
    echo -e "${RED}✗ Pull command failed${NC}"
    exit 1
fi

echo ""
echo "Waiting for pull to complete..."
sleep 5

# Verify image exists
echo ""
echo "Test 1.2: Verify image appeared in list"
IMAGE_COUNT=$(nanofuse --api-url $API_URL image list --json | jq '.images | length')
echo "Images found: $IMAGE_COUNT"

if [ "$IMAGE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Image successfully pulled and listed${NC}"
    nanofuse --api-url $API_URL image list
else
    echo -e "${RED}✗ No images found after pull${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== All Image Pull Tests PASSED ===${NC}"
