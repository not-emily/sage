# Phase 1: Provider Fields

> **Depends on:** None
> **Enables:** Phase 2 (Storage Format), Phase 3 (Client API)
>
> See: [Full Plan](../plan.md)

## Goal

Add a `Fields()` method to the Provider interface so each provider can declare its configuration requirements.

## Key Deliverables

- `ProviderField` struct type with key, label, required, secret, default
- Updated `Provider` interface with `Fields()` method
- `Fields()` implementations for OpenAI, Anthropic, and Ollama providers

## Files to Modify

- `pkg/sage/providers/provider.go` — Add ProviderField type and update interface
- `pkg/sage/providers/openai.go` — Implement Fields()
- `pkg/sage/providers/anthropic.go` — Implement Fields()
- `pkg/sage/providers/ollama.go` — Implement Fields()

## Dependencies

**Internal:** None

**External:** None (stdlib only)

## Implementation Notes

### ProviderField Struct

```go
type ProviderField struct {
    Key      string `json:"key"`               // e.g., "api_key", "base_url"
    Label    string `json:"label"`             // e.g., "API Key", "Base URL"
    Required bool   `json:"required"`
    Secret   bool   `json:"secret"`            // stored in encrypted secrets
    Default  string `json:"default,omitempty"` // e.g., "http://localhost:11434"
}
```

### Provider Interface Update

```go
type Provider interface {
    Name() string
    Complete(req Request) (*Response, error)
    CompleteStream(req Request) (<-chan Chunk, error)
    ListModels(apiKey, baseURL string) ([]Model, error)
    Fields() []ProviderField  // NEW
}
```

### Field Definitions Per Provider

**OpenAI:**
- api_key: required, secret
- base_url: optional, not secret, default "https://api.openai.com/v1"

**Anthropic:**
- api_key: required, secret

**Ollama:**
- base_url: required, not secret, default "http://localhost:11434"
- api_key: optional, secret (for remote Ollama instances)

### Ordering Convention

Fields should be ordered by importance/typical configuration flow:
1. Required fields first
2. Within required: most commonly customized first
3. Optional fields last

## Validation

How do we know this phase is complete?

- [ ] ProviderField type exists in provider.go
- [ ] Provider interface includes Fields() method
- [ ] All three providers implement Fields()
- [ ] `go build ./...` succeeds
- [ ] Existing tests pass (no breaking changes to current behavior)
