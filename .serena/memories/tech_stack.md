# Tech Stack

## Language & Version
- **Go**: 1.25.3
- Single binary compilation, no runtime dependencies
- Cross-platform support: macOS (Intel/ARM64), Linux, Windows

## Core Dependencies
```
github.com/google/go-dap v0.12.0  # DAP protocol implementation
```

## Standard Library Usage
- `encoding/json` - JSON marshaling/unmarshaling
- `net` - TCP networking
- `context` - Timeout and cancellation
- `bufio` - Buffered I/O for stdio communication
- `io`, `os` - File and stdio operations

## Optional Dependencies (for future consideration)
- `github.com/spf13/cobra` - CLI commands
- `github.com/sirupsen/logrus` - Structured logging

## Build System
- Standard Go toolchain
- Custom build script: `scripts/build.sh` for cross-compilation
- Module-based dependency management: `go.mod`

## Target Platforms
- macOS: darwin/amd64, darwin/arm64
- Linux: linux/amd64, linux/arm64
- Windows: windows/amd64

## Binary Size
- Approximate size: 2.9MB (development build)
- Static binary with all dependencies included
