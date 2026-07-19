package model

import "time"

// ProcessingMethod is how a honey batch was processed before certification.
type ProcessingMethod string

const (
	ProcessingMethodRaw         ProcessingMethod = "raw"
	ProcessingMethodFiltered    ProcessingMethod = "filtered"
	ProcessingMethodPasteurized ProcessingMethod = "pasteurized"
)

// IsValidProcessingMethod reports whether method is one of the known processing methods.
func IsValidProcessingMethod(method string) bool {
	switch ProcessingMethod(method) {
	case ProcessingMethodRaw, ProcessingMethodFiltered, ProcessingMethodPasteurized:
		return true
	default:
		return false
	}
}

// HoneyBatch represents a batch of harvested honey submitted for blockchain
// certification. It carries no blockchain state itself — see
// HoneyBatchCertification for the append-only certification history.
type HoneyBatch struct {
	ID                int64
	UserID            int64
	ApiaryID          int64
	VerificationToken string
	GatheringDate     time.Time
	AmountGrams       int64
	ProcessingMethod  string
	HoneyType         string
	LabPDFURL         string
	PDFFileHash       string
	MetadataHash      string
	DeletedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
