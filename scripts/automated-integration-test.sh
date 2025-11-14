#!/bin/bash

# Automated Integration Test for godot-dap-mcp-server
# This script launches Godot as a subprocess and runs integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Godot DAP MCP Server${NC}"
echo -e "${BLUE}Automated Integration Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Configuration
DAP_PORT_DEFAULT=6006
DAP_PORT_RANGE_START=6006
DAP_PORT_RANGE_END=6020
GODOT_STARTUP_TIMEOUT=10  # seconds to wait for Godot to start
GODOT_BIN="${GODOT_BIN:-godot}"  # Allow override via env var

# Function to check if a port is in use
is_port_in_use() {
    local port=$1
    nc -z 127.0.0.1 $port 2>/dev/null
    return $?
}

# Function to find an available port in range
find_available_port() {
    local start=$1
    local end=$2

    for port in $(seq $start $end); do
        if ! is_port_in_use $port; then
            echo $port
            return 0
        fi
    done

    return 1
}

# Check if godot-dap-mcp-server binary exists
if [ ! -f "./godot-dap-mcp-server" ]; then
    echo -e "${RED}✗ Error: godot-dap-mcp-server binary not found${NC}"
    echo "Please run: go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go"
    exit 1
fi
echo -e "${GREEN}✓ Found godot-dap-mcp-server binary${NC}"

# Check if Godot is installed
if ! command -v "$GODOT_BIN" &> /dev/null; then
    echo -e "${RED}✗ Error: Godot not found in PATH${NC}"
    echo ""
    echo "Please install Godot 4.2.2+ or set GODOT_BIN environment variable:"
    echo "  export GODOT_BIN=/path/to/godot"
    echo ""
    echo "Download from: https://godotengine.org/download"
    exit 1
fi

# Check Godot version
GODOT_VERSION=$("$GODOT_BIN" --version 2>&1 | head -1 || echo "unknown")
echo -e "${GREEN}✓ Found Godot: ${GODOT_VERSION}${NC}"

# Get absolute path to test project
TEST_PROJECT="$(cd "$(dirname "$0")/../tests/fixtures/test-project" && pwd)"

if [ ! -f "$TEST_PROJECT/project.godot" ]; then
    echo -e "${RED}✗ Error: Test project not found at $TEST_PROJECT${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Found test project: $TEST_PROJECT${NC}"
echo ""

# Check if default port is available, or find an alternative
echo -e "${YELLOW}Checking port availability...${NC}"
if is_port_in_use $DAP_PORT_DEFAULT; then
    echo -e "${YELLOW}⚠ Port $DAP_PORT_DEFAULT is already in use${NC}"
    echo "  (Godot editor may already be running with DAP enabled)"

    echo -e "${YELLOW}Finding next available port in range ${DAP_PORT_RANGE_START}-${DAP_PORT_RANGE_END}...${NC}"
    DAP_PORT=$(find_available_port $DAP_PORT_RANGE_START $DAP_PORT_RANGE_END)

    if [ -z "$DAP_PORT" ]; then
        echo -e "${RED}✗ Error: No available ports in range ${DAP_PORT_RANGE_START}-${DAP_PORT_RANGE_END}${NC}"
        echo ""
        echo "Please close any running Godot instances or processes using these ports."
        exit 1
    fi

    echo -e "${GREEN}✓ Found available port: $DAP_PORT${NC}"
else
    DAP_PORT=$DAP_PORT_DEFAULT
    echo -e "${GREEN}✓ Port $DAP_PORT is available${NC}"
fi
echo ""

# Function to check if port is available
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

# Function to send MCP request and capture response (using persistent server)
send_mcp_request() {
    local request="$1"

    # Send request via FD 3 (to MCP server stdin)
    echo "$request" >&3

    # Read response via FD 4 (from MCP server stdout)
    read -r response <&4
    echo "$response"
}

# Launch Godot in background
echo -e "${YELLOW}Launching Godot editor in headless mode...${NC}"
echo "  Command: $GODOT_BIN --headless --editor --path \"$TEST_PROJECT\" --dap-port $DAP_PORT"

"$GODOT_BIN" --headless --editor --path "$TEST_PROJECT" --dap-port $DAP_PORT > /tmp/godot-test.log 2>&1 &
GODOT_PID=$!

echo -e "${GREEN}✓ Godot started (PID: $GODOT_PID)${NC}"
echo ""

# Wait for Godot DAP server to be ready
if ! wait_for_port $DAP_PORT $GODOT_STARTUP_TIMEOUT; then
    echo -e "${RED}Failed to start Godot DAP server${NC}"
    echo ""
    echo "Godot log (last 20 lines):"
    tail -20 /tmp/godot-test.log || echo "No log available"
    exit 1
