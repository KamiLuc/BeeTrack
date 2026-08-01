package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/llm"
	"github.com/beetrack/backend/internal/mcp"
	"github.com/beetrack/backend/internal/model"
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
	responses []*llm.ChatCompletionStreamResult
	errs      []error
	calls     []llm.ChatCompletionRequest
	i         int
}

func (f *fakeTurnRunner) runTurn(_ context.Context, req llm.ChatCompletionRequest, w AssistantStreamWriter) (*llm.ChatCompletionStreamResult, error) {
	f.calls = append(f.calls, req)
	idx := f.i
	f.i++
	if idx < len(f.errs) && f.errs[idx] != nil {
		return nil, f.errs[idx]
	}
	w.WriteDelta("chunk")
	return f.responses[idx], nil
}

func textMessage(text string) *llm.ChatCompletionStreamResult {
	return &llm.ChatCompletionStreamResult{
		Message:      llm.ChatMessage{Role: "assistant", Content: text},
		FinishReason: "stop",
	}
}

func toolUseMessage(id, name string, input map[string]any) *llm.ChatCompletionStreamResult {
	data, _ := json.Marshal(input)
	return &llm.ChatCompletionStreamResult{
		Message: llm.ChatMessage{
			Role:      "assistant",
			ToolCalls: []llm.ToolCall{{ID: id, Type: "function", Function: llm.FunctionCall{Name: name, Arguments: string(data)}}},
		},
		FinishReason: "tool_calls",
	}
}

// fakeAssistantStore is an in-memory AssistantStore, keyed by incrementing IDs, good enough to drive the
// agent loop's persistence calls in tests without a real DB.
type fakeAssistantStore struct {
	conversations map[int64]*model.AssistantConversation
	messages      map[int64][]*model.AssistantMessageLog
	toolCalls     []*model.AssistantToolCall
	nextConvID    int64
	nextMsgID     int64
	createMsgErr  error
	listMsgErr    error
}

func newFakeAssistantStore() *fakeAssistantStore {
	return &fakeAssistantStore{
		conversations: make(map[int64]*model.AssistantConversation),
		messages:      make(map[int64][]*model.AssistantMessageLog),
	}
}

func (s *fakeAssistantStore) CreateConversation(_ context.Context, userID int64) (*model.AssistantConversation, error) {
	s.nextConvID++
	conv := &model.AssistantConversation{ID: s.nextConvID, UserID: userID}
	s.conversations[conv.ID] = conv
	return conv, nil
}

func (s *fakeAssistantStore) GetConversationByID(_ context.Context, id int64) (*model.AssistantConversation, error) {
	return s.conversations[id], nil
}

func (s *fakeAssistantStore) ListConversations(_ context.Context, userID int64) ([]*model.AssistantConversationSummary, error) {
	var summaries []*model.AssistantConversationSummary
	for _, conv := range s.conversations {
		if conv.UserID != userID {
			continue
		}
		var preview string
		for _, m := range s.messages[conv.ID] {
			if m.Role == model.AssistantMessageRoleUser {
				preview = m.Content
				break
			}
		}
		summaries = append(summaries, &model.AssistantConversationSummary{
			ID: conv.ID, CreatedAt: conv.CreatedAt, Preview: preview, MessageCount: int64(len(s.messages[conv.ID])),
		})
	}
	return summaries, nil
}

func (s *fakeAssistantStore) ListMessageLogs(_ context.Context, conversationID int64) ([]*model.AssistantMessageLog, error) {
	if s.listMsgErr != nil {
		return nil, s.listMsgErr
	}
	return s.messages[conversationID], nil
}

func (s *fakeAssistantStore) CountMessageLogs(_ context.Context, conversationID int64) (int64, error) {
	return int64(len(s.messages[conversationID])), nil
}

func (s *fakeAssistantStore) DeleteConversation(_ context.Context, id int64) error {
	delete(s.conversations, id)
	delete(s.messages, id)
	return nil
}

