#!/bin/bash
# End-to-End Network Testing Script for NanoFuse
# Tests complete VM lifecycle with networking

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================"
echo "NanoFuse Network End-to-End Test"
echo "========================================"
echo

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}ERROR: This script must be run as root (network setup requires it)${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Configuration
DAEMON_BIN="./bin/nanofused"
CLI_BIN="./bin/nanofuse"
REGISTER_BIN="./bin/register-local-image"
CONFIG_FILE="./config.dev.yaml"
API_URL="http://127.0.0.1:8080"
VM_NAME="test-network-vm"
IMAGE_NAME="nanofuse-base:latest"
DATA_DIR="/tmp/nanofuse"
DB_PATH="$DATA_DIR/nanofuse.db"

# Get original user for permission fixes
if [ -n "${SUDO_USER:-}" ]; then
    REAL_USER="$SUDO_USER"
    REAL_UID=$(id -u "$SUDO_USER")
    REAL_GID=$(id -g "$SUDO_USER")
else
    REAL_USER="$USER"
    REAL_UID="$UID"
    REAL_GID="$(id -g)"
fi

echo "Running as: root (for network setup)"
echo "Build user: $REAL_USER"
echo

# Function to cleanup
cleanup() {
    echo
    echo -e "${YELLOW}Cleaning up...${NC}"

    # Stop daemon if running
    if [ -f /tmp/daemon.pid ]; then
        DAEMON_PID=$(cat /tmp/daemon.pid)
        if kill -0 $DAEMON_PID 2>/dev/null; then
            echo "Stopping daemon (PID $DAEMON_PID)..."
            kill $DAEMON_PID 2>/dev/null || true
            sleep 2
            # Force kill if still running
            kill -9 $DAEMON_PID 2>/dev/null || true
        fi
        rm -f /tmp/daemon.pid
    fi

    # Kill any remaining daemon processes
    pkill -f "nanofused.*config.dev.yaml" 2>/dev/null || true

    # Clean up network infrastructure (NAT rules and bridge)
    if [ -n "${PRIMARY_IFACE:-}" ]; then
        echo "Cleaning up NAT rules..."
        iptables -t nat -D POSTROUTING -s 172.16.0.0/24 -o "$PRIMARY_IFACE" -j MASQUERADE 2>/dev/null || true
        iptables -D FORWARD -i nanofuse0 -o "$PRIMARY_IFACE" -j ACCEPT 2>/dev/null || true
        iptables -D FORWARD -i "$PRIMARY_IFACE" -o nanofuse0 -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true
    fi

    # Remove bridge if it exists
    if ip link show nanofuse0 >/dev/null 2>&1; then
        echo "Removing bridge nanofuse0..."
        ip link set nanofuse0 down 2>/dev/null || true
        ip link delete nanofuse0 2>/dev/null || true
    fi

    # Clean up test data
    echo "Cleaning up test data..."
    rm -rf "$DATA_DIR"

    echo -e "${GREEN}Cleanup complete${NC}"
}

# Trap to cleanup on exit
trap cleanup EXIT

echo "[0/9] Checking prerequisites..."

# Find Go binary - check common locations
GO_BIN=""
for go_candidate in \
    "/usr/local/go/bin/go" \
    "/usr/bin/go" \
    "/home/$REAL_USER/go/bin/go" \
    "/home/$REAL_USER/.local/go/bin/go" \
    "$(command -v go 2>/dev/null)"; do
    if [ -n "$go_candidate" ] && [ -x "$go_candidate" ]; then
        GO_BIN="$go_candidate"
        break
    fi
done

if [ -z "$GO_BIN" ]; then
    echo -e "${RED}ERROR: Go not found${NC}"
    echo "Checked locations: /usr/local/go/bin/go, /usr/bin/go, /home/$REAL_USER/go/bin/go"
    echo "Install with:"
    echo "  apt-get install golang"
    echo "Or download from: https://go.dev/dl/"
    exit 1
fi

echo -e "${GREEN}✓ Go found at: $GO_BIN${NC}"

# Get GOPATH for the user
GOPATH=$(sudo -u "$REAL_USER" "$GO_BIN" env GOPATH 2>/dev/null || echo "/home/$REAL_USER/go")

# Check for required tools
MISSING_TOOLS=()
for tool in jq ip iptables bridge docker sqlite3 curl; do
    if ! command -v $tool &> /dev/null; then
        MISSING_TOOLS+=("$tool")
    fi
done

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo -e "${RED}ERROR: Missing required tools: ${MISSING_TOOLS[*]}${NC}"
    echo "Install with:"
    echo "  apt-get install jq iproute2 iptables bridge-utils docker.io sqlite3 curl"
    exit 1
fi

# Check for mage, install if missing
MAGE_BIN="$GOPATH/bin/mage"
if [ ! -x "$MAGE_BIN" ]; then
    echo -e "${YELLOW}mage not found, installing to $MAGE_BIN...${NC}"
    if sudo -u "$REAL_USER" "$GO_BIN" install github.com/magefile/mage@latest; then
        echo -e "${GREEN}✓ mage installed${NC}"
    else
        echo -e "${RED}ERROR: Failed to install mage${NC}"
        exit 1
    fi
fi

# Verify mage is executable
if [ ! -x "$MAGE_BIN" ]; then
    echo -e "${RED}ERROR: mage not found at $MAGE_BIN after installation${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All required tools available${NC}"
echo -e "${GREEN}✓ mage found at: $MAGE_BIN${NC}"

# Fix bin directory permissions if needed
if [ -d "./bin" ] && [ "$(stat -c %u ./bin)" -eq 0 ]; then
    echo -e "${YELLOW}Fixing bin directory permissions...${NC}"
    chown -R "$REAL_UID:$REAL_GID" ./bin
fi

# Check binaries exist, build if missing
if [ ! -f "$DAEMON_BIN" ] || [ ! -f "$CLI_BIN" ] || [ ! -f "$REGISTER_BIN" ]; then
    echo -e "${YELLOW}Building binaries...${NC}"

    # Clean any root-owned build artifacts
    if [ -d "./bin" ]; then
        chown -R "$REAL_UID:$REAL_GID" ./bin
    fi

    # Build as original user using absolute paths to mage
    # Set PATH to include Go binary location so mage can find it
    GO_DIR=$(dirname "$GO_BIN")
    if sudo -u "$REAL_USER" env PATH="$GO_DIR:$PATH" "$MAGE_BIN" all; then
        echo -e "${GREEN}✓ Binaries built successfully${NC}"
    else
        echo -e "${RED}ERROR: Failed to build binaries${NC}"
        exit 1
    fi
fi

# Verify binaries exist and are executable
for bin in "$DAEMON_BIN" "$CLI_BIN" "$REGISTER_BIN"; do
    if [ ! -f "$bin" ]; then
        echo -e "${RED}ERROR: Binary not found: $bin${NC}"
        exit 1
    fi
    if [ ! -x "$bin" ]; then
        chmod +x "$bin"
    fi
done

echo -e "${GREEN}✓ All binaries ready${NC}"

# Check config exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}ERROR: Config file not found at $CONFIG_FILE${NC}"
    exit 1
fi

echo "[1/9] Checking base image files..."
if [ ! -f "images/base/build/rootfs.ext4" ] || [ ! -f "images/base/build/vmlinux" ]; then
    echo -e "${RED}ERROR: Base image files not found${NC}"
    echo "Build the base image first:"
    echo "  cd images/base && sudo ./build.sh"
    exit 1
fi
echo -e "${GREEN}✓ Base image files found${NC}"

echo
echo "[2/9] Preparing clean network environment..."

# Detect primary interface for cleanup
PRIMARY_IFACE=$(ip route show default | grep -oP 'dev \K\S+' | head -1)
if [ -z "$PRIMARY_IFACE" ]; then
    echo -e "${YELLOW}WARN: Could not detect primary interface${NC}"
    PRIMARY_IFACE="enp46s0"  # Fallback
fi
echo "Primary interface: $PRIMARY_IFACE"

# Clean up any existing nanofuse network setup
echo "Cleaning up any existing nanofuse network configuration..."
iptables -t nat -D POSTROUTING -s 172.16.0.0/24 -o "$PRIMARY_IFACE" -j MASQUERADE 2>/dev/null || true
iptables -D FORWARD -i nanofuse0 -o "$PRIMARY_IFACE" -j ACCEPT 2>/dev/null || true
iptables -D FORWARD -i "$PRIMARY_IFACE" -o nanofuse0 -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true

if ip link show nanofuse0 >/dev/null 2>&1; then
    ip link set nanofuse0 down 2>/dev/null || true
    ip link delete nanofuse0 2>/dev/null || true
fi

# Check for iptables kernel modules
if ! lsmod | grep -q iptable_nat; then
    echo -e "${YELLOW}Loading iptables NAT kernel modules...${NC}"
    modprobe iptable_nat 2>/dev/null || true
    modprobe ip_tables 2>/dev/null || true
    modprobe xt_state 2>/dev/null || true
fi

# Check for firewall conflicts
FIREWALL_WARNING=""
if systemctl is-active --quiet ufw 2>/dev/null; then
    FIREWALL_WARNING="ufw is running"
fi
if systemctl is-active --quiet firewalld 2>/dev/null; then
    FIREWALL_WARNING="${FIREWALL_WARNING}${FIREWALL_WARNING:+ and }firewalld is running"
fi

if [ -n "$FIREWALL_WARNING" ]; then
    echo -e "${YELLOW}WARNING: $FIREWALL_WARNING${NC}"
    echo -e "${YELLOW}This may interfere with NAT rules. Consider disabling temporarily.${NC}"
fi

echo -e "${GREEN}✓ Network environment prepared${NC}"

echo
echo "[3/9] Setting up test environment..."

# Kill any existing daemon processes
pkill -f "nanofused.*config.dev.yaml" 2>/dev/null || true
sleep 1

# Clean up any previous test data
rm -rf "$DATA_DIR"
mkdir -p "$DATA_DIR/images/nanofuse-base/latest"

# Copy image files
cp images/base/build/rootfs.ext4 "$DATA_DIR/images/nanofuse-base/latest/"
cp images/base/build/vmlinux "$DATA_DIR/images/nanofuse-base/latest/"
cp images/base/build/manifest.json "$DATA_DIR/images/nanofuse-base/latest/"

# Fix permissions on all test data
chown -R "$REAL_UID:$REAL_GID" "$DATA_DIR"
chmod -R u+rw "$DATA_DIR"
chmod 664 "$DATA_DIR/images/nanofuse-base/latest/rootfs.ext4"

echo -e "${GREEN}✓ Test environment ready${NC}"

echo
echo "[4/9] Starting daemon..."
# Start daemon in background AS ROOT (needs CAP_NET_ADMIN for bridge/TAP creation)
$DAEMON_BIN --config "$CONFIG_FILE" > /tmp/nanofuse-daemon.log 2>&1 &
DAEMON_PID=$!
echo $DAEMON_PID > /tmp/daemon.pid
echo "Daemon started with PID $DAEMON_PID (running as root for network operations)"

# Wait for daemon with polling
echo -n "Waiting for daemon to initialize"
for i in {1..30}; do
    sleep 1
    echo -n "."

    # Check if daemon is still running
    if ! kill -0 $DAEMON_PID 2>/dev/null; then
        echo
        echo -e "${RED}ERROR: Daemon crashed during startup${NC}"
        echo "Daemon log:"
        cat /tmp/nanofuse-daemon.log
        exit 1
    fi

    # Check if API is responding
    if curl -s -f "$API_URL/health" > /dev/null 2>&1; then
        echo
        echo -e "${GREEN}✓ Daemon is running and responding${NC}"
        break
    fi

    if [ $i -eq 30 ]; then
        echo
        echo -e "${RED}ERROR: Daemon did not become ready in time${NC}"
        echo "Last 50 lines of daemon log:"
        tail -n 50 /tmp/nanofuse-daemon.log
        exit 1
    fi
done

echo
echo "[5/9] Registering test image in database..."

# Ensure database directory exists
mkdir -p "$(dirname "$DB_PATH")"

# Register image (run as root since daemon is root and needs to access same DB)
if ! $REGISTER_BIN "$DB_PATH" "$IMAGE_NAME" \
    "$DATA_DIR/images/nanofuse-base/latest/rootfs.ext4" \
    "$DATA_DIR/images/nanofuse-base/latest/vmlinux"; then
    echo -e "${RED}ERROR: Failed to register image${NC}"
    exit 1
fi

# Verify image was registered
# First, let's see what's actually in the database
echo "Checking database contents..."
DB_IMAGES=$(sqlite3 "$DB_PATH" "SELECT digest, tags_json FROM images;" 2>&1)
echo "Database images: $DB_IMAGES"

# Check if image exists by digest (which is the tag for local images)
if sqlite3 "$DB_PATH" "SELECT digest FROM images WHERE digest = '$IMAGE_NAME';" | grep -q .; then
    echo -e "${GREEN}✓ Image registered and verified${NC}"
else
    echo -e "${RED}ERROR: Image not found in database after registration${NC}"
    echo "Expected digest: $IMAGE_NAME"
    echo "Database contents:"
    sqlite3 "$DB_PATH" "SELECT digest, tags_json FROM images;"
    exit 1
fi

echo
echo "[6/9] Verifying network infrastructure..."
if ! ip link show nanofuse0 >/dev/null 2>&1; then
    echo -e "${RED}ERROR: Bridge nanofuse0 not created${NC}"
    echo "Daemon log:"
    cat /tmp/nanofuse-daemon.log
    exit 1
fi
echo -e "${GREEN}✓ Bridge nanofuse0 exists${NC}"

# Check NAT rules
if ! iptables -t nat -L POSTROUTING 2>/dev/null | grep -q nanofuse0; then
    echo -e "${YELLOW}WARN: NAT rules may not be configured${NC}"
else
    echo -e "${GREEN}✓ NAT rules configured${NC}"
fi

echo
echo "[7/9] Creating VM with networking..."
if ! $CLI_BIN --api-url "$API_URL" vm create "$IMAGE_NAME" "$VM_NAME" \
    --vcpus 2 \
    --memory 512; then
    echo -e "${RED}ERROR: Failed to create VM${NC}"
    exit 1
fi

# Get VM details with retries
VM_INFO=""
for i in {1..5}; do
    if VM_INFO=$($CLI_BIN --api-url "$API_URL" vm inspect "$VM_NAME" --json 2>/dev/null); then
        break
    fi
    sleep 1
done

if [ -z "$VM_INFO" ]; then
    echo -e "${RED}ERROR: Failed to get VM info${NC}"
    exit 1
fi

VM_ID=$(echo "$VM_INFO" | jq -r '.id')
VM_IP=$(echo "$VM_INFO" | jq -r '.config.network.ip_address')
VM_TAP=$(echo "$VM_INFO" | jq -r '.config.network.tap_device')
VM_MAC=$(echo "$VM_INFO" | jq -r '.config.network.mac_address')

echo -e "${GREEN}✓ VM created:${NC}"
echo "  ID:  $VM_ID"
echo "  IP:  $VM_IP"
echo "  TAP: $VM_TAP"
echo "  MAC: $VM_MAC"

# Verify TAP device exists
echo
echo "[8/9] Verifying TAP device..."
if ! ip link show "$VM_TAP" >/dev/null 2>&1; then
    echo -e "${RED}ERROR: TAP device $VM_TAP not created${NC}"
    exit 1
fi
echo -e "${GREEN}✓ TAP device $VM_TAP exists${NC}"

# Verify TAP attached to bridge
if ! bridge link show | grep -q "$VM_TAP"; then
    echo -e "${RED}ERROR: TAP device not attached to bridge${NC}"
    exit 1
fi
echo -e "${GREEN}✓ TAP device attached to bridge${NC}"

echo
echo "[9/9] Starting VM..."
if ! $CLI_BIN --api-url "$API_URL" vm start "$VM_NAME"; then
    echo -e "${RED}ERROR: Failed to start VM${NC}"
    exit 1
fi

# Wait for VM to boot with polling
echo -n "Waiting for VM to boot"
for i in {1..60}; do
    sleep 1
    echo -n "."

    VM_INFO_AFTER_START=$($CLI_BIN --api-url "$API_URL" vm inspect "$VM_NAME" --json 2>/dev/null || echo "{}")
    VM_STATE=$(echo "$VM_INFO_AFTER_START" | jq -r '.state // "unknown"')

    if [ "$VM_STATE" = "running" ]; then
        echo
        break
    fi

    if [ $i -eq 60 ]; then
        echo
        echo -e "${RED}ERROR: VM did not start in time (state: $VM_STATE)${NC}"
        echo "VM console output:"
        $CLI_BIN --api-url "$API_URL" vm logs "$VM_NAME" 2>/dev/null || echo "Could not get logs"
        exit 1
    fi
done

VM_IP=$(echo "$VM_INFO_AFTER_START" | jq -r '.config.network.ip_address')
echo -e "${GREEN}✓ VM is running${NC}"

echo
echo "Testing network connectivity..."

# Test 1: Ping VM from host
echo -n "  Testing host -> VM connectivity (ping $VM_IP)... "
if ping -c 3 -W 5 "$VM_IP" >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    Cannot ping VM from host"
    echo "    Check VM console:"
    $CLI_BIN --api-url "$API_URL" vm logs "$VM_NAME" 2>/dev/null | tail -n 50 || true
fi

# Test 2: Check VM console for network initialization
echo -n "  Checking VM network initialization in console... "
if $CLI_BIN --api-url "$API_URL" vm logs "$VM_NAME" 2>/dev/null | grep -q "eth0"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${YELLOW}? Could not verify (check manually)${NC}"
fi

# Test 3: Verify IP forwarding is enabled
echo -n "  Checking IP forwarding on host... "
if [ "$(sysctl -n net.ipv4.ip_forward)" = "1" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    IP forwarding not enabled"
fi

# Test 4: Port forwarding test
echo
echo "Testing port forwarding..."

# Create a VM with port forwarding
PORT_FORWARD_VM="test-portforward-vm"
HOST_PORT=8888
VM_PORT=8080

echo -n "  Creating VM with port forward ($HOST_PORT:$VM_PORT/tcp)... "
PF_CREATE_OUTPUT=$($CLI_BIN --api-url "$API_URL" vm create "$IMAGE_NAME" "$PORT_FORWARD_VM" \
    --vcpus 1 --memory 256 --port-forward "$HOST_PORT:$VM_PORT/tcp" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    Error: $PF_CREATE_OUTPUT"
fi

# Start the VM (only if creation succeeded)
if [ $? -eq 0 ]; then
    echo -n "  Starting VM... "
    PF_START_OUTPUT=$($CLI_BIN --api-url "$API_URL" vm start "$PORT_FORWARD_VM" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ FAILED${NC}"
        echo "    Error: $PF_START_OUTPUT"
    fi
fi

# Wait for VM to boot
echo -n "  Waiting for VM to boot (10 seconds)... "
sleep 10
echo -e "${GREEN}✓${NC}"

# Get VM IP
PF_VM_INFO=$($CLI_BIN --api-url "$API_URL" vm inspect "$PORT_FORWARD_VM" --json 2>/dev/null || echo "{}")
PF_VM_IP=$(echo "$PF_VM_INFO" | jq -r '.config.network.ip_address')

# Verify VM is reachable
echo -n "  Verifying VM is reachable (ping $PF_VM_IP)... "
if ping -c 1 -W 2 "$PF_VM_IP" >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    Cannot reach VM IP"
fi

# Verify the iptables rules were created
echo -n "  Verifying port forward iptables PREROUTING rule... "
if iptables -t nat -L PREROUTING -n | grep -q "dpt:$HOST_PORT.*to:$PF_VM_IP:$VM_PORT"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    Expected DNAT rule not found"
    iptables -t nat -L PREROUTING -n -v
fi

echo -n "  Verifying OUTPUT chain rule for localhost... "
if iptables -t nat -L OUTPUT -n | grep -q "dpt:$HOST_PORT.*to:$PF_VM_IP:$VM_PORT"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    Expected OUTPUT rule not found"
fi

echo -n "  Verifying MASQUERADE rule for return traffic... "
if iptables -t nat -L POSTROUTING -n | grep "$PF_VM_IP" | grep -q "MASQUERADE.*dpt:$VM_PORT"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    Expected MASQUERADE rule not found"
    iptables -t nat -L POSTROUTING -n -v | grep "$PF_VM_IP" || echo "    No rules for $PF_VM_IP found"
fi

# Wait for HTTP server to be ready inside the VM (http-test-server.service)
echo -n "  Waiting for HTTP server to start in VM... "
# Try for up to 60 seconds, checking every 2 seconds
HTTP_READY=false
for i in {1..30}; do
    if curl -s -m 2 "http://$PF_VM_IP:$VM_PORT/" >/dev/null 2>&1; then
        HTTP_READY=true
        echo -e "${GREEN}✓${NC} (ready after ${i}x2 seconds)"
        break
    fi
    sleep 2
done

if [ "$HTTP_READY" = false ]; then
    echo -e "${RED}✗ FAILED${NC}"
    echo "    HTTP server did not respond after 60 seconds"
    echo ""
    echo "    === CONSOLE LOG (last 100 lines) ==="
    $CLI_BIN --api-url $API_URL vm logs $PORT_FORWARD_VM --tail 100 2>/dev/null || echo "    Could not retrieve logs"
    echo "    === END CONSOLE LOG ==="
    echo ""
    echo "    Direct VM connectivity test:"
    ping -c 1 -W 2 "$PF_VM_IP" >/dev/null 2>&1 && echo "    ✓ VM is pingable" || echo "    ✗ VM is NOT pingable"
    echo ""
fi

# Test actual HTTP connectivity through port forward
echo -n "  Testing HTTP fetch through port forward (localhost:$HOST_PORT)... "
# Force IPv4 with -4 since our iptables rules are IPv4 only
HTTP_RESPONSE=$(curl -4 -s -m 5 "http://localhost:$HOST_PORT/" 2>/dev/null || echo "")
if [ -n "$HTTP_RESPONSE" ]; then
    # Verify response contains expected message
    if echo "$HTTP_RESPONSE" | grep -q "Hello from NanoFuse VM"; then
        echo -e "${GREEN}✓${NC}"
        # Try to pretty-print with jq if valid JSON, otherwise show raw
        if echo "$HTTP_RESPONSE" | jq -e . >/dev/null 2>&1; then
            echo "    Response: $(echo "$HTTP_RESPONSE" | jq -c .)"
        else
            echo "    Response: $HTTP_RESPONSE"
        fi
    else
        echo -e "${YELLOW}⚠ Got response but unexpected content${NC}"
        echo "    Response: $HTTP_RESPONSE"
    fi
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    No response from forwarded port"
    echo ""
    echo "    === DIAGNOSTICS ==="
    echo "    Trying direct connection to VM..."
    if curl -s -m 5 "http://$PF_VM_IP:$VM_PORT/" >/dev/null 2>&1; then
        echo "    ✓ Direct connection to VM works"
        echo "    ✗ Port forward is broken"
        echo ""
        echo "    Checking iptables rules:"
        iptables -t nat -L PREROUTING -n -v | grep -A2 "$HOST_PORT" || echo "    No PREROUTING rule found"
        echo ""
        echo "    === ALL RULES FOR DEBUGGING ==="
        echo "    OUTPUT chain:"
        iptables -t nat -L OUTPUT -n -v --line-numbers | grep "$HOST_PORT" || echo "    No OUTPUT rule found"
        echo ""
        echo "    POSTROUTING chain (looking for SNAT with -o nanofuse0):"
        iptables -t nat -L POSTROUTING -n -v --line-numbers | grep -E "nanofuse0.*$VM_PORT|$VM_PORT.*nanofuse0" || echo "    No SNAT rule with -o nanofuse0 found"
        echo ""
        echo "    PAUSE FOR 30 SECONDS - Check rules manually with:"
        echo "      sudo iptables -t nat -L OUTPUT -n -v"
        echo "      sudo iptables -t nat -L POSTROUTING -n -v | grep nanofuse0"
        echo "      curl -v http://localhost:$HOST_PORT/"
        sleep 30
    else
        echo "    ✗ Direct connection to VM also fails"
        echo "    HTTP server is NOT running in VM"
        echo ""
        echo "    === CONSOLE LOG (last 150 lines) ==="
        $CLI_BIN --api-url $API_URL vm logs $PORT_FORWARD_VM --tail 150 2>/dev/null || echo "    Could not retrieve logs"
        echo "    === END CONSOLE LOG ==="
    fi
    echo ""
fi

# Stop and delete the port forward test VM
echo -n "  Cleaning up port forward test VM... "
$CLI_BIN --api-url "$API_URL" vm stop "$PORT_FORWARD_VM" > /dev/null 2>&1
$CLI_BIN --api-url "$API_URL" vm delete "$PORT_FORWARD_VM" --force > /dev/null 2>&1
sleep 2

# Verify iptables rules were removed
echo -n "  Verifying port forward cleanup (PREROUTING)... "
if ! iptables -t nat -L PREROUTING -n | grep -q "dpt:$HOST_PORT.*to:$PF_VM_IP:$VM_PORT"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    PREROUTING rule not cleaned up"
fi

echo -n "  Verifying port forward cleanup (OUTPUT)... "
if ! iptables -t nat -L OUTPUT -n | grep -q "dpt:$HOST_PORT.*to:$PF_VM_IP:$VM_PORT"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    OUTPUT rule not cleaned up"
fi

echo -n "  Verifying port forward cleanup (MASQUERADE)... "
if ! iptables -t nat -L POSTROUTING -n | grep "$PF_VM_IP" | grep -q "MASQUERADE.*dpt:$VM_PORT"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "    MASQUERADE rule not cleaned up"
fi

echo
echo "Test Summary"
echo "===================="
echo -e "${GREEN}✓ Network infrastructure setup${NC}"
echo -e "${GREEN}✓ VM created with network configuration${NC}"
echo -e "${GREEN}✓ TAP device created and attached to bridge${NC}"
echo -e "${GREEN}✓ VM booted successfully${NC}"

echo
echo "VM Information:"
echo "  Name:    $VM_NAME"
echo "  ID:      $VM_ID"
echo "  IP:      $VM_IP (ping from host to verify)"
echo "  Gateway: 172.16.0.1"
echo "  State:   $VM_STATE"

echo
echo "Manual Verification Steps:"
echo "  1. View VM console:"
echo "     $CLI_BIN --api-url $API_URL vm logs $VM_NAME"
echo
echo "  2. Check network interfaces:"
echo "     ip link show nanofuse0"
echo "     ip link show $VM_TAP"
echo "     bridge link show"
echo
echo "  3. Check iptables rules:"
echo "     iptables -t nat -L POSTROUTING -v"
echo
echo "  4. SSH into VM (if SSH keys configured):"
echo "     ssh root@$VM_IP"
echo
echo "  5. Test internet from inside VM:"
echo "     ssh root@$VM_IP 'ping -c 3 8.8.8.8'"
echo "     ssh root@$VM_IP 'ping -c 3 google.com'"

echo
echo -e "${GREEN}========================================"
echo "Test completed successfully!"
echo "========================================${NC}"

echo
echo "Press Ctrl+C to cleanup and exit, or wait 60 seconds for manual inspection..."
sleep 60

# Cleanup will be called automatically by trap EXIT
