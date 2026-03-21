#!/bin/bash
# Debug-Friendly Network E2E Script for NanoFuse
# - Boots a VM and keeps it running
# - Injects your SSH public key into the rootfs copy
# - Sets up port-forwarding to SSH (host:2222 -> vm:22)
# - Opens an SSH session (optional) so you can debug inside

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================"
echo "NanoFuse Network Debug (SSH)"
echo "========================================"
echo

if [ "${EUID}" -ne 0 ]; then
  echo -e "${RED}ERROR: Run as root (needs loop mounts and iptables)${NC}"
  echo "Try: sudo $0 [options]"
  exit 1
fi

# Defaults (override via env or flags)
DAEMON_BIN="./bin/nanofused"
CLI_BIN="./bin/nanofuse"
REGISTER_BIN="./bin/register-local-image"
CONFIG_FILE="./config.dev.yaml"
API_URL="http://127.0.0.1:8080"
DEFAULT_IMAGE_TAG="nanofuse-base:latest"
VM_NAME="debug-net-vm"
DATA_DIR="/tmp/nanofuse-debug"
# DB path will be read from config.dev.yaml to match the daemon
DB_PATH=""
HOST_SSH_PORT=2222
GUEST_SSH_PORT=22
AUTO_SSH=1
CLEANUP_ON_EXIT=0
PUBKEY_FILE=""
DEBUG_DUMP_DIR=""

usage() {
  cat <<EOF
Usage: sudo $0 [options]

Options:
  --name NAME            VM name (default: ${VM_NAME})
  --host-port PORT       Host SSH port to forward (default: ${HOST_SSH_PORT})
  --pubkey PATH          Public key to inject (default: autodetect from ~/.ssh)
  --no-ssh               Do not auto-open SSH session
  --cleanup-on-exit      Stop VM and daemon, remove temp data on exit
  --debug-dump DIR       Collect logs and state into DIR
  -h, --help             Show this help

This script:
  - Copies base rootfs and injects your SSH public key for root login
  - Starts nanofused and registers the local image copy
  - Creates VM with port-forward: HOST:${HOST_SSH_PORT} -> VM:${GUEST_SSH_PORT}
  - Waits for SSH availability and optionally opens SSH
  - Leaves VM running so you can investigate issues
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --name) VM_NAME="$2"; shift 2;;
    --host-port) HOST_SSH_PORT="$2"; shift 2;;
    --pubkey) PUBKEY_FILE="$2"; shift 2;;
    --no-ssh) AUTO_SSH=0; shift;;
    --cleanup-on-exit) CLEANUP_ON_EXIT=1; shift;;
    --debug-dump) DEBUG_DUMP_DIR="$2"; shift 2;;
    -h|--help) usage; exit 0;;
    *) echo -e "${RED}Unknown option: $1${NC}"; usage; exit 1;;
  esac
done

# Discover the real invoking user for ownership and key lookup
if [ -n "${SUDO_USER:-}" ]; then
  REAL_USER="$SUDO_USER"
  REAL_UID=$(id -u "$SUDO_USER")
  REAL_GID=$(id -g "$SUDO_USER")
else
  REAL_USER="$USER"
  REAL_UID="$UID"
  REAL_GID=$(id -g)
fi

echo "Running as root; invoking user: $REAL_USER"

