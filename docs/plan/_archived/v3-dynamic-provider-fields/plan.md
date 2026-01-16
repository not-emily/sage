# Dynamic Provider Fields

> **Status:** Planning complete | Last updated: 2026-01-15
>
> Phase files: [phases/](phases/)

## Overview

Enable providers to declare their configuration field requirements dynamically, allowing clients (CLI, hub-core, TUI) to render appropriate configuration forms without hardcoding provider-specific logic.

Currently, all providers are treated as if they need an "API Key", but Ollama primarily needs a "Base URL" while API key is optional. Azure OpenAI (future) needs endpoint, deployment_id, and api_version in addition to API key. This feature makes provider configuration requirements self-describing.

## Core Vision

- **Providers as source of truth**: Each provider declares what fields it needs - no external registry or hardcoding
- **Data characteristics, not rendering instructions**: Fields describe what data is needed (required, secret, default) - clients decide how to render
- **Per-account flexibility**: All fields stored per-account, allowing different configurations for different accounts of the same provider
- **Universal adapter**: Sage remains usable in any context - CLI-only projects ignore field metadata, apps with config UIs leverage it

## Requirements

### Must Have
- Provider interface includes `Fields()` method returning field requirements
- Each field has: key, label, required, secret, default
- `AddProviderAccount` accepts dynamic fields map instead of just apiKey
- Config storage updated to per-account fields
- Secrets key format updated to `provider:account:field`
- CLI dynamically prompts based on provider fields
- Clear error if old config format detected

### Nice to Have
- Field validation (e.g., URL format for base_url)
- Field grouping (e.g., "Authentication" vs "Connection")

### Out of Scope
- Config migration from old format (clean break)
- TUI changes (separate project, consumes hub-core API)
- New provider implementations (Azure, etc.) - this just enables them

## Constraints

- **Breaking change**: v0.3.0 will break existing configs and API
- **Sage stdlib-only**: No third-party dependencies (matches hub-core constraint)
- **Backward compatibility**: None - clean break for simplicity

## Success Metrics

- `sage provider add ollama` prompts for Base URL first, API Key optional
- `sage provider add openai` prompts for API Key required, Base URL optional
- Hub-core can query provider fields and expose via API
- Adding a new provider (e.g., Azure) requires zero client changes

## Architecture Decisions

### 1. Provider declares its own fields
**Choice:** Each provider implements `Fields() []ProviderField` method
**Rationale:** Keeps provider knowledge encapsulated - adding a new provider is self-contained
**Trade-offs:** Slightly more code per provider vs central registry

### 2. Per-account field storage
**Choice:** All non-secret fields stored per-account in config, secrets keyed by `provider:account:field`
**Rationale:** Different accounts may need different configurations (e.g., two Azure accounts with different endpoints)
**Trade-offs:** More complex storage structure, but more flexible

### 3. Clean break on config format
**Choice:** No migration - detect old format and error with helpful message
**Rationale:** Migration code is permanent cruft, only 1-2 users affected
**Trade-offs:** Users must re-add providers (2 minutes of work)

### 4. Include label in field metadata
**Choice:** Fields include `Label` for display purposes
**Rationale:** Reasonable default that clients can override - better than forcing all clients to derive labels from keys
**Trade-offs:** Slight mixing of data and presentation, but pragmatic

## Project Structure

```
sage/
├── pkg/sage/
│   ├── providers/
│   │   ├── provider.go      # Provider interface + ProviderField type
│   │   ├── openai.go        # Implements Fields()
│   │   ├── anthropic.go     # Implements Fields()
│   │   └── ollama.go        # Implements Fields()
│   ├── config.go            # Updated ProviderConfig struct
│   ├── secrets.go           # Updated key format
│   └── client.go            # Updated AddProviderAccount
├── internal/cli/
│   └── provider.go          # Dynamic field prompting
├── CHANGELOG.md             # New - track breaking changes
└── docs/plan/               # This plan
```

## Core Interfaces

### ProviderField
```go
type ProviderField struct {
    Key      string `json:"key"`               // e.g., "api_key", "base_url"
    Label    string `json:"label"`             // e.g., "API Key", "Base URL"
    Required bool   `json:"required"`
    Secret   bool   `json:"secret"`            // stored in encrypted secrets
    Default  string `json:"default,omitempty"` // e.g., "http://localhost:11434"
}
```

### Provider interface (updated)
```go
type Provider interface {
    Name() string
    Complete(req Request) (*Response, error)
    CompleteStream(req Request) (<-chan Chunk, error)
    ListModels(apiKey, baseURL string) ([]Model, error)
    Fields() []ProviderField  // NEW
}
```

### Client API (updated)
```go
// Changed signature
func (c *Client) AddProviderAccount(provider, account string, fields map[string]string) error

// New helper
func GetProviderFields(provider string) ([]ProviderField, error)
```

### Config structure (updated)
```go
type ProviderConfig struct {
    Accounts map[string]map[string]string `json:"accounts"` // account -> field -> value
}
```

## Implementation Phases

| Phase | Name | Scope | Depends On | Key Outputs |
|-------|------|-------|------------|-------------|
| 1 | Provider Fields | Add Fields() to interface + implementations | — | Field metadata available |
| 2 | Storage Format | Per-account config, new secrets key format | Phase 1 | New storage structure |
| 3 | Client API | Update AddProviderAccount, add GetProviderFields | Phase 2 | New client API |
| 4 | CLI Update | Dynamic field prompting | Phase 3 | Updated CLI |
| 5 | Documentation | CHANGELOG, README updates | Phase 4 | Release ready |

### Critical Path
All phases are sequential - each builds on the previous.

### Phase Details
- [Phase 1: Provider Fields](phases/phase-1.md)
- [Phase 2: Storage Format](phases/phase-2.md)
- [Phase 3: Client API](phases/phase-3.md)
- [Phase 4: CLI Update](phases/phase-4.md)
- [Phase 5: Documentation](phases/phase-5.md)

## Tech Stack

| Category | Choice | Notes |
|----------|--------|-------|
| Language | Go | Existing |
| Dependencies | stdlib only | No changes |

## Future Considerations

- **Field validation**: Could add `Validate` field with regex or validation type
- **Field groups**: Could group fields for better UI organization
- **Azure provider**: This work enables adding Azure OpenAI provider easily
- **Other providers**: Gemini, Cohere, etc. - just implement Provider interface with Fields()
