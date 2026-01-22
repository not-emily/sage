package cli

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/not-emily/sage/pkg/sage"
)

// mockClient implements sage.Client methods for testing
type mockClient struct {
	completeFunc       func(ctx context.Context, profile string, req sage.Request) (*sage.Response, error)
	completeStreamFunc func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error)
	listModelsFunc     func(ctx context.Context, provider, account string) ([]sage.ModelInfo, error)
	getProfileFunc     func(name string) (*sage.Profile, error)
	listProfilesFunc   func() []sage.Profile
	addProfileFunc     func(name string, p sage.Profile) error
	removeProfileFunc  func(name string) error
	listProvidersFunc  func() []sage.ProviderInfo
}

func (m *mockClient) Complete(ctx context.Context, profile string, req sage.Request) (*sage.Response, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, profile, req)
	}
	return &sage.Response{
		Content: "test response",
		Model:   "test-model",
		Usage: sage.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
		},
	}, nil
}

func (m *mockClient) CompleteStream(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
	if m.completeStreamFunc != nil {
		return m.completeStreamFunc(ctx, profile, req)
	}
	ch := make(chan sage.Chunk)
	go func() {
		defer close(ch)
		ch <- sage.Chunk{Content: "chunk1", Done: false}
		ch <- sage.Chunk{Content: "chunk2", Done: false}
		ch <- sage.Chunk{Done: true}
	}()
	return ch, nil
}

func (m *mockClient) ListModels(ctx context.Context, provider, account string) ([]sage.ModelInfo, error) {
	if m.listModelsFunc != nil {
		return m.listModelsFunc(ctx, provider, account)
	}
	return []sage.ModelInfo{
		{ID: "model-1", Name: "Test Model 1"},
		{ID: "model-2", Name: "Test Model 2"},
	}, nil
}

func (m *mockClient) GetProfile(name string) (*sage.Profile, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(name)
	}
	return &sage.Profile{
		Name:     name,
		Provider: "test-provider",
		Account:  "test-account",
		Model:    "test-model",
	}, nil
}

func (m *mockClient) ListProfiles() []sage.Profile {
	if m.listProfilesFunc != nil {
		return m.listProfilesFunc()
	}
	return []sage.Profile{
		{Name: "default", Provider: "test", Account: "test", Model: "test"},
	}
}

func (m *mockClient) AddProfile(name string, p sage.Profile) error {
	if m.addProfileFunc != nil {
		return m.addProfileFunc(name, p)
	}
	return nil
}

func (m *mockClient) RemoveProfile(name string) error {
	if m.removeProfileFunc != nil {
		return m.removeProfileFunc(name)
	}
	return nil
}

func (m *mockClient) ListProviders() []sage.ProviderInfo {
	if m.listProvidersFunc != nil {
		return m.listProvidersFunc()
	}
	return []sage.ProviderInfo{
		{Name: "test-provider", Accounts: []string{"test-account"}},
	}
}

// captureStdout captures stdout during function execution
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// captureStderr captures stderr during function execution
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// newMockResponse creates a test response
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

// newMockStreamChan creates a test streaming channel
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