cleanup() {
  echo
  echo -e "${YELLOW}Cleanup requested...${NC}"
  if [ -f /tmp/nanofused-debug.pid ]; then
    DAEMON_PID=$(cat /tmp/nanofused-debug.pid)
    if kill -0 "$DAEMON_PID" 2>/dev/null; then
      echo "Stopping daemon (PID $DAEMON_PID)..."
      kill "$DAEMON_PID" 2>/dev/null || true
      sleep 2
      kill -9 "$DAEMON_PID" 2>/dev/null || true
    fi
    rm -f /tmp/nanofused-debug.pid
  fi

  if [ -n "${PRIMARY_IFACE:-}" ]; then
    iptables -t nat -D POSTROUTING -s 172.16.0.0/24 -o "$PRIMARY_IFACE" -j MASQUERADE 2>/dev/null || true
    iptables -D FORWARD -i nanofuse0 -o "$PRIMARY_IFACE" -j ACCEPT 2>/dev/null || true
    iptables -D FORWARD -i "$PRIMARY_IFACE" -o nanofuse0 -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true
  fi

  if ip link show nanofuse0 >/dev/null 2>&1; then
    ip link set nanofuse0 down 2>/dev/null || true
    ip link delete nanofuse0 2>/dev/null || true
  fi

  if [ $CLEANUP_ON_EXIT -eq 1 ]; then
    echo "Stopping and deleting VM $VM_NAME (if present)..."
    $CLI_BIN --api-url "$API_URL" vm stop "$VM_NAME" >/dev/null 2>&1 || true
    $CLI_BIN --api-url "$API_URL" vm delete "$VM_NAME" --force >/dev/null 2>&1 || true
    echo "Removing $DATA_DIR..."
    rm -rf "$DATA_DIR"
  else
    echo -e "${YELLOW}Note: --cleanup-on-exit not set; keeping data in $DATA_DIR${NC}"
  fi
  echo -e "${GREEN}Cleanup complete${NC}"
}

if [ $CLEANUP_ON_EXIT -eq 1 ]; then
  trap cleanup EXIT
fi

echo -e "${YELLOW}[1/8] Checking prerequisites...${NC}"
for tool in jq ip iptables bridge sqlite3 curl losetup mount umount ssh; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo -e "${RED}Missing required tool: $tool${NC}"
    exit 1
  fi
done

# Ensure binaries exist; if not, try to build via mage (like the E2E script)
if [ ! -x "$DAEMON_BIN" ] || [ ! -x "$CLI_BIN" ] || [ ! -x "$REGISTER_BIN" ]; then
  echo -e "${YELLOW}Binaries missing; attempting to build with mage...${NC}"
  # Find go
  GO_BIN=$(command -v go || true)
  if [ -z "$GO_BIN" ]; then
    for c in /usr/local/go/bin/go /usr/bin/go "/home/$REAL_USER/go/bin/go" "/home/$REAL_USER/.local/go/bin/go"; do
      [ -x "$c" ] && GO_BIN="$c" && break
    done
  fi
  if [ -z "$GO_BIN" ]; then
    echo -e "${RED}Go not found; cannot build${NC}"; exit 1
  fi
  GOPATH=$(sudo -u "$REAL_USER" "$GO_BIN" env GOPATH 2>/dev/null || echo "/home/$REAL_USER/go")
  MAGE_BIN="$GOPATH/bin/mage"
  if [ ! -x "$MAGE_BIN" ]; then
    echo -e "${YELLOW}Installing mage...${NC}"
    sudo -u "$REAL_USER" "$GO_BIN" install github.com/magefile/mage@latest
  fi
  sudo -u "$REAL_USER" env PATH="$(dirname "$GO_BIN"):$PATH" "$MAGE_BIN" all
fi

for b in "$DAEMON_BIN" "$CLI_BIN" "$REGISTER_BIN"; do
  [ -x "$b" ] || chmod +x "$b"
done
echo -e "${GREEN}✓ Binaries ready${NC}"

if [ ! -f "$CONFIG_FILE" ]; then
  echo -e "${RED}Config file not found: $CONFIG_FILE${NC}"; exit 1
fi

# Derive DB path from config so we register into the daemon's database
DB_PATH=$(grep -E '^[[:space:]]*database:' "$CONFIG_FILE" | head -1 | sed -E "s/^[[:space:]]*database:[[:space:]]*//; s/[\"']//g")
if [ -z "$DB_PATH" ]; then
  echo -e "${YELLOW}Could not parse database path from $CONFIG_FILE; defaulting to /tmp/nanofuse/nanofuse.db${NC}"
  DB_PATH="/tmp/nanofuse/nanofuse.db"
