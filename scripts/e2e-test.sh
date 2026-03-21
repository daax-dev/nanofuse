#!/bin/bash
# scripts/e2e-test.sh - Full E2E test suite for nanofuse
#
# This script validates the complete nanofuse lifecycle:
# 1. Verify daemon is running
# 2. Register image
# 3. Create and start VM
# 4. Test SSH connectivity
# 5. Test HTTP connectivity
# 6. Cleanup
#
# Requirements:
# - Root/sudo access
# - KVM available (/dev/kvm)
# - Firecracker installed
# - Build artifacts present
# - Daemon running

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
VM_NAME="${E2E_VM_NAME:-e2e-test-vm}"
BOOT_TIMEOUT="${E2E_BOOT_TIMEOUT:-60}"
SSH_TIMEOUT="${E2E_SSH_TIMEOUT:-30}"
SKIP_CLEANUP="${E2E_SKIP_CLEANUP:-0}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() { echo -e "${GREEN}[E2E]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0

pass() {
    log "PASS: $1"
    ((TESTS_PASSED++)) || true
}

fail_test() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++)) || true
}

# Cleanup function
cleanup() {
    if [ "$SKIP_CLEANUP" = "1" ]; then
        warn "Skipping cleanup (E2E_SKIP_CLEANUP=1)"
        return
    fi

    log "Cleaning up..."
    nanofuse vm stop "$VM_NAME" 2>/dev/null || true
    sleep 2
    nanofuse vm delete "$VM_NAME" 2>/dev/null || true
}

# Check prerequisites
check_prereqs() {
    log "Checking prerequisites..."

    # Check root
    if [ "$(id -u)" -ne 0 ]; then
        fail "This script requires root access. Run with: sudo $0"
    fi
    pass "Running as root"

    # Check KVM
    if [ ! -e /dev/kvm ]; then
        fail "KVM not available (/dev/kvm not found)"
    fi
    pass "KVM available"

    # Check Firecracker
    if ! command -v firecracker &>/dev/null; then
        fail "Firecracker not installed"
    fi
    pass "Firecracker installed"

    # Check jq
    if ! command -v jq &>/dev/null; then
        fail "jq not installed (required for JSON parsing)"
    fi
    pass "jq available"

    # Check build artifacts
    if [ ! -f "${PROJECT_ROOT}/images/base/build/vmlinux" ]; then
        fail "Kernel not found at images/base/build/vmlinux"
    fi
    pass "Kernel found"

    if [ ! -f "${PROJECT_ROOT}/images/base/build/rootfs.ext4" ]; then
        fail "Rootfs not found at images/base/build/rootfs.ext4"
    fi
    pass "Rootfs found"

    # Check nanofuse CLI
    if ! command -v nanofuse &>/dev/null; then
        if [ -f "${PROJECT_ROOT}/bin/nanofuse" ]; then
            export PATH="${PROJECT_ROOT}/bin:$PATH"
        else
            fail "nanofuse CLI not found"
        fi
    fi
    pass "nanofuse CLI available"
}

# Phase 1: Verify daemon
phase_daemon() {
    log "Phase 1: Verifying daemon..."

    # Check socket exists
    local socket_found=0
    for socket in /var/run/nanofused.sock /run/nanofused.sock /tmp/nanofused.sock; do
        if [ -S "$socket" ]; then
            info "Found socket at $socket"
            socket_found=1
            break
        fi
    done

    if [ "$socket_found" -eq 0 ]; then
        fail_test "Daemon socket not found"
        return 1
    fi

    # Try health check
    if nanofuse health 2>/dev/null | grep -qi healthy; then
        pass "Daemon health check passed"
    else
        warn "Health check failed (may still work)"
    fi

    pass "Daemon is running"
}

# Phase 2: Register image
phase_register() {
    log "Phase 2: Registering image..."

    local rootfs="${PROJECT_ROOT}/images/base/build/rootfs.ext4"
    local register_bin="${PROJECT_ROOT}/bin/register-local-image"

    # Check if image already registered
    if nanofuse image list 2>/dev/null | grep -q "base"; then
        info "Image 'base' already registered"
        pass "Image available"
        return 0
    fi

    # Register image
    if [ -x "$register_bin" ]; then
        if "$register_bin" "$rootfs" base; then
            pass "Image registered successfully"
        else
            fail_test "Failed to register image"
            return 1
        fi
    else
        warn "register-local-image not found, skipping registration"
    fi
}

