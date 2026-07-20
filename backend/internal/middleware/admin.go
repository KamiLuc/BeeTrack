package middleware

import (
	"context"
	"net/http"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/pkg/respond"
)

// AdminUserStore is the minimal user lookup RequireAdmin needs.
type AdminUserStore interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

// RequireAdmin wraps a handler that already ran through Auth, rejecting
// non-admins with 403. The role is checked against the DB on every request
// rather than a JWT claim, so revoking admin access takes effect immediately.
func RequireAdmin(users AdminUserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
				return
			}
			user, err := users.GetByID(r.Context(), userID)
			if err != nil {
				respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
				return
			}
			if user == nil || !user.IsAdmin() {
				respond.Error(w, http.StatusForbidden, "NOT_ADMIN", "admin access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
