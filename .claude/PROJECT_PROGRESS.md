# Project Progress - sage

## Plan Files
Roadmap: None (v4 archived)
Current Phase: None
Latest Weekly Report: [weekly-2026-W01.md](../docs/reports/weekly-2026-W01.md)

Last Updated: 2026-01-16

## Current Focus
CLI JSON Output & Release Infrastructure (v0.4.0) - All phases complete, ready for release

## Active Tasks
- [NEXT] Tag and release v0.4.0

## Open Questions/Blockers
None

## Completed This Week
- Tagged and released v0.3.0
- Tagged and released v0.3.1 (OpenAI base_url fix + version injection)
- Phase 1: JSON Streaming (v0.4.0)
  - Added --stream flag to complete command
  - Implemented NDJSON output for --json --stream
  - Each chunk outputs {"content":"...", "done":false}
  - Final chunk includes model name
  - Usage stats deferred (requires provider changes)
- Phase 2: CLI JSON Coverage (v0.4.0)
  - Added --json flag to all commands (version, provider, profile)
  - Global JSONOutput flag for consistent error handling
  - JSON errors output to stdout for machine parsing
  - All commands return valid JSON on success or error
- Phase 3: Release Infrastructure (v0.4.0)
  - Created scripts/release.sh for cross-compilation
  - Builds for linux/darwin × amd64/arm64
  - Version injection via ldflags
  - Outputs gh release create command
- Phase 4: Documentation (v0.4.0)
  - Updated docs/cli-usage.md with JSON examples for all commands
  - Added NDJSON streaming documentation
  - Added error handling section
  - Updated docs/installation.md with binary downloads
  - Updated README.md with binary installation instructions
- Phase 1: Provider Fields
  - Added ProviderField type to provider.go
  - Updated Provider interface with Fields() method
  - Implemented Fields() for OpenAI, Anthropic, Ollama

- Phase 2: Storage Format
  - Updated ProviderConfig to use nested map (account -> field -> value)
  - Changed secrets key format to provider:account:field
  - Added old config format detection with helpful error message
  - Added helper functions (GetAccountFields, SetAccountFields, etc.)

- Phase 3: Client API
  - Changed AddProviderAccount signature to accept fields map
  - Added required field validation
  - Added GetProviderFields function
  - Re-exported ProviderField type for public API

- Phase 4: CLI Update
  - Dynamic field prompting in `sage provider add`
  - Fields sorted: required first, then optional
  - Default values displayed in prompt brackets
  - Defaults applied when user presses Enter

- Phase 5: Documentation
  - Created CHANGELOG.md with v0.3.0 breaking changes and features
  - Updated README.md with new Quick Start, Library Usage examples
  - All tests passing, build successful

## Next Session
Tag v0.4.0, create GitHub release with binaries, archive plan files.
