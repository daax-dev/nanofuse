#!/bin/bash
# Test todo-app VM functionality

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

VM_NAME="todo-app-test"
TESTS_PASSED=0
TESTS_FAILED=0

echo -e "${BLUE}=== NanoFuse Todo App Test Suite ===${NC}"
echo ""

# Get VM IP
echo -e "${YELLOW}Getting VM IP address...${NC}"
if command -v nanofuse &> /dev/null; then
    VM_IP=$(nanofuse vm info "$VM_NAME" 2>/dev/null | grep -oP 'IP.*:\s*\K[\d.]+' || echo "")
else
    echo -e "${RED}nanofuse command not found. Is it installed?${NC}"
    exit 1
fi

if [ -z "$VM_IP" ]; then
    echo -e "${RED}Could not get VM IP address${NC}"
    echo "Is the VM running? Check: nanofuse vm status $VM_NAME"
    exit 1
fi

echo -e "${GREEN}VM IP: $VM_IP${NC}"
echo ""

# Helper function for tests
run_test() {
    local test_name="$1"
    local test_command="$2"

    echo -e "${YELLOW}Testing: ${test_name}${NC}"

    if eval "$test_command"; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
    echo ""
}

# Test 1: Health Check
run_test "Health endpoint" \
    "curl -sf --connect-timeout 5 http://${VM_IP}:8080/health | jq -e '.status == \"healthy\"' > /dev/null"

# Test 2: Readiness Check
run_test "Readiness endpoint" \
    "curl -sf --connect-timeout 5 http://${VM_IP}:8080/ready | jq -e '.status == \"ready\"' > /dev/null"

# Test 3: List todos (should be empty initially)
run_test "List todos (empty)" \
    "curl -sf http://${VM_IP}:8080/api/v1/todos | jq -e '.total == 0' > /dev/null"

# Test 4: Create todo
echo -e "${YELLOW}Testing: Create todo${NC}"
TODO_RESPONSE=$(curl -sf -X POST http://${VM_IP}:8080/api/v1/todos \
    -H "Content-Type: application/json" \
    -d '{"title":"Test from script","description":"Automated test","priority":1,"tags":["test","automated"]}')

if echo "$TODO_RESPONSE" | jq -e '.id' > /dev/null; then
    TODO_ID=$(echo "$TODO_RESPONSE" | jq -r '.id')
    echo -e "${GREEN}✓ PASS (ID: $TODO_ID)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "Response: $TODO_RESPONSE"
    ((TESTS_FAILED++))
fi
echo ""

# Test 5: Get specific todo
if [ -n "$TODO_ID" ]; then
    run_test "Get todo by ID" \
        "curl -sf http://${VM_IP}:8080/api/v1/todos/${TODO_ID} | jq -e '.title == \"Test from script\"' > /dev/null"
fi

# Test 6: List todos (should have 1)
run_test "List todos (with data)" \
    "curl -sf http://${VM_IP}:8080/api/v1/todos | jq -e '.total == 1' > /dev/null"

# Test 7: Update todo
if [ -n "$TODO_ID" ]; then
    run_test "Update todo" \
        "curl -sf -X PUT http://${VM_IP}:8080/api/v1/todos/${TODO_ID} \
            -H 'Content-Type: application/json' \
            -d '{\"completed\":true}' | jq -e '.completed == true' > /dev/null"
fi

# Test 8: Delete todo
if [ -n "$TODO_ID" ]; then
    echo -e "${YELLOW}Testing: Delete todo${NC}"
    HTTP_CODE=$(curl -sf -X DELETE -w "%{http_code}" -o /dev/null http://${VM_IP}:8080/api/v1/todos/${TODO_ID})
    if [ "$HTTP_CODE" = "204" ]; then
        echo -e "${GREEN}✓ PASS (HTTP 204)${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL (HTTP $HTTP_CODE)${NC}"
        ((TESTS_FAILED++))
    fi
    echo ""
fi

# Test 9: List todos (should be empty again)
run_test "List todos (after delete)" \
    "curl -sf http://${VM_IP}:8080/api/v1/todos | jq -e '.total == 0' > /dev/null"

# Test 10: Metrics endpoint
run_test "Prometheus metrics" \
    "curl -sf http://${VM_IP}:8080/metrics | grep -q 'todo_app_todos_created_total'"

# Test 11: Frontend (Nginx)
run_test "Frontend (Nginx)" \
    "curl -sf http://${VM_IP}:80 | grep -q 'Todo App'"

# Summary
echo ""
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
echo -e "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
