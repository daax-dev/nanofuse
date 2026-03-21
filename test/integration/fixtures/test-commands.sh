#!/bin/bash
# Test Commands Script for Recording Integration Tests
# This script simulates typical terminal interactions for testing
# the recording agent's ability to capture terminal I/O.
#
# Requirements: bash, GNU coreutils (for date -Iseconds, date +%N, date +%s.%N)
# Note: This script uses bash-specific features and GNU extensions.
# It is designed for Linux test environments, not for POSIX portability.

set -e

echo "=== Recording Integration Test Commands ==="
echo "Timestamp: $(date -Iseconds)"
echo ""

# Basic echo commands to test output capture
echo "Test 1: Basic echo"
echo "Hello, World!"
echo ""

# Multi-line output
echo "Test 2: Multi-line output"
echo "Line 1"
echo "Line 2"
echo "Line 3"
echo ""

# Command with visible input (simulating user typing)
echo "Test 3: Interactive-style output"
printf "$ "
echo "ls -la"
ls -la /tmp 2>/dev/null || echo "(directory listing)"
echo ""

# File operations (for file event capture)
echo "Test 4: File operations"
TEST_FILE="/tmp/test-recording-$$"
echo "Creating test file: $TEST_FILE"
echo "Test content for recording" > "$TEST_FILE"
cat "$TEST_FILE"
rm -f "$TEST_FILE"
echo "File cleaned up"
echo ""

# Environment information
echo "Test 5: Environment info"
echo "USER: ${USER:-unknown}"
echo "PWD: ${PWD:-unknown}"
echo "SHELL: ${SHELL:-unknown}"
echo ""

# Long output (to test buffer handling)
echo "Test 6: Long output (100 lines)"
# Generate 50-character padding using bash printf (repeat 'x' 50 times)
PADDING=$(printf '%*s' 50 '' | tr ' ' 'x')
for i in $(seq 1 100); do
    echo "Output line $i: $PADDING"
done
echo ""

# Rapid output (for throughput testing)
echo "Test 7: Rapid output (1000 fast lines)"
for i in $(seq 1 1000); do
    echo "Rapid line $i"
done
echo ""

# Command timing
echo "Test 8: Timing test"
echo "Start: $(date +%s.%N)"
sleep 0.1
echo "End: $(date +%s.%N)"
echo ""

# Exit status capture
echo "Test 9: Exit status handling"
echo "Running command that succeeds..."
true
echo "Exit status: $?"
echo "Running command that fails..."
false || true
echo "Exit status after recovery: $?"
echo ""

# Final summary
echo "=== Test Commands Completed ==="
echo "All commands executed successfully."
echo "Check recording events for captured I/O."
