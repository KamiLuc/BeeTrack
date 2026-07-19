package blockchain

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// TestEmbeddedABI_Parses guards against a broken go:embed path or malformed
// HoneyCertification.abi silently breaking at runtime.
func TestEmbeddedABI_Parses(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(honeyCertificationABIJSON))
	if err != nil {
		t.Fatalf("failed to parse embedded ABI: %v", err)
	}

	for _, name := range []string{"certify", "getCertification", "minter", "setMinter"} {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("expected ABI to contain method %q", name)
		}
	}

	for _, name := range []string{"AlreadyCertified", "EmptyHashes", "NotCertified", "NotMinter", "ZeroAddress"} {
		if _, ok := parsed.Errors[name]; !ok {
			t.Errorf("expected ABI to contain error %q", name)
		}
	}

	for _, name := range []string{"CertificationCreated", "MinterUpdated"} {
		if _, ok := parsed.Events[name]; !ok {
			t.Errorf("expected ABI to contain event %q", name)
		}
	}
}
