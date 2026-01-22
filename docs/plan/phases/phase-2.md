# Phase 2: Context.Context Support

> **Depends on:** Phase 1 (HTTP client must be configurable)
> **Enables:** Phase 3 (stable interfaces for testing), Phase 4 (breaking changes to document)
>
> See: [Full Plan](../plan.md)

## Goal

Add context.Context to all Provider methods and Client methods to enable cancellation, timeouts, and future observability.

## Key Deliverables

- Update Provider interface to accept context in all methods
- Update all three provider implementations to accept and use context
- Update Client methods to accept and pass context
- Update CLI commands to create and pass context
- Implement context cancellation handling in streaming
- Use `http.NewRequestWithContext` in all HTTP calls

## Files to Modify

- `pkg/sage/providers/provider.go` - Update interface signatures
- `pkg/sage/providers/anthropic.go` - Accept ctx, use in HTTP requests, handle in streaming
- `pkg/sage/providers/openai.go` - Accept ctx, use in HTTP requests, handle in streaming
- `pkg/sage/providers/ollama.go` - Accept ctx, use in HTTP requests, handle in streaming
- `pkg/sage/client.go` - Accept ctx in methods, pass to providers
- `internal/cli/complete.go` - Create `context.Background()`, pass to client
- `internal/cli/models.go` - Create `context.Background()`, pass to client
- `internal/cli/init.go` - Update if it calls Client methods
- `internal/cli/profile.go` - Update if it calls Client methods
- `internal/cli/provider.go` - Update if it calls Client methods

## Dependencies

**Internal:** Phase 1 (HTTP client must be in place for testing)

**External:** `context` package (stdlib)

## Implementation Notes

### Context Pattern

Thread context through the entire call chain:

```
CLI command creates context.Background()
    ↓
Client methods accept ctx, pass to provider
    ↓
Provider methods accept ctx, use in HTTP requests
    ↓
http.NewRequestWithContext(ctx, ...) respects cancellation
```

### Provider Interface Update

```go
// pkg/sage/providers/provider.go
type Provider interface {
    Name() string
    Complete(ctx context.Context, req Request) (*Response, error)
    CompleteStream(ctx context.Context, req Request) (<-chan Chunk, error)
    ListModels(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error)
    Fields() []ProviderField
}
```

### Client Method Signatures

```go
// pkg/sage/client.go
func (c *Client) Complete(ctx context.Context, profile string, req Request) (*Response, error)
func (c *Client) CompleteStream(ctx context.Context, profile string, req Request) (<-chan Chunk, error)
func (c *Client) ListModels(ctx context.Context, providerName string) ([]ModelInfo, error)
```

### HTTP Request with Context

All HTTP requests must use context:

**Before:**
```go
httpReq, err := http.NewRequest("POST", url, body)
resp, err := req.HTTPClient.Do(httpReq)
```

**After:**
```go
httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
resp, err := req.HTTPClient.Do(httpReq)
```

### Streaming with Context

Streaming goroutines must respect `ctx.Done()`:

```go
func (p *Provider) CompleteStream(ctx context.Context, req Request) (<-chan Chunk, error) {
    chunks := make(chan Chunk)

    go func() {
        defer close(chunks)

        httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
        if err != nil {
            chunks <- Chunk{Error: err}
            return
        }

        resp, err := req.HTTPClient.Do(httpReq)
        if err != nil {
            chunks <- Chunk{Error: err}
            return
        }
        defer resp.Body.Close()

        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            // Check for cancellation
            select {
            case <-ctx.Done():
                chunks <- Chunk{Error: ctx.Err()}
                return
            default:
                // Process and send chunk
                chunk := parseChunk(scanner.Text())
                chunks <- chunk
            }
        }

        if err := scanner.Err(); err != nil {
            chunks <- Chunk{Error: err}
        }
    }()

    return chunks, nil
}
```

**Key points:**
- Check `ctx.Done()` in the loop
- Send `ctx.Err()` on cancellation
- Always close channel via defer
- HTTP request respects context automatically

### CLI Context Creation

All CLI commands create background context:

```go
// internal/cli/complete.go
func runComplete(args []string) error {
    // ... parse flags ...

    ctx := context.Background()

    client, err := sage.NewClient()
    if err != nil {
        return err
    }

    if *jsonOutput && *stream {
        return completeStreamJSON(ctx, client, *profile, req)
    }

    if *jsonOutput {
        return completeJSON(ctx, client, *profile, req)
    }

    return completeStream(ctx, client, *profile, req)
}
```

**Note:** CLI uses `context.Background()` without cancellation handling. Process exit is sufficient for CLI use case. Future enhancement could add signal handling for Ctrl+C.

## Sub-phases

### 2.1: Update Interfaces and Core Types

**Goal:** Change Provider interface and Client method signatures

**Deliverables:**
- Updated Provider interface in `provider.go`
- Updated Client method signatures in `client.go`
- Code won't compile until sub-phase 2.2 complete (expected)

