#!/bin/bash
# Complete dev rebuild and test script
# Usage: sudo ./scripts/dev-rebuild.sh [--test-ssh]
#
# This script:
# 1. Stops nanofused service
# 2. Rebuilds all binaries (mage)
# 3. Installs to /usr/local/bin
# 4. Rebuilds todo-app image (no cache)
# 5. Starts nanofused service
# 6. Optionally creates test VM and tests SSH

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
# Use real user's SSH key, not root's
REAL_USER_HOME=$(eval echo "~${SUDO_USER:-$USER}")
SSH_KEY="${SSH_KEY:-$REAL_USER_HOME/.ssh/id_ed25519.pub}"
TEST_SSH=false

# Parse args
while [[ $# -gt 0 ]]; do
    case $1 in
        --test-ssh)
            TEST_SSH=true
            shift
            ;;
        --ssh-key)
            SSH_KEY="$2"
            shift 2
            ;;
        *)
            echo "Usage: $0 [--test-ssh] [--ssh-key <path>]"
            exit 1
            ;;
    esac
done

# Must run as root
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root (sudo)"
    exit 1
fi

# Get the actual user (not root)
REAL_USER="${SUDO_USER:-$USER}"
REAL_HOME=$(eval echo "~$REAL_USER")

echo "=============================================="
echo "NanoFuse Dev Rebuild"
echo "=============================================="
echo "Project: $PROJECT_DIR"
echo "User: $REAL_USER"
echo ""

# Step 1: Stop service
echo "1. Stopping nanofused service..."
systemctl stop nanofused 2>/dev/null || true
echo "   ✓ Service stopped"

# Step 2: Clean and rebuild binaries
echo ""
echo "2. Rebuilding binaries..."
cd "$PROJECT_DIR"
# Run mage as real user with proper PATH (Go needs to be in PATH)
sudo -u "$REAL_USER" bash -c "export PATH=/usr/local/go/bin:\$HOME/go/bin:\$PATH && cd $PROJECT_DIR && $REAL_HOME/go/bin/mage clean" 2>/dev/null || true
sudo -u "$REAL_USER" bash -c "export PATH=/usr/local/go/bin:\$HOME/go/bin:\$PATH && cd $PROJECT_DIR && $REAL_HOME/go/bin/mage all"
echo "   ✓ Binaries built"

# Step 3: Install binaries
echo ""
echo "3. Installing binaries to /usr/local/bin..."
cp "$PROJECT_DIR/bin/nanofuse" /usr/local/bin/nanofuse
cp "$PROJECT_DIR/bin/nanofused" /usr/local/bin/nanofused
cp "$PROJECT_DIR/bin/register-local-image" /usr/local/bin/register-local-image
chmod +x /usr/local/bin/nanofuse /usr/local/bin/nanofused /usr/local/bin/register-local-image
echo "   ✓ Binaries installed"

# Step 4: Rebuild todo-app image (no cache)
echo ""
echo "4. Rebuilding todo-app image (no cache)..."
cd "$PROJECT_DIR/examples/todo-app"

# Clean previous build
rm -rf output/

# Build script handles docker build, export, ext4 creation, and registration
./build-nanofuse-image.sh --no-cache
echo "   ✓ Image rebuilt and registered"

# Step 5: Start service
echo ""
echo "5. Starting nanofused service..."
systemctl start nanofused
sleep 2
if systemctl is-active --quiet nanofused; then
    echo "   ✓ Service started"
else
    echo "   ✗ Service failed to start"
    journalctl -u nanofused -n 20 --no-pager
    exit 1
fi

# Step 6: Optionally test SSH
if [[ "$TEST_SSH" == "true" ]]; then
    echo ""
    echo "6. Testing SSH key injection..."

    # Check SSH key exists
    if [[ ! -f "$SSH_KEY" ]]; then
        echo "   ✗ SSH key not found: $SSH_KEY"
        exit 1
    fi
    echo "   Using SSH key: $SSH_KEY"

    # Delete existing test VM
    nanofuse vm delete ssh-test -f 2>/dev/null || true

    # Create and start VM with SSH key
    echo "   Creating test VM..."
    nanofuse vm run ghcr.io/peregrinesummit/nanofuse/todo-app:latest ssh-test --ssh-key "$SSH_KEY"

    # Get VM IP
    sleep 3
    VM_IP=$(nanofuse vm status ssh-test --json 2>/dev/null | jq -r '.runtime.network_info.guest_ip // empty')

    if [[ -z "$VM_IP" ]]; then
        echo "   ✗ Could not get VM IP"
        exit 1
    fi
    echo "   VM IP: $VM_IP"

    # Verify sshkey in kernel args
    echo "   Checking kernel args..."
    VM_ID=$(nanofuse vm status ssh-test --json | jq -r '.id')
    BOOT_ARGS=$(cat "/var/lib/nanofuse/vms/$VM_ID/config.json" | jq -r '.["boot-source"].boot_args')

    if echo "$BOOT_ARGS" | grep -q "sshkey="; then
        echo "   ✓ SSH key in kernel args"
    else
        echo "   ✗ SSH key NOT in kernel args!"
        echo "   Boot args: $BOOT_ARGS"
        exit 1
    fi

    # Remove old host key (IP gets reused, host key changes each rebuild)
    ssh-keygen -R "$VM_IP" 2>/dev/null || true

    # Wait for VM to boot and SSH to be ready
    echo "   Waiting for SSH..."
    for i in {1..30}; do
        if ssh -o ConnectTimeout=2 -o StrictHostKeyChecking=no -o BatchMode=yes "root@$VM_IP" "echo ok" 2>/dev/null; then
            echo "   ✓ SSH connection successful!"

            # Test HTTP endpoints
            echo "   Testing HTTP endpoints..."
            HEALTH=$(curl -sf "http://$VM_IP/health" 2>/dev/null)
            if [[ -n "$HEALTH" ]]; then
                echo "   ✓ /health: $HEALTH"
            else
                echo "   ✗ /health failed"
            fi

            STATIC=$(curl -sf "http://$VM_IP/" 2>/dev/null | head -c 80)
            if [[ -n "$STATIC" ]]; then
                echo "   ✓ /: ${STATIC}..."
            else
                echo "   ✗ / (static) failed"
            fi

            echo ""
            echo "=============================================="
            echo "SUCCESS!"
            echo "  curl http://$VM_IP/health"
            echo "  curl http://$VM_IP/"
            echo "  ssh root@$VM_IP"
            echo "=============================================="
            exit 0
        fi
        sleep 1
    done

    echo "   ✗ SSH connection failed after 30 seconds"
    echo "   Check console logs: nanofuse vm logs ssh-test"
    exit 1
fi

echo ""
echo "=============================================="
echo "Rebuild Complete!"
echo "=============================================="
echo ""
echo "To test manually:"
echo "  sudo nanofuse vm run ghcr.io/peregrinesummit/nanofuse/todo-app:latest test-vm --ssh-key ~/.ssh/id_ed25519.pub"
echo ""
echo "Or run with --test-ssh to auto-test:"
echo "  sudo $0 --test-ssh"
echo ""
