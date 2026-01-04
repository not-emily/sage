package providers

import (
	"net/http"
	"testing"
)

func TestOllama_Registered(t *testing.T) {
	if !Exists("ollama") {
		t.Fatal("ollama provider not registered")
	}

	p, err := Get("ollama")
	if err != nil {
		t.Fatalf("Get(ollama) error = %v", err)
	}

	if p.Name() != "ollama" {
		t.Errorf("Name() = %q, want %q", p.Name(), "ollama")
	}
}

func TestOllama_BuildRequest(t *testing.T) {
	o := &ollama{}

	req := Request{
		Model:  "llama3.1:8b",
		System: "You are helpful",
		Prompt: "Hello",
	}

	built := o.buildRequest(req, false)

	if built.Model != "llama3.1:8b" {
		t.Errorf("Model = %q, want %q", built.Model, "llama3.1:8b")
	}

	// Messages should have system + user
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

func TestOllama_BuildRequest_NoSystem(t *testing.T) {
	o := &ollama{}

	req := Request{
		Model:  "llama3.1:8b",
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

func TestOllama_Endpoint(t *testing.T) {
	o := &ollama{}

	// Default endpoint (localhost)
	req := Request{}
	expected := "http://localhost:11434/api/chat"
	if got := o.endpoint(req); got != expected {
		t.Errorf("endpoint() = %q, want %q", got, expected)
	}

	// Custom endpoint
	req.BaseURL = "http://remote-server:11434"
	expected = "http://remote-server:11434/api/chat"
	if got := o.endpoint(req); got != expected {
		t.Errorf("endpoint() = %q, want %q", got, expected)
	}

	// Custom endpoint with trailing slash
	req.BaseURL = "http://remote-server:11434/"
	if got := o.endpoint(req); got != expected {
		t.Errorf("endpoint() = %q, want %q", got, expected)
	}
}

func TestOllama_SetHeaders_NoAuth(t *testing.T) {
	o := &ollama{}

	req, _ := http.NewRequest("POST", "https://example.com", nil)
	o.setHeaders(req, "") // No API key

	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	// Should NOT have Authorization header
	if got := req.Header.Get("Authorization"); got != "" {
		t.Errorf("Authorization = %q, want empty (no auth)", got)
	}
}

func TestOllama_SetHeaders_WithAuth(t *testing.T) {
	o := &ollama{}

	req, _ := http.NewRequest("POST", "https://example.com", nil)
	o.setHeaders(req, "test-api-key")

	if got := req.Header.Get("Authorization"); got != "Bearer test-api-key" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer test-api-key")
	}
}
