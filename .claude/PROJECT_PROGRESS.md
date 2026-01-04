# Project Progress - sage

## Plan Files
Roadmap: [docs/plan/plan.md](../docs/plan/plan.md)
Current Phase: [Phase 13: Hub-core Sage Dependency](../docs/plan/phases/phase-13.md)
Latest Weekly Report: None

Last Updated: 2026-01-04

## Current Focus
Sage standalone complete. Beginning hub-core integration.

## Active Tasks
- [NEXT] Phase 13: Hub-core Sage Dependency

## Open Questions/Blockers
None

## Completed This Week
- Phase 1: Project Foundation
  - Created go.mod (github.com/not-emily/sage)
  - Directory structure (cmd/sage, pkg/sage, internal/cli, scripts)
  - Core types in pkg/sage/types.go
  - Build script (scripts/build.sh)
  - Basic CLI skeleton with version command
  - All validation checks pass
- Phase 2: Config Management
  - Config struct with Providers and Profiles maps
  - ConfigDir() creates ~/.config/sage/ with 0755 permissions
  - LoadConfig() returns empty config when file doesn't exist
  - Save() writes formatted JSON
  - GetProfile() with default fallback
  - GetProvider() helper
  - All 5 unit tests passing
- Phase 3: Secrets Management
  - InitSecrets() creates master.key with 0600 permissions
  - AES-256-GCM encryption with random nonce per encryption
  - LoadSecrets/SaveSecrets for encrypted storage
  - GetSecret/SetSecret/DeleteSecret/HasSecret helpers
  - Permission validation (refuses insecure key files)
  - All 14 unit tests passing
- Phase 4: Provider Interface
  - Provider interface (Name, Complete, CompleteStream)
  - Request/Response/Chunk/Usage types
  - Provider registry with Register/Get/List/Exists
  - All 19 unit tests passing
- Phase 5: OpenAI Provider
  - Complete() for blocking requests
  - CompleteStream() with SSE parsing
  - Error handling (401, 429, 500+)
  - BaseURL override for custom endpoints
  - Auto-registration via init()
  - All 23 unit tests passing
- Phase 6: Anthropic Provider
  - Separate system field (not in messages array)
  - x-api-key and anthropic-version headers
  - Event-based SSE streaming (content_block_delta)
  - Default max_tokens (1024) since required
  - All 28 unit tests passing
- Phase 7: Ollama Provider
  - NDJSON streaming (not SSE)
  - Optional authentication (only if API key provided)
  - Default localhost:11434 BaseURL
  - Token mapping (prompt_eval_count, eval_count)
  - All 34 unit tests passing
- Phase 8: Library API
  - Client struct with NewClient() constructor
  - Complete() and CompleteStream() with profile resolution
  - Profile management (Add/Remove/List/SetDefault)
  - Provider account management (Add/Remove/List/Has)
  - All 41 unit tests passing
- Phase 9: CLI Core
  - Subcommand routing in internal/cli/root.go
  - sage init creates config dir and master key
  - sage complete with streaming (default) and --json mode
  - stdin support for prompts
  - sage version and sage help commands
- Phase 10: Provider Commands
  - internal/cli/provider.go with subcommand routing
  - sage provider list shows configured providers
  - sage provider add with interactive prompt or --api-key-env
  - sage provider remove with --account flag
  - Fixed flag parsing order (reorderArgs helper)
  - scripts/test-provider.sh validation script
- Phase 11: Profile Commands
  - internal/cli/profile.go with subcommand routing
  - sage profile list shows profiles with default marker
  - sage profile add with --provider, --model, --account flags
  - sage profile remove and set-default commands
  - Validates provider account exists before creating profile
  - scripts/test-profile.sh validation script
- Phase 12: Documentation
  - README.md with quick start guide
  - docs/installation.md with build and install instructions
  - docs/cli-usage.md with full command reference
  - docs/library-usage.md with Go integration examples

## Next Session
Phase 13: Hub-core Sage Dependency - add sage as a Go module dependency to hub-core.
