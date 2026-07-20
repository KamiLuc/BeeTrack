package blockchain

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

// TestCanonicalMetadataHash_Golden pins this exact input to this exact hash
// forever — changing it means the hashing spec changed and every
// already-certified batch's hash is unreproducible from its DB row.
func TestCanonicalMetadataHash_Golden(t *testing.T) {
	batch := &model.HoneyBatch{
		ID:               42,
		GatheringDate:    time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		AmountGrams:      15000,
		ProcessingMethod: "raw",
		HoneyType:        "Lipowy",
		PDFFileHash:      "abc123",
	}

	want := "a93889e0bb3e26b147d9c2d9c60de3ff609bcff22fa73b5fd280ca640a757363"
	got := CanonicalMetadataHash(batch)

	if gotHex := hex.EncodeToString(got[:]); gotHex != want {
		t.Errorf("CanonicalMetadataHash() = %s, want %s", gotHex, want)
	}
}

func TestCanonicalMetadataHash_Deterministic(t *testing.T) {
	batch := &model.HoneyBatch{
		ID:               1,
		GatheringDate:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		AmountGrams:      1000,
		ProcessingMethod: "filtered",
		HoneyType:        "Wielokwiatowy",
		PDFFileHash:      "deadbeef",
	}

	a := CanonicalMetadataHash(batch)
	b := CanonicalMetadataHash(batch)
	if a != b {
		t.Errorf("CanonicalMetadataHash() is not deterministic: %x != %x", a, b)
	}
}

func TestCanonicalMetadataHash_DifferentAmountDifferentHash(t *testing.T) {
	base := &model.HoneyBatch{
		ID:               1,
		GatheringDate:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ProcessingMethod: "raw",
		HoneyType:        "Rzepakowy",
		PDFFileHash:      "deadbeef",
	}

	a := *base
	a.AmountGrams = 1000
	b := *base
	b.AmountGrams = 1001

	if CanonicalMetadataHash(&a) == CanonicalMetadataHash(&b) {
		t.Error("expected different amounts to produce different hashes")
	}
}
