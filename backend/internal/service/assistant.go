package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/beetrack/backend/internal/mcp"
)

const (
	assistantModel        = anthropic.ModelClaudeSonnet5
	assistantMaxTokens    = 2048
	assistantMaxToolTurns = 8
	assistantSystemPrompt = "You are the BeeTrack apiary assistant, helping a beekeeper understand their own hives and browse the public marketplace. Only use the information the tools return; never invent hive data, statuses, or listings."
)

var (
	ErrAssistantEmptyMessages    = errors.New("messages must not be empty")
	ErrAssistantLastMessageUser  = errors.New("the last message must be from the user")
	ErrAssistantTooManyToolTurns = errors.New("assistant exceeded the maximum number of tool-call turns")
)

// The client resends every prior turn each request — nothing is persisted server-side for v1.
type AssistantMessage struct {
	Role    string
	Content string
}

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

// userID always comes from the caller, never from the model's own input.
type AssistantService struct {
	turns    turnRunner
	registry *mcp.Registry
}

// messages is typically &llm.Client.Messages.
func NewAssistantService(messages messageStreamer, registry *mcp.Registry) *AssistantService {
	return &AssistantService{turns: streamTurnRunner{messages: messages}, registry: registry}
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

func toMessageParams(history []AssistantMessage) []anthropic.MessageParam {
	params := make([]anthropic.MessageParam, len(history))
	for i, m := range history {
		block := anthropic.NewTextBlock(m.Content)
		if m.Role == "assistant" {
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

// Text streams to w on every turn, not just the final one.
func (s *AssistantService) Run(ctx context.Context, userID int64, history []AssistantMessage, w AssistantStreamWriter) error {
	if len(history) == 0 {
		return ErrAssistantEmptyMessages
	}
	if history[len(history)-1].Role != "user" {
		return ErrAssistantLastMessageUser
	}

	messages := toMessageParams(history)
	tools := s.toolParams()

	for turn := 0; turn < assistantMaxToolTurns; turn++ {
		msg, err := s.turns.runTurn(ctx, anthropic.MessageNewParams{
			Model:     assistantModel,
			MaxTokens: assistantMaxTokens,
			System:    []anthropic.TextBlockParam{{Text: assistantSystemPrompt}},
			Messages:  messages,
			Tools:     tools,
		}, w)
		if err != nil {
			return err
		}
		if msg.StopReason != anthropic.StopReasonToolUse {
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
		}
		messages = append(messages, anthropic.NewUserMessage(results...))
	}
	return ErrAssistantTooManyToolTurns
}
