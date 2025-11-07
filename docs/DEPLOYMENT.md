# Godot DAP MCP Server - Deployment Guide

**Last Updated**: 2025-11-07

This document describes how to build, distribute, and deploy the godot-dap-mcp-server for end users.

---

## Table of Contents

1. [Build Process](#build-process)
2. [Cross-Platform Compilation](#cross-platform-compilation)
3. [Installation Methods](#installation-methods)
4. [Claude Code Integration](#claude-code-integration)
5. [GitHub Releases](#github-releases)
6. [Distribution Channels](#distribution-channels)

---

## Build Process

### Development Build

Quick build for local development:

```bash
# Build for current platform
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Test the binary
./godot-dap-mcp-server
# Should start and wait for stdin input (Ctrl+C to exit)
```

### Production Build

Build with version information:

```bash
# Get version from git
VERSION=$(git describe --tags --always)

# Build with version info
go build \
  -ldflags="-X main.Version=${VERSION}" \
  -o godot-dap-mcp-server \
  cmd/godot-dap-mcp-server/main.go
```

---

## Cross-Platform Compilation

### Automated Build Script

Location: `scripts/build.sh`

```bash
#!/bin/bash

set -e

VERSION=$(git describe --tags --always)
BUILD_DIR="build"

echo "Building godot-dap-mcp-server version ${VERSION}"

mkdir -p $BUILD_DIR

# macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build \
  -ldflags="-X main.Version=${VERSION}" \
  -o $BUILD_DIR/godot-dap-mcp-server-darwin-amd64 \
  cmd/godot-dap-mcp-server/main.go

# macOS (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build \
  -ldflags="-X main.Version=${VERSION}" \
  -o $BUILD_DIR/godot-dap-mcp-server-darwin-arm64 \
  cmd/godot-dap-mcp-server/main.go

# Linux (x86_64)
echo "Building for Linux (x86_64)..."
GOOS=linux GOARCH=amd64 go build \
  -ldflags="-X main.Version=${VERSION}" \
  -o $BUILD_DIR/godot-dap-mcp-server-linux-amd64 \
  cmd/godot-dap-mcp-server/main.go

# Windows (x86_64)
echo "Building for Windows (x86_64)..."
GOOS=windows GOARCH=amd64 go build \
  -ldflags="-X main.Version=${VERSION}" \
  -o $BUILD_DIR/godot-dap-mcp-server-windows-amd64.exe \
  cmd/godot-dap-mcp-server/main.go

echo ""
echo "Build complete: version ${VERSION}"
echo "Binaries created in ${BUILD_DIR}/"
ls -lh $BUILD_DIR/
```

### Manual Cross-Compilation

```bash
# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o godot-dap-mcp-server-darwin-amd64 cmd/godot-dap-mcp-server/main.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o godot-dap-mcp-server-darwin-arm64 cmd/godot-dap-mcp-server/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o godot-dap-mcp-server-linux-amd64 cmd/godot-dap-mcp-server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o godot-dap-mcp-server-windows-amd64.exe cmd/godot-dap-mcp-server/main.go
```

### Supported Platforms

| Platform | Architecture | Binary Name |
|----------|--------------|-------------|
| macOS | Intel (x86_64) | `godot-dap-mcp-server-darwin-amd64` |
| macOS | Apple Silicon (ARM64) | `godot-dap-mcp-server-darwin-arm64` |
| Linux | x86_64 | `godot-dap-mcp-server-linux-amd64` |
| Windows | x86_64 | `godot-dap-mcp-server-windows-amd64.exe` |

---

## Installation Methods

Unlike npm or Python MCP servers that can be installed on-the-fly, Go binary MCP servers require a two-step process:
1. Obtain the binary (download or build)
2. Register with Claude Code

### Prerequisites

- Godot 4.0+ with DAP enabled
- Claude Code CLI installed
- Terminal/command line access

### Option A: Install from GitHub Releases (Recommended)

**For macOS (Apple Silicon):**

```bash
# 1. Create directory for MCP servers
mkdir -p ~/.claude/mcp-servers/godot-dap/

# 2. Download binary
curl -L -o ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server \
  https://github.com/TransitionMatrix/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-darwin-arm64

# 3. Make executable
chmod +x ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server

# 4. Register with Claude Code
claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

**For macOS (Intel):**

```bash
mkdir -p ~/.claude/mcp-servers/godot-dap/

curl -L -o ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server \
  https://github.com/TransitionMatrix/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-darwin-amd64

chmod +x ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server

claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

**For Linux:**

```bash
mkdir -p ~/.claude/mcp-servers/godot-dap/

curl -L -o ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server \
  https://github.com/TransitionMatrix/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-linux-amd64

chmod +x ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server

claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

**For Windows (PowerShell):**

```powershell
# 1. Create directory
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.claude\mcp-servers\godot-dap"

# 2. Download binary
Invoke-WebRequest `
  -Uri "https://github.com/TransitionMatrix/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-windows-amd64.exe" `
  -OutFile "$env:USERPROFILE\.claude\mcp-servers\godot-dap\godot-dap-mcp-server.exe"

# 3. Register with Claude Code
claude mcp add godot-dap "$env:USERPROFILE\.claude\mcp-servers\godot-dap\godot-dap-mcp-server.exe"
```

### Option B: Build from Source

```bash
# 1. Clone repository
git clone https://github.com/TransitionMatrix/godot-dap-mcp-server.git
cd godot-dap-mcp-server

# 2. Install dependencies
go mod download

# 3. Build binary
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# 4. Create installation directory and move binary
mkdir -p ~/.claude/mcp-servers/godot-dap/
mv godot-dap-mcp-server ~/.claude/mcp-servers/godot-dap/

# 5. Register with Claude Code
claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

### Option C: System-Wide Installation

Install to system PATH for convenience:

```bash
# 1. Download or build binary (see above)

# 2. Move to system location
sudo mv godot-dap-mcp-server /usr/local/bin/

# 3. Make executable (if needed)
sudo chmod +x /usr/local/bin/godot-dap-mcp-server

# 4. Register with Claude Code
claude mcp add godot-dap /usr/local/bin/godot-dap-mcp-server
```

### Verify Installation

```bash
# Check MCP server is registered
claude mcp list

# Test the server can be spawned
~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
# Should start and wait for stdin input (Ctrl+C to exit)
```

---

## Claude Code Integration

### Manual Configuration

If `claude mcp add` doesn't work, manually edit `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "godot-dap": {
      "command": "/Users/username/.claude/mcp-servers/godot-dap/godot-dap-mcp-server",
      "args": [],
      "env": {
        "GODOT_DAP_DEBUG": "false",
        "GODOT_DAP_LOG_FILE": ""
      }
    }
  }
}
```

**Important**: Replace `/Users/username/` with your actual home directory path.

### Configuration Options

**Environment Variables:**

| Variable | Description | Default |
|----------|-------------|---------|
| `GODOT_DAP_DEBUG` | Enable debug logging | `false` |
| `GODOT_DAP_LOG_FILE` | Log file path (if set, logs to file instead of stderr) | `""` (stderr) |
| `GODOT_DAP_TIMEOUT` | Default command timeout in seconds | `30` |

**Example with debug logging:**

```json
{
  "mcpServers": {
    "godot-dap": {
      "command": "/Users/username/.claude/mcp-servers/godot-dap/godot-dap-mcp-server",
      "args": [],
      "env": {
        "GODOT_DAP_DEBUG": "true",
        "GODOT_DAP_LOG_FILE": "/tmp/godot-dap-mcp.log"
      }
    }
  }
}
```

### Testing Integration

```bash
# In Claude Code:
claude

# Type:
use godot-dap

# Test connection:
godot_connect(port=6006)
```

---

## GitHub Releases

### Automated Release Process

Location: `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest

    permissions:
      contents: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build binaries
        run: ./scripts/build.sh

      - name: Create checksums
        run: |
          cd build
          sha256sum * > SHA256SUMS
          cd ..

      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            build/*
          body_path: CHANGELOG.md
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Creating a Release

```bash
# 1. Update CHANGELOG.md with release notes

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. GitHub Actions will automatically:
#    - Build binaries for all platforms
#    - Create checksums
#    - Create GitHub release
#    - Upload binaries to release
```

### Manual Release

If GitHub Actions isn't available:

```bash
# 1. Build binaries
./scripts/build.sh

# 2. Create checksums
cd build
sha256sum * > SHA256SUMS
cd ..

# 3. Create release on GitHub manually
#    - Go to GitHub → Releases → Draft a new release
#    - Create tag: v1.0.0
#    - Upload all files from build/
#    - Add release notes from CHANGELOG.md
#    - Publish release
```

---

## Distribution Channels

### GitHub Releases (Primary)

- **URL**: https://github.com/TransitionMatrix/godot-dap-mcp-server/releases
- **Audience**: All users
- **Update frequency**: On version tags

### Homebrew (Future)

For easier macOS installation:

```ruby
# Formula: godot-dap-mcp-server.rb
class GodotDapMcpServer < Formula
  desc "MCP server for debugging Godot games via DAP"
  homepage "https://github.com/TransitionMatrix/godot-dap-mcp-server"
  url "https://github.com/TransitionMatrix/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-darwin-arm64"
  sha256 "..."
  version "1.0.0"

  def install
    bin.install "godot-dap-mcp-server-darwin-arm64" => "godot-dap-mcp-server"
  end

  test do
    system "#{bin}/godot-dap-mcp-server", "--version"
  end
end
```

Users would install with:

```bash
brew tap TransitionMatrix/godot-tools
brew install godot-dap-mcp-server
```

### Go Install (Future)

For Go developers:

```bash
go install github.com/TransitionMatrix/godot-dap-mcp-server/cmd/godot-dap-mcp-server@latest
```

Requires adding to `~/.claude/mcp.json`:

```bash
claude mcp add godot-dap $(go env GOPATH)/bin/godot-dap-mcp-server
```

---

## Binary Sizes

Expected sizes (approximate):

| Platform | Size | Compressed (gzip) |
|----------|------|-------------------|
| macOS (Intel) | ~8 MB | ~2.5 MB |
| macOS (ARM64) | ~8 MB | ~2.5 MB |
| Linux (x86_64) | ~8 MB | ~2.5 MB |
| Windows (x86_64) | ~8 MB | ~2.5 MB |

**Note**: Go produces static binaries with no runtime dependencies, resulting in larger file sizes but simpler deployment.

---

## Troubleshooting Deployment

### Binary Won't Execute

**macOS: "cannot be opened because the developer cannot be verified"**

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server

# Or allow in System Preferences:
# System Preferences → Security & Privacy → General → "Allow anyway"
```

**Linux: "Permission denied"**

```bash
# Make executable
chmod +x ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

### Claude Code Can't Find Binary

```bash
# Verify path is correct
ls -l ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server

# Check mcp.json configuration
cat ~/.claude/mcp.json

# Test binary directly
~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
# Should start and wait for input
```

### Wrong Architecture

**Error**: "exec format error" or "cannot execute binary file"

**Solution**: Download correct binary for your platform:

```bash
# Check your platform
uname -sm

# macOS ARM64: Darwin arm64
# macOS Intel: Darwin x86_64
# Linux: Linux x86_64
```

---

## References

- [ARCHITECTURE.md](ARCHITECTURE.md) - System design
- [TESTING.md](TESTING.md) - Testing procedures before release
- GitHub Actions Documentation: https://docs.github.com/en/actions