func (s *fakeAssistantStore) CreateMessageLog(_ context.Context, log *model.AssistantMessageLog) error {
	if s.createMsgErr != nil {
		return s.createMsgErr
	}
	s.nextMsgID++
	log.ID = s.nextMsgID
	s.messages[log.ConversationID] = append(s.messages[log.ConversationID], log)
	return nil
}

func (s *fakeAssistantStore) CreateToolCall(_ context.Context, call *model.AssistantToolCall) error {
	s.toolCalls = append(s.toolCalls, call)
	return nil
}

func newTestAssistant(runner turnRunner) (*AssistantService, *mcp.Registry, *fakeAssistantStore) {
	registry := mcp.NewRegistry()
	store := newFakeAssistantStore()
	return &AssistantService{turns: runner, registry: registry, repo: store}, registry, store
}

func testConversation(t *testing.T, store *fakeAssistantStore, userID int64) *model.AssistantConversation {
	t.Helper()
	conv, err := store.CreateConversation(context.Background(), userID)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	return conv
}

func TestAssistantRunRejectsEmptyMessage(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	conv := testConversation(t, store, 1)
	if err := svc.Run(context.Background(), 1, conv, "", &fakeStreamWriter{}); !errors.Is(err, ErrAssistantEmptyMessage) {
		t.Fatalf("expected ErrAssistantEmptyMessage, got %v", err)
	}
}

func TestAssistantRunEndsOnNonToolUseStopReason(t *testing.T) {
	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{textMessage("hello beekeeper")}}
	svc, _, store := newTestAssistant(runner)
	conv := testConversation(t, store, 1)
	w := &fakeStreamWriter{}

	err := svc.Run(context.Background(), 1, conv, "hi", w)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("expected exactly 1 turn, got %d", len(runner.calls))
	}
	if len(w.deltas) != 1 || w.deltas[0] != "chunk" {
		t.Errorf("expected the writer passed through to the turn runner, got %+v", w.deltas)
	}

	msgs := store.messages[conv.ID]
	if len(msgs) != 2 || msgs[0].Role != model.AssistantMessageRoleUser || msgs[1].Role != model.AssistantMessageRoleAssistant {
		t.Fatalf("expected [user, assistant] persisted, got %+v", msgs)
	}
	if msgs[1].Content != "hello beekeeper" {
		t.Errorf("expected persisted assistant content %q, got %q", "hello beekeeper", msgs[1].Content)
	}
}

func TestAssistantRunLoadsPriorHistoryIntoTheTurn(t *testing.T) {
	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{textMessage("still 3 hives")}}
	svc, _, store := newTestAssistant(runner)
	conv := testConversation(t, store, 1)
	store.messages[conv.ID] = []*model.AssistantMessageLog{
		{ID: 1, ConversationID: conv.ID, Role: model.AssistantMessageRoleUser, Content: "how many hives?"},
		{ID: 2, ConversationID: conv.ID, Role: model.AssistantMessageRoleAssistant, Content: "3 hives"},
	}

	if err := svc.Run(context.Background(), 1, conv, "still correct?", &fakeStreamWriter{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	sent := runner.calls[0].Messages
	if len(sent) != 4 {
		t.Fatalf("expected the system prompt + 2 history messages + the new one sent to the model, got %d", len(sent))
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

	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{
		toolUseMessage("toolu_1", "list_hives", map[string]any{"apiary_id": 5}),
		textMessage("you have 3 hives"),
	}}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 42)

	err := svc.Run(context.Background(), 42, conv, "how many hives?", &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 turns (tool call + final answer), got %d", len(runner.calls))
	}
	if gotUserID != 42 || gotName != "list_hives" || string(gotInput) != `{"apiary_id":5}` {
		t.Errorf("tool called with unexpected args: userID=%d name=%s input=%s", gotUserID, gotName, gotInput)
	}

	if len(store.toolCalls) != 1 || store.toolCalls[0].ToolName != "list_hives" || store.toolCalls[0].IsError {
		t.Fatalf("expected 1 persisted successful tool call, got %+v", store.toolCalls)
	}
	if store.toolCalls[0].MessageID == 0 {
		t.Errorf("expected the tool call linked to the persisted assistant message, got MessageID=0")
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

	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{
		toolUseMessage("toolu_1", "broken_tool", map[string]any{}),
		textMessage("sorry, that failed"),
	}}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 1)

	err := svc.Run(context.Background(), 1, conv, "hi", &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected the loop to continue past a tool error, got %d turns", len(runner.calls))
	}
	if len(store.toolCalls) != 1 || !store.toolCalls[0].IsError {
		t.Fatalf("expected 1 persisted tool call marked as an error, got %+v", store.toolCalls)
	}
}

