package model

import "time"

// Harvest represents a honey harvest from a hive.
type Harvest struct {
	ID              int64
	HiveID          int64
	HarvestedBy     int64
	HarvestedByName string `gorm:"-"`
	HarvestedAt     time.Time
	Frames          int
	HalfFrames      int
	Kilograms       float64
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
