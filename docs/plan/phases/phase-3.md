# Phase 3: CLI Test Coverage

> **Depends on:** Phase 1 (mockable HTTP), Phase 2 (stable interfaces)
> **Enables:** Phase 4 (validated changes ready to document)
>
> See: [Full Plan](../plan.md)

## Goal

Achieve 70%+ test coverage on `internal/cli/` with focus on JSON output format stability.

## Key Deliverables

- Create `internal/cli/testutil.go` with mock client and helpers
- Test `complete.go` - all flags, JSON output, streaming, stdin
- Test `models.go` - JSON output, error handling
- Test `init.go` - initialization flow
- Test `profile.go` - CRUD operations, JSON output
- Test `provider.go` - CRUD operations, JSON output
- Achieve 70%+ coverage on `internal/cli/`
- Document JSON output format stability

## Files to Create

- `internal/cli/testutil.go` - Mock client, test helpers
- `internal/cli/complete_test.go` - Complete command tests
- `internal/cli/models_test.go` - Models command tests
- `internal/cli/init_test.go` - Init command tests
- `internal/cli/profile_test.go` - Profile command tests
- `internal/cli/provider_test.go` - Provider command tests

## Dependencies

**Internal:** Phase 1 (mockable HTTP client), Phase 2 (stable context-aware interfaces)

**External:** `testing` package (stdlib)

## Implementation Notes

### Test Infrastructure

Create shared test utilities in `testutil.go`:

```go
// internal/cli/testutil.go
package cli

import (
    "context"
    "io"
    "os"
    "testing"

    "github.com/not-emily/sage/pkg/sage"
)

// mockClient implements sage.Client interface for testing
type mockClient struct {
    completeFunc       func(ctx context.Context, profile string, req sage.Request) (*sage.Response, error)
    completeStreamFunc func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error)
    listModelsFunc     func(ctx context.Context, provider string) ([]sage.ModelInfo, error)
    getProfileFunc     func(profile string) (*sage.Profile, error)
}

func (m *mockClient) Complete(ctx context.Context, profile string, req sage.Request) (*sage.Response, error) {
    if m.completeFunc != nil {
        return m.completeFunc(ctx, profile, req)
    }
    return newMockResponse("test response"), nil
}

func (m *mockClient) CompleteStream(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
    if m.completeStreamFunc != nil {
        return m.completeStreamFunc(ctx, profile, req)
    }
    return newMockStreamChan([]string{"chunk1", "chunk2"}), nil
}

// ... other methods ...

// Helper to create mock response
func newMockResponse(content string) *sage.Response {
    return &sage.Response{
        Content: content,
        Model:   "test-model",
        Usage: sage.Usage{
            PromptTokens:     10,
            CompletionTokens: 20,
        },
    }
}

// Helper to create mock stream
func newMockStreamChan(chunks []string) <-chan sage.Chunk {
    ch := make(chan sage.Chunk)
    go func() {
        defer close(ch)
        for _, content := range chunks {
            ch <- sage.Chunk{Content: content, Done: false}
        }
        ch <- sage.Chunk{Done: true}
    }()
    return ch
}

// Helper to capture stdout
func captureStdout(t *testing.T, fn func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    fn()

    w.Close()
    os.Stdout = old

    out, _ := io.ReadAll(r)
    return string(out)
}

// Helper to capture stderr
func captureStderr(t *testing.T, fn func()) string {
    old := os.Stderr
    r, w, _ := os.Pipe()
    os.Stderr = w

    fn()

    w.Close()
    os.Stderr = old

    out, _ := io.ReadAll(r)
    return string(out)
}
```

### Table-Driven Test Pattern

Use table-driven tests for multiple scenarios:

