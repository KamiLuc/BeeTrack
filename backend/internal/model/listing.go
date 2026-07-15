package model

import "time"

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
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Images          []ListingImage `gorm:"-"`
	ApiaryName      string         `gorm:"-"`
	ApiaryLat       *float64       `gorm:"-"`
	ApiaryLng       *float64       `gorm:"-"`
	ApiaryHiveCount int            `gorm:"-"`
	DistanceKm      *float64       `gorm:"-"`
}
