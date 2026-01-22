# Phase 1: HTTP Client Configurability

> **Depends on:** None (foundation phase)
> **Enables:** Phase 2 (context testing), Phase 3 (CLI mocking)
>
> See: [Full Plan](../plan.md)

## Goal

Make HTTP client configurable and provide sensible defaults instead of infinite timeouts.

## Key Deliverables

- Add `HTTPClient *http.Client` field to `Client` struct
- Implement `defaultHTTPClient()` with 5-minute timeout
- Add `HTTPClient` field to `Request` struct
- Update all three providers to use `req.HTTPClient` instead of `http.DefaultClient`
- Ensure `NewClient()` populates default if nil
- Ensure Client methods populate Request.HTTPClient before passing to providers

## Files to Modify

- `pkg/sage/client.go` - Add field, default client, populate in methods
- `pkg/sage/types.go` - Add HTTPClient to Request
- `pkg/sage/providers/anthropic.go` - Replace `http.DefaultClient` with `req.HTTPClient`
- `pkg/sage/providers/openai.go` - Replace `http.DefaultClient` with `req.HTTPClient`
- `pkg/sage/providers/ollama.go` - Replace `http.DefaultClient` with `req.HTTPClient`

## Dependencies

**Internal:** None (foundation phase)

**External:** None (stdlib only)

## Implementation Notes

### Default HTTP Client

Create a function that provides sensible defaults:

```go
// pkg/sage/client.go
func defaultHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 5 * time.Minute,  // Reasonable for LLM requests
        Transport: &http.Transport{
            MaxIdleConns:          100,
            MaxIdleConnsPerHost:   10,
            IdleConnTimeout:       90 * time.Second,
            TLSHandshakeTimeout:   10 * time.Second,
            ResponseHeaderTimeout: 10 * time.Second,
        },
    }
}
```

**Rationale:**
- 5-minute total timeout (better than infinite)
- Connection pooling for performance
- Reasonable TLS and header timeouts
- Users can override by setting `client.HTTPClient`

### Client Initialization

Update `NewClient()` to populate HTTPClient:

```go
func NewClient() (*Client, error) {
    // ... existing config/secrets loading ...

    client := &Client{
        config:  config,
        secrets: secrets,
    }

    // Set default HTTP client if not provided
    if client.HTTPClient == nil {
        client.HTTPClient = defaultHTTPClient()
    }

    return client, nil
}
```

### Request Population

Before passing Request to provider, ensure HTTPClient is set:

```go
func (c *Client) Complete(profile string, req Request) (*Response, error) {
    // ... existing profile lookup ...

    // Ensure request has HTTP client
    req.HTTPClient = c.HTTPClient

    // Pass to provider
    return provider.Complete(req)
}
```

### Provider Updates

Each provider replaces `http.DefaultClient` with `req.HTTPClient`:

**Before:**
```go
resp, err := http.DefaultClient.Do(httpReq)
```

**After:**
```go
resp, err := req.HTTPClient.Do(httpReq)
```

Search for all occurrences:
- `anthropic.go`: ~3 uses
- `openai.go`: ~3 uses
- `ollama.go`: ~2 uses

### Backward Compatibility

This change is fully backward compatible:
- Existing code works without modification
- nil HTTPClient is handled by `NewClient()`
- Users can optionally configure HTTPClient

## Validation

### Automated Tests
- [ ] All existing tests pass: `go test ./...`
- [ ] Can create client without setting HTTPClient (uses default)
- [ ] Can create client with custom HTTPClient (uses custom)
- [ ] Providers use the configured client

### Code Verification
- [ ] `grep -r "http.DefaultClient" pkg/sage/providers/` returns zero results
- [ ] All providers use `req.HTTPClient` instead

### Manual Tests
- [ ] CLI still works: `sage complete "hello"`
- [ ] Timeout works: Set 1-second timeout, verify it fires on slow request
- [ ] Custom client works: Create client with custom timeout, verify it's used

### Integration Test (Optional)

Create a simple test to verify timeout behavior:

```go
func TestHTTPClientTimeout(t *testing.T) {
    // Create client with very short timeout
    client := &Client{
        HTTPClient: &http.Client{Timeout: 1 * time.Millisecond},
    }

    // Make request (should timeout)
    _, err := client.Complete("profile", Request{Prompt: "test"})

    // Verify timeout error
    if err == nil {
        t.Error("Expected timeout error, got nil")
    }
}
```

## Success Criteria

Phase 1 is complete when:

- [ ] `Client` struct has `HTTPClient` field
- [ ] `Request` struct has `HTTPClient` field
- [ ] `defaultHTTPClient()` function exists with 5-minute timeout
- [ ] `NewClient()` populates default HTTPClient
- [ ] Client methods populate Request.HTTPClient
- [ ] All three providers use `req.HTTPClient` (zero uses of `http.DefaultClient`)
- [ ] All existing tests pass
- [ ] CLI works identically: `sage complete "test"` succeeds
- [ ] Can configure custom timeout via `client.HTTPClient`
- [ ] grep confirms no `http.DefaultClient` in providers

## Notes

**Why this phase first:**
- Non-breaking change (lower risk)
- Enables testing infrastructure immediately
- Makes Phase 2 easier to test (can mock HTTP for context tests)

**User impact:**
- CLI users: Automatic improvement (5min timeout vs infinite)
- Library users: Optional configuration capability
- Zero breaking changes

**Next phase:**
Once HTTP client is configurable, Phase 2 can add context support with confidence that timeouts and cancellation will work correctly.
