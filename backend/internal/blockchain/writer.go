package blockchain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/beetrack/backend/internal/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ErrAlreadyCertified means the contract already has a live certification
// for the batch — the worker treats this as success, not a failure.
var ErrAlreadyCertified = errors.New("batch already certified on-chain")

// HoneyCertWriter signs and broadcasts certify() transactions. Only the
// background worker should hold one — never the HTTP request path.
type HoneyCertWriter struct {
	client     *ethclient.Client
	contract   *bind.BoundContract
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
}

// NewHoneyCertWriter dials the configured Polygon RPC endpoint, loads the
// signing key, and returns a HoneyCertWriter bound to the configured
// contract address.
func NewHoneyCertWriter(cfg config.BlockchainConfig) (*HoneyCertWriter, error) {
	client, err := ethclient.Dial(cfg.PolygonRPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial RPC: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(honeyCertificationABIJSON))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("parse ABI: %w", err)
	}

	address := common.HexToAddress(cfg.ContractAddress)
	contract := bind.NewBoundContract(address, parsedABI, client, client, client)

	return &HoneyCertWriter{
		client:     client,
		contract:   contract,
		privateKey: privateKey,
		chainID:    big.NewInt(int64(cfg.ChainID)),
	}, nil
}

// Close releases the underlying RPC connection.
func (w *HoneyCertWriter) Close() {
	w.client.Close()
}

// CertifyBatch signs and broadcasts certify(batchID, pdfHash, metadataHash),
// returning the transaction hash immediately without waiting for
// confirmation. Returns ErrAlreadyCertified if batchID already has a live
// certification.
func (w *HoneyCertWriter) CertifyBatch(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (string, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(w.privateKey, w.chainID)
	if err != nil {
		return "", fmt.Errorf("build transactor: %w", err)
	}
	opts.Context = ctx

	tx, err := w.contract.Transact(opts, "certify", big.NewInt(batchID), pdfHash, metadataHash)
	if err != nil {
		if strings.Contains(err.Error(), "AlreadyCertified") {
			return "", ErrAlreadyCertified
		}
		return "", fmt.Errorf("send certify transaction: %w", err)
	}

	return tx.Hash().Hex(), nil
}
