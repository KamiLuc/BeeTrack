package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/beetrack/backend/internal/blockchain"
	"github.com/beetrack/backend/internal/model"
)

// ErrBlockchainNotConfigured means the API started without blockchain
// integration configured (see startHoneyCertificationWorker), so there's no
// RPC connection available to estimate against.
var ErrBlockchainNotConfigured = errors.New("blockchain integration is not configured")

// CertificationRequestDetailReader is the read-only slice of
// CertificationRequestStore CertificationGasEstimateService needs.
type CertificationRequestDetailReader interface {
	GetByID(ctx context.Context, id int64) (*model.HoneyBatchCertificationRequestDetail, error)
}

// GasReader is the read-only blockchain interface CertificationGasEstimateService
// needs — satisfied by *blockchain.HoneyCertReader.
type GasReader interface {
	EstimateCertifyGas(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (gasUnits uint64, gasPriceWei *big.Int, err error)
	AccountBalance(ctx context.Context) (*big.Int, error)
}

// GasEstimate is the cost preview shown to an admin before approving a
// certification request — a dry run, nothing is broadcast.
type GasEstimate struct {
	GasUnits          uint64
	GasPriceWei       *big.Int
	EstimatedCostWei  *big.Int
	AccountBalanceWei *big.Int
	MaticPLNRate      *float64 // nil if the price feed is unavailable
}

// CostMatic returns the estimated transaction cost in MATIC.
func (e *GasEstimate) CostMatic() float64 { return weiToMatic(e.EstimatedCostWei) }

// BalanceMatic returns the signing account's current balance in MATIC.
func (e *GasEstimate) BalanceMatic() float64 { return weiToMatic(e.AccountBalanceWei) }

func weiToMatic(wei *big.Int) float64 {
	f := new(big.Float).SetInt(wei)
	f.Quo(f, big.NewFloat(1e18))
	v, _ := f.Float64()
	return v
}

// CertificationGasEstimateService computes a pre-approval gas cost preview
// for a pending certification request, without submitting anything on-chain.
type CertificationGasEstimateService struct {
	requests CertificationRequestDetailReader
	batches  HoneyBatchRepository
	reader   GasReader
}

// NewCertificationGasEstimateService creates a CertificationGasEstimateService.
// reader may be a nil interface if blockchain integration isn't configured —
// Estimate then returns ErrBlockchainNotConfigured.
func NewCertificationGasEstimateService(requests CertificationRequestDetailReader, batches HoneyBatchRepository, reader GasReader) *CertificationGasEstimateService {
	return &CertificationGasEstimateService{requests: requests, batches: batches, reader: reader}
}

// Estimate looks up requestID's batch, decodes its stored hashes, and
// returns a dry-run gas estimate for the certify() call the worker would
// eventually submit once the request is approved.
func (s *CertificationGasEstimateService) Estimate(ctx context.Context, requestID int64) (*GasEstimate, error) {
	if s.reader == nil {
		return nil, ErrBlockchainNotConfigured
	}

	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("get certification request: %w", err)
	}
	if req == nil {
		return nil, ErrCertificationRequestNotFound
	}

	batch, err := s.batches.GetByID(ctx, req.BatchID)
	if err != nil {
		return nil, fmt.Errorf("get batch: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}

	pdfHash, err := blockchain.DecodeHash32(batch.PDFFileHash)
	if err != nil {
		return nil, fmt.Errorf("decode pdf hash: %w", err)
	}
	metadataHash, err := blockchain.DecodeHash32(batch.MetadataHash)
	if err != nil {
		return nil, fmt.Errorf("decode metadata hash: %w", err)
	}

	gasUnits, gasPriceWei, err := s.reader.EstimateCertifyGas(ctx, batch.ID, pdfHash, metadataHash)
	if err != nil {
		return nil, fmt.Errorf("estimate certify gas: %w", err)
	}

	balanceWei, err := s.reader.AccountBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account balance: %w", err)
	}

	estimatedCostWei := new(big.Int).Mul(gasPriceWei, new(big.Int).SetUint64(gasUnits))

	var rate *float64
	if price, err := blockchain.MaticPLNPrice(ctx); err == nil {
		rate = &price
	}

	return &GasEstimate{
		GasUnits:          gasUnits,
		GasPriceWei:       gasPriceWei,
		EstimatedCostWei:  estimatedCostWei,
		AccountBalanceWei: balanceWei,
		MaticPLNRate:      rate,
	}, nil
}
