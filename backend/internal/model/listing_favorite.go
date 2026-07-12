package model

import "time"

// ListingFavorite represents a user's saved (favorited) marketplace listing.
type ListingFavorite struct {
	ID        int64
	UserID    int64
	ListingID int64
	CreatedAt time.Time
}
