#!/bin/bash
# Build the sage CLI binary
set -euo pipefail

cd "$(dirname "$0")/.."

echo "Building sage..."
go build -o bin/sage ./cmd/sage

echo "Done: bin/sage"
