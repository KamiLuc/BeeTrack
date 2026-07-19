package config

import (
	"strings"
	"testing"
)

func TestBlockchainConfigValidate(t *testing.T) {
	validKey := strings.Repeat("ab", 32)
	validAddress := "0x1234567890123456789012345678901234567890"

	tests := []struct {
		name    string
		cfg     BlockchainConfig
		wantErr bool
	}{
		{
			name: "valid",
			cfg: BlockchainConfig{
				PolygonRPCURL:   "https://rpc-amoy.polygon.technology",
				PrivateKey:      validKey,
				ContractAddress: validAddress,
			},
			wantErr: false,
		},
		{
			name: "missing rpc url",
			cfg: BlockchainConfig{
				PolygonRPCURL:   "",
				PrivateKey:      validKey,
				ContractAddress: validAddress,
			},
			wantErr: true,
		},
		{
			name: "private key too short",
			cfg: BlockchainConfig{
				PolygonRPCURL:   "https://rpc-amoy.polygon.technology",
				PrivateKey:      "abc123",
				ContractAddress: validAddress,
			},
			wantErr: true,
		},
		{
			name: "contract address missing 0x prefix",
			cfg: BlockchainConfig{
				PolygonRPCURL:   "https://rpc-amoy.polygon.technology",
				PrivateKey:      validKey,
				ContractAddress: validAddress[2:],
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
