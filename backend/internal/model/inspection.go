package model

import "time"

var (
	ValidQueenStatuses   = []string{"not_seen", "seen"}
	ValidBroodPatterns   = []string{"excellent", "good", "none", "poor"}
	ValidAggressiveness  = []string{"calm", "mild", "aggressive", "very_aggressive"}
	ValidColonyStrengths = []string{"very_weak", "weak", "medium", "strong", "very_strong"}
)

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
	ColonyStrength        string
	FramesAddedFoundation *int
	FramesAddedDrawn      *int
	FramesAddedBrood      *int
	FramesAddedFeed       *int
	QueenAdded            bool
	BoxAdded              bool
	Notes                 string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
