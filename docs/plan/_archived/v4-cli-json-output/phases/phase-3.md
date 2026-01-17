# Phase 3: Release Infrastructure

> **Depends on:** None (can run in parallel with Phases 1-2)
> **Enables:** Hub-core can install sage without Go
>
> See: [Full Plan](../plan.md)

## Goal

Establish cross-compilation and GitHub release process so sage binaries can be downloaded and run without Go installed on the target machine.

## Key Deliverables

- Cross-compile script for all target platforms
- GitHub release with binaries attached
- Version tagged releases (v0.4.0, etc.)

## Target Platforms

| OS | Architecture | Binary Name |
|----|--------------|-------------|
| Linux | amd64 | `sage-linux-amd64` |
| Linux | arm64 | `sage-linux-arm64` |
| macOS | amd64 | `sage-darwin-amd64` |
| macOS | arm64 | `sage-darwin-arm64` |

## Files to Create

- `scripts/release.sh` — Cross-compile and prepare release artifacts

## Dependencies

**Internal:** None

**External:**
- GitHub CLI (`gh`) for creating releases (optional, can do manually)
- Go toolchain for cross-compilation

## Implementation Notes

**Release script (scripts/release.sh):**

```bash
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
```

**Makefile integration (optional):**

```makefile
.PHONY: release
release:
	./scripts/release.sh $(VERSION)
```

**Release process:**

1. Ensure all changes committed
2. Tag the release: `git tag v0.4.0`
3. Push tag: `git push origin v0.4.0`
4. Run release script: `./scripts/release.sh v0.4.0`
5. Create GitHub release: `gh release create v0.4.0 dist/* --title "v0.4.0"`

**Version injection:**

The script uses `-ldflags` to inject version at build time, same as existing `scripts/build.sh` and `scripts/install.sh`.

**Binary verification:**

Users can verify binaries work:
```bash
curl -L https://github.com/not-emily/sage/releases/download/v0.4.0/sage-linux-amd64 -o sage
chmod +x sage
./sage version
# Should output: 0.4.0
```

## Validation

- [ ] `scripts/release.sh v0.4.0` creates binaries in `dist/`
- [ ] All 4 platform binaries are created
- [ ] Each binary runs and shows correct version
- [ ] Linux binary runs on Linux (test on station)
- [ ] macOS binary runs on macOS (if available)
- [ ] GitHub release can be created with binaries attached
- [ ] Binaries downloadable via curl
