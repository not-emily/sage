# sage v0.6.0: Production-Ready Library

> **Status:** Planning complete | Last updated: 2026-01-22
>
> **Audit Reference:** This plan addresses high-priority technical debt identified in [docs/audits/project-audit-2026-01-22.md](../audits/project-audit-2026-01-22.md)
>
> Phase files: [phases/](phases/)

## Overview

This release focuses on making sage a production-ready Go library while maintaining its CLI excellence. We're addressing three critical gaps identified in the architectural audit:

**1. Context support** - Enable request cancellation, timeouts, and observability integration
**2. HTTP configurability** - Give users control over network behavior and enable comprehensive testing
**3. CLI test coverage** - Ensure the public CLI interface (including JSON output) is tested and stable

These improvements will unblock hub-core integration and make sage suitable for production use in long-running services. The changes maintain sage's zero-dependency philosophy while adding essential production capabilities.

## Core Vision

### 1. Library-First, CLI-Compatible
Sage must work equally well as an importable library and standalone CLI. Breaking changes to library interfaces are acceptable at v0.5.x, but CLI behavior must remain stable. Design decisions prioritize long-term library use cases (hub-core integration).

### 2. Production-Grade Reliability
Request lifecycle must be controllable (cancellation, timeouts). Network behavior must be configurable (testing, proxies, connection pooling). Public interfaces must be tested to prevent regressions.

### 3. Zero-Dependency Discipline
All improvements must maintain stdlib-only constraint. No external testing frameworks or HTTP mocking libraries. Use `httptest.Server` and standard testing patterns.

### 4. Backward Compatibility Where Possible
HTTP client changes should not break existing code (nil checks, defaults). CLI behavior must remain unchanged (flags, output formats, error messages). Only break library interfaces where necessary for correctness.

## Requirements

### Must Have
- Add `context.Context` parameter to all Provider interface methods
- Update all three provider implementations (Anthropic, OpenAI, Ollama) to accept and use context
- Update Client methods to accept and pass context
- Update CLI commands to create and pass appropriate contexts
- Respect context cancellation in streaming operations
- Handle context deadlines in HTTP requests
- Add `HTTPClient *http.Client` field to Client struct
- Pass HTTP client through to providers
- Default to sensible timeout settings (5 minutes, not infinite)
- Maintain backward compatibility (nil client → use default)
- Update all provider HTTP calls to use configurable client
- Test all CLI commands (complete, init, models, profile, provider, update)
- Test flag parsing and validation
- Test JSON output format stability (--json flag)
- Test NDJSON streaming output (--json --stream)
- Test error message formatting
- Test stdin input handling
- Achieve 70%+ coverage on `internal/cli/`

### Nice to Have
- Add integration tests for providers using `httptest.Server`
- Document context patterns in README
- Add example showing HTTP client configuration
- Basic benchmarks to ensure no performance regression

### Out of Scope
- **Structured error types** - Deferred to v0.7.0 (separate effort)
- **Logging interface** - Deferred to v0.7.0 (needs design discussion)
- **Provider test coverage improvements** - Only add what's needed for HTTP client testing
- **Retry logic** - Not adding automatic retries
- **Middleware/interceptor pattern** - Future consideration
- **Context-aware progress reporting** - Beyond current needs
- **Connection pooling tuning** - Users can configure via HTTPClient if needed

## Constraints

**Technical:**
- Must maintain zero external dependencies (stdlib only)
- Must not break CLI behavior or output formats
- Must maintain Go 1.21+ compatibility
- Must pass all existing tests before adding new ones

**Timeline:**
- Target completion: Today (2026-01-22)
- Phases will be implemented sequentially

**Team:**
- Single developer (Emily) with Claude assistance
- Must be well-documented for future contributors

## Success Metrics

### Audit Remediation (Critical)

