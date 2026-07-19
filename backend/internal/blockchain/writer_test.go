package blockchain

import (
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/config"
)

func TestNewHoneyCertWriter_InvalidPrivateKey(t *testing.T) {
	cfg := config.BlockchainConfig{
		PolygonRPCURL:   "http://localhost:8545",
		ContractAddress: "0x1234567890123456789012345678901234567890",
		PrivateKey:      "not-valid-hex",
		ChainID:         80002,
	}

	_, err := NewHoneyCertWriter(cfg)
	if err == nil {
		t.Fatal("expected error for invalid private key, got nil")
	}
	if !strings.Contains(err.Error(), "parse private key") {
		t.Errorf("expected error to mention private key parsing, got: %v", err)
	}
}
