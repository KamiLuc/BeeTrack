package model

import "time"

type User struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	Verified     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
