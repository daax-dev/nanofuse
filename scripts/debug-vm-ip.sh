#!/bin/bash
# Quick debug script to check VM IP in API responses

API_URL="http://127.0.0.1:8080"

echo "Creating VM..."
CREATE_RESPONSE=$(curl -s -X POST "$API_URL/vms" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "nanofuse-base:latest",
    "name": "debug-vm",
    "config": {
      "vcpus": 2,
      "memory_mib": 512
    }
  }')

echo "=== CREATE RESPONSE ==="
echo "$CREATE_RESPONSE" | jq '.'

echo ""
echo "=== NETWORK CONFIG FROM CREATE ==="
echo "$CREATE_RESPONSE" | jq '.config.network'

VM_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
echo ""
echo "VM ID: $VM_ID"

echo ""
echo "Inspecting VM..."
INSPECT_RESPONSE=$(curl -s "$API_URL/vms/$VM_ID")

echo "=== INSPECT RESPONSE ==="
echo "$INSPECT_RESPONSE" | jq '.'

echo ""
echo "=== NETWORK CONFIG FROM INSPECT ==="
echo "$INSPECT_RESPONSE" | jq '.config.network'
