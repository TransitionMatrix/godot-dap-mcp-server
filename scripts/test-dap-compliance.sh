#!/usr/bin/env bash
#
# Automated test script for running test-dap-protocol against Godot
#
# This script:
# 1. Starts Godot in the background with DAP enabled
# 2. Waits for DAP server to be ready
# 3. Runs test-dap-protocol in automated mode
# 4. Captures output for analysis
# 5. Cleans up Godot process
#
# Usage:
#   ./scripts/test-dap-compliance.sh
#
# Environment variables:
#   GODOT_BIN - Path to Godot binary (required)
#   PROJECT_PATH - Path to project.godot (default: tests/fixtures/test-project/project.godot)
#   DAP_PORT - DAP server port (default: 6006)
#   OUTPUT_FILE - Where to save output (default: /tmp/godot-dap-test-output.txt)
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DAP_PORT="${DAP_PORT:-6006}"
GODOT_STARTUP_TIMEOUT=15
PROJECT_PATH="${PROJECT_PATH:-tests/fixtures/test-project/project.godot}"
OUTPUT_FILE="${OUTPUT_FILE:-/tmp/godot-dap-test-output-$(date +%Y%m%d-%H%M%S).txt}"

# Validate GODOT_BIN
if [ -z "${GODOT_BIN:-}" ]; then
    echo -e "${RED}✗ GODOT_BIN environment variable not set${NC}"
    echo "Set it to the path of your Godot binary:"
    echo "  export GODOT_BIN=/path/to/godot.macos.editor.arm64"
    echo ""
    echo "Example:"
    echo "  export GODOT_BIN=/Users/adp/Projects/godot/bin/godot.macos.editor.arm64"
    echo "  ./scripts/test-dap-compliance.sh"
    exit 1
fi

if [ ! -f "$GODOT_BIN" ]; then
    echo -e "${RED}✗ Godot binary not found: $GODOT_BIN${NC}"
    exit 1
fi

if [ ! -x "$GODOT_BIN" ]; then
    echo -e "${RED}✗ Godot binary is not executable: $GODOT_BIN${NC}"
    exit 1
fi

# Resolve project path
if [ ! -f "$PROJECT_PATH" ]; then
    echo -e "${RED}✗ Project file not found: $PROJECT_PATH${NC}"
    exit 1
fi

PROJECT_DIR=$(dirname "$PROJECT_PATH")
PROJECT_DIR=$(cd "$PROJECT_DIR" && pwd)

# Get Godot version for display
GODOT_VERSION=$("$GODOT_BIN" --version 2>/dev/null | head -n1 || echo "unknown")

# Global variables for cleanup
GODOT_PID=""

