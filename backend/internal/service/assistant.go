package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/beetrack/backend/internal/llm"
	"github.com/beetrack/backend/internal/mcp"
	"github.com/beetrack/backend/internal/model"
	"gorm.io/datatypes"
)

const (
	defaultAssistantModel = "anthropic/claude-haiku-4.5"
	assistantMaxTokens    = 2048
	assistantMaxToolTurns = 8
	// assistantMaxMessagesPerConversation keeps conversations short enough to stay within a single LLM
	// context window and bound per-conversation API cost; counts logged rows (user + assistant), so 10 is
	// 5 round trips.
	assistantMaxMessagesPerConversation = 10
	assistantSystemPrompt               = "You are the BeeTrack apiary assistant, helping a beekeeper understand their own hives and browse the public marketplace. Only use the information the tools return; never invent hive data, statuses, or listings. Always reply in the same language the user's message is written in. Your scope covers three kinds of questions: (1) the caller's own data via tools (hives, apiaries, listings), (2) general beekeeping knowledge and practices (e.g. how to merge hives, treat varroa, judge brood pattern) — answer these directly from your own knowledge even when no tool applies, (3) the BeeTrack app itself. Only decline questions clearly outside all three, e.g. unrelated small talk, other topics, or requests to write code. When referring to a hive, apiary, or listing, always use its name/title from the tool result (e.g. \"Rapeseed Honey\") rather than its raw numeric id; only mention the id if the user explicitly asks for it or several results share the same name."
)

var (
	ErrAssistantEmptyMessage             = errors.New("message must not be empty")
	ErrAssistantTooManyToolTurns         = errors.New("assistant exceeded the maximum number of tool-call turns")
	ErrAssistantConversationNotFound     = errors.New("conversation not found")
	ErrAssistantConversationLimitReached = errors.New("conversation has reached the maximum number of messages")
)

type AssistantStreamWriter interface {
	WriteDelta(text string) error
}

// Satisfied by *llm.OpenRouterClient; kept as its own interface so the agent loop is testable without a
// real OpenRouter connection.
type chatStreamer interface {
	CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, onDelta func(string) error) (*llm.ChatCompletionStreamResult, error)
}

// Kept separate from AssistantService's tool-orchestration loop so that loop is testable with canned
// *llm.ChatCompletionStreamResult responses, without also having to fake the SSE parsing/reassembly
// internals (those live in and are tested by internal/llm's OpenRouterClient).
type turnRunner interface {
	runTurn(ctx context.Context, req llm.ChatCompletionRequest, w AssistantStreamWriter) (*llm.ChatCompletionStreamResult, error)
}

type streamTurnRunner struct {
	client chatStreamer
}

func (r streamTurnRunner) runTurn(ctx context.Context, req llm.ChatCompletionRequest, w AssistantStreamWriter) (*llm.ChatCompletionStreamResult, error) {
	return r.client.CreateChatCompletionStream(ctx, req, w.WriteDelta)
}

// AssistantStore is the persistence interface for assistant conversations and their message/tool-call
// trail, satisfied by *repository.AssistantRepository.
type AssistantStore interface {
	CreateConversation(ctx context.Context, userID int64) (*model.AssistantConversation, error)
	GetConversationByID(ctx context.Context, id int64) (*model.AssistantConversation, error)
	ListConversations(ctx context.Context, userID int64) ([]*model.AssistantConversationSummary, error)
	DeleteConversation(ctx context.Context, id int64) error
	ListMessageLogs(ctx context.Context, conversationID int64) ([]*model.AssistantMessageLog, error)
	CountMessageLogs(ctx context.Context, conversationID int64) (int64, error)
	CreateMessageLog(ctx context.Context, log *model.AssistantMessageLog) error
	CreateToolCall(ctx context.Context, call *model.AssistantToolCall) error
}

// userID always comes from the caller, never from the model's own input.
type AssistantService struct {
	turns    turnRunner
	registry *mcp.Registry
	repo     AssistantStore
	model    string
}