fi

if [ ! -f "images/base/build/rootfs.ext4" ] || [ ! -f "images/base/build/vmlinux" ]; then
  echo -e "${RED}Base image artifacts missing. Build with: (cd images/base && sudo ./build.sh)${NC}"
  exit 1
fi

# Validate Firecracker binary and KVM availability early
FC_BIN=$(grep -E '^[[:space:]]*binary_path:' "$CONFIG_FILE" | head -1 | sed -E "s/^[[:space:]]*binary_path:[[:space:]]*//; s/[\"']//g")
if [ -z "$FC_BIN" ] || [ ! -x "$FC_BIN" ]; then
  echo -e "${RED}Firecracker binary not found or not executable at path from config: '${FC_BIN:-<empty>}'${NC}"
  echo "Edit $CONFIG_FILE to point to your firecracker binary (e.g. /usr/bin/firecracker)."
  exit 1
fi
echo -n "Firecracker version: "
"$FC_BIN" --version || true

if [ ! -e /dev/kvm ]; then
  echo -e "${RED}/dev/kvm not present. KVM is required for Firecracker.${NC}"
  echo "Load KVM modules (host dependent): 'modprobe kvm' and 'modprobe kvm_intel' or 'kvm_amd'"
  exit 1
fi
ls -l /dev/kvm || true
lsmod | grep -E '(^kvm\b|kvm_(intel|amd))' || echo -e "${YELLOW}KVM modules not listed by lsmod; verify virtualization support is enabled${NC}"

echo -e "${YELLOW}[2/8] Preparing working image and injecting SSH key...${NC}"
rm -rf "$DATA_DIR"
mkdir -p "$DATA_DIR/images/nanofuse-base/latest"
cp images/base/build/rootfs.ext4 "$DATA_DIR/images/nanofuse-base/latest/"
cp images/base/build/vmlinux "$DATA_DIR/images/nanofuse-base/latest/"
cp images/base/build/manifest.json "$DATA_DIR/images/nanofuse-base/latest/" || true

# Detect a public key if none provided
if [ -z "$PUBKEY_FILE" ]; then
  for k in "/home/$REAL_USER/.ssh/id_ed25519.pub" "/home/$REAL_USER/.ssh/id_rsa.pub"; do
    if [ -f "$k" ]; then PUBKEY_FILE="$k"; break; fi
  done
fi

if [ ! -f "$PUBKEY_FILE" ]; then
  echo -e "${RED}No public key found. Specify with --pubkey PATH${NC}"
  exit 1
fi

ROOTFS_COPY="$DATA_DIR/images/nanofuse-base/latest/rootfs.ext4"
MNT_DIR="/mnt/nanofuse-rootfs-$$"
LOOPDEV=""
mkdir -p "$MNT_DIR"

set +e
LOOPDEV=$(losetup -f)
losetup "$LOOPDEV" "$ROOTFS_COPY"
RET=$?
set -e
if [ $RET -ne 0 ]; then
  echo -e "${RED}Failed to setup loop device for $ROOTFS_COPY${NC}"
  exit 1
fi

mount "$LOOPDEV" "$MNT_DIR"
trap 'umount "$MNT_DIR" 2>/dev/null || true; losetup -d "$LOOPDEV" 2>/dev/null || true; rmdir "$MNT_DIR" 2>/dev/null || true' EXIT

mkdir -p "$MNT_DIR/root/.ssh"
chmod 700 "$MNT_DIR/root/.ssh"
cat "$PUBKEY_FILE" > "$MNT_DIR/root/.ssh/authorized_keys"
chmod 600 "$MNT_DIR/root/.ssh/authorized_keys"

# Ensure the primary NIC is named eth0 so kernel ip= configuration applies
mkdir -p "$MNT_DIR/etc/systemd/network"
cat > "$MNT_DIR/etc/systemd/network/10-rename-eth0.link" <<'EOF'
[Match]
OriginalName=en*

