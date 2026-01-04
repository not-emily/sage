package providers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const ollamaDefaultURL = "http://localhost:11434"

func init() {
	Register("ollama", NewOllama)
}

type ollama struct{}

// NewOllama creates a new Ollama provider.
func NewOllama() Provider {
	return &ollama{}
}

func (o *ollama) Name() string {
	return "ollama"
}

// Ollama API request/response types

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
	Error           string        `json:"error,omitempty"`
}

func (o *ollama) Complete(req Request) (*Response, error) {
	body := o.buildRequest(req, false)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", o.endpoint(req), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	o.setHeaders(httpReq, req.APIKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, o.handleError(resp)
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return &Response{
		Content: ollamaResp.Message.Content,
		Model:   req.Model,
		Usage: Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
		},
	}, nil
}

func (o *ollama) CompleteStream(req Request) (<-chan Chunk, error) {
	body := o.buildRequest(req, true)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", o.endpoint(req), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	o.setHeaders(httpReq, req.APIKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, o.handleError(resp)
	}

	ch := make(chan Chunk)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			var streamResp ollamaResponse
			if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
				ch <- Chunk{Error: fmt.Errorf("failed to parse stream data: %w", err)}
				return
			}

			if streamResp.Error != "" {
				ch <- Chunk{Error: fmt.Errorf("ollama error: %s", streamResp.Error)}
				return
			}

			// Send content chunk
			if streamResp.Message.Content != "" {
				ch <- Chunk{Content: streamResp.Message.Content}
			}

			// Check for completion
			if streamResp.Done {
				ch <- Chunk{Done: true}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- Chunk{Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return ch, nil
}

func (o *ollama) buildRequest(req Request, stream bool) ollamaRequest {
	messages := []ollamaMessage{}

	if req.System != "" {
		messages = append(messages, ollamaMessage{
			Role:    "system",
			Content: req.System,
		})
	}

	messages = append(messages, ollamaMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	return ollamaRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   stream,
	}
}

func (o *ollama) endpoint(req Request) string {
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = ollamaDefaultURL
	}
	return strings.TrimSuffix(baseURL, "/") + "/api/chat"
}

func (o *ollama) setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")

	// Optional auth - only set if API key provided
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

func (o *ollama) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp ollamaResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("ollama error (%d): %s", resp.StatusCode, errResp.Error)
	}

	return fmt.Errorf("ollama error (%d): %s", resp.StatusCode, string(body))
}
