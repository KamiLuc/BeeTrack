package model

import "time"

const (
	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)

type User struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	Verified     bool
	Role         string `gorm:"default:user"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}
