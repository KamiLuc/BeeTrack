package llm

import (
	"context"
	"encoding/json"
	"errors"
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

func writeSSE(w http.ResponseWriter, lines ...string) {
	for _, l := range lines {
		w.Write([]byte("data: " + l + "\n\n"))
	}
}

func TestOpenRouterClientStreamForwardsTextDeltasAndAccumulatesContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var gotBody ChatCompletionRequest
		json.NewDecoder(r.Body).Decode(&gotBody)
		if !gotBody.Stream {
			t.Errorf("expected stream=true in request body")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"content":"Hello "},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"content":"world"},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			"[DONE]",
		)
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	var deltas []string
	result, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}, func(text string) error {
		deltas = append(deltas, text)
		return nil
	})
	if err != nil {
		t.Fatalf("CreateChatCompletionStream returned error: %v", err)
	}

	if len(deltas) != 2 || deltas[0] != "Hello " || deltas[1] != "world" {
		t.Errorf("expected 2 forwarded text deltas, got %+v", deltas)
	}
	if result.Message.Content != "Hello world" {
		t.Errorf("expected accumulated content %q, got %q", "Hello world", result.Message.Content)
	}
	if result.FinishReason != "stop" {
		t.Errorf("expected finish_reason %q, got %q", "stop", result.FinishReason)
	}
	if len(result.Message.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %+v", result.Message.ToolCalls)
	}
}

func TestOpenRouterClientStreamReassemblesToolCallArgumentsAcrossChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"list_hives","arguments":""}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"apiary_id\":"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"1}"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
			"[DONE]",
		)
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	result, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hive 3 looked good"}},
	}, func(string) error { return nil })
	if err != nil {
		t.Fatalf("CreateChatCompletionStream returned error: %v", err)
	}

	if result.FinishReason != "tool_calls" {
		t.Errorf("expected finish_reason %q, got %q", "tool_calls", result.FinishReason)
	}
	if len(result.Message.ToolCalls) != 1 {
		t.Fatalf("expected 1 reassembled tool call, got %d", len(result.Message.ToolCalls))
	}
	tc := result.Message.ToolCalls[0]
	if tc.ID != "call_1" || tc.Type != "function" || tc.Function.Name != "list_hives" {
		t.Errorf("unexpected tool call identity: %+v", tc)
	}
	if tc.Function.Arguments != `{"apiary_id":1}` {
		t.Errorf("expected reassembled arguments %q, got %q", `{"apiary_id":1}`, tc.Function.Arguments)
	}
}

func TestOpenRouterClientStreamReassemblesMultipleToolCallsInOneTurn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"list_hives","arguments":"{}"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"call_2","type":"function","function":{"name":"get_dashboard_summary","arguments":"{}"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
			"[DONE]",
		)
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	result, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}, func(string) error { return nil })
	if err != nil {
		t.Fatalf("CreateChatCompletionStream returned error: %v", err)
	}

	if len(result.Message.ToolCalls) != 2 {
		t.Fatalf("expected 2 reassembled tool calls, got %d", len(result.Message.ToolCalls))
	}
	if result.Message.ToolCalls[0].ID != "call_1" || result.Message.ToolCalls[0].Function.Name != "list_hives" {
		t.Errorf("unexpected first tool call: %+v", result.Message.ToolCalls[0])
	}
	if result.Message.ToolCalls[1].ID != "call_2" || result.Message.ToolCalls[1].Function.Name != "get_dashboard_summary" {
		t.Errorf("unexpected second tool call: %+v", result.Message.ToolCalls[1])
	}
}

func TestOpenRouterClientStreamReassemblesToolCallFieldsArrivingOutOfOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSE(w,
			// Index 1 (second tool call) is streamed before index 0 arrives at all.
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"call_2","type":"function","function":{"name":"get_dashboard_summary","arguments":""}}]},"finish_reason":null}]}`,
			// id/type/name for index 0 only arrive on a later chunk than its first argument fragment.
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"apiary_id\":"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"list_hives","arguments":"1}"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{}"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
			"[DONE]",
		)
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	result, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}, func(string) error { return nil })
	if err != nil {
		t.Fatalf("CreateChatCompletionStream returned error: %v", err)
	}

	if len(result.Message.ToolCalls) != 2 {
		t.Fatalf("expected 2 reassembled tool calls, got %d", len(result.Message.ToolCalls))
	}
	first, second := result.Message.ToolCalls[0], result.Message.ToolCalls[1]
	if first.ID != "call_1" || first.Type != "function" || first.Function.Name != "list_hives" || first.Function.Arguments != `{"apiary_id":1}` {
		t.Errorf("unexpected first tool call after out-of-order reassembly: %+v", first)
	}
	if second.ID != "call_2" || second.Function.Name != "get_dashboard_summary" || second.Function.Arguments != "{}" {
		t.Errorf("unexpected second tool call after out-of-order reassembly: %+v", second)
	}
}

func TestOpenRouterClientStreamPropagatesOnDeltaError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSE(w, `{"choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`)
	}))
	defer server.Close()

	client := NewOpenRouterClient("test-api-key", WithOpenRouterBaseURL(server.URL))

	_, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}, func(string) error { return errors.New("client disconnected") })
	if err == nil {
		t.Fatal("expected error when onDelta fails, got nil")
	}
}

func TestOpenRouterClientStreamNonOKResponseReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	client := NewOpenRouterClient("bad-api-key", WithOpenRouterBaseURL(server.URL))

	_, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	}, func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to mention status code 401, got: %v", err)
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