// client is typically *llm.OpenRouterClient. model selects the OpenRouter model slug for the agent loop --
// pass "" to use defaultAssistantModel (see config.OpenRouterModel for the env var that controls this in
// main.go).
func NewAssistantService(client chatStreamer, registry *mcp.Registry, repo AssistantStore, model string) *AssistantService {
	if model == "" {
		model = defaultAssistantModel
	}
	return &AssistantService{turns: streamTurnRunner{client: client}, registry: registry, repo: repo, model: model}
}

// ResolveConversation loads conversationID if given (rejecting one that doesn't exist or belongs to a
// different user), or starts a new conversation for userID if nil.
func (s *AssistantService) ResolveConversation(ctx context.Context, userID int64, conversationID *int64) (*model.AssistantConversation, error) {
	if conversationID == nil {
		return s.repo.CreateConversation(ctx, userID)
	}
	conv, err := s.repo.GetConversationByID(ctx, *conversationID)
	if err != nil {
		return nil, err
	}
	if conv == nil || conv.UserID != userID {
		return nil, ErrAssistantConversationNotFound
	}
	return conv, nil
}

// CheckMessageLimit rejects conversationID once it holds assistantMaxMessagesPerConversation logged
// messages. Callers must run this before committing SSE response headers, since a limit hit needs to come
// back as a normal JSON error rather than a mid-stream one.
func (s *AssistantService) CheckMessageLimit(ctx context.Context, conversationID int64) error {
	count, err := s.repo.CountMessageLogs(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("count conversation messages: %w", err)
	}
	if count >= assistantMaxMessagesPerConversation {
		return ErrAssistantConversationLimitReached
	}
	return nil
}

// ListConversations returns userID's conversations, newest first, for the chat history sidebar.
func (s *AssistantService) ListConversations(ctx context.Context, userID int64) ([]*model.AssistantConversationSummary, error) {
	return s.repo.ListConversations(ctx, userID)
}

// ConversationMessages returns conversationID's full message trail so the client can resume it, rejecting
// one that doesn't exist or belongs to a different user the same way ResolveConversation does.
func (s *AssistantService) ConversationMessages(ctx context.Context, userID, conversationID int64) ([]*model.AssistantMessageLog, error) {
	conv, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conv == nil || conv.UserID != userID {
		return nil, ErrAssistantConversationNotFound
	}
	return s.repo.ListMessageLogs(ctx, conversationID)
}

func (s *AssistantService) DeleteConversation(ctx context.Context, userID, conversationID int64) error {
	conv, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}
	if conv == nil || conv.UserID != userID {
		return ErrAssistantConversationNotFound
	}
	return s.repo.DeleteConversation(ctx, conversationID)
}

func (s *AssistantService) toolParams() []llm.Tool {
	tools := s.registry.All()
	params := make([]llm.Tool, len(tools))
	for i, t := range tools {
		schema, _ := json.Marshal(map[string]any{
			"type":       "object",
			"properties": t.InputSchema.Properties,
			"required":   t.InputSchema.Required,
		})
		params[i] = llm.Tool{
			Type: "function",
			Function: llm.FunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  schema,
			},
		}
	}
	return params
}

func toMessageParams(history []*model.AssistantMessageLog) []llm.ChatMessage {
	params := make([]llm.ChatMessage, len(history))
	for i, m := range history {
		role := "user"
		if m.Role == model.AssistantMessageRoleAssistant {
			role = "assistant"
		}
		params[i] = llm.ChatMessage{Role: role, Content: m.Content}
	}
	return params
}

// toolResultMessage builds the "tool" role message fed back for one tool call. Unlike Anthropic's
// tool_result content blocks, OpenAI-compatible schema has no separate is_error flag on a tool message --
// a failed call is just conveyed as error text in Content.
func toolResultMessage(toolCallID string, result any, callErr error) llm.ChatMessage {
	if callErr != nil {
		return llm.ChatMessage{Role: "tool", ToolCallID: toolCallID, Content: callErr.Error()}
	}
	data, err := json.Marshal(result)
	if err != nil {
		return llm.ChatMessage{Role: "tool", ToolCallID: toolCallID, Content: err.Error()}
	}
	return llm.ChatMessage{Role: "tool", ToolCallID: toolCallID, Content: string(data)}
}

