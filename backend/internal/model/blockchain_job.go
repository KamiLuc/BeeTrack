package model

import "time"

// BlockchainJob is one unit of durable, retryable blockchain work — currently
// always a "certify" job enqueued alongside a HoneyBatch. It survives process
// restarts, unlike an in-memory goroutine/channel queue.
type BlockchainJob struct {
	ID              int64
	BatchID         int64
	JobType         string
	Status          CertificationStatus
	AttemptCount    int
	NextRetryAt     time.Time
	LastError       *string
	CertificationID *int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
