# Library Usage

Sage can be imported as a Go library for programmatic LLM access.

## Installation

```bash
go get github.com/not-emily/sage/pkg/sage
```

## Prerequisites

Before using the library, ensure sage is initialized (this creates the config and encryption key):

```bash
sage init
sage provider add openai
sage profile add default --provider=openai --model=gpt-4o-mini
sage profile set-default default
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/not-emily/sage/pkg/sage"
)

func main() {
    // Create client (loads config and secrets)
    client, err := sage.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    // Send completion request using default profile
    resp, err := client.Complete("", sage.Request{
        Prompt: "What is the capital of France?",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
    fmt.Printf("Tokens: %d prompt, %d completion\n",
        resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
}
```

## Using a Specific Profile

```go
// Use a named profile instead of default
resp, err := client.Complete("claude", sage.Request{
    Prompt: "Explain monads simply",
})
```

## System Prompts

```go
resp, err := client.Complete("", sage.Request{
    System: "You are a helpful coding assistant. Be concise.",
    Prompt: "How do I reverse a string in Go?",
})
```

## Streaming Responses

```go
ch, err := client.CompleteStream("", sage.Request{
    Prompt: "Write a short poem about Go",
})
if err != nil {
    log.Fatal(err)
}

for chunk := range ch {
    if chunk.Error != nil {
        log.Fatal(chunk.Error)
    }
    fmt.Print(chunk.Content)
}
fmt.Println()
```

## Max Tokens

```go
resp, err := client.Complete("", sage.Request{
    Prompt:    "Explain quantum computing",
    MaxTokens: 100, // Limit response length
})
```

## Profile Management

```go
// List all profiles
profiles := client.ListProfiles()
for _, p := range profiles {
    fmt.Printf("%s: %s/%s\n", p.Name, p.Provider, p.Model)
}

// Get default profile name
defaultName := client.GetDefaultProfile()

// Get specific profile
profile, err := client.GetProfile("claude")

// Add a new profile
err = client.AddProfile("fast", sage.Profile{
    Provider: "openai",
    Account:  "default",
    Model:    "gpt-4o-mini",
})

// Remove a profile
err = client.RemoveProfile("old-profile")

// Set default
err = client.SetDefaultProfile("fast")
```

## Provider Account Management

```go
// List providers
providers := client.ListProviders()
for _, p := range providers {
    fmt.Printf("%s: %v\n", p.Name, p.Accounts)
}

// Check if account exists
if client.HasProviderAccount("openai", "work") {
    // ...
}

// Add provider account (stores encrypted API key)
err = client.AddProviderAccount("openai", "work", "sk-...")

// Remove provider account
err = client.RemoveProviderAccount("openai", "work")
```

## Types Reference

### Request

```go
type Request struct {
    System    string // System prompt (optional)
    Prompt    string // User prompt (required)
    MaxTokens int    // Max response tokens (0 = provider default)
}
```

### Response

```go
type Response struct {
    Content string // Response text
    Model   string // Model that generated response
    Usage   Usage  // Token usage
}

type Usage struct {
    PromptTokens     int
    CompletionTokens int
}
```

### Chunk (Streaming)

```go
type Chunk struct {
    Content string // Partial response text
    Done    bool   // True when stream is complete
    Error   error  // Non-nil if an error occurred
}
```

### Profile

```go
type Profile struct {
    Name     string // Profile name (set when retrieved)
    Provider string // Provider name (openai, anthropic, ollama)
    Account  string // Provider account name
    Model    string // Model identifier
}
```

## Error Handling

```go
resp, err := client.Complete("", sage.Request{
    Prompt: "Hello",
})
if err != nil {
    // Common errors:
    // - "profile not found: X" — profile doesn't exist
    // - "provider not configured: X" — provider account missing
    // - HTTP errors from provider (401, 429, 500, etc.)
    log.Fatal(err)
}
```

## Integration Pattern (Hub-core Example)

For applications that need role-based LLM access:

```go
type LLMUtility struct {
    client *sage.Client
    roles  map[string]string // role -> profile name
}

func NewLLMUtility(roles map[string]string) (*LLMUtility, error) {
    client, err := sage.NewClient()
    if err != nil {
        return nil, err
    }
    return &LLMUtility{client: client, roles: roles}, nil
}

func (u *LLMUtility) Complete(role string, prompt string) (string, error) {
    profileName, ok := u.roles[role]
    if !ok {
        return "", fmt.Errorf("unknown role: %s", role)
    }

    resp, err := u.client.Complete(profileName, sage.Request{
        Prompt: prompt,
    })
    if err != nil {
        return "", err
    }
    return resp.Content, nil
}

// Usage:
// llm, _ := NewLLMUtility(map[string]string{
//     "small_brain": "gpt-4o-mini",
//     "big_brain":   "claude-sonnet",
// })
// result, _ := llm.Complete("small_brain", "Summarize this...")
```
