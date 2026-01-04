package providers

import (
	"testing"
)

// mockProvider is a test provider implementation.
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Complete(req Request) (*Response, error) {
	return &Response{Content: "mock response", Model: req.Model}, nil
}

func (m *mockProvider) CompleteStream(req Request) (<-chan Chunk, error) {
	ch := make(chan Chunk, 1)
	ch <- Chunk{Content: "mock", Done: true}
	close(ch)
	return ch, nil
}

func TestRegisterAndGet(t *testing.T) {
	// Clear registry for test
	originalRegistry := registry
	registry = map[string]Constructor{}
	defer func() { registry = originalRegistry }()

	// Register a mock provider
	Register("mock", func() Provider {
		return &mockProvider{name: "mock"}
	})

	// Get it back
	p, err := Get("mock")
	if err != nil {
		t.Fatalf("Get(mock) error = %v", err)
	}

	if p.Name() != "mock" {
		t.Errorf("Name() = %q, want %q", p.Name(), "mock")
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("nonexistent-provider-xyz")
	if err == nil {
		t.Error("Get(unknown) should return error")
	}
}

func TestList(t *testing.T) {
	// Clear registry for test
	originalRegistry := registry
	registry = map[string]Constructor{}
	defer func() { registry = originalRegistry }()

	// Register multiple providers
	Register("zebra", func() Provider { return &mockProvider{name: "zebra"} })
	Register("alpha", func() Provider { return &mockProvider{name: "alpha"} })
	Register("beta", func() Provider { return &mockProvider{name: "beta"} })

	names := List()

	if len(names) != 3 {
		t.Fatalf("List() returned %d items, want 3", len(names))
	}

	// Should be sorted
	expected := []string{"alpha", "beta", "zebra"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("List()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestExists(t *testing.T) {
	// Clear registry for test
	originalRegistry := registry
	registry = map[string]Constructor{}
	defer func() { registry = originalRegistry }()

	Register("test", func() Provider { return &mockProvider{name: "test"} })

	if !Exists("test") {
		t.Error("Exists(test) should return true")
	}

	if Exists("unknown") {
		t.Error("Exists(unknown) should return false")
	}
}

func TestProviderInterface(t *testing.T) {
	// Verify mock provider satisfies the interface
	var _ Provider = (*mockProvider)(nil)

	p := &mockProvider{name: "test"}

	// Test Complete
	resp, err := p.Complete(Request{Model: "test-model", Prompt: "hello"})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "mock response" {
		t.Errorf("Content = %q, want %q", resp.Content, "mock response")
	}

	// Test CompleteStream
	ch, err := p.CompleteStream(Request{Model: "test-model", Prompt: "hello"})
	if err != nil {
		t.Fatalf("CompleteStream() error = %v", err)
	}

	chunk := <-ch
	if !chunk.Done {
		t.Error("chunk.Done should be true")
	}
}
