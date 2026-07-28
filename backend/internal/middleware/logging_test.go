package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beetrack/backend/pkg/logging"
	"github.com/google/uuid"
)

// captureLog points the global slog default at buf for the duration of the
// test and restores the previous default on cleanup.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	prev := slog.Default()
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func lastLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("expected valid JSON log line, got %q: %v", lines[len(lines)-1], err)
	}
	return entry
}

func TestLogging_GeneratesRequestIDWhenAbsent(t *testing.T) {
	captureLog(t)
	handler := Logging(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apiaries", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get(RequestIDHeader)
	if id == "" {
		t.Fatal("expected a generated X-Request-Id header")
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Errorf("expected X-Request-Id to be a UUID, got %q", id)
	}
}

func TestLogging_ReusesIncomingRequestID(t *testing.T) {
	captureLog(t)
	handler := Logging(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apiaries", nil)
	req.Header.Set(RequestIDHeader, "existing-request-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != "existing-request-id" {
		t.Errorf("expected X-Request-Id to be reused, got %q", got)
	}
}

func TestLogging_LogsMethodPathAndStatus(t *testing.T) {
	buf := captureLog(t)
	notFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	handler := Logging(notFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := lastLogLine(t, buf)
	if entry["method"] != http.MethodGet {
		t.Errorf("expected method %q, got %v", http.MethodGet, entry["method"])
	}
	if entry["path"] != "/api/v1/missing" {
		t.Errorf("expected path %q, got %v", "/api/v1/missing", entry["path"])
	}
	if entry["status"] != float64(http.StatusNotFound) {
		t.Errorf("expected status %d, got %v", http.StatusNotFound, entry["status"])
	}
	if entry["request_id"] == nil || entry["request_id"] == "" {
		t.Error("expected request_id to be present in the log line")
	}
}

func TestLogging_DefaultsStatusTo200WhenWriteHeaderNotCalled(t *testing.T) {
	buf := captureLog(t)
	implicit200 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	handler := Logging(implicit200)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := lastLogLine(t, buf)
	if entry["status"] != float64(http.StatusOK) {
		t.Errorf("expected default status %d, got %v", http.StatusOK, entry["status"])
	}
}

// nonFlushingResponseWriter implements only http.ResponseWriter, deliberately
// omitting Flush(), to exercise the no-op path in statusWriter.Flush().
type nonFlushingResponseWriter struct {
	header http.Header
}

func (w *nonFlushingResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *nonFlushingResponseWriter) Write(b []byte) (int, error) { return len(b), nil }

func (w *nonFlushingResponseWriter) WriteHeader(statusCode int) {}

func TestLogging_FlushesUnderlyingRecorderWhenSupported(t *testing.T) {
	captureLog(t)
	flushing := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected ResponseWriter passed to handler to implement http.Flusher")
		}
		flusher.Flush()
	})
	handler := Logging(flushing)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !rec.Flushed {
		t.Error("expected the underlying httptest.ResponseRecorder to have been flushed")
	}
}

func TestLogging_FlushNoOpsWhenUnderlyingWriterDoesNotSupportIt(t *testing.T) {
	captureLog(t)
	var asserted bool
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		asserted = ok
		if ok {
			flusher.Flush()
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(&nonFlushingResponseWriter{}, req)

	if !asserted {
		t.Fatal("expected statusWriter to still satisfy http.Flusher even when the underlying writer doesn't")
	}
}

func TestLogging_AttachesRequestScopedLoggerToContext(t *testing.T) {
	buf := captureLog(t)
	var gotSameLogger bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.FromContext(r.Context()).Info("inner handler line")
		gotSameLogger = logging.FromContext(r.Context()) != slog.Default()
		w.WriteHeader(http.StatusOK)
	})
	handler := Logging(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !gotSameLogger {
		t.Error("expected context to carry a request-scoped logger distinct from slog.Default()")
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines (inner handler + access log), got %d: %q", len(lines), buf.String())
	}

	var inner, access map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &inner); err != nil {
		t.Fatalf("inner log line is not valid JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &access); err != nil {
		t.Fatalf("access log line is not valid JSON: %v", err)
	}
	if inner["request_id"] == "" || inner["request_id"] != access["request_id"] {
		t.Errorf("expected both log lines to share request_id, got %v and %v", inner["request_id"], access["request_id"])
	}
}