[Link]
Name=eth0
MACAddressPolicy=persistent
EOF
chmod 644 "$MNT_DIR/etc/systemd/network/10-rename-eth0.link"

sync
umount "$MNT_DIR"
losetup -d "$LOOPDEV"
rmdir "$MNT_DIR"
trap - EXIT

chown -R "$REAL_UID:$REAL_GID" "$DATA_DIR" || true
chmod -R u+rw "$DATA_DIR"
chmod 664 "$ROOTFS_COPY" 2>/dev/null || true

echo -e "${GREEN}✓ SSH key injected for root login${NC}"

echo -e "${YELLOW}[3/8] Starting daemon...${NC}"
pkill -f "nanofused.*config.dev.yaml" 2>/dev/null || true
"$DAEMON_BIN" --config "$CONFIG_FILE" > /tmp/nanofuse-daemon-debug.log 2>&1 &
DAEMON_PID=$!
echo $DAEMON_PID > /tmp/nanofused-debug.pid
echo "Daemon PID: $DAEMON_PID"

echo -n "Waiting for API"
for i in {1..30}; do
  sleep 1; echo -n "."
  if ! kill -0 "$DAEMON_PID" 2>/dev/null; then
    echo; echo -e "${RED}Daemon crashed during startup${NC}"; tail -n 200 /tmp/nanofuse-daemon-debug.log; exit 1
  fi
  if curl -sf "$API_URL/health" >/dev/null; then echo; break; fi
  [ $i -eq 30 ] && { echo; echo -e "${RED}API didn't come ready${NC}"; exit 1; }
done
echo -e "${GREEN}✓ Daemon ready${NC}"

echo -e "${YELLOW}[4/8] Registering image (DB: $DB_PATH)...${NC}"
mkdir -p "$(dirname "$DB_PATH")"

ROOTFS_PATH="$DATA_DIR/images/nanofuse-base/latest/rootfs.ext4"
KERNEL_PATH="$DATA_DIR/images/nanofuse-base/latest/vmlinux"

# Decide which tag to use; prefer reusing latest if it already points to our paths
TAG_TO_USE="$DEFAULT_IMAGE_TAG"
EXIST_ROW=$(sqlite3 "$DB_PATH" "SELECT rootfs_path||'|'||kernel_path FROM images WHERE digest='$DEFAULT_IMAGE_TAG'" 2>/dev/null || true)
if [ -n "$EXIST_ROW" ]; then
  EXIST_ROOTFS="${EXIST_ROW%%|*}"
  EXIST_KERNEL="${EXIST_ROW##*|}"
  if [ "$EXIST_ROOTFS" = "$ROOTFS_PATH" ] && [ "$EXIST_KERNEL" = "$KERNEL_PATH" ]; then
    echo -e "${GREEN}✓ Image already registered and points to our debug copy (${DEFAULT_IMAGE_TAG})${NC}"
  else
    TAG_TO_USE="nanofuse-base:debug-$(date +%s)"
    echo -e "${YELLOW}Found existing ${DEFAULT_IMAGE_TAG} pointing elsewhere; using tag ${TAG_TO_USE}${NC}"
  fi
fi

# Register if the chosen tag doesn't exist yet
if ! sqlite3 "$DB_PATH" "SELECT 1 FROM images WHERE digest='$TAG_TO_USE'" | grep -q 1; then
  "$REGISTER_BIN" "$DB_PATH" "$TAG_TO_USE" "$ROOTFS_PATH" "$KERNEL_PATH"
  if ! sqlite3 "$DB_PATH" "SELECT digest FROM images WHERE digest='$TAG_TO_USE'" | grep -q .; then
    echo -e "${RED}Image registration failed for $TAG_TO_USE${NC}"; exit 1
  fi
  echo -e "${GREEN}✓ Image registered: $TAG_TO_USE${NC}"
else
  echo -e "${GREEN}✓ Using existing image: $TAG_TO_USE${NC}"
fi

