# Project Progress - sage

## Plan Files
Roadmap: None (v4 archived)
Current Phase: None
Latest Weekly Report: [weekly-2026-W03.md](../docs/reports/weekly-2026-W03.md)

Last Updated: 2026-01-23

## Current Focus
v0.8.0 released - CLI scripting improvements for programmatic provider configuration

## Active Tasks
None - awaiting direction

## Open Questions/Blockers
None

## Completed This Week
- Version detection and update improvements (v0.5.0)
  - Added runtime/debug.ReadBuildInfo() fallback for go install version detection
  - Implemented 'sage update' command for easy CLI updates
  - Tagged and released v0.5.0
- Debugged --json and --json --stream output
  - Identified flag ordering requirement (flags before positional args)
  - Confirmed both output modes work correctly
  - Documentation already accurate
- Comprehensive architectural audit
  - Ran deep project audit to identify technical debt
  - Generated detailed audit report (gitignored for security)
  - Identified three high-priority issues for immediate addressing
- v0.6.0: Context Support and HTTP Client Configurability (RELEASED)
  - Phase 1: HTTP Client Configurability
    - Added Client.HTTPClient field with 5-minute default timeout
    - Created defaultHTTPClient() with connection pooling
    - Updated all providers to use configurable HTTP client
    - Created comprehensive timeout tests verifying actual timeout behavior
  - Phase 2: Context.Context Support (breaking change for library API)
    - Updated Provider interface to accept context.Context
    - Updated all provider implementations (anthropic, openai, ollama)
    - Added context cancellation handling in streaming
    - Updated CLI commands to use context internally
    - Fixed all test files to support context
  - Phase 3: CLI Test Coverage
    - Created test infrastructure (mockClient, captureStdout helpers)
    - Added 14 test cases for complete command
    - Achieved 12.1% CLI coverage with foundation for expansion
  - Phase 4: Documentation
    - Updated CHANGELOG.md with v0.6.0 breaking changes
    - Enhanced README.md with context usage examples
    - Created comprehensive migration guide (docs/migration-v0.6.md)
    - Verified all code examples compile
  - All 70 tests passing
  - Tagged and released v0.6.0 with binaries for all platforms
  - Published release: https://github.com/not-emily/sage/releases/tag/v0.6.0
- v0.7.0: Provider Fields Command (RELEASED)
  - Added `sage provider fields <provider>` command
  - Lists configuration fields required by a provider (key, label, required, secret, default)
  - Supports `--json` output for scripting
  - Updated CLI and library documentation
  - Published release: https://github.com/not-emily/sage/releases/tag/v0.7.0
- v0.8.0: Stdin Support for Provider Add (RELEASED)
  - Added `--stdin` flag to `sage provider add`
  - Allows passing field values as JSON via stdin (keeps secrets out of bash history)
  - Pairs with `provider fields --json` for full programmatic provider configuration
  - Updated CLI documentation with scripting examples
  - Published release: https://github.com/not-emily/sage/releases/tag/v0.8.0

## Next Session
Integrate sage with hub-core, or continue CLI/library improvements.
