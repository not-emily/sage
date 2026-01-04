package providers

import (
	"testing"
)

func TestOpenAI_Registered(t *testing.T) {
	// Verify openai is registered via init()
	if !Exists("openai") {
		t.Fatal("openai provider not registered")
	}

	p, err := Get("openai")
	if err != nil {
		t.Fatalf("Get(openai) error = %v", err)
	}

	if p.Name() != "openai" {
		t.Errorf("Name() = %q, want %q", p.Name(), "openai")
	}
}

func TestOpenAI_BuildRequest(t *testing.T) {
	o := &openai{}

	req := Request{
		Model:     "gpt-4o-mini",
		System:    "You are helpful",
		Prompt:    "Hello",
		MaxTokens: 100,
	}

	built := o.buildRequest(req, false)

	if built.Model != "gpt-4o-mini" {
		t.Errorf("Model = %q, want %q", built.Model, "gpt-4o-mini")
	}

	if len(built.Messages) != 2 {
		t.Fatalf("Messages count = %d, want 2", len(built.Messages))
	}

	if built.Messages[0].Role != "system" {
		t.Errorf("Messages[0].Role = %q, want %q", built.Messages[0].Role, "system")
	}

	if built.Messages[1].Role != "user" {
		t.Errorf("Messages[1].Role = %q, want %q", built.Messages[1].Role, "user")
	}

	if built.Stream != false {
		t.Error("Stream should be false")
	}
}

func TestOpenAI_BuildRequest_NoSystem(t *testing.T) {
	o := &openai{}

	req := Request{
		Model:  "gpt-4o-mini",
		Prompt: "Hello",
	}

	built := o.buildRequest(req, true)

	if len(built.Messages) != 1 {
		t.Fatalf("Messages count = %d, want 1", len(built.Messages))
	}

	if built.Messages[0].Role != "user" {
		t.Errorf("Messages[0].Role = %q, want %q", built.Messages[0].Role, "user")
	}

	if built.Stream != true {
		t.Error("Stream should be true")
	}
}

func TestOpenAI_Endpoint(t *testing.T) {
	o := &openai{}

	// Default endpoint
	req := Request{}
	if got := o.endpoint(req); got != openaiDefaultURL {
		t.Errorf("endpoint() = %q, want %q", got, openaiDefaultURL)
	}

	// Custom endpoint
	req.BaseURL = "https://custom.api.com"
	expected := "https://custom.api.com/v1/chat/completions"
	if got := o.endpoint(req); got != expected {
		t.Errorf("endpoint() = %q, want %q", got, expected)
	}

	// Custom endpoint with trailing slash
	req.BaseURL = "https://custom.api.com/"
	if got := o.endpoint(req); got != expected {
		t.Errorf("endpoint() = %q, want %q", got, expected)
	}
}