echo -e "${YELLOW}[5/8] Creating VM with SSH port-forward (${HOST_SSH_PORT}:${GUEST_SSH_PORT}/tcp)...${NC}"
"$CLI_BIN" --api-url "$API_URL" vm delete "$VM_NAME" --force >/dev/null 2>&1 || true
"$CLI_BIN" --api-url "$API_URL" vm create "$TAG_TO_USE" "$VM_NAME" \
  --vcpus 2 --memory 512 \
  --port-forward "${HOST_SSH_PORT}:${GUEST_SSH_PORT}/tcp"

VM_INFO=$("$CLI_BIN" --api-url "$API_URL" vm inspect "$VM_NAME" --json)
VM_TAP=$(echo "$VM_INFO" | jq -r '.config.network.tap_device')
VM_IP=$(echo "$VM_INFO" | jq -r '.config.network.ip_address')
echo "TAP: $VM_TAP  IP: $VM_IP"

echo -e "${YELLOW}[6/8] Starting VM...${NC}"
"$CLI_BIN" --api-url "$API_URL" vm start "$VM_NAME"

echo -n "Waiting for VM state transition (created -> starting -> running)"
VM_INFO_JSON=$("$CLI_BIN" --api-url "$API_URL" vm inspect "$VM_NAME" --json)
VM_ID=$(echo "$VM_INFO_JSON" | jq -r .id)
VM_DIR="/tmp/nanofuse/vms/$VM_ID"
CONSOLE_LOG="$VM_DIR/console.log"
for i in {1..90}; do
  sleep 1; echo -n "."
  VM_INFO_JSON=$("$CLI_BIN" --api-url "$API_URL" vm inspect "$VM_NAME" --json || echo '{}')
  STATE=$(echo "$VM_INFO_JSON" | jq -r .state)

  if [ "$STATE" = "running" ]; then echo; echo -e "${GREEN}✓ VM running${NC}"; break; fi
  if [ "$STATE" = "failed" ]; then
    echo; echo -e "${RED}VM entered failed state${NC}"
    echo "Daemon log tail:"; tail -n 120 /tmp/nanofuse-daemon-debug.log || true
    [ -f "$CONSOLE_LOG" ] && { echo "Console log tail:"; tail -n 120 "$CONSOLE_LOG"; }
    exit 1
  fi
  if [ $i -eq 45 ]; then
    echo; echo -e "${YELLOW}Halfway: current state: ${STATE:-unknown}${NC}"
    [ -f "$CONSOLE_LOG" ] && { echo "Console log (last 60):"; tail -n 60 "$CONSOLE_LOG"; }
  fi
  if [ $i -eq 90 ]; then
    echo; echo -e "${RED}Timeout: VM didn't reach running (last state: ${STATE:-unknown})${NC}"
    [ -f "$CONSOLE_LOG" ] && { echo "Console log tail:"; tail -n 120 "$CONSOLE_LOG"; }
    tail -n 200 /tmp/nanofuse-daemon-debug.log || true
    exit 1
  fi
done

# Quick direct connectivity checks (bypass port-forward) to diagnose guest network
echo -n "Pinging guest $VM_IP ... "
if ping -c 1 -W 3 "$VM_IP" >/dev/null 2>&1; then
  echo -e "${GREEN}ok${NC}"
else
  echo -e "${RED}failed${NC}"
  echo "Guest may not have configured its IP. Kernel args may not match interface name."
  echo "Console tail:"; [ -f "$CONSOLE_LOG" ] && tail -n 80 "$CONSOLE_LOG" || true
fi

echo -n "Checking guest SSH port (22) directly ... "
if command -v nc >/dev/null 2>&1 && nc -vz -w 2 "$VM_IP" 22 2>/dev/null; then
  echo -e "${GREEN}open${NC}"
else
  echo -e "${YELLOW}closed/unreachable${NC}"
fi

echo -n "Checking guest HTTP port (8080) directly ... "
if curl -4 -s -m 2 "http://$VM_IP:8080/" >/dev/null 2>&1; then
  echo -e "${GREEN}responding${NC}"
