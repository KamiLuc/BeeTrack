package model

import "time"

type Inspection struct {
	ID                    int64
	HiveID                int64
	InspectedBy           int64
	InspectedByName       string `gorm:"-"`
	InspectedAt           time.Time
	QueenStatus           string
	BroodPattern          string
	FramesBrood           *int
	FramesFeed            *int
	FramesPollen          *int
	QueenCellsCount       *int
	Aggressiveness        string
	FramesAddedFoundation *int
	FramesAddedDrawn      *int
	FramesAddedBrood      *int
	FramesAddedFeed       *int
	QueenAdded            bool
	Notes                 string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
