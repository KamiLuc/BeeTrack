package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/pkg/token"
)

const testHelpersAuthSecret = "test-helpers-secret"

func TestRequireAuth_MissingToken(t *testing.T) {
	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := requireAuth(w, r); ok {
			t.Error("expected ok=false with no auth context")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "MISSING_TOKEN") {
		t.Errorf("expected MISSING_TOKEN in body, got %q", rec.Body.String())
	}
}

func TestRequireAuth_Success(t *testing.T) {
	var gotUserID int64
	var gotOK bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, gotOK = requireAuth(w, r)
	})
	handler := middleware.Auth(testHelpersAuthSecret)(inner)

	tokenStr, err := token.NewAccessToken(42, testHelpersAuthSecret, 5)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !gotOK {
		t.Fatal("expected ok=true with a valid token")
	}
	if gotUserID != 42 {
		t.Errorf("expected userID=42, got %d", gotUserID)
	}
}

func TestDecodeJSON_MalformedBody(t *testing.T) {
	var dest struct {
		Name string `json:"name"`
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if decodeJSON(w, r, &dest) {
			t.Error("expected decodeJSON to return false for malformed body")
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "INVALID_BODY") {
		t.Errorf("expected INVALID_BODY in body, got %q", rec.Body.String())
	}
}

func TestDecodeJSON_Success(t *testing.T) {
	var dest struct {
		Name string `json:"name"`
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !decodeJSON(w, r, &dest) {
			t.Fatal("expected decodeJSON to return true for valid body")
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Wildflower"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if dest.Name != "Wildflower" {
		t.Errorf("expected Name=%q, got %q", "Wildflower", dest.Name)
	}
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	var dest struct {
		Name string `json:"name"`
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if decodeJSON(w, r, &dest) {
			t.Error("expected decodeJSON to return false for an empty body")
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "INVALID_BODY") {
		t.Errorf("expected INVALID_BODY in body, got %q", rec.Body.String())
	}
}

func TestParsePathID_MissingParam(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := parsePathID(w, r, "id", "invalid apiary id"); ok {
			t.Error("expected ok=false when the path param is not set")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid apiary id") {
		t.Errorf("expected message %q in body, got %q", "invalid apiary id", rec.Body.String())
	}
}

func TestParsePathID_Invalid(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := parsePathID(w, r, "id", "invalid apiary id"); ok {
			t.Error("expected ok=false for a non-numeric path value")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetPathValue("id", "not-a-number")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid apiary id") {
		t.Errorf("expected message %q in body, got %q", "invalid apiary id", rec.Body.String())
	}
}

func TestParsePathID_Success(t *testing.T) {
	var gotID int64
	var gotOK bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, gotOK = parsePathID(w, r, "id", "invalid apiary id")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !gotOK {
		t.Fatal("expected ok=true for a valid numeric path value")
	}
	if gotID != 7 {
		t.Errorf("expected id=7, got %d", gotID)
	}
}
