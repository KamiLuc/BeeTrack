package model

import "time"

// Feeding represents a feeding (syrup, fondant, etc.) given to a hive.
type Feeding struct {
	ID        int64
	HiveID    int64
	FedBy     int64
	FedByName string `gorm:"-"`
	FedAt     time.Time
	FeedType  string
	Amount    string
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
