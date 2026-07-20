package model

import "time"

const (
	ListingStatusPending  = "pending"
	ListingStatusApproved = "approved"
	ListingStatusRejected = "rejected"
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
	Images          []ListingImage `gorm:"-"`
	ApiaryName      string         `gorm:"-"`
	ApiaryLat       *float64       `gorm:"-"`
	ApiaryLng       *float64       `gorm:"-"`
	ApiaryHiveCount int            `gorm:"-"`
	DistanceKm      *float64       `gorm:"-"`
}

// IsEdit distinguishes an edit of a previously-approved listing from a
// brand-new one — both sit at ListingStatusPending otherwise.
func (l *Listing) IsEdit() bool {
	return l.FirstApprovedAt != nil
}
