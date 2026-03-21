#!/bin/bash
# Hook: Update changelog on spec creation
# Triggered by: spec.created

set -e

echo "Updating CHANGELOG.md..."

# Parse event data
EVENT_TYPE=$(echo "$HOOK_EVENT" | jq -r '.event_type')
SPEC_ID=$(echo "$HOOK_EVENT" | jq -r '.feature // "unknown"')
TIMESTAMP=$(date +"%Y-%m-%d")

# Create CHANGELOG.md if it doesn't exist
if [ ! -f "CHANGELOG.md" ]; then
    echo "# Changelog" > CHANGELOG.md
    echo "" >> CHANGELOG.md
fi

# Add entry (simple implementation - customize as needed)
ENTRY="## [$SPEC_ID] - $TIMESTAMP\n\n- Feature specification created\n"
sed -i "3i $ENTRY" CHANGELOG.md

echo "CHANGELOG.md updated for $SPEC_ID"
