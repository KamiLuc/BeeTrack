package model

import "time"

const (
	CertificationRequestStatusPending  = "pending"
	CertificationRequestStatusApproved = "approved"
	CertificationRequestStatusRejected = "rejected"
)

// HoneyBatchCertificationRequest gates BlockchainJob creation behind admin
// review: only Approve creates the job the existing worker picks up.
type HoneyBatchCertificationRequest struct {
	ID              int64
	BatchID         int64
	RequestedBy     int64
	Status          string
	RejectionReason *string
	ReviewedBy      *int64
	ReviewedAt      *time.Time
	BlockchainJobID *int64
	CreatedAt       time.Time
}

// HoneyBatchCertificationRequestDetail adds the batch/requester fields the admin
// queue/detail views need, plus (once approved) the linked blockchain_jobs and
// honey_batch_certifications state so the admin panel can show on-chain progress
// without a separate lookup.
type HoneyBatchCertificationRequestDetail struct {
	HoneyBatchCertificationRequest
	GatheringDate         time.Time
	AmountGrams           int64
	HoneyType             string
	ProcessingMethod      string
	RequesterEmail        string
	JobStatus             *CertificationStatus
	JobLastError          *string
	TransactionHash       *string
	BlockNumber           *int64
	ConfirmationTimestamp *time.Time
}
