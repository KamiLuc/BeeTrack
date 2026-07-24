package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/token"
	"gorm.io/gorm"
)

const testReportAuthSecret = "test-report-secret"

type reportMockApiaryRepo struct{ apiary *model.Apiary }

func (m *reportMockApiaryRepo) GetMembership(_ context.Context, apiaryID, _ int64) (*model.Apiary, string, error) {
	if m.apiary == nil || m.apiary.ID != apiaryID {
		return nil, "", gorm.ErrRecordNotFound
	}
	return m.apiary, "member", nil
}

type reportMockHiveReader struct{ hives []*model.Hive }

func (m *reportMockHiveReader) ListByApiaryID(_ context.Context, _ int64) ([]*model.Hive, error) {
	return m.hives, nil
}

type reportMockInspectionReader struct{ inspections []*model.Inspection }

func (m *reportMockInspectionReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Inspection, error) {
	return m.inspections, nil
}

type reportMockTreatmentReader struct{}

func (m *reportMockTreatmentReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Treatment, error) {
	return nil, nil
}

type reportMockFeedingReader struct{}

func (m *reportMockFeedingReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Feeding, error) {
	return nil, nil
}

type reportMockHarvestReader struct{}

func (m *reportMockHarvestReader) ListByHiveIDsAndRange(_ context.Context, _ []int64, _, _ time.Time) ([]*model.Harvest, error) {
	return nil, nil
}

func newTestReportHandler() *ReportHandler {
	apiary := &model.Apiary{ID: 1, Name: "Pasieka Testowa"}
	hives := []*model.Hive{{ID: 10, ApiaryID: 1, Name: "Ul 1"}}
	inspections := []*model.Inspection{
		{ID: 1, HiveID: 10, InspectedAt: time.Now(), QueenStatus: "seen", Notes: "Test"},
	}
	svc := service.NewReportService(
		&reportMockApiaryRepo{apiary: apiary},
		&reportMockHiveReader{hives: hives},
		&reportMockInspectionReader{inspections: inspections},
		&reportMockTreatmentReader{},
		&reportMockFeedingReader{},
		&reportMockHarvestReader{},
	)
	return NewReportHandler(svc)
}

func authedReportRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	tokenStr, err := token.NewAccessToken(1, testReportAuthSecret, 5)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apiaries/1/report/pdf", strings.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	return req
}

func TestReportPDF_MissingToken(t *testing.T) {
	h := newTestReportHandler()
	handler := middleware.Auth(testReportAuthSecret)(http.HandlerFunc(h.PDF))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/apiaries/1/report/pdf", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestReportPDF_InvalidBody(t *testing.T) {
	h := newTestReportHandler()
	handler := middleware.Auth(testReportAuthSecret)(http.HandlerFunc(h.PDF))

	req := authedReportRequest(t, "not json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestReportPDF_InvalidDate(t *testing.T) {
	h := newTestReportHandler()
	handler := middleware.Auth(testReportAuthSecret)(http.HandlerFunc(h.PDF))

	body := `{"hive_ids":[10],"categories":["inspections"],"from":"not-a-date","to":"2026-07-24"}`
	req := authedReportRequest(t, body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestReportPDF_InvalidCategory(t *testing.T) {
	h := newTestReportHandler()
	handler := middleware.Auth(testReportAuthSecret)(http.HandlerFunc(h.PDF))

	body := `{"hive_ids":[10],"categories":["bogus"],"from":"2026-07-01","to":"2026-07-24"}`
	req := authedReportRequest(t, body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestReportPDF_Success(t *testing.T) {
	h := newTestReportHandler()
	handler := middleware.Auth(testReportAuthSecret)(http.HandlerFunc(h.PDF))

	body := `{"hive_ids":[10],"categories":["inspections"],"from":"2026-07-01","to":"2026-07-24"}`
	req := authedReportRequest(t, body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %q", ct)
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF")) {
		t.Errorf("expected response body to be a PDF, got %d bytes starting %q", rec.Body.Len(), rec.Body.Bytes()[:min(20, rec.Body.Len())])
	}
}
