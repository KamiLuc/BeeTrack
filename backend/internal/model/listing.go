package model

import "time"

const (
	ListingStatusPending  = "pending"
	ListingStatusApproved = "approved"
	ListingStatusRejected = "rejected"
	ListingStatusRemoved  = "removed"
)

// Listing represents a marketplace sale announcement posted by a user.
type Listing struct {
	ID              int64
	UserID          int64
	Title           string
	Description     string
	Category        string
	Price           *float64
	Quantity        string
	Address         string
	Lat             float64
	Lng             float64
	ApiaryID        *int64
	ContactPhone    string
	ContactEmail    string
	IsHidden        bool
	Status          string `gorm:"default:pending"`
	RejectionReason *string
	FirstApprovedAt *time.Time
	ReviewedBy      *int64
	ReviewedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	HoneyBatchID    *int64
	Images          []ListingImage `gorm:"-"`
	ApiaryName      string         `gorm:"-"`
	ApiaryLat       *float64       `gorm:"-"`
	ApiaryLng       *float64       `gorm:"-"`
	ApiaryHiveCount int            `gorm:"-"`
	DistanceKm      *float64       `gorm:"-"`
	OwnerEmail      string         `gorm:"-"`

	// HoneyBatch* fields are populated only by ListingService.Get, from the
	// attached honey_batches row (if any) — never scanned by gorm directly.
	HoneyBatchHoneyType           string     `gorm:"-"`
	HoneyBatchGatheringDate       *time.Time `gorm:"-"`
	HoneyBatchAmountGrams         *int64     `gorm:"-"`
	HoneyBatchProcessingMethod    string     `gorm:"-"`
	HoneyBatchCertificationStatus string     `gorm:"-"`
	HoneyBatchHasPDF              bool       `gorm:"-"`
	HoneyBatchVerificationURL     string     `gorm:"-"`
	HoneyBatchPDFURL              string     `gorm:"-"`

	// CertificationRequest* fields are populated only by
	// ListingModerationService.Get, from the attached honey batch's latest
	// certification request (if any) — used by the admin panel to link a
	// HONEY listing to its certification review.
	CertificationRequestID     *int64 `gorm:"-"`
	CertificationRequestStatus string `gorm:"-"`
}

// IsEdit distinguishes an edit of a previously-approved listing from a
// brand-new one — both sit at ListingStatusPending otherwise.
func (l *Listing) IsEdit() bool {
	return l.FirstApprovedAt != nil
}
