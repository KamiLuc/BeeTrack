package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// SHA256File computes the SHA256 hash of the file at filePath, streaming it
// rather than loading the whole file into memory.
func SHA256File(filePath string) ([32]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, fmt.Errorf("hash file: %w", err)
	}

	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return sum, nil
}

// DecodeHash32 decodes a lowercase-hex-encoded SHA256 hash (as stored on a
// HoneyBatch) into the fixed-size array the contract bindings expect.
func DecodeHash32(s string) ([32]byte, error) {
	var out [32]byte
	b, err := hex.DecodeString(s)
	if err != nil {
		return out, err
	}
	if len(b) != 32 {
		return out, fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	copy(out[:], b)
	return out, nil
}
