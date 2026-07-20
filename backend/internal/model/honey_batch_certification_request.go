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

// HoneyBatchCertificationRequestDetail adds the batch/requester fields the admin queue/detail views need.
type HoneyBatchCertificationRequestDetail struct {
	HoneyBatchCertificationRequest
	GatheringDate  time.Time
	AmountGrams    int64
	HoneyType      string
	RequesterEmail string
}
