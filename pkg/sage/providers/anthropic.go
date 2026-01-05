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

const (
	anthropicDefaultURL = "https://api.anthropic.com/v1/messages"
	anthropicVersion    = "2023-06-01"
)

func init() {
	Register("anthropic", NewAnthropic)
}

type anthropic struct{}

// NewAnthropic creates a new Anthropic provider.
func NewAnthropic() Provider {
	return &anthropic{}
}

func (a *anthropic) Name() string {
	return "anthropic"
}

// Anthropic API request/response types

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Usage   anthropicUsage     `json:"usage"`
	Error   *anthropicError    `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Streaming types
type anthropicStreamEvent struct {
	Type  string                `json:"type"`
	Delta *anthropicStreamDelta `json:"delta,omitempty"`
}

type anthropicStreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (a *anthropic) Complete(req Request) (*Response, error) {
	body := a.buildRequest(req, false)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", a.endpoint(req), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.setHeaders(httpReq, req.APIKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, a.handleError(resp)
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Extract text from first text content block
	var content string
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content = c.Text
			break
		}
	}

	return &Response{
		Content: content,
		Model:   req.Model,
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

func (a *anthropic) CompleteStream(req Request) (<-chan Chunk, error) {
	body := a.buildRequest(req, true)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", a.endpoint(req), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.setHeaders(httpReq, req.APIKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, a.handleError(resp)
	}

	ch := make(chan Chunk)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var currentEvent string

		for scanner.Scan() {
			line := scanner.Text()

			// Track event type
			if strings.HasPrefix(line, "event: ") {
				currentEvent = strings.TrimPrefix(line, "event: ")
				continue
			}

			// Skip empty lines
			if line == "" {
				continue
			}

			// Parse data line
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Handle message_stop event
			if currentEvent == "message_stop" {
				ch <- Chunk{Done: true}
				return
			}

			// Only process content_block_delta events
			if currentEvent != "content_block_delta" {
				continue
			}

			var event anthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				ch <- Chunk{Error: fmt.Errorf("failed to parse stream data: %w", err)}
				return
			}

			if event.Delta != nil && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				ch <- Chunk{Content: event.Delta.Text}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- Chunk{Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return ch, nil
}

func (a *anthropic) buildRequest(req Request, stream bool) anthropicRequest {
	messages := []anthropicMessage{
		{Role: "user", Content: req.Prompt},
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024 // Anthropic requires max_tokens
	}

	return anthropicRequest{
		Model:     req.Model,
		MaxTokens: maxTokens,
		System:    req.System, // Separate field, not in messages
		Messages:  messages,
		Stream:    stream,
	}
}

func (a *anthropic) endpoint(req Request) string {
	if req.BaseURL != "" {
		return strings.TrimSuffix(req.BaseURL, "/") + "/v1/messages"
	}
	return anthropicDefaultURL
}

func (a *anthropic) setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
}

func (a *anthropic) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Error *anthropicError `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("invalid API key: %s", errResp.Error.Message)
		case http.StatusTooManyRequests:
			return fmt.Errorf("rate limited: %s", errResp.Error.Message)
		default:
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error.Message)
		}
	}

	return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
}

// ListModels returns available Claude models.
// Anthropic doesn't have a models endpoint, so we return a hardcoded list.
func (a *anthropic) ListModels(apiKey, baseURL string) ([]ModelInfo, error) {
	// Hardcoded list of current Claude models
	return []ModelInfo{
		{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", Description: "Most capable model for complex tasks"},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Description: "Balanced performance and speed"},
		{ID: "claude-3-5-haiku-latest", Name: "Claude 3.5 Haiku", Description: "Fast and efficient for simple tasks"},
		{ID: "claude-3-5-sonnet-latest", Name: "Claude 3.5 Sonnet", Description: "Previous generation balanced model"},
		{ID: "claude-3-opus-latest", Name: "Claude 3 Opus", Description: "Previous generation top model"},
	}, nil
}
