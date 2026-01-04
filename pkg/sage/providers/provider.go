// Package providers defines the LLM provider interface and registry.
package providers

import (
	"fmt"
	"sort"
)

// Provider is implemented by each LLM provider.
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "anthropic").
	Name() string

	// Complete sends a request and returns the full response.
	Complete(req Request) (*Response, error)

	// CompleteStream sends a request and streams chunks.
	CompleteStream(req Request) (<-chan Chunk, error)
}

// Request is the normalized request format for providers.
type Request struct {
	Model     string
	System    string
	Prompt    string
	MaxTokens int
	APIKey    string // Decrypted, passed in by client
	BaseURL   string // Optional override
}

// Response is the normalized response from providers.
type Response struct {
	Content string
	Model   string
	Usage   Usage
}

// Usage contains token counts.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
}

// Chunk is a streaming response piece.
type Chunk struct {
	Content string
	Done    bool
	Error   error
}

// Constructor is a function that creates a new Provider instance.
type Constructor func() Provider

// registry maps provider names to constructors.
var registry = map[string]Constructor{}

// Register adds a provider constructor to the registry.
// This is typically called from provider init() functions.
func Register(name string, constructor Constructor) {
	registry[name] = constructor
}

// Get returns a provider by name.
func Get(name string) (Provider, error) {
	constructor, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return constructor(), nil
}

// List returns all available provider names in sorted order.
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Exists checks if a provider is registered.
func Exists(name string) bool {
	_, ok := registry[name]
	return ok
}
