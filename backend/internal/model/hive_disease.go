package model

import "time"

var ValidDiseases = []string{"american_foulbrood", "chalkbrood", "dwv", "european_foulbrood", "laying_workers", "nosema", "varroa"}

type HiveDisease struct {
	ID        int64
	HiveID    int64
	Disease   string
	CreatedAt time.Time
}
