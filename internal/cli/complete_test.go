package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/not-emily/sage/pkg/sage"
)

func TestCompleteJSON(t *testing.T) {
	tests := []struct {
		name    string
		client  *mockClient
		profile string
		prompt  string
		wantErr bool
		checkFn func(t *testing.T, output string)
	}{
		{
			name:    "basic completion",
			client:  &mockClient{},
			profile: "test",
			prompt:  "hello",
			wantErr: false,
			checkFn: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Invalid JSON output: %v", err)
				}

				if content, ok := result["content"].(string); !ok || content != "test response" {
					t.Errorf("Expected content 'test response', got %v", result["content"])
				}

				if model, ok := result["model"].(string); !ok || model != "test-model" {
					t.Errorf("Expected model 'test-model', got %v", result["model"])
				}

				if usage, ok := result["usage"].(map[string]interface{}); !ok {
					t.Error("Expected usage field")
				} else {
					if pt, ok := usage["prompt_tokens"].(float64); !ok || pt != 10 {
						t.Errorf("Expected prompt_tokens 10, got %v", usage["prompt_tokens"])
					}
				}
			},
		},
		{
			name: "error handling",
			client: &mockClient{
				completeFunc: func(ctx context.Context, profile string, req sage.Request) (*sage.Response, error) {
					return nil, fmt.Errorf("test error")
				},
			},
			profile: "test",
			prompt:  "hello",
			wantErr: true,
		},
		{
			name: "custom response",
			client: &mockClient{
				completeFunc: func(ctx context.Context, profile string, req sage.Request) (*sage.Response, error) {
					return &sage.Response{
						Content: "custom response",
						Model:   "custom-model",
						Usage: sage.Usage{
							PromptTokens:     5,
							CompletionTokens: 15,
						},
					}, nil
				},
			},
			profile: "test",
			prompt:  "hello",
			wantErr: false,
			checkFn: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Invalid JSON output: %v", err)
				}

				if content := result["content"]; content != "custom response" {
					t.Errorf("Expected 'custom response', got %v", content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := sage.Request{
				Prompt: tt.prompt,
			}

			output := captureStdout(t, func() {
				err := completeJSON(tt.client, tt.profile, req)
				if (err != nil) != tt.wantErr {
					t.Errorf("completeJSON() error = %v, wantErr %v", err, tt.wantErr)
				}
			})

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, output)
			}
		})
	}
}

func TestCompleteStreamJSON(t *testing.T) {
	tests := []struct {
		name    string
		client  *mockClient
		profile string
		model   string
		prompt  string
		wantErr bool
		checkFn func(t *testing.T, output string)
	}{
		{
			name:    "basic streaming",
			client:  &mockClient{},
			profile: "test",
			model:   "test-model",
			prompt:  "hello",
			wantErr: false,
			checkFn: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 2 {
					t.Fatalf("Expected at least 2 lines of output, got %d", len(lines))
				}

				// Check first chunk
				var firstChunk map[string]interface{}
				if err := json.Unmarshal([]byte(lines[0]), &firstChunk); err != nil {
					t.Fatalf("Invalid JSON in first chunk: %v", err)
				}

				if content := firstChunk["content"]; content != "chunk1" {
					t.Errorf("Expected first chunk 'chunk1', got %v", content)
				}

				if done, ok := firstChunk["done"].(bool); !ok || done {
					t.Error("First chunk should not be done")
				}

				// Check last chunk (done=true)
				var lastChunk map[string]interface{}
				if err := json.Unmarshal([]byte(lines[len(lines)-1]), &lastChunk); err != nil {
					t.Fatalf("Invalid JSON in last chunk: %v", err)
				}

				if done, ok := lastChunk["done"].(bool); !ok || !done {
					t.Error("Last chunk should be done")
				}

				if model := lastChunk["model"]; model != "test-model" {
					t.Errorf("Expected model in last chunk, got %v", model)
				}
			},
		},
		{
			name: "streaming error",
			client: &mockClient{
				completeStreamFunc: func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
					return nil, fmt.Errorf("stream error")
				},
			},
			profile: "test",
			model:   "test-model",
			prompt:  "hello",
			wantErr: true,
		},
		{
			name: "chunk with error",
			client: &mockClient{
				completeStreamFunc: func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
					ch := make(chan sage.Chunk)
					go func() {
						defer close(ch)
						ch <- sage.Chunk{Content: "chunk", Done: false}
						ch <- sage.Chunk{Error: fmt.Errorf("chunk error")}
					}()
					return ch, nil
				},
			},
			profile: "test",
			model:   "test-model",
			prompt:  "hello",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := sage.Request{
				Prompt: tt.prompt,
			}

			output := captureStdout(t, func() {
				err := completeStreamJSON(tt.client, tt.profile, tt.model, req)
				if (err != nil) != tt.wantErr {
					t.Errorf("completeStreamJSON() error = %v, wantErr %v", err, tt.wantErr)
				}
			})

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, output)
			}
		})
	}
}