# Cleanup function
cleanup() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Cleaning up...${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    if [ -n "$GODOT_PID" ]; then
        echo -e "${YELLOW}Stopping Godot (PID: $GODOT_PID)...${NC}"
        kill $GODOT_PID 2>/dev/null || true
        wait $GODOT_PID 2>/dev/null || true
        echo -e "${GREEN}✓ Godot stopped${NC}"
    fi

    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

# Register cleanup on exit
trap cleanup EXIT INT TERM

# Helper: Check if port is in use
is_port_in_use() {
    local port=$1
    nc -z 127.0.0.1 $port 2>/dev/null
}

# Helper: Wait for DAP server to be ready
wait_for_port() {
    local port=$1
    local timeout=$2
    local elapsed=0

    echo -e "${YELLOW}Waiting for DAP server on port $port (timeout: ${timeout}s)...${NC}"

    while [ $elapsed -lt $timeout ]; do
        if nc -z 127.0.0.1 $port 2>/dev/null; then
            echo -e "${GREEN}✓ DAP server is ready on port $port${NC}"
            return 0
        fi
        sleep 0.5
        elapsed=$((elapsed + 1))
    done

    echo -e "${RED}✗ Timeout waiting for DAP server on port $port${NC}"
    return 1
}

# Print header
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Godot DAP Compliance Test${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${GREEN}Configuration:${NC}"
echo "  Godot Binary: $GODOT_BIN"
echo "  Godot Version: $GODOT_VERSION"
echo "  Project: $PROJECT_DIR"
echo "  DAP Port: $DAP_PORT"
echo "  Output File: $OUTPUT_FILE"
echo ""

# Check if port is already in use
if is_port_in_use $DAP_PORT; then
    echo -e "${RED}✗ Port $DAP_PORT is already in use${NC}"
    echo "Either:"
    echo "  1. Stop the process using port $DAP_PORT"
    echo "  2. Set DAP_PORT to a different port: DAP_PORT=6007 ./scripts/test-dap-compliance.sh"
    exit 1
fi

# Step 1: Start Godot with DAP enabled
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 1: Starting Godot Editor${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Create a temporary editor settings override to enable DAP
# Godot 4 uses editor_settings-4.tres in the config directory
EDITOR_SETTINGS_OVERRIDE="/tmp/godot-dap-test-editor-settings-$(date +%s).tres"

cat > "$EDITOR_SETTINGS_OVERRIDE" <<EOF
[gd_resource type="EditorSettings" format=3]

[resource]
network/debug_adapter/remote_port = $DAP_PORT
network/debug_adapter/request_timeout = 1000
network/debug_adapter/sync_breakpoints = true
EOF

# Start Godot in headless mode with DAP enabled
# Use --editor flag to start editor, --quit to prevent opening project manager
echo -e "${YELLOW}Starting Godot with DAP server on port $DAP_PORT...${NC}"

"$GODOT_BIN" \
    --editor \
    --path "$PROJECT_DIR" \
    --headless \
    > "/tmp/godot-console-$$.log" 2>&1 &

GODOT_PID=$!
echo -e "${GREEN}✓ Godot started (PID: $GODOT_PID)${NC}"
echo -e "${YELLOW}Console output: /tmp/godot-console-$$.log${NC}"

# Wait for DAP server to be ready
if ! wait_for_port $DAP_PORT $GODOT_STARTUP_TIMEOUT; then
    echo -e "${RED}✗ Failed to start Godot DAP server${NC}"
    echo ""
    echo -e "${YELLOW}Godot console output (last 20 lines):${NC}"
    tail -n 20 "/tmp/godot-console-$$.log"
    exit 1
fi

# Step 2: Run test-dap-protocol in automated mode
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 2: Running test-dap-protocol${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Use 'yes' to automatically press ENTER for all prompts
# Redirect output to both stdout and file
echo -e "${YELLOW}Running compliance tests (automated mode)...${NC}"
echo ""

if yes "" | go run cmd/test-dap-protocol/main.go 2>&1 | tee "$OUTPUT_FILE"; then
    TEST_EXIT_CODE=0
else
    TEST_EXIT_CODE=$?
fi

# Step 3: Analyze Godot console for Dictionary errors
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 3: Analyzing Results${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Check for Dictionary errors in Godot console
DICT_ERRORS=$(grep -c "Dictionary::operator\[\]" "/tmp/godot-console-$$.log" || true)

if [ $DICT_ERRORS -gt 0 ]; then
    echo -e "${RED}✗ Found $DICT_ERRORS Dictionary error(s) in Godot console${NC}"
    echo ""
    echo -e "${YELLOW}Dictionary errors:${NC}"
    grep -A 2 "Dictionary::operator\[\]" "/tmp/godot-console-$$.log" || true
    echo ""
    EXIT_CODE=1
else
    echo -e "${GREEN}✓ No Dictionary errors found in Godot console${NC}"
    echo ""
    EXIT_CODE=0
fi

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "  Test Output: $OUTPUT_FILE"
echo "  Godot Console: /tmp/godot-console-$$.log"
echo "  Dictionary Errors: $DICT_ERRORS"
echo ""

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed - Godot is DAP spec-compliant${NC}"
else
    echo -e "${RED}✗ Tests failed - Dictionary errors detected${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "  1. Review output files above"
    echo "  2. Compare against upstream Godot to see which functions need fixes"
    echo "  3. Use findings to create GitHub issues per docs/godot-upstream/STRATEGY.md"
fi

echo ""

exit $EXIT_CODE
