package model

import "time"

// Treatment represents a medication or treatment applied to a hive.
type Treatment struct {
	ID            int64
	HiveID        int64
	TreatedBy     int64
	TreatedByName string `gorm:"-"`
	TreatedAt     time.Time
	MedicineName  string
	Dose          string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