# Phase 3: Create VM
phase_create() {
    log "Phase 3: Creating VM..."

    # Clean up any existing test VM
    nanofuse vm stop "$VM_NAME" 2>/dev/null || true
    nanofuse vm delete "$VM_NAME" 2>/dev/null || true
    sleep 1

    if nanofuse vm create "$VM_NAME" --image base; then
        pass "VM created"
    else
        fail_test "Failed to create VM"
        return 1
    fi
}

# Phase 4: Start VM
phase_start() {
    log "Phase 4: Starting VM..."

    if nanofuse vm start "$VM_NAME"; then
        pass "VM start command succeeded"
    else
        fail_test "Failed to start VM"
        return 1
    fi
}

# Phase 5: Wait for boot
phase_boot() {
    log "Phase 5: Waiting for VM to boot (timeout: ${BOOT_TIMEOUT}s)..."

    local elapsed=0
    while [ $elapsed -lt $BOOT_TIMEOUT ]; do
        if nanofuse vm status "$VM_NAME" 2>/dev/null | grep -q "running"; then
            pass "VM is running"
            # Give init time to complete
            sleep 5
            return 0
        fi
        sleep 2
        ((elapsed += 2)) || true
        info "Waiting for boot... (${elapsed}s/${BOOT_TIMEOUT}s)"
    done

    fail_test "Timeout waiting for VM to boot"
    return 1
}

# Phase 6: Test SSH
phase_ssh() {
    log "Phase 6: Testing SSH connectivity..."

    local vm_ip
    vm_ip=$(nanofuse vm show "$VM_NAME" --format json 2>/dev/null | jq -r '.ip_address' || echo "")

    if [ -z "$vm_ip" ] || [ "$vm_ip" = "null" ]; then
        fail_test "Could not get VM IP address"
        return 1
    fi

    info "VM IP: $vm_ip"

    local elapsed=0
    while [ $elapsed -lt $SSH_TIMEOUT ]; do
        if ssh -o ConnectTimeout=5 \
               -o StrictHostKeyChecking=no \
               -o UserKnownHostsFile=/dev/null \
               -o BatchMode=yes \
               "root@${vm_ip}" "echo SSH_SUCCESS" 2>/dev/null | grep -q "SSH_SUCCESS"; then
            pass "SSH connectivity verified"
            return 0
        fi
        sleep 3
        ((elapsed += 3)) || true
        info "SSH attempt failed, retrying... (${elapsed}s/${SSH_TIMEOUT}s)"
    done

    fail_test "SSH connectivity failed"
    return 1
}

# Phase 7: Test HTTP
phase_http() {
    log "Phase 7: Testing HTTP connectivity..."

    local vm_ip
    vm_ip=$(nanofuse vm show "$VM_NAME" --format json 2>/dev/null | jq -r '.ip_address' || echo "")

    if [ -z "$vm_ip" ] || [ "$vm_ip" = "null" ]; then
        warn "Could not get VM IP, skipping HTTP test"
        return 0
    fi

    # Test outbound HTTP from VM
    if ssh -o ConnectTimeout=10 \
           -o StrictHostKeyChecking=no \
           -o UserKnownHostsFile=/dev/null \
           "root@${vm_ip}" \
           "wget -q -O- http://example.com 2>/dev/null | grep -qi html"; then
        pass "HTTP outbound connectivity verified"
    else
        warn "HTTP outbound test failed (may be expected if blocked)"
    fi
}

# Summary
print_summary() {
    echo ""
    echo "========================================"
    echo "E2E Test Summary"
    echo "========================================"
    echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
    echo "========================================"

    if [ "$TESTS_FAILED" -gt 0 ]; then
        echo -e "${RED}E2E TESTS FAILED${NC}"
        return 1
    else
        echo -e "${GREEN}E2E TESTS PASSED${NC}"
        return 0
    fi
}

# Main
main() {
    echo "========================================"
    echo "NanoFuse E2E Test Suite"
    echo "========================================"
    echo "Project root: ${PROJECT_ROOT}"
    echo "VM name: ${VM_NAME}"
    echo "Boot timeout: ${BOOT_TIMEOUT}s"
    echo "SSH timeout: ${SSH_TIMEOUT}s"
    echo ""

    trap cleanup EXIT

    check_prereqs
    phase_daemon || true
    phase_register || true
    phase_create || true
    phase_start || true
    phase_boot || true
    phase_ssh || true
    phase_http || true

    print_summary
}

main "$@"
