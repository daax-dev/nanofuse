#!/bin/bash
# Hook: Run tests after implementation
# Triggered by: implement.completed

set -e

echo "Running test suite..."

# Parse event data from HOOK_EVENT env var
EVENT_TYPE=$(echo "$HOOK_EVENT" | jq -r '.event_type')
SPEC_ID=$(echo "$HOOK_EVENT" | jq -r '.feature // "unknown"')

echo "Event: $EVENT_TYPE"
echo "Spec: $SPEC_ID"

# Run your test command here
# Examples:
# pytest tests/
# npm test
# go test ./...
# mvn test

echo "Tests completed successfully"
