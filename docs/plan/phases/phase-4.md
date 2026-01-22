# Phase 4: Documentation Updates

> **Depends on:** Phase 2 (context changes complete), Phase 3 (validation complete)
> **Enables:** Public release of v0.6.0
>
> See: [Full Plan](../plan.md)

## Goal

Update documentation to reflect breaking changes and new capabilities without over-documenting.

## Key Deliverables

- Update CHANGELOG.md with v0.6.0 breaking changes
- Update README.md with context usage examples
- Add migration guide for library users (v0.5 → v0.6)
- Update code examples in README to use context
- Document HTTP client configuration options
- Verify all existing examples still work

## Files to Modify

- `CHANGELOG.md` - Add v0.6.0 section
- `README.md` - Update examples, add configuration section

## Files to Create

- `docs/migration-v0.6.md` - Migration guide for library users

## Dependencies

**Internal:** Phase 2 (context changes complete), Phase 3 (validation complete)

**External:** None

## Implementation Notes

### Documentation Principles

**Keep it concise:**
- Only document what changed
- Focus on library users (they're affected by breaking changes)
- CLI users don't need migration docs (no changes for them)

**Show examples:**
- Before/after code for breaking changes
- Working code snippets (not pseudocode)
- Real use cases (timeout, cancellation, custom client)

**No over-documentation:**
- Don't document internal implementation details
- Don't duplicate information across files
- Link to migration guide instead of repeating

## CHANGELOG.md Updates

Add v0.6.0 section at the top:

```markdown
## [0.6.0] - 2026-01-22

### Breaking Changes

**Library API**: All Client and Provider methods now require `context.Context` as first parameter.

- `Complete(ctx context.Context, profile string, req Request)`
- `CompleteStream(ctx context.Context, profile string, req Request)`
- `ListModels(ctx context.Context, provider string)`

See [docs/migration-v0.6.md](docs/migration-v0.6.md) for upgrading from v0.5.

**CLI users**: No breaking changes. CLI behavior is identical to v0.5.

### Added

- **Context support**: Enable request cancellation, timeouts, and tracing integration
- **HTTP client configurability**: Configure timeouts, proxies, and connection pooling via `Client.HTTPClient` field
- **Sensible default timeouts**: 5-minute timeout instead of infinite (configurable)
- **CLI test coverage**: 70%+ coverage on CLI commands, ensuring stability

### Fixed

- Infinite HTTP timeouts could cause hangs (now defaults to 5 minutes)
- No way to cancel long-running requests (now supports context cancellation)
- Providers hardcoded `http.DefaultClient` (now configurable)

### Security

- Default HTTP client includes timeouts to prevent resource exhaustion
- Connection pooling limits prevent excessive resource usage

## [0.5.0] - 2026-01-16

...
```

## README.md Updates

### Add Library Usage Section

Add a new "Library Usage" section after CLI usage:

```markdown
## Library Usage

Sage can be used as a Go library for programmatic access to LLM providers.

### Installation

```bash
go get github.com/not-emily/sage
```

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/not-emily/sage/pkg/sage"
)

func main() {
    client, err := sage.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    resp, err := client.Complete(ctx, "gpt4", sage.Request{
        Prompt: "What is the meaning of life?",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
```

### With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.Complete(ctx, "gpt4", sage.Request{
    Prompt: "Explain quantum computing",
})
```

### With Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel on user interrupt
go func() {
    <-sigChan  // Wait for signal
    cancel()
}()

resp, err := client.Complete(ctx, "gpt4", req)
if err == context.Canceled {
    fmt.Println("Request cancelled by user")
}
```

### Custom HTTP Client

Configure timeouts, proxies, or connection pooling:

```go
client := &sage.Client{
    HTTPClient: &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            Proxy:               http.ProxyFromEnvironment,
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    },
}
```

### Streaming

```go
ctx := context.Background()
chunks, err := client.CompleteStream(ctx, "gpt4", sage.Request{
    Prompt: "Write a story",
})
if err != nil {
    log.Fatal(err)
}

for chunk := range chunks {
    if chunk.Error != nil {
        log.Fatal(chunk.Error)
    }
    if chunk.Done {
        break
    }
    fmt.Print(chunk.Content)
}
```

### Migration from v0.5

If you're upgrading from v0.5, see [docs/migration-v0.6.md](docs/migration-v0.6.md).
```

### Update Existing Examples

Search for any existing code examples and add `ctx` parameter:

**Before:**
```go
resp, err := client.Complete("gpt4", req)
```

**After:**
```go
ctx := context.Background()
resp, err := client.Complete(ctx, "gpt4", req)
```

## Migration Guide (NEW FILE)

Create `docs/migration-v0.6.md`:

```markdown
# Migration Guide: v0.5 → v0.6

## Overview

Version 0.6 adds context.Context support to enable request cancellation, timeouts, and future observability features. This is a **breaking change for library users only**. CLI users are not affected.

## Breaking Changes

### All Client Methods Require Context

**Before (v0.5):**
```go
resp, err := client.Complete("gpt4", req)
chunks, err := client.CompleteStream("gpt4", req)
models, err := client.ListModels("openai")
```

