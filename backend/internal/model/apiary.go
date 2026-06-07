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

type ApiaryInvitation struct {
	ID              int64
	ApiaryID        int64
	InvitedByUserID int64
	InvitedEmail    string
	Status          string
	CreatedAt       time.Time
}

// MyInvitationView is the enriched read model returned to an invited user,
// joining apiary and inviter name for display.
type MyInvitationView struct {
	ID            int64
	ApiaryID      int64
	ApiaryName    string
	InvitedByName string
	CreatedAt     time.Time
}

type ApiaryMemberInfo struct {
	UserID   int64
	Name     string
	Email    string
	Role     string
	JoinedAt time.Time
}
