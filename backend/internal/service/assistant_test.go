package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/beetrack/backend/internal/mcp"
)

type fakeStreamWriter struct {
	deltas []string
}

func (w *fakeStreamWriter) WriteDelta(text string) error {
	w.deltas = append(w.deltas, text)
	return nil
}

// fakeTurnRunner replays canned responses in order, one per call to runTurn.
type fakeTurnRunner struct {
	responses []*anthropic.Message
	errs      []error
	calls     []anthropic.MessageNewParams
	i         int
}

func (f *fakeTurnRunner) runTurn(_ context.Context, params anthropic.MessageNewParams, w AssistantStreamWriter) (*anthropic.Message, error) {
	f.calls = append(f.calls, params)
	idx := f.i
	f.i++
	if idx < len(f.errs) && f.errs[idx] != nil {
		return nil, f.errs[idx]
	}
	w.WriteDelta("chunk")
	return f.responses[idx], nil
}

func textMessage(text string) *anthropic.Message {
	return &anthropic.Message{
		Content:    []anthropic.ContentBlockUnion{{Type: "text", Text: text}},
		StopReason: anthropic.StopReasonEndTurn,
	}
}

func toolUseMessage(id, name string, input map[string]any) *anthropic.Message {
	data, _ := json.Marshal(input)
	return &anthropic.Message{
		Content:    []anthropic.ContentBlockUnion{{Type: "tool_use", ID: id, Name: name, Input: data}},
		StopReason: anthropic.StopReasonToolUse,
	}
}

func newTestAssistant(runner turnRunner) (*AssistantService, *mcp.Registry) {
	registry := mcp.NewRegistry()
	return &AssistantService{turns: runner, registry: registry}, registry
}

func TestAssistantRunRejectsEmptyHistory(t *testing.T) {
	svc, _ := newTestAssistant(&fakeTurnRunner{})
	if err := svc.Run(context.Background(), 1, nil, &fakeStreamWriter{}); !errors.Is(err, ErrAssistantEmptyMessages) {
		t.Fatalf("expected ErrAssistantEmptyMessages, got %v", err)
	}
}

func TestAssistantRunRejectsNonUserLastMessage(t *testing.T) {
	svc, _ := newTestAssistant(&fakeTurnRunner{})
	history := []AssistantMessage{{Role: "user", Content: "hi"}, {Role: "assistant", Content: "hello"}}
	if err := svc.Run(context.Background(), 1, history, &fakeStreamWriter{}); !errors.Is(err, ErrAssistantLastMessageUser) {
		t.Fatalf("expected ErrAssistantLastMessageUser, got %v", err)
	}
}

func TestAssistantRunEndsOnNonToolUseStopReason(t *testing.T) {
	runner := &fakeTurnRunner{responses: []*anthropic.Message{textMessage("hello beekeeper")}}
	svc, _ := newTestAssistant(runner)
	w := &fakeStreamWriter{}

	err := svc.Run(context.Background(), 1, []AssistantMessage{{Role: "user", Content: "hi"}}, w)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("expected exactly 1 turn, got %d", len(runner.calls))
	}
	if len(w.deltas) != 1 || w.deltas[0] != "chunk" {
		t.Errorf("expected the writer passed through to the turn runner, got %+v", w.deltas)
	}
}

func TestAssistantRunCallsToolThenReturnsFinalAnswer(t *testing.T) {
	var gotUserID int64
	var gotName string
	var gotInput json.RawMessage
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name: "list_hives",
		Handler: func(_ context.Context, userID int64, input json.RawMessage) (any, error) {
			gotUserID, gotName, gotInput = userID, "list_hives", input
			return map[string]string{"result": "ok"}, nil
		},
	})

	runner := &fakeTurnRunner{responses: []*anthropic.Message{
		toolUseMessage("toolu_1", "list_hives", map[string]any{"apiary_id": 5}),
		textMessage("you have 3 hives"),
	}}
	svc := &AssistantService{turns: runner, registry: registry}

	err := svc.Run(context.Background(), 42, []AssistantMessage{{Role: "user", Content: "how many hives?"}}, &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 turns (tool call + final answer), got %d", len(runner.calls))
	}
	if gotUserID != 42 || gotName != "list_hives" || string(gotInput) != `{"apiary_id":5}` {
		t.Errorf("tool called with unexpected args: userID=%d name=%s input=%s", gotUserID, gotName, gotInput)
	}
}

func TestAssistantRunToolErrorIsRelayedAsToolResultAndLoopContinues(t *testing.T) {
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name: "broken_tool",
		Handler: func(_ context.Context, _ int64, _ json.RawMessage) (any, error) {
			return nil, errors.New("boom")
		},
	})

	runner := &fakeTurnRunner{responses: []*anthropic.Message{
		toolUseMessage("toolu_1", "broken_tool", map[string]any{}),
		textMessage("sorry, that failed"),
	}}
	svc := &AssistantService{turns: runner, registry: registry}

	err := svc.Run(context.Background(), 1, []AssistantMessage{{Role: "user", Content: "hi"}}, &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected the loop to continue past a tool error, got %d turns", len(runner.calls))
	}
}

func multiToolUseMessage(calls ...[2]string) *anthropic.Message {
	content := make([]anthropic.ContentBlockUnion, len(calls))
	for i, c := range calls {
		content[i] = anthropic.ContentBlockUnion{Type: "tool_use", ID: c[0], Name: c[1], Input: json.RawMessage(`{}`)}
	}
	return &anthropic.Message{Content: content, StopReason: anthropic.StopReasonToolUse}
}

