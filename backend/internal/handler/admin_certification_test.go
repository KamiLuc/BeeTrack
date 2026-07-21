package handler

import (
	"net/http/httptest"
	"testing"
)

func TestParseCertificationReviewQuery(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus string
		wantKey    string
		wantSort   string
		wantOK     bool
	}{
		{name: "no params defaults to all statuses and asc", url: "/certification-requests", wantStatus: "", wantKey: "", wantSort: "asc", wantOK: true},
		{name: "pending status accepted", url: "/certification-requests?status=pending", wantStatus: "pending", wantSort: "asc", wantOK: true},
		{name: "approved status accepted", url: "/certification-requests?status=approved", wantStatus: "approved", wantSort: "asc", wantOK: true},
		{name: "rejected status accepted", url: "/certification-requests?status=rejected", wantStatus: "rejected", wantSort: "asc", wantOK: true},
		{name: "invalid status rejected", url: "/certification-requests?status=bogus", wantOK: false},
		{name: "desc sort accepted", url: "/certification-requests?sort=desc", wantSort: "desc", wantOK: true},
		{name: "unknown sort falls back to asc", url: "/certification-requests?sort=sideways", wantSort: "asc", wantOK: true},
		{name: "keyword is trimmed", url: "/certification-requests?q=%20clover%20", wantKey: "clover", wantSort: "asc", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)

			status, keyword, sortDir, ok := parseCertificationReviewQuery(r)

			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if !tt.wantOK {
				return
			}
			if status != tt.wantStatus {
				t.Errorf("expected status %q, got %q", tt.wantStatus, status)
			}
			if keyword != tt.wantKey {
				t.Errorf("expected keyword %q, got %q", tt.wantKey, keyword)
			}
			if sortDir != tt.wantSort {
				t.Errorf("expected sortDir %q, got %q", tt.wantSort, sortDir)
			}
		})
	}
}
