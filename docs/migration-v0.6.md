# Migration Guide: v0.5 to v0.6

This guide helps you migrate from sage v0.5 to v0.6.

## Summary

**CLI users**: No changes required. The CLI behavior is identical to v0.5.

**Library users**: Breaking changes to method signatures. All `Client` and `Provider` methods now require `context.Context` as the first parameter.

## Breaking Changes

### Client Methods

All `Client` methods now accept `context.Context` as the first parameter:

#### Before (v0.5)
```go
resp, err := client.Complete(profile, req)
chunks, err := client.CompleteStream(profile, req)
models, err := client.ListModels(provider, account)
```

#### After (v0.6)
```go
ctx := context.Background()
resp, err := client.Complete(ctx, profile, req)
chunks, err := client.CompleteStream(ctx, profile, req)
models, err := client.ListModels(ctx, provider, account)
```

### Provider Interface

If you've implemented custom providers, update the `Provider` interface:

#### Before (v0.5)
```go
type Provider interface {
    Name() string
    Complete(req Request) (*Response, error)
    CompleteStream(req Request) (<-chan Chunk, error)
    ListModels(apiKey, baseURL string) ([]ModelInfo, error)
    Fields() []ProviderField
}
```

#### After (v0.6)
```go
type Provider interface {
    Name() string
    Complete(ctx context.Context, req Request) (*Response, error)
    CompleteStream(ctx context.Context, req Request) (<-chan Chunk, error)
    ListModels(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error)
    Fields() []ProviderField
}
```

## Migration Patterns

### Basic Migration

The simplest migration is to add `context.Background()`:

```go
import "context"

// Old code
resp, err := client.Complete("myprofile", sage.Request{
    Prompt: "Hello!",
})

// New code
ctx := context.Background()
resp, err := client.Complete(ctx, "myprofile", sage.Request{
    Prompt: "Hello!",
})
```

### Adding Timeouts

If your code had custom timeout logic, you can now use context timeouts:

```go
import (
    "context"
    "time"
)

// Old approach (custom HTTP client setup)
// ... complex HTTP client configuration ...

// New approach (simple context timeout)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.Complete(ctx, "myprofile", sage.Request{
    Prompt: "Generate a long response...",
})
if err == context.DeadlineExceeded {
    fmt.Println("Request timed out after 30 seconds")
}
```

### Request Cancellation

Enable user-initiated cancellation:

```go
import (
    "context"
    "os"
    "os/signal"
)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Cancel on Ctrl+C
go func() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
    cancel()
}()

resp, err := client.Complete(ctx, "myprofile", sage.Request{
    Prompt: "This is a long request...",
})
if err == context.Canceled {
    fmt.Println("Request was canceled by user")
}
```

### Streaming with Context

Streaming now respects context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
defer cancel()

chunks, err := client.CompleteStream(ctx, "myprofile", sage.Request{
    Prompt: "Stream a story...",
})
if err != nil {
    panic(err)
}

for chunk := range chunks {
    if chunk.Error != nil {
        if chunk.Error == context.DeadlineExceeded {
            fmt.Println("Streaming timed out")
        }
        break
    }
    if chunk.Done {
        break
    }
    fmt.Print(chunk.Content)
}
```

### HTTP Request Context

If you're passing requests through to other services:

```go
import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
    // Use the HTTP request's context for automatic cancellation
    // when the client disconnects
    resp, err := client.Complete(r.Context(), "myprofile", sage.Request{
        Prompt: "Handle this request...",
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write([]byte(resp.Content))
}
```

## New Features (Non-Breaking)

### HTTP Client Configuration

You can now configure the HTTP client used for provider requests:

```go
import (
    "net/http"
    "time"
)

client, _ := sage.NewClient()

// Custom HTTP client with specific timeouts
client.HTTPClient = &http.Client{
    Timeout: 2 * time.Minute,
    Transport: &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        MaxIdleConns: 50,
        IdleConnTimeout: 60 * time.Second,
    },
}
```

**Default behavior**: If not set, sage uses a default client with:
- 5-minute request timeout
- Connection pooling (100 max idle connections, 10 per host)
- 90-second idle timeout
- 10-second TLS handshake timeout

### Benefits of Context Support

1. **Timeouts**: Set deadlines for long-running requests
2. **Cancellation**: Cancel in-flight requests when they're no longer needed
3. **Tracing**: Integrate with distributed tracing systems (OpenTelemetry, etc.)
4. **Request scoping**: Pass request-scoped values through the call chain
5. **Resource cleanup**: Automatic cleanup when requests are canceled

## Common Issues

### "not enough arguments in call"

```
Error: not enough arguments in call to client.Complete
    have (string, sage.Request)
    want (context.Context, string, sage.Request)
```

**Fix**: Add `context.Background()` as the first argument.

### "cannot use ctx (variable of type context.Context) as string value"

This means you added context but didn't shift the other parameters:

```go
// Wrong
client.Complete(ctx, req, profile)

// Correct
client.Complete(ctx, profile, req)
```

### Testing without real API calls

Use a context with a very short timeout to simulate failures:

```go
import "testing"

func TestClientTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
    defer cancel()

    time.Sleep(10 * time.Millisecond) // Ensure timeout fires

    _, err := client.Complete(ctx, "profile", req)
    if err != context.DeadlineExceeded {
        t.Errorf("Expected timeout, got: %v", err)
    }
}
```

## Rollback Plan

If you need to rollback to v0.5:

```bash
# Go module
go get github.com/not-emily/sage@v0.5.0

# CLI binary
curl -L https://github.com/not-emily/sage/releases/download/v0.5.0/sage-linux-amd64 \
  -o /usr/local/bin/sage && chmod +x /usr/local/bin/sage
```

Your v0.5 code will work without modification.

## Questions?

If you encounter migration issues not covered here, please [open an issue](https://github.com/not-emily/sage/issues).