// toolCallLog builds the assistant_tool_calls row for one MCP call — Result holds the marshaled tool
// result, or the error message when callErr is set, mirroring toolResultMessage's own success/error shape.
func toolCallLog(name string, input json.RawMessage, result any, callErr error) *model.AssistantToolCall {
	var resultJSON []byte
	if callErr != nil {
		resultJSON, _ = json.Marshal(map[string]string{"error": callErr.Error()})
	} else {
		resultJSON, _ = json.Marshal(result)
	}
	return &model.AssistantToolCall{
		ToolName: name,
		Input:    datatypes.JSON(input),
		Result:   datatypes.JSON(resultJSON),
		IsError:  callErr != nil,
	}
}

// Logging failures are reported but never returned, so a DB hiccup here can't fail an SSE response
// already in flight.
func (s *AssistantService) persistUserMessage(ctx context.Context, conversationID int64, message string) {
	log := &model.AssistantMessageLog{ConversationID: conversationID, Role: model.AssistantMessageRoleUser, Content: message}
	if err := s.repo.CreateMessageLog(ctx, log); err != nil {
		slog.Error("persist assistant user message", "conversation_id", conversationID, "error", err)
	}
}

func (s *AssistantService) persistAssistantTurn(ctx context.Context, conversationID int64, text string, toolCalls []*model.AssistantToolCall) {
	log := &model.AssistantMessageLog{ConversationID: conversationID, Role: model.AssistantMessageRoleAssistant, Content: text}
	if err := s.repo.CreateMessageLog(ctx, log); err != nil {
		slog.Error("persist assistant reply", "conversation_id", conversationID, "error", err)
		return
	}
	for _, call := range toolCalls {
		call.MessageID = log.ID
		if err := s.repo.CreateToolCall(ctx, call); err != nil {
			slog.Error("persist assistant tool call", "conversation_id", conversationID, "message_id", log.ID, "tool_name", call.ToolName, "error", err)
		}
	}
}

// Text streams to w on every turn, not just the final one. conv must already be resolved (and its
// ownership checked) via ResolveConversation before calling Run.
func (s *AssistantService) Run(ctx context.Context, userID int64, conv *model.AssistantConversation, message string, w AssistantStreamWriter) error {
	if message == "" {
		return ErrAssistantEmptyMessage
	}

	history, err := s.repo.ListMessageLogs(ctx, conv.ID)
	if err != nil {
		return fmt.Errorf("load conversation history: %w", err)
	}
	s.persistUserMessage(ctx, conv.ID, message)

	messages := append([]llm.ChatMessage{{Role: "system", Content: assistantSystemPrompt}}, toMessageParams(history)...)
	messages = append(messages, llm.ChatMessage{Role: "user", Content: message})
	tools := s.toolParams()
	var toolCalls []*model.AssistantToolCall

	for turn := 0; turn < assistantMaxToolTurns; turn++ {
		result, err := s.turns.runTurn(ctx, llm.ChatCompletionRequest{
			Model:     s.model,
			MaxTokens: assistantMaxTokens,
			Messages:  messages,
			Tools:     tools,
		}, w)
		if err != nil {
			return err
		}
		if result.FinishReason != "tool_calls" {
			s.persistAssistantTurn(ctx, conv.ID, result.Message.Content, toolCalls)
			return nil
		}

		messages = append(messages, result.Message)
		for _, call := range result.Message.ToolCalls {
			args := json.RawMessage(call.Function.Arguments)
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}
			toolResult, callErr := s.registry.Call(ctx, userID, call.Function.Name, args)
			messages = append(messages, toolResultMessage(call.ID, toolResult, callErr))
			toolCalls = append(toolCalls, toolCallLog(call.Function.Name, args, toolResult, callErr))
		}

		// If this turn had any filler text before its tool call(s) (e.g. "Let me check that:"),
		// separate it from the next turn's text so they don't run together mid-sentence once
		// concatenated client-side — the stream has no other boundary marker between turns.
		if result.Message.Content != "" {
			if err := w.WriteDelta("\n\n"); err != nil {
				return fmt.Errorf("write turn separator: %w", err)
			}
		}
	}
	return ErrAssistantTooManyToolTurns
}
