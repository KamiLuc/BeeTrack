package middleware

import (
	"context"
	"net/http"
	"strings"

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

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
