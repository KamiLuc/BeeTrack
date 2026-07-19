package model

import "time"

// CertificationStatus is the lifecycle state of a single certification attempt.
// These values are the single source of truth mirrored by the DB CHECK
// constraints on honey_batch_certifications.status and blockchain_jobs.status,
// the API JSON, and the Dart CertificationStatus enum.
type CertificationStatus string

const (
	CertificationStatusQueued              CertificationStatus = "queued"
	CertificationStatusSubmitting          CertificationStatus = "submitting"
	CertificationStatusSubmitted           CertificationStatus = "submitted"
	CertificationStatusPendingConfirmation CertificationStatus = "pending_confirmation"
	CertificationStatusConfirmed           CertificationStatus = "confirmed"
	CertificationStatusFailed              CertificationStatus = "failed"
	CertificationStatusReverted            CertificationStatus = "reverted"
)

// IsTerminal reports whether the status will never transition again.
func (s CertificationStatus) IsTerminal() bool {
	switch s {
	case CertificationStatusConfirmed, CertificationStatusFailed, CertificationStatusReverted:
		return true
	default:
		return false
	}
}

// IsLive reports whether a certification in this status already occupies (or
// will occupy) the "one live certification per batch" slot enforced by the
// partial unique index on honey_batch_certifications. Used by the worker's
// idempotency check before submitting a new certification attempt.
func (s CertificationStatus) IsLive() bool {
	switch s {
	case CertificationStatusSubmitted, CertificationStatusPendingConfirmation, CertificationStatusConfirmed:
		return true
	default:
		return false
	}
}

// HoneyBatchCertification is a single certification attempt for a batch.
// Rows are append-only: a batch may accumulate multiple certifications across
// retries, so the current state for display purposes is simply the most
// recent row by CreatedAt.
type HoneyBatchCertification struct {
	ID                    int64
	BatchID               int64
	ChainID               int
	ContractAddress       string
	TransactionHash       *string
	BlockNumber           *int64
	Status                CertificationStatus
	GasUsed               *int64
	ConfirmationTimestamp *time.Time
	CreatedAt             time.Time
}
