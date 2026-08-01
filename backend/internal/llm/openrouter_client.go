package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	Stream    bool          `json:"stream,omitempty"`
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

// ChatCompletionStreamResult is the fully-accumulated turn assembled from a chat-completions SSE stream --
// the streaming equivalent of one Choice from ChatCompletionResponse.
type ChatCompletionStreamResult struct {
	Message      ChatMessage
	FinishReason string
}

type chunkDelta struct {
	Content   string          `json:"content"`
	ToolCalls []toolCallDelta `json:"tool_calls"`
}

// ToolCallDelta.Index groups deltas belonging to the same tool call across chunks -- id/type/name arrive
// once (typically the first chunk for that index) and arguments accumulate incrementally.
type toolCallDelta struct {
	Index    int               `json:"index"`
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function functionCallDelta `json:"function"`
}

type functionCallDelta struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatCompletionChunk struct {
	Choices []struct {
		Delta        chunkDelta `json:"delta"`
		FinishReason string     `json:"finish_reason"`
	} `json:"choices"`
}

type toolCallAccumulator struct {
	id, callType, name string
	arguments          strings.Builder
}

// CreateChatCompletionStream streams a chat-completions turn, invoking onDelta with each piece of
// assistant text as it arrives, and returns the fully-accumulated message once the stream ends (a
// `data: [DONE]` line or EOF). Tool-call arguments are split across many chunks by OpenAI's own streaming
// schema, keyed by ToolCallDelta.Index, and are reassembled here rather than left for the caller to do.
func (c *OpenRouterClient) CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest, onDelta func(string) error) (*ChatCompletionStreamResult, error) {
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &OpenRouterAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var content strings.Builder
	var toolCalls []*toolCallAccumulator
	finishReason := ""

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}
		if data == "[DONE]" {
			break
		}

		var chunk chatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return nil, fmt.Errorf("decode stream chunk: %w", err)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		choice := chunk.Choices[0]

		if choice.Delta.Content != "" {
			content.WriteString(choice.Delta.Content)
			if err := onDelta(choice.Delta.Content); err != nil {
				return nil, fmt.Errorf("write delta: %w", err)
			}
		}
		for _, tc := range choice.Delta.ToolCalls {
			for len(toolCalls) <= tc.Index {
				toolCalls = append(toolCalls, &toolCallAccumulator{})
			}
			acc := toolCalls[tc.Index]
			if tc.ID != "" {
				acc.id = tc.ID
			}
			if tc.Type != "" {
				acc.callType = tc.Type
			}
			if tc.Function.Name != "" {
				acc.name = tc.Function.Name
			}
			acc.arguments.WriteString(tc.Function.Arguments)
		}
		if choice.FinishReason != "" {
			finishReason = choice.FinishReason
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stream: %w", err)
	}

	result := &ChatCompletionStreamResult{
		Message:      ChatMessage{Role: "assistant", Content: content.String()},
		FinishReason: finishReason,
	}
	for _, acc := range toolCalls {
		result.Message.ToolCalls = append(result.Message.ToolCalls, ToolCall{
			ID:       acc.id,
			Type:     acc.callType,
			Function: FunctionCall{Name: acc.name, Arguments: acc.arguments.String()},
		})
	}
	return result, nil
}
