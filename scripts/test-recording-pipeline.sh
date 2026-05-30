#!/bin/bash
# Test verification script for recording pipeline
# Runs all tests related to the recording functionality with proper quality gates
#
# Addresses all issues from test-feedback.md:
# 1. No redundant test execution
# 2. No eval usage - direct command execution
# 3. No silent error suppression
# 4. Coverage threshold enforced
# 5. Race detection, shuffle, timeout flags
# 6. Full bash strict mode
# 7. Working directory validation
# 8. Consistent output handling
# 9. No redundant build steps (go test compiles)
# 10. Trap-based cleanup
# 11. Static analysis included

set -euo pipefail

# Determine script and repo locations
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
COVERAGE_THRESHOLD=20  # Start realistic, increase over time as more tests are added
TEST_TIMEOUT="5m"
COVERAGE_FILE="coverage-recording.out"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Cleanup function - runs on exit
cleanup() {
    local exit_code=$?
    rm -f "$REPO_ROOT/$COVERAGE_FILE"
    exit $exit_code
}
trap cleanup EXIT

# Validate we're in the right directory
validate_repo() {
    if [[ ! -f "$REPO_ROOT/go.mod" ]]; then
        echo -e "${RED}ERROR: Cannot find go.mod. Run from repo root or scripts/ directory.${NC}" >&2
        exit 1
    fi

    if ! grep -q "github.com/daax-dev/nanofuse" "$REPO_ROOT/go.mod"; then
        echo -e "${RED}ERROR: This doesn't appear to be the nanofuse repository.${NC}" >&2
        exit 1
    fi
}

# Print section header
section() {
    echo ""
    echo "========================================"
    echo "$1"
    echo "========================================"
}

# Run a test step with proper error handling
run_step() {
    local name="$1"
    shift
    local cmd=("$@")

    echo -n "Running $name... "

    local output
    local exit_code=0

    # Capture output and exit code without eval
    if output=$("${cmd[@]}" 2>&1); then
        echo -e "${GREEN}PASS${NC}"
        return 0
    else
        exit_code=$?
        echo -e "${RED}FAIL${NC}"
        echo "Command: ${cmd[*]}"
        echo "Exit code: $exit_code"
        echo "Output:"
        echo "$output" | tail -50
        return $exit_code
    fi
}

main() {
    echo "========================================"
    echo "Recording Pipeline Test Verification"
    echo "========================================"
    echo "Repo: $REPO_ROOT"
    echo "Coverage threshold: ${COVERAGE_THRESHOLD}%"
    echo "Test timeout: $TEST_TIMEOUT"

    cd "$REPO_ROOT"
    validate_repo

    section "1. Static Analysis"
    run_step "go vet" go vet ./internal/recording/... ./internal/api/...

    # Run staticcheck if available
    if command -v staticcheck &> /dev/null; then
        run_step "staticcheck" staticcheck ./internal/recording/... ./internal/api/...
    else
        echo -e "${YELLOW}SKIP: staticcheck not installed${NC}"
    fi

    section "2. Build Verification"
    # Single build check - go test will also compile, but this separates compile errors
    run_step "build" go build ./...

    section "3. Unit Tests with Race Detection"
    # Single comprehensive test run with all quality flags
    echo "Running tests with -race -shuffle=on -timeout=$TEST_TIMEOUT"

    if ! go test \
        -race \
        -shuffle=on \
        -timeout="$TEST_TIMEOUT" \
        -coverprofile="$COVERAGE_FILE" \
        -v \
        ./internal/recording/... \
        ./internal/api/... \
        ./internal/firecracker/... 2>&1 | tee /tmp/test-output.txt; then

        echo -e "${RED}FAIL: Tests failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}PASS: All tests passed${NC}"

    section "4. Coverage Enforcement"

    if [[ ! -f "$COVERAGE_FILE" ]]; then
        echo -e "${RED}FAIL: Coverage file not generated${NC}"
        exit 1
    fi

    # Extract coverage percentage
    COVERAGE_LINE=$(go tool cover -func="$COVERAGE_FILE" | grep total || true)

    if [[ -z "$COVERAGE_LINE" ]]; then
        echo -e "${RED}FAIL: Could not extract coverage${NC}"
        exit 1
    fi

    COVERAGE=$(echo "$COVERAGE_LINE" | awk '{print $3}' | tr -d '%')

    echo "Coverage: ${COVERAGE}%"
    echo "Threshold: ${COVERAGE_THRESHOLD}%"

    # Compare coverage to threshold
    if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
        echo -e "${RED}FAIL: Coverage ${COVERAGE}% is below threshold ${COVERAGE_THRESHOLD}%${NC}"
        exit 1
    fi

    echo -e "${GREEN}PASS: Coverage meets threshold${NC}"

    section "5. Summary"
    echo -e "${GREEN}All quality gates passed!${NC}"
    echo ""
    echo "Results:"
    echo "  - Static analysis: PASS"
    echo "  - Build: PASS"
    echo "  - Tests (with race detection): PASS"
    echo "  - Coverage: ${COVERAGE}% (threshold: ${COVERAGE_THRESHOLD}%)"
}

main "$@"
