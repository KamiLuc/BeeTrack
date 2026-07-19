package blockchain

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/beetrack/backend/internal/model"
)

// TestCanonicalMetadataHash_Golden locks the field order/format/separator
// spec in place: this exact input must always produce this exact hash, in
// any language, forever. If this test ever needs to change, the hashing
// spec itself changed and every already-certified batch's on-chain hash is
// now unreproducible from its DB row.
func TestCanonicalMetadataHash_Golden(t *testing.T) {
	batch := &model.HoneyBatch{
		ID:               42,
		ApiaryID:         7,
		GatheringDate:    time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		AmountGrams:      15000,
		ProcessingMethod: "raw",
		HoneyType:        "Lipowy",
		PDFFileHash:      "abc123",
	}

	want := "cf86220ecd8f403200488ac7cf758d14bbf7e434c1c4a1cd6866d67bcef41bc6"
	got := CanonicalMetadataHash(batch)

	if gotHex := hex.EncodeToString(got[:]); gotHex != want {
		t.Errorf("CanonicalMetadataHash() = %s, want %s", gotHex, want)
	}
}

func TestCanonicalMetadataHash_Deterministic(t *testing.T) {
	batch := &model.HoneyBatch{
		ID:               1,
		ApiaryID:         2,
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
		ApiaryID:         2,
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
