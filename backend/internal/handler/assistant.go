package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	ConversationID *int64 `json:"conversation_id"`
	Message        string `json:"message"`
}

// The client distinguishes conversation/delta/done/error via the SSE `event:` field.
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

	var req assistantMessageRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Message == "" {
		respond.Error(w, http.StatusBadRequest, "EMPTY_MESSAGE", "message must not be empty")
		return
	}

	conv, err := h.assistant.ResolveConversation(r.Context(), userID, req.ConversationID)
	if errors.Is(err, service.ErrAssistantConversationNotFound) {
		respond.Error(w, http.StatusNotFound, "CONVERSATION_NOT_FOUND", "conversation not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve conversation")
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
	_ = sse.writeEvent("conversation", map[string]int64{"conversation_id": conv.ID})
	if err := h.assistant.Run(r.Context(), userID, conv, req.Message, sse); err != nil {
		slog.ErrorContext(r.Context(), "assistant run failed", "conversation_id", conv.ID, "error", err)
		_ = sse.writeEvent("error", map[string]string{"message": err.Error()})
		return
	}
	_ = sse.writeEvent("done", map[string]bool{"done": true})
}