fi
echo ""

# Start MCP server with persistent stdin/stdout for session
echo -e "${YELLOW}Starting MCP server...${NC}"

# Create named pipes for input and output
rm -f /tmp/mcp-stdin /tmp/mcp-stdout
mkfifo /tmp/mcp-stdin
mkfifo /tmp/mcp-stdout

# Open pipes with file descriptors to keep them alive
exec 3<>/tmp/mcp-stdin   # FD 3 for writing requests
exec 4<>/tmp/mcp-stdout  # FD 4 for reading responses

# Start MCP server with piped stdin/stdout
./godot-dap-mcp-server </tmp/mcp-stdin >/tmp/mcp-stdout 2>/dev/null &
MCP_PID=$!

# Update trap to clean up all processes and files
trap "echo ''; echo -e '${YELLOW}Cleaning up...${NC}'; exec 3>&- 4>&-; kill $MCP_PID 2>/dev/null || true; kill $GODOT_PID 2>/dev/null || true; wait $MCP_PID 2>/dev/null || true; wait $GODOT_PID 2>/dev/null || true; rm -f /tmp/mcp-stdin /tmp/mcp-stdout; echo -e '${GREEN}✓ Cleanup complete${NC}'" EXIT

# Give MCP server a moment to initialize
sleep 0.2

# Verify MCP server is running
if ! kill -0 $MCP_PID 2>/dev/null; then
    echo -e "${RED}✗ MCP server died immediately after starting${NC}"
    exit 1
fi

echo -e "${GREEN}✓ MCP server started (PID: $MCP_PID)${NC}"
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Running Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Test 1: Initialize
echo -e "${YELLOW}Test 1: Initialize MCP server${NC}"
INIT_REQUEST='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0","capabilities":{},"clientInfo":{"name":"integration-test","version":"1.0"}}}'
INIT_RESPONSE=$(send_mcp_request "$INIT_REQUEST")

if echo "$INIT_RESPONSE" | grep -q '"result"'; then
    echo -e "${GREEN}✓ Initialize successful${NC}"
else
    echo -e "${RED}✗ Initialize failed${NC}"
    echo "Response: $INIT_RESPONSE"
    exit 1
fi
echo ""

# Test 2: List tools
echo -e "${YELLOW}Test 2: List available tools${NC}"
TOOLS_REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
TOOLS_RESPONSE=$(send_mcp_request "$TOOLS_REQUEST")

if echo "$TOOLS_RESPONSE" | grep -q 'godot_connect'; then
    echo -e "${GREEN}✓ Tools list successful${NC}"
else
    echo -e "${RED}✗ godot_connect not found in tools list${NC}"
    exit 1
fi

# Check for Phase 3 tools (Core Debugging)
PHASE3_TOOLS=("godot_connect" "godot_disconnect" "godot_continue" "godot_step_over" "godot_step_into" "godot_set_breakpoint" "godot_clear_breakpoint")
echo "  Phase 3 (Core Debugging):"
for tool in "${PHASE3_TOOLS[@]}"; do
    if echo "$TOOLS_RESPONSE" | grep -q "$tool"; then
        echo -e "${GREEN}    ✓ $tool${NC}"
    else
        echo -e "${RED}    ✗ $tool missing${NC}"
        exit 1
    fi
done

# Check for Phase 4 tools (Runtime Inspection)
PHASE4_TOOLS=("godot_get_threads" "godot_get_stack_trace" "godot_get_scopes" "godot_get_variables" "godot_evaluate")
echo "  Phase 4 (Runtime Inspection):"
for tool in "${PHASE4_TOOLS[@]}"; do
    if echo "$TOOLS_RESPONSE" | grep -q "$tool"; then
        echo -e "${GREEN}    ✓ $tool${NC}"
    else
        echo -e "${RED}    ✗ $tool missing${NC}"
        exit 1
    fi
done

# Check for Phase 6 tools (Advanced Debugging)
PHASE6_TOOLS=("godot_pause" "godot_set_variable")
echo "  Phase 6 (Advanced Debugging):"
for tool in "${PHASE6_TOOLS[@]}"; do
    if echo "$TOOLS_RESPONSE" | grep -q "$tool"; then
        echo -e "${GREEN}    ✓ $tool${NC}"
    else
        echo -e "${RED}    ✗ $tool missing${NC}"
        exit 1
    fi
done
echo ""

