#!/bin/bash
# Build and install sage to GOBIN (or default Go bin path)
set -euo pipefail

cd "$(dirname "$0")/.."

# Get version from git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

echo "Installing sage ${VERSION}..."
go install -ldflags "-X 'github.com/not-emily/sage/internal/cli.Version=${VERSION}'" ./cmd/sage

echo "Done: sage installed (${VERSION})"
