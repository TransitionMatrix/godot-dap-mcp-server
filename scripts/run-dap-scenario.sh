#!/usr/bin/env bash
#
# Wrapper script to run a specific DAP scenario against a Godot instance
#
# This script:
# 1. Starts Godot in the background with DAP enabled (headless editor)
# 2. Waits for DAP server to be ready
# 3. Runs the test-dap-runner with the specified scenario file
# 4. Cleans up the Godot process on exit
#
# Usage:
#   ./scripts/run-dap-scenario.sh <path_to_scenario_file>
#
# Environment variables:
#   GODOT_BIN - Path to Godot binary (required)
#   PROJECT_PATH - Path to project.godot (default: tests/fixtures/test-project/project.godot)
#   DAP_PORT - DAP server port (default: 6006)
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check arguments
if [ $# -ne 1 ]; then
    echo -e "${RED}Usage: $0 <path_to_scenario_file>${NC}"
    exit 1
fi

SCENARIO_FILE="$1"

# Configuration
DAP_PORT="${DAP_PORT:-6006}"
GODOT_STARTUP_TIMEOUT=15
PROJECT_PATH="${PROJECT_PATH:-tests/fixtures/test-project/project.godot}"

# Validate Scenario File
if [ ! -f "$SCENARIO_FILE" ]; then
    echo -e "${RED}✗ Scenario file not found: $SCENARIO_FILE${NC}"
    exit 1
fi

# Validate GODOT_BIN
if [ -z "${GODOT_BIN:-}" ]; then
    echo -e "${RED}✗ GODOT_BIN environment variable not set${NC}"
    echo "Set it to the path of your Godot binary:"
    echo "  export GODOT_BIN=/path/to/godot.macos.editor.arm64"
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
}

# Register cleanup on exit
trap cleanup EXIT INT TERM

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

# Helper: Check if port is in use
is_port_in_use() {
    local port=$1
    nc -z 127.0.0.1 $port 2>/dev/null
}

# Check if port is already in use
if is_port_in_use $DAP_PORT; then
    echo -e "${RED}✗ Port $DAP_PORT is already in use${NC}"
    echo "The wrapper expects to start its own Godot instance."
    echo "Please stop the existing process or use a different port."
    exit 1
fi

# Start Godot
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 1: Starting Godot Editor${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Create a temporary editor settings override to enable DAP
EDITOR_SETTINGS_OVERRIDE="/tmp/godot-dap-runner-settings-$(date +%s).tres"

cat > "$EDITOR_SETTINGS_OVERRIDE" <<EOF
[gd_resource type="EditorSettings" format=3]

[resource]
network/debug_adapter/remote_port = $DAP_PORT
network/debug_adapter/request_timeout = 1000
network/debug_adapter/sync_breakpoints = true
network/debug/remote_port = 6007
EOF

echo -e "${YELLOW}Starting Godot with DAP server on port $DAP_PORT...${NC}"

"$GODOT_BIN" \
    --editor \
    --path "$PROJECT_DIR" \
    --headless \
    > "/tmp/godot-console-runner-$$.log" 2>&1 &

GODOT_PID=$!
echo -e "${GREEN}✓ Godot started (PID: $GODOT_PID)${NC}"

# Wait for DAP server
if ! wait_for_port $DAP_PORT $GODOT_STARTUP_TIMEOUT; then
    echo -e "${RED}✗ Failed to start Godot DAP server${NC}"
    tail -n 20 "/tmp/godot-console-runner-$$.log"
    exit 1
fi

# Special handling for 'attach' scenarios: Start the game instance manually
if [[ "$SCENARIO_FILE" == *"attach"* ]]; then
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Step 1.5: Starting Game Instance (for Attach test)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Use --breakpoints to pause the game immediately so we can attach
    # Using line 4 in test_script.gd (inside _ready)
    GAME_BREAKPOINT="res://test_script.gd:4"
    DEBUG_URI="tcp://127.0.0.1:6007"
    
    echo -e "${YELLOW}Launching game with breakpoint at $GAME_BREAKPOINT...${NC}"
    echo -e "${YELLOW}Connecting to debugger at $DEBUG_URI...${NC}"
    
    "$GODOT_BIN" \
        --path "$PROJECT_DIR" \
        --breakpoints "$GAME_BREAKPOINT" \
        --remote-debug "$DEBUG_URI" \
        > "/tmp/godot-game-$$.log" 2>&1 &
        
    GAME_PID=$!
    echo -e "${GREEN}✓ Game started (PID: $GAME_PID)${NC}"
    
    # Register game for cleanup
    trap 'kill $GAME_PID 2>/dev/null || true; cleanup' EXIT INT TERM
    
    # Give the game a moment to connect to the editor
    sleep 2
fi

# Run Runner
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 2: Running Scenario: $SCENARIO_FILE${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Run interactively
go run cmd/test-dap-runner/main.go "$SCENARIO_FILE"

EXIT_CODE=$?
echo ""
exit $EXIT_CODE
