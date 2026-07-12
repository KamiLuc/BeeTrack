package model

import "time"

// ListingImage represents an image attached to a marketplace listing.
type ListingImage struct {
	ID           int64
	ListingID    int64
	ImageURL     string
	DisplayOrder int
	CreatedAt    time.Time
}
