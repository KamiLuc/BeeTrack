package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenRouterClientSendsToolUseRequestAndParsesToolCallResponse(t *testing.T) {
	var gotAuth string
	var gotBody ChatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			ID: "gen_test",
			Choices: []Choice{
				{
					Message: ChatMessage{
						Role: "assistant",
						ToolCalls: []ToolCall{
							{ID: "call_1", Type: "function", Function: FunctionCall{Name: "list_hives", Arguments: `{"apiary_id":1}`}},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		})
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	resp, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model: "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hive 3 looked good"},
		},
		Tools: []Tool{
			{Type: "function", Function: FunctionDef{Name: "list_hives"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateChatCompletion returned error: %v", err)
	}

	if gotAuth != "Bearer test-api-key" {
		t.Errorf("expected Authorization header %q, got %q", "Bearer test-api-key", gotAuth)
	}
	if len(gotBody.Tools) != 1 {
		t.Fatalf("expected 1 tool in request, got %d", len(gotBody.Tools))
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	choice := resp.Choices[0]
	if choice.FinishReason != "tool_calls" {
		t.Errorf("expected finish_reason %q, got %q", "tool_calls", choice.FinishReason)
	}
	if len(choice.Message.ToolCalls) != 1 || choice.Message.ToolCalls[0].Function.Name != "list_hives" {
		t.Fatalf("expected a list_hives tool call, got %+v", choice.Message.ToolCalls)
	}
}

func TestOpenRouterClientTextOnlyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			ID: "gen_test",
			Choices: []Choice{
				{
					Message:      ChatMessage{Role: "assistant", Content: "Sounds like a healthy hive."},
					FinishReason: "stop",
				},
			},
		})
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	resp, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "Hive 3 looked good"}},
	})
	if err != nil {
		t.Fatalf("CreateChatCompletion returned error: %v", err)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	choice := resp.Choices[0]
	if choice.Message.Content != "Sounds like a healthy hive." {
		t.Errorf("expected text content, got %q", choice.Message.Content)
	}
	if len(choice.Message.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %+v", choice.Message.ToolCalls)
	}
}

func TestOpenRouterClientEmptyChoicesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatCompletionResponse{ID: "gen_test", Choices: []Choice{}})
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	resp, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("CreateChatCompletion returned error: %v", err)
	}
	if len(resp.Choices) != 0 {
		t.Fatalf("expected 0 choices, got %d", len(resp.Choices))
	}
}

func TestOpenRouterClientMalformedResponseBodyReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": "gen_test", "choices": [`))
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected error for malformed response body, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("expected error to mention response decoding, got: %v", err)
	}
}

func TestOpenRouterClientDefaultBaseURL(t *testing.T) {
	client := NewOpenRouterClient("test-api-key")
	if client.baseURL != defaultOpenRouterBaseURL {
		t.Errorf("expected default base URL %q, got %q", defaultOpenRouterBaseURL, client.baseURL)
	}
}

func TestOpenRouterClientNonOKResponseReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	client := NewOpenRouterClient("bad-api-key", WithOpenRouterBaseURL(server.URL))

	_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to mention status code 401, got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("expected error to include response body, got: %v", err)
	}
}
