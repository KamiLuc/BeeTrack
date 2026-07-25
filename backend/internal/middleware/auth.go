package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/beetrack/backend/pkg/logging"
	"github.com/beetrack/backend/pkg/respond"
	"github.com/beetrack/backend/pkg/token"
)

type contextKey string

const userIDKey contextKey = "userID"

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := token.ParseAccessToken(tokenStr, secret)
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token")
				return
			}

			next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID)))
		})
	}
}

// OptionalAuth attaches the userID to the request context when a valid Bearer
// token is present, but allows the request to proceed anonymously otherwise.
func OptionalAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				if userID, err := token.ParseAccessToken(tokenStr, secret); err == nil {
					r = r.WithContext(withUserID(r.Context(), userID))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

// withUserID attaches userID to ctx and enriches ctx's request-scoped logger
// with it, so any handler/service that logs via logging.FromContext(ctx)
// after this point automatically includes it.
func withUserID(ctx context.Context, userID int64) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return logging.WithContext(ctx, logging.FromContext(ctx).With("user_id", userID))
}