**After (v0.6):**
```go
ctx := context.Background()
resp, err := client.Complete(ctx, "gpt4", req)
chunks, err := client.CompleteStream(ctx, "gpt4", req)
models, err := client.ListModels(ctx, "openai")
```

### Provider Interface Changes

If you've implemented custom providers, update the interface:

**Before (v0.5):**
```go
func (p *MyProvider) Complete(req Request) (*Response, error) {
    // ...
}
```

**After (v0.6):**
```go
func (p *MyProvider) Complete(ctx context.Context, req Request) (*Response, error) {
    // Use context in HTTP requests
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
    // ...
}
```

## New Capabilities

### Request Timeouts

Set per-request timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.Complete(ctx, "gpt4", req)
if err == context.DeadlineExceeded {
    fmt.Println("Request timed out")
}
```

### Request Cancellation

Cancel in-flight requests:

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(5 * time.Second)
    cancel()  // Cancel after 5 seconds
}()

resp, err := client.Complete(ctx, "gpt4", req)
if err == context.Canceled {
    fmt.Println("Request cancelled")
}
```

### Custom HTTP Client

Configure HTTP behavior:

```go
client := &sage.Client{
    HTTPClient: &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            Proxy:            http.ProxyFromEnvironment,
            MaxIdleConns:     100,
            IdleConnTimeout:  90 * time.Second,
        },
    },
}
```

## CLI Users

**No changes required.** The CLI works identically to v0.5:

```bash
sage complete "hello"
sage complete --json "hello"
sage models openai
```

## Migration Checklist

- [ ] Add `context.Context` parameter to all `Complete()` calls
- [ ] Add `context.Context` parameter to all `CompleteStream()` calls
- [ ] Add `context.Context` parameter to all `ListModels()` calls
- [ ] Consider adding timeouts where appropriate
- [ ] Consider adding cancellation for long-running operations
- [ ] Update custom provider implementations (if any)
- [ ] Run tests: `go test ./...`
- [ ] Update to v0.6: `go get github.com/not-emily/sage@v0.6.0`

## Benefits

- **Cancellation**: Stop requests mid-flight when no longer needed
- **Timeouts**: Prevent requests from hanging indefinitely
- **Observability**: Future support for distributed tracing via context
- **Production-ready**: Essential for web servers and long-running services

## Questions?

- See examples in [README.md](../README.md#library-usage)
- Check [CHANGELOG.md](../CHANGELOG.md) for full v0.6.0 changes
- Open an issue if you encounter migration problems
```

## Validation

### Documentation Quality
- [ ] CHANGELOG.md accurately describes all changes
- [ ] README.md examples compile and run
- [ ] Migration guide covers all breaking changes
- [ ] Code examples use correct syntax
- [ ] No outdated examples remain

### Links and References
- [ ] Internal links work (migration guide, changelog)
- [ ] Code blocks have proper syntax highlighting
- [ ] Examples are copy-pasteable (no placeholders)

### Completeness
- [ ] Breaking changes documented in CHANGELOG
- [ ] New features documented in CHANGELOG
- [ ] Bug fixes documented in CHANGELOG
- [ ] Library usage examples in README
- [ ] Migration path clear in migration guide
- [ ] CLI users know they're unaffected

### Code Examples
- [ ] Basic example compiles
- [ ] Timeout example compiles
- [ ] Cancellation example compiles
- [ ] Custom HTTP client example compiles
- [ ] Streaming example compiles

### Testing Examples

Verify examples actually work:

```bash
# Create test file from README example
cat > /tmp/test_sage.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/not-emily/sage/pkg/sage"
)

func main() {
    client, err := sage.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    resp, err := client.Complete(ctx, "default", sage.Request{
        Prompt: "test",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
EOF

# Test compilation
cd /tmp && go mod init test && go get github.com/not-emily/sage@latest && go build test_sage.go
```

## Success Criteria

Phase 4 is complete when:

- [ ] CHANGELOG.md has v0.6.0 section with breaking changes
- [ ] README.md has Library Usage section with context examples
- [ ] `docs/migration-v0.6.md` exists with migration guide
- [ ] All code examples compile
- [ ] All code examples use context correctly
- [ ] Links between docs work
- [ ] No outdated v0.5 examples remain
- [ ] HTTP client configuration is documented
- [ ] CLI users know they're unaffected
- [ ] Library users have clear migration path

## Notes

**Keep it practical:**
- Users want copy-paste examples, not explanations
- Show real use cases (timeout, cancellation)
- Don't over-document internal details

**Link, don't duplicate:**
- CHANGELOG points to migration guide
- README points to migration guide
- Don't repeat breaking change details everywhere

**Version numbering:**
- This is v0.6.0 (breaking change in pre-1.0)
- If we were post-1.0, this would be v2.0.0
- Semantic versioning: MAJOR.MINOR.PATCH

**Release notes:**
When creating GitHub release, include:
- Breaking changes summary
- Link to migration guide
- Benefits of upgrading
- CLI users unaffected
