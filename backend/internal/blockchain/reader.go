package blockchain

import (
	_ "embed"

	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/beetrack/backend/internal/config"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:embed contracts/HoneyCertification.abi
var honeyCertificationABIJSON string

// ErrBatchNotCertified means no on-chain certification exists yet for the batch id.
var ErrBatchNotCertified = errors.New("batch not certified on-chain")

// CertificationRecord mirrors the contract's stored Certification struct.
type CertificationRecord struct {
	PDFHash      [32]byte
	MetadataHash [32]byte
	Timestamp    time.Time
	CertifiedBy  common.Address
}

// HoneyCertReader provides read-only access to the HoneyCertification contract.
type HoneyCertReader struct {
	client   *ethclient.Client
	contract *bind.BoundContract
}

// NewHoneyCertReader dials the configured Polygon RPC endpoint and returns a
// HoneyCertReader bound to the configured contract address.
func NewHoneyCertReader(cfg config.BlockchainConfig) (*HoneyCertReader, error) {
	client, err := ethclient.Dial(cfg.PolygonRPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial RPC: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(honeyCertificationABIJSON))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("parse ABI: %w", err)
	}

	address := common.HexToAddress(cfg.ContractAddress)
	contract := bind.NewBoundContract(address, parsedABI, client, client, client)

	return &HoneyCertReader{client: client, contract: contract}, nil
}

// Close releases the underlying RPC connection.
func (r *HoneyCertReader) Close() {
	r.client.Close()
}

// GetCertification returns the stored record for batchID, or
// ErrBatchNotCertified if the batch hasn't been certified.
func (r *HoneyCertReader) GetCertification(ctx context.Context, batchID int64) (*CertificationRecord, error) {
	var result []interface{}
	err := r.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getCertification", big.NewInt(batchID))
	if err != nil {
		if strings.Contains(err.Error(), "NotCertified") {
			return nil, ErrBatchNotCertified
		}
		return nil, fmt.Errorf("call getCertification: %w", err)
	}

	pdfHash := result[0].([32]byte)
	metadataHash := result[1].([32]byte)
	timestamp := result[2].(*big.Int)
	certifiedBy := result[3].(common.Address)

	if timestamp.Sign() == 0 {
		return nil, ErrBatchNotCertified
	}

	return &CertificationRecord{
		PDFHash:      pdfHash,
		MetadataHash: metadataHash,
		Timestamp:    time.Unix(timestamp.Int64(), 0).UTC(),
		CertifiedBy:  certifiedBy,
	}, nil
}

// GetTransactionStatus reports whether txHash is mined, reverted, its block
// number, gas used, and confirmation count. mined=false with a nil error
// means the transaction just hasn't landed in a block yet — not a failure.
func (r *HoneyCertReader) GetTransactionStatus(ctx context.Context, txHash string) (mined, reverted bool, blockNumber, gasUsed, confirmations uint64, err error) {
	receipt, err := r.client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if errors.Is(err, ethereum.NotFound) {
		return false, false, 0, 0, 0, nil
	}
	if err != nil {
		return false, false, 0, 0, 0, fmt.Errorf("get transaction receipt: %w", err)
	}

	currentBlock, err := r.client.BlockNumber(ctx)
	if err != nil {
		return false, false, 0, 0, 0, fmt.Errorf("get current block number: %w", err)
	}

	confs := currentBlock - receipt.BlockNumber.Uint64() + 1
	return true, receipt.Status == 0, receipt.BlockNumber.Uint64(), receipt.GasUsed, confs, nil
}
