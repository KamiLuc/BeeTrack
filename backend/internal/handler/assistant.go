package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

type AssistantHandler struct {
	assistant *service.AssistantService
}

func NewAssistantHandler(assistant *service.AssistantService) *AssistantHandler {
	return &AssistantHandler{assistant: assistant}
}

type assistantMessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type assistantMessagesRequest struct {
	Messages []assistantMessageRequest `json:"messages"`
}

func (req assistantMessagesRequest) toHistory() []service.AssistantMessage {
	history := make([]service.AssistantMessage, len(req.Messages))
	for i, m := range req.Messages {
		history[i] = service.AssistantMessage{Role: m.Role, Content: m.Content}
	}
	return history
}

// The client distinguishes delta/done/error via the SSE `event:` field.
type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (s sseWriter) WriteDelta(text string) error {
	return s.writeEvent("delta", map[string]string{"text": text})
}

func (s sseWriter) writeEvent(eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal sse payload: %w", err)
	}
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventType, data); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// Messages handles POST /api/v1/assistant/messages — runs the agent loop and
// streams the response back as Server-Sent Events.
func (h *AssistantHandler) Messages(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuth(w, r)
	if !ok {
		return
	}

	var req assistantMessagesRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	history := req.toHistory()
	if len(history) == 0 {
		respond.Error(w, http.StatusBadRequest, "EMPTY_MESSAGES", "messages must not be empty")
		return
	}
	if history[len(history)-1].Role != "user" {
		respond.Error(w, http.StatusBadRequest, "INVALID_LAST_MESSAGE", "the last message must be from the user")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		respond.Error(w, http.StatusInternalServerError, "STREAMING_UNSUPPORTED", "streaming is not supported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	sse := sseWriter{w: w, flusher: flusher}
	if err := h.assistant.Run(r.Context(), userID, history, sse); err != nil {
		_ = sse.writeEvent("error", map[string]string{"message": err.Error()})
		return
	}
	_ = sse.writeEvent("done", map[string]bool{"done": true})
}
