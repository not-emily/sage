// Package sage provides a unified interface for LLM providers.
package sage

// Request is the input for a completion.
type Request struct {
	Prompt    string
	System    string
	MaxTokens int
}

// Response is the result of a completion.
type Response struct {
	Content string
	Model   string
	Usage   Usage
}

// Chunk is a streaming response piece.
type Chunk struct {
	Content string
	Done    bool
	Error   error
}

// Usage contains token counts.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
}

// Profile defines an LLM configuration.
type Profile struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Account  string `json:"account"`
	Model    string `json:"model"`
}

// ProviderAccount stores credentials for a provider account.
type ProviderAccount struct {
	Name   string `json:"name"`
	APIKey string `json:"api_key"` // Encrypted in storage
}

// ProviderInfo describes a configured provider.
type ProviderInfo struct {
	Name     string   `json:"name"`
	Accounts []string `json:"accounts"`
	BaseURL  string   `json:"base_url,omitempty"`
}
