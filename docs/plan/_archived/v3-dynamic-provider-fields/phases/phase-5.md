# Phase 5: Documentation

> **Depends on:** Phase 4 (CLI Update)
> **Enables:** Release ready
>
> See: [Full Plan](../plan.md)

## Goal

Document the breaking changes and new features for the v0.3.0 release.

## Key Deliverables

- CHANGELOG.md created with v0.3.0 breaking changes
- README.md updated with new API usage
- Clean removal of plan files (optional)

## Files to Create/Modify

- `CHANGELOG.md` — Create new file
- `README.md` — Update API documentation

## Dependencies

**Internal:** All previous phases complete

**External:** None

## Implementation Notes

### CHANGELOG.md

Create a new CHANGELOG.md following Keep a Changelog format:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - YYYY-MM-DD

### Breaking Changes

- **Config format changed**: Provider accounts now store multiple fields instead of just API key reference. Old config files are incompatible - delete `~/.sage/config.json` and re-add providers.
- **Secrets key format changed**: Now uses `provider:account:field` format (was `provider:account`). Old secrets will not be found - re-add providers to create new secrets.
- **AddProviderAccount signature changed**: Now accepts `fields map[string]string` instead of `apiKey string`.

### Added

- `ProviderField` type for declaring provider configuration requirements
- `Provider.Fields()` method - providers now declare their own field requirements
- `GetProviderFields(provider)` function to query field requirements
- Per-account field storage - different accounts can have different configurations
- Dynamic CLI prompting based on provider field requirements

### Changed

- OpenAI provider: API Key required, Base URL optional (default: https://api.openai.com/v1)
- Anthropic provider: API Key required
- Ollama provider: Base URL required (default: http://localhost:11434), API Key optional

## [0.2.0] - 2026-01-15

### Changed

- Secrets auto-initialize on first use (no longer requires `sage init`)

## [0.1.0] - Initial Release

- Initial release with OpenAI, Anthropic, and Ollama providers
- Encrypted secrets storage
- Multi-account support per provider
```

### README.md Updates

Update the README to reflect the new API:

**Library Usage section:**
```go
// Query provider field requirements
fields, _ := sage.GetProviderFields("openai")
for _, f := range fields {
    fmt.Printf("%s (required=%v, secret=%v)\n", f.Label, f.Required, f.Secret)
}

// Add provider with dynamic fields
client.AddProviderAccount("openai", "default", map[string]string{
    "api_key":  "sk-...",
    "base_url": "https://api.openai.com/v1",
})

// Add Ollama (different fields)
client.AddProviderAccount("ollama", "local", map[string]string{
    "base_url": "http://localhost:11434",
})
```

**CLI Usage section:**
```
# Add OpenAI (prompts for API Key required, Base URL optional)
sage provider add openai

# Add Ollama (prompts for Base URL required, API Key optional)
sage provider add ollama

# Add named account
sage provider add openai work
```

### Plan Files

After release, the plan files in `docs/plan/` can be:
1. Kept for historical reference
2. Moved to a `docs/archive/` directory
3. Deleted

This is a team preference decision. The plan served its purpose during implementation.

## Validation

How do we know this phase is complete?

- [ ] CHANGELOG.md exists with v0.3.0 changes documented
- [ ] README.md reflects new API and CLI behavior
- [ ] Breaking changes are clearly called out
- [ ] Migration path is documented (delete config, re-add providers)
- [ ] Tag v0.3.0 created and pushed
