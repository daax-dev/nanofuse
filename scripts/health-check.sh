#!/bin/bash
# NanoFuse VM Health Check Script
# Validates VM state, network, and services
#
# Usage: ./scripts/health-check.sh <vm-name> [--json] [--timeout <seconds>]
#
# Exit codes:
#   0 - All checks passed
#   1 - One or more checks failed

set -euo pipefail

# Colors (disabled if not a terminal or --json mode)
if [[ -t 1 ]] && [[ "${JSON_OUTPUT:-}" != "1" ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Default configuration
TIMEOUT=30
JSON_OUTPUT=0
VM_NAME=""

# Parse arguments
show_help() {
    cat << EOF
Usage: $(basename "$0") <vm-name> [OPTIONS]

Validates VM health: state, network connectivity, and service endpoints.

Options:
  --json              Output results as JSON
  --timeout <sec>     Timeout for health checks (default: 30)
  -h, --help          Show this help message

Checks performed:
  1. VM state (must be "running")
  2. Network connectivity (VM has IP, responds to ping)
  3. HTTP service (port 80 responds with content)
  4. Health endpoint (port 80 /health responds)

Exit codes:
  0 - All checks passed
  1 - One or more checks failed

Examples:
  $(basename "$0") my-vm
  $(basename "$0") my-vm --json
  $(basename "$0") my-vm --timeout 60
EOF
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --json)
            JSON_OUTPUT=1
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
        -*)
            echo "Unknown option: $1" >&2
            show_help
            exit 1
            ;;
        *)
            if [[ -z "$VM_NAME" ]]; then
                VM_NAME="$1"
            else
                echo "Error: Multiple VM names provided" >&2
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate VM name provided
if [[ -z "$VM_NAME" ]]; then
    echo "Error: VM name required" >&2
    show_help
    exit 1
fi

# Track results
RESULTS=()
PASSED=0
FAILED=0
START_TIME=$(date +%s.%N)

# Helper functions
log_check() {
    local name="$1"
    local status="$2"
    local message="$3"

    if [[ "$JSON_OUTPUT" == "1" ]]; then
        RESULTS+=("{\"check\":\"$name\",\"status\":\"$status\",\"message\":\"$message\"}")
    else
        if [[ "$status" == "OK" ]]; then
            echo -e "${GREEN}$name: OK${NC} - $message"
        else
            echo -e "${RED}$name: FAILED${NC} - $message"
        fi
    fi

    if [[ "$status" == "OK" ]]; then
        ((PASSED++))
    else
        ((FAILED++))
    fi
}

output_json() {
    local end_time=$(date +%s.%N)
    local elapsed=$(echo "$end_time - $START_TIME" | bc)
    local status="healthy"
    [[ $FAILED -gt 0 ]] && status="unhealthy"

    echo "{"
    echo "  \"vm\": \"$VM_NAME\","
    echo "  \"status\": \"$status\","
    echo "  \"checks_passed\": $PASSED,"
    echo "  \"checks_failed\": $FAILED,"
    echo "  \"elapsed_seconds\": $elapsed,"
    echo "  \"checks\": ["
    local first=1
    for result in "${RESULTS[@]}"; do
        [[ $first -eq 0 ]] && echo ","
        echo -n "    $result"
        first=0
    done
    echo ""
    echo "  ]"
    echo "}"
}

# Check 1: VM exists and get state
check_vm_state() {
    local vm_info

    # Try to get VM info (use --json for parsing)
    if ! vm_info=$(sudo nanofuse vm inspect "$VM_NAME" --json 2>&1); then
        if echo "$vm_info" | grep -qi "not found"; then
            log_check "VM state" "FAILED" "VM not found: $VM_NAME"
            return 1
        else
            log_check "VM state" "FAILED" "Error inspecting VM: $vm_info"
            return 1
        fi
    fi

    # Extract state from JSON
    local state
    state=$(echo "$vm_info" | jq -r '.state // "unknown"' 2>/dev/null || echo "unknown")

    if [[ "$state" == "running" ]]; then
        log_check "VM state" "OK" "running"
        return 0
    else
        log_check "VM state" "FAILED" "State is '$state', expected 'running'"
        return 1
    fi
}

