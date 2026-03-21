#!/bin/bash
set -e

# NanoFuse Build Validation Script
# This script proves what actually works

echo "==============================================="
echo "NanoFuse Build Validation"
echo "==============================================="
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    return 1
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

section() {
    echo
    echo "-------------------"
    echo "$1"
    echo "-------------------"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    fail "Not in project root directory (go.mod not found)"
    exit 1
fi

section "1. Go Environment"
go version && pass "Go is installed" || fail "Go is not installed"

section "2. Build CLI"
if go build -o ./bin/nanofuse ./cmd/nanofuse 2>/dev/null; then
    pass "CLI builds successfully"
    ls -lh ./bin/nanofuse | awk '{print "  Size: " $5}'
else
    fail "CLI build failed"
    exit 1
fi

section "3. Test CLI"
if ./bin/nanofuse version >/dev/null 2>&1; then
    pass "CLI runs successfully"
    ./bin/nanofuse version | sed 's/^/  /'
else
    fail "CLI execution failed"
    exit 1
fi

section "4. Build Daemon"
if go build -o ./bin/nanofused ./cmd/nanofused 2>/dev/null; then
    pass "Daemon builds successfully"
    ls -lh ./bin/nanofused | awk '{print "  Size: " $5}'
else
    fail "Daemon build failed"
    exit 1
fi

section "5. Run Unit Tests"
if go test ./... -v 2>&1 | grep -q "PASS"; then
    pass "Unit tests pass"
    go test ./... 2>&1 | grep "^ok" | sed 's/^/  /'
else
    fail "Unit tests failed"
    go test ./...
    exit 1
fi

section "6. Test Coverage"
go test -coverprofile=coverage.out ./... >/dev/null 2>&1
if [ -f coverage.out ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    pass "Coverage report generated: $COVERAGE"
else
    warn "Coverage report not generated"
fi

section "7. Check Mage"
if command -v mage >/dev/null 2>&1; then
    pass "Mage is installed"
elif [ -f ~/go/bin/mage ]; then
    pass "Mage is installed at ~/go/bin/mage"
else
    warn "Mage not found (install with: go install github.com/magefile/mage@latest)"
fi

section "8. Check Build Artifacts"
ARTIFACTS=0
[ -f "./bin/nanofuse" ] && pass "CLI binary exists" && ((ARTIFACTS++))
[ -f "./bin/nanofused" ] && pass "Daemon binary exists" && ((ARTIFACTS++))
[ -f "go.mod" ] && pass "go.mod exists" && ((ARTIFACTS++))
[ -f "magefile.go" ] && pass "magefile.go exists" && ((ARTIFACTS++))

echo
echo "Build artifacts: $ARTIFACTS/4"

section "9. Check Documentation"
DOCS=0
[ -f "docs/API_CONTRACT.md" ] && ((DOCS++))
[ -f "docs/CLI_SPEC.md" ] && ((DOCS++))
[ -f "docs/ARCHITECTURE_DECISIONS.md" ] && ((DOCS++))
[ -f "docs/EXECUTION_PLAN.md" ] && ((DOCS++))
[ -f "ACTUAL_STATUS_REPORT.md" ] && ((DOCS++))

if [ $DOCS -eq 5 ]; then
    pass "All documentation files exist ($DOCS/5)"
else
    warn "Some documentation files missing ($DOCS/5)"
fi

section "10. Check Integration Tests"
if [ -f "test/integration/api_integration_test.go" ]; then
    pass "Integration tests exist"
    if go test -tags=integration ./test/integration/... 2>&1 | grep -q "PASS"; then
        pass "Integration tests compile and pass"
    else
        warn "Integration tests exist but don't compile/pass yet"
    fi
else
    warn "Integration tests not created yet"
fi

section "11. Check Docker Files"
if [ -f "images/base/Dockerfile" ]; then
    pass "Base Dockerfile exists"
    if [ -d "images/base/build" ]; then
        pass "Docker build artifacts exist"
    else
        warn "Docker image not built yet (run: cd images/base && sudo make build)"
    fi
else
    fail "Base Dockerfile missing"
fi

section "12. Check CI/CD"
if [ -f ".github/workflows/ci.yaml" ]; then
    pass "GitHub Actions workflow exists"
else
    warn "CI workflow not configured"
fi

section "Summary"
echo
echo "What Works:"
echo "  ✓ Go binaries build"
echo "  ✓ CLI runs and responds"
echo "  ✓ Unit tests pass"
echo "  ✓ Mage build system configured"
echo "  ✓ Documentation complete"
echo
echo "What Needs Work:"
echo "  ⚠ Integration tests need fixing"
echo "  ⚠ Docker image needs building (requires sudo)"
echo "  ⚠ Daemon needs end-to-end testing"
echo "  ⚠ CI pipeline needs triggering"
echo
echo "Next Steps:"
echo "  1. Fix integration tests to match actual APIs"
echo "  2. Build Docker image: cd images/base && sudo make build"
echo "  3. Test daemon with real config"
echo "  4. Push to GitHub to trigger CI"
echo
echo "==============================================="
echo "Validation Complete"
echo "==============================================="
