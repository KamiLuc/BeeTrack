package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/beetrack/backend/pkg/logging"
	"github.com/google/uuid"
)

// RequestIDHeader is the header used to propagate a request ID from an
// upstream proxy, or to hand a generated one back to the client.
const RequestIDHeader = "X-Request-Id"

// Logging returns a middleware that assigns a request ID (reusing the one in
// the incoming X-Request-Id header if present), attaches a request-scoped
// logger to the request context for downstream handlers/services, and logs
// one structured line per request with the method, path, status, duration,
// and remote address.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set(RequestIDHeader, requestID)

		logger := slog.Default().With("request_id", requestID)
		r = r.WithContext(logging.WithContext(r.Context(), logger))

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sw, r)

		logger.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// statusWriter records the status code passed to WriteHeader so it can be
// logged after the handler returns; http.ResponseWriter doesn't expose it.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
