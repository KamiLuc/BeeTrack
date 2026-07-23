package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beetrack/backend/internal/blockchain"
	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/token"
)

// stubCertsBatchRepo is a minimal HoneyBatchRepository stub for testing
// HoneyBatchHandler.Certifications, indexed by batch ID so ownership checks
// (batch.UserID) can be exercised.
type stubCertsBatchRepo struct {
	byID map[int64]*model.HoneyBatch
}

func (r *stubCertsBatchRepo) CreateWithCertificationRequest(context.Context, *model.HoneyBatch, *model.HoneyBatchCertificationRequest) error {
	return nil
}
func (r *stubCertsBatchRepo) GetByID(ctx context.Context, id int64) (*model.HoneyBatch, error) {
	return r.byID[id], nil
}
func (r *stubCertsBatchRepo) GetByVerificationToken(context.Context, string) (*model.HoneyBatch, error) {
	return nil, nil
}
func (r *stubCertsBatchRepo) ListByUserID(context.Context, int64, int, int) ([]*model.HoneyBatch, error) {
	return nil, nil
}
func (r *stubCertsBatchRepo) CountByUserID(context.Context, int64) (int64, error)   { return 0, nil }
func (r *stubCertsBatchRepo) UpdateFields(context.Context, *model.HoneyBatch) error { return nil }
func (r *stubCertsBatchRepo) SoftDelete(context.Context, int64) error               { return nil }
func (r *stubCertsBatchRepo) HardDelete(context.Context, int64) error               { return nil }

// stubCertsHistoryRepo is a minimal HoneyBatchCertificationRepository stub
// returning a fixed certification history per batch ID.
type stubCertsHistoryRepo struct {
	history map[int64][]*model.HoneyBatchCertification
}

func (r *stubCertsHistoryRepo) GetLatestByBatchID(context.Context, int64) (*model.HoneyBatchCertification, error) {
	return nil, nil
}
func (r *stubCertsHistoryRepo) ListByBatchID(ctx context.Context, batchID int64) ([]*model.HoneyBatchCertification, error) {
	return r.history[batchID], nil
}

const certsTestSecret = "test-secret"

