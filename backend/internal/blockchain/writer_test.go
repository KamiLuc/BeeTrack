package blockchain

import (
	"errors"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/config"
)

// dataError implements the ErrorData() side channel go-ethereum's RPC errors
// use to carry raw revert bytes — the plain Error() text never includes the
// decoded custom error name.
type dataError struct {
	data string
}

func (e dataError) Error() string          { return "execution reverted" }
func (e dataError) ErrorData() interface{} { return e.data }

func TestIsAlreadyCertifiedRevert(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "AlreadyCertified(4) revert data",
			err:  dataError{data: "0x3b71c3b00000000000000000000000000000000000000000000000000000000000000004"},
			want: true,
		},
		{
			name: "AlreadyCertified(5) revert data",
			err:  dataError{data: "0x3b71c3b00000000000000000000000000000000000000000000000000000000000000005"},
			want: true,
		},
		{
			name: "a different custom error's selector (EmptyHashes)",
			err:  dataError{data: "0x4e9c062e"},
			want: false,
		},
		{
			name: "plain error with no ErrorData at all",
			err:  errors.New("execution reverted"),
			want: false,
		},
		{
			name: "ErrorData present but not a hex string",
			err:  dataError{data: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlreadyCertifiedRevert(tt.err); got != tt.want {
				t.Errorf("isAlreadyCertifiedRevert() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
