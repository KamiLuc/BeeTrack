package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

type fakeAdminUserStore struct {
	users map[int64]*model.User
	err   error
}

func (f *fakeAdminUserStore) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.users[id], nil
}

func TestRequireAdmin_NoUserIDInContext(t *testing.T) {
	store := &fakeAdminUserStore{}
	handler := RequireAdmin(store)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAdmin_UserLookupError(t *testing.T) {
	store := &fakeAdminUserStore{err: errors.New("db down")}
	handler := RequireAdmin(store)(okHandler())

	req := requestWithUserID(1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRequireAdmin_UserNotFound(t *testing.T) {
	store := &fakeAdminUserStore{users: map[int64]*model.User{}}
	handler := RequireAdmin(store)(okHandler())

	req := requestWithUserID(1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireAdmin_NonAdminUserRejected(t *testing.T) {
	store := &fakeAdminUserStore{users: map[int64]*model.User{
		1: {ID: 1, Role: model.UserRoleUser},
	}}
	handler := RequireAdmin(store)(okHandler())

	req := requestWithUserID(1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "NOT_ADMIN") {
		t.Errorf("expected NOT_ADMIN in body, got %q", body)
	}
}

func TestRequireAdmin_AdminUserAllowedThrough(t *testing.T) {
	store := &fakeAdminUserStore{users: map[int64]*model.User{
		1: {ID: 1, Role: model.UserRoleAdmin},
	}}
	var calledWithUserID int64
	var calledOK bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledWithUserID, calledOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAdmin(store)(next)

	req := requestWithUserID(1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !calledOK || calledWithUserID != 1 {
		t.Errorf("expected next handler to see userID=1, got ok=%v userID=%d", calledOK, calledWithUserID)
	}
}

// requestWithUserID builds a request whose context already carries userID, as
// if Auth had already run — RequireAdmin is always chained after Auth in
// practice (see the `admin` wrapper in cmd/api/main.go), never used alone.
func requestWithUserID(userID int64) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}
