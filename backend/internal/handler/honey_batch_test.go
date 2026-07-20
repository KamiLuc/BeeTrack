package handler

import (
	"time"

	"testing"

	"github.com/beetrack/backend/internal/model"
)

func TestHoneyBatchJSON_NilCertification(t *testing.T) {
	b := &model.HoneyBatch{
		ID:               5,
		GatheringDate:    time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		AmountGrams:      15000,
		ProcessingMethod: "raw",
		HoneyType:        "Lipowy",
	}

	got := honeyBatchJSON(b, nil, nil)

	if got["certification"] != nil {
		t.Errorf("expected nil certification, got %v", got["certification"])
	}
	if got["honey_type"] != "Lipowy" {
		t.Errorf("expected honey_type Lipowy, got %v", got["honey_type"])
	}
}

func TestHoneyBatchJSON_WithCertification(t *testing.T) {
	b := &model.HoneyBatch{ID: 5}
	txHash := "0xabc"
	blockNum := int64(42)
	cert := &model.HoneyBatchCertification{
		Status:          model.CertificationStatusConfirmed,
		TransactionHash: &txHash,
		BlockNumber:     &blockNum,
	}

	got := honeyBatchJSON(b, cert, nil)

	certJSON, ok := got["certification"].(map[string]any)
	if !ok {
		t.Fatalf("expected certification to be a map, got %T", got["certification"])
	}
	if certJSON["status"] != model.CertificationStatusConfirmed {
		t.Errorf("expected status confirmed, got %v", certJSON["status"])
	}
	if certJSON["transaction_hash"] != &txHash {
		t.Errorf("expected transaction_hash %v, got %v", &txHash, certJSON["transaction_hash"])
	}
}