# Test 3: Connect to Godot DAP server
echo -e "${YELLOW}Test 3: Connect to Godot DAP server${NC}"
CONNECT_REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"godot_connect","arguments":{"port":'$DAP_PORT'}}}'
CONNECT_RESPONSE=$(send_mcp_request "$CONNECT_REQUEST")

# Check for "status":"connected" (handles both escaped and unescaped JSON)
if echo "$CONNECT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")connected'; then
    echo -e "${GREEN}✓ Connection successful${NC}"
    # Extract state from response (handles escaped JSON)
    STATE=$(echo "$CONNECT_RESPONSE" | grep -oE '(\\"|")state(\\"|"):(\\"|")[^"\\]*' | tail -1 | sed 's/.*://' | tr -d '"\\')
    if [ -n "$STATE" ]; then
        echo "  State: $STATE"
    fi
elif echo "$CONNECT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")already_connected'; then
    echo -e "${GREEN}✓ Already connected (reusing session)${NC}"
else
    echo -e "${RED}✗ Connection failed${NC}"
    echo "Response: $CONNECT_RESPONSE"
    exit 1
fi
echo ""

# Test 4: Set breakpoint
echo -e "${YELLOW}Test 4: Set breakpoint${NC}"
# Use absolute path (Godot DAP requires absolute paths, not res:// paths)
TEST_SCRIPT_PATH="$TEST_PROJECT/test_script.gd"
BREAKPOINT_REQUEST="{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"godot_set_breakpoint\",\"arguments\":{\"file\":\"$TEST_SCRIPT_PATH\",\"line\":13}}}"
BREAKPOINT_RESPONSE=$(send_mcp_request "$BREAKPOINT_REQUEST")

if echo "$BREAKPOINT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")verified'; then
    echo -e "${GREEN}✓ Breakpoint set and verified${NC}"
    echo "  Location: $TEST_SCRIPT_PATH:13"
elif echo "$BREAKPOINT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")unverified'; then
    echo -e "${YELLOW}⚠ Breakpoint set but not verified${NC}"
    echo "  (This is normal - file loads when scene runs)"
else
    echo -e "${RED}✗ Failed to set breakpoint${NC}"
    echo "Response: $BREAKPOINT_RESPONSE"
    exit 1
fi
echo ""

# Test 5: Clear breakpoint
echo -e "${YELLOW}Test 5: Clear breakpoint${NC}"
CLEAR_REQUEST="{\"jsonrpc\":\"2.0\",\"id\":5,\"method\":\"tools/call\",\"params\":{\"name\":\"godot_clear_breakpoint\",\"arguments\":{\"file\":\"$TEST_SCRIPT_PATH\"}}}"
CLEAR_RESPONSE=$(send_mcp_request "$CLEAR_REQUEST")

if echo "$CLEAR_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")cleared'; then
    echo -e "${GREEN}✓ Breakpoint cleared${NC}"
else
    echo -e "${RED}✗ Failed to clear breakpoint${NC}"
    echo "Response: $CLEAR_RESPONSE"
    exit 1
fi
echo ""

# Test 6: Disconnect
echo -e "${YELLOW}Test 6: Disconnect from Godot${NC}"
DISCONNECT_REQUEST='{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"godot_disconnect","arguments":{}}}'
DISCONNECT_RESPONSE=$(send_mcp_request "$DISCONNECT_REQUEST")

if echo "$DISCONNECT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")disconnected'; then
    echo -e "${GREEN}✓ Disconnection successful${NC}"
else
    echo -e "${RED}✗ Disconnection failed${NC}"
    echo "Response: $DISCONNECT_RESPONSE"
    exit 1
fi
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}All Integration Tests Passed! ✓${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${GREEN}Summary:${NC}"
echo "  ✓ Godot launched as subprocess (headless mode)"
echo "  ✓ DAP server started automatically"
echo "  ✓ MCP server initialized"
echo "  ✓ All 7 Phase 3 tools registered (Core Debugging)"
echo "  ✓ All 5 Phase 4 tools registered (Runtime Inspection)"
echo "  ✓ All 2 Phase 6 tools registered (Advanced Debugging)"
echo "  ✓ Connected to Godot DAP server"
echo "  ✓ Set and cleared breakpoint"
echo "  ✓ Clean disconnection"
echo ""

echo -e "${YELLOW}Note:${NC} Some tools require an active debugging session:"
echo "  • Execution control (continue, step) - requires paused game at breakpoint"
echo "  • Runtime inspection (stack, variables, evaluate) - requires paused game"
echo "  • Advanced debugging (pause, set_variable) - requires running/paused game"
echo "  Run scene (F5) with breakpoint to test these interactively"
echo ""