func TestCompleteStream(t *testing.T) {
	tests := []struct {
		name    string
		client  *mockClient
		profile string
		prompt  string
		wantErr bool
		checkFn func(t *testing.T, output string)
	}{
		{
			name:    "basic streaming output",
			client:  &mockClient{},
			profile: "test",
			prompt:  "hello",
			wantErr: false,
			checkFn: func(t *testing.T, output string) {
				if !strings.Contains(output, "chunk1") {
					t.Error("Expected output to contain 'chunk1'")
				}
				if !strings.Contains(output, "chunk2") {
					t.Error("Expected output to contain 'chunk2'")
				}
			},
		},
		{
			name: "streaming error",
			client: &mockClient{
				completeStreamFunc: func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
					return nil, fmt.Errorf("stream error")
				},
			},
			profile: "test",
			prompt:  "hello",
			wantErr: true,
		},
		{
			name: "chunk error",
			client: &mockClient{
				completeStreamFunc: func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
					ch := make(chan sage.Chunk)
					go func() {
						defer close(ch)
						ch <- sage.Chunk{Content: "start", Done: false}
						ch <- sage.Chunk{Error: fmt.Errorf("chunk error")}
					}()
					return ch, nil
				},
			},
			profile: "test",
			prompt:  "hello",
			wantErr: true,
			checkFn: func(t *testing.T, output string) {
				if !strings.Contains(output, "start") {
					t.Error("Expected partial output before error")
				}
			},
		},
		{
			name: "empty chunks",
			client: &mockClient{
				completeStreamFunc: func(ctx context.Context, profile string, req sage.Request) (<-chan sage.Chunk, error) {
					ch := make(chan sage.Chunk)
					go func() {
						defer close(ch)
						ch <- sage.Chunk{Content: "", Done: false}
						ch <- sage.Chunk{Content: "text", Done: false}
						ch <- sage.Chunk{Done: true}
					}()
					return ch, nil
				},
			},
			profile: "test",
			prompt:  "hello",
			wantErr: false,
			checkFn: func(t *testing.T, output string) {
				if !strings.Contains(output, "text") {
					t.Error("Expected output to contain 'text'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := sage.Request{
				Prompt: tt.prompt,
			}

			output := captureStdout(t, func() {
				err := completeStream(tt.client, tt.profile, req)
				if (err != nil) != tt.wantErr {
					t.Errorf("completeStream() error = %v, wantErr %v", err, tt.wantErr)
				}
			})

			if tt.checkFn != nil {
				tt.checkFn(t, output)
			}
		})
	}
}

func TestGetPrompt(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "single arg",
			args: []string{"hello"},
			want: "hello",
		},
		{
			name: "multiple args",
			args: []string{"hello", "world"},
			want: "hello world",
		},
		{
			name: "empty args",
			args: []string{},
			want: "",
		},
		{
			name: "args with spaces",
			args: []string{"hello", "beautiful", "world"},
			want: "hello beautiful world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPrompt(tt.args)
			if got != tt.want {
				t.Errorf("getPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}
