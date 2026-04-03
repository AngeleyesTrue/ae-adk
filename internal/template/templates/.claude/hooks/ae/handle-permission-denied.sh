#!/usr/bin/env bash
# handle-permission-denied.sh
# Logs permission denied events for security auditing.
# Part of AE-ADK hook system.

set -euo pipefail

# Read event data from stdin
EVENT_DATA=$(cat)

# Extract relevant fields
TOOL_NAME=$(printf '%s\n' "$EVENT_DATA" | grep -o '"toolName":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "unknown")
SESSION_ID=$(printf '%s\n' "$EVENT_DATA" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 2>/dev/null || echo "unknown")

# Log to stderr (captured by Claude Code)
echo "[AE] Permission denied: tool=$TOOL_NAME session=$SESSION_ID" >&2

# Exit 0 to not block the flow
exit 0