func multiToolUseMessage(calls ...[2]string) *llm.ChatCompletionStreamResult {
	toolCalls := make([]llm.ToolCall, len(calls))
	for i, c := range calls {
		toolCalls[i] = llm.ToolCall{ID: c[0], Type: "function", Function: llm.FunctionCall{Name: c[1], Arguments: "{}"}}
	}
	return &llm.ChatCompletionStreamResult{
		Message:      llm.ChatMessage{Role: "assistant", ToolCalls: toolCalls},
		FinishReason: "tool_calls",
	}
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

	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{
		multiToolUseMessage([2]string{"toolu_1", "list_hives"}, [2]string{"toolu_2", "get_dashboard_summary"}),
		textMessage("here you go"),
	}}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 1)

	err := svc.Run(context.Background(), 1, conv, "hi", &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(calledNames) != 2 || calledNames[0] != "list_hives" || calledNames[1] != "get_dashboard_summary" {
		t.Errorf("expected both tools called in order, got %+v", calledNames)
	}

	sentMessages := runner.calls[1].Messages
	toolResultMsgs := sentMessages[len(sentMessages)-2:]
	for _, m := range toolResultMsgs {
		if m.Role != "tool" {
			t.Errorf("expected a tool-role message per call, got %+v", m)
		}
	}
	if toolResultMsgs[0].ToolCallID != "toolu_1" || toolResultMsgs[1].ToolCallID != "toolu_2" {
		t.Errorf("expected one tool message per tool_call_id in order, got %+v", toolResultMsgs)
	}
	if len(store.toolCalls) != 2 {
		t.Fatalf("expected 2 persisted tool calls, got %d", len(store.toolCalls))
	}
}

func TestAssistantRunUnregisteredToolNameRelayedAsToolResultAndLoopContinues(t *testing.T) {
	registry := mcp.NewRegistry()
	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{
		toolUseMessage("toolu_1", "does_not_exist", map[string]any{}),
		textMessage("sorry, I couldn't do that"),
	}}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 1)

	err := svc.Run(context.Background(), 1, conv, "hi", &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected the loop to continue past an unregistered tool name, got %d turns", len(runner.calls))
	}
	toolResultMsg := runner.calls[1].Messages[len(runner.calls[1].Messages)-1]
	if toolResultMsg.Role != "tool" || toolResultMsg.ToolCallID != "toolu_1" || !strings.Contains(toolResultMsg.Content, "unknown tool") {
		t.Errorf("expected a tool message carrying the registry error, got %+v", toolResultMsg)
	}
}

func TestAssistantRunTreatsEmptyToolCallArgumentsAsEmptyObject(t *testing.T) {
	var gotInput json.RawMessage
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name: "list_hives",
		Handler: func(_ context.Context, _ int64, input json.RawMessage) (any, error) {
			gotInput = input
			return "ok", nil
		},
	})

	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{
		{
			Message: llm.ChatMessage{
				Role:      "assistant",
				ToolCalls: []llm.ToolCall{{ID: "toolu_1", Type: "function", Function: llm.FunctionCall{Name: "list_hives", Arguments: ""}}},
			},
			FinishReason: "tool_calls",
		},
		textMessage("you have 3 hives"),
	}}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 1)

	if err := svc.Run(context.Background(), 1, conv, "how many hives?", &fakeStreamWriter{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if string(gotInput) != "{}" {
		t.Errorf("expected empty tool-call arguments normalized to {}, got %q", gotInput)
	}
	if len(store.toolCalls) != 1 || string(store.toolCalls[0].Input) != "{}" {
		t.Errorf("expected the persisted tool call input normalized to {}, got %+v", store.toolCalls)
	}
}

func TestAssistantRunExceedingMaxToolTurnsReturnsError(t *testing.T) {
	responses := make([]*llm.ChatCompletionStreamResult, assistantMaxToolTurns)
	for i := range responses {
		responses[i] = toolUseMessage("toolu", "list_hives", map[string]any{})
	}
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name:    "list_hives",
		Handler: func(_ context.Context, _ int64, _ json.RawMessage) (any, error) { return nil, nil },
	})
	runner := &fakeTurnRunner{responses: responses}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 1)

	err := svc.Run(context.Background(), 1, conv, "hi", &fakeStreamWriter{})
	if !errors.Is(err, ErrAssistantTooManyToolTurns) {
		t.Fatalf("expected ErrAssistantTooManyToolTurns, got %v", err)
	}
	if len(runner.calls) != assistantMaxToolTurns {
		t.Errorf("expected exactly %d turns, got %d", assistantMaxToolTurns, len(runner.calls))
	}
}