**Steps:**
1. Update `Provider` interface to accept `context.Context` as first parameter
2. Update Client method signatures to accept `context.Context` as first parameter
3. Commit (code won't compile yet, but interface is defined)

### 2.2: Update Provider Implementations

**Goal:** Make all providers context-aware

**Deliverables:**
- `anthropic.go` uses context in all methods
- `openai.go` uses context in all methods
- `ollama.go` uses context in all methods
- All provider tests pass

**Steps:**
1. Update `anthropic.go`:
   - Add `ctx context.Context` to method signatures
   - Replace `http.NewRequest` with `http.NewRequestWithContext`
   - Add context cancellation handling in streaming
2. Update `openai.go` (same pattern)
3. Update `ollama.go` (same pattern)
4. Run provider tests: `go test ./pkg/sage/providers/...`
5. Fix any compilation errors
6. Commit

### 2.3: Update CLI Commands

**Goal:** CLI creates and passes contexts

**Deliverables:**
- All CLI commands create `context.Background()`
- All CLI commands pass context to Client methods
- CLI works: `sage complete "test"` succeeds

**Steps:**
1. Update `complete.go` to create and pass context
2. Update `models.go` to create and pass context
3. Update other CLI files if they call Client methods
4. Test CLI: `./bin/sage complete "hello"`
5. Commit

## Validation

### Compilation
- [ ] Code compiles without errors: `go build ./...`
- [ ] All packages build successfully

### Automated Tests
- [ ] All existing tests pass: `go test ./...`
- [ ] Provider tests pass: `go test ./pkg/sage/providers/...`
- [ ] Client tests pass: `go test ./pkg/sage/...`

### Context Functionality Tests

**Timeout test:**
```go
func TestContextTimeout(t *testing.T) {
    client, _ := sage.NewClient()

    // Create context with very short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    // Should timeout
    _, err := client.Complete(ctx, "profile", sage.Request{Prompt: "test"})

    if !errors.Is(err, context.DeadlineExceeded) {
        t.Errorf("Expected DeadlineExceeded, got %v", err)
    }
}
```

**Cancellation test:**
```go
func TestContextCancellation(t *testing.T) {
    client, _ := sage.NewClient()

    ctx, cancel := context.WithCancel(context.Background())

    // Cancel immediately
    cancel()

    // Should fail with context.Canceled
    _, err := client.Complete(ctx, "profile", sage.Request{Prompt: "test"})

    if !errors.Is(err, context.Canceled) {
        t.Errorf("Expected Canceled, got %v", err)
    }
}
```

**Streaming cancellation test:**
```go
func TestStreamingCancellation(t *testing.T) {
    client, _ := sage.NewClient()

    ctx, cancel := context.WithCancel(context.Background())

    chunks, err := client.CompleteStream(ctx, "profile", sage.Request{Prompt: "test"})
    if err != nil {
        t.Fatal(err)
    }

    // Cancel after first chunk
    firstChunk := <-chunks
    cancel()

    // Should receive cancellation error
    var gotCancellation bool
    for chunk := range chunks {
        if chunk.Error != nil && errors.Is(chunk.Error, context.Canceled) {
            gotCancellation = true
        }
    }

    if !gotCancellation {
        t.Error("Expected cancellation error in stream")
    }
}
```

### Manual Tests
- [ ] CLI works: `sage complete "hello"`
- [ ] CLI with all flags: `sage complete --json --profile=test "hello"`
- [ ] Models command: `sage models openai`
- [ ] Long request can be cancelled (if signal handling added)

### Goroutine Leak Test

Run repeated cancellations and check goroutine count:

```go
func TestNoGoroutineLeaks(t *testing.T) {
    initialGoroutines := runtime.NumGoroutine()

    for i := 0; i < 100; i++ {
        ctx, cancel := context.WithCancel(context.Background())
        client, _ := sage.NewClient()

        chunks, _ := client.CompleteStream(ctx, "profile", sage.Request{Prompt: "test"})
        cancel()  // Cancel immediately

        // Drain channel
        for range chunks {
        }
    }

    // Allow goroutines to clean up
    time.Sleep(100 * time.Millisecond)
    runtime.GC()

    finalGoroutines := runtime.NumGoroutine()

    // Should not have leaked goroutines
    if finalGoroutines > initialGoroutines+5 {
        t.Errorf("Goroutine leak: started with %d, ended with %d",
            initialGoroutines, finalGoroutines)
    }
}
```

## Success Criteria

Phase 2 is complete when:

- [ ] Provider interface accepts `context.Context` in all methods
- [ ] All three providers implement context-aware methods
- [ ] Client methods accept and pass context
- [ ] CLI commands create and pass context
- [ ] All HTTP requests use `http.NewRequestWithContext`
- [ ] Streaming respects `ctx.Done()` for cancellation
- [ ] All existing tests pass
- [ ] Context timeout test passes
- [ ] Context cancellation test passes
- [ ] Streaming cancellation test passes
- [ ] No goroutine leaks under cancellation
- [ ] CLI works identically: `sage complete "test"` succeeds
- [ ] Code compiles without errors

## Breaking Changes

**Library API Changes:**
- `Complete(profile, req)` → `Complete(ctx, profile, req)`
- `CompleteStream(profile, req)` → `CompleteStream(ctx, profile, req)`
- `ListModels(provider)` → `ListModels(ctx, provider)`

**CLI Changes:**
- None (CLI behavior is identical)

**Migration Path:**
Library users must update all calls to pass `context.Background()` or appropriate context.

See Phase 4 for migration documentation.

## Notes

**Why this is a breaking change:**
Go doesn't support optional parameters. Context must be first parameter by convention. All callers must update.

**Why break now:**
- Library is at v0.5.x (pre-1.0)
- Hub-core hasn't integrated yet (limited blast radius)
- Context is essential for production use
- Better to break early than after wider adoption

**Future enhancements:**
- Signal handling in CLI (Ctrl+C cancels requests)
- Context-aware progress reporting
- Trace ID propagation via context
