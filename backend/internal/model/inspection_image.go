package model

import "time"

type InspectionImage struct {
	ID           int64
	InspectionID int64
	Filename     string
	MimeType     string
	CreatedAt    time.Time
}
