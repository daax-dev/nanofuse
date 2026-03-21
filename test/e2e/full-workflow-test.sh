#!/bin/bash
# NanoFuse End-to-End Full Workflow Test
#
# Tests the complete pull-to-running cycle:
# 1. Authenticate to GHCR (optional, uses GITHUB_TOKEN)
# 2. Pull or use local image
# 3. Create and start VM
# 4. Validate services respond
# 5. Clean up resources
#
# Usage: ./test/e2e/full-workflow-test.sh [OPTIONS]
#
# Exit codes:
#   0 - All tests passed
#   1 - Test failure
#   2 - Invalid arguments or environment

set -euo pipefail

# Configuration
DEFAULT_IMAGE="todo-app:latest"
VM_NAME="e2e-test-vm"
TIMEOUT=90
VERBOSE=0
SKIP_CLEANUP=0

# Colors (disabled if not a terminal)
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Timing
START_TIME=$(date +%s)

show_help() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS]

End-to-end test for NanoFuse VM workflow: pull → create → start → verify → cleanup

Options:
  --image <name>        Image to test (default: $DEFAULT_IMAGE)
  --verbose             Show detailed output
  --skip-cleanup        Don't delete test VM on completion (for debugging)
  --timeout <seconds>   Total test timeout (default: $TIMEOUT)
  -h, --help            Show this help message

Environment:
  GITHUB_TOKEN          Optional, for GHCR authentication

Examples:
  $(basename "$0")
  $(basename "$0") --image ghcr.io/peregrinesummit/nanofuse/todo-app:latest
  $(basename "$0") --verbose --skip-cleanup
EOF
}

log() {
    local level="$1"
    shift
    local msg="$*"
    local elapsed=$(($(date +%s) - START_TIME))

    case "$level" in
        INFO)  echo -e "${BLUE}[${elapsed}s]${NC} $msg" ;;
        OK)    echo -e "${GREEN}[${elapsed}s] ✓${NC} $msg" ;;
        FAIL)  echo -e "${RED}[${elapsed}s] ✗${NC} $msg" ;;
        WARN)  echo -e "${YELLOW}[${elapsed}s] !${NC} $msg" ;;
        DEBUG) [[ $VERBOSE -eq 1 ]] && echo -e "[${elapsed}s] $msg" ;;
    esac
}

cleanup() {
    if [[ $SKIP_CLEANUP -eq 1 ]]; then
        log WARN "Skipping cleanup (--skip-cleanup). VM '$VM_NAME' left running."
        return 0
    fi

    log INFO "Cleaning up test VM..."
    sudo nanofuse vm delete "$VM_NAME" -f 2>/dev/null || true
}

fail() {
    log FAIL "$1"
    cleanup
    exit 1
}

# Parse arguments
IMAGE="$DEFAULT_IMAGE"
while [[ $# -gt 0 ]]; do
    case $1 in
        --image)
            IMAGE="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=1
            shift
            ;;
        --skip-cleanup)
            SKIP_CLEANUP=1
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            show_help
            exit 2
            ;;
    esac
done

# Must run as root for VM operations
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root (sudo)" >&2
    exit 2
fi

# Trap for cleanup on exit
trap cleanup EXIT

echo "========================================"
echo "NanoFuse E2E Full Workflow Test"
echo "========================================"
echo "Image: $IMAGE"
echo "Timeout: ${TIMEOUT}s"
echo "========================================"
echo ""

# Step 1: Ensure clean state (idempotent)
log INFO "Step 1/6: Ensuring clean state..."
sudo nanofuse vm delete "$VM_NAME" -f 2>/dev/null || true
log OK "Clean state ensured"

