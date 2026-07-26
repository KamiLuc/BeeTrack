package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func TestClientSendsToolUseRequestAndParsesToolCallResponse(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("x-api-key")

		var body anthropic.MessageNewParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.Tools) != 1 {
			t.Fatalf("expected 1 tool in request, got %d", len(body.Tools))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(anthropic.Message{
			ID: "msg_test",
			Content: []anthropic.ContentBlockUnion{
				{Type: "tool_use", ID: "toolu_1", Name: "list_hives", Input: json.RawMessage(`{"apiary_id":1}`)},
			},
			StopReason: anthropic.StopReasonToolUse,
		})
	}))
	defer server.Close()

	client := NewClient("test-api-key", option.WithBaseURL(server.URL))

	message, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Hive 3 looked good")),
		},
		Tools: []anthropic.ToolUnionParam{
			anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{}, "list_hives"),
		},
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotAuth != "test-api-key" {
		t.Errorf("expected x-api-key header %q, got %q", "test-api-key", gotAuth)
	}
	if message.StopReason != anthropic.StopReasonToolUse {
		t.Errorf("expected stop reason %q, got %q", anthropic.StopReasonToolUse, message.StopReason)
	}
	if len(message.Content) != 1 || message.Content[0].Name != "list_hives" {
		t.Fatalf("expected a list_hives tool_use block, got %+v", message.Content)
	}
}
