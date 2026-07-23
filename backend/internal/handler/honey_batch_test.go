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

	got := honeyBatchJSON(b, nil, nil, "https://example.com/verify/tok")

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

	got := honeyBatchJSON(b, cert, nil, "https://example.com/verify/tok")

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
	if certJSON["chain_id"] != cert.ChainID {
		t.Errorf("expected chain_id %v, got %v", cert.ChainID, certJSON["chain_id"])
	}
	if certJSON["contract_address"] != cert.ContractAddress {
		t.Errorf("expected contract_address %v, got %v", cert.ContractAddress, certJSON["contract_address"])
	}
}

func TestHoneyBatchJSON_IncludesIDAndVerificationURL(t *testing.T) {
	b := &model.HoneyBatch{ID: 5, VerificationToken: "tok"}

	got := honeyBatchJSON(b, nil, nil, "https://example.com/verify/tok")

	if got["id"] != int64(5) {
		t.Errorf("expected id 5, got %v", got["id"])
	}
	if got["verification_url"] != "https://example.com/verify/tok" {
		t.Errorf("expected verification_url set, got %v", got["verification_url"])
	}
	if _, ok := got["certification_request"]; !ok {
		t.Error("expected certification_request key to be present")
	}
}

func TestPublicHoneyBatchJSON_ExcludesIDAndCertificationRequest(t *testing.T) {
	b := &model.HoneyBatch{ID: 5, VerificationToken: "tok", MetadataHash: "hash"}

	got := publicHoneyBatchJSON(b, nil, "https://example.com/verify/tok")

	if _, ok := got["id"]; ok {
		t.Errorf("expected no id key in public JSON, got %v", got["id"])
	}
	if _, ok := got["certification_request"]; ok {
		t.Errorf("expected no certification_request key in public JSON, got %v", got["certification_request"])
	}
	if got["metadata_hash"] != "hash" {
		t.Errorf("expected metadata_hash to be set, got %v", got["metadata_hash"])
	}
	if got["verification_url"] != "https://example.com/verify/tok" {
		t.Errorf("expected verification_url set, got %v", got["verification_url"])
	}
}
