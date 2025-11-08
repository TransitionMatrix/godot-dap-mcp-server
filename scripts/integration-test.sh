#!/bin/bash

# Integration test for godot-dap-mcp-server
# This script tests the MCP server with a running Godot editor

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Godot DAP MCP Server - Integration Test${NC}"
echo -e "${YELLOW}(Manual Setup Required)${NC}"
echo ""

# Configuration
DAP_PORT="${DAP_PORT:-6006}"  # Default to 6006, allow override via env var

# Check if godot-dap-mcp-server binary exists
if [ ! -f "./godot-dap-mcp-server" ]; then
    echo -e "${RED}Error: godot-dap-mcp-server binary not found${NC}"
    echo "Please run: go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go"
    exit 1
fi

# Get absolute path to test project
TEST_PROJECT="$(cd "$(dirname "$0")/../tests/fixtures/test-project" && pwd)"

echo -e "${YELLOW}Test Setup:${NC}"
echo "  Test project: $TEST_PROJECT"
echo "  DAP Port: $DAP_PORT (override with: export DAP_PORT=xxxx)"
echo ""

echo -e "${YELLOW}Prerequisites:${NC}"
echo "  1. Godot editor must be running"
echo "  2. DAP server must be enabled:"
echo "     Editor → Editor Settings → Network → Debug Adapter"
echo "     - Enable Debug Adapter Server"
echo "     - Remote Port = $DAP_PORT"
echo "  3. Test project should be opened in Godot: $TEST_PROJECT"
echo ""

read -p "Press Enter when Godot is ready, or Ctrl+C to cancel..."
echo ""

# Test project path (convert to absolute if needed)
TEST_SCRIPT="$TEST_PROJECT/test_script.gd"

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

# Set up cleanup trap
trap "echo ''; echo -e '${YELLOW}Cleaning up...${NC}'; exec 3>&- 4>&-; kill $MCP_PID 2>/dev/null || true; wait $MCP_PID 2>/dev/null || true; rm -f /tmp/mcp-stdin /tmp/mcp-stdout; echo -e '${GREEN}✓ Cleanup complete${NC}'" EXIT

# Give MCP server a moment to initialize
sleep 0.2

# Verify MCP server is running
if ! kill -0 $MCP_PID 2>/dev/null; then
    echo -e "${RED}✗ MCP server died immediately after starting${NC}"
    exit 1
fi

echo -e "${GREEN}✓ MCP server started (PID: $MCP_PID)${NC}"
echo ""

echo -e "${GREEN}Starting integration tests...${NC}"
echo ""

# Function to send MCP request and capture response (using persistent server)
send_mcp_request() {
    local request="$1"
    # Send request via FD 3 (to MCP server stdin)
    echo "$request" >&3
    # Read response via FD 4 (from MCP server stdout)
    read -r response <&4
    echo "$response"
}

# Test 1: Initialize
echo -e "${YELLOW}Test 1: Initialize MCP server${NC}"
INIT_REQUEST='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0","capabilities":{},"clientInfo":{"name":"integration-test","version":"1.0"}}}'
INIT_RESPONSE=$(send_mcp_request "$INIT_REQUEST")
echo "Response: $INIT_RESPONSE"

if echo "$INIT_RESPONSE" | grep -q '"result"'; then
    echo -e "${GREEN}✓ Initialize successful${NC}"
else
    echo -e "${RED}✗ Initialize failed${NC}"
    exit 1
fi
echo ""

# Test 2: List tools
echo -e "${YELLOW}Test 2: List available tools${NC}"
TOOLS_REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
TOOLS_RESPONSE=$(send_mcp_request "$TOOLS_REQUEST")
echo "Response (truncated): $(echo "$TOOLS_RESPONSE" | cut -c1-200)..."

if echo "$TOOLS_RESPONSE" | grep -q 'godot_connect'; then
    echo -e "${GREEN}✓ Tools list includes godot_connect${NC}"
else
    echo -e "${RED}✗ godot_connect not found in tools list${NC}"
    exit 1
fi

# Check for all Phase 3 tools
for tool in "godot_connect" "godot_disconnect" "godot_continue" "godot_step_over" "godot_step_into" "godot_set_breakpoint" "godot_clear_breakpoint"; do
    if echo "$TOOLS_RESPONSE" | grep -q "$tool"; then
        echo -e "${GREEN}  ✓ $tool${NC}"
    else
        echo -e "${RED}  ✗ $tool missing${NC}"
        exit 1
    fi
done
echo ""

# Test 3: Connect to Godot DAP server
echo -e "${YELLOW}Test 3: Connect to Godot DAP server${NC}"
CONNECT_REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"godot_connect","arguments":{"port":'$DAP_PORT'}}}'
CONNECT_RESPONSE=$(send_mcp_request "$CONNECT_REQUEST")
echo "Response: $CONNECT_RESPONSE"

# Check for "status":"connected" (handles both escaped and unescaped JSON)
if echo "$CONNECT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")connected'; then
    echo -e "${GREEN}✓ Connection successful${NC}"
elif echo "$CONNECT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")already_connected'; then
    echo -e "${GREEN}✓ Already connected (reusing session)${NC}"
else
    echo -e "${RED}✗ Connection failed${NC}"
    echo "Make sure Godot editor is running with DAP server enabled on port $DAP_PORT"
    exit 1
fi
echo ""

# Test 4: Set breakpoint
echo -e "${YELLOW}Test 4: Set breakpoint${NC}"
# Use absolute path (Godot DAP requires absolute paths, not res:// paths)
TEST_SCRIPT_PATH="$TEST_PROJECT/test_script.gd"
BREAKPOINT_REQUEST="{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"godot_set_breakpoint\",\"arguments\":{\"file\":\"$TEST_SCRIPT_PATH\",\"line\":13}}}"
BREAKPOINT_RESPONSE=$(send_mcp_request "$BREAKPOINT_REQUEST")
echo "Response: $BREAKPOINT_RESPONSE"

# Check for "status":"verified" (handles both escaped and unescaped JSON)
if echo "$BREAKPOINT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")verified'; then
    echo -e "${GREEN}✓ Breakpoint set and verified${NC}"
elif echo "$BREAKPOINT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")unverified'; then
    echo -e "${YELLOW}⚠ Breakpoint set but not verified (file may not be loaded yet)${NC}"
else
    echo -e "${RED}✗ Failed to set breakpoint${NC}"
    exit 1
fi
echo ""

# Test 5: Disconnect
echo -e "${YELLOW}Test 5: Disconnect from Godot${NC}"
DISCONNECT_REQUEST='{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"godot_disconnect","arguments":{}}}'
DISCONNECT_RESPONSE=$(send_mcp_request "$DISCONNECT_REQUEST")
echo "Response: $DISCONNECT_RESPONSE"

# Check for "status":"disconnected" (handles both escaped and unescaped JSON)
if echo "$DISCONNECT_RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")disconnected'; then
    echo -e "${GREEN}✓ Disconnection successful${NC}"
else
    echo -e "${RED}✗ Disconnection failed${NC}"
    exit 1
fi
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}All integration tests passed! ✓${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Open test project in Godot: $TEST_PROJECT"
echo "  2. Press F5 to run the scene"
echo "  3. Manually test stepping commands when breakpoint hits"
echo ""