// certsAuthedRequest builds a GET request for path, attaching a Bearer token
// for userID (skipped entirely if userID is 0, to exercise the unauthenticated case).
func certsAuthedRequest(t *testing.T, path string, userID int64) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if userID != 0 {
		tok, err := token.NewAccessToken(userID, certsTestSecret, 5)
		if err != nil {
			t.Fatalf("NewAccessToken() error = %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	return req
}

// newCertsTestHandler wires a real HoneyBatchService (via stub repos) behind
// middleware.Auth, exposing HoneyBatchHandler.Certifications for a batch and
// its certification history.
func newCertsTestHandler(t *testing.T, batch *model.HoneyBatch, history []*model.HoneyBatchCertification, certReader ChainCertReader) http.Handler {
	t.Helper()
	byID := map[int64]*model.HoneyBatch{}
	if batch != nil {
		byID[batch.ID] = batch
	}
	batches := &stubCertsBatchRepo{byID: byID}
	certs := &stubCertsHistoryRepo{history: map[int64][]*model.HoneyBatchCertification{}}
	if batch != nil {
		certs.history[batch.ID] = history
	}
	svc := service.NewHoneyBatchService(batches, certs, &stubVerifyCertRequestRepo{}, &stubVerifyQRCodeRepo{}, &stubVerifyJobRepo{}, "https://example.com", t.TempDir())
	h := NewHoneyBatchHandler(svc, certReader)
	return middleware.Auth(certsTestSecret)(http.HandlerFunc(h.Certifications))
}

func TestCertifications_Unauthorized(t *testing.T) {
	handler := newCertsTestHandler(t, &model.HoneyBatch{ID: 1, UserID: 42}, nil, nil)

	req := certsAuthedRequest(t, "/api/v1/honey-batches/1/certifications", 0)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestCertifications_InvalidID(t *testing.T) {
	handler := newCertsTestHandler(t, &model.HoneyBatch{ID: 1, UserID: 42}, nil, nil)

	req := certsAuthedRequest(t, "/api/v1/honey-batches/abc/certifications", 42)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCertifications_NotOwner(t *testing.T) {
	handler := newCertsTestHandler(t, &model.HoneyBatch{ID: 1, UserID: 42}, nil, nil)

	req := certsAuthedRequest(t, "/api/v1/honey-batches/1/certifications", 999)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCertifications_ReturnsHistoryMostRecentFirst(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	cert1 := &model.HoneyBatchCertification{ID: 10, Status: model.CertificationStatusConfirmed}
	cert2 := &model.HoneyBatchCertification{ID: 9, Status: model.CertificationStatusFailed}
	handler := newCertsTestHandler(t, batch, []*model.HoneyBatchCertification{cert1, cert2}, nil)

	req := certsAuthedRequest(t, "/api/v1/honey-batches/1/certifications", 42)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(body.Items))
	}
	if body.Items[0]["status"] != string(model.CertificationStatusConfirmed) {
		t.Errorf("expected the first item to be the most recent (confirmed), got %v", body.Items[0]["status"])
	}
	if body.Items[1]["status"] != string(model.CertificationStatusFailed) {
		t.Errorf("expected the second item to be failed, got %v", body.Items[1]["status"])
	}
}

func TestCertifications_AttachesOnChainHashesOnlyForLiveRow(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	live := &model.HoneyBatchCertification{ID: 10, Status: model.CertificationStatusConfirmed}
	failed := &model.HoneyBatchCertification{ID: 9, Status: model.CertificationStatusFailed}
	var pdfHash, metadataHash [32]byte
	pdfHash[0] = 0xAB
	metadataHash[0] = 0xCD
	reader := &stubChainCertReader{record: &blockchain.CertificationRecord{PDFHash: pdfHash, MetadataHash: metadataHash}}
	handler := newCertsTestHandler(t, batch, []*model.HoneyBatchCertification{live, failed}, reader)

	req := certsAuthedRequest(t, "/api/v1/honey-batches/1/certifications", 42)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Items[0]["on_chain_pdf_hash"] == nil || body.Items[0]["on_chain_metadata_hash"] == nil {
		t.Errorf("expected on-chain hashes attached to the live row, got %+v", body.Items[0])
	}
	if body.Items[1]["on_chain_pdf_hash"] != nil || body.Items[1]["on_chain_metadata_hash"] != nil {
		t.Errorf("did not expect on-chain hashes on the non-live row, got %+v", body.Items[1])
	}
}

func TestCertifications_OmitsHashesWhenReaderNil(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	live := &model.HoneyBatchCertification{ID: 10, Status: model.CertificationStatusConfirmed}
	handler := newCertsTestHandler(t, batch, []*model.HoneyBatchCertification{live}, nil)

	req := certsAuthedRequest(t, "/api/v1/honey-batches/1/certifications", 42)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Items[0]["on_chain_pdf_hash"] != nil || body.Items[0]["on_chain_metadata_hash"] != nil {
		t.Errorf("did not expect on-chain hashes with a nil reader, got %+v", body.Items[0])
	}
}

func TestCertifications_OmitsHashesWhenReaderErrors(t *testing.T) {
	batch := &model.HoneyBatch{ID: 1, UserID: 42}
	live := &model.HoneyBatchCertification{ID: 10, Status: model.CertificationStatusConfirmed}
	reader := &stubChainCertReader{err: errors.New("rpc timeout")}
	handler := newCertsTestHandler(t, batch, []*model.HoneyBatchCertification{live}, reader)

	req := certsAuthedRequest(t, "/api/v1/honey-batches/1/certifications", 42)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Items[0]["on_chain_pdf_hash"] != nil || body.Items[0]["on_chain_metadata_hash"] != nil {
		t.Errorf("did not expect on-chain hashes when the reader errors, got %+v", body.Items[0])
	}
}