```go
func TestCompleteCommand(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        mockClient  *mockClient
        wantOut     string
        wantErr     bool
        wantJSONKey string  // For JSON output validation
    }{
        {
            name: "basic completion",
            args: []string{"complete", "hello"},
            mockClient: &mockClient{
                completeStreamFunc: func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
                    return newMockStreamChan([]string{"response"}), nil
                },
            },
            wantOut: "response",
            wantErr: false,
        },
        {
            name: "json output",
            args: []string{"complete", "--json", "hello"},
            mockClient: &mockClient{
                completeFunc: func(ctx context.Context, profile string, req sage.Request) (*sage.Response, error) {
                    return newMockResponse("response"), nil
                },
            },
            wantJSONKey: "content",
            wantErr:     false,
        },
        {
            name:    "missing prompt",
            args:    []string{"complete"},
            wantErr: true,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Key Test Scenarios

**Complete Command:**
- Basic completion (default streaming)
- JSON output (`--json`)
- NDJSON streaming (`--json --stream`)
- With profile (`--profile=name`)
- With system message (`--system="You are..."`)
- With max tokens (`--max-tokens=100`)
- Stdin input (pipe mode)
- Missing prompt (error case)
- Invalid profile (error case)
- Flag ordering (flags must come before prompt)

**Models Command:**
- List models for provider
- JSON output (`--json`)
- Invalid provider (error case)

**Init Command:**
- Initialize new config
- Already initialized (error case)

**Profile Commands:**
- List profiles
- Add profile
- Remove profile
- Set default profile
- JSON output (`--json`)
- Invalid operations (error cases)

**Provider Commands:**
- List providers
- Add provider account
- Remove provider account
- JSON output (`--json`)
- Invalid operations (error cases)

## Sub-phases

### 3.1: Create Test Infrastructure

**Goal:** Build reusable test helpers

**Deliverables:**
- `testutil.go` with mock client
- Helper functions for stdout/stderr capture
- Common test fixtures

**Steps:**
1. Create `internal/cli/testutil.go`
2. Implement `mockClient` struct with all Client methods
3. Add helper functions (`newMockResponse`, `newMockStreamChan`)
4. Add stdout/stderr capture helpers
5. Write a simple smoke test to verify helpers work
6. Commit

### 3.2: Test Complete Command

**Goal:** Comprehensive coverage of complete command

**Deliverables:**
- `complete_test.go` with 15+ test cases
- Coverage of all flags and output modes
- JSON format validation

**Steps:**
1. Create `internal/cli/complete_test.go`
2. Write table-driven tests for:
   - Basic streaming completion
   - JSON output
   - NDJSON streaming
   - All flag combinations
   - Stdin input
   - Error cases
3. Validate JSON output structure
4. Run: `go test ./internal/cli/ -run TestComplete`
5. Commit

**JSON Output Validation:**
```go
func TestCompleteJSONOutput(t *testing.T) {
    // ... setup mock client ...

    output := captureStdout(t, func() {
        runComplete([]string{"--json", "test"})
    })

    var result map[string]interface{}
    if err := json.Unmarshal([]byte(output), &result); err != nil {
        t.Fatalf("Invalid JSON output: %v", err)
    }

    // Validate structure
    if _, ok := result["content"]; !ok {
        t.Error("Missing 'content' field")
    }
    if _, ok := result["model"]; !ok {
        t.Error("Missing 'model' field")
    }
    if _, ok := result["usage"]; !ok {
        t.Error("Missing 'usage' field")
    }
}
```

### 3.3: Test Other Commands

**Goal:** Coverage for models, init, profile, provider commands

**Deliverables:**
- `models_test.go` with JSON output validation
- `init_test.go` with initialization flow
- `profile_test.go` with CRUD operations
- `provider_test.go` with CRUD operations

**Steps:**
1. Create test files for each command
2. Write table-driven tests
3. Validate JSON outputs where applicable
4. Test error cases
5. Run: `go test ./internal/cli/`
6. Commit

## Validation

### Coverage Target
- [ ] Run: `go test -cover ./internal/cli/`
- [ ] Coverage reaches 70%+ on all CLI files
- [ ] `complete.go`: 70%+ coverage
- [ ] `models.go`: 70%+ coverage
- [ ] Other command files: 50%+ coverage (lower priority)

### Test Execution
- [ ] All tests pass: `go test ./internal/cli/`
- [ ] No flaky tests (run 10 times, all pass)
- [ ] Tests run quickly (< 1 second for full suite)

### JSON Output Validation
- [ ] Complete JSON structure tested
- [ ] Complete NDJSON streaming tested
- [ ] Models JSON structure tested
- [ ] Profile list JSON structure tested
- [ ] Provider list JSON structure tested

### Edge Cases Covered
- [ ] Missing required arguments
- [ ] Invalid flag values
- [ ] Invalid profiles/providers
- [ ] Stdin input (pipe mode)
- [ ] Flag ordering (flags before prompt)
- [ ] Empty input
- [ ] Error message formatting

### No External Dependencies
- [ ] Tests don't hit real APIs
- [ ] Tests don't require Ollama running
- [ ] Tests don't create real config files (use temp dirs)
- [ ] Tests are fully isolated (can run in parallel)

## Success Criteria

Phase 3 is complete when:

- [ ] `testutil.go` exists with mock client and helpers
- [ ] All 5 command files have corresponding test files
- [ ] Overall coverage reaches 70%+ on `internal/cli/`
- [ ] All tests pass: `go test ./internal/cli/`
- [ ] JSON output formats are validated in tests
- [ ] NDJSON streaming format is validated
- [ ] Flag parsing edge cases are tested
- [ ] Error messages are tested
- [ ] Stdin input handling is tested
- [ ] No tests hit real APIs or create real files
- [ ] Tests run in < 1 second
- [ ] Coverage report shows 70%+:
  ```bash
  go test -coverprofile=coverage.out ./internal/cli/
  go tool cover -func=coverage.out
  ```

## Testing Strategy

### What to Test

**High Priority:**
- JSON output formats (public API)
- Flag parsing (user-facing behavior)
- Error messages (user-facing)
- Basic command functionality

**Medium Priority:**
- Edge cases (empty input, invalid values)
- Stdin handling
- Profile/provider CRUD operations

**Low Priority:**
- Internal helper functions (covered by integration)
- Stdout formatting details

### What Not to Test

- Real LLM API responses (that's provider testing)
- Real file I/O (mock the Client)
- Network behavior (mock HTTP)
- Provider-specific logic (tested in provider tests)

### Mock Strategy

Mock at the `sage.Client` interface level:
- ✅ Mock Client methods (Complete, CompleteStream, etc.)
- ❌ Don't mock HTTP transport (too low-level)
- ❌ Don't mock providers directly (test through Client)

This gives us:
- Fast tests (no network)
- Isolated tests (no external dependencies)
- Focused tests (CLI logic only)

## Example Test

```go
func TestCompleteWithProfile(t *testing.T) {
    mock := &mockClient{
        completeStreamFunc: func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
            // Verify profile was passed correctly
            if profile != "test-profile" {
                t.Errorf("Expected profile 'test-profile', got '%s'", profile)
            }

            // Verify request
            if req.Prompt != "hello" {
                t.Errorf("Expected prompt 'hello', got '%s'", req.Prompt)
            }

            return newMockStreamChan([]string{"response"}), nil
        },
    }

    // TODO: Inject mock client into command
    // This may require refactoring CLI to accept client interface

    output := captureStdout(t, func() {
        runComplete([]string{"--profile=test-profile", "hello"})
    })

    if !strings.Contains(output, "response") {
        t.Errorf("Expected output to contain 'response', got '%s'", output)
    }
}
```

## Notes

**CLI Refactoring:**
May need to refactor CLI commands to accept a client interface for testing. Options:
1. Add global variable that can be overridden in tests (simplest)
2. Refactor commands to accept client parameter (cleaner, more work)
3. Use build tags to swap implementations (complex)

Recommend option 1 for pragmatism.

**Coverage vs Quality:**
70% coverage is a target, not a goal. Focus on testing important behavior:
- Public APIs (JSON output)
- User-facing behavior (flags, errors)
- Regression prevention (bugs found in use)

Don't chase 100% coverage on internal helpers.

**Future Enhancements:**
- Integration tests against real Ollama (optional, slow)
- Fuzzing for flag parsing
- Property-based tests for JSON output
