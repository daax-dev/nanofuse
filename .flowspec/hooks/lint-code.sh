#!/bin/bash
# Hook: Run linter on task completion
# Triggered by: task.completed

set -e

echo "Running code linter..."

# Parse event data
TASK_ID=$(echo "$HOOK_EVENT" | jq -r '.context.task_id // "unknown"')

echo "Task: $TASK_ID"

# Run your linter
# Examples:
# ruff check . --fix
# eslint --fix .
# golangci-lint run
# mvn checkstyle:check

echo "Linting completed"
