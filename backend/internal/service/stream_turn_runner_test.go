package service

import (
	"context"
	"errors"
	"testing"

	"github.com/beetrack/backend/internal/llm"
)

// fakeChatStreamer stands in for *llm.OpenRouterClient -- reassembly of streamed chunks into a final
// ChatCompletionStreamResult is internal/llm's job (and tested there), so streamTurnRunner itself only
// needs to prove it forwards the writer's WriteDelta as the onDelta callback and passes results through.
type fakeChatStreamer struct {
	result  *llm.ChatCompletionStreamResult
	err     error
	gotReq  llm.ChatCompletionRequest
	deltaFn func(string) error
}

func (f *fakeChatStreamer) CreateChatCompletionStream(_ context.Context, req llm.ChatCompletionRequest, onDelta func(string) error) (*llm.ChatCompletionStreamResult, error) {
	f.gotReq = req
	f.deltaFn = onDelta
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestStreamTurnRunnerForwardsWriteDeltaAndResult(t *testing.T) {
	streamer := &fakeChatStreamer{result: &llm.ChatCompletionStreamResult{
		Message:      llm.ChatMessage{Role: "assistant", Content: "hello world"},
		FinishReason: "stop",
	}}
	runner := streamTurnRunner{client: streamer}
	w := &fakeStreamWriter{}

	result, err := runner.runTurn(context.Background(), llm.ChatCompletionRequest{Model: "anthropic/claude-haiku-4.5"}, w)
	if err != nil {
		t.Fatalf("runTurn returned error: %v", err)
	}
	if result.Message.Content != "hello world" || result.FinishReason != "stop" {
		t.Errorf("expected the streamer's result passed through unchanged, got %+v", result)
	}

	if err := streamer.deltaFn("chunk"); err != nil {
		t.Fatalf("forwarded onDelta returned error: %v", err)
	}
	if len(w.deltas) != 1 || w.deltas[0] != "chunk" {
		t.Errorf("expected onDelta forwarded to w.WriteDelta, got %+v", w.deltas)
	}
}

func TestStreamTurnRunnerPropagatesStreamerError(t *testing.T) {
	runner := streamTurnRunner{client: &fakeChatStreamer{err: errors.New("stream broke")}}

	_, err := runner.runTurn(context.Background(), llm.ChatCompletionRequest{}, &fakeStreamWriter{})
	if err == nil {
		t.Fatal("expected error when the streamer fails, got nil")
	}
}
