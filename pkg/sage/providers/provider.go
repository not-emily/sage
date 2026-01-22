// Package providers defines the LLM provider interface and registry.
package providers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
)

// ProviderField describes a configuration field required by a provider.
type ProviderField struct {
	Key      string `json:"key"`               // e.g., "api_key", "base_url"
	Label    string `json:"label"`             // e.g., "API Key", "Base URL"
	Required bool   `json:"required"`
	Secret   bool   `json:"secret"`            // stored in encrypted secrets
	Default  string `json:"default,omitempty"` // e.g., "http://localhost:11434"
}

// Provider is implemented by each LLM provider.
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "anthropic").
	Name() string

	// Complete sends a request and returns the full response.
	// ctx is used for cancellation and timeout control.
	Complete(ctx context.Context, req Request) (*Response, error)

	// CompleteStream sends a request and streams chunks.
	// ctx is used for cancellation; goroutine must respect ctx.Done().
	CompleteStream(ctx context.Context, req Request) (<-chan Chunk, error)

	// ListModels returns available models from this provider.
	// ctx is used for cancellation and timeout control.
	ListModels(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error)

	// Fields returns the configuration fields required by this provider.
	Fields() []ProviderField
}

// ModelInfo describes an available model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Request is the normalized request format for providers.
type Request struct {
	Model      string
	System     string
	Prompt     string
	MaxTokens  int
	APIKey     string       // Decrypted, passed in by client
	BaseURL    string       // Optional override
	HTTPClient *http.Client // HTTP client to use for requests
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
