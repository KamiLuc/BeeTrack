package model

import "time"

type Hive struct {
	ID                         int64
	ApiaryID                   int64
	Name                       string
	Type                       string
	Active                     bool
	ReadyForHarvest            bool
	ReadyForHarvestSince       *time.Time
	QueenNeedsReplacement      bool
	QueenNeedsReplacementSince *time.Time
	NeedsFood                  bool
	NeedsFoodSince             *time.Time
	BoxNeedsAdding             bool
	BoxNeedsAddingSince        *time.Time
	GridRow                    int
	GridCol                    int
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}
