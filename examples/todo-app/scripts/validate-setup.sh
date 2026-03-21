#!/bin/bash
# Validate setup before deployment

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== NanoFuse Todo App - Setup Validation ===${NC}"
echo ""

ERRORS=0
WARNINGS=0

# Check 1: nanofuse binaries
echo -e "${YELLOW}[1/8] Checking NanoFuse binaries...${NC}"
if command -v nanofuse &> /dev/null && command -v nanofused &> /dev/null; then
    echo -e "${GREEN}✓ nanofuse and nanofused found${NC}"
    nanofuse version 2>/dev/null || echo "  Version: $(nanofuse --version 2>&1 | head -1)"
else
    echo -e "${RED}✗ nanofuse or nanofused not found${NC}"
    echo "  Install from: https://github.com/peregrinesummit/nanofuse"
    ((ERRORS++))
fi

# Check 2: Docker
echo -e "${YELLOW}[2/8] Checking Docker...${NC}"
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✓ Docker found${NC}"
    docker --version
else
    echo -e "${RED}✗ Docker not found${NC}"
    ((ERRORS++))
fi

# Check 3: KVM support
echo -e "${YELLOW}[3/8] Checking KVM support...${NC}"
if [ -e /dev/kvm ]; then
    echo -e "${GREEN}✓ /dev/kvm exists${NC}"
else
    echo -e "${RED}✗ /dev/kvm not found${NC}"
    echo "  KVM virtualization not available"
    ((ERRORS++))
fi

# Check 4: Firecracker
echo -e "${YELLOW}[4/8] Checking Firecracker...${NC}"
if command -v firecracker &> /dev/null; then
    echo -e "${GREEN}✓ Firecracker found${NC}"
    firecracker --version 2>&1 | head -1
else
    echo -e "${YELLOW}⚠ Firecracker not found in PATH${NC}"
    echo "  NanoFuse may have it in a different location"
    ((WARNINGS++))
fi

# Check 5: Backend binary
echo -e "${YELLOW}[5/8] Checking backend binary...${NC}"
if [ -f "backend/bin/todo-server" ]; then
    echo -e "${GREEN}✓ Backend binary exists${NC}"
    ls -lh backend/bin/todo-server | awk '{print "  Size: " $5}'
else
    echo -e "${YELLOW}⚠ Backend not built yet${NC}"
    echo "  Run: make build"
    ((WARNINGS++))
fi

# Check 6: Container image
echo -e "${YELLOW}[6/8] Checking container image...${NC}"
if docker images | grep -q "nanofuse-todo-app.*test"; then
    echo -e "${GREEN}✓ Container image exists${NC}"
    docker images | grep "nanofuse-todo-app.*test"
else
    echo -e "${YELLOW}⚠ Container image not built${NC}"
    echo "  Run: docker build -f docker/Dockerfile -t nanofuse-todo-app:test ."
    ((WARNINGS++))
fi

# Check 7: Required scripts
echo -e "${YELLOW}[7/8] Checking scripts...${NC}"
SCRIPTS=("deploy-vm.sh" "test-vm.sh" "stability-test.sh")
ALL_SCRIPTS_OK=true
for script in "${SCRIPTS[@]}"; do
    if [ -x "scripts/$script" ]; then
        echo -e "${GREEN}  ✓ scripts/$script${NC}"
    else
        echo -e "${RED}  ✗ scripts/$script not found or not executable${NC}"
        ALL_SCRIPTS_OK=false
        ((ERRORS++))
    fi
done

# Check 8: Dependencies
echo -e "${YELLOW}[8/8] Checking dependencies...${NC}"
DEPS_OK=true
for cmd in curl jq bc; do
    if command -v $cmd &> /dev/null; then
        echo -e "${GREEN}  ✓ $cmd${NC}"
    else
        echo -e "${RED}  ✗ $cmd not found${NC}"
        DEPS_OK=false
        ((ERRORS++))
    fi
done

echo ""
echo -e "${BLUE}=== Summary ===${NC}"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo ""
    echo "Ready to deploy:"
    echo "  sudo ./scripts/deploy-vm.sh"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ ${WARNINGS} warnings${NC}"
    echo ""
    echo "You can proceed, but some features may not work perfectly"
    echo "  sudo ./scripts/deploy-vm.sh"
    exit 0
else
    echo -e "${RED}✗ ${ERRORS} errors, ${WARNINGS} warnings${NC}"
    echo ""
    echo "Please fix errors before deploying"
    exit 1
fi
