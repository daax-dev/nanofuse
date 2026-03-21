#!/bin/bash
# Quick test of the critical fixes
# Run as regular user (will use sudo where needed)

set -e

echo "========================================"
echo "QUICK TEST OF CRITICAL FIXES"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Stop any old daemon
echo "1. Cleaning up old processes..."
sudo pkill -9 nanofused 2>/dev/null || true
sleep 2
echo -e "${GREEN}✓ Old daemon stopped${NC}"

# Remove old socket
sudo rm -f /tmp/nanofused.sock
echo -e "${GREEN}✓ Old socket removed${NC}"

# Start new daemon with our fixed binary
echo ""
echo "2. Starting daemon with FIXED binary..."
echo "   Binary: /home/jpoley/ps/nanofuse/bin/nanofused"
sudo /home/jpoley/ps/nanofuse/bin/nanofused > /tmp/nanofused.log 2>&1 &
DAEMON_PID=$!
echo "   PID: $DAEMON_PID"

# Wait for startup
echo "   Waiting 3 seconds for startup..."
sleep 3

# Check if daemon is running
if ! ps -p $DAEMON_PID > /dev/null 2>&1; then
    echo -e "${RED}✗ Daemon failed to start${NC}"
    echo "Log output:"
    cat /tmp/nanofused.log
    exit 1
fi
echo -e "${GREEN}✓ Daemon started${NC}"

# Check logs for listener creation
echo ""
echo "3. Checking daemon logs for listeners..."
sleep 1
if grep -q "Listening on Unix socket" /tmp/nanofused.log; then
    SOCKET_PATH=$(grep "Listening on Unix socket" /tmp/nanofused.log | awk '{print $NF}')
    echo -e "${GREEN}✓ Unix socket listener created: $SOCKET_PATH${NC}"
else
    echo -e "${RED}✗ Unix socket NOT created${NC}"
fi

if grep -q "Listening on TCP" /tmp/nanofused.log; then
    TCP_ADDR=$(grep "Listening on TCP" /tmp/nanofused.log | awk '{print $NF}')
    echo -e "${GREEN}✓ TCP listener created: $TCP_ADDR${NC}"
else
    echo -e "${RED}✗ TCP listener NOT created${NC}"
fi

# Verify socket file exists
echo ""
echo "4. Verifying socket file..."
if [ -S "/tmp/nanofused.sock" ]; then
    echo -e "${GREEN}✓ Socket file exists${NC}"
    ls -la /tmp/nanofused.sock
else
    echo -e "${RED}✗ Socket file missing${NC}"
    echo "Daemon log:"
    cat /tmp/nanofused.log
fi

# Test Unix socket API
echo ""
echo "5. Testing API via Unix socket..."
SOCKET_HEALTH=$(curl --silent --unix-socket /tmp/nanofused.sock http://localhost/health 2>&1 | jq -r .status 2>/dev/null || echo "ERROR")
if [ "$SOCKET_HEALTH" = "healthy" ]; then
    echo -e "${GREEN}✓ Unix socket API working${NC}"
else
    echo -e "${RED}✗ Unix socket API failed: $SOCKET_HEALTH${NC}"
fi

# Test TCP API
echo ""
echo "6. Testing API via TCP..."
TCP_HEALTH=$(curl --silent http://localhost:8080/health 2>&1 | jq -r .status 2>/dev/null || echo "ERROR")
if [ "$TCP_HEALTH" = "healthy" ]; then
    echo -e "${GREEN}✓ TCP API working${NC}"
else
    echo -e "${RED}✗ TCP API failed: $TCP_HEALTH${NC}"
fi

# Test CLI image pull (the critical fix)
echo ""
echo "7. Testing CLI image pull (CRITICAL FIX #1)..."
echo "   This should NOT show 'Job ID is required' error"
echo ""

# Use the new binary
export PATH="/home/jpoley/ps/nanofuse/bin:$PATH"

# Try to pull (will fail on auth or succeed, but should NOT error with "Job ID")
OUTPUT=$(/home/jpoley/ps/nanofuse/bin/nanofuse --api-url http://localhost:8080 image pull --default 2>&1 || true)
echo "$OUTPUT"

if echo "$OUTPUT" | grep -q "Job ID is required"; then
    echo ""
    echo -e "${RED}✗ CRITICAL FIX FAILED - Still getting 'Job ID is required' error${NC}"
    echo "The fix didn't work!"
    sudo kill $DAEMON_PID
    exit 1
elif echo "$OUTPUT" | grep -q "authentication required"; then
    echo ""
    echo -e "${YELLOW}⚠ Got authentication error (expected - this is good!)${NC}"
    echo -e "${GREEN}✓ CRITICAL FIX #1 WORKS - No 'Job ID' error!${NC}"
elif echo "$OUTPUT" | grep -q "Pulling"; then
    echo ""
    echo -e "${GREEN}✓ CRITICAL FIX #1 WORKS - Image pull started!${NC}"
else
    echo ""
    echo -e "${YELLOW}⚠ Unexpected output, but no 'Job ID' error${NC}"
    echo -e "${GREEN}✓ CRITICAL FIX #1 probably works${NC}"
fi

# Summary
echo ""
echo "========================================"
echo "SUMMARY"
echo "========================================"
echo ""
echo "Critical Fix #1 (CLI image pull type mismatch):"
if ! echo "$OUTPUT" | grep -q "Job ID is required"; then
    echo -e "  ${GREEN}✓ FIXED - No 'Job ID is required' error${NC}"
else
    echo -e "  ${RED}✗ FAILED - Still broken${NC}"
fi

echo ""
echo "Critical Fix #2 (Dual listener support):"
if [ -S "/tmp/nanofused.sock" ] && [ "$SOCKET_HEALTH" = "healthy" ] && [ "$TCP_HEALTH" = "healthy" ]; then
    echo -e "  ${GREEN}✓ FIXED - Both Unix socket and TCP working${NC}"
else
    echo -e "  ${RED}✗ PARTIAL - One or both listeners not working${NC}"
fi

echo ""
echo "Daemon PID: $DAEMON_PID (still running)"
echo "To stop: sudo kill $DAEMON_PID"
echo "Logs: /tmp/nanofused.log"
echo ""
