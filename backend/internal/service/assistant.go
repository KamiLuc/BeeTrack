package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/beetrack/backend/internal/mcp"
	"github.com/beetrack/backend/internal/model"
	"gorm.io/datatypes"
)

const (
	defaultAssistantModel = anthropic.ModelClaudeHaiku4_5
	assistantMaxTokens    = 2048
	assistantMaxToolTurns = 8
	// assistantMaxMessagesPerConversation keeps conversations short enough to stay within a single Claude
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

// Satisfied by &llm.Client.Messages; kept as its own interface so the agent loop is testable without a real
// Anthropic connection.
type messageStreamer interface {
	NewStreaming(ctx context.Context, params anthropic.MessageNewParams, opts ...option.RequestOption) *ssestream.Stream[anthropic.MessageStreamEventUnion]
}

// Kept separate from AssistantService's tool-orchestration loop so that loop is testable with canned
// *anthropic.Message responses, without also having to fake the SDK's own SSE-accumulation internals.
type turnRunner interface {
	runTurn(ctx context.Context, params anthropic.MessageNewParams, w AssistantStreamWriter) (*anthropic.Message, error)
}

type streamTurnRunner struct {
	messages messageStreamer
}

func (r streamTurnRunner) runTurn(ctx context.Context, params anthropic.MessageNewParams, w AssistantStreamWriter) (*anthropic.Message, error) {
	stream := r.messages.NewStreaming(ctx, params)
	var msg anthropic.Message
	for stream.Next() {
		event := stream.Current()
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
			if err := w.WriteDelta(event.Delta.Text); err != nil {
				return nil, fmt.Errorf("write delta: %w", err)
			}
		}
		if err := msg.Accumulate(event); err != nil {
			return nil, fmt.Errorf("accumulate stream event: %w", err)
		}
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("stream claude response: %w", err)
	}
	return &msg, nil
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
	model    anthropic.Model
}

// messages is typically &llm.Client.Messages. model selects the Claude model for the agent loop — pass ""
// to use defaultAssistantModel (see config.AnthropicModel for the env var that controls this in main.go).
func NewAssistantService(messages messageStreamer, registry *mcp.Registry, repo AssistantStore, model string) *AssistantService {
	m := anthropic.Model(model)
	if m == "" {
		m = defaultAssistantModel
	}
	return &AssistantService{turns: streamTurnRunner{messages: messages}, registry: registry, repo: repo, model: m}
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

func (s *AssistantService) toolParams() []anthropic.ToolUnionParam {
	tools := s.registry.All()
	params := make([]anthropic.ToolUnionParam, len(tools))
	for i, t := range tools {
		params[i] = anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{
			Name:        t.Name,
			Description: param.NewOpt(t.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: t.InputSchema.Properties,
				Required:   t.InputSchema.Required,
			},
		}}
	}
	return params
}

func toMessageParams(history []*model.AssistantMessageLog) []anthropic.MessageParam {
	params := make([]anthropic.MessageParam, len(history))
	for i, m := range history {
		block := anthropic.NewTextBlock(m.Content)
		if m.Role == model.AssistantMessageRoleAssistant {
			params[i] = anthropic.NewAssistantMessage(block)
		} else {
			params[i] = anthropic.NewUserMessage(block)
		}
	}
	return params
}

func toolResultParam(toolUseID string, result any, callErr error) anthropic.ContentBlockParamUnion {
	if callErr != nil {
		return anthropic.NewToolResultBlock(toolUseID, callErr.Error(), true)
	}
	data, err := json.Marshal(result)
	if err != nil {
		return anthropic.NewToolResultBlock(toolUseID, err.Error(), true)
	}
	return anthropic.NewToolResultBlock(toolUseID, string(data), false)
}

// toolCallLog builds the assistant_tool_calls row for one MCP call — Result holds the marshaled tool
// result, or the error message when callErr is set, mirroring toolResultParam's own success/error shape.
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

func finalText(msg *anthropic.Message) string {
	var sb strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String()
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

	messages := append(toMessageParams(history), anthropic.NewUserMessage(anthropic.NewTextBlock(message)))
	tools := s.toolParams()
	var toolCalls []*model.AssistantToolCall

	for turn := 0; turn < assistantMaxToolTurns; turn++ {
		msg, err := s.turns.runTurn(ctx, anthropic.MessageNewParams{
			Model:     s.model,
			MaxTokens: assistantMaxTokens,
			System:    []anthropic.TextBlockParam{{Text: assistantSystemPrompt}},
			Messages:  messages,
			Tools:     tools,
		}, w)
		if err != nil {
			return err
		}
		if msg.StopReason != anthropic.StopReasonToolUse {
			s.persistAssistantTurn(ctx, conv.ID, finalText(msg), toolCalls)
			return nil
		}

		messages = append(messages, msg.ToParam())
		var results []anthropic.ContentBlockParamUnion
		for _, block := range msg.Content {
			if block.Type != "tool_use" {
				continue
			}
			result, callErr := s.registry.Call(ctx, userID, block.Name, block.Input)
			results = append(results, toolResultParam(block.ID, result, callErr))
			toolCalls = append(toolCalls, toolCallLog(block.Name, block.Input, result, callErr))
		}
		messages = append(messages, anthropic.NewUserMessage(results...))

		// If this turn had any filler text before its tool call(s) (e.g. "Let me check that:"),
		// separate it from the next turn's text so they don't run together mid-sentence once
		// concatenated client-side — the stream has no other boundary marker between turns.
		if finalText(msg) != "" {
			if err := w.WriteDelta("\n\n"); err != nil {
				return fmt.Errorf("write turn separator: %w", err)
			}
		}
	}
	return ErrAssistantTooManyToolTurns
}
