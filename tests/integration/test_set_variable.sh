#!/bin/bash
set -e

# Setup
GODOT_BIN=${GODOT_BIN:-"/Applications/Godot.app/Contents/MacOS/Godot"}
PROJECT_PATH="tests/fixtures/test-project/project.godot"
DAP_PORT=6006
SERVER_BIN="./godot-dap-mcp-server"

# Start Godot
echo "Starting Godot..."
# Use port 6007 for the remote debug server (game -> editor connection)
# DAP server will listen on default port 6006 (configured in project.godot or defaults)
"$GODOT_BIN" --path "tests/fixtures/test-project" --headless --editor --debug-server tcp://127.0.0.1:6007 > /tmp/godot.log 2>&1 &
GODOT_PID=$!

# Cleanup trap
cleanup() {
    echo "Cleaning up..."
    kill $GODOT_PID 2>/dev/null || true
    rm -f /tmp/godot.log
    rm -f /tmp/mcp-stdin /tmp/mcp-stdout /tmp/mcp-stderr
}
trap cleanup EXIT

# Wait for Godot
echo "Waiting for Godot DAP..."
for i in {1..30}; do
    if nc -z 127.0.0.1 $DAP_PORT 2>/dev/null; then
        echo "Godot ready."
        sleep 5  # Wait for DAP server to be fully responsive
        break
    fi
    sleep 0.5
done

# Setup Named Pipes
rm -f /tmp/mcp-stdin /tmp/mcp-stdout
mkfifo /tmp/mcp-stdin /tmp/mcp-stdout

# Start Server
"$SERVER_BIN" < /tmp/mcp-stdin > /tmp/mcp-stdout 2>&1 &
SERVER_PID=$!

# Open FDs
exec 3> /tmp/mcp-stdin
exec 4< /tmp/mcp-stdout

# Helper to send and wait
send_request() {
    local req="$1"
    echo "Sending: $req"
    echo "$req" >&3
    
    # Read line by line until we get a response (which starts with {)
    # Note: Server might output logs first. We assume response is a JSON object on a single line.
    # This is a simplification; robust parsing would buffer.
    # But MCP server outputs logs to stderr (Wait, I redirected 2>&1 to stdout above? Yes.)
    # So logs are mixed.
    # I should separate stderr logs from stdout response.
}

# Restart Server with stderr separation
exec 3>&-
exec 4<&-
kill $SERVER_PID 2>/dev/null || true
rm -f /tmp/mcp-stdin /tmp/mcp-stdout
mkfifo /tmp/mcp-stdin /tmp/mcp-stdout

"$SERVER_BIN" < /tmp/mcp-stdin > /tmp/mcp-stdout 2> /tmp/mcp-stderr &
SERVER_PID=$!
exec 3> /tmp/mcp-stdin
exec 4< /tmp/mcp-stdout

echo "Server started (PID: $SERVER_PID)"

# Send Connect
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"godot_connect","arguments":{"port":6006}}}' >&3
echo "Sent connect request..."

# Read response (blocking)
read -r RESPONSE <&4
echo "Received: $RESPONSE"

if [[ "$RESPONSE" != *"success"* && "$RESPONSE" != *"connected"* ]]; then
    echo "Connect failed? $RESPONSE"
    # Continue anyway to check set_variable error
fi

# Send Set Variable
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"godot_set_variable","arguments":{"variable_name":"sum","value":100}}}' >&3
echo "Sent set_variable request..."

# Read response
read -r RESPONSE_SET <&4
echo "Received: $RESPONSE_SET"

# Check Output
if echo "$RESPONSE_SET" | grep -q "godot_set_variable is currently unavailable"; then
    echo "SUCCESS: godot_set_variable correctly returned unavailability error."
else
    echo "FAILURE: godot_set_variable did NOT return expected error."
    echo "Response: $RESPONSE_SET"
    exit 1
fi
