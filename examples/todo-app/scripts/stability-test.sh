#!/bin/bash
# 24-hour stability test for todo-app VM

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
VM_NAME="todo-app-test"
TEST_DURATION_HOURS=${TEST_DURATION_HOURS:-24}
TEST_INTERVAL_SECONDS=${TEST_INTERVAL_SECONDS:-60}
LOG_FILE="/tmp/todo-app-stability-$(date +%Y%m%d-%H%M%S).log"
METRICS_FILE="/tmp/todo-app-metrics-$(date +%Y%m%d-%H%M%S).csv"

echo -e "${BLUE}=== 24-Hour Stability Test ===${NC}"
echo "VM: $VM_NAME"
echo "Duration: ${TEST_DURATION_HOURS} hours"
echo "Check interval: ${TEST_INTERVAL_SECONDS} seconds"
echo "Log file: $LOG_FILE"
echo "Metrics file: $METRICS_FILE"
echo ""

# Get VM IP
VM_IP=$(nanofuse vm info "$VM_NAME" 2>/dev/null | grep -oP 'IP.*:\s*\K[\d.]+' || echo "")
if [ -z "$VM_IP" ]; then
    echo -e "${RED}Error: Could not get VM IP${NC}"
    exit 1
fi
echo -e "${GREEN}VM IP: $VM_IP${NC}"
echo ""

# Initialize metrics CSV
echo "timestamp,iteration,health_status,health_response_ms,api_status,api_response_ms,todos_created,todos_completed,todos_deleted,todos_active,memory_check,cpu_check,errors" > "$METRICS_FILE"

# Calculate end time
START_TIME=$(date +%s)
END_TIME=$((START_TIME + TEST_DURATION_HOURS * 3600))
ITERATION=0
TOTAL_ERRORS=0
CONSECUTIVE_ERRORS=0

echo -e "${YELLOW}Starting continuous monitoring...${NC}"
echo "Press Ctrl+C to stop early"
echo ""

# Trap for clean exit
trap 'echo ""; echo "Test interrupted. Check logs: $LOG_FILE"; exit 0' INT TERM

