package service

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

type mockGasReader struct {
	gasUnits    uint64
	gasPriceWei *big.Int
	estimateErr error

	balanceWei *big.Int
	balanceErr error

	gotBatchID      int64
	gotPDFHash      [32]byte
	gotMetadataHash [32]byte
}

func (m *mockGasReader) EstimateCertifyGas(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (uint64, *big.Int, error) {
	m.gotBatchID = batchID
	m.gotPDFHash = pdfHash
	m.gotMetadataHash = metadataHash
	if m.estimateErr != nil {
		return 0, nil, m.estimateErr
	}
	return m.gasUnits, m.gasPriceWei, nil
}

func (m *mockGasReader) AccountBalance(ctx context.Context) (*big.Int, error) {
	if m.balanceErr != nil {
		return nil, m.balanceErr
	}
	return m.balanceWei, nil
}

const validHash = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func newTestGasEstimateService() (*CertificationGasEstimateService, *mockCertificationRequestStore, *mockHoneyBatchRepo, *mockGasReader) {
	requests := &mockCertificationRequestStore{
		byID: map[int64]*model.HoneyBatchCertificationRequestDetail{
			1: {HoneyBatchCertificationRequest: model.HoneyBatchCertificationRequest{ID: 1, BatchID: 5}},
		},
	}
	batches := &mockHoneyBatchRepo{
		byID: map[int64]*model.HoneyBatch{
			5: {ID: 5, PDFFileHash: validHash, MetadataHash: validHash},
		},
	}
	reader := &mockGasReader{
		gasUnits:    21000,
		gasPriceWei: big.NewInt(30_000_000_000),
		balanceWei:  big.NewInt(1_000_000_000_000_000_000),
	}
	svc := NewCertificationGasEstimateService(requests, batches, reader)
	return svc, requests, batches, reader
}

func TestCertificationGasEstimate_Estimate_NilPLNRateWhenPriceFeedErrors(t *testing.T) {
	svc, _, _, _ := newTestGasEstimateService()

	// A pre-canceled context makes blockchain.MaticPLNPrice's HTTP call fail
	// deterministically (context.Canceled) without touching the network,
	// while the hand-written stubs below ignore ctx entirely.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	est, err := svc.Estimate(ctx, 1)
	if err != nil {
		t.Fatalf("Estimate() error = %v", err)
	}
	if est.MaticPLNRate != nil {
		t.Errorf("expected nil MaticPLNRate when price feed errors, got %v", *est.MaticPLNRate)
	}
}

// TestCertificationGasEstimate_Estimate_Success uses a pre-canceled context
// so blockchain.MaticPLNPrice's HTTP call fails deterministically instead of
// hitting the real CoinGecko endpoint — MaticPLNPrice isn't behind an
// injectable interface (a package-level function backed by a hardcoded URL),
// so a live network dependency in this suite would otherwise be unavoidable.
// The nil-vs-non-nil rate behavior itself is covered by
// TestCertificationGasEstimate_Estimate_NilPLNRateWhenPriceFeedErrors; this
// test only asserts the numeric estimate fields.
func TestCertificationGasEstimate_Estimate_Success(t *testing.T) {
	svc, _, _, reader := newTestGasEstimateService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	est, err := svc.Estimate(ctx, 1)
	if err != nil {
		t.Fatalf("Estimate() error = %v", err)
	}
	if est.GasUnits != 21000 {
		t.Errorf("expected GasUnits=21000, got %d", est.GasUnits)
	}
	wantCost := new(big.Int).Mul(reader.gasPriceWei, big.NewInt(21000))
	if est.EstimatedCostWei.Cmp(wantCost) != 0 {
		t.Errorf("expected EstimatedCostWei=%s, got %s", wantCost, est.EstimatedCostWei)
	}
	if est.AccountBalanceWei.Cmp(reader.balanceWei) != 0 {
		t.Errorf("expected AccountBalanceWei=%s, got %s", reader.balanceWei, est.AccountBalanceWei)
	}
	if reader.gotBatchID != 5 {
		t.Errorf("expected reader called with batchID=5, got %d", reader.gotBatchID)
	}
}

func TestCertificationGasEstimate_Estimate_RequestNotFound(t *testing.T) {
	svc, _, _, _ := newTestGasEstimateService()

	_, err := svc.Estimate(context.Background(), 999)
	if err != ErrCertificationRequestNotFound {
		t.Errorf("expected ErrCertificationRequestNotFound, got %v", err)
	}
}

func TestCertificationGasEstimate_Estimate_BatchNotFound(t *testing.T) {
	requests := &mockCertificationRequestStore{
		byID: map[int64]*model.HoneyBatchCertificationRequestDetail{
			1: {HoneyBatchCertificationRequest: model.HoneyBatchCertificationRequest{ID: 1, BatchID: 5}},
		},
	}
	batches := &mockHoneyBatchRepo{byID: map[int64]*model.HoneyBatch{}}
	reader := &mockGasReader{}
	svc := NewCertificationGasEstimateService(requests, batches, reader)

	_, err := svc.Estimate(context.Background(), 1)
	if err != ErrBatchNotFound {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestCertificationGasEstimate_Estimate_NilReader(t *testing.T) {
	requests := &mockCertificationRequestStore{}
	batches := &mockHoneyBatchRepo{}
	svc := NewCertificationGasEstimateService(requests, batches, nil)

	_, err := svc.Estimate(context.Background(), 1)
	if err != ErrBlockchainNotConfigured {
		t.Errorf("expected ErrBlockchainNotConfigured, got %v", err)
	}
}

func TestCertificationGasEstimate_Estimate_BadStoredHash(t *testing.T) {
	requests := &mockCertificationRequestStore{
		byID: map[int64]*model.HoneyBatchCertificationRequestDetail{
			1: {HoneyBatchCertificationRequest: model.HoneyBatchCertificationRequest{ID: 1, BatchID: 5}},
		},
	}
	batches := &mockHoneyBatchRepo{
		byID: map[int64]*model.HoneyBatch{
			5: {ID: 5, PDFFileHash: "not-hex", MetadataHash: validHash},
		},
	}
	reader := &mockGasReader{}
	svc := NewCertificationGasEstimateService(requests, batches, reader)

	_, err := svc.Estimate(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "decode pdf hash") {
		t.Errorf("expected decode pdf hash error, got %v", err)
	}
}

func TestCertificationGasEstimate_Estimate_EstimateGasError(t *testing.T) {
	svc, _, _, reader := newTestGasEstimateService()
	reader.estimateErr = errors.New("rpc down")

	_, err := svc.Estimate(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
