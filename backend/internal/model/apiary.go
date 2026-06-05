package model

import "time"

type Apiary struct {
	ID           int64
	OwnerUserID  int64
	Name     string
	Lat      *float64
	Lng          *float64
	GridRows     int
	GridCols     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ApiaryMember struct {
	ApiaryID int64
	UserID   int64
	Role     string
	JoinedAt time.Time
}

type ApiaryMembership struct {
	Apiary          *Apiary
	HiveCount       int
	UserRole        string
	LastInspectedAt *time.Time
}
