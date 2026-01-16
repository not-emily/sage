#!/bin/bash
# Build the sage CLI binary
set -euo pipefail

cd "$(dirname "$0")/.."

# Get version from git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

echo "Building sage ${VERSION}..."
go build -ldflags "-X 'github.com/not-emily/sage/internal/cli.Version=${VERSION}'" -o bin/sage ./cmd/sage

echo "Done: bin/sage (${VERSION})"
