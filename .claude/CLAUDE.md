# Claude Context - sage

## Overview
Standalone CLI and Go library for unified LLM provider access. Manages providers (OpenAI, Anthropic, Ollama), profiles, and encrypted credential storage.

## Constraints
- **Go stdlib only** — no third-party dependencies
- **JSON for config** — stdlib has no YAML parser
- **AES-256-GCM** for credential encryption

## Key Patterns
- `pkg/sage/` — Public library API (importable by hub-core)
- `internal/` — CLI internals, not exported
- `cmd/sage/` — CLI entrypoint
- Config location: `~/.config/sage/`

## Development
```bash
./scripts/build.sh    # Build binary to bin/sage
go test ./...         # Run tests
go build ./pkg/sage/  # Build library only
```

## Project Structure
```
sage/
├── cmd/sage/           # CLI binary
├── pkg/sage/           # Public library
│   └── providers/      # Provider implementations
├── internal/cli/       # CLI internals
├── scripts/            # Build/test scripts
└── docs/               # Documentation & plan
```

## Helper Scripts
Scripts in `scripts/` are reusable helpers. **Before writing repetitive bash commands:**
1. Check if a script already exists in `scripts/`
2. If not, consider creating one for sequences you'll run again

This reduces permission prompts and ensures consistency.

Available scripts:
- `scripts/build.sh` — Build the sage CLI binary to `bin/sage`
- `scripts/release.sh [version]` — Cross-compile for all platforms to `dist/`

## Release Process
**CRITICAL: Tag first, then build!** The version is embedded at build time via ldflags.

```bash
# 1. Create and push the tag FIRST
git tag v0.x.x
git push origin v0.x.x

# 2. Build binaries (will embed the tag version)
./scripts/release.sh v0.x.x

# 3. Create GitHub release with binaries
gh release create v0.x.x dist/* \
  --title "v0.x.x" \
  --notes "Release notes here"
```

If you build before tagging, binaries will have the wrong version embedded and must be rebuilt.
