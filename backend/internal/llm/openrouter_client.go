package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"

type OpenRouterAPIError struct {
	StatusCode int
	Body       string
}

func (e *OpenRouterAPIError) Error() string {
	return fmt.Sprintf("OpenRouter API returned status %d: %s", e.StatusCode, e.Body)
}

type OpenRouterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type OpenRouterOption func(*OpenRouterClient)

func WithOpenRouterBaseURL(url string) OpenRouterOption {
	return func(c *OpenRouterClient) { c.baseURL = url }
}

func NewOpenRouterClient(apiKey string, opts ...OpenRouterOption) *OpenRouterClient {
	c := &OpenRouterClient{
		apiKey:     apiKey,
		baseURL:    defaultOpenRouterBaseURL,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ChatMessage is one turn in an OpenAI-compatible chat-completions request or
// response: "system"/"user" messages set Content; "assistant" messages set
// either Content or ToolCalls; "tool" messages set ToolCallID and Content
// (the tool's result).
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall.Arguments is a JSON-encoded string (not a raw object), per the
// OpenAI tool-calling schema.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type ChatCompletionRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	Tools     []Tool        `json:"tools,omitempty"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

func (c *OpenRouterClient) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &OpenRouterAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
