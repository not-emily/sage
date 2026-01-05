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

const openaiDefaultURL = "https://api.openai.com/v1/chat/completions"

func init() {
	Register("openai", NewOpenAI)
}

type openai struct{}

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI() Provider {
	return &openai{}
}

func (o *openai) Name() string {
	return "openai"
}

// OpenAI API request/response types

type openaiRequest struct {
	Model               string          `json:"model"`
	Messages            []openaiMessage `json:"messages"`
	MaxTokens           int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
	Stream              bool            `json:"stream,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
	Usage   openaiUsage    `json:"usage"`
	Error   *openaiError   `json:"error,omitempty"`
}

type openaiChoice struct {
	Message openaiMessage `json:"message"`
	Delta   openaiMessage `json:"delta"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type openaiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func (o *openai) Complete(req Request) (*Response, error) {
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

	var openaiResp openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &Response{
		Content: openaiResp.Choices[0].Message.Content,
		Model:   req.Model,
		Usage: Usage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
		},
	}, nil
}

func (o *openai) CompleteStream(req Request) (<-chan Chunk, error) {
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

			// Check for end of stream
			if line == "data: [DONE]" {
				ch <- Chunk{Done: true}
				return
			}

			// Parse SSE data line
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var streamResp openaiResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				ch <- Chunk{Error: fmt.Errorf("failed to parse stream data: %w", err)}
				return
			}

			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					ch <- Chunk{Content: content}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- Chunk{Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return ch, nil
}

func (o *openai) buildRequest(req Request, stream bool) openaiRequest {
	messages := []openaiMessage{}

	if req.System != "" {
		messages = append(messages, openaiMessage{
			Role:    "system",
			Content: req.System,
		})
	}

	messages = append(messages, openaiMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	r := openaiRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   stream,
	}

	// Newer models (o1, o3, gpt-4o) use max_completion_tokens instead of max_tokens
	if req.MaxTokens > 0 {
		if o.usesMaxCompletionTokens(req.Model) {
			r.MaxCompletionTokens = req.MaxTokens
		} else {
			r.MaxTokens = req.MaxTokens
		}
	}

	return r
}

// usesMaxCompletionTokens returns true for models that require max_completion_tokens.
func (o *openai) usesMaxCompletionTokens(model string) bool {
	// Newer models use max_completion_tokens instead of max_tokens:
	// - o1, o3 reasoning models
	// - gpt-4o and newer (gpt-5, etc.)
	return strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.Contains(model, "gpt-4o") ||
		strings.Contains(model, "gpt-5")
}

func (o *openai) endpoint(req Request) string {
	if req.BaseURL != "" {
		return strings.TrimSuffix(req.BaseURL, "/") + "/v1/chat/completions"
	}
	return openaiDefaultURL
}

func (o *openai) setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
}

func (o *openai) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp openaiResponse
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

// ListModels returns available models from OpenAI.
func (o *openai) ListModels(apiKey, baseURL string) ([]ModelInfo, error) {
	endpoint := "https://api.openai.com/v1/models"
	if baseURL != "" {
		endpoint = strings.TrimSuffix(baseURL, "/") + "/v1/models"
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result openaiModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		// Filter to chat models (skip embeddings, audio, etc.)
		if strings.Contains(m.ID, "gpt") || strings.Contains(m.ID, "o1") || strings.Contains(m.ID, "o3") {
			models = append(models, ModelInfo{
				ID:   m.ID,
				Name: m.ID,
			})
		}
	}

	return models, nil
}

type openaiModelsResponse struct {
	Data []openaiModel `json:"data"`
}

type openaiModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}
