package config

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// BlockchainConfig holds everything needed to talk to the Polygon smart
// contract: connection details for the writer/reader (internal/blockchain)
// and the timing knobs for the background worker (internal/worker).
type BlockchainConfig struct {
	PolygonRPCURL            string
	ContractAddress          string
	PrivateKey               string
	ChainID                  int
	JobPollInterval          time.Duration
	ConfirmationPollInterval time.Duration
	RequiredConfirmations    int
}

var (
	privateKeyPattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
	addressPattern    = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// LoadBlockchainConfig reads blockchain settings from the environment and
// validates them. Defaults (JobPollInterval 5s, ConfirmationPollInterval 30s,
// RequiredConfirmations 12, ChainID 80002 for Amoy testnet) are not
// environment-configurable — they're fixed for this thesis's scope.
func LoadBlockchainConfig() (BlockchainConfig, error) {
	chainID, err := strconv.Atoi(getEnv("CHAIN_ID", "80002"))
	if err != nil {
		return BlockchainConfig{}, fmt.Errorf("invalid CHAIN_ID: %w", err)
	}

	cfg := BlockchainConfig{
		PolygonRPCURL:            getEnv("POLYGON_RPC_URL", ""),
		ContractAddress:          getEnv("CONTRACT_ADDRESS", ""),
		PrivateKey:               getEnv("BLOCKCHAIN_PRIVATE_KEY", ""),
		ChainID:                  chainID,
		JobPollInterval:          5 * time.Second,
		ConfirmationPollInterval: 30 * time.Second,
		RequiredConfirmations:    12,
	}

	if err := cfg.validate(); err != nil {
		return BlockchainConfig{}, err
	}
	return cfg, nil
}

// validate rejects obviously-malformed connection details before the worker
// ever attempts to use them, rather than failing on the first blockchain call.
func (c BlockchainConfig) validate() error {
	if c.PolygonRPCURL == "" {
		return fmt.Errorf("POLYGON_RPC_URL is required")
	}
	if !privateKeyPattern.MatchString(c.PrivateKey) {
		return fmt.Errorf("BLOCKCHAIN_PRIVATE_KEY must be 64 hex characters")
	}
	if !addressPattern.MatchString(c.ContractAddress) {
		return fmt.Errorf("CONTRACT_ADDRESS must be a 0x-prefixed 40 hex character address")
	}
	return nil
}
