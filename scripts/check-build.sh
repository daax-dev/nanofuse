#!/bin/bash
# Check if build passes locally before pushing
# This script runs the same checks as CI/CD pipeline

set -e

echo "🔍 NanoFuse Build Checker"
echo "========================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track failures
FAILED=0

# Function to run check
run_check() {
    local name=$1
    local cmd=$2

    echo -n "⏳ $name... "
    if eval "$cmd" > /tmp/nanofuse-check.log 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        echo "   Error details:"
        tail -n 5 /tmp/nanofuse-check.log | sed 's/^/   /'
        FAILED=1
    fi
}

echo "1️⃣  Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓ Go installed: $GO_VERSION${NC}"
echo ""

echo "2️⃣  Checking dependencies..."
run_check "Download dependencies" "go mod download"
run_check "Verify dependencies" "go mod verify"
echo ""

echo "3️⃣  Running go vet..."
run_check "Go vet" "go vet ./..."
echo ""

echo "4️⃣  Running tests..."
run_check "Unit tests" "go test -race ./..."
echo ""

echo "5️⃣  Building binaries..."
run_check "Build CLI" "cd cmd/nanofuse && CGO_ENABLED=0 go build -o nanofuse ."
run_check "Build daemon" "cd cmd/nanofused && CGO_ENABLED=0 go build -o nanofused ."
echo ""

echo "6️⃣  Running linter (if installed)..."
if command -v golangci-lint &> /dev/null; then
    run_check "golangci-lint" "golangci-lint run --timeout=5m"
else
    echo -e "${YELLOW}⚠ golangci-lint not installed, skipping${NC}"
fi
echo ""

# Summary
echo "========================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed!${NC}"
    echo "Safe to push to GitHub."
    exit 0
else
    echo -e "${RED}❌ Some checks failed!${NC}"
    echo "Please fix the errors before pushing."
    exit 1
fi
