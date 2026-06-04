package model

import "time"

type Hive struct {
	ID              int64
	ApiaryID        int64
	Name            string
	Type            string
	Active          bool
	ReadyForHarvest bool
	Queenless       bool
	GridRow         int
	GridCol         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