else
  echo -e "${YELLOW}no response${NC}"
fi

echo -n "Waiting for SSH on 127.0.0.1:${HOST_SSH_PORT} (forcing IPv4)"
for i in {1..60}; do
  sleep 1; echo -n "."
  if ssh -4 -o BatchMode=yes -o StrictHostKeyChecking=no -p "$HOST_SSH_PORT" root@127.0.0.1 true 2>/dev/null; then
    echo; echo -e "${GREEN}✓ SSH is ready${NC}"; break
  fi
  [ $i -eq 60 ] && { echo; echo -e "${YELLOW}SSH not ready yet; you can still try manually${NC}"; }
done

echo
echo "Debugging tips:"
echo "  - VM logs: $CLI_BIN --api-url $API_URL vm logs $VM_NAME | tail -n 100"
echo "  - Bridge:  ip link show nanofuse0 && bridge link show"
echo "  - Rules:   iptables -t nat -L -n -v | sed -n '1,120p'"
echo "  - SSH:     ssh -4 -o StrictHostKeyChecking=no -p $HOST_SSH_PORT root@127.0.0.1"
echo "  - Direct:   ssh -o StrictHostKeyChecking=no root@$VM_IP (if route exists)"
echo

if [ -n "$DEBUG_DUMP_DIR" ]; then
  echo -e "${YELLOW}Collecting debug bundle in $DEBUG_DUMP_DIR ...${NC}"
  mkdir -p "$DEBUG_DUMP_DIR"
  {
    echo "===== VM INSPECT =====";
    "$CLI_BIN" --api-url "$API_URL" vm inspect "$VM_NAME" --json || true;
    echo;
    echo "===== VM LOGS (last 300) =====";
    "$CLI_BIN" --api-url "$API_URL" vm logs "$VM_NAME" 2>/dev/null | tail -n 300 || true;
    echo;
    echo "===== DAEMON LOG TAIL =====";
    tail -n 300 /tmp/nanofuse-daemon-debug.log 2>/dev/null || true;
    echo;
    echo "===== HOST NET =====";
    ip a; echo; ip r; echo; iptables -t nat -L -n -v;
  } > "$DEBUG_DUMP_DIR/debug.txt" 2>&1

  if [ -n "${VM_DIR:-}" ] && [ -d "$VM_DIR" ]; then
    cp -f "$VM_DIR/config.json" "$DEBUG_DUMP_DIR/" 2>/dev/null || true
    cp -f "$VM_DIR/console.log" "$DEBUG_DUMP_DIR/" 2>/dev/null || true
  fi
  echo -e "${GREEN}✓ Debug bundle written to $DEBUG_DUMP_DIR${NC}"
fi

if [ $AUTO_SSH -eq 1 ]; then
  echo -e "${YELLOW}Opening SSH session (Ctrl+D to exit, VM stays running)...${NC}"
  ssh -4 -o StrictHostKeyChecking=no -p "$HOST_SSH_PORT" root@127.0.0.1 || true
fi

echo
echo -e "${GREEN}VM is running and ready for inspection${NC}"
echo "When finished debugging:"
echo "  - Stop VM:   $CLI_BIN --api-url $API_URL vm stop $VM_NAME"
echo "  - Delete VM: $CLI_BIN --api-url $API_URL vm delete $VM_NAME --force"
echo "  - Stop daemon (PID $(cat /tmp/nanofused-debug.pid 2>/dev/null || echo '?')): kill \\$(cat /tmp/nanofused-debug.pid)"
echo
if [ $CLEANUP_ON_EXIT -eq 1 ]; then
  echo -e "${YELLOW}--cleanup-on-exit set; press Ctrl+C to cleanup and exit...${NC}"
  while true; do sleep 1; done
else
  echo -e "${YELLOW}Leaving everything running for investigation. Script will now exit.${NC}"
  echo -e "${YELLOW}Note: Daemon logs at /tmp/nanofuse-daemon-debug.log${NC}"
fi
