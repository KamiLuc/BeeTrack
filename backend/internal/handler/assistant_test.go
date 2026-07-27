package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/beetrack/backend/internal/mcp"
	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/token"
)

const testAssistantAuthSecret = "test-assistant-secret"

// fakeAssistantDecoder replays one canned text turn ending in end_turn, just
// enough to exercise the handler's SSE framing without a real Anthropic call.
type fakeAssistantDecoder struct {
	events []ssestream.Event
	i      int
}

func (d *fakeAssistantDecoder) Next() bool {
	if d.i >= len(d.events) {
		return false
	}
	d.i++
	return true
}
func (d *fakeAssistantDecoder) Event() ssestream.Event { return d.events[d.i-1] }
func (d *fakeAssistantDecoder) Close() error           { return nil }
func (d *fakeAssistantDecoder) Err() error             { return nil }

type fakeAssistantMessageStreamer struct{}

func (f *fakeAssistantMessageStreamer) NewStreaming(_ context.Context, _ anthropic.MessageNewParams, _ ...option.RequestOption) *ssestream.Stream[anthropic.MessageStreamEventUnion] {
	events := []ssestream.Event{
		{Type: "content_block_start", Data: []byte(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)},
		{Type: "content_block_delta", Data: []byte(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hi there"}}`)},
		{Type: "message_delta", Data: []byte(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":3}}`)},
	}
	return ssestream.NewStream[anthropic.MessageStreamEventUnion](&fakeAssistantDecoder{events: events}, nil)
}

func newTestAssistantHandler() *AssistantHandler {
	svc := service.NewAssistantService(&fakeAssistantMessageStreamer{}, mcp.NewRegistry())
	return NewAssistantHandler(svc)
}

func assistantRequest(t *testing.T, userID int64, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := newTestAssistantHandler()
	handler := middleware.Auth(testAssistantAuthSecret)(http.HandlerFunc(h.Messages))

	r := httptest.NewRequest("POST", "/api/v1/assistant/messages", strings.NewReader(body))
	if userID != 0 {
		tok, err := token.NewAccessToken(userID, testAssistantAuthSecret, 5)
		if err != nil {
			t.Fatalf("generate token: %v", err)
		}
		r.Header.Set("Authorization", "Bearer "+tok)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestAssistantMessagesRequiresAuth(t *testing.T) {
	w := assistantRequest(t, 0, `{"messages":[{"role":"user","content":"hi"}]}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAssistantMessagesRejectsEmptyMessages(t *testing.T) {
	w := assistantRequest(t, 7, `{"messages":[]}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAssistantMessagesRejectsNonUserLastMessage(t *testing.T) {
	w := assistantRequest(t, 7, `{"messages":[{"role":"user","content":"hi"},{"role":"assistant","content":"hello"}]}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// erroringAssistantDecoder fails immediately, simulating an upstream Claude/stream error.
type erroringAssistantDecoder struct{}

func (erroringAssistantDecoder) Next() bool         { return false }
func (erroringAssistantDecoder) Event() ssestream.Event { return ssestream.Event{} }
func (erroringAssistantDecoder) Close() error       { return nil }
func (erroringAssistantDecoder) Err() error         { return context.DeadlineExceeded }

type erroringAssistantMessageStreamer struct{}

func (f *erroringAssistantMessageStreamer) NewStreaming(_ context.Context, _ anthropic.MessageNewParams, _ ...option.RequestOption) *ssestream.Stream[anthropic.MessageStreamEventUnion] {
	return ssestream.NewStream[anthropic.MessageStreamEventUnion](erroringAssistantDecoder{}, nil)
}

func TestAssistantMessagesStreamsErrorEventOnTurnRunnerFailure(t *testing.T) {
	svc := service.NewAssistantService(&erroringAssistantMessageStreamer{}, mcp.NewRegistry())
	h := NewAssistantHandler(svc)
	handler := middleware.Auth(testAssistantAuthSecret)(http.HandlerFunc(h.Messages))

	r := httptest.NewRequest("POST", "/api/v1/assistant/messages", strings.NewReader(`{"messages":[{"role":"user","content":"hi"}]}`))
	tok, err := token.NewAccessToken(7, testAssistantAuthSecret, 5)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (headers already sent before the stream failed), got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "event: error") {
		t.Errorf("expected an error event, got: %s", w.Body.String())
	}
}

func TestAssistantMessagesStreamsSSEDeltasAndDone(t *testing.T) {
	w := assistantRequest(t, 7, `{"messages":[{"role":"user","content":"hi"}]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected text/event-stream content type, got %q", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "event: delta") || !strings.Contains(body, `"hi there"`) {
		t.Errorf("expected a delta event with the streamed text, got: %s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Errorf("expected a done event, got: %s", body)
	}
}
