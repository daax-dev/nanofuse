#!/bin/bash
# Hook: Quality gate before validation
# Triggered by: validate.started
# Fail mode: stop (blocks workflow if quality checks fail)

set -e

echo "Running quality gate checks..."

SPEC_ID=$(echo "$HOOK_EVENT" | jq -r '.feature // "unknown"')
echo "Spec: $SPEC_ID"

# Example quality checks
echo "Checking code coverage..."
# coverage report --fail-under=80

echo "Checking complexity..."
# radon cc . -a -nb

echo "Checking security..."
# bandit -r src/

echo "Quality gate passed"