# Step 2: GHCR Authentication (if pulling from ghcr.io)
if [[ "$IMAGE" == ghcr.io/* ]]; then
    log INFO "Step 2/6: Authenticating to GHCR..."
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        if echo "$GITHUB_TOKEN" | docker login ghcr.io -u "x-access-token" --password-stdin >/dev/null 2>&1; then
            log OK "Authenticated to GHCR"
        else
            fail "GHCR authentication failed"
        fi
    else
        log WARN "No GITHUB_TOKEN set, skipping GHCR auth (may fail for private images)"
    fi
else
    log INFO "Step 2/6: Using local image, skipping GHCR auth"
    log OK "Local image mode"
fi

# Step 3: Verify image exists (or pull)
log INFO "Step 3/6: Checking image availability..."
if sudo nanofuse image list 2>/dev/null | grep -q "${IMAGE%:*}"; then
    log OK "Image available: $IMAGE"
else
    log WARN "Image not found locally, attempting to pull..."
    if [[ "$IMAGE" == ghcr.io/* ]]; then
        # For GHCR images, would need nanofuse image pull (if implemented)
        fail "Image not available and remote pull not yet implemented for: $IMAGE"
    else
        fail "Local image not found: $IMAGE"
    fi
fi

# Step 4: Create and start VM
log INFO "Step 4/6: Creating and starting VM..."
log DEBUG "Running: nanofuse vm run $IMAGE $VM_NAME"

if ! sudo nanofuse vm run "$IMAGE" "$VM_NAME" 2>&1; then
    fail "Failed to create/start VM"
fi

# Wait for VM to get IP
log DEBUG "Waiting for VM to get IP address..."
VM_IP=""
for i in {1..30}; do
    VM_IP=$(sudo nanofuse vm status "$VM_NAME" --json 2>/dev/null | jq -r '.runtime.network_info.guest_ip // empty' 2>/dev/null || echo "")
    if [[ -n "$VM_IP" ]] && [[ "$VM_IP" != "null" ]]; then
        break
    fi
    sleep 1
done

if [[ -z "$VM_IP" ]] || [[ "$VM_IP" == "null" ]]; then
    fail "VM failed to get IP address within 30 seconds"
fi

log OK "VM started with IP: $VM_IP"

# Step 5: Wait for services and validate
log INFO "Step 5/6: Validating services..."

# Calculate remaining time for service checks
ELAPSED=$(($(date +%s) - START_TIME))
REMAINING=$((TIMEOUT - ELAPSED - 10))  # Reserve 10s for cleanup
[[ $REMAINING -lt 10 ]] && REMAINING=10

log DEBUG "Service validation timeout: ${REMAINING}s"

# Poll for HTTP service
HTTP_OK=0
HEALTH_OK=0

for i in $(seq 1 $REMAINING); do
    # Check HTTP (static files)
    if [[ $HTTP_OK -eq 0 ]]; then
        if curl -sf --max-time 3 "http://${VM_IP}/" 2>/dev/null | grep -qi '<html>'; then
            HTTP_OK=1
            log OK "HTTP service responding (static files)"
        fi
    fi

    # Check health endpoint
    if [[ $HEALTH_OK -eq 0 ]]; then
        HEALTH_RESPONSE=$(curl -sf --max-time 3 "http://${VM_IP}/health" 2>/dev/null || echo "")
        if [[ -n "$HEALTH_RESPONSE" ]]; then
            HEALTH_OK=1
            log OK "Health endpoint responding"
            log DEBUG "Health response: $HEALTH_RESPONSE"
        fi
    fi

    # Both checks passed
    if [[ $HTTP_OK -eq 1 ]] && [[ $HEALTH_OK -eq 1 ]]; then
        break
    fi

    sleep 1
done

# Verify all checks passed
if [[ $HTTP_OK -eq 0 ]]; then
    fail "HTTP service (port 80) did not respond within timeout"
fi

if [[ $HEALTH_OK -eq 0 ]]; then
    fail "Health endpoint did not respond within timeout"
fi

log OK "All service validations passed"

# Step 6: Cleanup (handled by trap, but log it)
log INFO "Step 6/6: Cleanup..."
# Cleanup happens via trap

# Final summary
TOTAL_TIME=$(($(date +%s) - START_TIME))
echo ""
echo "========================================"
log OK "All tests passed!"
echo "Total time: ${TOTAL_TIME}s"
echo "========================================"

# Disable trap since we're exiting successfully
trap - EXIT
cleanup

exit 0
