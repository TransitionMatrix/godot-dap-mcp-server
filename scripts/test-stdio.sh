#!/bin/bash

# Test script for manual stdio testing of the MCP server
# This script sends test requests to the server and shows responses

set -e

echo "Testing MCP Server via stdio..."
echo ""

# Start server in background and connect to its stdio
{
    echo "=== Test 1: Initialize ===" >&2
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
    sleep 0.5

    echo "=== Test 2: List Tools ===" >&2
    echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
    sleep 0.5

    echo "=== Test 3: Call godot_ping with default ===" >&2
    echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"godot_ping","arguments":{}}}'
    sleep 0.5

    echo "=== Test 4: Call godot_ping with message ===" >&2
    echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"godot_ping","arguments":{"message":"Hello from test!"}}}'
    sleep 0.5

    echo "=== Test 5: Invalid method ===" >&2
    echo '{"jsonrpc":"2.0","id":5,"method":"invalid/method","params":{}}'
    sleep 0.5

    echo "=== Test 6: Invalid tool ===" >&2
    echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"nonexistent_tool","arguments":{}}}'
    sleep 0.5

    # Send EOF to cleanly shutdown server
    echo "=== Sending EOF to shutdown server ===" >&2
} | ./godot-dap-mcp-server 2>&1

echo ""
echo "Tests complete!"
