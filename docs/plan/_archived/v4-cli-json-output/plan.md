# Sage: CLI JSON Output and Release Infrastructure

> **Status:** Planning complete | Last updated: 2026-01-16
>
> Phase files: [phases/](phases/)
>
> **Enables:** [Hub-core CLI Adapter](../../../hub-core/docs/plan/plan.md)

## Overview

Add machine-readable JSON output to sage CLI commands and establish release infrastructure for pre-compiled binaries. This enables hub-core to use sage as a CLI tool (via subprocess) rather than a Go library, which is required for user-scoped execution where each user's `~/.config/sage/` credentials are used.

Currently, sage CLI streams plain text by default. Hub-core needs structured JSON output to parse responses programmatically, and NDJSON streaming for real-time chat experiences.

## Core Vision

- **CLI as the interface**: Hub-core will call sage via subprocess, not import as Go library
- **User-scoped execution**: Running `sudo -u emily sage ...` uses emily's config
- **Machine-readable output**: All commands support `--json` for structured output
- **Pre-compiled distribution**: No Go dependency required on hub stations

## Requirements

### Must Have
- `--json` flag for structured output (non-streaming, buffered)
- `--json --stream` for NDJSON streaming output
- All subcommands support `--json` (complete, provider, profile, version)
- Cross-compiled binaries for linux/darwin × amd64/arm64
- GitHub releases with downloadable binaries

### Nice to Have
- Checksum files for binary verification
- Install script that detects platform

### Out of Scope
- Go library changes (hub-core will stop using it)
- New CLI commands beyond JSON support
- Breaking changes to existing CLI behavior (--json is additive)

## Constraints

- **Backward compatible**: Existing CLI behavior unchanged, --json is opt-in
- **No new dependencies**: Keep sage dependency-free
- **Cross-platform**: Must work on Linux and macOS

## Success Metrics

- `sage complete "hi" --json` returns valid JSON
- `sage complete "hi" --json --stream` streams NDJSON lines
- `sage provider list --json` returns structured provider data
- GitHub release contains binaries for all target platforms
- Binary can be downloaded and run without Go installed

## Architecture Decisions

### 1. NDJSON for streaming
**Choice:** `--json --stream` outputs newline-delimited JSON (one object per line)
**Rationale:** Standard format for streaming structured data, easy to parse line-by-line
**Trade-offs:** Slightly more complex than plain text streaming

### 2. --json implies non-streaming by default
**Choice:** `--json` alone buffers and outputs single JSON object
**Rationale:** Most JSON consumers expect complete document, streaming is opt-in
**Trade-offs:** Need two flags for streaming JSON

### 3. Pre-compiled binaries via GitHub releases
**Choice:** Cross-compile and upload to GitHub releases
**Rationale:** No Go required on target machines, easy curl-based install
**Trade-offs:** Must maintain release process

## Implementation Phases

| Phase | Name | Scope | Depends On | Key Outputs |
|-------|------|-------|------------|-------------|
| 1 | JSON Streaming | Add `--stream` flag for NDJSON | — | Streaming JSON works |
| 2 | CLI JSON Coverage | All commands support `--json` | Phase 1 | Full CLI machine-readable |
| 3 | Release Infrastructure | Cross-compile, GitHub releases | — | Binaries downloadable |
| 4 | Documentation | CLI reference with JSON examples | Phase 2 | Usage documented |

### Critical Path
- Phases 1 → 2 → 4 (JSON features sequential)
- Phase 3 can run in parallel with 1-2

### Phase Details
- [Phase 1: JSON Streaming](phases/phase-1.md)
- [Phase 2: CLI JSON Coverage](phases/phase-2.md)
- [Phase 3: Release Infrastructure](phases/phase-3.md)
- [Phase 4: Documentation](phases/phase-4.md)

## Tech Stack

| Category | Choice | Notes |
|----------|--------|-------|
| Language | Go | Existing |
| Release | GitHub Releases | Standard, curl-friendly |
| Build | goreleaser or shell script | Cross-compilation |

## Future Considerations

- **Homebrew formula**: Could add `brew install sage` support
- **Shell completions**: Could generate bash/zsh completions
- **Config migration**: If config format changes, could add migration command
