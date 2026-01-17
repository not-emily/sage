#!/bin/bash
# Build sage binaries for all platforms and create a release
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION=${1:-$(git describe --tags --always)}
OUTDIR="dist"

echo "Building sage ${VERSION}..."

rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

# Build for each platform
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for platform in "${platforms[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"
    output="$OUTDIR/sage-${GOOS}-${GOARCH}"

    echo "  Building ${GOOS}/${GOARCH}..."
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X 'github.com/not-emily/sage/internal/cli.Version=${VERSION}'" \
        -o "$output" \
        ./cmd/sage
done

echo ""
echo "Binaries built in $OUTDIR/"
ls -la "$OUTDIR/"

echo ""
echo "To create a GitHub release:"
echo "  gh release create ${VERSION} ${OUTDIR}/* --title \"${VERSION}\" --notes \"Release ${VERSION}\""