**Context support (#1 High Priority):** "No way to cancel requests, implement timeouts, or pass request-scoped values"
- [ ] Can cancel in-flight requests via context cancellation
- [ ] Can implement request timeouts via `context.WithTimeout`
- [ ] Can pass request-scoped values (e.g., trace IDs)
- [ ] No longer blocks adoption in production services

**HTTP client configurability (#2 High Priority):** "http.DefaultClient used directly in 8 places"
- [ ] HTTP client is configurable via Client struct
- [ ] Can configure timeouts, connection pooling
- [ ] Can inject custom transport for testing
- [ ] Can configure proxy or custom TLS settings
- [ ] Providers use shared HTTP client, not `http.DefaultClient`

**CLI test coverage (#3 High Priority):** "1,207 lines, 0 tests"
- [ ] All commands have test coverage
- [ ] JSON output format is tested (prevents breaking integrations)
- [ ] Argument parsing is tested
- [ ] Flag handling is tested
- [ ] Error formatting is tested
- [ ] Achieve 70%+ coverage on `internal/cli/`

### Audit Recommendations Met
- [ ] Breaking changes done at v0.5.x (audit: "Better to break now than later") ✓
- [ ] HTTP client passed to providers (audit recommendation) ✓
- [ ] Tests use HTTP mocking via configurable client ✓
- [ ] Context parameters enable future observability integration ✓

### Scalability Concerns Addressed
- [ ] HTTP client reuse (audit: "Each request creates its own HTTP connection")
  - Configurable client enables connection pooling
  - Users can tune `MaxIdleConns`, `IdleConnTimeout`, etc.

### Functionality
- [ ] All Provider methods accept and respect context.Context
- [ ] Context cancellation stops in-flight HTTP requests
- [ ] HTTP client is configurable via Client struct
- [ ] Default HTTP client has reasonable timeouts (5 minutes, not infinite)
- [ ] All CLI commands have test coverage

### Quality
- [ ] All existing tests pass
- [ ] No performance regressions (streaming still fast)
- [ ] Zero new external dependencies added
- [ ] `go test ./...` passes clean

### Documentation
- [ ] CHANGELOG.md updated with breaking changes
- [ ] README.md shows context usage examples
- [ ] README.md shows HTTP client configuration
- [ ] Migration guide for library users (v0.5 → v0.6)

### Integration
- [ ] sage CLI still works identically for end users
- [ ] Hub-core integration path is clearer (context support ready)

## Architecture Decisions

### 1. Context Threading Pattern

**Choice:** Thread `context.Context` through the entire call chain: CLI → Client → Provider → HTTP

**Rationale:** Standard Go pattern for cancellation/timeouts. Enables future observability (trace IDs, deadlines). Required for production use (web servers, long-running services).

**Trade-offs:** Breaking change to Provider interface (acceptable at v0.5.x). All provider implementations must be updated (3 files). CLI code must create and pass contexts (minimal impact).

**Alternatives Considered:**
- Add separate methods like `CompleteWithContext()` - Rejected because it creates interface bloat and half the library would be context-aware
- Make context optional via options pattern - Rejected because context should be ubiquitous in Go, not optional

### 2. HTTP Client Injection

**Choice:** Add `HTTPClient` field to Client struct, pass to providers via Request struct

**Implementation Pattern:**
```go
type Client struct {
    config     *Config
    secrets    *Secrets
    HTTPClient *http.Client  // nil = use sensible default
}

type Request struct {
    Prompt     string
    System     string
    MaxTokens  int
    HTTPClient *http.Client  // Provider uses this for all requests
}
```

**Rationale:** Users can configure timeouts, proxies, TLS, connection pooling. Enables testing via `httptest.Server` or custom transports. Solves audit concern about 8 hardcoded `http.DefaultClient` uses.

**Trade-offs:** Slight API surface increase. Providers need to handle nil client (use sensible default).

**User Impact:**
- **CLI users:** Nothing changes, get better defaults automatically
- **Library users:** Optional configuration, enables testing

### 3. Default HTTP Client Configuration

**Choice:** Provide sensible defaults with timeouts instead of infinite waits

**Implementation:**
```go
func defaultHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 5 * time.Minute,
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

**Rationale:** Current `http.DefaultClient` has infinite read timeout (dangerous). 5 minutes is reasonable for even slow LLM responses. Connection pooling improves performance for repeated requests. Users can override by setting their own HTTPClient.

**Trade-offs:** Changes default behavior (currently no timeout → now 5min timeout). Could break long-running requests (but unlikely, and configurable).

### 4. CLI Testing Architecture

**Choice:** Table-driven tests with mocked `sage.Client`

**Implementation Pattern:**
```go
type mockClient struct {
    completeFunc func(ctx context.Context, profile string, req Request) (*Response, error)
}

func TestCompleteCommand(t *testing.T) {
    tests := []struct {
        name       string
        args       []string
        mockClient *mockClient
        wantOut    string
        wantErr    bool
    }{
        // test cases
    }
}
```

**Rationale:** Can test CLI logic without hitting real LLM APIs. Table-driven tests make it easy to add edge cases. Standard Go testing patterns (no external frameworks).

**Trade-offs:** May need to refactor CLI to accept client interface. Tests won't catch real API integration issues (but that's what provider tests are for).

### 5. Context Cancellation in Streaming

**Choice:** Respect context cancellation in streaming goroutines, close channels properly

**Implementation Pattern:**
```go
for scanner.Scan() {
    select {
    case <-ctx.Done():
        chunks <- Chunk{Error: ctx.Err()}
        return
    default:
        // Process and send chunk
    }
}
```

**Rationale:** Prevents goroutine leaks when requests are cancelled. Immediate cancellation (doesn't wait for next chunk). Follows Go concurrency best practices.

**Trade-offs:** Slightly more complex streaming code (but safer).

## Project Structure

### Files to Modify

```
pkg/sage/
├── client.go              # ADD: HTTPClient field, update methods to accept context
├── types.go               # ADD: HTTPClient field to Request struct
└── providers/
    ├── provider.go        # MODIFY: Provider interface to accept context
    ├── anthropic.go       # MODIFY: Use ctx and req.HTTPClient
    ├── openai.go          # MODIFY: Use ctx and req.HTTPClient
    └── ollama.go          # MODIFY: Use ctx and req.HTTPClient

internal/cli/
├── complete.go            # MODIFY: Create and pass context
├── init.go                # MODIFY: Create and pass context (if needed)
├── models.go              # MODIFY: Create and pass context
├── profile.go             # MODIFY: Create and pass context (if needed)
├── provider.go            # MODIFY: Create and pass context (if needed)
└── root.go                # No changes needed

docs/
├── CHANGELOG.md           # UPDATE: v0.6.0 section
└── README.md              # UPDATE: Examples, configuration
```

### Files to Create

```
pkg/sage/providers/
├── anthropic_test.go      # NEW: HTTP client tests (optional, minimal)
├── openai_test.go         # NEW: HTTP client tests (optional, minimal)
└── ollama_test.go         # NEW: HTTP client tests (optional, minimal)

internal/cli/
├── complete_test.go       # NEW: Test completion command
├── init_test.go           # NEW: Test init command
├── models_test.go         # NEW: Test models command
├── profile_test.go        # NEW: Test profile commands
├── provider_test.go       # NEW: Test provider commands
└── testutil.go            # NEW: Shared test helpers (mock client, etc.)

docs/
├── plan/
│   ├── plan.md            # THIS FILE
│   └── phases/
│       ├── phase-1.md     # HTTP client configurability
│       ├── phase-2.md     # Context support
│       ├── phase-3.md     # CLI test coverage
│       └── phase-4.md     # Documentation updates
└── migration-v0.6.md      # NEW: Migration guide for library users
```

### Key Patterns

**Context Creation (CLI):**
- CLI commands create `context.Background()` at command entry point
- Pass through to client methods
- No cancellation handling needed (process exit is sufficient for CLI)

**HTTP Client Default:**
- `client.go` provides `defaultHTTPClient()` function
- Used when `Client.HTTPClient` is nil
- Request struct always has non-nil HTTPClient when passed to providers

**Test Helpers:**
- `testutil.go` provides mock client
- Shared test fixtures for common scenarios
- Helper functions for capturing stdout/stderr in tests

## Core Interfaces

### Provider Interface (Breaking Change)

```go
// pkg/sage/providers/provider.go
type Provider interface {
    Name() string

    // Complete performs a non-streaming completion request
    // ctx: Used for cancellation and timeout control
    Complete(ctx context.Context, req Request) (*Response, error)

    // CompleteStream performs a streaming completion request
    // ctx: Used for cancellation; goroutine must respect ctx.Done()
    CompleteStream(ctx context.Context, req Request) (<-chan Chunk, error)

    // ListModels retrieves available models for this provider
    // ctx: Used for cancellation and timeout control
    ListModels(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error)

    // Fields returns configuration fields required by this provider
    Fields() []ProviderField
}
```

**Breaking changes from v0.5.0:**
- `Complete`: Added `ctx context.Context` as first parameter
- `CompleteStream`: Added `ctx context.Context` as first parameter
- `ListModels`: Added `ctx context.Context` as first parameter

### Request Type (Non-Breaking Addition)

```go
// pkg/sage/types.go
type Request struct {
    Prompt     string
    System     string
    MaxTokens  int
    HTTPClient *http.Client  // NEW: Provider uses this for all HTTP requests
}
```

### Client Methods (Breaking Change for Library Users)

```go
// pkg/sage/client.go
type Client struct {
    config     *Config
    secrets    *Secrets
    HTTPClient *http.Client  // NEW: Optional; nil = use defaultHTTPClient()
}

// Complete performs a non-streaming completion
// ctx: Passed through to provider; enables cancellation and timeouts
func (c *Client) Complete(ctx context.Context, profile string, req Request) (*Response, error)

// CompleteStream performs a streaming completion
// ctx: Passed through to provider; enables cancellation
func (c *Client) CompleteStream(ctx context.Context, profile string, req Request) (<-chan Chunk, error)

// ListModels lists available models for a provider
// ctx: Passed through to provider; enables cancellation and timeouts
func (c *Client) ListModels(ctx context.Context, providerName string) ([]ModelInfo, error)
```

**Breaking changes from v0.5.0:**
- All methods now require `ctx context.Context` as first parameter
- `Client` struct has new `HTTPClient` field (non-breaking addition)

### Context Usage Patterns

**CLI Context Creation:**
```go
// All CLI commands create background context
ctx := context.Background()
```

**Library Context Usage:**
```go
// User controls timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
resp, err := client.Complete(ctx, "gpt4", req)

// User controls cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(5 * time.Second)
    cancel()  // Cancel after 5 seconds
}()
resp, err := client.Complete(ctx, "gpt4", req)
```

**Provider Context Handling:**
```go
// Respect context in HTTP requests
httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
resp, err := req.HTTPClient.Do(httpReq)

// Respect context in streaming
for scanner.Scan() {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process chunk
    }
}
```

## Implementation Phases

| Phase | Name | Scope | Depends On | Key Outputs |
|-------|------|-------|------------|-------------|
| 1 | HTTP Client Configurability | Add HTTPClient field, use in providers | — | Configurable client, sensible defaults |
| 2 | Context.Context Support | Add ctx parameter to all interfaces | Phase 1 | Context-aware API, cancellation support |
| 3 | CLI Test Coverage | Test all CLI commands | Phase 1, 2 | 70%+ CLI coverage, tested JSON output |
| 4 | Documentation Updates | Update docs for breaking changes | Phase 2, 3 | CHANGELOG, README, migration guide |

### Critical Path

**Sequential dependencies:**
- Phase 1 must complete before Phase 2 (same files, HTTP client needed for context testing)
- Phase 3 depends on Phase 1 (needs mockable HTTP) and Phase 2 (needs stable interfaces)
- Phase 4 depends on Phase 2 (breaking changes) and Phase 3 (validation complete)

**Cannot parallelize:** All phases touch overlapping files and the Provider interface.

### Phase Details
- [Phase 1: HTTP Client Configurability](phases/phase-1.md)
- [Phase 2: Context.Context Support](phases/phase-2.md)
- [Phase 3: CLI Test Coverage](phases/phase-3.md)
- [Phase 4: Documentation Updates](phases/phase-4.md)

## Tech Stack

| Category | Choice | Notes |
|----------|--------|-------|
| Language | Go 1.21+ | stdlib only, no dependencies |
| Testing | `testing` package | Table-driven tests, `httptest.Server` |
| HTTP | `net/http` | Configurable client, sensible defaults |
| Context | `context` | Standard Go cancellation pattern |

## Future Considerations

Items explicitly deferred from scope but architecturally supported:

- **Structured error types** (v0.7.0) - Context support enables attaching error metadata
- **Logging interface** (v0.7.0) - Context can carry logger instances
- **Retry logic** (future) - Context deadlines inform retry decisions
- **Middleware pattern** (future) - Context enables request/response interception
- **Observability hooks** (future) - Context carries trace IDs, spans
