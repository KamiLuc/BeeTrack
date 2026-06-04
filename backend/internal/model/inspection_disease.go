package model

import "time"

type InspectionDisease struct {
	ID           int64
	InspectionID int64
	Disease      string
	Notes        string
	CreatedAt    time.Time
}
