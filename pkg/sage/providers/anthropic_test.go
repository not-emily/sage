package providers

import (
	"net/http"
	"testing"
)

func TestAnthropic_Registered(t *testing.T) {
	if !Exists("anthropic") {
		t.Fatal("anthropic provider not registered")
	}

	p, err := Get("anthropic")
	if err != nil {
		t.Fatalf("Get(anthropic) error = %v", err)
	}

	if p.Name() != "anthropic" {
		t.Errorf("Name() = %q, want %q", p.Name(), "anthropic")
	}
}

func TestAnthropic_BuildRequest(t *testing.T) {
	a := &anthropic{}

	req := Request{
		Model:     "claude-3-5-sonnet-20241022",
		System:    "You are helpful",
		Prompt:    "Hello",
		MaxTokens: 100,
	}

	built := a.buildRequest(req, false)

	if built.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Model = %q, want %q", built.Model, "claude-3-5-sonnet-20241022")
	}

	// System should be separate field, not in messages
	if built.System != "You are helpful" {
		t.Errorf("System = %q, want %q", built.System, "You are helpful")
	}

	// Messages should only have user message
	if len(built.Messages) != 1 {
		t.Fatalf("Messages count = %d, want 1", len(built.Messages))
	}

	if built.Messages[0].Role != "user" {
		t.Errorf("Messages[0].Role = %q, want %q", built.Messages[0].Role, "user")
	}

	if built.MaxTokens != 100 {
		t.Errorf("MaxTokens = %d, want %d", built.MaxTokens, 100)
	}

	if built.Stream != false {
		t.Error("Stream should be false")
	}
}

func TestAnthropic_BuildRequest_DefaultMaxTokens(t *testing.T) {
	a := &anthropic{}

	req := Request{
		Model:  "claude-3-5-sonnet-20241022",
		Prompt: "Hello",
		// MaxTokens not set
	}

	built := a.buildRequest(req, true)

	// Should default to 1024 since Anthropic requires max_tokens
	if built.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %d, want %d (default)", built.MaxTokens, 1024)
	}

	if built.Stream != true {
		t.Error("Stream should be true")
	}
}

func TestAnthropic_Endpoint(t *testing.T) {
	a := &anthropic{}

	// Default endpoint
	req := Request{}
	if got := a.endpoint(req); got != anthropicDefaultURL {
		t.Errorf("endpoint() = %q, want %q", got, anthropicDefaultURL)
	}

	// Custom endpoint
	req.BaseURL = "https://custom.api.com"
	expected := "https://custom.api.com/v1/messages"
	if got := a.endpoint(req); got != expected {
		t.Errorf("endpoint() = %q, want %q", got, expected)
	}

	// Custom endpoint with trailing slash
	req.BaseURL = "https://custom.api.com/"
	if got := a.endpoint(req); got != expected {
		t.Errorf("endpoint() = %q, want %q", got, expected)
	}
}

func TestAnthropic_SetHeaders(t *testing.T) {
	a := &anthropic{}

	req, _ := http.NewRequest("POST", "https://example.com", nil)
	a.setHeaders(req, "test-api-key")

	if got := req.Header.Get("x-api-key"); got != "test-api-key" {
		t.Errorf("x-api-key = %q, want %q", got, "test-api-key")
	}

	if got := req.Header.Get("anthropic-version"); got != anthropicVersion {
		t.Errorf("anthropic-version = %q, want %q", got, anthropicVersion)
	}

	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
}