func TestAssistantRunPropagatesTurnRunnerError(t *testing.T) {
	runner := &fakeTurnRunner{errs: []error{errors.New("stream broke")}, responses: []*llm.ChatCompletionStreamResult{nil}}
	svc, _, store := newTestAssistant(runner)
	conv := testConversation(t, store, 1)

	err := svc.Run(context.Background(), 1, conv, "hi", &fakeStreamWriter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAssistantRunSucceedsDespitePersistFailure(t *testing.T) {
	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{textMessage("hello beekeeper")}}
	svc, _, store := newTestAssistant(runner)
	conv := testConversation(t, store, 1)
	store.createMsgErr = errors.New("db down")

	err := svc.Run(context.Background(), 1, conv, "hi", &fakeStreamWriter{})
	if err != nil {
		t.Fatalf("expected a logging failure to not abort the response, got %v", err)
	}
	if len(store.messages[conv.ID]) != 0 {
		t.Errorf("expected no messages persisted when CreateMessageLog always fails, got %+v", store.messages[conv.ID])
	}
}

func TestAssistantRunPropagatesHistoryLoadError(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	conv := testConversation(t, store, 1)
	store.listMsgErr = errors.New("db down")

	err := svc.Run(context.Background(), 1, conv, "hi", &fakeStreamWriter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveConversationCreatesNewWhenNilGiven(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	conv, err := svc.ResolveConversation(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("ResolveConversation returned error: %v", err)
	}
	if conv.UserID != 7 {
		t.Errorf("expected new conversation owned by userID 7, got %d", conv.UserID)
	}
	if _, ok := store.conversations[conv.ID]; !ok {
		t.Errorf("expected the new conversation to be persisted")
	}
}

func TestResolveConversationLoadsExistingOwnedConversation(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	existing := testConversation(t, store, 7)

	conv, err := svc.ResolveConversation(context.Background(), 7, &existing.ID)
	if err != nil {
		t.Fatalf("ResolveConversation returned error: %v", err)
	}
	if conv.ID != existing.ID {
		t.Errorf("expected the existing conversation returned, got %+v", conv)
	}
}

func TestResolveConversationRejectsUnknownID(t *testing.T) {
	svc, _, _ := newTestAssistant(&fakeTurnRunner{})
	unknown := int64(999)
	_, err := svc.ResolveConversation(context.Background(), 7, &unknown)
	if !errors.Is(err, ErrAssistantConversationNotFound) {
		t.Fatalf("expected ErrAssistantConversationNotFound, got %v", err)
	}
}

func TestResolveConversationRejectsConversationOwnedByAnotherUser(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	other := testConversation(t, store, 99)

	_, err := svc.ResolveConversation(context.Background(), 7, &other.ID)
	if !errors.Is(err, ErrAssistantConversationNotFound) {
		t.Fatalf("expected ErrAssistantConversationNotFound, got %v", err)
	}
}

func TestCheckMessageLimitAllowsUnderLimit(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	conv := testConversation(t, store, 1)
	store.messages[conv.ID] = make([]*model.AssistantMessageLog, assistantMaxMessagesPerConversation-1)

	if err := svc.CheckMessageLimit(context.Background(), conv.ID); err != nil {
		t.Fatalf("expected no error under the limit, got %v", err)
	}
}

func TestCheckMessageLimitRejectsAtLimit(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	conv := testConversation(t, store, 1)
	store.messages[conv.ID] = make([]*model.AssistantMessageLog, assistantMaxMessagesPerConversation)

	err := svc.CheckMessageLimit(context.Background(), conv.ID)
	if !errors.Is(err, ErrAssistantConversationLimitReached) {
		t.Fatalf("expected ErrAssistantConversationLimitReached, got %v", err)
	}
}

func TestListConversationsReturnsOnlyCallersConversations(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	mine := testConversation(t, store, 7)
	testConversation(t, store, 99)
	store.messages[mine.ID] = []*model.AssistantMessageLog{
		{ID: 1, ConversationID: mine.ID, Role: model.AssistantMessageRoleUser, Content: "hi"},
	}

	summaries, err := svc.ListConversations(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if len(summaries) != 1 || summaries[0].ID != mine.ID || summaries[0].Preview != "hi" {
		t.Fatalf("expected only the caller's conversation with its preview, got %+v", summaries)
	}
}

func TestConversationMessagesReturnsHistoryForOwnedConversation(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	conv := testConversation(t, store, 7)
	store.messages[conv.ID] = []*model.AssistantMessageLog{
		{ID: 1, ConversationID: conv.ID, Role: model.AssistantMessageRoleUser, Content: "hi"},
		{ID: 2, ConversationID: conv.ID, Role: model.AssistantMessageRoleAssistant, Content: "hello"},
	}

	logs, err := svc.ConversationMessages(context.Background(), 7, conv.ID)
	if err != nil {
		t.Fatalf("ConversationMessages returned error: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(logs))
	}
}

func TestConversationMessagesRejectsConversationOwnedByAnotherUser(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	other := testConversation(t, store, 99)

	_, err := svc.ConversationMessages(context.Background(), 7, other.ID)
	if !errors.Is(err, ErrAssistantConversationNotFound) {
		t.Fatalf("expected ErrAssistantConversationNotFound, got %v", err)
	}
}

func TestConversationMessagesRejectsUnknownConversation(t *testing.T) {
	svc, _, _ := newTestAssistant(&fakeTurnRunner{})

	_, err := svc.ConversationMessages(context.Background(), 7, 999)
	if !errors.Is(err, ErrAssistantConversationNotFound) {
		t.Fatalf("expected ErrAssistantConversationNotFound, got %v", err)
	}
}

func TestDeleteConversationRemovesOwnedConversation(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	conv := testConversation(t, store, 7)

	if err := svc.DeleteConversation(context.Background(), 7, conv.ID); err != nil {
		t.Fatalf("DeleteConversation returned error: %v", err)
	}
	if _, ok := store.conversations[conv.ID]; ok {
		t.Errorf("expected the conversation to be deleted")
	}
}

func TestDeleteConversationRejectsConversationOwnedByAnotherUser(t *testing.T) {
	svc, _, store := newTestAssistant(&fakeTurnRunner{})
	other := testConversation(t, store, 99)

	err := svc.DeleteConversation(context.Background(), 7, other.ID)
	if !errors.Is(err, ErrAssistantConversationNotFound) {
		t.Fatalf("expected ErrAssistantConversationNotFound, got %v", err)
	}
	if _, ok := store.conversations[other.ID]; !ok {
		t.Errorf("expected the other user's conversation to remain")
	}
}

func TestDeleteConversationRejectsUnknownConversation(t *testing.T) {
	svc, _, _ := newTestAssistant(&fakeTurnRunner{})

	err := svc.DeleteConversation(context.Background(), 7, 999)
	if !errors.Is(err, ErrAssistantConversationNotFound) {
		t.Fatalf("expected ErrAssistantConversationNotFound, got %v", err)
	}
}

func TestToolParamsConvertsRegistryToolsToOpenAISchema(t *testing.T) {
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
	if len(params) != 1 || params[0].Type != "function" {
		t.Fatalf("expected 1 converted function tool, got %+v", params)
	}
	tool := params[0].Function
	if tool.Name != "list_hives" || tool.Description != "lists hives" {
		t.Errorf("unexpected tool conversion: %+v", tool)
	}
	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(tool.Parameters, &schema); err != nil {
		t.Fatalf("unmarshal parameters schema: %v", err)
	}
	if len(schema.Required) != 1 || schema.Required[0] != "apiary_id" {
		t.Errorf("expected required apiary_id, got %+v", schema.Required)
	}
}

func TestToMessageParamsMapsRoles(t *testing.T) {
	history := []*model.AssistantMessageLog{
		{Role: model.AssistantMessageRoleUser, Content: "hi"},
		{Role: model.AssistantMessageRoleAssistant, Content: "hello"},
	}
	params := toMessageParams(history)
	if len(params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(params))
	}
	if params[0].Role != "user" {
		t.Errorf("expected first message role user, got %v", params[0].Role)
	}
	if params[1].Role != "assistant" {
		t.Errorf("expected second message role assistant, got %v", params[1].Role)
	}
}

func TestToolResultMessageCarriesErrorTextInContent(t *testing.T) {
	okMsg := toolResultMessage("toolu_1", map[string]string{"a": "b"}, nil)
	if okMsg.Role != "tool" || okMsg.ToolCallID != "toolu_1" || strings.Contains(okMsg.Content, "nope") {
		t.Errorf("expected a non-error tool message, got %+v", okMsg)
	}

	errMsg := toolResultMessage("toolu_2", nil, errors.New("nope"))
	if errMsg.Role != "tool" || errMsg.Content != "nope" {
		t.Errorf("expected the tool message content to carry the error text, got %+v", errMsg)
	}
}

// toolUseMessageWithText simulates a turn where the model writes some filler text (e.g. "Let me check
// that:") before calling a tool, which real models can do.
func toolUseMessageWithText(text, id, name string) *llm.ChatCompletionStreamResult {
	return &llm.ChatCompletionStreamResult{
		Message: llm.ChatMessage{
			Role:      "assistant",
			Content:   text,
			ToolCalls: []llm.ToolCall{{ID: id, Type: "function", Function: llm.FunctionCall{Name: name, Arguments: "{}"}}},
		},
		FinishReason: "tool_calls",
	}
}

func TestAssistantRunWritesSeparatorAfterATurnWithFillerTextBeforeToolUse(t *testing.T) {
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name:    "list_hives",
		Handler: func(_ context.Context, _ int64, _ json.RawMessage) (any, error) { return "ok", nil },
	})

	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{
		toolUseMessageWithText("Let me check that:", "toolu_1", "list_hives"),
		textMessage("you have 3 hives"),
	}}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 1)

	w := &fakeStreamWriter{}
	if err := svc.Run(context.Background(), 1, conv, "how many hives?", w); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	found := false
	for _, d := range w.deltas {
		if d == "\n\n" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a \\n\\n separator delta between turns, got deltas: %+v", w.deltas)
	}
}

func TestAssistantRunSkipsSeparatorWhenToolUseTurnHasNoText(t *testing.T) {
	registry := mcp.NewRegistry()
	registry.Register(mcp.Tool{
		Name:    "list_hives",
		Handler: func(_ context.Context, _ int64, _ json.RawMessage) (any, error) { return "ok", nil },
	})

	runner := &fakeTurnRunner{responses: []*llm.ChatCompletionStreamResult{
		toolUseMessage("toolu_1", "list_hives", map[string]any{}),
		textMessage("you have 3 hives"),
	}}
	store := newFakeAssistantStore()
	svc := &AssistantService{turns: runner, registry: registry, repo: store}
	conv := testConversation(t, store, 1)

	w := &fakeStreamWriter{}
	if err := svc.Run(context.Background(), 1, conv, "how many hives?", w); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	for _, d := range w.deltas {
		if d == "\n\n" {
			t.Errorf("expected no separator when the tool-use turn had no filler text, got deltas: %+v", w.deltas)
		}
	}
}
