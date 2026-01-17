# Weekly Report - Sage - Week of 2026-01-13

## Week Overview
Major release week: shipped v0.3.0 (Dynamic Provider Fields), v0.3.1 (bugfix), and v0.4.0 (CLI JSON Output). The sage CLI is now fully machine-readable with `--json` support on all commands, NDJSON streaming, and pre-compiled binaries available for download.

## Key Accomplishments

### Dynamic Provider Fields (v0.3.0)
- Added `ProviderField` type and `Fields()` method to Provider interface
- Updated storage format to nested maps (provider → account → field → value)
- Changed `AddProviderAccount` API to accept fields map with validation
- Dynamic field prompting in CLI with defaults support

### CLI JSON Output (v0.4.0)
- Added `--json` flag to all commands (version, provider, profile, complete)
- Implemented `--stream` flag for NDJSON streaming output
- Global `JSONOutput` flag for consistent JSON error handling
- Errors returned as JSON to stdout when `--json` is used

### Release Infrastructure (v0.4.0)
- Created `scripts/release.sh` for cross-compilation
- Builds for linux/darwin × amd64/arm64
- Version injection via ldflags
- First GitHub release with pre-compiled binaries

### Documentation
- Updated CLI docs with JSON examples for all commands
- Added binary installation instructions to README and installation.md
- Created CHANGELOG.md

## Decisions This Week

1. **Global JSONOutput flag** - Cleaner than per-command error handling → Prevents forgetting to handle errors, centralizes error formatting
2. **Deferred usage stats in streaming** - Requires provider-layer changes → Will revisit when needed

## Challenges Encountered
- Streaming doesn't include usage stats (OpenAI requires `stream_options.include_usage`) - deferred for future enhancement

## Metrics
- Commits: 7
- Releases: 3 (v0.3.0, v0.3.1, v0.4.0)
- New files: ~15 (including archived plan files)

## Next Week Priorities

1. Integrate sage with hub-core (CLI subprocess approach)
2. Test JSON output in real integration scenarios
3. Consider adding usage stats to streaming if needed