while [ $(date +%s) -lt $END_TIME ]; do
    ((ITERATION++))
    CURRENT_TIME=$(date +"%Y-%m-%d %H:%M:%S")
    ERRORS_THIS_ITERATION=0

    # Health check
    HEALTH_START=$(date +%s%3N)
    if HEALTH_RESPONSE=$(curl -sf --connect-timeout 5 http://${VM_IP}:8080/health 2>/dev/null); then
        HEALTH_END=$(date +%s%3N)
        HEALTH_MS=$((HEALTH_END - HEALTH_START))
        HEALTH_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.status' 2>/dev/null || echo "error")

        if [ "$HEALTH_STATUS" = "healthy" ]; then
            HEALTH_OK="OK"
        else
            HEALTH_OK="FAIL"
            ((ERRORS_THIS_ITERATION++))
        fi
    else
        HEALTH_MS="timeout"
        HEALTH_OK="FAIL"
        ((ERRORS_THIS_ITERATION++))
    fi

    # API Test - Create a todo
    API_START=$(date +%s%3N)
    if API_RESPONSE=$(curl -sf -X POST http://${VM_IP}:8080/api/v1/todos \
        -H "Content-Type: application/json" \
        -d "{\"title\":\"Stability test iteration $ITERATION\",\"description\":\"$(date)\",\"priority\":0,\"tags\":[\"test\"]}" 2>/dev/null); then

        API_END=$(date +%s%3N)
        API_MS=$((API_END - API_START))
        TODO_ID=$(echo "$API_RESPONSE" | jq -r '.id' 2>/dev/null || echo "")

        if [ -n "$TODO_ID" ] && [ "$TODO_ID" != "null" ]; then
            API_OK="OK"

            # Mark as completed
            curl -sf -X PUT http://${VM_IP}:8080/api/v1/todos/${TODO_ID} \
                -H "Content-Type: application/json" \
                -d '{"completed":true}' > /dev/null 2>&1 || true

            # Delete it to keep DB clean
            curl -sf -X DELETE http://${VM_IP}:8080/api/v1/todos/${TODO_ID} > /dev/null 2>&1 || true
        else
            API_OK="FAIL"
            ((ERRORS_THIS_ITERATION++))
        fi
    else
        API_MS="timeout"
        API_OK="FAIL"
        ((ERRORS_THIS_ITERATION++))
    fi

    # Get metrics
    if METRICS_RESPONSE=$(curl -sf http://${VM_IP}:8080/metrics 2>/dev/null); then
        TODOS_CREATED=$(echo "$METRICS_RESPONSE" | grep '^todo_app_todos_created_total' | awk '{print $2}' || echo "0")
        TODOS_COMPLETED=$(echo "$METRICS_RESPONSE" | grep '^todo_app_todos_completed_total' | awk '{print $2}' || echo "0")
        TODOS_DELETED=$(echo "$METRICS_RESPONSE" | grep '^todo_app_todos_deleted_total' | awk '{print $2}' || echo "0")
        TODOS_ACTIVE=$(echo "$METRICS_RESPONSE" | grep '^todo_app_todos_active' | awk '{print $2}' || echo "0")
    else
        TODOS_CREATED="N/A"
        TODOS_COMPLETED="N/A"
        TODOS_DELETED="N/A"
        TODOS_ACTIVE="N/A"
    fi

    # Memory and CPU checks (placeholder - would need VM monitoring)
    MEMORY_CHECK="OK"
    CPU_CHECK="OK"

    # Update error counters
    if [ $ERRORS_THIS_ITERATION -gt 0 ]; then
        ((TOTAL_ERRORS += ERRORS_THIS_ITERATION))
        ((CONSECUTIVE_ERRORS++))
    else
        CONSECUTIVE_ERRORS=0
    fi

    # Log to CSV
    echo "${CURRENT_TIME},${ITERATION},${HEALTH_OK},${HEALTH_MS},${API_OK},${API_MS},${TODOS_CREATED},${TODOS_COMPLETED},${TODOS_DELETED},${TODOS_ACTIVE},${MEMORY_CHECK},${CPU_CHECK},${ERRORS_THIS_ITERATION}" >> "$METRICS_FILE"

    # Log details
    echo "[$(date +"%H:%M:%S")] Iteration $ITERATION | Health: $HEALTH_OK (${HEALTH_MS}ms) | API: $API_OK (${API_MS}ms) | Errors: $ERRORS_THIS_ITERATION/$TOTAL_ERRORS" | tee -a "$LOG_FILE"

    # Alert on consecutive errors
    if [ $CONSECUTIVE_ERRORS -ge 3 ]; then
        echo -e "${RED}[ALERT] $CONSECUTIVE_ERRORS consecutive errors detected!${NC}" | tee -a "$LOG_FILE"
    fi

    # Calculate time remaining
    CURRENT=$(date +%s)
    ELAPSED=$((CURRENT - START_TIME))
    REMAINING=$((END_TIME - CURRENT))
    ELAPSED_HOURS=$((ELAPSED / 3600))
    REMAINING_HOURS=$((REMAINING / 3600))

    # Status update every hour
    if [ $((ITERATION % 60)) -eq 0 ]; then
        echo ""
        echo -e "${BLUE}=== Status Update ===${NC}"
        echo "Elapsed: ${ELAPSED_HOURS}h | Remaining: ${REMAINING_HOURS}h"
        echo "Iterations: $ITERATION | Total Errors: $TOTAL_ERRORS"
        echo "Success Rate: $(awk "BEGIN {printf \"%.2f%%\", 100 - ($TOTAL_ERRORS / $ITERATION * 100)}")"
        echo ""
    fi

    # Sleep until next check
    sleep $TEST_INTERVAL_SECONDS
done

# Final report
echo ""
echo -e "${BLUE}=== Test Complete ===${NC}"
echo "Total iterations: $ITERATION"
echo "Total errors: $TOTAL_ERRORS"
SUCCESS_RATE=$(awk "BEGIN {printf \"%.2f\", 100 - ($TOTAL_ERRORS / $ITERATION * 100)}")
echo "Success rate: ${SUCCESS_RATE}%"
echo ""
echo "Logs: $LOG_FILE"
echo "Metrics: $METRICS_FILE"
echo ""

if [ $TOTAL_ERRORS -eq 0 ]; then
    echo -e "${GREEN}✓ PERFECT: Zero errors during $TEST_DURATION_HOURS hour test!${NC}"
    exit 0
elif [ $(echo "$SUCCESS_RATE > 99" | bc) -eq 1 ]; then
    echo -e "${GREEN}✓ EXCELLENT: ${SUCCESS_RATE}% success rate${NC}"
    exit 0
elif [ $(echo "$SUCCESS_RATE > 95" | bc) -eq 1 ]; then
    echo -e "${YELLOW}⚠ GOOD: ${SUCCESS_RATE}% success rate (some issues)${NC}"
    exit 0
else
    echo -e "${RED}✗ POOR: ${SUCCESS_RATE}% success rate (significant issues)${NC}"
    exit 1
fi
