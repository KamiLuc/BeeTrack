package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRegistryCallDispatchesToRegisteredTool(t *testing.T) {
	r := NewRegistry()
	var gotUserID int64
	var gotInput string
	r.Register(Tool{
		Name: "echo",
		Handler: func(_ context.Context, userID int64, input json.RawMessage) (any, error) {
			gotUserID = userID
			gotInput = string(input)
			return "ok", nil
		},
	})

	result, err := r.Call(context.Background(), 42, "echo", json.RawMessage(`{"foo":"bar"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if result != "ok" {
		t.Errorf("expected result %q, got %v", "ok", result)
	}
	if gotUserID != 42 {
		t.Errorf("expected userID 42, got %d", gotUserID)
	}
	if gotInput != `{"foo":"bar"}` {
		t.Errorf("expected input %q, got %q", `{"foo":"bar"}`, gotInput)
	}
}

func TestRegistryCallUnknownToolReturnsError(t *testing.T) {
	r := NewRegistry()

	if _, err := r.Call(context.Background(), 1, "does_not_exist", nil); err == nil {
		t.Fatal("expected an error for an unregistered tool, got nil")
	}
}

func TestRegistryAllReturnsEveryRegisteredTool(t *testing.T) {
	r := NewRegistry()
	r.Register(Tool{Name: "a", Handler: func(context.Context, int64, json.RawMessage) (any, error) { return nil, nil }})
	r.Register(Tool{Name: "b", Handler: func(context.Context, int64, json.RawMessage) (any, error) { return nil, nil }})

	tools := r.All()
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
}

func TestRegistryGetMissingToolReturnsFalse(t *testing.T) {
	r := NewRegistry()

	if _, ok := r.Get("missing"); ok {
		t.Fatal("expected ok=false for a missing tool")
	}
}
