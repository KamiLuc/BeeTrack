package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Handler runs one tool call; userID is injected by the caller, never taken from the model's input.
type Handler func(ctx context.Context, userID int64, input json.RawMessage) (any, error)

type InputSchema struct {
	Properties map[string]any
	Required   []string
}

type Tool struct {
	Name        string
	Description string
	InputSchema InputSchema
	Handler     Handler
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register replaces any existing tool with the same name.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) All() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

func (r *Registry) Call(ctx context.Context, userID int64, name string, input json.RawMessage) (any, error) {
	t, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return t.Handler(ctx, userID, input)
}