# Check 2: Network connectivity
check_network() {
    # Get VM info as JSON
    local vm_info
    vm_info=$(sudo nanofuse vm inspect "$VM_NAME" --json 2>/dev/null || echo "{}")

    # Get VM IP from JSON and set global
    VM_IP=$(echo "$vm_info" | jq -r '.config.network.ip_address // ""' 2>/dev/null || echo "")

    if [[ -z "$VM_IP" ]] || [[ "$VM_IP" == "null" ]]; then
        log_check "Network" "FAILED" "No IP address assigned"
        return 1
    fi

    # Ping test
    if ping -c 1 -W 2 "$VM_IP" > /dev/null 2>&1; then
        log_check "Network" "OK" "IP $VM_IP responds to ping"
        return 0
    else
        log_check "Network" "FAILED" "IP $VM_IP does not respond to ping"
        return 1
    fi
}

# Check 3: HTTP service (port 80 - static files)
check_http() {
    local vm_ip="$1"

    if [[ -z "$vm_ip" ]]; then
        log_check "HTTP (port 80)" "FAILED" "No IP address to check"
        return 1
    fi

    local response
    if response=$(curl -sf --max-time 5 "http://${vm_ip}:80/" 2>&1); then
        if echo "$response" | grep -qi '<html>'; then
            log_check "HTTP (port 80)" "OK" "Static files served"
            return 0
        else
            log_check "HTTP (port 80)" "OK" "HTTP response received"
            return 0
        fi
    else
        log_check "HTTP (port 80)" "FAILED" "No HTTP response from port 80"
        return 1
    fi
}

# Check 4: Health endpoint (port 80 /health)
check_health_endpoint() {
    local vm_ip="$1"

    if [[ -z "$vm_ip" ]]; then
        log_check "Health endpoint" "FAILED" "No IP address to check"
        return 1
    fi

    local response
    if response=$(curl -sf --max-time 5 "http://${vm_ip}:80/health" 2>&1); then
        # Check for healthy/ok status in JSON
        if echo "$response" | grep -qiE '"status"\s*:\s*"(healthy|ok)"'; then
            log_check "Health endpoint" "OK" "Health check passed"
            return 0
        else
            # Service responded but status unclear
            log_check "Health endpoint" "OK" "Endpoint responded"
            return 0
        fi
    else
        log_check "Health endpoint" "FAILED" "No response from /health endpoint"
        return 1
    fi
}

# Global to pass IP between functions
VM_IP=""

# Main execution
main() {
    if [[ "$JSON_OUTPUT" != "1" ]]; then
        echo "Health check for VM: $VM_NAME"
        echo "========================================"
    fi

    # Check 1: VM state
    check_vm_state || true

    # Check 2: Network (sets VM_IP global)
    check_network || true

    # Check 3 & 4: Services (only if we have an IP)
    if [[ -n "$VM_IP" ]] && [[ "$VM_IP" != "null" ]]; then
        check_http "$VM_IP" || true
        check_health_endpoint "$VM_IP" || true
    else
        log_check "HTTP (port 80)" "FAILED" "Skipped - no network"
        log_check "Health endpoint" "FAILED" "Skipped - no network"
    fi

    # Calculate elapsed time
    local end_time=$(date +%s.%N)
    local elapsed=$(echo "$end_time - $START_TIME" | bc)

    # Output results
    if [[ "$JSON_OUTPUT" == "1" ]]; then
        output_json
    else
        echo "========================================"
        echo -e "Boot time: ${elapsed}s"
        echo -e "Checks passed: ${GREEN}$PASSED${NC}"
        echo -e "Checks failed: ${RED}$FAILED${NC}"

        if [[ $FAILED -eq 0 ]]; then
            echo -e "${GREEN}Overall: HEALTHY${NC}"
        else
            echo -e "${RED}Overall: UNHEALTHY${NC}"
        fi
    fi

    # Exit code
    [[ $FAILED -eq 0 ]] && exit 0 || exit 1
}

main
