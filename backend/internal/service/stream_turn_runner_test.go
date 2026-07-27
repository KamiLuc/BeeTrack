package service

import (
	"context"
	"errors"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

// erroringStreamWriter simulates a client disconnecting mid-stream: every
// WriteDelta call fails.
type erroringStreamWriter struct{}

func (erroringStreamWriter) WriteDelta(string) error { return errors.New("client disconnected") }

// fakeSSEDecoder replays a canned sequence of raw SSE events for ssestream.NewStream.
type fakeSSEDecoder struct {
	events []ssestream.Event
	i      int
}

func (d *fakeSSEDecoder) Next() bool {
	if d.i >= len(d.events) {
		return false
	}
	d.i++
	return true
}
func (d *fakeSSEDecoder) Event() ssestream.Event { return d.events[d.i-1] }
func (d *fakeSSEDecoder) Close() error           { return nil }
func (d *fakeSSEDecoder) Err() error             { return nil }

type fakeMessageStreamer struct {
	events []ssestream.Event
}

func (f *fakeMessageStreamer) NewStreaming(_ context.Context, _ anthropic.MessageNewParams, _ ...option.RequestOption) *ssestream.Stream[anthropic.MessageStreamEventUnion] {
	return ssestream.NewStream[anthropic.MessageStreamEventUnion](&fakeSSEDecoder{events: f.events}, nil)
}

func TestStreamTurnRunnerForwardsTextDeltasAndAccumulatesStopReason(t *testing.T) {
	events := []ssestream.Event{
		{Type: "content_block_start", Data: []byte(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)},
		{Type: "content_block_delta", Data: []byte(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello "}}`)},
		{Type: "content_block_delta", Data: []byte(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"world"}}`)},
		{Type: "message_delta", Data: []byte(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}`)},
	}
	runner := streamTurnRunner{messages: &fakeMessageStreamer{events: events}}
	w := &fakeStreamWriter{}

	msg, err := runner.runTurn(context.Background(), anthropic.MessageNewParams{}, w)
	if err != nil {
		t.Fatalf("runTurn returned error: %v", err)
	}
	if len(w.deltas) != 2 || w.deltas[0] != "Hello " || w.deltas[1] != "world" {
		t.Errorf("expected 2 forwarded text deltas, got %+v", w.deltas)
	}
	if msg.StopReason != anthropic.StopReasonEndTurn {
		t.Errorf("expected accumulated stop_reason end_turn, got %v", msg.StopReason)
	}
}

func TestStreamTurnRunnerPropagatesWriteDeltaError(t *testing.T) {
	events := []ssestream.Event{
		{Type: "content_block_start", Data: []byte(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)},
		{Type: "content_block_delta", Data: []byte(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`)},
	}
	runner := streamTurnRunner{messages: &fakeMessageStreamer{events: events}}

	_, err := runner.runTurn(context.Background(), anthropic.MessageNewParams{}, erroringStreamWriter{})
	if err == nil {
		t.Fatal("expected an error when the stream writer fails, got nil")
	}
}
