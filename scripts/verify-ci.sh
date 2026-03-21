#!/bin/bash
# Verify CI/CD pipeline configuration
# Checks that all required files exist and are valid

set -e

echo "🔍 NanoFuse CI/CD Verification"
echo "=============================="
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

FAILED=0

check_file() {
    local file=$1
    local description=$2

    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $description: $file"
    else
        echo -e "${RED}✗${NC} Missing $description: $file"
        FAILED=1
    fi
}

check_yaml_syntax() {
    local file=$1

    if command -v yamllint &> /dev/null; then
        if yamllint -d relaxed "$file" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} YAML syntax valid: $file"
        else
            echo -e "${YELLOW}⚠${NC} YAML syntax issues in: $file"
        fi
    fi
}

echo "1️⃣  Checking CI/CD workflow files..."
check_file ".github/workflows/ci.yaml" "Main CI/CD workflow"
check_file ".github/workflows/pr-comment.yaml" "PR comment workflow"
check_file ".github/dependabot.yml" "Dependabot config"
echo ""

echo "2️⃣  Checking configuration files..."
check_file ".golangci.yml" "golangci-lint config"
check_file "magefile.go" "Magefile"
check_file ".gitignore" "Git ignore file"
echo ""

echo "3️⃣  Checking documentation..."
check_file "README.md" "README"
check_file "CONTRIBUTING.md" "Contributing guide"
check_file "LICENSE" "License file"
check_file "docs/CI_CD.md" "CI/CD documentation"
check_file "docs/TESTING.md" "Testing guide"
echo ""

echo "4️⃣  Checking source structure..."
check_file "cmd/nanofuse/main.go" "CLI main"
check_file "cmd/nanofused/main.go" "Daemon main"
check_file "internal/api/server.go" "API server"
check_file "go.mod" "Go module file"
echo ""

echo "5️⃣  Checking Docker files..."
check_file "images/base/Dockerfile" "Base image Dockerfile"
echo ""

echo "6️⃣  Checking test files..."
check_file "internal/api/server_test.go" "API tests"
check_file "cmd/nanofuse/main_test.go" "CLI tests"
echo ""

echo "7️⃣  Validating YAML syntax..."
if command -v yamllint &> /dev/null; then
    check_yaml_syntax ".github/workflows/ci.yaml"
    check_yaml_syntax ".github/workflows/pr-comment.yaml"
    check_yaml_syntax ".github/dependabot.yml"
    check_yaml_syntax ".golangci.yml"
else
    echo -e "${YELLOW}⚠${NC} yamllint not installed, skipping YAML validation"
fi
echo ""

echo "8️⃣  Checking Go module..."
if go mod verify > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Go modules verified"
else
    echo -e "${RED}✗${NC} Go module verification failed"
    FAILED=1
fi
echo ""

echo "9️⃣  Checking binary builds..."
if cd cmd/nanofuse && CGO_ENABLED=0 go build -o /tmp/nanofuse . > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} CLI binary builds successfully"
    rm -f /tmp/nanofuse
else
    echo -e "${RED}✗${NC} CLI binary build failed"
    FAILED=1
fi

if cd ../nanofused && CGO_ENABLED=0 go build -o /tmp/nanofused . > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Daemon binary builds successfully"
    rm -f /tmp/nanofused
else
    echo -e "${RED}✗${NC} Daemon binary build failed"
    FAILED=1
fi
cd ../..
echo ""

echo "🔟 Checking required GitHub Actions..."
echo "   The following actions are used:"
grep -h "uses:" .github/workflows/*.yaml | sed 's/.*uses: //' | sort -u | while read action; do
    echo "   - $action"
done
echo ""

echo "=============================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All CI/CD components verified!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Commit and push changes to trigger CI"
    echo "  2. Create a PR to test the pipeline"
    echo "  3. Merge to main to publish artifacts"
    echo "  4. Create a tag (v0.1.0) to create a release"
    exit 0
else
    echo -e "${RED}❌ Some components are missing or invalid!${NC}"
    echo "Please fix the issues above."
    exit 1
fi
