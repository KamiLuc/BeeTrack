package handler

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http/httptest"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
)

// stubGasEstimateRequestReader is a minimal service.CertificationRequestDetailReader
// stub, indexed by request ID.
type stubGasEstimateRequestReader struct {
	byID map[int64]*model.HoneyBatchCertificationRequestDetail
	err  error
}

func (r *stubGasEstimateRequestReader) GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertificationRequestDetail, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.byID[id], nil
}

// stubGasEstimateReader is a minimal service.GasReader stub.
type stubGasEstimateReader struct {
	gasUnits    uint64
	gasPriceWei *big.Int
	estimateErr error

	balanceWei *big.Int
	balanceErr error
}

func (r *stubGasEstimateReader) EstimateCertifyGas(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (uint64, *big.Int, error) {
	if r.estimateErr != nil {
		return 0, nil, r.estimateErr
	}
	return r.gasUnits, r.gasPriceWei, nil
}

func (r *stubGasEstimateReader) AccountBalance(ctx context.Context) (*big.Int, error) {
	if r.balanceErr != nil {
		return nil, r.balanceErr
	}
	return r.balanceWei, nil
}

const gasEstimateTestHash = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

// newGasEstimateTestHandler wires an AdminCertificationHandler whose review
// and batches dependencies are unused by EstimateGas, so they're left nil.
func newGasEstimateTestHandler(requests *stubGasEstimateRequestReader, batches *stubCertsBatchRepo, reader *stubGasEstimateReader) *AdminCertificationHandler {
	var gasEstimate *service.CertificationGasEstimateService
	if reader == nil {
		gasEstimate = service.NewCertificationGasEstimateService(requests, batches, nil)
	} else {
		gasEstimate = service.NewCertificationGasEstimateService(requests, batches, reader)
	}
	return NewAdminCertificationHandler(nil, nil, gasEstimate)
}

func TestAdminCertification_EstimateGas_InvalidID(t *testing.T) {
	h := newGasEstimateTestHandler(&stubGasEstimateRequestReader{}, &stubCertsBatchRepo{}, &stubGasEstimateReader{})

	req := httptest.NewRequest("GET", "/api/v1/admin/certification-requests/abc/estimate-gas", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	h.EstimateGas(rec, req)

	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAdminCertification_EstimateGas_NotFound(t *testing.T) {
	h := newGasEstimateTestHandler(&stubGasEstimateRequestReader{byID: map[int64]*model.HoneyBatchCertificationRequestDetail{}}, &stubCertsBatchRepo{}, &stubGasEstimateReader{})

	req := httptest.NewRequest("GET", "/api/v1/admin/certification-requests/1/estimate-gas", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.EstimateGas(rec, req)

	if rec.Code != 404 {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestAdminCertification_EstimateGas_BlockchainNotConfigured(t *testing.T) {
	requests := &stubGasEstimateRequestReader{byID: map[int64]*model.HoneyBatchCertificationRequestDetail{
		1: {HoneyBatchCertificationRequest: model.HoneyBatchCertificationRequest{ID: 1, BatchID: 5}},
	}}
	batches := &stubCertsBatchRepo{byID: map[int64]*model.HoneyBatch{5: {ID: 5, PDFFileHash: gasEstimateTestHash, MetadataHash: gasEstimateTestHash}}}
	h := newGasEstimateTestHandler(requests, batches, nil)

	req := httptest.NewRequest("GET", "/api/v1/admin/certification-requests/1/estimate-gas", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.EstimateGas(rec, req)

	if rec.Code != 503 {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestAdminCertification_EstimateGas_BatchNotFound(t *testing.T) {
	requests := &stubGasEstimateRequestReader{byID: map[int64]*model.HoneyBatchCertificationRequestDetail{
		1: {HoneyBatchCertificationRequest: model.HoneyBatchCertificationRequest{ID: 1, BatchID: 5}},
	}}
	batches := &stubCertsBatchRepo{byID: map[int64]*model.HoneyBatch{}}
	h := newGasEstimateTestHandler(requests, batches, &stubGasEstimateReader{})

	req := httptest.NewRequest("GET", "/api/v1/admin/certification-requests/1/estimate-gas", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.EstimateGas(rec, req)

	if rec.Code != 404 {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestAdminCertification_EstimateGas_InternalError(t *testing.T) {
	requests := &stubGasEstimateRequestReader{err: errors.New("db down")}
	h := newGasEstimateTestHandler(requests, &stubCertsBatchRepo{}, &stubGasEstimateReader{})

	req := httptest.NewRequest("GET", "/api/v1/admin/certification-requests/1/estimate-gas", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.EstimateGas(rec, req)

	if rec.Code != 500 {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestAdminCertification_EstimateGas_JSONShape(t *testing.T) {
	requests := &stubGasEstimateRequestReader{byID: map[int64]*model.HoneyBatchCertificationRequestDetail{
		1: {HoneyBatchCertificationRequest: model.HoneyBatchCertificationRequest{ID: 1, BatchID: 5}},
	}}
	batches := &stubCertsBatchRepo{byID: map[int64]*model.HoneyBatch{5: {ID: 5, PDFFileHash: gasEstimateTestHash, MetadataHash: gasEstimateTestHash}}}
	reader := &stubGasEstimateReader{
		gasUnits:    21000,
		gasPriceWei: big.NewInt(30_000_000_000),
		balanceWei:  big.NewInt(1_000_000_000_000_000_000),
	}
	h := newGasEstimateTestHandler(requests, batches, reader)

	// A pre-canceled context deterministically fails the MaticPLNPrice HTTP
	// call (without touching the network) so we can assert the nullable PLN
	// fields stay nil, rather than depending on a live CoinGecko response.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest("GET", "/api/v1/admin/certification-requests/1/estimate-gas", nil).WithContext(ctx)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.EstimateGas(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["gas_price_wei"] != "30000000000" {
		t.Errorf("expected gas_price_wei as decimal string, got %v (%T)", body["gas_price_wei"], body["gas_price_wei"])
	}
	if body["estimated_cost_wei"] != "630000000000000" {
		t.Errorf("expected estimated_cost_wei=%q, got %v", "630000000000000", body["estimated_cost_wei"])
	}
	if body["account_balance_wei"] != "1000000000000000000" {
		t.Errorf("expected account_balance_wei as decimal string, got %v", body["account_balance_wei"])
	}
	if body["matic_pln_rate"] != nil {
		t.Errorf("expected nil matic_pln_rate when price feed errors, got %v", body["matic_pln_rate"])
	}
	if body["estimated_cost_pln"] != nil {
		t.Errorf("expected nil estimated_cost_pln when price feed errors, got %v", body["estimated_cost_pln"])
	}
	if body["account_balance_pln"] != nil {
		t.Errorf("expected nil account_balance_pln when price feed errors, got %v", body["account_balance_pln"])
	}
}
