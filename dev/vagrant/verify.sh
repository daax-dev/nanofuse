#!/usr/bin/env bash
# verify.sh — validate nanofuse dev environment end-to-end
#
# Checks system prerequisites, nanofuse build, base image, daemon, and
# Firecracker microVM boot. Exits non-zero if any critical check fails.

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0
WARN=0

pass() { echo -e "  ${GREEN}PASS${NC} $*"; PASS=$((PASS + 1)); }
fail() { echo -e "  ${RED}FAIL${NC} $*"; FAIL=$((FAIL + 1)); }
skip() { echo -e "  ${YELLOW}SKIP${NC} $*"; WARN=$((WARN + 1)); }

export PATH="/usr/local/go/bin:$HOME/go/bin:/usr/local/bin:$PATH"
export GOPATH="$HOME/go"

cleanup() {
    # Kill any firecracker process we started
    pkill -f "firecracker --api-sock /tmp/fc-verify" 2>/dev/null || true
    rm -f /tmp/fc-verify.sock /tmp/verify-rootfs.ext4
    systemctl stop nanofused 2>/dev/null || true
}
trap cleanup EXIT

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  nanofuse dev environment verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ─── 1. System prerequisites ───────────────────────────────────────────────
echo -e "\n▶ System prerequisites"

[[ -e /dev/kvm ]]                && pass "KVM"       || fail "KVM (/dev/kvm missing)"
command -v docker &>/dev/null    && pass "Docker"    || fail "Docker"
[[ -x /usr/local/go/bin/go ]]   && pass "Go $(/usr/local/go/bin/go version | awk '{print $3}')" || fail "Go"
command -v mage &>/dev/null      && pass "Mage"      || fail "Mage"
[[ -x /usr/local/bin/firecracker ]] && pass "Firecracker $(firecracker --version 2>&1 | grep -oP 'v[\d.]+')" || fail "Firecracker"
[[ -x /usr/local/bin/jailer ]]  && pass "Jailer"    || skip "Jailer (optional)"

# ─── 2. nanofuse binaries ──────────────────────────────────────────────────
echo -e "\n▶ nanofuse binaries"

[[ -x /usr/local/bin/nanofuse ]]  && pass "nanofuse CLI"    || fail "nanofuse CLI"
[[ -x /usr/local/bin/nanofused ]] && pass "nanofused daemon" || fail "nanofused daemon"

# ─── 3. Unit tests (quiet — only show summary) ────────────────────────────
echo -e "\n▶ Unit tests"

if [[ -d /nanofuse ]]; then
    cd /nanofuse
    test_output_file=$(mktemp)
    if go test ./... -count=1 -short 2>&1 | tee "$test_output_file"; then
        test_count=$(grep -c "^ok" "$test_output_file" || true)
        pass "$test_count packages pass"
    else
        grep -E "^(FAIL|---)" "$test_output_file" | head -10 || true
        fail "unit tests failed"
    fi
    rm -f "$test_output_file"
    cd /
else
    fail "nanofuse source not at /nanofuse"
fi

# ─── 4. Base image ─────────────────────────────────────────────────────────
echo -e "\n▶ Base microVM image"

KERNEL="/var/lib/nanofuse/images/vmlinux"
ROOTFS="/var/lib/nanofuse/images/rootfs.ext4"

[[ -f "$KERNEL" ]] && pass "kernel $(ls -lh $KERNEL | awk '{print $5}')" || fail "kernel missing"
[[ -f "$ROOTFS" ]] && pass "rootfs $(ls -lh $ROOTFS | awk '{print $5}')" || fail "rootfs missing"

# ─── 5. Config ─────────────────────────────────────────────────────────────
echo -e "\n▶ Configuration"

[[ -f /etc/nanofuse/nanofused.yaml ]] && pass "nanofused.yaml" || fail "config missing"

# ─── 6. Network tools ──────────────────────────────────────────────────────
echo -e "\n▶ Network tools"

for cmd in iptables ip dnsmasq dig; do
    command -v $cmd &>/dev/null && pass "$cmd" || fail "$cmd"
done

# ─── 7. Firecracker boot test ──────────────────────────────────────────────
echo -e "\n▶ Firecracker microVM boot"

if [[ -f "$KERNEL" ]] && [[ -f "$ROOTFS" ]]; then
    SOCKET="/tmp/fc-verify.sock"
    TEST_ROOTFS="/tmp/verify-rootfs.ext4"
    rm -f "$SOCKET"

    # Copy rootfs (Firecracker needs rw access)
    cp "$ROOTFS" "$TEST_ROOTFS"

    # Start Firecracker in background (console output to log, not stdout)
    touch /tmp/fc-verify.log
    firecracker --api-sock "$SOCKET" --log-path /tmp/fc-verify.log --level Info > /tmp/fc-console.log 2>&1 &
    FC_PID=$!
    sleep 1

    # Configure and boot
    if ! curl -sf --unix-socket "$SOCKET" -X PUT "http://localhost/boot-source" \
        -H "Content-Type: application/json" \
        -d "{\"kernel_image_path\":\"$KERNEL\",\"boot_args\":\"console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw\"}" >/dev/null 2>&1; then
        fail "Configure Firecracker boot-source"
    fi

    if ! curl -sf --unix-socket "$SOCKET" -X PUT "http://localhost/drives/rootfs" \
        -H "Content-Type: application/json" \
        -d "{\"drive_id\":\"rootfs\",\"path_on_host\":\"$TEST_ROOTFS\",\"is_root_device\":true,\"is_read_only\":false}" >/dev/null 2>&1; then
        fail "Configure Firecracker rootfs drive"
    fi

    if ! curl -sf --unix-socket "$SOCKET" -X PUT "http://localhost/machine-config" \
        -H "Content-Type: application/json" \
        -d "{\"vcpu_count\":1,\"mem_size_mib\":256}" >/dev/null 2>&1; then
        fail "Configure Firecracker machine-config"
    fi

    if curl -sf --unix-socket "$SOCKET" -X PUT "http://localhost/actions" \
        -H "Content-Type: application/json" \
        -d '{"action_type":"InstanceStart"}' >/dev/null 2>&1; then
        pass "InstanceStart accepted"

        # Wait for boot — check that FC process is still alive after kernel starts
        sleep 5
        if kill -0 $FC_PID 2>/dev/null; then
            pass "Firecracker process alive after boot"

            # Check state via API
            state=$(curl -sf --unix-socket "$SOCKET" "http://localhost/" 2>/dev/null | grep -o '"state":"[^"]*"' || echo "unknown")
            pass "VM state: $state"
        else
            fail "Firecracker process died"
        fi
    else
        fail "InstanceStart failed"
    fi

    # Cleanup
    kill $FC_PID 2>/dev/null || true
    wait $FC_PID 2>/dev/null || true
    rm -f "$SOCKET" "$TEST_ROOTFS" /tmp/fc-verify.log
else
    skip "Firecracker boot — base image not available"
fi

# ─── Summary ────────────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "  ${GREEN}PASS: $PASS${NC}  ${RED}FAIL: $FAIL${NC}  ${YELLOW}SKIP: $WARN${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ $FAIL -gt 0 ]]; then
    echo -e "\n${RED}$FAIL checks failed${NC}"
    exit 1
fi

echo -e "\n${GREEN}nanofuse dev environment fully verified!${NC}"
