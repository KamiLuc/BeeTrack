package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestInit_WritesJSONAtConfiguredLevel(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, "warn")

	slog.Info("should be filtered out")
	slog.Warn("something odd", "key", "value")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d: %q", len(lines), buf.String())
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if entry["msg"] != "something odd" {
		t.Errorf("expected msg %q, got %v", "something odd", entry["msg"])
	}
	if entry["key"] != "value" {
		t.Errorf("expected key %q, got %v", "value", entry["key"])
	}
}

func TestInit_UnrecognizedLevelDefaultsToInfo(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, "not-a-real-level")

	slog.Info("visible at default level")

	if !strings.Contains(buf.String(), "visible at default level") {
		t.Errorf("expected info line to be written, got %q", buf.String())
	}
}

func TestFromContext_ReturnsDefaultWhenUnset(t *testing.T) {
	logger := FromContext(context.Background())
	if logger != slog.Default() {
		t.Error("expected slog.Default() when context carries no logger")
	}
}

func TestWithContext_RoundTrips(t *testing.T) {
	var buf bytes.Buffer
	want := slog.New(slog.NewJSONHandler(&buf, nil)).With("request_id", "abc123")

	ctx := WithContext(context.Background(), want)
	got := FromContext(ctx)

	if got != want {
		t.Error("expected FromContext to return the exact logger stored by WithContext")
	}
}