func TestAssistantRunHandlesMultipleToolUseBlocksInOneTurn(t *testing.T) {
	var calledNames []string
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name: "list_hives",
		Handler: func(_ context.Context, _ int64, _ json.RawMessage) (any, error) {
			calledNames = append(calledNames, "list_hives")
			return "hives", nil
		},
	})
	registry.Register(mcp.Tool{
		Name: "get_dashboard_summary",
		Handler: func(_ context.Context, _ int64, _ json.RawMessage) (any, error) {
			calledNames = append(calledNames, "get_dashboard_summary")
			return "summary", nil
		},
	})

	runner := &fakeTurnRunner{responses: []*anthropic.Message{
		multiToolUseMessage([2]string{"toolu_1", "list_hives"}, [2]string{"toolu_2", "get_dashboard_summary"}),
		textMessage("here you go"),
	}}
	svc := &AssistantService{turns: runner, registry: registry}

	err := svc.Run(context.Background(), 1, []AssistantMessage{{Role: "user", Content: "hi"}}, &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(calledNames) != 2 || calledNames[0] != "list_hives" || calledNames[1] != "get_dashboard_summary" {
		t.Errorf("expected both tools called in order, got %+v", calledNames)
	}

	toolResultsMsg := runner.calls[1].Messages[len(runner.calls[1].Messages)-1]
	if len(toolResultsMsg.Content) != 2 {
		t.Fatalf("expected 2 tool_result blocks fed back in one user message, got %d", len(toolResultsMsg.Content))
	}
}

func TestAssistantRunUnregisteredToolNameRelayedAsToolResultAndLoopContinues(t *testing.T) {
	registry := mcp.NewRegistry()
	runner := &fakeTurnRunner{responses: []*anthropic.Message{
		toolUseMessage("toolu_1", "does_not_exist", map[string]any{}),
		textMessage("sorry, I couldn't do that"),
	}}
	svc := &AssistantService{turns: runner, registry: registry}

	err := svc.Run(context.Background(), 1, []AssistantMessage{{Role: "user", Content: "hi"}}, &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected the loop to continue past an unregistered tool name, got %d turns", len(runner.calls))
	}
	toolResultsMsg := runner.calls[1].Messages[len(runner.calls[1].Messages)-1]
	if len(toolResultsMsg.Content) != 1 || !toolResultsMsg.Content[0].OfToolResult.IsError.Value {
		t.Errorf("expected an is_error tool result for the unregistered tool, got %+v", toolResultsMsg.Content)
	}
}

func TestAssistantRunExceedingMaxToolTurnsReturnsError(t *testing.T) {
	responses := make([]*anthropic.Message, assistantMaxToolTurns)
	for i := range responses {
		responses[i] = toolUseMessage("toolu", "list_hives", map[string]any{})
	}
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name:    "list_hives",
		Handler: func(_ context.Context, _ int64, _ json.RawMessage) (any, error) { return nil, nil },
	})
	runner := &fakeTurnRunner{responses: responses}
	svc := &AssistantService{turns: runner, registry: registry}

	err := svc.Run(context.Background(), 1, []AssistantMessage{{Role: "user", Content: "hi"}}, &fakeStreamWriter{})
	if !errors.Is(err, ErrAssistantTooManyToolTurns) {
		t.Fatalf("expected ErrAssistantTooManyToolTurns, got %v", err)
	}
	if len(runner.calls) != assistantMaxToolTurns {
		t.Errorf("expected exactly %d turns, got %d", assistantMaxToolTurns, len(runner.calls))
	}
}

func TestAssistantRunPropagatesTurnRunnerError(t *testing.T) {
	runner := &fakeTurnRunner{errs: []error{errors.New("stream broke")}, responses: []*anthropic.Message{nil}}
	svc, _ := newTestAssistant(runner)

	err := svc.Run(context.Background(), 1, []AssistantMessage{{Role: "user", Content: "hi"}}, &fakeStreamWriter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestToolParamsConvertsRegistryToolsToAnthropicSchema(t *testing.T) {
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name:        "list_hives",
		Description: "lists hives",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{"apiary_id": map[string]any{"type": "integer"}},
			Required:   []string{"apiary_id"},
		},
	})
	svc := &AssistantService{registry: registry}

	params := svc.toolParams()
	if len(params) != 1 || params[0].OfTool == nil {
		t.Fatalf("expected 1 converted tool, got %+v", params)
	}
	tool := params[0].OfTool
	if tool.Name != "list_hives" || tool.Description.Value != "lists hives" {
		t.Errorf("unexpected tool conversion: %+v", tool)
	}
	if tool.InputSchema.Required[0] != "apiary_id" {
		t.Errorf("expected required apiary_id, got %+v", tool.InputSchema.Required)
	}
}

func TestToMessageParamsMapsRoles(t *testing.T) {
	history := []AssistantMessage{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
	}
	params := toMessageParams(history)
	if len(params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(params))
	}
	if params[0].Role != anthropic.MessageParamRoleUser {
		t.Errorf("expected first message role user, got %v", params[0].Role)
	}
	if params[1].Role != anthropic.MessageParamRoleAssistant {
		t.Errorf("expected second message role assistant, got %v", params[1].Role)
	}
}

func TestToolResultParamMarksErrorsAsIsError(t *testing.T) {
	okBlock := toolResultParam("toolu_1", map[string]string{"a": "b"}, nil)
	if okBlock.OfToolResult == nil || okBlock.OfToolResult.IsError.Value {
		t.Errorf("expected a non-error tool result, got %+v", okBlock.OfToolResult)
	}

	errBlock := toolResultParam("toolu_2", nil, errors.New("nope"))
	if errBlock.OfToolResult == nil || !errBlock.OfToolResult.IsError.Value {
		t.Errorf("expected an error tool result, got %+v", errBlock.OfToolResult)
	}
}
