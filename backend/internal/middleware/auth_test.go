package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beetrack/backend/pkg/token"
)

func TestAuth_MissingTokenReturns401(t *testing.T) {
	handler := Auth("secret")(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_InvalidTokenReturns401(t *testing.T) {
	handler := Auth("secret")(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ValidTokenAttachesUserID(t *testing.T) {
	tokenStr, err := token.NewAccessToken(42, "secret", 5)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	var gotUserID int64
	var gotOK bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, gotOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth("secret")(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !gotOK {
		t.Fatal("expected userID to be attached to context")
	}
	if gotUserID != 42 {
		t.Errorf("expected userID 42, got %d", gotUserID)
	}
}

func TestOptionalAuth_NoTokenProceedsAnonymously(t *testing.T) {
	var gotOK bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, gotOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalAuth("secret")(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotOK {
		t.Error("expected no userID in context for anonymous request")
	}
}

func TestOptionalAuth_InvalidTokenProceedsAnonymously(t *testing.T) {
	var gotOK bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, gotOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalAuth("secret")(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotOK {
		t.Error("expected no userID in context for invalid token")
	}
}

func TestOptionalAuth_ValidTokenAttachesUserID(t *testing.T) {
	tokenStr, err := token.NewAccessToken(7, "secret", 5)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	var gotUserID int64
	var gotOK bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, gotOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalAuth("secret")(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !gotOK {
		t.Fatal("expected userID to be attached to context")
	}
	if gotUserID != 7 {
		t.Errorf("expected userID 7, got %d", gotUserID)
	}
}
