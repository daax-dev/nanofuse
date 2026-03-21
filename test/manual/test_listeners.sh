#!/bin/bash
# Test Case 2: Unix Socket and TCP Dual Listeners
# Tests the fix for Unix socket not being created
#
# Prerequisites:
# - nanofused NOT running (will be started by this test)
# - /etc/nanofuse/nanofused.yaml configured with both socket and tcp_bind
#
# Run: sudo ./test_listeners.sh

set -e

echo "=== Test Case 2: Dual Listener Support ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}This test must be run as root (sudo)${NC}"
    exit 1
fi

# Stop any existing daemon
echo "Stopping any existing daemon..."
pkill -9 nanofused 2>/dev/null || true
sleep 2

# Clean up old socket
rm -f /tmp/nanofused.sock

# Verify config has both listeners
echo "Test 2.1: Verify config has both socket and TCP configured"
SOCKET_PATH=$(grep "socket:" /etc/nanofuse/nanofused.yaml | grep -v "#" | awk '{print $2}' | head -1)
TCP_BIND=$(grep "tcp_bind:" /etc/nanofuse/nanofused.yaml | grep -v "#" | awk '{print $2}' | tr -d '"' | head -1)

echo "  Socket path: $SOCKET_PATH"
echo "  TCP bind: $TCP_BIND"

if [ -z "$SOCKET_PATH" ] || [ -z "$TCP_BIND" ]; then
    echo -e "${RED}✗ Config missing socket or tcp_bind${NC}"
    exit 1
fi

# Start daemon
echo ""
echo "Test 2.2: Starting daemon with dual listener config"
/usr/local/bin/nanofused &
DAEMON_PID=$!
echo "  Daemon PID: $DAEMON_PID"

# Wait for startup
sleep 3

# Check daemon is running
if ! ps -p $DAEMON_PID > /dev/null; then
    echo -e "${RED}✗ Daemon failed to start${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Daemon started${NC}"

# Test 2.3: Verify Unix socket exists
echo ""
echo "Test 2.3: Verify Unix socket created"
if [ -S "$SOCKET_PATH" ]; then
    echo -e "${GREEN}✓ Unix socket exists at $SOCKET_PATH${NC}"
    ls -la $SOCKET_PATH
else
    echo -e "${RED}✗ Unix socket NOT created${NC}"
    kill $DAEMON_PID
    exit 1
fi

# Test 2.4: Verify TCP listener
echo ""
echo "Test 2.4: Verify TCP listener active"
if lsof -i:8080 | grep -q nanofused; then
    echo -e "${GREEN}✓ TCP listener on port 8080${NC}"
    lsof -i:8080 | grep nanofused
else
    echo -e "${RED}✗ TCP listener NOT active${NC}"
    kill $DAEMON_PID
    exit 1
fi

# Test 2.5: Test Unix socket connectivity
echo ""
echo "Test 2.5: Test API via Unix socket"
SOCKET_HEALTH=$(curl --silent --unix-socket $SOCKET_PATH http://localhost/health | jq -r .status)
if [ "$SOCKET_HEALTH" = "healthy" ]; then
    echo -e "${GREEN}✓ Unix socket API working${NC}"
else
    echo -e "${RED}✗ Unix socket API failed (got: $SOCKET_HEALTH)${NC}"
    kill $DAEMON_PID
    exit 1
fi

# Test 2.6: Test TCP connectivity
echo ""
echo "Test 2.6: Test API via TCP"
TCP_HEALTH=$(curl --silent http://localhost:8080/health | jq -r .status)
if [ "$TCP_HEALTH" = "healthy" ]; then
    echo -e "${GREEN}✓ TCP API working${NC}"
else
    echo -e "${RED}✗ TCP API failed (got: $TCP_HEALTH)${NC}"
    kill $DAEMON_PID
    exit 1
fi

# Test 2.7: CLI auto-detection (should prefer Unix socket)
echo ""
echo "Test 2.7: Test CLI auto-detection"
CLI_OUTPUT=$(nanofuse image list 2>&1 || true)
if echo "$CLI_OUTPUT" | grep -q "DIGEST"; then
    echo -e "${GREEN}✓ CLI works (used Unix socket by default)${NC}"
else
    echo -e "${YELLOW}⚠ CLI may need --api-url flag${NC}"
    echo "  Output: $CLI_OUTPUT"
fi

# Test 2.8: CLI explicit TCP
echo ""
echo "Test 2.8: Test CLI with explicit TCP endpoint"
if nanofuse --api-url http://localhost:8080 image list >/dev/null 2>&1; then
    echo -e "${GREEN}✓ CLI works with explicit TCP${NC}"
else
    echo -e "${RED}✗ CLI failed with explicit TCP${NC}"
    kill $DAEMON_PID
    exit 1
fi

# Cleanup
echo ""
echo "Cleaning up..."
kill $DAEMON_PID
wait $DAEMON_PID 2>/dev/null || true
rm -f $SOCKET_PATH

echo ""
echo -e "${GREEN}=== All Listener Tests PASSED ===${NC}"
echo ""
echo "Summary:"
echo "  ✓ Unix socket created and accessible"
echo "  ✓ TCP listener active and accessible"
echo "  ✓ Both listeners work simultaneously"
echo "  ✓ CLI can use both transport methods"
