package model

import "time"

type HiveDisease struct {
	ID        int64
	HiveID    int64
	Disease   string
	CreatedAt time.Time
}
